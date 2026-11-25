/**
 * Pre-activation lookup of the reproducer factory's own sticky comment.
 *
 * The agent reads the result as draft input (`prior_reproducer_comment.md`), and the safe-output
 * handler performs the same lookup again when it publishes. Both go through `findBotCommentByMarker`
 * so the bot+marker filter has one definition.
 *
 * A thin per-factory copy of `research-factory/fetch-prior-research-comment.js` — the only
 * factory-specific values are the marker and the output name, both below. See that file's header for
 * the three rules this encodes (paginate without sort/direction; REST only; resolve the bot login with
 * `||`, not a default parameter). A lookup failure warns rather than fails: losing the prior comment
 * costs the agent its draft input, but the run can still produce a reproduction.
 */
'use strict';

const { findBotCommentByMarker } = require('../lib/sanitize-context.js');
const { positiveIssueNumber } = require('../lib/issue-number.js');
const { REPRODUCER_COMMENT_MARKER } = require('../lib/intake/reproducer-factory-constants.js');
const { boundEnvPayload } = require('../lib/env-payload.js');

module.exports = async function ({ github, context, core }) {
  const { owner, repo } = context.repo;
  const issueNumber = positiveIssueNumber(process.env.INPUT_ISSUE_NUMBER);
  const botLogin = process.env.GH_AW_BOT_LOGIN || 'github-actions[bot]';

  if (!issueNumber) {
    core.setOutput('prior_reproducer_comment', '');
    core.info('No prior reproducer comment lookup: invalid issue number.');
    return;
  }

  try {
    const allComments = await github.paginate(github.rest.issues.listComments, {
      owner,
      repo,
      issue_number: issueNumber,
      per_page: 100,
    });

    const latest = findBotCommentByMarker(allComments, REPRODUCER_COMMENT_MARKER, botLogin);

    if (latest) {
      // Bounded in bytes before it becomes an environment variable: a large body written in a
      // multi-byte script can exceed the kernel's per-environment-entry limit.
      core.setOutput('prior_reproducer_comment', boundEnvPayload(latest.body));
      core.info(`Found prior reproducer comment ${latest.id} for issue #${issueNumber}`);
    } else {
      core.setOutput('prior_reproducer_comment', '');
      core.info(`No prior reproducer comment found for issue #${issueNumber}`);
    }
  } catch (err) {
    core.setOutput('prior_reproducer_comment', '');
    core.warning(`Could not fetch prior reproducer comment for issue #${issueNumber}: ${err.message}`);
  }
};
