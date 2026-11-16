/**
 * Shared deterministic helpers for code-factory and change-factory issue intake workflows.
 * Workflow-specific configuration is passed via {@link createFactoryIssueIntake}.
 */

/** GitHub-recognized issue-closing keywords (case-insensitive). See https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/using-keywords-in-issues-and-pull-requests */
const GITHUB_ISSUE_CLOSING_KEYWORDS = '(?:close|closes|closed|fix|fixes|fixed|resolve|resolves|resolved)';

/**
 * GitHub closing-keyword reference: `#` is immediately followed by the issue digits (no whitespace),
 * per GitHub keyword syntax. Case-insensitive keywords; `(?![0-9])` avoids matching `#42` inside `#420`.
 *
 * @param {number} issueNumber
 * @returns {RegExp}
 */
function issueClosingReferencePattern(issueNumber) {
  return new RegExp(
    `\\b${GITHUB_ISSUE_CLOSING_KEYWORDS}\\s*#${issueNumber}(?![0-9])`,
    'i',
  );
}

/**
 * @param {{ eventName: string, eventAction: string, labelName: string, issueLabels: string[] | null | undefined, factoryLabel: string, issueOpenedNotEligibleReason: string, commentBody?: string, isPullRequest?: boolean }} params
 * @returns {{ event_eligible: boolean, event_eligible_reason: string }}
 */
function factoryQualifyTriggerEvent({
  eventName,
  eventAction,
  labelName,
  issueLabels,
  factoryLabel,
  issueOpenedNotEligibleReason,
  commentBody,
  isPullRequest,
}) {
  if (eventName === 'issue_comment') {
    // Pre-activation steps run before gh-aw's own activation gate, so the body and the
    // issue-vs-pull-request discriminator have to be re-checked here: without them any
    // comment by any reader would qualify and drive the issues: write steps.
    // Truthy, not `=== true`: a caller that forwards the raw `payload.issue.pull_request` object
    // rather than a boolean must still be rejected. Omitting the param stays undefined → falsy.
    if (isPullRequest) {
      return {
        event_eligible: false,
        event_eligible_reason: `Issue comment event does not qualify because the comment is on a pull request, not an issue.`,
      };
    }

    const command = `/${factoryLabel}`;
    const body = typeof commentBody === 'string' ? commentBody : '';
    // Mirror gh-aw's activation delimiters: raw startsWith, never trimStart(). Trimming here
    // would qualify a comment gh-aw then refuses, and pre-activation would write labels for a
    // run the agent never gets. Case-sensitivity deliberately differs: GitHub Actions
    // expressions compare case-insensitively, so `/Change-Factory` activates gh-aw but is
    // rejected here — stricter than the gate it guards, which is the safe direction.
    const matches = body === command
      || body.startsWith(`${command} `)
      || body.startsWith(`${command}\n`)
      || body.startsWith(`${command}\r`);

    if (!matches) {
      return {
        event_eligible: false,
        event_eligible_reason: `Issue comment event does not qualify because the comment body does not start with '${command}'.`,
      };
    }

    return {
      event_eligible: true,
      event_eligible_reason: `Issue comment event qualifies because the comment body starts with '${command}'.`,
    };
  }

  if (eventName !== 'issues') {
    return {
      event_eligible: false,
      event_eligible_reason: `Unsupported event '${eventName || '(empty)'}'; expected 'issues'.`,
    };
  }

  if (eventAction === 'labeled') {
    if (labelName === factoryLabel) {
      return {
        event_eligible: true,
        event_eligible_reason: `Issue labeled event qualifies because the applied label is ${factoryLabel}.`,
      };
    }

    return {
      event_eligible: false,
      event_eligible_reason: `Issue labeled event does not qualify because the applied label is '${labelName || '(empty)'}', not '${factoryLabel}'.`,
    };
  }

  if (eventAction === 'opened') {
    if (Array.isArray(issueLabels) && issueLabels.includes(factoryLabel)) {
      return {
        event_eligible: true,
        event_eligible_reason: `Issue opened event qualifies because the issue already has the ${factoryLabel} label.`,
      };
    }

    return {
      event_eligible: false,
      event_eligible_reason: issueOpenedNotEligibleReason,
    };
  }

  return {
    event_eligible: false,
    event_eligible_reason: `Issue event action '${eventAction || '(empty)'}' is not eligible; expected 'opened' or 'labeled'.`,
  };
}

/**
 * @param {{ issueNumber: number, pullRequests: Array<{ number: number, state: string, head_branch: string, labels: string[], body: string, html_url: string }>, branchPrefix: string, prLabel: string, duplicateLinkageMode: 'closes-literal' | 'related-literal' | 'github-keywords' }} params
 * @returns {{ duplicate_pr_found: boolean, duplicate_pr_url: string | null, gate_reason: string }}
 */
function factoryCheckDuplicatePR({
  issueNumber,
  pullRequests,
  branchPrefix,
  prLabel,
  duplicateLinkageMode,
}) {
  const expectedBranch = `${branchPrefix}${issueNumber}`;
  const expectedClosesExample = `Closes #${issueNumber}`;
  const expectedRelatedExample = `Related to #${issueNumber}`;

  let bodyPattern;
  if (duplicateLinkageMode === 'closes-literal') {
    bodyPattern = new RegExp(`\\bCloses #${issueNumber}(?![0-9])`);
  } else if (duplicateLinkageMode === 'related-literal') {
    bodyPattern = new RegExp(`\\bRelated to #${issueNumber}(?![0-9])`);
  } else {
    bodyPattern = issueClosingReferencePattern(issueNumber);
  }

  const duplicate = (pullRequests || []).find(pr => (
    pr.state === 'open' &&
    Array.isArray(pr.labels) && pr.labels.includes(prLabel) &&
    pr.head_branch === expectedBranch &&
    bodyPattern.test(String(pr.body || ''))
  ));

  if (duplicate) {
    const url = duplicate.html_url ?? null;
    let linkagePhrase;
    if (duplicateLinkageMode === 'closes-literal') {
      linkagePhrase = `canonical linkage '${expectedClosesExample}'`;
    } else if (duplicateLinkageMode === 'related-literal') {
      linkagePhrase = `literal linkage \`${expectedRelatedExample}\``;
    } else {
      linkagePhrase = `issue-closing reference such as '${expectedClosesExample}'`;
    }

    return {
      duplicate_pr_found: true,
      duplicate_pr_url: url,
      gate_reason: `Found existing linked ${prLabel} PR #${duplicate.number} (${url ?? '(unknown URL)'}) for issue #${issueNumber} on branch '${expectedBranch}' with ${linkagePhrase}.`,
    };
  }

  // The no-match reason is what the agent reads as gate_reason on the all-gates-passed path, so
  // it must not advertise the label and linkage predicates: check-duplicate-pr.js fails closed on
  // any open PR on the branch, whether or not it carries them.
  return {
    duplicate_pr_found: false,
    duplicate_pr_url: null,
    gate_reason: `No open pull request found on branch '${expectedBranch}' for issue #${issueNumber}.`,
  };
}

/**
 * @param {{ eventEligible: boolean, eventEligibleReason: string, actorTrusted: boolean | null, actorTrustedReason: string | null, duplicatePrFound: boolean | null, duplicatePrUrl: string | null, duplicateCheckGateReason: string | null }} params
 * @param {string} factoryLabel
 * @returns {{ gate_reason: string }}
 */
function factoryComputeGateReason({
  eventEligible,
  eventEligibleReason,
  actorTrusted,
  actorTrustedReason,
  duplicatePrFound,
  duplicatePrUrl,
  duplicateCheckGateReason,
}, factoryLabel) {
  if (!eventEligible) {
    return { gate_reason: eventEligibleReason };
  }

  if (actorTrusted === false) {
    return { gate_reason: actorTrustedReason || 'Trigger actor is not trusted.' };
  }

  if (actorTrusted == null) {
    return { gate_reason: 'Actor trust could not be determined; the trust check step did not produce an output.' };
  }

  if (duplicatePrFound === true) {
    return {
      gate_reason: duplicateCheckGateReason || `Found existing linked ${factoryLabel} PR: ${duplicatePrUrl || '(unknown URL)'}.`,
    };
  }

  if (duplicatePrFound == null) {
    return { gate_reason: 'Duplicate PR check did not complete; the check step did not produce an output.' };
  }

  return {
    gate_reason: duplicateCheckGateReason || `All deterministic gates passed: event eligible, actor trusted, and no linked ${factoryLabel} PR found.`,
  };
}

/**
 * @param {string | undefined} raw
 * @returns {boolean | null}
 */
function factoryParseOptionalTriStateFromEnv(raw) {
  if (raw === 'true') {
    return true;
  }
  if (raw === 'false') {
    return false;
  }
  return null;
}

/**
 * @param {Record<string, string | undefined>} env
 */
function factoryParseFinalizeGateEnv(env) {
  const e = env || {};
  return {
    eventEligible: e.EVENT_ELIGIBLE === 'true',
    eventEligibleReason: e.EVENT_ELIGIBLE_REASON ?? '',
    actorTrusted: factoryParseOptionalTriStateFromEnv(e.ACTOR_TRUSTED),
    actorTrustedReason: e.ACTOR_TRUSTED_REASON ?? null,
    duplicatePrFound: factoryParseOptionalTriStateFromEnv(e.DUPLICATE_PR_FOUND),
    duplicatePrUrl: e.DUPLICATE_PR_URL && e.DUPLICATE_PR_URL !== '' ? e.DUPLICATE_PR_URL : null,
    duplicateCheckGateReason: e.DUPLICATE_GATE_REASON ?? null,
  };
}

/**
 * @param {{
 *   branchPrefix: string,
 *   factoryLabel: string,
 *   issueOpenedNotEligibleReason: string,
 *   duplicateLinkageMode: 'closes-literal' | 'related-literal' | 'github-keywords',
 * }} config
 */
function createFactoryIssueIntake(config) {
  const {
    branchPrefix,
    factoryLabel,
    issueOpenedNotEligibleReason,
    duplicateLinkageMode,
  } = config;

  function issueBranchName(issueNumber) {
    return `${branchPrefix}${issueNumber}`;
  }

  /**
   * @param {{ eventName: string, eventAction: string, labelName: string, issueLabels?: string[] | null | undefined }} params
   */
  function qualifyTriggerEvent(params) {
    return factoryQualifyTriggerEvent({
      ...params,
      factoryLabel,
      issueOpenedNotEligibleReason,
    });
  }

  function checkDuplicatePR(params) {
    return factoryCheckDuplicatePR({
      ...params,
      branchPrefix,
      prLabel: factoryLabel,
      duplicateLinkageMode,
    });
  }

  function computeGateReason(params) {
    return factoryComputeGateReason(params, factoryLabel);
  }

  return {
    issueBranchName,
    qualifyTriggerEvent,
    checkDuplicatePR,
    computeGateReason,
  };
}

/**
 * @param {{
 *   branchPrefix: string,
 *   factoryLabel: string,
 *   issueOpenedNotEligibleReason: string,
 *   duplicateLinkageMode: 'closes-literal' | 'related-literal' | 'github-keywords',
 *   issueBranchNameAliases?: string[],
 * }} config
 */
function createFactoryIssueModule(config) {
  const intake = createFactoryIssueIntake(config);
  const issueBranchNameAliases = config.issueBranchNameAliases || [];
  const factoryIssueModule = {
    ...intake,
    parseOptionalTriStateFromEnv: factoryParseOptionalTriStateFromEnv,
    parseFinalizeGateEnv: factoryParseFinalizeGateEnv,
  };

  for (const alias of issueBranchNameAliases) {
    factoryIssueModule[alias] = intake.issueBranchName;
  }

  return factoryIssueModule;
}

if (typeof module !== 'undefined') {
  module.exports = {
    issueClosingReferencePattern,
    factoryQualifyTriggerEvent,
    factoryCheckDuplicatePR,
    factoryComputeGateReason,
    factoryParseOptionalTriStateFromEnv,
    factoryParseFinalizeGateEnv,
    createFactoryIssueIntake,
    createFactoryIssueModule,
  };
}
