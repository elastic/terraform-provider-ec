/**
 * Applies the completion phase label to an issue, removing any other phase-* label.
 *
 * Runs in a safe-output job, after the factory's comment has been published, so the issue number
 * arrives raw from `github.event.issue.number || inputs.issue_number` (a custom safe-output job
 * cannot depend on `pre_activation`, where the dispatch input was validated).
 *
 * Fails the job on anything short of full application. `setPhaseLabel` reports
 * `phase_label_set: true` when the label was added but a stale `phase-*` could not be listed or
 * removed, which is half the acceptance criterion; only the full-success reason is accepted here.
 * Setting the label is idempotent, so a red job after a published comment re-runs cleanly.
 */
'use strict';

const { setPhaseLabel } = require('../phase-label.js');
const { positiveIssueNumber } = require('../issue-number.js');

module.exports = async function ({ github, context, core }) {

  const issueNumber = positiveIssueNumber(process.env.INPUT_ISSUE_NUMBER);
  const phaseLabelName = process.env.PHASE_LABEL_NAME || '';

  if (!issueNumber) {
    core.setOutput('phase_label_set', 'false');
    core.setFailed(`Cannot set phase label: invalid issue number '${process.env.INPUT_ISSUE_NUMBER || ''}'.`);
    return;
  }

  let result;
  try {
    result = await setPhaseLabel({ github, context, issueNumber, phaseLabelName, core });
  } catch (err) {
    core.setOutput('phase_label_set', 'false');
    core.setFailed(`Failed to set phase label ${phaseLabelName} on #${issueNumber}: ${err.message}`);
    return;
  }

  // Positive allowlist: setPhaseLabel composes its full-success reason as
  // `Set phase label <name>. ...`. Every partial outcome ("Added label X but failed to ...")
  // and every refusal uses a different prefix, so anything else fails closed.
  const fullyApplied = result.phase_label_set === true && result.reason.startsWith(`Set phase label ${result.phase_label_name}.`);

  core.setOutput('phase_label_set', fullyApplied ? 'true' : 'false');
  core.setOutput('phase_label_name', result.phase_label_name);
  core.setOutput('stale_labels_removed', (result.stale_labels_removed || []).join(','));

  if (!fullyApplied) {
    core.setFailed(`Phase label not fully applied to #${issueNumber}: ${result.reason}`);
    return;
  }

  core.info(`#${issueNumber}: ${result.reason}`);
};
