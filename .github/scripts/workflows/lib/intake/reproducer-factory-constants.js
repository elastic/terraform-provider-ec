/**
 * Reproducer-factory issue intake configuration.
 *
 * Unlike research-factory, reproducer-factory opens a PR carrying a repro test, so the duplicate-PR
 * gate stays wired: there is no `DISABLE_DUPLICATE_GATE` here, and `ISSUE_BRANCH_PREFIX` is live. The
 * prefix is the single source of truth for the repro branch name.
 * `.buildkite/acceptance-test-selection.sh` narrows the acceptance run by matching
 * `^reproducer-factory/issue-([1-9][0-9]*)$`, so the two must stay in lockstep;
 * `buildkite-acceptance-test-selection.test.mjs` builds its branch from this constant and fails if
 * they drift.
 */
'use strict';

const ISSUE_BRANCH_PREFIX = 'reproducer-factory/issue-';
const FACTORY_LABEL = 'reproducer-factory';
const DUPLICATE_LINKAGE_MODE = 'related-literal';
const ISSUE_OPENED_NOT_ELIGIBLE_REASON =
  'Issue opened event does not qualify because the issue was created without the reproducer-factory label.';
// Sticky-comment marker. The pre-activation lookup and the safe-output handler both match on it,
// so a one-character drift here makes every run post a new comment instead of editing the old one.
const REPRODUCER_COMMENT_MARKER = '<!-- gha-reproducer-factory -->';
// This factory is label-triggered only (no slash_command), so notify-duplicate-blocked.js must not
// tell a maintainer to retry with a `/reproducer-factory` comment.
const HAS_SLASH_COMMAND = false;

if (typeof module !== 'undefined') {
  module.exports = {
    ISSUE_BRANCH_PREFIX,
    FACTORY_LABEL,
    DUPLICATE_LINKAGE_MODE,
    ISSUE_OPENED_NOT_ELIGIBLE_REASON,
    REPRODUCER_COMMENT_MARKER,
    HAS_SLASH_COMMAND,
  };
}
