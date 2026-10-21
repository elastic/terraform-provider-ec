/**
 * Pre-activation lookup of the research factory's own sticky comment.
 *
 * The agent reads the result as draft input (`prior_research_comment.md`), and the safe-output
 * handler performs the same lookup again when it publishes. Both go through
 * `findResearchComment` so the bot+marker filter has one definition.
 *
 * Three rules this encodes:
 * - paginate without `sort`/`direction`: the helper's "oldest first, take the last" contract rests
 *   on the REST API defaults (created ascending);
 * - REST only: `findResearchComment` reads `c.author ?? c.user?.login`, and a GraphQL comment's
 *   `author` is an object, which would short-circuit the `??` and match nothing — producing a
 *   duplicate comment rather than an edit;
 * - resolve the bot login with `||`, not a default parameter: a set-but-empty `GH_AW_BOT_LOGIN`
 *   propagates `''`, which matches no author.
 *
 * A lookup failure warns rather than fails: losing the prior comment costs the agent its draft
 * input, but the run can still produce research.
 */
'use strict';

const { findResearchComment } = require('../lib/sanitize-context.js');
const { positiveIssueNumber } = require('../lib/issue-number.js');
const { RESEARCH_COMMENT_MARKER } = require('../lib/intake/research-factory-constants.js');

module.exports = async function ({ github, context, core }) {

  const { owner, repo } = context.repo;
  const issueNumber = positiveIssueNumber(process.env.INPUT_ISSUE_NUMBER);
  const botLogin = process.env.GH_AW_BOT_LOGIN || 'github-actions[bot]';

  if (!issueNumber) {
    core.setOutput('prior_research_comment', '');
    core.info('No prior research comment lookup: invalid issue number.');
    return;
  }

  try {
    const allComments = await github.paginate(github.rest.issues.listComments, {
      owner,
      repo,
      issue_number: issueNumber,
      per_page: 100,
    });

    const latest = findResearchComment(allComments, RESEARCH_COMMENT_MARKER, botLogin);

    if (latest) {
      core.setOutput('prior_research_comment', latest.body);
      core.info(`Found prior research comment ${latest.id} for issue #${issueNumber}`);
    } else {
      core.setOutput('prior_research_comment', '');
      core.info(`No prior research comment found for issue #${issueNumber}`);
    }
  } catch (err) {
    core.setOutput('prior_research_comment', '');
    core.warning(`Could not fetch prior research comment for issue #${issueNumber}: ${err.message}`);
  }
};
