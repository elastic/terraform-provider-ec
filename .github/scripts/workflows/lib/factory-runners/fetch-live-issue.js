/**
 * Fetches the live issue title and body from the GitHub API (not the webhook payload).
 *
 * Outputs issue_number, issue_title, issue_body, and fetch_error are the raw API values.
 * This module does not sanitize them. Callers pass the body through the sanitize and
 * write runners before an agent reads it. The title is not written to disk here.
 * Do not expand issue_title or issue_body inside a shell `run:` step.
 *
 * Pull requests are refused outright: every factory sharing this runner is issue-only today. A
 * future PR-aware factory needs an explicit opt-in here rather than a silent widening.
 */

const { positiveIssueNumber } = require('../issue-number.js');
const { boundEnvPayload } = require('../env-payload.js');

/**
 * Collapses newlines so a title is safe to export with a bare `echo` into `GITHUB_OUTPUT`.
 *
 * @param {unknown} title
 * @returns {string}
 */
function flattenTitle(title) {
  return String(title ?? '').replace(/[\r\n]+/g, ' ');
}

module.exports = async function ({ github, context, core }) {
  const { owner, repo } = context.repo;
  const issueNumber = positiveIssueNumber(process.env.INPUT_ISSUE_NUMBER);

  if (!issueNumber) {
    core.setOutput('issue_number', '');
    core.setOutput('issue_title', '');
    core.setOutput('issue_body', '');
    core.setOutput('fetch_error', 'Invalid issue number in dispatch inputs.');
    core.setFailed('Cannot fetch live issue: invalid issue number.');
  } else {
    try {
      const { data } = await github.rest.issues.get({
        owner,
        repo,
        issue_number: issueNumber,
      });
      if (data.pull_request) {
        // A dispatch can name any number, and the Issues API returns pull requests too — as do the
        // comment and label endpoints downstream. Without this check an issue-only factory will
        // research a pull request and stamp its phase label on it.
        core.setOutput('issue_number', '');
        core.setOutput('issue_title', '');
        core.setOutput('issue_body', '');
        core.setOutput('fetch_error', `#${data.number} is a pull request, not an issue.`);
        core.setFailed(`Refusing to run: #${data.number} is a pull request, not an issue.`);
        return;
      }
      core.setOutput('issue_number', String(data.number));
      // Flattened: `normalize_context` exports the title with a bare `echo` into GITHUB_OUTPUT,
      // where a newline would start a second key=value line. The title is written *after*
      // issue_number and a later assignment wins, so an injected `issue_number=` line could
      // redirect the whole run at another issue; an injected heredoc opener would instead break
      // the output file, after the trigger label has already been consumed.
      core.setOutput('issue_title', flattenTitle(data.title));
      core.setOutput('issue_body', boundEnvPayload(data.body ?? ''));
      core.setOutput('fetch_error', '');
      core.info(`Fetched live issue #${data.number}: ${data.title}`);
    } catch (err) {
      core.setOutput('issue_number', '');
      core.setOutput('issue_title', '');
      core.setOutput('issue_body', '');
      core.setOutput('fetch_error', err.message);
      core.setFailed(`Failed to fetch issue #${issueNumber}: ${err.message}`);
    }
  }
};

module.exports.flattenTitle = flattenTitle;
