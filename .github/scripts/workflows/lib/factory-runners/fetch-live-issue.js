/**
 * Fetches the live issue title and body from the GitHub API (not the webhook payload).
 *
 * Outputs issue_number, issue_title, issue_body, and fetch_error are the raw API values.
 * This module does not sanitize them. Callers pass title and body through the sanitize
 * and write runners before an agent reads them, and do not expand issue_title or
 * issue_body inside a shell `run:` step.
 */
module.exports = async function ({ github, context, core }) {
  const { owner, repo } = context.repo;
  const issueNumber = parseInt(process.env.INPUT_ISSUE_NUMBER, 10);

  if (!issueNumber || issueNumber <= 0) {
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
      core.setOutput('issue_number', String(data.number));
      core.setOutput('issue_title', data.title ?? '');
      core.setOutput('issue_body', data.body ?? '');
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
