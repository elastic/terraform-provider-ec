import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const {
  factoryFetchIssueComments,
  serializeIssueComments,
  COMMENT_CONTEXT_BUDGET,
} = require('./factory-issue-comments.js');

function githubWith(comments) {
  const calls = [];
  return {
    calls,
    github: {
      rest: { issues: { listComments: 'listComments' } },
      async paginate(route, params) {
        calls.push({ route, params });
        return comments;
      },
    },
  };
}

function comment(login, body, createdAt = '2026-01-01T00:00:00Z') {
  return { user: { login }, body, created_at: createdAt };
}

describe('factoryFetchIssueComments', () => {
  test('paginates with per_page 100 and no sort or direction override', async () => {
    const { github, calls } = githubWith([comment('alice', 'hello')]);
    const result = await factoryFetchIssueComments({
      github,
      owner: 'elastic',
      repo: 'terraform-provider-ec',
      issueNumber: 7,
    });

    assert.equal(calls.length, 1);
    assert.deepEqual(calls[0].params, {
      owner: 'elastic',
      repo: 'terraform-provider-ec',
      issue_number: 7,
      per_page: 100,
    });
    assert.equal(result.comments.length, 1);
    assert.equal(result.truncated, false);
  });

  test('drops every [bot] author and keeps humans in order', async () => {
    const { github } = githubWith([
      comment('github-actions[bot]', 'status'),
      comment('alice', 'first'),
      comment('renovate[bot]', 'deps'),
      comment('bob', 'second'),
    ]);
    const { comments } = await factoryFetchIssueComments({
      github,
      owner: 'o',
      repo: 'r',
      issueNumber: 1,
    });

    assert.deepEqual(
      comments.map((c) => c.author),
      ['alice', 'bob'],
    );
    assert.equal(comments[0].body, 'first');
    assert.equal(comments[0].createdAt, '2026-01-01T00:00:00Z');
  });

  test('caps at 200 human comments and reports truncation', async () => {
    const many = Array.from({ length: 205 }, (_, i) => comment(`user${i}`, `body ${i}`));
    const { github } = githubWith(many);
    const { comments, truncated } = await factoryFetchIssueComments({
      github,
      owner: 'o',
      repo: 'r',
      issueNumber: 1,
    });

    assert.equal(comments.length, 200);
    assert.equal(truncated, true);
  });

  test('tolerates comments with no user and no body', async () => {
    const { github } = githubWith([{}, { user: {}, body: null }]);
    const { comments } = await factoryFetchIssueComments({
      github,
      owner: 'o',
      repo: 'r',
      issueNumber: 1,
    });

    assert.equal(comments.length, 2);
    assert.deepEqual(comments[0], { author: '', createdAt: '', body: '' });
  });
});

describe('serializeIssueComments', () => {
  test('returns an empty string for no comments', () => {
    assert.equal(serializeIssueComments({ comments: [], truncated: false }), '');
    assert.equal(serializeIssueComments({ comments: null, truncated: false }), '');
  });

  test('serializes author, timestamp and body with a separator', () => {
    const out = serializeIssueComments({
      comments: [{ author: 'alice', createdAt: '2026-02-03T04:05:06Z', body: 'hello' }],
      truncated: false,
    });

    assert.equal(out, '**@alice** (2026-02-03T04:05:06Z):\n\nhello\n\n---\n');
  });

  test('stays within the context budget and reports the dropped remainder', () => {
    const big = 'x'.repeat(20_000);
    const comments = Array.from({ length: 6 }, (_, i) => ({
      author: `user${i}`,
      createdAt: '2026-01-01T00:00:00Z',
      body: big,
    }));

    const out = serializeIssueComments({ comments, truncated: false });

    assert.ok(out.length <= COMMENT_CONTEXT_BUDGET, `serialized ${out.length} chars`);
    assert.match(out, /\[\.\.\. \d+ more comments truncated for context budget\]/);
  });

  test('appends the 200-comment marker when the fetch truncated', () => {
    const out = serializeIssueComments({
      comments: [{ author: 'alice', createdAt: '', body: 'hi' }],
      truncated: true,
    });

    assert.match(out, /\[\.\.\. comment history truncated at 200 comments\]/);
  });
});
