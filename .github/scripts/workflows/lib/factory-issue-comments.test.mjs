import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const {
  factoryFetchIssueComments,
  serializeIssueComments,
  COMMENT_CONTEXT_BUDGET,
} = require('./factory-issue-comments.js');

// A github.paginate stub that returns the given array regardless of the listComments ref.
function mockGithub(comments) {
  return {
    rest: { issues: { listComments: 'listComments' } },
    paginate: async (_fn, _params) => comments,
  };
}

// ---------------------------------------------------------------------------
// factoryFetchIssueComments
// ---------------------------------------------------------------------------

test('factoryFetchIssueComments filters out bot comments', async () => {
  const github = mockGithub([
    { user: { login: 'alice' }, created_at: '2026-01-01T00:00:00Z', body: 'human one' },
    { user: { login: 'github-actions[bot]' }, created_at: '2026-01-02T00:00:00Z', body: 'bot noise' },
    { user: { login: 'bob' }, created_at: '2026-01-03T00:00:00Z', body: 'human two' },
  ]);
  const { comments, truncated } = await factoryFetchIssueComments({
    github, owner: 'o', repo: 'r', issueNumber: 1,
  });
  assert.equal(comments.length, 2);
  assert.deepEqual(comments.map((c) => c.author), ['alice', 'bob']);
  assert.equal(truncated, false);
});

test('factoryFetchIssueComments caps at 200 human comments and sets truncated', async () => {
  const many = [];
  for (let i = 0; i < 250; i++) {
    many.push({ user: { login: `user${i}` }, created_at: '2026-01-01T00:00:00Z', body: `c${i}` });
  }
  const { comments, truncated } = await factoryFetchIssueComments({
    github: mockGithub(many), owner: 'o', repo: 'r', issueNumber: 1,
  });
  assert.equal(comments.length, 200);
  assert.equal(truncated, true);
});

test('factoryFetchIssueComments tolerates missing user/body fields', async () => {
  const github = mockGithub([
    { created_at: '2026-01-01T00:00:00Z' },
    { user: {}, body: 'no login' },
  ]);
  const { comments } = await factoryFetchIssueComments({
    github, owner: 'o', repo: 'r', issueNumber: 1,
  });
  assert.equal(comments.length, 2);
  assert.equal(comments[0].author, '');
  assert.equal(comments[0].body, '');
  assert.equal(comments[1].body, 'no login');
});

// ---------------------------------------------------------------------------
// serializeIssueComments
// ---------------------------------------------------------------------------

test('serializeIssueComments returns empty string for no comments', () => {
  assert.equal(serializeIssueComments({ comments: [], truncated: false }), '');
  assert.equal(serializeIssueComments({ comments: null, truncated: false }), '');
});

test('serializeIssueComments renders author, timestamp and body', () => {
  const out = serializeIssueComments({
    comments: [{ author: 'alice', createdAt: '2026-01-01T00:00:00Z', body: 'hello' }],
    truncated: false,
  });
  assert.match(out, /\*\*@alice\*\* \(2026-01-01T00:00:00Z\):/);
  assert.match(out, /hello/);
  assert.match(out, /\n---\n/);
});

test('serializeIssueComments appends the 200-cap marker when truncated', () => {
  const out = serializeIssueComments({
    comments: [{ author: 'a', createdAt: '', body: 'x' }],
    truncated: true,
  });
  assert.match(out, /truncated at 200 comments/);
});

test('serializeIssueComments stays within the UTF-8 byte budget and reports remaining', () => {
  const big = 'x'.repeat(40_000);
  const comments = [
    { author: 'a', createdAt: '', body: big },
    { author: 'b', createdAt: '', body: big },
    { author: 'c', createdAt: '', body: big },
  ];
  const out = serializeIssueComments({ comments, truncated: false });
  assert.ok(Buffer.byteLength(out, 'utf8') <= COMMENT_CONTEXT_BUDGET, `budget exceeded: ${Buffer.byteLength(out, 'utf8')}`);
  assert.match(out, /more comments truncated for context budget/);
});

test('serializeIssueComments bounds a CJK-heavy thread by bytes (no E2BIG via env var)', () => {
  // 3 bytes/char in UTF-8; a code-unit budget would let this exceed 128 KB.
  const cjk = '中'.repeat(30_000);
  const comments = [
    { author: 'a', createdAt: '', body: cjk },
    { author: 'b', createdAt: '', body: cjk },
    { author: 'c', createdAt: '', body: cjk },
  ];
  const out = serializeIssueComments({ comments, truncated: false });
  const bytes = Buffer.byteLength(out, 'utf8');
  assert.ok(bytes <= COMMENT_CONTEXT_BUDGET, `byte budget exceeded: ${bytes}`);
  // And comfortably under Linux MAX_ARG_STRLEN (~128 KB).
  assert.ok(bytes < 128 * 1024, `env-var transport unsafe: ${bytes} bytes`);
});

test('serializeIssueComments truncates without splitting a multi-byte character', () => {
  const cjk = '中'.repeat(60_000);
  const out = serializeIssueComments({ comments: [{ author: 'a', createdAt: '', body: cjk }], truncated: false });
  // Round-tripping through UTF-8 must not introduce the replacement char U+FFFD.
  assert.doesNotMatch(out, /�/);
});
