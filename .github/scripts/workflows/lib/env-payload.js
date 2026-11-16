/**
 * Byte-bounded handoff for the large, user-controlled payloads that pre-activation passes between
 * steps as environment variables (issue body, comment history, prior research comment).
 *
 * Linux caps a single `execve` argument or environment entry at `MAX_ARG_STRLEN` = 128 KiB. The
 * budgets elsewhere in this layer count *characters*, which is the same thing only for ASCII: 49 000
 * CJK characters occupy 147 000 bytes and the step that exports them dies with `E2BIG`
 * ("Argument list too long"). That failure lands in `normalize_context` — after `remove_trigger_label`
 * has already consumed the trigger label — so the run cannot be retried without a human re-applying
 * it. GitHub's own 65 536-character issue-body limit does not help: in a multi-byte script it is
 * three times over the kernel's limit.
 *
 * The cap is deliberately well above any ASCII payload this layer can produce, so it only ever bites
 * on genuinely multi-byte content.
 */
'use strict';

/** 100 KiB, leaving ~28 KiB of head-room under the kernel's 128 KiB per-entry limit. */
const MAX_ENV_PAYLOAD_BYTES = 100 * 1024;

const TRUNCATION_NOTICE = '\n\n[... truncated: payload exceeded the environment-variable byte budget]';
/**
 * Appended when the payload is cut at an unterminated `<!--`.
 *
 * `write-context-files.js` runs `sanitizeUserContent` over every payload, and its
 * `stripHtmlComments` deletes from an orphaned opener to the end of the *string*. An issue body is
 * raw user input and this repository's own bug-report template carries 27 HTML comments, so a
 * reporter deleting one `-->` while filling it in is ordinary: measured on that template, the body
 * the agent received went from 782 characters to 559 — losing the whole environment-and-version
 * section — with nothing to show it had happened. (782, not the template's 2 413: the terminated
 * comments are stripped either way. The loss is the content *after* the orphan.)
 */
const OPENER_NOTICE = '\n\n[... truncated: the rest followed an unterminated HTML comment]';

/**
 * UTF-8 byte length of a string.
 *
 * @param {string} text
 * @returns {number}
 */
function byteLength(text) {
  return Buffer.byteLength(String(text ?? ''), 'utf8');
}

/**
 * Truncates to a UTF-8 byte budget, never splitting a code point.
 *
 * @param {string} text
 * @param {number} maxBytes
 * @returns {string}
 */
function sliceToBytes(text, maxBytes) {
  const value = String(text ?? '');
  if (maxBytes <= 0) {
    return '';
  }
  if (byteLength(value) <= maxBytes) {
    return value;
  }
  let out = '';
  let used = 0;
  // Iterating the string yields whole code points, so a surrogate pair is never cut in half.
  for (const char of value) {
    const size = byteLength(char);
    if (used + size > maxBytes) {
      break;
    }
    out += char;
    used += size;
  }
  return out;
}

/**
 * Bounds a payload for safe export as an environment variable, appending a visible notice when it
 * had to be cut — a silently shortened issue body reads as a complete one to the agent.
 *
 * @param {string} text
 * @param {number} [maxBytes]
 * @returns {string}
 */
function boundEnvPayload(text, maxBytes = MAX_ENV_PAYLOAD_BYTES) {
  const raw = String(text ?? '');

  // Always, not only when the budget truncates: an orphaned opener anywhere in the payload makes
  // the downstream sanitiser delete from there to the end of the file, and a payload well under
  // budget never reached the truncation branch at all.
  const whole = dropOrphanedCommentOpener(raw);
  const value = whole === raw ? raw : `${whole}${OPENER_NOTICE}`;

  if (byteLength(value) <= maxBytes) {
    return value;
  }
  if (maxBytes <= byteLength(TRUNCATION_NOTICE)) {
    // Smaller than the notice itself: honour the budget rather than overshoot it announcing that
    // we cut. Unreachable from the production call sites, which all use the 100 KiB default — but
    // still drop an opener the slice orphans, so this branch cannot be the one place that hands
    // the sanitiser a payload it will silently eat.
    return dropOrphanedCommentOpener(sliceToBytes(value, maxBytes));
  }
  const room = maxBytes - byteLength(TRUNCATION_NOTICE);
  return `${dropOrphanedCommentOpener(sliceToBytes(value, room))}${TRUNCATION_NOTICE}`;
}

/**
 * Removes a trailing HTML comment opener left unterminated by the cut.
 *
 * These payloads are sanitised downstream by `sanitizeUserContent`, whose `stripHtmlComments` is
 * `/<!--[\s\S]*?(?:-->|$)/g` — the `|$` deletes from an orphaned opener to the end of the string.
 * Cutting mid-comment therefore swallows the truncation notice as well, and the agent reads a
 * silently shortened payload: exactly what the notice exists to prevent.
 *
 * @param {string} text
 * @returns {string}
 */
function dropOrphanedCommentOpener(text) {
  // Walks the string the way `stripHtmlComments` does rather than approximating it. Two shapes
  // defeat the shorter version: `<!-->` (the closer overlaps the opener, so a backwards search
  // thinks it is terminated while the regex, resuming after the 4-character opener, does not), and
  // a terminated comment sitting to the left of the orphaned one.
  let opener = text.indexOf('<!--');
  while (opener !== -1) {
    const closer = text.indexOf('-->', opener + 4);
    if (closer === -1) {
      return text.slice(0, opener);
    }
    opener = text.indexOf('<!--', closer + 3);
  }
  return text;
}

if (typeof module !== 'undefined') {
  module.exports = {
    MAX_ENV_PAYLOAD_BYTES,
    TRUNCATION_NOTICE,
    OPENER_NOTICE,
    byteLength,
    sliceToBytes,
    boundEnvPayload,
    dropOrphanedCommentOpener,
  };
}
