import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const {
  issueClosingReferencePattern,
  factoryQualifyTriggerEvent,
  factoryParseOptionalTriStateFromEnv,
  factoryParseFinalizeGateEnv,
  factoryCheckDuplicatePR,
  factoryComputeGateReason,
  createFactoryIssueIntake,
  createFactoryIssueModule,
} = require('./factory-issue-shared.js');
test('factoryParseOptionalTriStateFromEnv treats missing and empty as null', () => {
  assert.equal(factoryParseOptionalTriStateFromEnv(undefined), null);
  assert.equal(factoryParseOptionalTriStateFromEnv(''), null);
});

test('factoryParseOptionalTriStateFromEnv parses true only for exact true string', () => {
  assert.equal(factoryParseOptionalTriStateFromEnv('true'), true);
  assert.equal(factoryParseOptionalTriStateFromEnv('false'), false);
  assert.equal(factoryParseOptionalTriStateFromEnv('TRUE'), false);
});

test('factoryParseFinalizeGateEnv matches finalize_gate env semantics', () => {
  assert.deepEqual(
    factoryParseFinalizeGateEnv({}),
    {
      eventEligible: false,
      eventEligibleReason: '',
      actorTrusted: null,
      actorTrustedReason: null,
      duplicatePrFound: null,
      duplicatePrUrl: null,
      duplicateCheckGateReason: null,
    },
  );

  assert.deepEqual(
    factoryParseFinalizeGateEnv({
      EVENT_ELIGIBLE: 'true',
      EVENT_ELIGIBLE_REASON: 'ok',
      ACTOR_TRUSTED: 'true',
      ACTOR_TRUSTED_REASON: '',
      DUPLICATE_PR_FOUND: 'false',
      DUPLICATE_PR_URL: '',
      DUPLICATE_GATE_REASON: 'x',
    }),
    {
      eventEligible: true,
      eventEligibleReason: 'ok',
      actorTrusted: true,
      actorTrustedReason: '',
      duplicatePrFound: false,
      duplicatePrUrl: null,
      duplicateCheckGateReason: 'x',
    },
  );
});

test('factoryParseFinalizeGateEnv feeds factoryComputeGateReason for an all-pass path', () => {
  const parsed = factoryParseFinalizeGateEnv({
    EVENT_ELIGIBLE: 'true',
    EVENT_ELIGIBLE_REASON: 'eligible',
    ACTOR_TRUSTED: 'true',
    ACTOR_TRUSTED_REASON: 'trusted',
    DUPLICATE_PR_FOUND: 'false',
    DUPLICATE_PR_URL: 'https://example.com/pr/1',
    DUPLICATE_GATE_REASON: null,
  });
  const result = factoryComputeGateReason(parsed, 'code-factory');

  assert.match(result.gate_reason, /All deterministic gates passed/);
});

test('factoryComputeGateReason uses generic untrusted text when actorTrusted is false with falsy reason', () => {
  const result = factoryComputeGateReason({
    eventEligible: true,
    eventEligibleReason: 'Event is eligible.',
    actorTrusted: false,
    actorTrustedReason: '',
    duplicatePrFound: false,
    duplicatePrUrl: null,
    duplicateCheckGateReason: null,
  }, 'change-factory');

  assert.equal(result.gate_reason, 'Trigger actor is not trusted.');
});

test('factoryComputeGateReason falls back to unknown URL when duplicate found without URL or override', () => {
  const result = factoryComputeGateReason({
    eventEligible: true,
    eventEligibleReason: 'Event is eligible.',
    actorTrusted: true,
    actorTrustedReason: 'trusted',
    duplicatePrFound: true,
    duplicatePrUrl: null,
    duplicateCheckGateReason: null,
  }, 'change-factory');

  assert.match(result.gate_reason, /Found existing linked change-factory PR: \(unknown URL\)\./);
});

test('factoryQualifyTriggerEvent rejects non-issues events', () => {
  const result = factoryQualifyTriggerEvent({
    eventName: 'pull_request',
    eventAction: 'opened',
    labelName: '',
    issueLabels: [],
    factoryLabel: 'demo-factory',
    issueOpenedNotEligibleReason: 'n/a',
  });
  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /expected 'issues'/);
});

test('factoryQualifyTriggerEvent accepts issue_comment event', () => {
  const result = factoryQualifyTriggerEvent({
    eventName: 'issue_comment',
    eventAction: '',
    labelName: '',
    issueLabels: [],
    factoryLabel: 'change-factory',
    issueOpenedNotEligibleReason: 'n/a',
  });
  assert.equal(result.event_eligible, true);
  assert.match(result.event_eligible_reason, /issue_comment/);
});

test('issueClosingReferencePattern matches GitHub closing keywords but not longer issue numbers', () => {
  const p = issueClosingReferencePattern(42);
  assert.equal(p.test('See fixes #42\n'), true);
  assert.equal(p.test('fixes #420'), false);
});

test('issueClosingReferencePattern does not match whitespace between # and the issue number', () => {
  const p = issueClosingReferencePattern(123);
  assert.equal(p.test('closes #123'), true);
  assert.equal(p.test('closes # 123'), false);
  assert.equal(p.test('Closes # 123'), false);
});

test('factoryCheckDuplicatePR coalesces html_url for closes-literal mode when duplicate html_url is missing', () => {
  const result = factoryCheckDuplicatePR({
    issueNumber: 42,
    pullRequests: [{
      number: 101,
      state: 'open',
      head_branch: 'code-factory/issue-42',
      labels: ['code-factory'],
      body: 'Closes #42',
      html_url: undefined,
    }],
    branchPrefix: 'code-factory/issue-',
    prLabel: 'code-factory',
    duplicateLinkageMode: 'closes-literal',
  });
  assert.equal(result.duplicate_pr_found, true);
  assert.equal(result.duplicate_pr_url, null);
  assert.match(result.gate_reason, /\(unknown URL\)/);
});

// Keyword linkage is exercised at this pure-function boundary because no
// factory wires github-keywords into check-duplicate-pr.js. research-factory
// sets the mode but does not run that runner, and finalize-gate forces its
// duplicate result off.
test('factoryCheckDuplicatePR coalesces html_url for github-keywords mode when duplicate html_url is missing', () => {
  const result = factoryCheckDuplicatePR({
    issueNumber: 42,
    pullRequests: [{
      number: 101,
      state: 'open',
      head_branch: 'change-factory/issue-42',
      labels: ['change-factory'],
      body: 'Closes #42',
      html_url: undefined,
    }],
    branchPrefix: 'change-factory/issue-',
    prLabel: 'change-factory',
    duplicateLinkageMode: 'github-keywords',
  });
  assert.equal(result.duplicate_pr_found, true);
  assert.equal(result.duplicate_pr_url, null);
  assert.match(result.gate_reason, /\(unknown URL\)/);
});

test('factoryCheckDuplicatePR github-keywords matches closing keywords but not Related to or a longer issue number', () => {
  const base = {
    issueNumber: 42,
    branchPrefix: 'change-factory/issue-',
    prLabel: 'change-factory',
    duplicateLinkageMode: 'github-keywords',
  };
  const pr = (body) => ({
    number: 1,
    state: 'open',
    head_branch: 'change-factory/issue-42',
    labels: ['change-factory'],
    body,
    html_url: 'https://example.com/pr/1',
  });

  assert.equal(factoryCheckDuplicatePR({ ...base, pullRequests: [pr('Fixes #42')] }).duplicate_pr_found, true);
  assert.equal(factoryCheckDuplicatePR({ ...base, pullRequests: [pr('Related to #42')] }).duplicate_pr_found, false);
  assert.equal(factoryCheckDuplicatePR({ ...base, pullRequests: [pr('fixes #420')] }).duplicate_pr_found, false);
});

test('factoryCheckDuplicatePR coalesces html_url for related-literal mode when duplicate html_url is missing', () => {
  const result = factoryCheckDuplicatePR({
    issueNumber: 42,
    pullRequests: [{
      number: 101,
      state: 'open',
      head_branch: 'reproducer-factory/issue-42',
      labels: ['reproducer-factory'],
      body: 'Related to #42',
      html_url: undefined,
    }],
    branchPrefix: 'reproducer-factory/issue-',
    prLabel: 'reproducer-factory',
    duplicateLinkageMode: 'related-literal',
  });
  assert.equal(result.duplicate_pr_found, true);
  assert.equal(result.duplicate_pr_url, null);
  assert.match(result.gate_reason, /\(unknown URL\)/);
});

test('factoryCheckDuplicatePR related-literal matches Related to #42 but not Closes #42', () => {
  const base = {
    issueNumber: 42,
    branchPrefix: 'reproducer-factory/issue-',
    prLabel: 'reproducer-factory',
    duplicateLinkageMode: 'related-literal',
  };
  const related = factoryCheckDuplicatePR({
    ...base,
    pullRequests: [{
      number: 1,
      state: 'open',
      head_branch: 'reproducer-factory/issue-42',
      labels: ['reproducer-factory'],
      body: 'Related to #42',
      html_url: 'https://example.com/pr/1',
    }],
  });
  assert.equal(related.duplicate_pr_found, true);

  const closes = factoryCheckDuplicatePR({
    ...base,
    pullRequests: [{
      number: 2,
      state: 'open',
      head_branch: 'reproducer-factory/issue-42',
      labels: ['reproducer-factory'],
      body: 'Closes #42',
      html_url: 'https://example.com/pr/2',
    }],
  });
  assert.equal(closes.duplicate_pr_found, false);

  for (const body of ['related to #42', 'Related to #420']) {
    const miss = factoryCheckDuplicatePR({
      ...base,
      pullRequests: [{
        number: 3,
        state: 'open',
        head_branch: 'reproducer-factory/issue-42',
        labels: ['reproducer-factory'],
        body,
        html_url: 'https://example.com/pr/3',
      }],
    });
    assert.equal(miss.duplicate_pr_found, false, body);
  }

  const wrongBranch = factoryCheckDuplicatePR({
    ...base,
    pullRequests: [{
      number: 4,
      state: 'open',
      head_branch: 'reproducer-factory/issue-421',
      labels: ['reproducer-factory'],
      body: 'Related to #42',
      html_url: 'https://example.com/pr/4',
    }],
  });
  assert.equal(wrongBranch.duplicate_pr_found, false);
});

test('factoryCheckDuplicatePR ignores a closed PR even when linkage, label, and branch match', () => {
  const result = factoryCheckDuplicatePR({
    issueNumber: 42,
    pullRequests: [{
      number: 101,
      state: 'closed',
      head_branch: 'code-factory/issue-42',
      labels: ['code-factory'],
      body: 'Closes #42',
      html_url: 'https://example.com/pr/101',
    }],
    branchPrefix: 'code-factory/issue-',
    prLabel: 'code-factory',
    duplicateLinkageMode: 'closes-literal',
  });
  assert.equal(result.duplicate_pr_found, false);
});

test('factoryCheckDuplicatePR closes-literal matches Closes #42 but not Related to #42', () => {
  const base = {
    issueNumber: 42,
    branchPrefix: 'code-factory/issue-',
    prLabel: 'code-factory',
    duplicateLinkageMode: 'closes-literal',
  };
  const closes = factoryCheckDuplicatePR({
    ...base,
    pullRequests: [{
      number: 1,
      state: 'open',
      head_branch: 'code-factory/issue-42',
      labels: ['code-factory'],
      body: 'Closes #42',
      html_url: 'https://example.com/pr/1',
    }],
  });
  assert.equal(closes.duplicate_pr_found, true);

  const afterWord = factoryCheckDuplicatePR({
    ...base,
    pullRequests: [{
      number: 5,
      state: 'open',
      head_branch: 'code-factory/issue-42',
      labels: ['code-factory'],
      body: 'See Closes #42',
      html_url: 'https://example.com/pr/5',
    }],
  });
  assert.equal(afterWord.duplicate_pr_found, true);

  const related = factoryCheckDuplicatePR({
    ...base,
    pullRequests: [{
      number: 2,
      state: 'open',
      head_branch: 'code-factory/issue-42',
      labels: ['code-factory'],
      body: 'Related to #42',
      html_url: 'https://example.com/pr/2',
    }],
  });
  assert.equal(related.duplicate_pr_found, false);

  for (const body of ['closes #42', 'Closes #420', 'DisCloses #42']) {
    const miss = factoryCheckDuplicatePR({
      ...base,
      pullRequests: [{
        number: 3,
        state: 'open',
        head_branch: 'code-factory/issue-42',
        labels: ['code-factory'],
        body,
        html_url: 'https://example.com/pr/3',
      }],
    });
    assert.equal(miss.duplicate_pr_found, false, body);
  }

  const wrongBranch = factoryCheckDuplicatePR({
    ...base,
    pullRequests: [{
      number: 4,
      state: 'open',
      head_branch: 'code-factory/issue-421',
      labels: ['code-factory'],
      body: 'Closes #42',
      html_url: 'https://example.com/pr/4',
    }],
  });
  assert.equal(wrongBranch.duplicate_pr_found, false);
});

test('createFactoryIssueIntake: duplicateLinkageMode selects duplicate PR URL handling', () => {
  const code = createFactoryIssueIntake({
    branchPrefix: 'code-factory/issue-',
    factoryLabel: 'code-factory',
    issueOpenedNotEligibleReason: 'x',
    duplicateLinkageMode: 'closes-literal',
  });
  const change = createFactoryIssueIntake({
    branchPrefix: 'change-factory/issue-',
    factoryLabel: 'change-factory',
    issueOpenedNotEligibleReason: 'y',
    duplicateLinkageMode: 'github-keywords',
  });
  const pr = {
    number: 1,
    state: 'open',
    head_branch: 'code-factory/issue-9',
    labels: ['code-factory'],
    body: 'Closes #9',
    html_url: undefined,
  };
  const c = code.checkDuplicatePR({ issueNumber: 9, pullRequests: [pr] });
  assert.equal(c.duplicate_pr_url, null);
  const p = {
    ...pr,
    head_branch: 'change-factory/issue-9',
    labels: ['change-factory'],
  };
  const ch = change.checkDuplicatePR({ issueNumber: 9, pullRequests: [p] });
  assert.equal(ch.duplicate_pr_url, null);
});

test('createFactoryIssueModule binds shared exports and branch aliases', () => {
  const mod = createFactoryIssueModule({
    branchPrefix: 'demo-factory/issue-',
    factoryLabel: 'demo-factory',
    issueOpenedNotEligibleReason: 'not eligible',
    duplicateLinkageMode: 'closes-literal',
    issueBranchNameAliases: ['demoFactoryIssueBranchName'],
  });

  assert.equal(mod.issueBranchName(42), 'demo-factory/issue-42');
  assert.equal(mod.demoFactoryIssueBranchName(42), 'demo-factory/issue-42');
  assert.equal(mod.parseOptionalTriStateFromEnv('true'), true);
  assert.deepEqual(mod.parseFinalizeGateEnv({}), factoryParseFinalizeGateEnv({}));

  const event = mod.qualifyTriggerEvent({
    eventName: 'issues',
    eventAction: 'labeled',
    labelName: 'demo-factory',
    issueLabels: [],
  });
  assert.equal(event.event_eligible, true);
});
