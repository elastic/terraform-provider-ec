import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const fetchIssueComments = require('./fetch-issue-comments.js');

function mockCore() {
  const outputs = {};
  return {
    outputs,
    core: {
      setOutput(key, value) { outputs[key] = value; },
      info() {},
    },
  };
}

const context = { repo: { owner: 'o', repo: 'r' } };

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

test('fetch-issue-comments emits empty output when issue number is missing', async () => {
  await withEnv({ INPUT_ISSUE_NUMBER: undefined }, async () => {
    const { outputs, core } = mockCore();
    await fetchIssueComments({ github: {}, context, core });
    assert.equal(outputs.human_comments, '');
  });
});

test('fetch-issue-comments emits empty output when issue number is zero or negative', async () => {
  for (const val of ['0', '-5', 'notanumber']) {
    await withEnv({ INPUT_ISSUE_NUMBER: val }, async () => {
      const { outputs, core } = mockCore();
      await fetchIssueComments({ github: {}, context, core });
      assert.equal(outputs.human_comments, '', `expected empty for ${val}`);
    });
  }
});

test('fetch-issue-comments serializes fetched human comments', async () => {
  const github = {
    rest: { issues: { listComments: 'listComments' } },
    paginate: async () => [
      { user: { login: 'alice' }, created_at: '2026-01-01T00:00:00Z', body: 'hello' },
      { user: { login: 'dependabot[bot]' }, created_at: '2026-01-02T00:00:00Z', body: 'bot' },
    ],
  };
  await withEnv({ INPUT_ISSUE_NUMBER: '42' }, async () => {
    const { outputs, core } = mockCore();
    await fetchIssueComments({ github, context, core });
    assert.match(outputs.human_comments, /@alice/);
    assert.doesNotMatch(outputs.human_comments, /dependabot/);
  });
});
