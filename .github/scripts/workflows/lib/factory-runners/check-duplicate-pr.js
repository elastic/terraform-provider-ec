const { getFactoryConstants, getFactoryModule } = require('./_factory-context.js');

module.exports = async function ({ github, context, core }) {

  const { checkDuplicatePR, issueBranchName } = getFactoryModule();
  const { FACTORY_LABEL: factoryLabel, DUPLICATE_LINKAGE_MODE: linkageMode } = getFactoryConstants();

  const { owner, repo } = context.repo;
  const intakeMode = context.eventName === 'workflow_dispatch' ? 'dispatch' : 'issue-event';
  let issueNumber;

  if (intakeMode === 'dispatch') {
    issueNumber = parseInt(context.payload.inputs?.issue_number, 10) || null;
  } else {
    issueNumber = context.payload.issue?.number || null;
  }

  if (!issueNumber) {
    core.setOutput('duplicate_pr_found', 'false');
    core.setOutput('duplicate_pr_contract_match', 'false');
    core.setOutput('duplicate_pr_url', '');
    core.setOutput('gate_reason', 'No issue number available for duplicate PR check.');
    core.info('Duplicate PR check skipped: issue number is not available.');
    return;
  }

  const expectedBranch = issueBranchName(issueNumber);

  const pulls = await github.paginate(github.rest.pulls.list, {
    owner,
    repo,
    state: 'open',
    head: `${owner}:${expectedBranch}`,
    per_page: 100,
  });

  const pullRequests = pulls.map(pr => ({
    number: pr.number,
    state: pr.state,
    head_branch: pr.head.ref,
    labels: pr.labels.map(l => l.name),
    body: pr.body ?? '',
    html_url: pr.html_url,
  }));

  const result = checkDuplicatePR({ issueNumber, pullRequests });

  let { duplicate_pr_found: duplicatePrFound, duplicate_pr_url: duplicatePrUrl, gate_reason: gateReason } = result;
  // A PR that matches the full contract (branch + label + linkage) is one this factory can claim
  // it opened; a bare branch match may be anybody's. The caller words its comment off this.
  let contractMatch = result.duplicate_pr_found;

  // Fail closed. The contract predicates (label, `Related to #N`) are both removable by ordinary
  // triage, and the workflow runs with recreate-ref: true — so trusting them would let a stripped
  // label force-delete the ref under a live open PR.
  if (!duplicatePrFound) {
    // Re-assert state and branch rather than trusting the API's state/head filter.
    const onBranch = pullRequests.find(
      pr => pr.state === 'open' && pr.head_branch === expectedBranch,
    );
    if (onBranch) {
      const linkage = linkageMode === 'closes-literal' ? `Closes #${issueNumber}` : `Related to #${issueNumber}`;
      duplicatePrFound = true;
      duplicatePrUrl = onBranch.html_url ?? '';
      contractMatch = false;
      gateReason = `Refusing to open a second pull request: PR #${onBranch.number} `
        + `(${onBranch.html_url ?? '(unknown URL)'}) is already open on branch '${expectedBranch}', `
        + `although it does not carry the '${factoryLabel}' label or the literal linkage '${linkage}'.`;
    }
  }

  core.setOutput('duplicate_pr_found', duplicatePrFound ? 'true' : 'false');
  core.setOutput('duplicate_pr_contract_match', contractMatch ? 'true' : 'false');
  core.setOutput('duplicate_pr_url', duplicatePrUrl ?? '');
  core.setOutput('gate_reason', gateReason);

  if (duplicatePrFound) {
    core.info(`Duplicate PR found: ${gateReason}`);
  } else {
    core.info(`No duplicate PR: ${gateReason}`);
  }
};
