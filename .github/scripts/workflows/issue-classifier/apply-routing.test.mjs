import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
import { mkdtempSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';

const require = createRequire(import.meta.url);
const { applyRouting } = require('./apply-routing.js');

function run(state, items, env = {}) {
  const dir = mkdtempSync(path.join(tmpdir(), 'apply-routing-'));
  const output = path.join(dir, 'agent_output.json');
  writeFileSync(output, JSON.stringify({ items }));
  const calls = [];
  const labels = structuredClone(state.labels);
  function execFile(cmd, args, options) {
    calls.push(args.join(' '));
    const rest = args.slice(1);
    let method = 'GET';
    let apiPath = '';
    for (let i = 0; i < rest.length; i++) {
      if (rest[i] === '--method') {
        method = rest[++i];
        continue;
      }
      if (rest[i] === '--input' || rest[i] === '--paginate' || rest[i] === '--slurp') {
        if (rest[i] === '--input') i++;
        continue;
      }
      if (!rest[i].startsWith('-') && !apiPath) apiPath = rest[i];
    }
    const issue = apiPath.match(/issues\/(\d+)(?:\/(labels|comments)(?:\/([^/]+))?)?$/);
    const num = issue[1];
    const kind = issue[2];
    if (!kind && method === 'GET') {
      return JSON.stringify(state.prs?.includes(Number(num)) ? { pull_request: {} } : { number: Number(num) });
    }
    if (kind === 'labels' && method === 'GET') {
      const pages = state.pages?.[num] ?? [labels[num] ?? []];
      return JSON.stringify(pages.map((page) => page.map((name) => ({ name }))));
    }
    if (kind === 'labels' && method === 'DELETE') {
      const name = decodeURIComponent(issue[3]);
      if (state.failDelete?.includes(name)) {
        const error = new Error('delete failed');
        throw error;
      }
      labels[num] = (labels[num] ?? []).filter((label) => label !== name);
      if (state.pages?.[num]) {
        state.pages[num] = state.pages[num].map((page) => page.filter((label) => label !== name));
      }
      return '';
    }
    if (method === 'POST') {
      calls.push(options.input);
      if (kind === 'labels') {
        labels[num] = [...(labels[num] ?? []), ...JSON.parse(options.input).labels];
      }
      return '';
    }
    throw new Error(`unexpected ${args.join(' ')}`);
  }
  let thrown;
  try {
    applyRouting({
      env: {
        DETECTION_RESULT: 'success',
        GH_AW_DETECTION_CONCLUSION: '',
        GITHUB_REPOSITORY: 'elastic/terraform-provider-ec',
        GH_AW_AGENT_OUTPUT: output,
        ...env,
      },
      execFile,
    });
  } catch (err) {
    thrown = err;
  }
  return { calls, labels, thrown };
}

const item = {
  type: 'apply_routing',
  issue_number: 15,
  category: 'needs-research',
  comment: 'Thanks',
};

test('a detection warning writes nothing', () => {
  const result = run({ labels: { 15: ['needs-human'] } }, [item], { GH_AW_DETECTION_CONCLUSION: 'warning' });
  assert.equal(result.thrown.routingFailure, true);
  assert.deepEqual(result.calls, []);
  assert.deepEqual(result.labels[15], ['needs-human']);
});

test('a pull request number is rejected before any label write', () => {
  const result = run({ labels: { 9: [] }, prs: [9] }, [{ ...item, issue_number: 9 }]);
  assert.match(result.thrown.message, /pull request/);
  assert.deepEqual(result.calls, ['api repos/elastic/terraform-provider-ec/issues/9']);
});

test('conflicting categories write nothing', () => {
  const result = run({ labels: { 15: [] } }, [item, { ...item, category: 'needs-spec' }]);
  assert.match(result.thrown.message, /conflicting/);
  assert.deepEqual(result.calls, []);
});

test('stale routing labels on a later page are removed before the comment', () => {
  const result = run({
    labels: { 15: ['bug'] },
    pages: { 15: [['bug'], ['needs-human']] },
  }, [item]);
  assert.equal(result.thrown, undefined);
  assert.deepEqual(result.calls.map((call) => call.startsWith('{') ? JSON.parse(call) : call), [
    'api repos/elastic/terraform-provider-ec/issues/15',
    'api --paginate --slurp repos/elastic/terraform-provider-ec/issues/15/labels',
    'api --method DELETE repos/elastic/terraform-provider-ec/issues/15/labels/needs-human',
    'api --paginate --slurp repos/elastic/terraform-provider-ec/issues/15/labels',
    'api --method POST repos/elastic/terraform-provider-ec/issues/15/labels --input -',
    { labels: ['needs-research', 'triaged'] },
    'api --method POST repos/elastic/terraform-provider-ec/issues/15/comments --input -',
    { body: '<!-- gha-issue-classifier -->\nThanks' },
  ]);
});

test('a failed delete does not post labels or a comment', () => {
  const result = run({
    labels: { 15: ['needs-human'] },
    failDelete: ['needs-human'],
  }, [item]);
  assert.match(result.thrown.message, /failed to remove needs-human/);
  assert.equal(result.calls.some((call) => call.includes('--method POST')), false);
});
