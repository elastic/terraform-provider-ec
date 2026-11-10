import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import fs from 'node:fs';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const {
  factoryFetchIssueComments,
  serializeIssueComments,
  COMMENT_CONTEXT_BUDGET,
  COMMENT_CONTEXT_MARKER_OVERHEAD,
  COMMENT_BODY_TRUNCATED_NOTICE,
  COMMENT_BODY_OPENER_NOTICE,
} = require('./factory-issue-comments.js');
const { dropOrphanedCommentOpener, sliceToBytes, OPENER_NOTICE, TRUNCATION_NOTICE } = require('./env-payload.js');
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

describe('truncation markers', () => {
  // The prompt documents these verbatim in its marker table, and the agent is told to read one as
  // "there was more here". Importing the constants removed the only literal pin, so pin both
  // directions here: the string, and the prompt that quotes it.
  const prompt = fs.readFileSync(
    new URL('../../../workflows/research-factory-issue.md', import.meta.url),
    'utf8',
  );

  test('the comment markers read as truncation notices', () => {
    assert.equal(COMMENT_BODY_TRUNCATED_NOTICE.trim(), '[... this comment was truncated for context budget]');
    assert.equal(COMMENT_BODY_OPENER_NOTICE.trim(), '[... this comment was cut at an unterminated HTML comment]');
  });

  test('the prompt documents every marker the code can emit', () => {
    for (const marker of [COMMENT_BODY_TRUNCATED_NOTICE, COMMENT_BODY_OPENER_NOTICE, OPENER_NOTICE, TRUNCATION_NOTICE]) {
      assert.ok(prompt.includes(marker.trim()), `undocumented marker: ${marker.trim()}`);
    }
    assert.ok(prompt.includes('comment history truncated at 200 comments'));
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

  test('one comment ending in an unterminated opener cannot erase the comments after it', () => {
    // The whole serialized history goes through sanitizeUserContent, whose stripHtmlComments
    // deletes from an orphaned `<!--` to the end of the *string*. Before this, a single reporter
    // comment reading "Example input: <!--" silently removed every maintainer comment below it.
    const out = serializeIssueComments({
      comments: [
        { author: 'reporter', createdAt: 't1', body: 'Example input: <!--' },
        { author: 'maintainer', createdAt: 't2', body: 'IMPORTANT: this must also support ECE.' },
        { author: 'maintainer', createdAt: 't3', body: 'And the serverless variant.' },
      ],
      truncated: false,
    });
    const sanitized = sanitizeUserContent(out);

    assert.ok(sanitized.includes('must also support ECE'), 'later feedback survives');
    assert.ok(sanitized.includes('serverless variant'), 'and the comment after that');
    assert.ok(sanitized.includes(COMMENT_BODY_OPENER_NOTICE.trim()));
  });

  test('leaves a terminated comment in a body alone', () => {
    const out = serializeIssueComments({
      comments: [{ author: 'alice', createdAt: 't', body: 'see <!-- a note --> here' }],
      truncated: false,
    });

    assert.ok(out.includes('see <!-- a note --> here'), 'untouched by the orphan drop');
    assert.ok(!out.includes('cut at an unterminated HTML comment'));
  });

  test('keeps both truncation markers readable through the downstream sanitiser', () => {
    // The byte cut must land *inside* a terminated HTML comment — routine, since issue templates
    // leave them in bodies — which orphans the opener at slice time. Without dropping it,
    // sanitizeUserContent deletes from there to the end, taking the in-block notice and the
    // trailing drop marker with it.
    //
    // The offset is derived, not guessed: the slice happens at
    // (budget - marker overhead) - frame - notice, so the comment has to straddle that point. A
    // fixture whose comment sits earlier leaves this path untested while still passing.
    const frame = Buffer.byteLength('**@alice** (t):\n\n') + Buffer.byteLength('\n\n---\n');
    const cut = COMMENT_CONTEXT_BUDGET - COMMENT_CONTEXT_MARKER_OVERHEAD - frame - Buffer.byteLength(COMMENT_BODY_TRUNCATED_NOTICE);
    const body = `${'q'.repeat(cut - 10)}<!-- a template block that straddles the cut -->${'r'.repeat(5_000)}`;
    // Non-vacuity: if the arithmetic ever stops straddling the cut, the assertions below would all
    // pass while testing nothing. Fail here instead.
    const slice = sliceToBytes(body, cut);
    assert.notEqual(dropOrphanedCommentOpener(slice), slice, 'fixture must straddle the cut');
    const out = serializeIssueComments({
      comments: [
        { author: 'alice', createdAt: 't', body },
        { author: 'bob', createdAt: 't', body: 'second' },
      ],
      truncated: false,
    });
    assert.ok(!out.includes('<!--'), 'the slice orphaned the opener and it was dropped');
    const sanitized = sanitizeUserContent(out);

    assert.match(sanitized, /\[\.\.\. this comment was truncated for context budget\]/);
    assert.match(sanitized, /\[\.\.\. 1 more comments truncated for context budget\]/);
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
