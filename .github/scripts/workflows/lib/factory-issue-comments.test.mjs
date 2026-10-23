import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const {
  serializeIssueComments,
  COMMENT_CONTEXT_BUDGET,
  COMMENT_CONTEXT_MARKER_OVERHEAD,
} = require('./factory-issue-comments.js');

describe('serializeIssueComments', () => {
  test('returns an empty string for an empty or non-array input', () => {
    assert.equal(serializeIssueComments({ comments: [], truncated: false }), '');
    assert.equal(serializeIssueComments({ comments: undefined, truncated: false }), '');
    assert.equal(serializeIssueComments({ comments: null, truncated: false }), '');
    assert.equal(serializeIssueComments({ comments: 'nope', truncated: false }), '');
  });

  test('frames each comment deterministically', () => {
    const serialized = serializeIssueComments({
      comments: [
        { author: 'alice', createdAt: '2026-01-02T03:04:05Z', body: 'first' },
        { author: 'bob', createdAt: '2026-01-03T03:04:05Z', body: 'second' },
      ],
      truncated: false,
    });
    assert.equal(
      serialized,
      '**@alice** (2026-01-02T03:04:05Z):\n\nfirst\n\n---\n'
      + '**@bob** (2026-01-03T03:04:05Z):\n\nsecond\n\n---\n',
    );
  });

  test('tolerates missing author, date, and body fields', () => {
    assert.equal(serializeIssueComments({ comments: [{}], truncated: false }), '**@** ():\n\n\n\n---\n');
  });

  test('stops at the context budget and reports how many comments were dropped', () => {
    const big = 'x'.repeat(20_000);
    const comments = Array.from({ length: 5 }, (_, i) => ({
      author: `user${i}`,
      createdAt: '2026-01-01T00:00:00Z',
      body: big,
    }));
    const serialized = serializeIssueComments({ comments, truncated: false });
    assert.ok(serialized.length <= COMMENT_CONTEXT_BUDGET);
    assert.ok(serialized.length > COMMENT_CONTEXT_BUDGET - COMMENT_CONTEXT_MARKER_OVERHEAD - 2000);
    assert.match(serialized, /\[\.\.\. 2 more comments truncated for context budget\]\n/);
    // The last included comment is also body-truncated here, so that is flagged too.
    assert.match(serialized, /\[\.\.\. last included comment truncated for context budget\]\n$/);
  });

  test('flags partial-body truncation of a single oversized comment', () => {
    const serialized = serializeIssueComments({
      comments: [{ author: 'alice', createdAt: '2026-01-01T00:00:00Z', body: 'x'.repeat(60_000) }],
      truncated: false,
    });
    assert.ok(serialized.length <= COMMENT_CONTEXT_BUDGET);
    // Without the fix this reads as a complete comment; the notice is the only signal its tail was cut.
    assert.match(serialized, /\[\.\.\. last included comment truncated for context budget\]\n$/);
    assert.doesNotMatch(serialized, /more comments truncated for context budget/);
  });

  test('does not flag partial truncation when every included comment fits whole', () => {
    const serialized = serializeIssueComments({
      comments: [
        { author: 'alice', createdAt: '2026-01-01T00:00:00Z', body: 'short one' },
        { author: 'bob', createdAt: '2026-01-02T00:00:00Z', body: 'short two' },
      ],
      truncated: false,
    });
    assert.doesNotMatch(serialized, /last included comment truncated/);
  });

  test('appends the history marker when the caller reports truncation', () => {
    const serialized = serializeIssueComments({
      comments: [{ author: 'alice', createdAt: '2026-01-01T00:00:00Z', body: 'hi' }],
      truncated: true,
    });
    assert.match(serialized, /\[\.\.\. comment history truncated at 200 comments\]\n$/);
    assert.doesNotMatch(serialized, /more comments truncated for context budget/);
  });

  test('can append both truncation markers', () => {
    const big = 'x'.repeat(30_000);
    const comments = Array.from({ length: 4 }, (_, i) => ({
      author: `user${i}`,
      createdAt: '2026-01-01T00:00:00Z',
      body: big,
    }));
    const serialized = serializeIssueComments({ comments, truncated: true });
    // Order: omitted-comments notice, then partial-body notice (the last fit comment is cut),
    // then the history-cap notice.
    assert.match(serialized, /more comments truncated for context budget\]\n/);
    assert.match(serialized, /\[\.\.\. last included comment truncated for context budget\]\n\[\.\.\. comment history truncated at 200 comments\]\n$/);
  });
});
