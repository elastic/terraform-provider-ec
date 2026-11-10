/**
 * Explains on the issue why the factory did not open a pull request.
 *
 * The duplicate gate fails closed on ANY open pull request on the factory branch, so the
 * blocker may be a pull request a human opened. That case gets its own wording: "just close
 * it" wrongly assumes someone else's pull request is disposable.
 */
const { getFactoryName, getFactoryConstants } = require('./_factory-context.js');

/**
 * @param {string} url
 * @returns {string}
 */
function extractPrNumber(url) {
  return url.match(/\/(\d+)(?:\/|$)/)?.[1] ?? url;
}

module.exports = async function ({ github, context, core }) {

  const factoryName = getFactoryName();
  const { ISSUE_BRANCH_PREFIX: branchPrefix, DUPLICATE_LINKAGE_MODE: linkageMode } = getFactoryConstants();
  const { owner, repo } = context.repo;

  const duplicatePrUrl = process.env.DUPLICATE_PR_URL || '';
  const contractMatch = process.env.DUPLICATE_PR_CONTRACT_MATCH || '';
  const issueNumberRaw = process.env.ISSUE_NUMBER || '';

  if (!duplicatePrUrl || !issueNumberRaw) {
    core.info(`DUPLICATE_PR_URL or ISSUE_NUMBER is empty; skipping the ${factoryName} duplicate-blocked notification.`);
    return;
  }

  const issueNumber = parseInt(issueNumberRaw, 10);
  const prNumber = extractPrNumber(duplicatePrUrl);
  const branch = `${branchPrefix}${issueNumber}`;
  const linkage = linkageMode === 'closes-literal' ? `Closes #${issueNumber}` : `Related to #${issueNumber}`;

  const body = contractMatch === 'true'
    ? `⚠️ **${factoryName} skipped** — PR #${prNumber} is already open for this issue. `
      + `Close it, then re-apply the \`${factoryName}\` label (or comment \`/${factoryName}\`) to retry.`
    : `⚠️ **${factoryName} skipped** — PR #${prNumber} is already open on branch \`${branch}\`, but it is `
      + `missing the \`${factoryName}\` label or the \`${linkage}\` line. Close it **or move it to another `
      + `branch**, then re-apply the \`${factoryName}\` label (or comment \`/${factoryName}\`) to retry.`;

  // This step runs on the already-blocked path, where the only work left is bookkeeping. Letting
  // a rate-limited createComment fail the job would skip the trigger-label removal that follows,
  // leaving the label on the issue and making the documented "re-apply the label" retry a no-op.
  try {
    await github.rest.issues.createComment({
      owner,
      repo,
      issue_number: issueNumber,
      body,
    });
    core.info(`Posted a duplicate-blocked comment on issue #${issueNumber} referencing ${duplicatePrUrl}`);
  } catch (err) {
    const message = `Could not post the duplicate-blocked comment on issue #${issueNumber}: ${err.message}`;
    if (typeof core.warning === 'function') {
      core.warning(message);
    } else {
      core.info(message);
    }
  }
};
