/**
 * Census of the research-factory grounding inputs, for the threat-detection job.
 *
 * A small per-file summary (byte counts and marker counts) staged beside the grounding files. Counts
 * are computed over the raw payload as received by the context writer, before sanitisation.
 *
 * Pure and deterministic: no I/O, no `Date`. Depends only on `env-payload.js` for byte handling.
 */
'use strict';

const { byteLength, sliceToBytes } = require('./env-payload.js');

/** Hard ceiling on the rendered `census.md`. */
const CENSUS_MAX_BYTES = 8 * 1024;

/** Appended when the rendered census has to be cut to honour `CENSUS_MAX_BYTES`. */
const CENSUS_TRUNCATION_NOTICE = '\n\n[... census truncated: exceeded CENSUS_MAX_BYTES]';

/**
 * Builds a counter for "these words appear in this order within one line". Counts one per line on
 * which `steps[0]` matches, then `steps[1]` after it, and so on.
 *
 * Runs in linear time. Taking the earliest match at each step leaves the most room for the steps
 * after it, so a line counts if and only if some in-order match exists on it.
 *
 * @param {...RegExp} steps non-global, case-insensitive alternations of literal words
 * @returns {(text: string) => number}
 */
function inOrderOnOneLine(...steps) {
  if (steps.some((s) => s.global || s.sticky)) {
    throw new Error('inOrderOnOneLine: steps must be non-global, non-sticky regexes');
  }
  return function countLines(text) {
    let total = 0;
    for (const line of text.split('\n')) {
      let rest = line;
      let found = true;
      for (const step of steps) {
        const m = step.exec(rest); // non-global regex: exec carries no lastIndex state
        if (!m) {
          found = false;
          break;
        }
        rest = rest.slice(m.index + m[0].length);
      }
      if (found) total += 1;
    }
    return total;
  };
}

/**
 * Injection-marker definitions — the single source of truth. Tests iterate this array, so a new
 * marker is added here and nowhere else, and the census table columns follow this order.
 *
 * Each entry is `{ name, pattern, scope }`:
 * - `pattern` is a case-insensitive `g` regex, a counting function `(text) => number`, or an
 *   array of those; the count is the sum over all of them. Every pattern must run in linear time
 *   on large inputs; use `inOrderOnOneLine` for "words in order on one line".
 * - `scope` is `'all'`, or a list of source `name`s the marker applies to; out-of-scope sources
 *   report `'n/a'`, never `0`.
 *
 * Counting uses `String.prototype.match`, which resets `lastIndex` for a `g` regex, so these shared
 * regex objects can be reused across sources without the `lastIndex`-carry mis-count that
 * `test()`/`exec()` would cause.
 */
const INJECTION_MARKERS = [
  {
    name: 'html_comment_openers',
    scope: 'all',
    // Literal `<!--`. Issue-template reports carry a baseline of these; the detector prompt says so.
    pattern: /<!--/gi,
  },
  {
    name: 'fence_lines',
    scope: 'all',
    // A line opening or closing a code fence; one fenced block counts as 2.
    pattern: /^[ \t]{0,3}(?:```|~~~)/gim,
  },
  {
    name: 'role_markers',
    scope: 'all',
    pattern: [
      // A line led by a chat role label, allowing Markdown decoration before it.
      /^[#*> \t]*(?:system|assistant|user|human|developer):/gim,
      // Chat-template control tokens.
      /<\|im_start\|>|<\|system\|>|\[INST\]|<<SYS>>/gi,
      // A heading that names a role or an instruction block. `### System:` is already counted by
      // the role-label pattern above, so the heading pattern skips it to avoid counting it twice.
      /^#{1,6}[ \t]*(?:system\b(?!:)|instructions?\b)/gim,
    ],
  },
  {
    name: 'agent_directives',
    scope: 'all',
    // Heuristic, with a small non-zero baseline on ordinary prose.
    pattern: [
      // ignore/disregard/... then previous/prior/... then instruction/prompt/... on one line.
      inOrderOnOneLine(
        /ignore|disregard|forget|override/i,
        /previous|prior|above|earlier|all|your/i,
        /instruction|prompt|rule|guideline|context/i,
      ),
      /you (?:are|must|will) now (?:act|behave|respond|answer|ignore|follow)/gi,
      /your new (?:instructions?|task|role|persona)/gi,
      /(?:system|developer) (?:prompt|message)/gi,
      /do not (?:tell|mention|reveal|disclose)/gi,
    ],
  },
  {
    name: 'workflow_steering',
    // This workflow's own control vocabulary. A signal to weigh, not a finding: maintainers discuss
    // the factory in issue bodies and comments. Excluded from the prior research comment, which
    // always contains `ready-for-change-factory`.
    scope: ['issue_body.md', 'issue_comments.md'],
    pattern: /update_research_comment|ready-for-change-factory|gate\.outcome|research-critic|critic-rubric|safe[_ -]?outputs?|phase-research/gi,
  },
  {
    name: 'template_macros',
    scope: 'all',
    pattern: /\{\{[#/]/gi,
  },
  {
    name: 'placeholder_tokens',
    scope: 'all',
    pattern: /__GH_AW_/gi,
  },
  {
    name: 'invisible_chars',
    scope: 'all',
    pattern: new RegExp('[' + [[0x200b, 0x200f], [0x2028, 0x202e], [0x2060, 0x2064], [0xfeff, 0xfeff]].map(function (r) { return String.fromCharCode(r[0]) + '-' + String.fromCharCode(r[1]); }).join('') + ']', 'gi'),
  },
  {
    name: 'system_tags',
    scope: 'all',
    // Weak heuristic.
    pattern: /<\/?system>/gi,
  },
];

/**
 * Counts matches of one marker's pattern(s) in `text`.
 *
 * @param {string} text
 * @param {RegExp|RegExp[]} pattern
 * @returns {number}
 */
function matchCount(text, pattern) {
  const value = String(text ?? '');
  const patterns = Array.isArray(pattern) ? pattern : [pattern];
  let total = 0;
  for (const p of patterns) {
    if (typeof p === 'function') {
      total += p(value);
      continue;
    }
    const matches = value.match(p);
    total += matches ? matches.length : 0;
  }
  return total;
}

/**
 * Counts every marker in `text`, attributing the counts to `sourceName`. Markers out of scope for
 * the source report `'n/a'`.
 *
 * @param {string} text
 * @param {string} sourceName
 * @returns {{ [name: string]: number | 'n/a' }}
 */
function countMarkers(text, sourceName) {
  const counts = {};
  for (const marker of INJECTION_MARKERS) {
    const inScope = marker.scope === 'all' || marker.scope.includes(sourceName);
    counts[marker.name] = inScope ? matchCount(text, marker.pattern) : 'n/a';
  }
  return counts;
}

/**
 * Bounds the rendered census to `CENSUS_MAX_BYTES`, reserving room for the notice first (as
 * `boundEnvPayload` does) so the ceiling is never exceeded and no code point is split. Exported so
 * the test can exercise the guard directly.
 *
 * @param {string} text
 * @returns {string}
 */
function boundCensus(text) {
  const value = String(text ?? '');
  if (byteLength(value) <= CENSUS_MAX_BYTES) {
    return value;
  }
  const room = CENSUS_MAX_BYTES - byteLength(CENSUS_TRUNCATION_NOTICE);
  return `${sliceToBytes(value, room)}${CENSUS_TRUNCATION_NOTICE}`;
}

/**
 * Renders `census.md` (Markdown table; the exact shape is pinned by the tests). Pure and
 * deterministic.
 *
 * @param {{ sources: Array<{ name: string, raw?: string, sanitized?: string }> }} args
 *   `sources` in fixed order: `issue_body.md`, `issue_comments.md`, `prior_research_comment.md`.
 *   `raw` is the pre-sanitisation payload; `sanitized` is what was written to disk. Either may be
 *   empty or undefined.
 * @returns {string} bounded to `CENSUS_MAX_BYTES`, ending with exactly one newline.
 */
function renderCensus({ sources }) {
  const markerNames = INJECTION_MARKERS.map((m) => m.name);

  const headerCols = ['file', 'raw bytes', 'sanitized bytes', ...markerNames];
  const header = `| ${headerCols.join(' | ')} |`;

  const sepCols = ['---', '---:', '---:', ...markerNames.map(() => '---:')];
  const separator = `| ${sepCols.join(' | ')} |`;

  const rows = sources.map((source) => {
    const counts = countMarkers(source.raw, source.name);
    const cells = [
      source.name,
      String(byteLength(source.raw)),
      String(byteLength(source.sanitized)),
      ...markerNames.map((name) => String(counts[name])),
    ];
    return `| ${cells.join(' | ')} |`;
  });

  const note = [
    'Counts are over the payload as received by the context writer (after the upstream byte clamps,',
    'before sanitisation); the files beside this one are the SANITISED copies the agent read.',
    '`raw` > `sanitized` means sanitisation removed something (HTML comments, control or invisible characters).',
  ].join('\n');

  const body = [
    '# Grounding-input census (research-factory)',
    '',
    note,
    '',
    header,
    separator,
    ...rows,
  ].join('\n') + '\n';

  return boundCensus(body);
}

module.exports = {
  CENSUS_MAX_BYTES,
  CENSUS_TRUNCATION_NOTICE,
  INJECTION_MARKERS,
  inOrderOnOneLine,
  countMarkers,
  boundCensus,
  renderCensus,
};
