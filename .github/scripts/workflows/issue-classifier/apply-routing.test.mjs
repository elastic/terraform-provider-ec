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
    if (method === 'PATCH') {
      // Issue comment edit: repos/.../issues/comments/{id}
      if (state.failPatch) {
        throw new Error('comment edit failed');
      }
      calls.push(options.input);
      return '';
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
    if (kind === 'comments' && method === 'GET') {
      const existing = state.comments?.[num] ?? [];
      return JSON.stringify([existing.map((c, idx) =>
        typeof c === 'string'
          ? { id: 1000 + idx, body: c, user: { login: 'github-actions[bot]' } }
          : { id: c.id ?? 1000 + idx, body: c.body, user: { login: c.login } })]);
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
      if (state.failPost?.includes(kind)) {
        throw new Error(`${kind} post failed`);
      }
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

test('stale routing labels on a later page are removed, then category, comment, triaged', () => {
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
    { labels: ['needs-research'] },
    'api --paginate --slurp repos/elastic/terraform-provider-ec/issues/15/comments',
    'api --method POST repos/elastic/terraform-provider-ec/issues/15/comments --input -',
    { body: '<!-- gha-issue-classifier -->\nThanks' },
    'api --method POST repos/elastic/terraform-provider-ec/issues/15/labels --input -',
    { labels: ['triaged'] },
  ]);
});

test('triaged is written after the category label and comment', () => {
  const result = run({ labels: { 15: [] } }, [item]);
  assert.equal(result.thrown, undefined);
  const posts = result.calls.filter((call) => call.startsWith('{')).map((call) => JSON.parse(call));
  assert.deepEqual(posts, [
    { labels: ['needs-research'] },
    { body: '<!-- gha-issue-classifier -->\nThanks' },
    { labels: ['triaged'] },
  ]);
  assert.deepEqual(result.labels[15], ['needs-research', 'triaged']);
});

test('an existing identical bot comment is neither reposted nor edited', () => {
  const result = run({
    labels: { 15: ['needs-research'] },
    comments: { 15: ['<!-- gha-issue-classifier -->\nThanks'] },
  }, [item]);
  assert.equal(result.thrown, undefined);
  assert.equal(result.calls.some((call) => call.includes('/comments --input') || call.includes('/issues/comments/')), false);
  assert.deepEqual(result.labels[15], ['needs-research', 'triaged']);
});

test('a marker comment from a non-bot author does not suppress the bot explanation', () => {
  const result = run({
    labels: { 15: [] },
    comments: { 15: [{ body: '<!-- gha-issue-classifier -->\nfake', login: 'reporter' }] },
  }, [item]);
  assert.equal(result.thrown, undefined);
  // The reporter's marker comment is ignored; the real explanation is still posted.
  assert.equal(result.calls.some((call) => call.includes('--method POST') && call.includes('/comments --input')), true);
  assert.equal(result.calls.some((call) => call.includes('--method PATCH')), false);
  assert.deepEqual(result.labels[15], ['needs-research', 'triaged']);
});

test('a retry that reclassified to a different category reconciles the existing comment', () => {
  const result = run({
    labels: { 15: ['needs-research'] },
    comments: { 15: [{ id: 42, body: '<!-- gha-issue-classifier -->\nResearch', login: 'github-actions[bot]' }] },
  }, [{ ...item, category: 'needs-spec', comment: 'Spec' }]);
  assert.equal(result.thrown, undefined);
  // Old label removed, new category applied, triaged last.
  assert.deepEqual(result.labels[15], ['needs-spec', 'triaged']);
  // The bot comment is edited (not reposted) so the explanation matches the new label.
  assert.equal(result.calls.some((call) => call.includes('--method PATCH') && call.includes('/issues/comments/42')), true);
  assert.equal(result.calls.some((call) => call.includes('--method POST') && call.includes('/comments --input')), false);
  const patched = result.calls.filter((call) => call.startsWith('{')).map((call) => JSON.parse(call)).find((obj) => obj.body);
  assert.equal(patched.body, '<!-- gha-issue-classifier -->\nSpec');
});

test('a failed comment edit on a changed category leaves the issue untriaged', () => {
  const result = run({
    labels: { 15: ['needs-research'] },
    comments: { 15: [{ id: 42, body: '<!-- gha-issue-classifier -->\nResearch', login: 'github-actions[bot]' }] },
    failPatch: true,
  }, [{ ...item, category: 'needs-spec', comment: 'Spec' }]);
  assert.notEqual(result.thrown, undefined);
  assert.equal(result.labels[15].includes('triaged'), false);
});

test('GH_AW_BOT_LOGIN overrides the recognised comment author', () => {
  // Under the override, a comment by custom[bot] is the bot's own -> idempotent skip.
  const recognised = run({
    labels: { 15: ['needs-research'] },
    comments: { 15: [{ body: '<!-- gha-issue-classifier -->\nThanks', login: 'custom[bot]' }] },
  }, [item], { GH_AW_BOT_LOGIN: 'custom[bot]' });
  assert.equal(recognised.thrown, undefined);
  assert.equal(recognised.calls.some((call) => call.includes('/comments --input') || call.includes('/issues/comments/')), false);

  // A default-login comment is NOT the bot's own under the override -> real comment is posted.
  const ignored = run({
    labels: { 15: ['needs-research'] },
    comments: { 15: [{ body: '<!-- gha-issue-classifier -->\nThanks', login: 'github-actions[bot]' }] },
  }, [item], { GH_AW_BOT_LOGIN: 'custom[bot]' });
  assert.equal(ignored.thrown, undefined);
  assert.equal(ignored.calls.some((call) => call.includes('--method POST') && call.includes('/comments --input')), true);
});

test('an issue that already has triaged is left unchanged', () => {
  const result = run({ labels: { 15: ['triaged', 'needs-human'] } }, [item]);
  assert.equal(result.thrown, undefined);
  assert.equal(result.calls.some((call) => call.includes('DELETE') || call.includes('POST')), false);
  assert.deepEqual(result.labels[15], ['triaged', 'needs-human']);
});

test('a failed delete does not post labels or a comment', () => {
  const result = run({
    labels: { 15: ['needs-human'] },
    failDelete: ['needs-human'],
  }, [item]);
  assert.match(result.thrown.message, /failed to remove needs-human/);
  assert.equal(result.calls.some((call) => call.includes('--method POST')), false);
});

test('a failed comment post leaves the issue untriaged for a clean retry', () => {
  const result = run({
    labels: { 15: ['needs-research'] },
    failPost: ['comments'],
  }, [item]);
  assert.notEqual(result.thrown, undefined);
  // triaged is the completion marker; it must not be written when the comment fails.
  assert.equal(result.labels[15].includes('triaged'), false);
  assert.deepEqual(result.labels[15], ['needs-research']);
});

test('a non-classifier comment does not suppress the classification comment', () => {
  const result = run({
    labels: { 15: [] },
    comments: { 15: ['just a user comment'] },
  }, [item]);
  assert.equal(result.thrown, undefined);
  assert.equal(result.calls.filter((call) => call.includes('/comments --input')).length, 1);
  assert.deepEqual(result.labels[15], ['needs-research', 'triaged']);
});

test('a category already present from a partial run is not re-applied; comment and triaged complete', () => {
  const result = run({ labels: { 15: ['needs-research'] } }, [item]);
  assert.equal(result.thrown, undefined);
  const labelPosts = result.calls
    .filter((call) => call.startsWith('{') && JSON.parse(call).labels)
    .map((call) => JSON.parse(call).labels);
  assert.deepEqual(labelPosts, [['triaged']]);
  assert.equal(result.calls.filter((call) => call.includes('/comments --input')).length, 1);
  assert.deepEqual(result.labels[15], ['needs-research', 'triaged']);
});
