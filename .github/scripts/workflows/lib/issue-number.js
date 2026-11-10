/**
 * Issue-number validation shared by the pre-activation runners and the safe-output handlers.
 *
 * Safe-output jobs cannot depend on `pre_activation`, so they read the issue number straight from
 * `github.event.issue.number || inputs.issue_number`. A `workflow_dispatch` input is raw text that
 * GitHub does not validate, and `parseInt('1.5', 10)` is `1` — enough to stamp the wrong issue.
 * Three other factory runners parse an issue number without this guard, none of them a gap:
 * `validate-dispatch-inputs.js` applies the equivalent canonical-form check from
 * `code-factory-dispatch.js` (it is where a dispatch input gets validated in the first place);
 * `fetch-issue-comments.js` only ever receives an already-validated number, from that runner or
 * straight off the webhook payload; and `check-duplicate-pr.js` reads a raw dispatch input but is
 * wired into no workflow here.
 */
'use strict';

// The whole string must be a safe positive integer.
function positiveIssueNumber(raw) {
  if (typeof raw !== 'string' || !/^[1-9][0-9]*$/.test(raw)) return 0;
  const issueNumber = Number(raw);
  if (!Number.isSafeInteger(issueNumber)) return 0;
  return issueNumber;
}

if (typeof module !== 'undefined') {
  module.exports = { positiveIssueNumber };
}
