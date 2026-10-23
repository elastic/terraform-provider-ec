/**
 * HTML comment sanitisation, control/invisible-char removal, and research-comment lookup helpers.
 *
 * Strips HTML comments and the control and invisible characters listed on the helpers below.
 * Collapsed details blocks, link reference definitions, and image title or alt text are left
 * as plain text.
 */

/**
 * Removes all HTML comment sequences (<code>&lt;!--</code> through the next <code>--&gt;</code>).
 * If an opening sequence has no closing counterpart, everything from the opener to the end of
 * the string is removed.
 *
 * @param {string} text
 * @returns {string}
 */
function stripHtmlComments(text) {
  if (typeof text !== 'string') return '';
  return text.replace(/<!--[\s\S]*?(?:-->|$)/g, '');
}

/**
 * Removes non-printable ASCII control characters while preserving tab, newline,
 * and carriage return. Also strips Unicode line/paragraph separators.
 *
 * Stripped ASCII:  \x00-\x08, \x0B, \x0C, \x0E-\x1F, \x7F
 * Preserved:       \x09 (tab), \x0A (LF), \x0D (CR)
 * Unicode:         \u2028, \u2029
 *
 * @param {string} text
 * @returns {string}
 */
function stripControlChars(text) {
  if (typeof text !== 'string') return '';
  return text.replace(/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F\u2028\u2029]/g, '');
}

/**
 * Removes invisible Unicode characters that have no legitimate use in issue content:
 * zero-width spaces/joiners, bidirectional marks, word/function/invisible operators,
 * and the BOM (byte order mark).
 *
 * Ranges stripped:
 *   \u200B-\u200F  — zero-width space, non-joiner, joiner, LTR mark, RTL mark
 *   \u2060-\u2064  — word joiner, function application, invisible times/separator/plus
 *   \uFEFF           — BOM / zero-width no-break space
 *
 * @param {string} text
 * @returns {string}
 */
function stripInvisibleUnicode(text) {
  if (typeof text !== 'string') return '';
  return text.replace(/[\u200B-\u200F\u2060-\u2064\uFEFF]/g, '');
}

/**
 * Composed sanitisation pipeline. Runs all three filters in sequence:
 *
 *   input → stripHtmlComments → stripControlChars → stripInvisibleUnicode → output
 *
 * Idempotent except where a later filter reassembles a comment opener
 * (a null or zero-width character inside "<!--", or a nested "<!-" chain). See the residue tests.
 *
 * @param {string} text
 * @returns {string}
 */
function sanitizeUserContent(text) {
  if (typeof text !== 'string') return '';
  return stripInvisibleUnicode(stripControlChars(stripHtmlComments(text)));
}

/** Upper bound on sanitisation passes. Legitimate content reaches a fixed point on the first. */
const SANITIZE_FIXED_POINT_MAX_PASSES = 20;

/**
 * Sanitisation for a value that will be <strong>concatenated with others</strong> and sanitized
 * again downstream — a comment body joined into a multi-author blob, for example.
 *
 * {@link sanitizeUserContent} is deliberately a single pass and is not idempotent: removing an
 * invisible character or an inner comment can reassemble a literal <code>&lt;!--</code> in its
 * own output (<code>hello &lt;!-&lt;!-- --&gt;-</code> becomes <code>hello &lt;!--</code>). In one
 * standalone document that residue is inert text, but once it is concatenated the next
 * sanitisation pass treats it as an unterminated opener and deletes everything after it — one
 * comment can then erase every later comment and its attribution.
 *
 * Running to a fixed point removes that residue: <code>stripHtmlComments</code> strictly shortens
 * the string whenever it fires, so a surviving <code>&lt;!--</code> would mean the next pass
 * differs and the fixed point has not been reached. The pass bound keeps a crafted input from
 * running long; if it is ever hit, every <code>&lt;</code> is dropped so the result still cannot
 * open a comment.
 *
 * @param {string} text
 * @returns {string} Sanitized text guaranteed to contain no <code>&lt;!--</code>.
 */
function sanitizeForConcatenation(text) {
  let current = typeof text === 'string' ? text : '';
  for (let pass = 0; pass < SANITIZE_FIXED_POINT_MAX_PASSES; pass++) {
    const next = sanitizeUserContent(current);
    if (next === current) break;
    current = next;
  }
  return current.includes('<!--') ? current.replace(/</g, '') : current;
}

/**
 * Finds the most recently created matching research comment written by
 * <code>github-actions[bot]</code> whose body starts with <code>marker</code>.
 *
 * @param {Array<{author: string, body: string}>} comments Ordered oldest-first.
 * @param {string} marker
 * @returns {{author: string, body: string} | null}
 */
function findResearchComment(comments, marker) {
  if (!Array.isArray(comments)) {
    return null;
  }
  const matches = comments.filter(
    (c) =>
      c != null &&
      typeof c.body === 'string' &&
      (c.author ?? c.user?.login) === 'github-actions[bot]' &&
      c.body.trimStart().startsWith(marker),
  );
  return matches.length > 0 ? matches[matches.length - 1] : null;
}

if (typeof module !== 'undefined') {
  module.exports = {
    stripHtmlComments,
    stripControlChars,
    stripInvisibleUnicode,
    sanitizeUserContent,
    sanitizeForConcatenation,
    SANITIZE_FIXED_POINT_MAX_PASSES,
    findResearchComment,
  };
}
