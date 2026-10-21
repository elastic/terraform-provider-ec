import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { boundEnvPayload, sliceToBytes, byteLength, dropOrphanedCommentOpener, MAX_ENV_PAYLOAD_BYTES, TRUNCATION_NOTICE } = require('./env-payload.js');
const { sanitizeUserContent } = require('./sanitize-context.js');

describe('byteLength', () => {
  test('counts UTF-8 bytes, not characters', () => {
    assert.equal(byteLength('abc'), 3);
    assert.equal(byteLength('漢'), 3);
    assert.equal(byteLength('😀'), 4);
    assert.equal(byteLength(undefined), 0);
  });
});

describe('sliceToBytes', () => {
  test('returns the input when it already fits', () => {
    assert.equal(sliceToBytes('hello', 10), 'hello');
  });

  test('never splits a multi-byte code point', () => {
    // 3 bytes of budget fits one 漢; 5 bytes still fits only one.
    assert.equal(sliceToBytes('漢漢', 3), '漢');
    assert.equal(sliceToBytes('漢漢', 5), '漢');
    assert.equal(sliceToBytes('漢漢', 6), '漢漢');
  });

  test('never splits a surrogate pair', () => {
    const out = sliceToBytes('😀😀', 7);
    assert.equal(out, '😀');
    assert.ok(out.isWellFormed());
  });

  test('returns empty for a non-positive budget', () => {
    assert.equal(sliceToBytes('abc', 0), '');
    assert.equal(sliceToBytes('abc', -1), '');
  });
});

describe('MAX_ENV_PAYLOAD_BYTES', () => {
  test('is 100 KiB: under the kernel limit with real head-room', () => {
    // MAX_ARG_STRLEN is 128 KiB per entry. The margin is the point — several of these payloads are
    // exported by one step, and a value chosen just under the limit leaves nothing for the rest.
    assert.equal(MAX_ENV_PAYLOAD_BYTES, 100 * 1024);
    assert.ok(MAX_ENV_PAYLOAD_BYTES < 128 * 1024 - 16 * 1024);
  });
});

describe('dropOrphanedCommentOpener', () => {
  test('removes an opener the cut left unterminated', () => {
    assert.equal(dropOrphanedCommentOpener('text <!-- note'), 'text ');
  });

  test('leaves a terminated comment, and text with no opener, alone', () => {
    assert.equal(dropOrphanedCommentOpener('text <!-- note --> more'), 'text <!-- note --> more');
    assert.equal(dropOrphanedCommentOpener('plain text'), 'plain text');
  });

  test('handles the overlapping `<!-->` shape the regex treats as unterminated', () => {
    // `indexOf('-->', opener)` finds the closer *inside* the opener and judges it terminated, while
    // stripHtmlComments resumes after the 4-character opener, finds nothing, and deletes to the end.
    assert.equal(dropOrphanedCommentOpener('text <!-->'), 'text ');
    const bounded = boundEnvPayload(`${'a'.repeat(120)}<!-->${'z'.repeat(400)}`, 200);
    assert.ok(sanitizeUserContent(bounded).includes('truncated: payload exceeded'));
  });

  test('treats `<!--->` as unterminated, the way the sanitiser does', () => {
    // The closer must be sought from `opener + 4`: from +3 the trailing `-->` of `<!--->` reads as
    // a terminator, while stripHtmlComments does not see one and deletes to the end of the string.
    assert.equal(dropOrphanedCommentOpener('text <!--->'), 'text ');
    const bounded = boundEnvPayload(`${'a'.repeat(120)}<!--->${'z'.repeat(400)}`, 200);
    assert.ok(sanitizeUserContent(bounded).includes('truncated: payload exceeded'));
  });

  test('scans past a terminated comment to find a later orphan', () => {
    // A backwards search from the end also works here, but not when the cut lands in the *second*
    // comment of a body whose first one is closed — the common shape for issue templates.
    assert.equal(dropOrphanedCommentOpener('intro <!-- tmpl --> body <!-- cut here'), 'intro <!-- tmpl --> body ');
  });

  test('keeps the truncation notice readable through the downstream sanitiser', () => {
    // sanitizeUserContent deletes from an orphaned `<!--` to end of string, taking the notice with
    // it — the payload would then look complete to the agent.
    const body = `${'漢'.repeat(34_108)}<!-- note -->${'漢'.repeat(5_000)}`;
    const bounded = boundEnvPayload(body);
    assert.ok(sanitizeUserContent(bounded).includes('truncated: payload exceeded'));
  });
});

describe('boundEnvPayload', () => {
  test('leaves any ASCII payload this layer can produce untouched', () => {
    // GitHub caps an issue body at 65 536 characters; in ASCII that is 65 536 bytes.
    const body = 'x'.repeat(65_536);
    assert.equal(boundEnvPayload(body), body);
  });

  test('bounds a multi-byte payload that would exceed the kernel limit', () => {
    // 49 000 CJK characters are 147 000 bytes — over Linux's 128 KiB per-entry limit, which kills
    // the exporting step with E2BIG after the trigger label has already been consumed.
    const cjk = '漢'.repeat(49_000);
    assert.ok(byteLength(cjk) > 128 * 1024);

    const bounded = boundEnvPayload(cjk);
    assert.ok(byteLength(bounded) <= MAX_ENV_PAYLOAD_BYTES);
    assert.ok(byteLength(bounded) < 128 * 1024);
    assert.ok(bounded.endsWith(TRUNCATION_NOTICE), 'the cut is visible, not silent');
  });

  test('a bounded payload stays well-formed', () => {
    const bounded = boundEnvPayload('😀'.repeat(40_000));
    assert.ok(bounded.isWellFormed());
    assert.ok(byteLength(bounded) <= MAX_ENV_PAYLOAD_BYTES);
  });

  test('returns a payload of exactly the budget untouched', () => {
    const exact = 'x'.repeat(MAX_ENV_PAYLOAD_BYTES);
    assert.equal(boundEnvPayload(exact), exact, 'the boundary is inclusive');
    assert.ok(byteLength(boundEnvPayload(`${exact}y`)) < byteLength(`${exact}y`), 'one byte over is cut');
  });

  test('honours a budget smaller than the notice instead of overshooting it', () => {
    assert.ok(byteLength(boundEnvPayload('x'.repeat(100), 10)) <= 10);
  });

  test('honours an explicit budget and tolerates empty input', () => {
    assert.ok(byteLength(boundEnvPayload('漢'.repeat(500), 200)) <= 200);
    assert.equal(boundEnvPayload(''), '');
    assert.equal(boundEnvPayload(null), '');
  });
});
