/**
 * What the compiled reproducer-factory workflow is able to do, asserted on the lock.
 *
 * The prompt asks the agent to open a pull request only when the bug reproduces and to put only the
 * repro test in it. A prompt is a request, not a control, so everything that has to hold regardless
 * of what the agent does is pinned here against `reproducer-factory-issue.lock.yml`: which safe
 * outputs exist, what the pull request may contain, and that the agent job itself cannot write.
 *
 * What is NOT here, because it is not in the lock: whether the paid acceptance suite runs. Opening
 * the PR starts no Buildkite build only while the PR's author has no standing in
 * `.buildkite/pull-requests.json`. The last test below pins the part of that which lives in the
 * repository; the other part is which identity the token behind `GH_AW_GITHUB_TOKEN` belongs to,
 * and no test here can see a secret.
 */
import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const here = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(here, '../../../..');
const lock = readFileSync(path.join(repoRoot, '.github/workflows/reproducer-factory-issue.lock.yml'), 'utf8');

function manifest() {
  const line = lock.split('\n').find((l) => l.startsWith('# gh-aw-manifest: '));
  assert.ok(line, 'the lock has no gh-aw-manifest header');
  return JSON.parse(line.slice('# gh-aw-manifest: '.length));
}

/** The safe-output handler config, as the safe_outputs job receives it. */
function handlerConfig() {
  const match = /^\s*GH_AW_SAFE_OUTPUTS_HANDLER_CONFIG: (".*")\s*$/m.exec(lock);
  assert.ok(match, 'the lock has no GH_AW_SAFE_OUTPUTS_HANDLER_CONFIG');
  return JSON.parse(JSON.parse(match[1]));
}

/** The text of one top-level job, from its `  <name>:` line to the next job. */
function jobBlock(name) {
  const lines = lock.split('\n');
  const start = lines.indexOf(`  ${name}:`);
  assert.notEqual(start, -1, `the lock has no ${name} job`);
  let end = lines.length;
  for (let i = start + 1; i < lines.length; i++) {
    if (/^  [A-Za-z_][\w-]*:\s*$/.test(lines[i])) { end = i; break; }
  }
  return lines.slice(start, end).join('\n');
}

/** The job-level `permissions:` map of a job. */
function jobPermissions(name) {
  const lines = jobBlock(name).split('\n');
  const start = lines.indexOf('    permissions:');
  assert.notEqual(start, -1, `job ${name} declares no permissions`);
  const permissions = {};
  for (let i = start + 1; i < lines.length; i++) {
    const entry = /^      ([\w-]+): (\w+)\s*$/.exec(lines[i]);
    if (!entry) break;
    permissions[entry[1]] = entry[2];
  }
  return permissions;
}

test('the agent can emit exactly three safe outputs', () => {
  const safeoutputs = manifest().mcp_servers.find((s) => s.name === 'safeoutputs');
  assert.deepEqual(safeoutputs.tools, ['create_pull_request', 'noop', 'update_reproducer_comment']);
});

test('the pull request is a draft on the exact repro branch and never closes the issue', () => {
  const pr = handlerConfig().create_pull_request;
  assert.equal(pr.draft, true);
  assert.equal(pr.max, 1);
  assert.equal(pr.auto_close_issue, false);
  assert.deepEqual(pr.labels, ['reproducer-factory']);
  // Without these the branch is salted, and neither the duplicate gate nor the Buildkite narrowing
  // would recognise it.
  assert.equal(pr.preserve_branch_name, true);
  assert.equal(pr.recreate_ref, true);
});

test('the pull request branch is bound to the triggering issue, and a failed PR opens no issue', () => {
  const pr = handlerConfig().create_pull_request;
  // A glob here would let a run for one issue force-recreate another issue's open PR branch: the
  // duplicate gate only looks at the branch of the issue that triggered the run.
  assert.deepEqual(pr.allowed_branches, [
    'reproducer-factory/issue-${{ github.event.issue.number || inputs.issue_number }}',
  ]);
  // The branch expression must be the one `concurrency` keys on, so both name the same issue.
  assert.ok(
    lock.includes('group: "reproducer-factory-issue-${{ github.event.issue.number || inputs.issue_number }}"') ||
      lock.includes('group: reproducer-factory-issue-${{ github.event.issue.number || inputs.issue_number }}'),
    'the concurrency group no longer uses the same issue-number expression as allowed_branches',
  );
  assert.equal(pr.fallback_as_issue, false);
});

test('the pull request can contain a repro acceptance test and nothing else', () => {
  const pr = handlerConfig().create_pull_request;
  // gh-aw's schema says the whole patch is refused when any file falls outside this list. The one
  // path is bound to the triggering issue with the same expression as the branch, so a run cannot
  // carry another issue's test. .buildkite/acceptance-test-selection.sh does not rely on this: it
  // re-checks the diff itself before narrowing the paid run.
  assert.deepEqual(pr.allowed_files, [
    'ec/acc/issue_${{ github.event.issue.number || inputs.issue_number }}_acc_test.go',
  ]);
  assert.equal(pr.protect_top_level_dot_folders, true);
  assert.ok(pr.protected_files.includes('go.mod'));
  assert.ok(pr.protected_files.includes('go.sum'));
});

test('the agent job itself is read-only', () => {
  assert.deepEqual(jobPermissions('agent'), { contents: 'read', issues: 'read', 'pull-requests': 'read' });
});

test('only the gh-aw safe_outputs job can push, and only after threat detection succeeded', () => {
  assert.equal(jobPermissions('safe_outputs').contents, 'write');
  assert.match(jobBlock('safe_outputs'), /^    if: .*needs\.detection\.result == 'success'/m);
  // The comment job must never gain the ability to push or to touch pull requests.
  assert.deepEqual(jobPermissions('update_reproducer_comment'), { contents: 'read', issues: 'write' });
  assert.deepEqual(jobPermissions('pre_activation'), { contents: 'read', issues: 'write', 'pull-requests': 'read' });
});

test('the agent only starts when no pull request is open on the repro branch', () => {
  // recreate_ref force-recreates the branch, so this clause is what protects an open PR.
  assert.match(jobBlock('activation'), /needs\.pre_activation\.outputs\.duplicate_pr_found != 'true'/);
});

test('the workflow runs on Claude through OpenRouter, not on the Copilot default', () => {
  const header = lock.split('\n').find((l) => l.startsWith('# gh-aw-metadata: '));
  assert.equal(JSON.parse(header.slice('# gh-aw-metadata: '.length)).agent_id, 'claude');
  assert.ok(lock.includes('OPENROUTER_API_KEY'));
});

test('the pull request author has no standing with Buildkite, as far as the repository can tell', () => {
  // Buildkite builds a PR on open for an author who is in allowed_list, is an org member
  // (allow_org_users), or has one of allowed_repo_permissions. The factory's PRs are opened with
  // `GH_AW_GITHUB_TOKEN || GITHUB_TOKEN`. With no such secret that is github-actions[bot], which is
  // none of the three, so a maintainer's `buildkite test this` comment is the only thing that
  // spends. This test pins what the repository controls:
  //   - the bot is not added to allowed_list (a deliberate, reviewed decision if it ever is);
  //   - the PR is still opened with that token expression.
  // It cannot see the secret. Setting GH_AW_GITHUB_TOKEN to a token of an org member or of a user
  // with write access makes every repro PR build on open, with this test still green: that is a
  // spend decision, documented in dev-docs/high-level/reproducer-factory.md.
  const config = JSON.parse(readFileSync(path.join(repoRoot, '.buildkite/pull-requests.json'), 'utf8'));
  const job = config.jobs.find((j) => j.pipeline_slug === 'terraform-provider-ec-acceptance');
  assert.deepEqual(job.allowed_list, ['renovate[bot]']);
  assert.equal(job.build_on_comment, true);
  assert.ok(new RegExp(job.trigger_comment_regex).test('buildkite test this'));

  const tokens = [...jobBlock('safe_outputs').matchAll(/^\s*github-token: (.+)$/gm)].map((m) => m[1].trim());
  assert.ok(tokens.length > 0, 'the safe_outputs job sets no github-token');
  for (const token of tokens) {
    assert.equal(token, '${{ secrets.GH_AW_GITHUB_TOKEN || secrets.GITHUB_TOKEN }}');
  }
});
