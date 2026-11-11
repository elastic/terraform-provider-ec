/**
 * Renders and validates the `.changelog/{PR}.txt` fragment for the code-factory `add_changelog`
 * safe-job.
 *
 * The code-factory workflow's `add_changelog` custom safe-job invokes this with `--render-only`:
 * it reads the agent's safe-output items from `$GH_AW_AGENT_OUTPUT`, asserts exactly one
 * `add_changelog` item, validates the release-note type and body against the same contract
 * `scripts/changelogcheck` (Go) enforces, and prints the fenced fragment to stdout. The workflow
 * step then commits that fragment to `.changelog/{PR}.txt` on the PR's own branch (derived from the
 * trusted created-PR number, not from any agent-supplied value). This module makes no network/API
 * call; it fails (non-zero exit) on a malformed item so a bad agent call never reaches a commit.
 *
 * ALLOWED_TYPES MUST stay in sync with `allowedTypes` in
 * `scripts/changelogcheck/validate.go`. A drift test
 * (`code-factory-changelog.test.mjs`) asserts the two lists match; do not edit one without the
 * other.
 */
'use strict';

const ALLOWED_TYPES = [
  'breaking-change',
  'note',
  'feature',
  'new-resource',
  'new-data-source',
  'new-guide',
  'enhancement',
  'bug',
  'none',
];

const ITEM_TYPE = 'add_changelog';

/**
 * Selects the single `add_changelog` item from a parsed `$GH_AW_AGENT_OUTPUT` object.
 *
 * A custom safe-job runs once per workflow run; every tool call is collected into `.items[]`, so
 * the agent calling the tool twice is a hard error rather than a silently dropped second call.
 *
 * @param {unknown} parsed - the JSON parsed from $GH_AW_AGENT_OUTPUT
 * @returns {{ type: string, issue_number: unknown, release_note_type: unknown, release_note_body: unknown }}
 * @throws {Error} when there is not exactly one add_changelog item
 */
function selectChangelogItem(parsed) {
  if (parsed === null || typeof parsed !== 'object' || !Array.isArray(parsed.items)) {
    throw new Error('agent output is missing an "items" array');
  }
  const items = parsed.items.filter((it) => it && it.type === ITEM_TYPE);
  if (items.length !== 1) {
    throw new Error(`expected exactly one ${ITEM_TYPE} item, found ${items.length}`);
  }
  return items[0];
}

/**
 * Validates a changelog item's type and body against the release-note contract.
 *
 * @param {{ release_note_type: unknown, release_note_body: unknown }} item
 * @returns {{ type: string, body: string }}
 * @throws {Error} on any contract violation
 */
function validateChangelogItem(item) {
  const type = String(item.release_note_type ?? '').trim();
  if (!ALLOWED_TYPES.includes(type)) {
    throw new Error(
      `release_note_type ${JSON.stringify(item.release_note_type)} is not one of: ${ALLOWED_TYPES.join(', ')}`,
    );
  }

  const body = String(item.release_note_body ?? '').trim();
  // Mirror scripts/changelogcheck/validate.go: none => empty body, otherwise non-empty body.
  if (type === 'none' && body !== '') {
    throw new Error('release-note:none must have an empty body');
  }
  if (type !== 'none') {
    if (body === '') {
      throw new Error(`release-note:${type} must have a non-empty body`);
    }
    // The body is rendered verbatim between the ```release-note:<type> fences, so it must be a
    // single line and must not open/close a code fence — otherwise the rendered fragment is
    // malformed (an empty note, or an extra/`releasenote` fence) and scripts/changelogcheck rejects
    // it. Enforce the documented one-line, no-fence contract before rendering.
    if (/[\r\n]/.test(body)) {
      throw new Error(`release-note:${type} body must be a single line (no newlines)`);
    }
    if (body.includes('```')) {
      throw new Error(`release-note:${type} body must not contain a code fence (\`\`\`)`);
    }
  }

  return { type, body };
}

/**
 * Renders the fenced release-note fragment. `none` yields an empty body between the fences.
 *
 * @param {{ type: string, body: string }} validated
 * @returns {string}
 */
function renderFragment({ type, body }) {
  if (type === 'none') {
    return '```release-note:none\n```\n';
  }
  return `\`\`\`release-note:${type}\n${body}\n\`\`\`\n`;
}

/**
 * Reads $GH_AW_AGENT_OUTPUT, selects + validates the single add_changelog item, and returns the
 * rendered fragment.
 *
 * @param {NodeJS.ProcessEnv} env
 * @param {typeof import('fs')} fs
 * @returns {{ fragment: string }}
 */
function buildFromEnv(env, fs) {
  const outputPath = env.GH_AW_AGENT_OUTPUT;
  if (!outputPath) {
    throw new Error('GH_AW_AGENT_OUTPUT is not set');
  }
  const parsed = JSON.parse(fs.readFileSync(outputPath, 'utf8'));
  const item = selectChangelogItem(parsed);
  const validated = validateChangelogItem(item);
  return { fragment: renderFragment(validated) };
}

function main(argv, env, fs, stdout, stderr) {
  const renderOnly = argv.includes('--render-only');
  if (!renderOnly) {
    stderr.write('usage: code-factory-changelog.js --render-only\n');
    return 2;
  }
  try {
    const { fragment } = buildFromEnv(env, fs);
    stdout.write(fragment);
    return 0;
  } catch (err) {
    stderr.write(`code-factory-changelog: ${err.message}\n`);
    return 1;
  }
}

if (require.main === module) {
  // eslint-disable-next-line global-require
  const fs = require('fs');
  process.exit(main(process.argv.slice(2), process.env, fs, process.stdout, process.stderr));
}

module.exports = {
  ALLOWED_TYPES,
  ITEM_TYPE,
  selectChangelogItem,
  validateChangelogItem,
  renderFragment,
  buildFromEnv,
  main,
};
