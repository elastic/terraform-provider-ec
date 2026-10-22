/**
 * Stamps the issue with a phase-* label and clears any other phase-* label.
 *
 * Deliberately not FACTORY_NAME-keyed: every factory sets a different phase, so the label is
 * passed in through PHASE_LABEL_NAME and this runner is reused verbatim.
 *
 * Soft failures are reported through the outputs, not raised: a label that could not be set is
 * not a reason to fail the run.
 */
const { setPhaseLabel } = require('../phase-label.js');

// The whole string must be a safe positive integer, so '1.5' or '007' cannot select an issue.
function positiveIssueNumber(raw) {
  if (typeof raw !== 'string' || !/^[1-9][0-9]*$/.test(raw)) return 0;
  const issueNumber = Number(raw);
  if (!Number.isSafeInteger(issueNumber)) return 0;
  return issueNumber;
}

module.exports = async function ({ github, context, core }) {

  const rawIssueNumber = process.env.INPUT_ISSUE_NUMBER || '';
  const issueNumber = positiveIssueNumber(rawIssueNumber);
  const phaseLabelName = process.env.PHASE_LABEL_NAME || '';

  const result = issueNumber
    ? await setPhaseLabel({ github, context, issueNumber, phaseLabelName, core })
    : {
      phase_label_set: false,
      phase_label_name: phaseLabelName,
      stale_labels_removed: [],
      reason: `No usable issue number: expected a positive integer, got '${rawIssueNumber}'.`,
    };

  core.setOutput('phase_label_set', result.phase_label_set ? 'true' : 'false');
  core.setOutput('phase_label_name', result.phase_label_name);
  core.setOutput('stale_labels_removed', JSON.stringify(result.stale_labels_removed));
  core.setOutput('reason', result.reason);
  core.info(result.reason);
};
