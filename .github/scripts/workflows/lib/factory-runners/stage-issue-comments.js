/**
 * Fetches every comment on the triggering issue and stages it on disk for a later runner.
 *
 * Bot comments are kept: the research-factory sticky comment a change-factory run consumes is
 * authored by github-actions[bot]. When the history exceeds the cap the NEWEST comments are
 * kept for discussion context, PLUS the latest research-marked bot comment even when it falls
 * outside that window — on a busy issue the sticky comment is created early and is not the newest,
 * and editing it does not change its creation order, so a plain tail cap would drop the research
 * baseline and the agent would fall back to the issue despite one existing.
 *
 * The staged JSON is raw, unsanitized comment text, so it is written under RUNNER_TEMP and
 * never into the factory context directory that is uploaded to the agent.
 */
const { getFactoryName } = require('./_factory-context.js');
const { findResearchComment } = require('../sanitize-context.js');
const { RESEARCH_COMMENT_MARKER } = require('../intake/research-factory-constants.js');

const MAX_COMMENTS = 200;

module.exports = async function ({ github, context, core }) {

  const fs = require('fs');
  const factoryName = getFactoryName();
  const { owner, repo } = context.repo;
  const issueNumber = context.payload.issue?.number;

  if (!issueNumber) {
    core.setOutput('issue_comments_path', '');
    core.setOutput('issue_comments_count', '0');
    core.setOutput('comments_truncated', 'false');
    core.info('No issue number in payload; skipping comment fetch.');
    return;
  }

  const allComments = await github.paginate(github.rest.issues.listComments, {
    owner,
    repo,
    issue_number: issueNumber,
    per_page: 100,
  });

  const mapped = (allComments ?? []).map(c => ({
    author: c.user?.login ?? '',
    createdAt: c.created_at ?? '',
    body: c.body ?? '',
  }));

  const comments = mapped.slice(-MAX_COMMENTS);
  const truncated = mapped.length > MAX_COMMENTS;

  // If the newest research-marked bot comment was dropped by the tail cap, prepend it so the
  // extraction step still sees the research baseline. Order is preserved (it is older than the
  // retained window). Harmless no-op for factories whose issues carry no such comment.
  const research = findResearchComment(mapped, RESEARCH_COMMENT_MARKER, process.env.GH_AW_BOT_LOGIN || 'github-actions[bot]');
  if (research && !comments.includes(research)) {
    comments.unshift(research);
  }

  // Unset RUNNER_TEMP would interpolate the literal 'undefined' and write raw comment text into
  // the checkout. Fall back to the OS temp dir: still outside the uploaded context directory.
  const runnerTemp = process.env.RUNNER_TEMP || require('os').tmpdir();
  if (!process.env.RUNNER_TEMP) {
    core.info(`RUNNER_TEMP is not set; staging comments under ${runnerTemp} instead.`);
  }
  const dir = `${runnerTemp}/${factoryName}-raw`;
  const commentsPath = `${dir}/comments.json`;
  fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(commentsPath, JSON.stringify({ comments, truncated }));

  core.setOutput('issue_comments_path', commentsPath);
  core.setOutput('issue_comments_count', String(comments.length));
  core.setOutput('comments_truncated', truncated ? 'true' : 'false');
  core.info(
    `Staged ${comments.length} comments for issue #${issueNumber} at ${commentsPath}`
    + `${truncated ? ` (kept the newest ${MAX_COMMENTS} of ${mapped.length})` : ''}`,
  );
};
