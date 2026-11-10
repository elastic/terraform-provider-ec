/**
 * Locates and parses the research comment's pipeline-metadata JSON block.
 *
 * Ported from the Elastic Stack provider's research-factory gate (`gate.js`), which solves the
 * same problem: find the one `json` fence inside the one `<details><summary>🤖 Pipeline
 * metadata</summary>` envelope, without being fooled by markdown around it.
 *
 * Why this is not a regex over the body:
 * - the envelope is anchored on the `<summary>` text, never on the first bare `<details>`, so a
 *   collapsed "Pros and cons" block earlier in the comment cannot win;
 * - `<details>` and `<summary>` tags inside fenced code blocks are ignored, so an approach that
 *   quotes this very markup does not create a second envelope;
 * - `json` fences elsewhere in the body (an Elastic Cloud API response, for instance) are
 *   explicitly permitted — only fences inside the envelope count.
 *
 * `locateMetadata` returns the content range of that fence, which is what makes a structural
 * rewrite possible: the caller can replace exactly the JSON text and nothing else.
 */
'use strict';

const FENCE_LINE = /^ {0,3}(`{3,}|~{3,})([^`~\s]*)?(.*)$/;
const DETAILS_TAG = /<details(?:\s[^>]*)?>|<\/details\s*>/gi;
const SUMMARY_TEXT = /<summary(?:\s[^>]*)?>(.*?)<\/summary\s*>/i;

/**
 * Parses a line as a code-fence delimiter.
 *
 * @param {string} line
 * @returns {{char: string, length: number, info: string} | null}
 */
function fenceOf(line) {
  const m = FENCE_LINE.exec(line.replace(/\r?\n$/, ''));
  if (!m) {
    return null;
  }
  const info = `${m[2] || ''}${m[3] || ''}`;
  if (m[1][0] === '`' && info.includes('`')) {
    return null;
  }
  return { char: m[1][0], length: m[1].length, info: info.trim() };
}

/**
 * Locates the content range of the single json fence inside the single "Pipeline metadata"
 * `<details>` envelope. Fence-aware: tags inside code fences are ignored.
 *
 * @param {string} body
 * @returns {{start: number, end: number, text: string} | {error: string}}
 */
function locateMetadata(body) {
  const text = String(body ?? '');
  const stack = [];
  const envelopes = [];
  let fence = null;

  for (const match of text.matchAll(/[^\n]*\n|[^\n]+/g)) {
    const line = match[0];
    const index = match.index;
    const parsed = fenceOf(line);

    if (fence) {
      if (parsed && parsed.char === fence.char && parsed.length >= fence.length && parsed.info === '') {
        if (fence.json && fence.envelope) {
          const end = text[index - 1] === '\n' ? index - 1 : index;
          fence.envelope.fences.push({ start: fence.start, end: Math.max(fence.start, text[end - 1] === '\r' ? end - 1 : end) });
        }
        fence = null;
      }
      continue;
    }
    if (parsed) {
      fence = {
        ...parsed,
        json: parsed.info.toLowerCase() === 'json',
        start: index + line.length,
        envelope: [...stack].reverse().find((entry) => entry.meta),
      };
      continue;
    }

    const summary = SUMMARY_TEXT.exec(line);
    for (const tag of line.matchAll(DETAILS_TAG)) {
      if (tag[0][1] === '/') {
        const entry = stack.pop();
        if (entry && entry.meta) {
          envelopes.push(entry);
        }
      } else {
        stack.push({ meta: false, fences: [], summarySeen: false });
      }
    }
    if (summary && stack.length > 0) {
      const top = stack[stack.length - 1];
      if (!top.summarySeen) {
        top.summarySeen = true;
        top.meta = /pipeline metadata/i.test(summary[1]);
      }
    }
  }

  if (envelopes.length === 0) {
    return { error: 'Pipeline metadata <details> envelope not found in comment' };
  }
  if (envelopes.length > 1) {
    return { error: 'multiple Pipeline metadata <details> envelopes found in comment' };
  }
  const fences = envelopes[0].fences;
  if (fences.length !== 1) {
    return { error: `Pipeline metadata envelope must contain exactly one json block, found ${fences.length}` };
  }
  return { start: fences[0].start, end: fences[0].end, text: text.slice(fences[0].start, fences[0].end) };
}

/**
 * Walks the body line by line, yielding only the lines that are outside fenced code blocks.
 *
 * The same fence state machine `locateMetadata` runs, exposed so the comment handler counts
 * headings the way this module counts fences: a `#### Approach` quoted inside an ```` ```markdown ````
 * fence is an example, not an approach, and a `### References` inside one does not satisfy the
 * mandatory-section check.
 *
 * @param {string} body
 * @yields {{line: string, index: number}} The line including its trailing newline, and its offset.
 */
function* linesOutsideFences(body) {
  const text = String(body ?? '');
  let fence = null;

  for (const match of text.matchAll(/[^\n]*\n|[^\n]+/g)) {
    const line = match[0];
    const parsed = fenceOf(line);

    if (fence) {
      if (parsed && parsed.char === fence.char && parsed.length >= fence.length && parsed.info === '') {
        fence = null;
      }
      continue;
    }
    if (parsed) {
      fence = parsed;
      continue;
    }
    yield { line, index: match.index };
  }
}

/**
 * Locates and parses the pipeline metadata.
 *
 * @param {string} body
 * @returns {{metadata: object} | {error: string}}
 */
function extractMetadata(body) {
  const located = locateMetadata(body);
  if (located.error) {
    return { error: located.error };
  }
  try {
    return { metadata: JSON.parse(located.text) };
  } catch (err) {
    return { error: `Pipeline metadata JSON could not be parsed: ${err.message}` };
  }
}

module.exports = { fenceOf, locateMetadata, linesOutsideFences, extractMetadata };
