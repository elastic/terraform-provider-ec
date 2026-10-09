/**
 * Computes issue slot availability for a labeled issue bucket.
 * @param {{ label: string, issueCap: number | string, openIssueCount: number }} params
 * @returns {{ open_issues: number, issue_slots_available: number, gate_reason: string }}
 */
function computeIssueSlots({ label, issueCap, openIssueCount }) {
  if (typeof label !== 'string') {
    throw new Error('issue slot label must be a non-empty string');
  }
  const normalizedLabel = label.trim();
  if (!normalizedLabel) {
    throw new Error('issue slot label must be a non-empty string');
  }

  // Only number | non-blank string. Number(true) === 1 and Number('  ') === 0
  // would otherwise look like valid caps.
  if (typeof issueCap === 'string') {
    if (issueCap.trim() === '') {
      throw new Error(`issue cap must be a non-negative integer, got: ${issueCap}`);
    }
  } else if (typeof issueCap !== 'number') {
    throw new Error(`issue cap must be a non-negative integer, got: ${issueCap}`);
  }
  const cap = Number(issueCap);
  if (!Number.isInteger(cap) || cap < 0) {
    throw new Error(`issue cap must be a non-negative integer, got: ${issueCap}`);
  }

  if (typeof openIssueCount !== 'number' || !Number.isInteger(openIssueCount) || openIssueCount < 0) {
    throw new Error(`open issue count must be a non-negative integer, got: ${openIssueCount}`);
  }

  const slotsAvailable = Math.max(0, cap - openIssueCount);

  let gateReason;
  if (slotsAvailable === 0) {
    gateReason = `Issue cap reached: ${openIssueCount} open ${normalizedLabel} issue(s), cap is ${cap}. Agent job will be skipped.`;
  } else {
    gateReason = `${slotsAvailable} slot(s) available: ${openIssueCount} open ${normalizedLabel} issue(s), cap is ${cap}.`;
  }

  return {
    open_issues: openIssueCount,
    issue_slots_available: slotsAvailable,
    gate_reason: gateReason,
  };
}

if (typeof module !== 'undefined') {
  module.exports = { computeIssueSlots };
}
