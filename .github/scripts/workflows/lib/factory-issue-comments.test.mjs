import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const {
  factoryFetchIssueComments,
  serializeIssueComments,
  COMMENT_CONTEXT_BUDGET,
} = require('./factory-issue-comments.js');
const { sanitizeUserContent } = require('./sanitize-context.js');

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

  test('marks a comment whose body was sliced, even when it is the only one', () => {
    // Without the notice, a sliced comment is indistinguishable from a complete one: the trailing
    // "N more comments truncated" marker does not fire when nothing was dropped wholesale.
    const out = serializeIssueComments({
      comments: [{ author: 'alice', createdAt: '2026-01-01T00:00:00Z', body: 'y'.repeat(COMMENT_CONTEXT_BUDGET * 2) }],
      truncated: false,
    });

    assert.ok(Buffer.byteLength(out) <= COMMENT_CONTEXT_BUDGET);
    assert.match(out, /\[\.\.\. this comment was truncated for context budget\]/);
    assert.match(out, /\[\.\.\. a comment above was truncated for context budget\]/);
  });

  test('reserves room for the truncation notice inside the budget', () => {
    // Without the reservation the sliced block overshoots `available` by the notice length and the
    // total only stays legal because the marker overhead happens to absorb it.
    const comments = Array.from({ length: 3 }, (_, i) => ({
      author: `user${i}`,
      createdAt: '2026-01-01T00:00:00Z',
      body: 'z'.repeat(COMMENT_CONTEXT_BUDGET),
    }));
    const out = serializeIssueComments({ comments, truncated: false });

    assert.ok(Buffer.byteLength(out) <= COMMENT_CONTEXT_BUDGET - 100, `used ${Buffer.byteLength(out)} of ${COMMENT_CONTEXT_BUDGET}`);
    assert.match(out, /\[\.\.\. this comment was truncated for context budget\]/);
  });

  test('keeps both truncation markers readable through the downstream sanitiser', () => {
    // The cut lands inside an HTML comment — routine, since issue templates leave them in bodies.
    // Without dropping the orphaned opener, sanitizeUserContent deletes from there to the end,
    // taking the in-block notice and the trailing drop marker with it.
    const out = serializeIssueComments({
      comments: [
        { author: 'alice', createdAt: 't', body: `${'q'.repeat(49_000)}<!-- template ${'r'.repeat(5_000)}` },
        { author: 'bob', createdAt: 't', body: 'second' },
      ],
      truncated: false,
    });
    const sanitized = sanitizeUserContent(out);

    assert.match(sanitized, /\[\.\.\. this comment was truncated for context budget\]/);
    assert.match(sanitized, /\[\.\.\. a comment above was truncated for context budget\]/);
    assert.ok(sanitized.includes('**@bob**'), 'the later comment survives the orphaned opener');
  });

  test('accounts for whole blocks in bytes, not characters', () => {
    // Twelve comments that each fit whole: measuring the block with `.length` keeps every one of
    // them and emits ~130 KB — past the kernel limit this whole budget exists to stay under.
    const comments = Array.from({ length: 12 }, (_, i) => ({
      author: `user${i}`,
      createdAt: '2026-01-01T00:00:00Z',
      body: '漢'.repeat(4_000),
    }));
    const out = serializeIssueComments({ comments, truncated: false });

    assert.ok(Buffer.byteLength(out) <= COMMENT_CONTEXT_BUDGET, `${Buffer.byteLength(out)} bytes`);
    assert.ok(Buffer.byteLength(out) < 128 * 1024);
  });

  test('accounts for the budget in bytes across several multi-byte comments', () => {
    // Two comments: tracking `result.length` instead of bytes, or measuring the block with
    // `.length`, both overshoot here while agreeing on a single-comment fixture.
    const comments = Array.from({ length: 2 }, (_, i) => ({
      author: `user${i}`,
      createdAt: '2026-01-01T00:00:00Z',
      body: '漢'.repeat(20_000),
    }));
    const out = serializeIssueComments({ comments, truncated: false });

    assert.ok(Buffer.byteLength(out) <= COMMENT_CONTEXT_BUDGET, `${Buffer.byteLength(out)} bytes`);
  });

  test('drops a comment whose frame and notice alone would not fit', () => {
    // The `room < 0` guard. Sized so the second comment has room for its frame (19 bytes) but not
    // for frame + notice: without the guard an empty-bodied block is emitted that overshoots
    // `available`, and the comment is miscounted as included.
    const bodyBudget = COMMENT_CONTEXT_BUDGET - 200;
    const frame = '**@a** (t):\n\n'.length + '\n\n---\n'.length;
    const first = { author: 'a', createdAt: 't', body: 'x'.repeat(bodyBudget - 40 - frame) };
    const second = { author: 'b', createdAt: 't', body: 'y'.repeat(100) };
    const out = serializeIssueComments({ comments: [first, second], truncated: false });

    assert.ok(Buffer.byteLength(out) <= COMMENT_CONTEXT_BUDGET);
    assert.match(out, /\[\.\.\. 1 more comments truncated for context budget\]/);
    assert.ok(!out.includes('**@b**'), 'the comment that could not fit its own frame is dropped, not stubbed');
  });

  test('budgets in UTF-8 bytes, not characters', () => {
    // 50 000 CJK characters are 150 000 bytes. Budgeting by character would hand that to an
    // environment variable and blow Linux's 128 KiB per-entry limit.
    const out = serializeIssueComments({
      comments: [{ author: 'alice', createdAt: '', body: '漢'.repeat(50_000) }],
      truncated: false,
    });

    assert.ok(Buffer.byteLength(out) <= COMMENT_CONTEXT_BUDGET, `${Buffer.byteLength(out)} bytes`);
    assert.ok(out.isWellFormed(), 'no code point was cut in half');
  });

  test('appends the 200-comment marker when the fetch truncated', () => {
    const out = serializeIssueComments({
      comments: [{ author: 'alice', createdAt: '', body: 'hi' }],
      truncated: true,
    });

    assert.match(out, /\[\.\.\. comment history truncated at 200 comments\]/);
  });
});
