/**
 * Safe-output handler: publishes the agent's reproduction draft as the issue's sticky comment.
 *
 * The draft reaches this job as an artifact, not a string input (the safe-output gateway caps string
 * inputs at 10 KiB and a reproduction comment that inlines a unit test or `.tf` can exceed that). The
 * job takes no inputs: the resolved draft path is the whole contract. This mirrors
 * `research-factory/update-research-comment.js`, but the reproducer contract is far smaller — there is
 * no critique loop, no quality-gate forcing, and no structural rewrite of the body.
 *
 * Pipeline: detection-gate -> resolve draft -> reject empty -> sanitise (strip HTML comments and
 * control/invisible chars) -> neutralise @mentions -> prepend the marker -> validate (byte cap +
 * metadata enum) on the exact string that will be published -> upsert the one sticky comment.
 *
 * Two fatal checks, deliberately narrow (D5): the byte cap and the pipeline-metadata enum. The prose
 * half of the contract (section names) stays advisory — failing a paid agent run over a missing `###`
 * heading is a worse outcome than publishing a slightly imperfect comment. What is NOT advisory is the
 * `outcome` value: an LLM spelling an enum is not an acceptance criterion, so an `outcome` outside the
 * three values (D1) fails the run and posts nothing.
 */
'use strict';

const fs = require('fs');
const path = require('path');

const { stripHtmlComments, stripControlChars, stripInvisibleUnicode, neutralizeMentions, findBotCommentByMarker } = require('../lib/sanitize-context.js');
const { positiveIssueNumber } = require('../lib/issue-number.js');
const { REPRODUCER_COMMENT_MARKER } = require('../lib/intake/reproducer-factory-constants.js');
const { validateCommentBody } = require('./validate-metadata.js');

/** Exactly what the safe job's `artifacts:` entry names. */
const REPRO_DIR = '/tmp/gh-aw/agent/repro/';
const DRAFT_FINAL = 'draft-final.md';
/** upload-artifact strips this common ancestor from the uploaded paths. */
const ARTIFACT_ROOT = '/tmp/gh-aw/';

/** GitHub's own ceiling is 65 536; the rest is reserved for the marker and any editor chrome. */
const MAX_COMMENT_CHARS = 60000;

/**
 * Resolves the agent's final draft.
 *
 * The agent writes `/tmp/gh-aw/agent/repro/draft-final.md`; `upload-artifact` strips the `/tmp/gh-aw/`
 * common ancestor, and the safe job downloads the artifact into `${RUNNER_TEMP}/gh-aw/safe-jobs/`, so
 * the same file arrives as `<safe-jobs>/agent/repro/draft-final.md`. Both forms are tried, then a
 * bounded search of the download directory, so a gh-aw layout change degrades into a slower lookup
 * rather than a discarded run.
 *
 * @param {NodeJS.ProcessEnv} env
 * @returns {{path: string, candidates: string[], fallback?: string[]} | {error: string, candidates: string[]}}
 */
function resolveDraftPath(env) {
  const safeJobsDir = env.GH_AW_AGENT_OUTPUT
    ? path.dirname(env.GH_AW_AGENT_OUTPUT)
    : path.join(env.RUNNER_TEMP || '/tmp', 'gh-aw', 'safe-jobs');
  const relative = path.join(path.relative(ARTIFACT_ROOT, REPRO_DIR), DRAFT_FINAL);

  const candidates = [path.join(safeJobsDir, relative), path.join(REPRO_DIR, DRAFT_FINAL)];

  for (const candidate of candidates) {
    if (fs.existsSync(candidate) && fs.statSync(candidate).isFile()) {
      return { path: candidate, candidates };
    }
  }

  const found = findDraftUnder(safeJobsDir, 6);
  if (found.length > 0) {
    return { path: found.sort()[0], candidates, fallback: found };
  }

  return { error: `no ${DRAFT_FINAL} in the agent artifact`, candidates };
}

/**
 * Bounded recursive search for the final draft under a downloaded artifact.
 *
 * @param {string} dir
 * @param {number} depth
 * @returns {string[]}
 */
function findDraftUnder(dir, depth) {
  if (depth < 0) {
    return [];
  }
  let dirEntries;
  try {
    dirEntries = fs.readdirSync(dir, { withFileTypes: true });
  } catch {
    return [];
  }
  const matches = [];
  for (const entry of dirEntries) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      matches.push(...findDraftUnder(full, depth - 1));
    } else if (entry.isFile() && entry.name === DRAFT_FINAL) {
      matches.push(full);
    }
  }
  return matches;
}

/**
 * Turns a raw draft into the exact string to publish, or explains why it cannot.
 *
 * @param {string} draft
 * @returns {{body: string | null, errors: string[]}}
 */
function preparePublishedBody(draft) {
  if (typeof draft !== 'string' || draft.trim() === '') {
    return { body: null, errors: ['draft is empty'] };
  }

  // D9: strip HTML comments so an injected `<!-- gha-issue-classifier -->` in quoted issue text cannot
  // hijack another workflow's sticky comment, then strip control/invisible characters. The marker is
  // prepended afterwards, so it is the only `<!--` that survives. If a stray unterminated `<!--` in the
  // draft caused the strip to eat the metadata envelope, the validate step below fails closed.
  const sanitized = stripInvisibleUnicode(stripControlChars(stripHtmlComments(draft)));
  // Neutralise every @mention so a quoted `@org/team` cannot mass-ping. This runs after the invisible
  // characters are stripped, because it works by inserting one. It parses no markdown, so an `@` in
  // a string of the metadata envelope gets the same invisible character; the envelope stays valid
  // JSON and none of its validated fields can contain an `@`.
  const neutralized = neutralizeMentions(sanitized);

  const body = `${REPRODUCER_COMMENT_MARKER}\n${neutralized.replace(/^\n+/, '')}`;

  const errors = [];
  if (body.length > MAX_COMMENT_CHARS) {
    errors.push(`comment body is ${body.length} characters, over the ${MAX_COMMENT_CHARS} limit`);
  }

  // The marker must be the one and only HTML comment opener, and must lead the body.
  const openerCount = body.split('<!--').length - 1;
  if (openerCount !== 1 || !body.startsWith(REPRODUCER_COMMENT_MARKER)) {
    errors.push(`the body carries ${openerCount} HTML comment opener(s); only the leading marker is allowed`);
  }

  // D1/D5: the pipeline metadata must be present, parseable, and carry an in-enum outcome. This is the
  // machine-checkable half of AC #1 and is fatal.
  const { errors: metadataErrors } = validateCommentBody(body);
  errors.push(...metadataErrors);

  if (errors.length > 0) {
    return { body: null, errors };
  }
  return { body, errors: [] };
}

module.exports = async function ({ github, context, core }) {
  const env = process.env;

  // Detection steps run with continue-on-error, so the job can report result=success while the
  // detector warned, failed, or emitted nothing. Allow exactly a clean verdict; fail closed on
  // warning, failure, an empty conclusion, or any unknown value. gh-aw does not inject
  // DETECTION_RESULT, so the workflow passes it explicitly.
  if (env.DETECTION_RESULT !== 'success' || !['success', 'skipped'].includes(env.GH_AW_DETECTION_CONCLUSION)) {
    core.setFailed(
      `refusing to publish: detection result=${env.DETECTION_RESULT} conclusion=${env.GH_AW_DETECTION_CONCLUSION}`,
    );
    return;
  }

  const issueNumber = positiveIssueNumber(env.INPUT_ISSUE_NUMBER);
  if (!issueNumber) {
    core.setFailed(`refusing to publish: invalid issue number '${env.INPUT_ISSUE_NUMBER || ''}'.`);
    return;
  }

  const resolved = resolveDraftPath(env);
  if (resolved.error) {
    core.setFailed(
      `${resolved.error}. Looked in: ${resolved.candidates.join(', ')}. ` +
        'A missing draft can also mean secret redaction failed, which drops the whole agent artifact.',
    );
    return;
  }
  if (resolved.fallback) {
    core.warning(
      `resolved the draft by searching the artifact: using ${resolved.path}` +
        (resolved.fallback.length > 1 ? ` (${resolved.fallback.length} matches: ${resolved.fallback.join(', ')})` : ''),
    );
  }

  let draft;
  try {
    draft = fs.readFileSync(resolved.path, 'utf8');
  } catch (err) {
    core.setFailed(`cannot read ${resolved.path}: ${err.message}`);
    return;
  }

  let prepared;
  try {
    prepared = preparePublishedBody(draft);
  } catch (err) {
    core.setFailed(`refusing to publish ${resolved.path}: the handler threw while preparing the body: ${err.stack || err.message}`);
    return;
  }
  const { body, errors } = prepared;
  if (!body) {
    core.setFailed(
      `refusing to publish ${resolved.path}: ${errors.join('; ')}. ` +
        'Re-running this job cannot fix it — the artifact carries the same draft; re-apply the reproducer-factory label for a fresh run.',
    );
    return;
  }

  const { owner, repo } = context.repo;
  const botLogin = env.GH_AW_BOT_LOGIN || 'github-actions[bot]';

  try {
    // The number is re-read raw here (a custom safe-output job cannot depend on pre_activation), so the
    // pull-request check is repeated at the boundary that writes. The comment and label endpoints
    // accept pull requests, so without this a dispatch naming a PR gets a reproduction comment stamped.
    const { data: target } = await github.rest.issues.get({ owner, repo, issue_number: issueNumber });
    if (target.pull_request) {
      core.setFailed(`refusing to publish: #${issueNumber} is a pull request, not an issue.`);
      return;
    }
  } catch (err) {
    core.setFailed(`cannot read #${issueNumber}: ${err.message}`);
    return;
  }

  let existing = null;
  try {
    const comments = await github.paginate(github.rest.issues.listComments, {
      owner,
      repo,
      issue_number: issueNumber,
      per_page: 100,
    });
    existing = findBotCommentByMarker(comments, REPRODUCER_COMMENT_MARKER, botLogin);
  } catch (err) {
    core.setFailed(`cannot list comments on #${issueNumber}: ${err.message}`);
    return;
  }

  try {
    if (existing) {
      await github.rest.issues.updateComment({ owner, repo, comment_id: existing.id, body });
      core.info(`Updated reproducer comment ${existing.id} on #${issueNumber} (${body.length} chars).`);
    } else {
      const { data } = await github.rest.issues.createComment({ owner, repo, issue_number: issueNumber, body });
      core.info(`Created reproducer comment ${data.id} on #${issueNumber} (${body.length} chars).`);
    }
  } catch (err) {
    core.setFailed(`failed to publish the reproducer comment on #${issueNumber}: ${err.message}`);
    return;
  }

  core.setOutput('reproducer_comment_published', 'true');
};

/**
 * Dual-use CLI, for the agent's pre-flight: `node update-reproducer-comment.js <draft.md>` runs the
 * exact body checks this job runs at publish time and exits non-zero if the draft would be rejected.
 * github-script loads this file with `require`, so `require.main` is the github-script runner and
 * `main()` never fires there.
 */
function main() {
  const draftPath = process.argv[2];

  if (!draftPath) {
    console.error('usage: node update-reproducer-comment.js <draft.md>');
    process.exit(2);
  }

  let draft;
  try {
    draft = fs.readFileSync(draftPath, 'utf8');
  } catch (err) {
    console.error(`cannot read ${draftPath}: ${err.message}`);
    process.exit(2);
  }

  const { body, errors } = preparePublishedBody(draft);
  if (!body) {
    console.error(`${draftPath} would be REJECTED by the publishing job:`);
    for (const error of errors) {
      console.error(`  - ${error}`);
    }
    process.exit(1);
  }

  console.log(`${draftPath}: would publish ${body.length} of ${MAX_COMMENT_CHARS} characters.`);
}

if (require.main === module) {
  main();
}

module.exports.preparePublishedBody = preparePublishedBody;
module.exports.resolveDraftPath = resolveDraftPath;
module.exports.REPRO_DIR = REPRO_DIR;
module.exports.DRAFT_FINAL = DRAFT_FINAL;
module.exports.MAX_COMMENT_CHARS = MAX_COMMENT_CHARS;
