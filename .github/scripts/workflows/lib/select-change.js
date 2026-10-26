const CHANGE_PATTERN = /^openspec\/changes\/([^/]+)\/.+$/;
const ARCHIVE_PATTERN = /^openspec\/changes\/archive\//;

const ALLOWED_STATUSES = new Set(['added', 'modified']);

function ineligible(selection_reason) {
  return {
    selection_status: 'ineligible',
    selection_reason,
    selected_change: '',
    review_disposition: '',
    disposition_reason: '',
  };
}

/**
 * Paths on a PR file that can place it under an active (non-archive) change.
 * Includes previous_filename so a rename out of openspec/changes/<id>/ is not missed.
 * @param {{ filename: string, previous_filename?: string }} file
 * @returns {string[]}
 */
function activeChangePaths(file) {
  const paths = [file.filename];
  if (typeof file.previous_filename === 'string' && file.previous_filename) {
    paths.push(file.previous_filename);
  }
  return paths.filter(
    (path) => CHANGE_PATTERN.test(path) && !ARCHIVE_PATTERN.test(path),
  );
}

function selectChangeFromFiles(files) {
  const relevantFiles = files.filter((file) => activeChangePaths(file).length > 0);

  if (relevantFiles.length === 0) {
    return ineligible('No files under openspec/changes/ (non-archive) found in this PR');
  }

  const unsupported = relevantFiles.filter(file => !ALLOWED_STATUSES.has(file.status));
  if (unsupported.length > 0) {
    return ineligible(
      `Unsupported file status under openspec/changes/: ${unsupported
        .map(file => `${file.filename} (${file.status})`)
        .join(', ')}`
    );
  }

  const changeIds = new Set();
  for (const file of relevantFiles) {
    for (const path of activeChangePaths(file)) {
      changeIds.add(path.match(CHANGE_PATTERN)[1]);
    }
  }

  if (changeIds.size > 1) {
    return ineligible(`Multiple active change ids: ${Array.from(changeIds).sort().join(', ')}`);
  }

  const selectedChange = Array.from(changeIds)[0];
  const hasAdded = relevantFiles.some(file => file.status === 'added');
  const reviewDisposition = hasAdded ? 'comment-only' : 'approval-eligible';

  const dispositionReason = hasAdded
    ? 'The selected change includes one or more added files (net-new spec change material). APPROVE is not permitted; submit COMMENT only, even if verification passes with no blocking issues.'
    : 'Every file under the selected change is a modification. APPROVE is permitted when verification finds zero CRITICAL issues and zero unassociated files.';

  return {
    selection_status: 'eligible',
    selection_reason: `Selected change: ${selectedChange}`,
    selected_change: selectedChange,
    review_disposition: reviewDisposition,
    disposition_reason: dispositionReason,
  };
}

function selectChangeForPullRequest({ prNumber, files = [] }) {
  if (!prNumber) {
    return ineligible('No pull request number in event payload');
  }

  return selectChangeFromFiles(files);
}

if (typeof module !== 'undefined') {
  module.exports = {
    ARCHIVE_PATTERN,
    CHANGE_PATTERN,
    selectChangeForPullRequest,
    selectChangeFromFiles,
  };
}
