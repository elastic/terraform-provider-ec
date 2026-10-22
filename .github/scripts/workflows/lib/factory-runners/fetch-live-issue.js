/**
 * Fetches the live issue title and body from the GitHub API (not the webhook payload).
 *
 * Outputs issue_number, issue_title, issue_body, and fetch_error are the raw API values.
 * This module does not sanitize them. Callers pass the body through the sanitize and
 * write runners before an agent reads it. The title is not written to disk here.
 * Do not expand issue_title or issue_body inside a shell `run:` step.
 *
 * The body is clamped to a UTF-8 byte budget before it is emitted: downstream steps receive it
 * through an environment variable, and a single env-var string above Linux's MAX_ARG_STRLEN
 * (~128 KB) makes the next process launch fail with E2BIG. GitHub issue bodies cap at 65,536
 * characters, which is up to ~256 KB of UTF-8 for CJK/emoji-heavy text.
 */

const { clampUtf8Bytes } = require('../utf8.js');

// Max UTF-8 bytes for the issue body passed via env to normalize/sanitize steps.
const ISSUE_BODY_MAX_BYTES = 96_000;

// The whole string must be a safe positive integer.
function positiveIssueNumber(raw) {
  if (typeof raw !== 'string' || !/^[1-9][0-9]*$/.test(raw)) return 0;
  const issueNumber = Number(raw);
  if (!Number.isSafeInteger(issueNumber)) return 0;
  return issueNumber;
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
      const { text: body, truncated } = clampUtf8Bytes(data.body ?? '', ISSUE_BODY_MAX_BYTES);
      core.setOutput('issue_number', String(data.number));
      core.setOutput('issue_title', data.title ?? '');
      core.setOutput('issue_body', body);
      core.setOutput('fetch_error', '');
      core.info(
        `Fetched live issue #${data.number}: ${data.title}` +
          (truncated ? ` (body clamped to ${ISSUE_BODY_MAX_BYTES} bytes)` : ''),
      );
    } catch (err) {
      core.setOutput('issue_number', '');
      core.setOutput('issue_title', '');
      core.setOutput('issue_body', '');
      core.setOutput('fetch_error', err.message);
      core.setFailed(`Failed to fetch issue #${issueNumber}: ${err.message}`);
    }
  }
};
