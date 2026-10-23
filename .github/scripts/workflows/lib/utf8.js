/**
 * UTF-8 byte-size helpers shared by factory intake runners.
 *
 * Issue bodies and comment history are transported to downstream steps through environment
 * variables (e.g. `ISSUE_BODY_EVENT`, `HUMAN_COMMENTS`). A single env-var string above Linux's
 * `MAX_ARG_STRLEN` (~128 KB) makes the next process launch fail with `E2BIG`, so these values must
 * be bounded by UTF-8 *bytes*, not JS string length (UTF-16 code units): a CJK/emoji-heavy string
 * is 3-4 bytes/char, so a code-unit bound can still blow the byte limit.
 */
'use strict';

/** UTF-8 byte length of a string. */
function utf8Len(str) {
  return Buffer.byteLength(str, 'utf8');
}

/**
 * Returns the longest prefix of `str` whose UTF-8 encoding is <= `maxBytes`, never splitting a
 * multi-byte character (so the result is always valid UTF-8, never emitting U+FFFD).
 *
 * @param {string} str
 * @param {number} maxBytes
 * @returns {string}
 */
function sliceToByteBudget(str, maxBytes) {
  if (maxBytes <= 0) return '';
  const full = Buffer.from(str, 'utf8');
  if (full.length <= maxBytes) return str;
  // The cut falls between byte maxBytes-1 (last kept) and byte maxBytes (first excluded). Only
  // back up if that cut splits a character — i.e. the first EXCLUDED byte is a UTF-8 continuation
  // byte (10xxxxxx). When it is a lead/ASCII byte, maxBytes already lands exactly on a character
  // boundary and the whole prefix is kept (no over-truncation of a character that fits).
  let end = maxBytes;
  if ((full[end] & 0xc0) === 0x80) {
    while (end > 0 && (full[end - 1] & 0xc0) === 0x80) end--;
    if (end > 0) end--; // drop the lead byte of the split character
  }
  return full.subarray(0, end).toString('utf8');
}

/**
 * Clamps `str` to at most `maxBytes` UTF-8 bytes, appending a short truncation marker when it had
 * to cut. The marker itself is kept within the budget.
 *
 * @param {string} str
 * @param {number} maxBytes
 * @returns {{ text: string, truncated: boolean }}
 */
function clampUtf8Bytes(str, maxBytes) {
  const s = typeof str === 'string' ? str : '';
  if (utf8Len(s) <= maxBytes) {
    return { text: s, truncated: false };
  }
  const marker = '\n\n[... truncated to fit the processing budget]\n';
  const markerBytes = utf8Len(marker);
  if (maxBytes <= markerBytes) {
    // No room for the marker; hard-slice to the budget so the result never exceeds maxBytes.
    return { text: sliceToByteBudget(s, maxBytes), truncated: true };
  }
  const body = sliceToByteBudget(s, maxBytes - markerBytes);
  return { text: body + marker, truncated: true };
}

if (typeof module !== 'undefined') {
  module.exports = { utf8Len, sliceToByteBudget, clampUtf8Bytes };
}
