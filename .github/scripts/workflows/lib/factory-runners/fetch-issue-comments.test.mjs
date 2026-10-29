import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const fetchIssueComments = require('./fetch-issue-comments.js');

function mockCore() {
  const outputs = {};
  const failed = [];
  return {
    outputs,
    failed,
    core: {
      setOutput(key, value) {
        outputs[key] = value;
      },
      setFailed(message) {
        failed.push(message);
      },
      info() {},
    },
  };
}

const context = { repo: { owner: 'elastic', repo: 'terraform-provider-ec' } };

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

function githubReturning(comments) {
  return {
    rest: { issues: { listComments: 'listComments' } },
    async paginate() {
      return comments;
    },
  };
}

describe('fetch-issue-comments runner', { concurrency: false }, () => {
  test('writes the serialized history to GITHUB_OUTPUT with a heredoc delimiter', async () => {
    const outputFile = path.join(fs.mkdtempSync(path.join(os.tmpdir(), 'fic-')), 'out');
    fs.writeFileSync(outputFile, '');
    const { core, outputs, failed } = mockCore();
    const github = githubReturning([
      { user: { login: 'alice' }, body: 'needs a data source', created_at: '2026-03-01T00:00:00Z' },
      { user: { login: 'github-actions[bot]' }, body: 'status', created_at: '2026-03-02T00:00:00Z' },
    ]);

    await withEnv({ INPUT_ISSUE_NUMBER: '42', GITHUB_OUTPUT: outputFile }, () =>
      fetchIssueComments({ github, context, core }),
    );

    assert.deepEqual(failed, []);
    assert.equal(outputs.comment_count, '1');
    const written = fs.readFileSync(outputFile, 'utf8');
    const [header] = written.split('\n');
    assert.match(header, /^issue_comments<<EOF_[0-9a-f]{32}$/);
    assert.ok(written.includes('**@alice** (2026-03-01T00:00:00Z):'));
    assert.ok(!written.includes('github-actions[bot]'));
  });

  test('fails closed on an invalid issue number', async () => {
    const { core, outputs, failed } = mockCore();
    await withEnv({ INPUT_ISSUE_NUMBER: '', GITHUB_OUTPUT: undefined }, () =>
      fetchIssueComments({ github: githubReturning([]), context, core }),
    );

    assert.equal(outputs.comment_count, '0');
    assert.equal(outputs.issue_comments, '');
    assert.equal(failed.length, 1);
    assert.match(failed[0], /invalid issue number/);
  });

  test('fails closed when the API call throws', async () => {
    const { core, outputs, failed } = mockCore();
    const github = {
      rest: { issues: { listComments: 'listComments' } },
      async paginate() {
        throw new Error('boom');
      },
    };

    await withEnv({ INPUT_ISSUE_NUMBER: '42', GITHUB_OUTPUT: undefined }, () =>
      fetchIssueComments({ github, context, core }),
    );

    assert.equal(outputs.comment_count, '0');
    assert.equal(outputs.issue_comments, '');
    assert.equal(failed.length, 1);
    assert.match(failed[0], /Failed to fetch issue comments for #42: boom/);
  });
});
