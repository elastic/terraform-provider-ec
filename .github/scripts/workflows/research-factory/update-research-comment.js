/**
 * Safe-output handler: publishes the research draft as the issue's sticky comment.
 *
 * The agent's draft reaches this job as an artifact, not as a string input (the safe-output
 * gateway caps string inputs at 10 KiB). The job takes no inputs at all: the resolved draft path
 * is the whole contract.
 *
 * The pipeline is deliberately **decide -> apply -> re-assert**:
 *
 *   resolve -> fail closed if missing/empty -> reject (never mutate) a body carrying `<!--`
 *   -> transform length-preservingly -> prepend the marker -> evaluate every check and collect a
 *   disjoint edit set -> abort on any fatal -> apply edits in descending offset order
 *   -> re-run the fatal assertions and re-parse the metadata on the exact published string
 *   -> publish.
 *
 * Two rules are load-bearing and easy to "simplify" into bugs:
 *
 * - **Reject, do not mutate, a body containing `<!--`.** `sanitizeUserContent` is an *input*
 *   sanitiser: its `stripHtmlComments` is `/<!--[\s\S]*?(?:-->|$)/g`, and the `|$` deletes from an
 *   unterminated `<!--` to the end of the string. Measured on a realistic draft with one stray
 *   `<!--` inside an `hcl` fence: 337 chars -> 125, with `### Recommendation`, `### References`,
 *   the `<details>` block and `schema_version` all gone and no error raised. So this handler only
 *   ever applies the two length-preserving filters, and refuses a body it would have to cut.
 * - **The version drop is structural, never textual.** Dropping `"schema_version"` with a string
 *   edit can leave a trailing comma, and "metadata parses" is only a warn row -- the handler would
 *   publish JSON it broke itself while warning into a log nobody reads. The fence content is
 *   replaced with `JSON.stringify(metaWithoutSchemaVersion, null, 2)` instead, and the result is
 *   re-parsed before anything is published. The drop is skipped, with a warning, in the two cases
 *   where applying it would cost the run — re-serialising would introduce an HTML comment opener,
 *   or would breach the comment size limit — so a published comment can still carry a
 *   `schema_version` it does not meet, which is the cheaper of the two failures.
 *
 * Severity (see the plan's table): structure and size are fatal; metadata problems publish; a
 * `### Quality gate` section that disagrees with the machine-readable gate publishes with the
 * outcome forced and an annotation, because the human reading that section *is* a consumer.
 */
'use strict';

const fs = require('fs');
const path = require('path');

const { stripControlChars, stripInvisibleUnicode, findResearchComment } = require('../lib/sanitize-context.js');
const { positiveIssueNumber } = require('../lib/issue-number.js');
const { RESEARCH_COMMENT_MARKER } = require('../lib/intake/research-factory-constants.js');
const { locateMetadata, extractMetadata, linesOutsideFences } = require('./locate-metadata.js');
const { validateMetadata, SCHEMA_VERSION, CHECKLIST_ITEMS } = require('./validate-metadata.js');

/** Exactly what the safe job's `artifacts:` entry names. */
const RESEARCH_DIR = '/tmp/gh-aw/agent/research/';
const DRAFT_FINAL = 'draft-final.md';
/** upload-artifact strips this common ancestor from the uploaded paths. */
const ARTIFACT_ROOT = '/tmp/gh-aw/';

/** GitHub's own ceiling is 65 536; the comment contract reserves the rest for the marker. */
const MAX_COMMENT_CHARS = 60000;

/**
 * Anchors ported from the upstream research-factory gate (`gate.js:17-18`).
 *
 * Matched per line against lines outside code fences, never against the raw section: a decoy
 * `**Outcome:** \`research-needs-human\`` inside a ```text fence would otherwise be read as the
 * prose outcome, and the real, disagreeing one below it would publish unannotated — the one check
 * the severity table exists to guarantee for a human reader.
 *
 * Anchored at the start of the line on purpose: unanchored it would also match prose such as
 * "see the **Outcome:** line below", which since the stray-line check would force an outcome on a
 * draft that never stated one.
 */
const OUTCOME_LINE = /^\*\*Outcome:\*\* .*$/;
const QUALITY_GATE_HEADING = 'Quality gate';
const APPROACHES_HEADING = 'Approaches considered';
const RECOMMENDATION_HEADING = 'Recommendation';
const REFERENCES_HEADING = 'References';

const FORCED_OUTCOME = 'research-needs-human';
const READY_OUTCOME = 'ready-for-change-factory';
const MIN_APPROACHES = 2;
/** The rubric's convergence threshold, mirrored here only to judge a readiness claim. */
const SCORE_THRESHOLD = 85;
/**
 * Longest agent-authored fragment quoted inside an annotation. Both the forced outcome line and
 * the ⚠️ line quote the prose outcome and the declared gate outcome, so without a clamp the
 * handler's own edit grows the body by an unbounded amount — measured at +8 393 characters from a
 * draft with a 2 000-character outcome token — and the 60 000 re-assert then discards a run that
 * was within the limit before the edit. Clamped, the annotation is bounded per live outcome line —
 * it rewrites each one, so the total scales with how many there are, and the size filter keeps
 * that under the ceiling.
 */
const MAX_QUOTED_FRAGMENT = 80;

/**
 * Renders an agent-authored fragment safe to quote inside the annotation.
 *
 * The annotation is the one place the handler copies draft-controlled text into new markdown, so
 * the fragment is neutralised three ways, each closing a reproduced failure:
 * - `<` and `>` are removed, so a `gate.outcome` of `"<!-- x"` cannot reintroduce the HTML comment
 *   opener the body was just checked for (which would fail the re-assert and discard the run), and
 *   `"x</summary></details>…<details><summary>Pipeline metadata</summary>"` cannot forge a second
 *   metadata envelope (which would break the metadata re-parse and discard the run);
 * - backticks are removed, so the fragment cannot escape the inline code span it is quoted in;
 * - every whitespace run collapses to one space, so a newline cannot inject a heading or list item.
 * Then it is clamped, because both the forced outcome line and the ⚠️ line quote it: unclamped, a
 * 2 000-character outcome token grew the published body by 8 393 characters. Clamped, the
 * annotation is bounded per live outcome line (it rewrites each one), not by a single constant —
 * the size filter below is what keeps the total under the ceiling.
 *
 * @param {string} text
 * @returns {string}
 */
function quoteFragment(text) {
  const flattened = String(text)
    .replace(/[`<>]/g, '')
    .replace(/\s+/g, ' ')
    .trim();
  if (flattened === '') {
    return '(empty)';
  }
  // Clamped by code point, not UTF-16 unit: slicing mid-surrogate would put a lone half into the
  // comment body, which octokit then serialises as an unpaired escape.
  const points = [...flattened];
  return points.length <= MAX_QUOTED_FRAGMENT ? flattened : `${points.slice(0, MAX_QUOTED_FRAGMENT).join('')}…`;
}

const HEADING_LINE = /^ {0,3}(#{1,6})[^\S\n]+(.*?)[^\S\n]*$/;

/** A plain JSON object: not null, not an array. Mirrors the predicate in validate-metadata.js. */
function isObject(value) {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

/**
 * Fence-aware heading scan.
 *
 * @param {string} body
 * @returns {Array<{level: number, text: string, start: number, end: number}>}
 */
function headings(body) {
  const found = [];
  for (const { line, index } of linesOutsideFences(body)) {
    const match = HEADING_LINE.exec(line.replace(/\r?\n$/, ''));
    if (match) {
      found.push({ level: match[1].length, text: match[2].trim(), start: index, end: index + line.length });
    }
  }
  return found;
}

/**
 * Range of a section: from its heading to the next heading at the same level or above.
 *
 * @param {string} body
 * @param {string} text Heading text to find at level 3.
 * @returns {{start: number, end: number} | null}
 */
function sectionRange(body, text) {
  const all = headings(body);
  const index = all.findIndex((h) => h.level === 3 && h.text === text);
  if (index === -1) {
    return null;
  }
  const next = all.slice(index + 1).find((h) => h.level <= 3);
  return { start: all[index].start, end: next ? next.start : body.length };
}

/**
 * Counts the `####` approach headings inside `### Approaches considered`, ignoring fenced regions.
 *
 * @param {string} body
 * @returns {{count: number, sectionFound: boolean}}
 */
function countApproaches(body) {
  const range = sectionRange(body, APPROACHES_HEADING);
  if (!range) {
    return { count: 0, sectionFound: false };
  }
  const count = headings(body).filter((h) => h.level === 4 && h.start >= range.start && h.start < range.end).length;
  return { count, sectionFound: true };
}

/**
 * 1-based line number of an offset, for an error message a human can act on.
 *
 * @param {string} text
 * @param {number} offset
 * @returns {number}
 */
function lineNumberOf(text, offset) {
  return text.slice(0, offset).split('\n').length;
}

/**
 * The fatal contract: everything whose absence makes the published comment wrong rather than
 * merely imperfect. Re-run verbatim on the final string, which is the point of keeping it here.
 *
 * @param {string} body The full comment body, marker included.
 * @returns {string[]} Fatal messages; empty when the body satisfies the contract.
 */
function fatalContractViolations(body) {
  const fatals = [];

  if (body.length > MAX_COMMENT_CHARS) {
    fatals.push(`comment body is ${body.length} characters, over the ${MAX_COMMENT_CHARS} limit`);
  }

  const { count, sectionFound } = countApproaches(body);
  if (!sectionFound) {
    fatals.push(`missing mandatory section: ### ${APPROACHES_HEADING}`);
  } else if (count < MIN_APPROACHES) {
    fatals.push(`### ${APPROACHES_HEADING} has ${count} #### approach heading(s) outside code fences, needs at least ${MIN_APPROACHES}`);
  }

  for (const heading of [RECOMMENDATION_HEADING, REFERENCES_HEADING]) {
    if (!sectionRange(body, heading)) {
      fatals.push(`missing mandatory section: ### ${heading}`);
    }
  }

  return fatals;
}

/**
 * Rewrites the `### Quality gate` section: forces the prose outcome and appends the annotation.
 * One edit over one range, so it cannot interleave with the metadata rewrite.
 *
 * @param {string} section
 * @param {string[]} causes
 * @returns {string}
 */
function annotateQualityGate(section, causes, { compact = false, minimal = false } = {}) {
  // The minimal form carries no causes at all — just the corrected verdict. It is 35 characters and
  // usually *shorter* than the line it replaces, so it fits wherever the draft itself fits. Without
  // it, a draft near the size ceiling made both richer forms unaffordable, the edit was skipped,
  // and a readiness claim the handler had positively judged unearned published with nothing in the
  // body to say so.
  const forced = minimal
    ? `**Outcome:** \`${FORCED_OUTCOME}\``
    : `**Outcome:** \`${FORCED_OUTCOME}\` - forced by the comment handler: ${causes.join(' ')}`;
  const annotation = `> ⚠️ ${causes.join(' ')} Treat the outcome above as unverified.`;
  const lines = outcomeLines(section);

  let rewritten;
  if (lines.length === 0) {
    // No outcome line to correct: state the forced outcome right under the heading.
    const firstBreak = section.indexOf('\n');
    rewritten = firstBreak === -1 ? `${section}\n\n${forced}\n` : `${section.slice(0, firstBreak + 1)}\n${forced}\n${section.slice(firstBreak + 1)}`;
  } else {
    // Every live outcome line is replaced, not just the first: leaving a second one behind would
    // publish a section whose forced outcome is contradicted a few lines further down. Applied by
    // descending offset, for the same reason the edit set is.
    rewritten = section;
    for (const line of [...lines].sort((a, b) => b.start - a.start)) {
      rewritten = rewritten.slice(0, line.start) + forced + rewritten.slice(line.end);
    }
  }

  // The compact form carries the correction without the appended block: near length-neutral per
  // live outcome line (it swaps each one for a forced line of similar length), so it still fits
  // where the full annotation would breach the ceiling. With many live outcome lines it is not
  // bounded by a constant, and the size filter then degrades it to a skip rather than a lost run.
  return compact || minimal ? rewritten : `${rewritten.replace(/\s*$/, '')}\n\n${annotation}\n\n`;
}

/**
 * Reasons the metadata's own fields contradict a `ready-for-change-factory` claim.
 *
 * This is not the upstream deterministic gate — that exists to defend a machine-readable label this
 * workflow does not apply, and porting it was rejected. This is narrower: the `### Quality gate`
 * prose is the one thing a human reads to decide whether the research is ready, so a draft claiming
 * readiness while its own gate object records an unavailable critic, zero rounds, a failed checklist
 * item or a blocking open question is claiming something it has not earned. Treated exactly like a
 * prose/JSON disagreement: publish, force the prose outcome, annotate.
 *
 * @param {object} metadata
 * @returns {string[]} Empty when the claim holds, or when no readiness is claimed.
 */
function unearnedReadiness(metadata) {
  const gate = metadata?.gate;
  if (!isObject(gate) || gate.outcome !== READY_OUTCOME) {
    return [];
  }

  // Every predicate makes the claim *earn* its pass rather than looking for a disproof: a field
  // that is absent, null or written with the wrong JSON type is not evidence of readiness, and
  // treating it as such let a one-character change (`0` -> `"0"`) defeat the whole check. The
  // schema validator already rejects those shapes, but a 1.1 violation only drops the claim — it
  // does not annotate — so the comment would still publish an unqualified readiness claim.
  const reasons = [];

  const checklist = isObject(gate.checklist) ? gate.checklist : null;
  const unproven = checklist === null ? CHECKLIST_ITEMS : CHECKLIST_ITEMS.filter((item) => checklist[item] !== true);
  if (unproven.length > 0) {
    reasons.push(`checklist item(s) ${unproven.join(', ')} did not pass`);
  }
  if (!isObject(gate.critic) || gate.critic.status !== 'ok') {
    const status = isObject(gate.critic) ? gate.critic.status : undefined;
    // Agent-authored text: quoted through the same neutraliser as every other fragment the
    // annotation embeds. Unsanitised, a multi-line status reproduced a live `**Outcome:**` line —
    // republishing the very claim this check exists to force down — and a long one grew the
    // annotation past the size ceiling so the edit was skipped entirely.
    reasons.push(`the critic did not report status \`ok\` (got \`${quoteFragment(status === undefined ? 'nothing' : status)}\`)`);
  }
  if (gate.converged !== true) {
    reasons.push('the critique loop is not recorded as converged');
  }
  if (!Number.isInteger(gate.rounds) || gate.rounds < 1) {
    reasons.push('no critique round is recorded as completed');
  }
  if (typeof gate.score !== 'number' || !Number.isFinite(gate.score) || gate.score < SCORE_THRESHOLD) {
    reasons.push(`the final score is not recorded at or above the threshold of ${SCORE_THRESHOLD}`);
  }
  const questions = metadata.open_questions;
  if (questions !== undefined && !Array.isArray(questions)) {
    // Present but not a list: the schema validator catches the type, but a 1.1 violation only drops
    // the claim — it never annotates — so treating this like "absent" would publish the readiness
    // claim unqualified, which is the whole failure this function exists to stop.
    reasons.push('the open questions are not recorded as a list');
  } else if (Array.isArray(questions)) {
    const blocking = questions.filter((question) => !isObject(question) || question.blocking !== false).length;
    if (blocking > 0) {
      reasons.push(`${blocking} open question(s) are not recorded as non-blocking`);
    }
  }
  return reasons;
}

/**
 * Every `**Outcome:**` line in a section that is not inside a code fence, with its range.
 *
 * @param {string} section
 * @returns {Array<{text: string, start: number, end: number}>}
 */
function outcomeLines(section) {
  const found = [];
  for (const { line, index } of linesOutsideFences(section)) {
    const text = line.replace(/\r?\n$/, '');
    if (OUTCOME_LINE.test(text)) {
      found.push({ text, start: index, end: index + text.length });
    }
  }
  return found;
}

/**
 * Reads the outcome token out of a `**Outcome:**` line.
 *
 * @param {string} line
 * @returns {string | null}
 */
function quotedToken(line) {
  const quoted = /`([^`]+)`/.exec(line);
  return quoted ? quoted[1].trim() : null;
}

/**
 * Resolves the agent's final draft.
 *
 * The agent writes `/tmp/gh-aw/agent/research/draft-final.md`; `upload-artifact` strips the
 * `/tmp/gh-aw/` common ancestor, and the safe job downloads the artifact into
 * `${RUNNER_TEMP}/gh-aw/safe-jobs/`, so the same file arrives here as
 * `<safe-jobs>/agent/research/draft-final.md`. Both forms are tried, then a bounded search of the
 * download directory, so a gh-aw layout change degrades into a slower lookup rather than a
 * discarded run.
 *
 * @param {NodeJS.ProcessEnv} env
 * @returns {{path: string, candidates: string[]} | {error: string, candidates: string[]}}
 */
function resolveDraftPath(env) {
  const safeJobsDir = env.GH_AW_AGENT_OUTPUT
    ? path.dirname(env.GH_AW_AGENT_OUTPUT)
    : path.join(env.RUNNER_TEMP || '/tmp', 'gh-aw', 'safe-jobs');
  const relative = path.join(path.relative(ARTIFACT_ROOT, RESEARCH_DIR), DRAFT_FINAL);

  const candidates = [path.join(safeJobsDir, relative), path.join(RESEARCH_DIR, DRAFT_FINAL)];

  for (const candidate of candidates) {
    if (fs.existsSync(candidate) && fs.statSync(candidate).isFile()) {
      return { path: candidate, candidates };
    }
  }

  const found = findDraftUnder(safeJobsDir, 6);
  if (found.length > 0) {
    return { path: found.sort()[0], candidates, fallback: found };
  }

  return { error: `no ${DRAFT_FINAL} in the agent artifact`, candidates };
}

/**
 * Bounded recursive search for the final draft under a downloaded artifact.
 *
 * @param {string} dir
 * @param {number} depth
 * @returns {string[]}
 */
function findDraftUnder(dir, depth) {
  if (depth < 0) {
    return [];
  }
  let dirEntries;
  try {
    dirEntries = fs.readdirSync(dir, { withFileTypes: true });
  } catch {
    return [];
  }
  const matches = [];
  for (const entry of dirEntries) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      matches.push(...findDraftUnder(full, depth - 1));
    } else if (entry.isFile() && entry.name === DRAFT_FINAL) {
      matches.push(full);
    }
  }
  return matches;
}

/**
 * Re-runs every fatal assertion on the exact string that will be published.
 *
 * The decide phase already evaluated all of these on the pre-edit body; this is the backstop that
 * catches the handler breaking its own output, which is where five of the plan's eleven review
 * rounds found defects. It is a last resort by design — the applicability filter is supposed to
 * keep an unsafe edit out of the set in the first place — so it is exported and tested directly
 * rather than through an input that is no longer supposed to reach it.
 *
 * @param {string} published
 * @param {{metadataWasParsed: boolean, metadataDropped: boolean}} state
 * @returns {string[]} Fatal messages; empty when the published string holds up.
 */
function reassertPublished(published, { metadataWasParsed, metadataDropped }) {
  const fatals = fatalContractViolations(published);

  const markerCount = published.split('<!--').length - 1;
  if (markerCount !== 1 || !published.startsWith(RESEARCH_COMMENT_MARKER)) {
    fatals.push(`the body carries ${markerCount} HTML comment opener(s); only the leading marker is allowed`);
  }

  if (metadataWasParsed) {
    const reparsed = extractMetadata(published);
    if (reparsed.error) {
      fatals.push(`the pipeline metadata no longer parses: ${reparsed.error}`);
    } else if (metadataDropped && reparsed.metadata.schema_version !== undefined) {
      fatals.push('the schema_version claim survived the drop');
    }
  }

  return fatals;
}

/**
 * Turns a raw draft into the exact string to publish.
 *
 * @param {string} draft
 * @returns {{body: string | null, fatals: string[], warnings: string[], edits: number}}
 */
function preparePublishedBody(draft) {
  const fatals = [];
  const warnings = [];

  if (typeof draft !== 'string' || draft.trim() === '') {
    return { body: null, fatals: ['draft is empty'], warnings, edits: 0 };
  }

  // Length-preserving in the sense that matters: these two filters delete only control and
  // invisible characters, never a span of visible content. They run *before* the `<!--` check
  // because stripping a NUL or zero-width space can reassemble an opener (`<!\x00--` -> `<!--`),
  // and the check must see the string that would actually be published.
  const transformed = stripInvisibleUnicode(stripControlChars(draft));

  const commentIndex = transformed.indexOf('<!--');
  if (commentIndex !== -1) {
    // Rejected, never stripped: removing it is what destroys the comment.
    return {
      body: null,
      fatals: [`draft contains an HTML comment opener "<!--" on line ${lineNumberOf(transformed, commentIndex)}; the handler rejects rather than strips it`],
      warnings,
      edits: 0,
    };
  }

  const body = `${RESEARCH_COMMENT_MARKER}\n${transformed.replace(/^\n+/, '')}`;

  // ---- decide -------------------------------------------------------------------------------
  fatals.push(...fatalContractViolations(body));

  /** Candidate edits, filtered below into the set that is actually applied. */
  const candidates = [];

  const located = locateMetadata(body);
  const extracted = extractMetadata(body);
  let metadataWasParsed = false;
  let metadataDropped = false;
  // Named separately because the severity table treats them differently: a 1.1 violation drops the
  // claim and publishes *without* touching the prose; only a gate disagreement annotates. When both
  // fire, the single annotation names both causes.
  let dropCause = null;
  let gateCause = null;
  let strayCause = null;
  let readinessCause = null;

  if (extracted.error) {
    warnings.push(`pipeline metadata unusable: ${extracted.error} (publishing anyway; nothing consumes it today)`);
  } else if (!isObject(extracted.metadata)) {
    // Valid JSON that is not an object: `null`, an array, a string, a number. There is no
    // schema_version claim to drop and no `gate` to compare against, and spreading one of these
    // would either throw (null) or invent an object the agent never wrote (`[1,2]` ->
    // `{"0":1,"1":2}`). Same severity as unparseable JSON: warn, publish, touch nothing.
    warnings.push(
      `pipeline metadata is ${extracted.metadata === null ? 'null' : `a ${Array.isArray(extracted.metadata) ? 'JSON array' : typeof extracted.metadata}`}, not an object; leaving it untouched (publishing anyway; nothing consumes it today)`,
    );
  } else {
    metadataWasParsed = true;
    const { errors } = validateMetadata(extracted.metadata);
    if (errors.length > 0) {
      warnings.push(`pipeline metadata does not satisfy schema_version ${SCHEMA_VERSION}: ${errors.join('; ')}`);
      // Stated as a fact about the metadata, not as a claim about what the handler did: the
      // annotation is composed before the applicability filter decides whether the drop survives.
      dropCause = `The pipeline metadata does not satisfy schema_version ${SCHEMA_VERSION}.`;
      const { schema_version: _dropped, ...withoutSchemaVersion } = extracted.metadata;
      candidates.push({
        name: 'drop schema_version',
        start: located.start,
        end: located.end,
        text: JSON.stringify(withoutSchemaVersion, null, 2),
      });
    }
  }

  if (metadataWasParsed) {
    const unearned = unearnedReadiness(extracted.metadata);
    if (unearned.length > 0) {
      readinessCause = `The gate claims \`${READY_OUTCOME}\` but ${unearned.join('; ')}.`;
      warnings.push(`readiness claim not earned: ${unearned.join('; ')}`);
    }
  }

  const gateRange = sectionRange(body, QUALITY_GATE_HEADING);

  // An `**Outcome:**` line outside the gate section is a machine-readable verdict a human reads as
  // one, sitting where nothing checks it. Same class as a fenced decoy: the published comment would
  // carry a prominent outcome contradicting both the gate section and the gate JSON, with nothing
  // in the log. It is therefore a cause in its own right — it forces and annotates the gate
  // section. The stray line itself is *not* rewritten. That is a risk-budget decision rather than a
  // structural limit — the edit machinery already rewrites one agent-authored range and would
  // support more — but every additional rewritten region is somewhere else this handler can get a
  // published comment wrong, which the plan's own history says is where its defects land. The
  // prompt asks for exactly one outcome line, and the annotation names any stray it finds — except
  // at the minimal rung, where there is only room for the verdict and the causes stay in the log.
  const strayOutcomes = outcomeLines(body).filter(
    (line) => !gateRange || line.start < gateRange.start || line.start >= gateRange.end,
  );
  if (strayOutcomes.length > 0) {
    strayCause = `${strayOutcomes.length} **Outcome:** line(s) sit outside the ### ${QUALITY_GATE_HEADING} section.`;
    warnings.push(`the body carries ${strayOutcomes.length} **Outcome:** line(s) outside the ### ${QUALITY_GATE_HEADING} section`);
  }

  if (!gateRange) {
    warnings.push(`no ### ${QUALITY_GATE_HEADING} section in the draft`);
  } else {
    const section = body.slice(gateRange.start, gateRange.end);
    const lines = outcomeLines(section);
    const prose = lines.length === 1 ? quotedToken(lines[0].text) : null;

    if (lines.length > 1) {
      // Structural: two live outcome lines contradict each other whatever the metadata says.
      gateCause = `The ### ${QUALITY_GATE_HEADING} section states ${lines.length} outcomes.`;
    } else if (!metadataWasParsed) {
      // Unreadable metadata is not a reason to trust the prose. Without this, one stray comma in
      // the JSON fence — or a missing envelope, or a fence holding `null` — switched off both this
      // check and the readiness check, and a `ready-for-change-factory` claim published with
      // nothing supporting it and nothing in the body to say so.
      //
      // Stated as "anything that is not already the conservative verdict", not as "equals the one
      // bad value": matching `ready-for-change-factory` exactly would let `Ready-For-Change-Factory`
      // or an unbackticked line through — the same one-character evasion the predicates above were
      // rewritten to close, one level up. A draft that states no outcome line at all is left alone:
      // there is nothing to correct, and inventing a verdict is not this branch's job.
      if (lines.length === 1 && prose !== FORCED_OUTCOME) {
        gateCause = `The ### ${QUALITY_GATE_HEADING} section states an outcome but the pipeline metadata could not be read, so nothing supports it.`;
      }
    } else if (lines.length === 0 || prose === null) {
      gateCause = `The ### ${QUALITY_GATE_HEADING} section states no machine-readable outcome.`;
    } else {
      const declared = extracted.metadata?.gate?.outcome;
      if (typeof declared !== 'string' || prose !== declared) {
        gateCause = `The ### ${QUALITY_GATE_HEADING} section says \`${quoteFragment(prose)}\` but the machine-readable gate says ${declared === undefined ? 'nothing' : `\`${quoteFragment(declared)}\``}.`;
      }
    }
  }

  if (gateCause) {
    // Logged whether or not the edit survives the filter below: an operator reading "skipped the
    // annotation" needs to know what disagreed.
    warnings.push(`the ### ${QUALITY_GATE_HEADING} section disagrees with the machine-readable gate: ${gateCause}`);
  }

  // One list, used for the annotation text, its fallback and the log line, so the three cannot
  // drift apart — a log that named a subset of the causes once printed an empty reason.
  const causes = [gateCause, readinessCause, strayCause, dropCause].filter(Boolean);

  if (gateRange && (gateCause || readinessCause || strayCause)) {
    // One edit over one range for every cause, so the forced outcome and the annotation can never
    // be applied half-way. The compact variant is the fallback when the full one does not fit.
    const section = body.slice(gateRange.start, gateRange.end);
    candidates.unshift({
      name: 'annotate quality gate',
      start: gateRange.start,
      end: gateRange.end,
      text: annotateQualityGate(section, causes),
      fallbackText: annotateQualityGate(section, causes, { compact: true }),
      minimalText: annotateQualityGate(section, causes, { minimal: true }),
    });
  } else if (!gateRange && (readinessCause || strayCause)) {
    // F3: no section to annotate, but the body still carries a machine-readable verdict a reader
    // will take at face value. Append the warning at the end rather than leave it only in a log
    // nobody reads. Zero-length range, so it cannot overlap the metadata edit.
    candidates.unshift({
      name: 'append stray-outcome notice',
      start: body.length,
      end: body.length,
      text: `\n> ⚠️ ${causes.join(' ')} Treat any outcome above as unverified.\n`,
      minimalText: '\n> ⚠️ Outcome unverified; see the workflow run log.\n',
    });
  }

  // Every edit here is non-fatal by the severity table, so none of them may cost the run. Two ways
  // one could: by introducing an HTML comment opener (the published body would then fail its own
  // contract), or by growing the body past the ceiling — a compact metadata block re-serialised
  // with two-space indent can add thousands of characters. A skipped edit publishes a comment that
  // is merely less annotated; a kept one can lose a 60-minute run. Candidates are considered in
  // priority order: the annotation first, because a human reads it and its size is bounded, then
  // the schema_version drop.
  const edits = [];
  let projected = body.length;
  for (const edit of candidates) {
    if (edit.text.includes('<!--')) {
      warnings.push(`skipped the "${edit.name}" edit: applying it would introduce an HTML comment opener`);
      continue;
    }
    let { text } = edit;
    let delta = text.length - (edit.end - edit.start);
    // Degrade rather than drop: full -> compact (correction, no appended block) -> minimal (the
    // corrected verdict alone). Each is tried only if the previous one breaches the ceiling.
    for (const [label, candidate] of [['compact', edit.fallbackText], ['minimal', edit.minimalText]]) {
      if (projected + delta <= MAX_COMMENT_CHARS) {
        break;
      }
      if (!candidate || candidate.includes('<!--')) {
        continue;
      }
      const candidateDelta = candidate.length - (edit.end - edit.start);
      if (projected + candidateDelta <= MAX_COMMENT_CHARS) {
        warnings.push(`used the ${label} "${edit.name}" edit: the fuller form would have breached the ${MAX_COMMENT_CHARS} limit`);
        text = candidate;
        delta = candidateDelta;
      }
    }
    if (projected + delta > MAX_COMMENT_CHARS) {
      warnings.push(
        `skipped the "${edit.name}" edit: applying it would take the body to ${projected + delta} characters, over the ${MAX_COMMENT_CHARS} limit`,
      );
      continue;
    }
    projected += delta;
    edits.push({ ...edit, text });
  }
  metadataDropped = edits.some((edit) => edit.name === 'drop schema_version');
  if (edits.some((edit) => edit.name === 'annotate quality gate')) {
    warnings.push(`forced the ### ${QUALITY_GATE_HEADING} outcome to ${FORCED_OUTCOME}: ${causes.join(' ')}`);
  }

  const ordered = [...edits].sort((a, b) => a.start - b.start);
  for (let i = 1; i < ordered.length; i += 1) {
    if (ordered[i].start < ordered[i - 1].end) {
      // Reachable when the draft puts the metadata envelope inside the ### Quality gate section —
      // i.e. the mandatory section order is wrong. Fail closed rather than apply half an edit set.
      fatals.push(
        `edits "${ordered[i - 1].name}" and "${ordered[i].name}" overlap; check the mandatory section order (### ${QUALITY_GATE_HEADING} before ### ${REFERENCES_HEADING} and the metadata envelope)`,
      );
    }
  }

  if (fatals.length > 0) {
    return { body: null, fatals, warnings, edits: edits.length };
  }

  // ---- apply --------------------------------------------------------------------------------
  let published = body;
  for (const edit of [...edits].sort((a, b) => b.start - a.start)) {
    published = published.slice(0, edit.start) + edit.text + published.slice(edit.end);
  }

  // ---- re-assert, on the exact string that will be published --------------------------------
  const reasserted = reassertPublished(published, { metadataWasParsed, metadataDropped });
  if (reasserted.length > 0) {
    return {
      body: null,
      fatals: reasserted.map((message) => `after applying ${edits.length} edit(s): ${message}`),
      warnings,
      edits: edits.length,
    };
  }

  return { body: published, fatals: [], warnings, edits: edits.length };
}

module.exports = async function ({ github, context, core }) {

  const env = process.env;

  // Detection steps run with continue-on-error, so the job can report result=success while the
  // detector warned, failed, or emitted nothing. Allow exactly a clean verdict; fail closed on
  // warning, failure, an empty conclusion, or any unknown value. gh-aw does not inject
  // DETECTION_RESULT, so the workflow passes it explicitly.
  if (env.DETECTION_RESULT !== 'success' || !['success', 'skipped'].includes(env.GH_AW_DETECTION_CONCLUSION)) {
    core.setFailed(
      `refusing to publish: detection result=${env.DETECTION_RESULT} conclusion=${env.GH_AW_DETECTION_CONCLUSION}`,
    );
    return;
  }

  const issueNumber = positiveIssueNumber(env.INPUT_ISSUE_NUMBER);
  if (!issueNumber) {
    core.setFailed(`refusing to publish: invalid issue number '${env.INPUT_ISSUE_NUMBER || ''}'.`);
    return;
  }

  const resolved = resolveDraftPath(env);
  if (resolved.error) {
    core.setFailed(
      `${resolved.error}. Looked in: ${resolved.candidates.join(', ')}. ` +
        'A missing draft can also mean secret redaction failed, which drops the whole agent artifact.',
    );
    return;
  }
  if (resolved.fallback) {
    core.warning(
      `resolved the draft by searching the artifact: using ${resolved.path}` +
        (resolved.fallback.length > 1 ? ` (${resolved.fallback.length} matches: ${resolved.fallback.join(', ')})` : ''),
    );
  }

  let draft;
  try {
    draft = fs.readFileSync(resolved.path, 'utf8');
  } catch (err) {
    core.setFailed(`cannot read ${resolved.path}: ${err.message}`);
    return;
  }

  let prepared;
  try {
    prepared = preparePublishedBody(draft);
  } catch (err) {
    // Any unexpected throw here would otherwise surface as an unhandled github-script error, with
    // no comment, no phase label and a consumed trigger label. Fail with something actionable.
    core.setFailed(`refusing to publish ${resolved.path}: the handler threw while preparing the body: ${err.stack || err.message}`);
    return;
  }
  const { body, fatals, warnings } = prepared;
  for (const warning of warnings) {
    core.warning(warning);
  }
  if (!body) {
    core.setFailed(
      `refusing to publish ${resolved.path}: ${fatals.join('; ')}. ` +
        'Re-running this job cannot fix it — the artifact carries the same draft; re-apply the research-factory label for a fresh run.',
    );
    return;
  }

  const { owner, repo } = context.repo;
  const botLogin = env.GH_AW_BOT_LOGIN || 'github-actions[bot]';

  try {
    // The number is re-read raw here (a custom safe-output job cannot depend on pre_activation), so
    // the pull-request check is repeated at the boundary that writes. The comment and label
    // endpoints accept pull requests, so without this a dispatch naming a PR gets researched and
    // stamped. Mirrors `issue-classifier/apply-routing.js`'s assertIssue.
    const { data: target } = await github.rest.issues.get({ owner, repo, issue_number: issueNumber });
    if (target.pull_request) {
      core.setFailed(`refusing to publish: #${issueNumber} is a pull request, not an issue.`);
      return;
    }
  } catch (err) {
    core.setFailed(`cannot read #${issueNumber}: ${err.message}`);
    return;
  }

  let existing = null;
  try {
    const comments = await github.paginate(github.rest.issues.listComments, {
      owner,
      repo,
      issue_number: issueNumber,
      per_page: 100,
    });
    existing = findResearchComment(comments, RESEARCH_COMMENT_MARKER, botLogin);
  } catch (err) {
    core.setFailed(`cannot list comments on #${issueNumber}: ${err.message}`);
    return;
  }

  try {
    if (existing) {
      await github.rest.issues.updateComment({ owner, repo, comment_id: existing.id, body });
      core.info(`Updated research comment ${existing.id} on #${issueNumber} (${body.length} chars).`);
    } else {
      const { data } = await github.rest.issues.createComment({ owner, repo, issue_number: issueNumber, body });
      core.info(`Created research comment ${data.id} on #${issueNumber} (${body.length} chars).`);
    }
  } catch (err) {
    core.setFailed(`failed to publish the research comment on #${issueNumber}: ${err.message}`);
    return;
  }

  core.setOutput('research_comment_published', 'true');
};

/**
 * Dual-use CLI, for the agent's pre-flight: `node update-research-comment.js <draft.md>` runs the
 * exact checks this job runs at publish time and exits non-zero if the draft would be rejected.
 *
 * Same argument as validate-metadata.js: a pre-flight that re-implements the handler's checks
 * drifts from them, and every divergence is a draft the agent believes passed and the handler
 * discards — unrecoverable, because the artifact keeps the same draft and the trigger label is
 * already consumed. github-script loads this file with `require`, so `require.main` is the
 * github-script runner and `main()` never fires there.
 */
function main() {
  const draftPath = process.argv[2];

  if (!draftPath) {
    console.error('usage: node update-research-comment.js <draft.md>');
    process.exit(2);
  }

  let draft;
  try {
    draft = fs.readFileSync(draftPath, 'utf8');
  } catch (err) {
    console.error(`cannot read ${draftPath}: ${err.message}`);
    process.exit(2);
  }

  const { body, fatals, warnings } = preparePublishedBody(draft);
  for (const warning of warnings) {
    console.error(`warning: ${warning}`);
  }

  if (!body) {
    console.error(`${draftPath} would be REJECTED by the publishing job:`);
    for (const fatal of fatals) {
      console.error(`  - ${fatal}`);
    }
    process.exit(1);
  }

  console.log(`${draftPath}: would publish ${body.length} of ${MAX_COMMENT_CHARS} characters.`);
}

if (require.main === module) {
  main();
}

module.exports.preparePublishedBody = preparePublishedBody;
module.exports.resolveDraftPath = resolveDraftPath;
module.exports.fatalContractViolations = fatalContractViolations;
module.exports.reassertPublished = reassertPublished;
module.exports.quoteFragment = quoteFragment;
module.exports.RESEARCH_DIR = RESEARCH_DIR;
module.exports.DRAFT_FINAL = DRAFT_FINAL;
module.exports.MAX_COMMENT_CHARS = MAX_COMMENT_CHARS;
