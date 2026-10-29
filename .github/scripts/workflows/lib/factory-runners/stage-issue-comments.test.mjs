import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const fetchIssueComments = require('./stage-issue-comments.js');

async function withEnv(vars, fn) {
  const previous = {};
  for (const key of Object.keys(vars)) {
    previous[key] = process.env[key];
    if (vars[key] === undefined) delete process.env[key];
    else process.env[key] = vars[key];
  }
  try {
    return await fn();
  } finally {
    for (const key of Object.keys(previous)) {
      if (previous[key] === undefined) delete process.env[key];
      else process.env[key] = previous[key];
    }
  }
}

function apiComment({ login, createdAt, body }) {
  return { user: { login }, created_at: createdAt, body };
}

async function run({ factory = 'change-factory', payload, apiComments = [] }) {
  const runnerTemp = fs.mkdtempSync(path.join(os.tmpdir(), `runner-temp-${crypto.randomUUID()}-`));
  const outputs = {};
  const calls = [];
  const listComments = { marker: 'issues.listComments' };
  const github = {
    rest: { issues: { listComments } },
    paginate: async (fn, params) => {
      calls.push({ fn, params });
      return apiComments;
    },
  };
  try {
    await withEnv({ FACTORY_NAME: factory, RUNNER_TEMP: runnerTemp }, () =>
      fetchIssueComments({
        github,
        context: { eventName: 'issues', payload, repo: { owner: 'elastic', repo: 'terraform-provider-ec' } },
        core: {
          setOutput(key, value) { outputs[key] = value; },
          info() {},
        },
      }),
    );
    const staged = outputs.issue_comments_path
      ? JSON.parse(fs.readFileSync(outputs.issue_comments_path, 'utf8'))
      : null;
    return { outputs, calls, listComments, staged, runnerTemp };
  } finally {
    fs.rmSync(runnerTemp, { recursive: true, force: true });
  }
}

describe('stage-issue-comments', { concurrency: false }, () => {
  test('paginates the full comment list for the triggering issue', async () => {
    const { calls, listComments } = await run({
      payload: { issue: { number: 42 } },
      apiComments: [apiComment({ login: 'alice', createdAt: '2026-01-01T00:00:00Z', body: 'hi' })],
    });
    assert.equal(calls.length, 1);
    assert.equal(calls[0].fn, listComments);
    assert.deepEqual(calls[0].params, {
      owner: 'elastic',
      repo: 'terraform-provider-ec',
      issue_number: 42,
      per_page: 100,
    });
  });

  test('keeps every comment when the history is below the cap', async () => {
    const apiComments = Array.from({ length: 5 }, (_, i) =>
      apiComment({ login: `user${i}`, createdAt: '2026-01-01T00:00:00Z', body: `c${i}` }));
    const { outputs, staged } = await run({ payload: { issue: { number: 42 } }, apiComments });
    assert.equal(outputs.issue_comments_count, '5');
    assert.equal(outputs.comments_truncated, 'false');
    assert.equal(staged.truncated, false);
    assert.deepEqual(staged.comments.map(c => c.body), ['c0', 'c1', 'c2', 'c3', 'c4']);
  });

  // The research sticky comment is the newest one, so an oldest-first cap would drop exactly
  // the comment this factory exists to read.
  test('keeps the NEWEST 200 comments and reports the truncation', async () => {
    const apiComments = Array.from({ length: 250 }, (_, i) =>
      apiComment({ login: `user${i}`, createdAt: '2026-01-01T00:00:00Z', body: `c${i}` }));
    const { outputs, staged } = await run({ payload: { issue: { number: 42 } }, apiComments });
    assert.equal(outputs.issue_comments_count, '200');
    assert.equal(outputs.comments_truncated, 'true');
    assert.equal(staged.truncated, true);
    assert.equal(staged.comments.length, 200);
    assert.equal(staged.comments[0].body, 'c50');
    assert.equal(staged.comments.at(-1).body, 'c249');
  });

  test('retains bot comments', async () => {
    const { staged } = await run({
      payload: { issue: { number: 42 } },
      apiComments: [
        apiComment({ login: 'alice', createdAt: '2026-01-01T00:00:00Z', body: 'human' }),
        apiComment({ login: 'github-actions[bot]', createdAt: '2026-01-02T00:00:00Z', body: '<!-- gha-research-factory -->' }),
      ],
    });
    assert.deepEqual(staged.comments.map(c => c.author), ['alice', 'github-actions[bot]']);
  });

  // The research sticky comment is created early, so on a busy issue it falls outside the tail
  // cap. It must still be staged, or the agent loses the research baseline.
  test('retains the research comment even when the tail cap would drop it', async () => {
    const research = apiComment({
      login: 'github-actions[bot]',
      createdAt: '2026-01-01T00:00:00Z',
      body: '<!-- gha-research-factory -->\n## Implementation research\nbaseline',
    });
    const later = Array.from({ length: 249 }, (_, i) =>
      apiComment({ login: `user${i}`, createdAt: '2026-02-01T00:00:00Z', body: `c${i}` }));
    const { outputs, staged } = await run({
      payload: { issue: { number: 42 } },
      apiComments: [research, ...later], // research is the oldest, so -200 would drop it
    });
    assert.equal(outputs.comments_truncated, 'true');
    assert.equal(staged.comments.length, 201, 'research comment retained on top of the 200 cap');
    assert.equal(staged.comments[0].author, 'github-actions[bot]');
    assert.match(staged.comments[0].body, /gha-research-factory/);
    assert.equal(staged.comments.filter(c => c.author === 'github-actions[bot]').length, 1);
  });

  test('does not duplicate the research comment when it is already within the cap', async () => {
    const older = Array.from({ length: 5 }, (_, i) =>
      apiComment({ login: `user${i}`, createdAt: '2026-01-01T00:00:00Z', body: `c${i}` }));
    const research = apiComment({
      login: 'github-actions[bot]',
      createdAt: '2026-01-09T00:00:00Z',
      body: '<!-- gha-research-factory -->\nbaseline',
    });
    const { staged } = await run({
      payload: { issue: { number: 42 } },
      apiComments: [...older, research],
    });
    assert.equal(staged.comments.length, 6);
    assert.equal(staged.comments.filter(c => c.author === 'github-actions[bot]').length, 1);
  });

  test('tolerates comments with no user, body, or timestamp', async () => {
    const { staged } = await run({
      payload: { issue: { number: 42 } },
      apiComments: [{}],
    });
    assert.deepEqual(staged.comments, [{ author: '', createdAt: '', body: '' }]);
  });

  // D17: the staged file is raw, unsanitized comment text, so it must not land in the directory
  // that is uploaded to the agent.
  test('writes under RUNNER_TEMP, never into the uploaded context dir', async () => {
    const { outputs, runnerTemp } = await run({
      payload: { issue: { number: 42 } },
      apiComments: [apiComment({ login: 'alice', createdAt: '2026-01-01T00:00:00Z', body: 'hi' })],
    });
    assert.equal(outputs.issue_comments_path, `${runnerTemp}/change-factory-raw/comments.json`);
    assert.equal(outputs.issue_comments_path.startsWith('/tmp/change-factory-context'), false);
  });

  test('keys the staging directory off FACTORY_NAME', async () => {
    const { outputs, runnerTemp } = await run({
      factory: 'research-factory',
      payload: { issue: { number: 42 } },
      apiComments: [],
    });
    assert.equal(outputs.issue_comments_path, `${runnerTemp}/research-factory-raw/comments.json`);
  });

  test('falls back to the OS temp dir when RUNNER_TEMP is unset', async () => {
    const outputs = {};
    await withEnv({ FACTORY_NAME: 'change-factory', RUNNER_TEMP: undefined }, () =>
      fetchIssueComments({
        github: {
          rest: { issues: { listComments: {} } },
          paginate: async () => [apiComment({ login: 'alice', createdAt: '2026-01-01T00:00:00Z', body: 'hi' })],
        },
        context: { eventName: 'issues', payload: { issue: { number: 42 } }, repo: { owner: 'elastic', repo: 'terraform-provider-ec' } },
        core: { setOutput(key, value) { outputs[key] = value; }, info() {} },
      }),
    );
    assert.equal(outputs.issue_comments_path, `${os.tmpdir()}/change-factory-raw/comments.json`);
    assert.equal(outputs.issue_comments_path.includes('undefined'), false);
    assert.equal(path.isAbsolute(outputs.issue_comments_path), true);
    assert.equal(outputs.issue_comments_path.startsWith('/tmp/change-factory-context'), false);
    fs.rmSync(`${os.tmpdir()}/change-factory-raw`, { recursive: true, force: true });
  });

  test('returns empty outputs without throwing when the payload has no issue number', async () => {
    const { outputs, calls } = await run({ payload: {}, apiComments: [] });
    assert.equal(calls.length, 0);
    assert.equal(outputs.issue_comments_path, '');
    assert.equal(outputs.issue_comments_count, '0');
    assert.equal(outputs.comments_truncated, 'false');
  });
});
