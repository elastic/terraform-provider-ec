import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const handler = require('./update-research-comment.js');
const { preparePublishedBody, resolveDraftPath, reassertPublished, quoteFragment, MAX_COMMENT_CHARS } = handler;
const { RESEARCH_COMMENT_MARKER } = require('../lib/intake/research-factory-constants.js');
const { extractMetadata } = require('./locate-metadata.js');

function validMetadata() {
  return {
    schema_version: '1.1',
    recommendation: { spine: 'extend-deployment-schema', confidence: 'medium', approach_index: 0 },
    open_questions: [],
    affected_capabilities: ['ec_deployment'],
    estimated_scope: 'medium',
    references: [{ type: 'repo-path', path: 'ec/provider.go' }],
    gate: {
      outcome: 'ready-for-change-factory',
      checklist: { grounded: true, mapped: true, compatible: true, versioned: true, testable: true, idiomatic: true },
      score: 88,
      scores: [86, 88],
      converged: true,
      rounds: 2,
      outstanding_feedback: [],
      author_model: 'anthropic/claude-sonnet-5',
      critic: { model: 'openai/gpt-6.1-sol', status: 'ok' },
    },
  };
}

/**
 * `metadata` is overloaded on purpose: an object is serialised into the fence, a **string** is
 * written into the fence verbatim (so `'null'` produces a fence containing `null`), and `null`
 * omits the `<details>` envelope entirely.
 *
 * @param {{approaches?: number, proseOutcome?: string|null, metadata?: object|null|string,
 *          filler?: string, qualityGate?: boolean, references?: boolean, extra?: string}} opts
 */
function makeDraft(opts = {}) {
  const {
    approaches = 2,
    proseOutcome = 'ready-for-change-factory',
    metadata = validMetadata(),
    filler = '',
    qualityGate = true,
    references = true,
    extra = '',
  } = opts;

  const approachBlocks = Array.from({ length: approaches }, (_, i) =>
    [`#### Approach ${String.fromCharCode(65 + i)}: variant ${i}`, '', `Sketch for variant ${i}.`, ''].join('\n'),
  ).join('');

  const gate = qualityGate
    ? [
        '### Quality gate',
        '',
        proseOutcome === null ? 'The gate did not complete.' : `**Outcome:** \`${proseOutcome}\` - all checklist items pass.`,
        '',
        '| Check | Result |',
        '| --- | --- |',
        '| Grounded | pass |',
        '| Mapped | pass |',
        '| Compatible | pass |',
        '| Versioned | pass |',
        '| Testable | pass |',
        '| Idiomatic | pass |',
        '',
        'Final score 88 against the threshold of 85, over 2 rounds.',
        '',
      ].join('\n')
    : '';

  const metadataBlock =
    metadata === null
      ? ''
      : [
          '<details><summary>🤖 Pipeline metadata</summary>',
          '',
          '```json',
          typeof metadata === 'string' ? metadata : JSON.stringify(metadata, null, 2),
          '```',
          '',
          '</details>',
          '',
        ].join('\n');

  return [
    '## Implementation research',
    '',
    'Run: https://github.com/elastic/terraform-provider-ec/actions/runs/1',
    '',
    '> Edits inside this comment are read as input on the next run but are not preserved verbatim.',
    '',
    '### Problem framing',
    '',
    `The issue asks for a new attribute.${filler}`,
    '',
    '### Approaches considered',
    '',
    approachBlocks,
    '### Recommendation',
    '',
    'Approach A, because it matches the existing schema.',
    '',
    '### Open questions',
    '',
    '- none',
    '',
    '### Out of scope',
    '',
    '- none',
    '',
    gate,
    references ? '### References\n\n- `ec/provider.go`\n' : '',
    extra,
    metadataBlock,
  ].join('\n');
}

describe('preparePublishedBody — happy path', () => {
  test('prepends the marker and publishes an unedited conformant draft', () => {
    const { body, fatals, warnings, edits } = preparePublishedBody(makeDraft());

    assert.deepEqual(fatals, []);
    assert.deepEqual(warnings, []);
    assert.equal(edits, 0);
    assert.ok(body.startsWith(`${RESEARCH_COMMENT_MARKER}\n`));
    assert.ok(body.includes('### Recommendation'));
    assert.equal(extractMetadata(body).metadata.schema_version, '1.1');
  });

  test('strips control and invisible characters without cutting content', () => {
    const draft = makeDraft().replace('### Recommendation', '### Recommendation​\x07');
    const { body, fatals } = preparePublishedBody(draft);

    assert.deepEqual(fatals, []);
    assert.ok(body.includes('### Recommendation\n'));
    assert.ok(!body.includes('​'));
  });

  test('accepts a json fence outside the metadata envelope', () => {
    const draft = makeDraft({ extra: '```json\n{"example": "api response"}\n```\n' });
    const { body, fatals, warnings } = preparePublishedBody(draft);

    assert.deepEqual(fatals, []);
    assert.deepEqual(warnings, []);
    assert.ok(body.includes('{"example": "api response"}'));
  });

  test('ignores a decoy details block before the metadata envelope', () => {
    const draft = makeDraft({
      extra: '<details><summary>Pros and cons</summary>\n\n```json\n{"decoy": true}\n```\n\n</details>\n',
    });
    const { fatals, warnings } = preparePublishedBody(draft);

    assert.deepEqual(fatals, []);
    assert.deepEqual(warnings, []);
  });
});

describe('preparePublishedBody — fatal rejections', () => {
  test('rejects an empty or whitespace-only draft', () => {
    assert.deepEqual(preparePublishedBody('').fatals, ['draft is empty']);
    assert.deepEqual(preparePublishedBody('   \n\t\n').fatals, ['draft is empty']);
    assert.deepEqual(preparePublishedBody(undefined).fatals, ['draft is empty']);
  });

  test('rejects an HTML comment opener and names its line', () => {
    const draft = makeDraft().replace('### Problem framing', '<!-- a note -->\n### Problem framing');
    const { body, fatals } = preparePublishedBody(draft);

    assert.equal(body, null);
    assert.equal(fatals.length, 1);
    assert.match(fatals[0], /HTML comment opener "<!--" on line 7/);
  });

  test('rejects an unterminated opener inside a code fence — the case the sanitiser would eat', () => {
    const draft = makeDraft({ extra: '```hcl\nresource "ec_deployment" "x" {\n  # <!-- stray\n}\n```\n' });
    const { body, fatals } = preparePublishedBody(draft);

    assert.equal(body, null);
    assert.match(fatals[0], /HTML comment opener/);
  });

  test('rejects an opener that the control-character strip reassembles', () => {
    const draft = makeDraft().replace('### Problem framing', '<!​-- hidden -->\n### Problem framing');
    const { body, fatals } = preparePublishedBody(draft);

    assert.equal(body, null);
    assert.match(fatals[0], /HTML comment opener/);
  });

  test('rejects an opener that only the control-character strip makes visible', () => {
    // Pins the order: the scan must run on the transformed text. Scanning the raw draft instead
    // lets `<!\x00--` and `<!\u200B--` through, and the body then publishes an opener it was
    // checked for. (reassertPublished catches it afterwards, but as a discarded run.)
    for (const hidden of ['<!\x00-- hidden -->', '<!\u200B-- hidden -->', '<\u200B!-- hidden -->']) {
      const draft = makeDraft().replace('### Problem framing', () => `${hidden}\n### Problem framing`);
      const { body, fatals } = preparePublishedBody(draft);
      const text = fatals.join('\n');

      assert.equal(body, null, `not rejected: ${JSON.stringify(hidden)}`);
      // The *input* rejection, not the post-apply backstop: scanning the raw draft instead lets
      // these through the front door and only `reassertPublished` stops them, which discards the
      // run instead of telling the agent its draft is bad.
      assert.match(text, /rejects rather than strips it/);
      assert.ok(!text.includes('after applying'), `caught by the backstop, not the input scan: ${text}`);
    }
  });

  test('rejects a single-approach draft', () => {
    const { body, fatals } = preparePublishedBody(makeDraft({ approaches: 1 }));

    assert.equal(body, null);
    assert.match(fatals.join('\n'), /1 #### approach heading\(s\).*at least 2/);
  });

  test('does not count #### headings inside a code fence as approaches', () => {
    // The fenced heading has to sit INSIDE ### Approaches considered, or the test passes for the
    // wrong reason (anything after ### References is out of the section regardless of fences).
    const draft = makeDraft({ approaches: 2 }).replace(
      '#### Approach B: variant 1',
      () => '```markdown\n#### Approach B: fenced example\n```',
    );
    assert.ok(draft.includes('```markdown'), 'fixture rewrote the second approach');
    const { body, fatals } = preparePublishedBody(draft);

    assert.equal(body, null, 'a fenced #### is an example, not an approach');
    assert.match(fatals.join('\n'), /has 1 #### approach heading\(s\) outside code fences/);
  });

  test('counts a #### that follows a closed fence inside the section', () => {
    const draft = makeDraft({ approaches: 2 }).replace(
      '#### Approach B: variant 1',
      () => '```hcl\nresource "ec_deployment" "x" {}\n```\n\n#### Approach B: variant 1',
    );
    const { body, fatals } = preparePublishedBody(draft);

    assert.deepEqual(fatals, []);
    assert.ok(body);
  });

  test('does not count #### headings outside the approaches section', () => {
    const draft = makeDraft({ approaches: 1, extra: '#### A stray heading under References\n' });
    const { fatals } = preparePublishedBody(draft);

    assert.match(fatals.join('\n'), /needs at least 2/);
  });

  test('rejects a draft missing ### Recommendation or ### References', () => {
    const noRecommendation = makeDraft().replace('### Recommendation', '### Chosen approach');
    assert.match(preparePublishedBody(noRecommendation).fatals.join('\n'), /missing mandatory section: ### Recommendation/);

    const noReferences = makeDraft({ references: false });
    assert.match(preparePublishedBody(noReferences).fatals.join('\n'), /missing mandatory section: ### References/);
  });

  test('rejects a draft missing ### Approaches considered', () => {
    const draft = makeDraft().replace('### Approaches considered', '### Options');
    assert.match(preparePublishedBody(draft).fatals.join('\n'), /missing mandatory section: ### Approaches considered/);
  });

  test('rejects a body over 60 000 characters', () => {
    const { body, fatals } = preparePublishedBody(makeDraft({ filler: 'x'.repeat(MAX_COMMENT_CHARS) }));

    assert.equal(body, null);
    assert.match(fatals.join('\n'), /over the 60000 limit/);
  });

  test('skips an annotation that would breach the ceiling rather than discarding the run', () => {
    const metadata = validMetadata();
    metadata.gate.outcome = 'research-needs-human';
    const base = makeDraft({ metadata, proseOutcome: 'ready-for-change-factory' });
    // Sit just under the limit so only the annotation can push the published body over it.
    const headroom = MAX_COMMENT_CHARS - base.length - RESEARCH_COMMENT_MARKER.length - 60;
    assert.ok(headroom > 0);
    const draft = makeDraft({ metadata, proseOutcome: 'ready-for-change-factory', filler: 'x'.repeat(headroom) });

    const beforeEdits = draft.length + RESEARCH_COMMENT_MARKER.length + 1;
    assert.ok(beforeEdits <= MAX_COMMENT_CHARS, `pre-edit body already over the limit: ${beforeEdits}`);

    const { body, fatals, warnings, edits } = preparePublishedBody(draft);
    assert.deepEqual(fatals, []);
    assert.equal(edits, 0);
    assert.ok(body.length <= MAX_COMMENT_CHARS);
    assert.match(warnings.join('\n'), /skipped the "annotate quality gate" edit: applying it would take the body to \d+ characters/);
  });

  test('skips a schema_version drop that would breach the ceiling — a compact block re-indents much larger', () => {
    const metadata = validMetadata();
    metadata.schema_version = '1.0'; // a warn row: the claim would be dropped
    metadata.references = Array.from({ length: 400 }, (_, i) => ({ type: 'repo-path', path: `ec/internal/file-${i}.go` }));
    // Compact in the draft, pretty-printed by the rewrite: nothing tells the agent to pretty-print.
    const draft = makeDraft({ metadata: JSON.stringify(metadata) });
    const padding = MAX_COMMENT_CHARS - draft.length - RESEARCH_COMMENT_MARKER.length - 60;
    assert.ok(padding > 0);
    const padded = makeDraft({ metadata: JSON.stringify(metadata), filler: 'x'.repeat(padding) });
    assert.ok(padded.length + RESEARCH_COMMENT_MARKER.length + 1 <= MAX_COMMENT_CHARS, 'the draft itself fits');

    const { body, fatals, warnings, edits } = preparePublishedBody(padded);

    assert.deepEqual(fatals, []);
    assert.equal(edits, 0);
    assert.ok(body.length <= MAX_COMMENT_CHARS);
    assert.match(warnings.join('\n'), /skipped the "drop schema_version" edit: applying it would take the body to \d+ characters/);
    assert.equal(extractMetadata(body).metadata.schema_version, '1.0', 'the claim survives, unchanged');
  });

  test('rejects an edit set whose ranges overlap', () => {
    const metadata = validMetadata();
    metadata.gate.rounds = 99;
    metadata.gate.outcome = 'research-needs-human';
    // Metadata envelope inside the quality-gate section: the mandatory section order is wrong.
    const draft = makeDraft({ metadata: null, proseOutcome: 'ready-for-change-factory' })
      .replace(
        'Final score 88 against the threshold of 85, over 2 rounds.',
        [
          'Final score 88 against the threshold of 85, over 2 rounds.',
          '',
          '<details><summary>🤖 Pipeline metadata</summary>',
          '',
          '```json',
          JSON.stringify(metadata, null, 2),
          '```',
          '',
          '</details>',
        ].join('\n'),
      );

    const { body, fatals } = preparePublishedBody(draft);
    assert.equal(body, null);
    assert.match(fatals.join('\n'), /overlap; check the mandatory section order/);
  });
});

describe('preparePublishedBody — metadata severity', () => {
  test('publishes with a warning when the metadata envelope is missing', () => {
    const { body, fatals, warnings } = preparePublishedBody(makeDraft({ metadata: null }));

    assert.deepEqual(fatals, []);
    assert.ok(body);
    assert.match(warnings.join('\n'), /pipeline metadata unusable/);
  });

  test('publishes with a warning when the metadata JSON does not parse', () => {
    const { body, warnings } = preparePublishedBody(makeDraft({ metadata: '{"schema_version": "1.1",}' }));

    assert.ok(body);
    assert.match(warnings.join('\n'), /could not be parsed/);
    assert.ok(body.includes('{"schema_version": "1.1",}'), 'an unparseable block is left untouched');
  });

  test('drops the schema_version claim structurally, leaving parseable JSON', () => {
    const metadata = validMetadata();
    metadata.recommendation.approach_index = '0';
    const { body, fatals, warnings, edits } = preparePublishedBody(makeDraft({ metadata }));

    assert.deepEqual(fatals, []);
    assert.equal(edits, 1);
    assert.match(warnings.join('\n'), /does not satisfy schema_version 1\.1/);

    const reparsed = extractMetadata(body);
    assert.equal(reparsed.error, undefined);
    assert.equal(reparsed.metadata.schema_version, undefined);
    assert.equal(reparsed.metadata.recommendation.spine, 'extend-deployment-schema');
    assert.equal(reparsed.metadata.gate.outcome, 'ready-for-change-factory');
  });

  test('a schema_version violation alone neither annotates nor forces the outcome', () => {
    const metadata = validMetadata();
    metadata.estimated_scope = 'enormous';
    const { body, warnings } = preparePublishedBody(makeDraft({ metadata }));

    assert.ok(body.includes('**Outcome:** `ready-for-change-factory`'));
    assert.ok(!body.includes('⚠️'));
    assert.ok(!warnings.join('\n').includes('forced the'));
  });
});

describe('preparePublishedBody — metadata that is not an object', () => {
  // Valid JSON, not an object. Spreading one of these used to throw (null) or invent an object the
  // agent never wrote ([1,2] -> {"0":1,"1":2}), losing or corrupting a 60-minute run over what the
  // severity table calls a warn row.
  for (const [label, fence] of [
    ['null', 'null'],
    ['an array', '[1, 2, 3]'],
    ['a string', '"hello"'],
    ['a number', '42'],
    ['a boolean', 'true'],
  ]) {
    test(`publishes untouched when the metadata fence holds ${label}`, () => {
      const { body, fatals, warnings, edits } = preparePublishedBody(makeDraft({ metadata: fence }));

      assert.deepEqual(fatals, []);
      assert.equal(edits, 0);
      assert.ok(body.includes(`\n${fence}\n`), 'the fence content is left exactly as written');
      assert.match(warnings.join('\n'), /not an object; leaving it untouched/);
      assert.ok(!body.includes('⚠️'), 'no gate annotation: there is no machine-readable gate to disagree with');
    });
  }

  test('keeps the schema_version claim rather than discarding the run when the rewrite would introduce an opener', () => {
    const metadata = validMetadata();
    // Escaped in the draft (so the literal `<!--` check passes) but literal once JSON.stringify
    // re-emits it. The handler must not let its own mutation escalate a warn row into a lost run.
    metadata.gate.outstanding_feedback = ['<!-- smuggled'];
    metadata.gate.rounds = 99; // violates rounds === scores.length, so the drop would otherwise fire
    const { body, fatals, warnings, edits } = preparePublishedBody(
      makeDraft({ metadata: JSON.stringify(metadata, null, 2).replace('<!--', '\\u003c!--') }),
    );

    assert.deepEqual(fatals, []);
    assert.equal(edits, 0);
    assert.ok(body, 'the comment is still published');
    assert.match(warnings.join('\n'), /skipped the "drop schema_version" edit: applying it would introduce an HTML comment opener/);
    assert.equal(extractMetadata(body).metadata.schema_version, '1.1', 'the claim survives untouched');
    assert.equal(body.split('<!--').length - 1, 1, 'only the leading marker');
  });
});

describe('preparePublishedBody — quality-gate annotation', () => {
  test('forces the outcome and annotates when the prose disagrees with the gate', () => {
    const metadata = validMetadata();
    metadata.gate.outcome = 'research-needs-human';
    const { body, fatals, edits } = preparePublishedBody(
      makeDraft({ metadata, proseOutcome: 'ready-for-change-factory' }),
    );

    assert.deepEqual(fatals, []);
    assert.equal(edits, 1);
    assert.ok(body.includes('**Outcome:** `research-needs-human` - forced by the comment handler'));
    assert.ok(!body.includes('**Outcome:** `ready-for-change-factory`'));
    assert.match(body, /> ⚠️ .*machine-readable gate says `research-needs-human`/);
    assert.ok(body.includes('### References'), 'the following section survives the rewrite');
    assert.ok(body.includes('| Grounded | pass |'), 'the checklist table survives the rewrite');
  });

  test('does not double the sentence-ending period on the forced line', () => {
    const metadata = validMetadata();
    metadata.gate.outcome = 'research-needs-human';
    const { body } = preparePublishedBody(makeDraft({ metadata, proseOutcome: 'ready-for-change-factory' }));
    const forced = body.split('\n').find((line) => line.includes('forced by the comment handler'));

    assert.ok(forced.endsWith('.'), forced);
    assert.ok(!forced.endsWith('..'), `double period: ${forced}`);
  });

  test('annotates when the quality gate states no outcome at all', () => {
    const { body, edits } = preparePublishedBody(makeDraft({ proseOutcome: null }));

    assert.equal(edits, 1);
    assert.match(body, /### Quality gate\n\n\*\*Outcome:\*\* `research-needs-human` - forced/);
    assert.match(body, /states no machine-readable outcome/);
  });

  test('emits one combined annotation when the gate disagrees and the metadata is non-conformant', () => {
    const metadata = validMetadata();
    metadata.gate.outcome = 'research-needs-human';
    metadata.gate.rounds = 4; // violates rounds === scores.length
    const { body, edits, warnings } = preparePublishedBody(
      makeDraft({ metadata, proseOutcome: 'ready-for-change-factory' }),
    );

    assert.equal(edits, 2);
    const annotations = body.match(/> ⚠️/g) || [];
    assert.equal(annotations.length, 1, 'exactly one annotation naming both causes');
    assert.match(body, /machine-readable gate says `research-needs-human`/);
    assert.match(body, /does not satisfy schema_version 1\.1/);

    const reparsed = extractMetadata(body);
    assert.equal(reparsed.error, undefined, 'the descending-order apply left the metadata intact');
    assert.equal(reparsed.metadata.schema_version, undefined);
    assert.equal(reparsed.metadata.gate.rounds, 4);
    assert.match(warnings.join('\n'), /forced the ### Quality gate outcome/);
  });

  test('does not expand a $-pattern smuggled into the prose outcome', () => {
    const metadata = validMetadata();
    metadata.gate.outcome = 'research-needs-human';
    const draft = makeDraft({ metadata, proseOutcome: 'ready-for-change-factory' }).replace(
      '**Outcome:** `ready-for-change-factory` - all checklist items pass.',
      // Replacer function: the `$&` has to reach the draft literally, not expand here.
      () => '**Outcome:** `$&` - all checklist items pass.',
    );
    const { body } = preparePublishedBody(draft);

    assert.ok(body.includes('**Outcome:** `research-needs-human` - forced by the comment handler'));
    assert.ok(!body.includes('all checklist items pass'), 'the forced line replaced the original');
  });

  test('bounds how much the annotation can grow the body, whatever the draft quotes', () => {
    // The prompt tells the agent to stay at or under 59 000 characters because the handler may add
    // characters after the pre-flight has run. That margin is only real if the annotation cannot
    // quote an unbounded fragment back: before the clamp, a 2 000-character outcome token grew the
    // body by 8 393 characters and the 60 000 re-assert then discarded a draft that fitted.
    const MARGIN = 1000;
    for (const quoted of [0, 40, 500, 2000]) {
      const metadata = validMetadata();
      metadata.gate.outcome = 'Z'.repeat(3000); // also violates the outcome enum, so both edits fire
      metadata.gate.rounds = 4;
      const draft = makeDraft({ metadata, proseOutcome: 'Y'.repeat(quoted) });
      const { body, fatals } = preparePublishedBody(draft);

      assert.deepEqual(fatals, []);
      const growth = body.length - (draft.length + RESEARCH_COMMENT_MARKER.length + 1);
      assert.ok(growth <= MARGIN, `quoting ${quoted} chars grew the body by ${growth}, over the ${MARGIN} margin`);
    }
  });

  test('warns without annotating when there is no quality-gate section', () => {
    const { body, fatals, warnings } = preparePublishedBody(makeDraft({ qualityGate: false }));

    assert.deepEqual(fatals, []);
    assert.ok(body);
    assert.match(warnings.join('\n'), /no ### Quality gate section/);
    assert.ok(!body.includes('⚠️'));
  });
});

describe('quoteFragment', () => {
  test('strips the characters that let a quoted fragment escape its context', () => {
    assert.equal(quoteFragment('<!-- smuggled'), '!-- smuggled');
    assert.equal(quoteFragment('x</summary></details>'), 'x/summary/details');
    assert.equal(quoteFragment('a`b`c'), 'abc');
  });

  test('collapses whitespace so a newline cannot inject markdown', () => {
    assert.equal(quoteFragment('x\n\n#### Fake approach\n\n### Recommendation\n'), 'x #### Fake approach ### Recommendation');
  });

  test('clamps a long fragment and labels an empty one', () => {
    const clamped = quoteFragment('Y'.repeat(500));
    assert.equal(clamped.length, 81);
    assert.ok(clamped.endsWith('…'));
    assert.equal(quoteFragment('   '), '(empty)');
    assert.equal(quoteFragment('<>`'), '(empty)');
  });

  test('clamps by code point, so it cannot split a surrogate pair', () => {
    // A lone surrogate half reaches createComment's body and octokit serialises it as an unpaired
    // escape; clamping by UTF-16 unit at an 80-unit boundary is how it would get there.
    const clamped = quoteFragment(`a${'😀'.repeat(100)}`);
    assert.ok(clamped.isWellFormed(), 'no lone surrogate');
    assert.equal([...clamped].length, 81);
  });
});

describe('preparePublishedBody — the annotation cannot be used as an injection vector', () => {
  function withGateOutcome(outcome, { escapeOpener = false } = {}) {
    const metadata = validMetadata();
    metadata.gate.outcome = outcome;
    const json = JSON.stringify(metadata, null, 2);
    return makeDraft({
      metadata: escapeOpener ? json.replace('<!--', '\\u003c!--') : json,
      proseOutcome: 'ready-for-change-factory',
    });
  }

  test('a gate outcome carrying a JSON-escaped opener publishes, with only the marker', () => {
    // The draft holds no literal `<!--`, so it passes the input check; the drop edit would
    // re-serialise it literal, and the annotation would quote it. Both paths are now guarded.
    const { body, fatals, warnings } = preparePublishedBody(withGateOutcome('<!-- smuggled', { escapeOpener: true }));

    assert.deepEqual(fatals, []);
    assert.equal(body.split('<!--').length - 1, 1, 'only the leading marker');
    assert.match(warnings.join('\n'), /skipped the "drop schema_version" edit: applying it would introduce an HTML comment opener/);
  });

  test('a gate outcome forging a second metadata envelope publishes, and the metadata still parses', () => {
    const forged = 'x</summary></details>\n\n<details><summary>🤖 Pipeline metadata</summary>\n\n```json\n{"a":1}\n```\n\n</details>';
    const { body, fatals } = preparePublishedBody(withGateOutcome(forged));

    assert.deepEqual(fatals, []);
    const reparsed = extractMetadata(body);
    assert.equal(reparsed.error, undefined, 'exactly one envelope survives');
    assert.ok(!body.includes('</details>\n\n<details>'));
  });

  test('a gate outcome carrying markdown cannot inject a heading into the published body', () => {
    const { body, fatals } = preparePublishedBody(withGateOutcome('x\n\n#### Fake approach\n\n### Recommendation\n'));

    assert.deepEqual(fatals, []);
    assert.equal((body.match(/^#### /gm) || []).length, 2, 'still exactly the two real approaches');
    assert.equal((body.match(/^### Recommendation$/gm) || []).length, 1);
  });
});

describe('reassertPublished', () => {
  // The backstop: the applicability filter is supposed to stop an unsafe edit reaching this, so it
  // is exercised directly rather than through an input that no longer gets here.
  const clean = `${RESEARCH_COMMENT_MARKER}\n${makeDraft()}`;

  test('passes the string the handler would publish', () => {
    assert.deepEqual(reassertPublished(clean, { metadataWasParsed: true, metadataDropped: false }), []);
  });

  test('catches a second HTML comment opener', () => {
    const broken = clean.replace('### Recommendation', '<!-- injected -->\n### Recommendation');
    assert.match(reassertPublished(broken, { metadataWasParsed: true, metadataDropped: false }).join('\n'), /carries 2 HTML comment opener\(s\)/);
  });

  test('catches a missing leading marker', () => {
    const broken = clean.slice(RESEARCH_COMMENT_MARKER.length + 1);
    assert.match(reassertPublished(broken, { metadataWasParsed: false, metadataDropped: false }).join('\n'), /carries 0 HTML comment opener\(s\)/);
  });

  test('catches metadata the edit broke', () => {
    const broken = clean.replace('"schema_version": "1.1",', '"schema_version": "1.1",,');
    assert.match(reassertPublished(broken, { metadataWasParsed: true, metadataDropped: false }).join('\n'), /no longer parses/);
  });

  test('catches a schema_version claim that survived the drop', () => {
    assert.match(reassertPublished(clean, { metadataWasParsed: true, metadataDropped: true }).join('\n'), /schema_version claim survived the drop/);
  });

  test('collects every violation rather than stopping at the first', () => {
    const broken = clean
      .replace('### References', '<!-- injected -->\n### Sources')
      .replace('"schema_version": "1.1",', '"schema_version": "1.1",,');
    const fatals = reassertPublished(broken, { metadataWasParsed: true, metadataDropped: false });

    assert.ok(fatals.length >= 3, `expected several, got ${fatals.length}: ${fatals.join(' | ')}`);
    assert.match(fatals.join('\n'), /missing mandatory section: ### References/);
    assert.match(fatals.join('\n'), /HTML comment opener/);
    assert.match(fatals.join('\n'), /no longer parses/);
  });

  test('catches a structural violation introduced after the edits', () => {
    const broken = clean.replace('### References', '### Sources');
    assert.match(reassertPublished(broken, { metadataWasParsed: true, metadataDropped: false }).join('\n'), /missing mandatory section: ### References/);
  });
});

describe('preparePublishedBody — the outcome line is read outside code fences', () => {
  function gateSection(lines, declaredOutcome) {
    const metadata = validMetadata();
    metadata.gate.outcome = declaredOutcome;
    return makeDraft({ metadata }).replace(
      '**Outcome:** `ready-for-change-factory` - all checklist items pass.',
      () => lines.join('\n'),
    );
  }

  test('a fenced decoy outcome cannot suppress the disagreement check', () => {
    // Without fence-awareness the decoy is read as THE prose outcome, it agrees with the gate, and
    // the real disagreeing line below publishes with no annotation and no warning at all.
    const draft = gateSection(
      ['```text', '**Outcome:** `research-needs-human` - decoy', '```', '', '**Outcome:** `ready-for-change-factory` - all checklist items pass.'],
      'research-needs-human',
    );
    const { body, fatals, warnings, edits } = preparePublishedBody(draft);

    assert.deepEqual(fatals, []);
    assert.equal(edits, 1);
    assert.ok(body.includes('**Outcome:** `research-needs-human` - forced by the comment handler'));
    assert.ok(!body.includes('**Outcome:** `ready-for-change-factory`'), 'the live line was corrected');
    assert.ok(body.includes('- decoy'), 'the fenced example is left alone');
    assert.match(warnings.join('\n'), /disagrees with the machine-readable gate/);
  });

  test('every live outcome line is corrected, not just the first', () => {
    const draft = gateSection(
      ['**Outcome:** `ready-for-change-factory` - all checklist items pass.', '', '**Outcome:** `definitely-ship-it` - stray.'],
      'ready-for-change-factory',
    );
    const { body, fatals } = preparePublishedBody(draft);

    assert.deepEqual(fatals, []);
    assert.equal((body.match(/forced by the comment handler/g) || []).length, 2);
    assert.ok(!body.includes('definitely-ship-it'), 'no contradictory outcome survives');
  });

  test('two outcome lines are themselves a disagreement, whatever they say', () => {
    const draft = gateSection(
      ['**Outcome:** `research-needs-human` - one.', '', '**Outcome:** `research-needs-human` - two.'],
      'research-needs-human',
    );
    const { warnings, edits } = preparePublishedBody(draft);

    assert.equal(edits, 1);
    assert.match(warnings.join('\n'), /states 2 outcomes/);
  });

  test('an outcome line outside the quality-gate section is a cause of its own', () => {
    // A machine-readable verdict a human reads as one, sitting where nothing checks it. Before this
    // it published with no annotation and nothing in the log.
    const metadata = validMetadata();
    metadata.gate.outcome = 'research-needs-human';
    const draft = makeDraft({ metadata, proseOutcome: 'research-needs-human' }).replace(
      '### Open questions',
      () => '### Open questions\n\n**Outcome:** `ready-for-change-factory` - all green',
    );
    const { body, fatals, warnings, edits } = preparePublishedBody(draft);

    assert.deepEqual(fatals, []);
    assert.equal(edits, 1);
    assert.match(warnings.join('\n'), /carries 1 \*\*Outcome:\*\* line\(s\) outside the ### Quality gate section/);
    assert.ok(body.includes('**Outcome:** `research-needs-human` - forced by the comment handler'));
    assert.match(body, /> ⚠️ .*sit outside the ### Quality gate section/);
  });

  test('names every cause in the log, including a stray line on its own', () => {
    // The log line and the annotation are built from one list. When they were built separately a
    // stray-only run logged "forced the ### Quality gate outcome to research-needs-human: ".
    const metadata = validMetadata();
    const draft = makeDraft({ metadata }).replace(
      '### Open questions',
      () => '### Open questions\n\n**Outcome:** `ready-for-change-factory` - stray',
    );
    const { warnings, edits } = preparePublishedBody(draft);
    const forcedLog = warnings.find((w) => w.startsWith('forced the'));

    assert.equal(edits, 1);
    assert.ok(forcedLog, 'the forced outcome is logged');
    assert.match(forcedLog, /sit outside the ### Quality gate section\.$/, `empty reason: ${forcedLog}`);
  });

  test('the compact fallback names the cause too, when a stray line is the only one', () => {
    // The compact form is built from the same `causes` list as the full one. Built separately, a
    // stray-only run publishes "forced by the comment handler:" with an empty reason — the same
    // defect that reached the log, but this time in the comment a human reads.
    const metadata = validMetadata();
    const base = makeDraft({ metadata }).replace(
      '### Open questions',
      () => '### Open questions\n\n**Outcome:** `ready-for-change-factory` - stray',
    );
    // ~120 characters left: enough for the correction alone, not for the appended ⚠️ block.
    const filler = 'x'.repeat(MAX_COMMENT_CHARS - base.length - RESEARCH_COMMENT_MARKER.length - 120);
    const draft = makeDraft({ metadata, filler }).replace(
      '### Open questions',
      () => '### Open questions\n\n**Outcome:** `ready-for-change-factory` - stray',
    );
    const { body, fatals, warnings } = preparePublishedBody(draft);

    assert.deepEqual(fatals, []);
    assert.match(warnings.join('\n'), /used the compact "annotate quality gate" edit/);
    const forced = body.split('\n').find((line) => line.includes('forced by the comment handler'));
    assert.ok(forced, 'the correction is present');
    assert.match(forced, /sit outside the ### Quality gate section\.$/, `empty reason: ${forced}`);
  });

  test('the appended notice names every cause, not just the stray line', () => {
    const metadata = validMetadata();
    metadata.gate.rounds = 4; // also violates 1.1, so the drop fires alongside the stray cause
    const draft = makeDraft({ metadata, qualityGate: false }).replace(
      '### Open questions',
      () => '### Open questions\n\n**Outcome:** `ready-for-change-factory` - stray',
    );
    const { body, fatals } = preparePublishedBody(draft);

    assert.deepEqual(fatals, []);
    const notice = body.split('\n').find((line) => line.startsWith('> ⚠️'));
    assert.ok(notice);
    assert.match(notice, /sit outside the ### Quality gate section\./);
    assert.match(notice, /does not satisfy schema_version 1\.1\./);
  });

  test('appends a notice when there is no quality-gate section to annotate', () => {
    const metadata = validMetadata();
    const draft = makeDraft({ metadata, qualityGate: false }).replace(
      '### Open questions',
      () => '### Open questions\n\n**Outcome:** `ready-for-change-factory` - stray',
    );
    const { body, fatals, edits } = preparePublishedBody(draft);

    assert.deepEqual(fatals, []);
    assert.equal(edits, 1);
    assert.match(body, /> ⚠️ .*sit outside the ### Quality gate section\. Treat any outcome above as unverified\./);
  });

  test('prose mentioning **Outcome:** mid-sentence is not an outcome line', () => {
    // The match is anchored at the start of the line; unanchored, this forces an outcome on a
    // draft that never stated one.
    const draft = makeDraft().replace(
      '### Open questions',
      () => '### Open questions\n\n- See the **Outcome:** row in the table above for the verdict.',
    );
    const { fatals, warnings, edits } = preparePublishedBody(draft);

    assert.deepEqual(fatals, []);
    assert.deepEqual(warnings, []);
    assert.equal(edits, 0);
  });

  test('a fenced outcome line outside the section is an example, not a stray', () => {
    const metadata = validMetadata();
    const draft = makeDraft({ metadata }).replace(
      '### Open questions',
      () => '### Open questions\n\n```text\n**Outcome:** `ready-for-change-factory` - example\n```',
    );
    const { fatals, warnings, edits } = preparePublishedBody(draft);

    assert.deepEqual(fatals, []);
    assert.deepEqual(warnings, []);
    assert.equal(edits, 0);
  });

  test('an agreeing single outcome line is left untouched', () => {
    const metadata = validMetadata();
    const { body, fatals, warnings, edits } = preparePublishedBody(makeDraft({ metadata }));

    assert.deepEqual(fatals, []);
    assert.deepEqual(warnings, []);
    assert.equal(edits, 0);
    assert.ok(body.includes('**Outcome:** `ready-for-change-factory` - all checklist items pass.'));
  });
});

describe('preparePublishedBody — the annotation degrades instead of disappearing', () => {
  /** A draft sized so exactly `headroom` characters remain under the published-size ceiling. */
  function draftWithHeadroom(headroom) {
    const metadata = validMetadata();
    metadata.gate.outcome = 'research-needs-human';
    const base = makeDraft({ metadata, proseOutcome: 'ready-for-change-factory' });
    const filler = 'x'.repeat(MAX_COMMENT_CHARS - base.length - RESEARCH_COMMENT_MARKER.length - headroom);
    return makeDraft({ metadata, proseOutcome: 'ready-for-change-factory', filler });
  }

  test('falls back to the compact correction when the full annotation would breach the limit', () => {
    // ~200 characters left: the full annotation (~370) does not fit, the correction alone does.
    const { body, fatals, warnings, edits } = preparePublishedBody(draftWithHeadroom(200));

    assert.deepEqual(fatals, []);
    assert.equal(edits, 1);
    assert.ok(body.length <= MAX_COMMENT_CHARS);
    assert.ok(body.includes('**Outcome:** `research-needs-human` - forced by the comment handler'), 'the correction survives');
    assert.ok(!body.includes('⚠️'), 'the appended block is what was dropped');
    assert.match(warnings.join('\n'), /used the compact "annotate quality gate" edit/);
  });

  test('keeps the full annotation when it fits', () => {
    const { body, edits, warnings } = preparePublishedBody(draftWithHeadroom(600));

    assert.equal(edits, 1);
    assert.ok(body.includes('⚠️'));
    assert.ok(!warnings.join('\n').includes('used the compact'));
  });

  test('logs the cause even when neither form fits and the edit is skipped', () => {
    // ~60 characters left: nothing fits. The comment still publishes, and the log still says what
    // disagreed — an operator reading only "skipped … characters" would have no cause to act on.
    const { body, fatals, warnings, edits } = preparePublishedBody(draftWithHeadroom(60));

    assert.deepEqual(fatals, []);
    assert.equal(edits, 0);
    assert.ok(body.length <= MAX_COMMENT_CHARS);
    assert.match(warnings.join('\n'), /disagrees with the machine-readable gate: .*ready-for-change-factory/);
    assert.match(warnings.join('\n'), /skipped the "annotate quality gate" edit/);
  });

  test('logs what disagreed when there is no quality-gate section to annotate', () => {
    const metadata = validMetadata();
    metadata.gate.outcome = 'research-needs-human';
    const { body, warnings } = preparePublishedBody(makeDraft({ metadata, qualityGate: false }));

    assert.ok(body);
    assert.match(warnings.join('\n'), /no ### Quality gate section/);
  });
});

describe('preparePublishedBody — edit priority', () => {
  test('applies both edits when the budget allows', () => {
    const metadata = validMetadata();
    metadata.gate.outcome = 'research-needs-human';
    metadata.gate.rounds = 4;
    const { body, edits } = preparePublishedBody(makeDraft({ metadata, proseOutcome: 'ready-for-change-factory' }));

    assert.equal(edits, 2);
    assert.ok(body.includes('forced by the comment handler'));
    assert.equal(extractMetadata(body).metadata.schema_version, undefined);
  });

  test('when only one edit fits, the annotation wins and the schema_version claim survives', () => {
    // The order decides which edit survives a tight budget. The annotation is bounded and a human
    // reads it; the drop can be arbitrarily large and nothing consumes the metadata today. With
    // the order reversed the drop lands first, the full annotation no longer fits, and the reader
    // loses the ⚠️ — which is the failure this ordering exists to prevent.
    const metadata = validMetadata();
    metadata.schema_version = '1.0'; // a warn row: the drop fires
    metadata.gate.outcome = 'research-needs-human';
    const compact = JSON.stringify(metadata); // re-serialising at indent 2 grows it
    const base = makeDraft({ metadata: compact, proseOutcome: 'ready-for-change-factory' });
    const filler = 'x'.repeat(MAX_COMMENT_CHARS - base.length - RESEARCH_COMMENT_MARKER.length - 500);
    const { body, fatals, warnings, edits } = preparePublishedBody(
      makeDraft({ metadata: compact, proseOutcome: 'ready-for-change-factory', filler }),
    );

    assert.deepEqual(fatals, []);
    assert.equal(edits, 1);
    assert.ok(body.includes('⚠️'), 'the full annotation was applied first');
    assert.equal(extractMetadata(body).metadata.schema_version, '1.0', 'the drop is what got skipped');
    assert.match(warnings.join('\n'), /skipped the "drop schema_version" edit/);
  });
});

describe('update-research-comment CLI', { concurrency: false }, () => {
  // The agent's pre-flight is told "must exit 0". A rejected draft that exits 0 costs an
  // unrecoverable run: the artifact keeps the same draft and the trigger label is already consumed.
  const cli = path.join(path.dirname(fileURLToPath(import.meta.url)), 'update-research-comment.js');

  function run(args) {
    const result = spawnSync(process.execPath, [cli, ...args], { encoding: 'utf8' });
    return { status: result.status, stdout: result.stdout ?? '', stderr: result.stderr ?? '' };
  }

  function draftFile(body) {
    const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'urc-cli-'));
    const file = path.join(dir, 'draft-final.md');
    fs.writeFileSync(file, body);
    return file;
  }

  test('exits 0 and reports the published size for a conformant draft', () => {
    const { status, stdout } = run([draftFile(makeDraft())]);
    assert.equal(status, 0);
    assert.match(stdout, /would publish \d+ of 60000 characters/);
  });

  test('exits 1 and lists every fatal for a draft the job would reject', () => {
    const { status, stderr } = run([draftFile(makeDraft({ approaches: 1, references: false }))]);

    assert.equal(status, 1);
    assert.match(stderr, /would be REJECTED by the publishing job/);
    assert.match(stderr, /needs at least 2/);
    assert.match(stderr, /missing mandatory section: ### References/);
  });

  test('exits 1 on an HTML comment opener, naming the line', () => {
    const { status, stderr } = run([draftFile(makeDraft().replace('### Problem framing', '<!-- note -->\n### Problem framing'))]);
    assert.equal(status, 1);
    assert.match(stderr, /HTML comment opener "<!--" on line \d+/);
  });

  test('exits 0 but reports warnings for a publishable draft with metadata problems', () => {
    const metadata = validMetadata();
    metadata.estimated_scope = 'enormous';
    const { status, stderr } = run([draftFile(makeDraft({ metadata }))]);

    assert.equal(status, 0);
    assert.match(stderr, /warning: pipeline metadata does not satisfy schema_version 1\.1/);
  });

  test('exits 1 on an empty draft and 2 on a usage or read error', () => {
    assert.equal(run([draftFile('   \n')]).status, 1);
    assert.equal(run([]).status, 2);
    assert.equal(run([path.join(os.tmpdir(), 'no-such-draft-4381.md')]).status, 2);
  });
});

describe('resolveDraftPath', () => {
  function layout() {
    const root = fs.mkdtempSync(path.join(os.tmpdir(), 'urc-'));
    const safeJobs = path.join(root, 'gh-aw', 'safe-jobs');
    fs.mkdirSync(path.join(safeJobs, 'agent', 'research'), { recursive: true });
    return { root, safeJobs };
  }

  test('finds the draft at the artifact-relative path', () => {
    const { root, safeJobs } = layout();
    const draft = path.join(safeJobs, 'agent', 'research', 'draft-final.md');
    fs.writeFileSync(draft, 'body');

    const resolved = resolveDraftPath({ RUNNER_TEMP: root });
    assert.equal(resolved.path, draft);
    assert.equal(resolved.fallback, undefined);
  });

  test('derives the download directory from GH_AW_AGENT_OUTPUT', () => {
    const { safeJobs } = layout();
    const draft = path.join(safeJobs, 'agent', 'research', 'draft-final.md');
    fs.writeFileSync(draft, 'body');

    const resolved = resolveDraftPath({ GH_AW_AGENT_OUTPUT: path.join(safeJobs, 'agent_output.json') });
    assert.equal(resolved.path, draft);
  });

  test('falls back to a bounded search when the layout changes', () => {
    const { root, safeJobs } = layout();
    const moved = path.join(safeJobs, 'nested', 'elsewhere');
    fs.mkdirSync(moved, { recursive: true });
    const draft = path.join(moved, 'draft-final.md');
    fs.writeFileSync(draft, 'body');

    const resolved = resolveDraftPath({ RUNNER_TEMP: root });
    assert.equal(resolved.path, draft);
    assert.deepEqual(resolved.fallback, [draft]);
  });

  test('reports the candidates it tried when nothing is found', () => {
    const { root } = layout();
    const resolved = resolveDraftPath({ RUNNER_TEMP: root });

    assert.match(resolved.error, /no draft-final\.md/);
    assert.equal(resolved.candidates.length, 2);
    assert.ok(resolved.candidates[0].endsWith(path.join('safe-jobs', 'agent', 'research', 'draft-final.md')));
    assert.equal(resolved.candidates[1], path.join('/tmp/gh-aw/agent/research', 'draft-final.md'));
  });
});

describe('update-research-comment runner', { concurrency: false }, () => {
  function mockCore() {
    const outputs = {};
    const failed = [];
    const warnings = [];
    return {
      outputs,
      failed,
      warnings,
      core: {
        setOutput(key, value) {
          outputs[key] = value;
        },
        setFailed(message) {
          failed.push(message);
        },
        warning(message) {
          warnings.push(message);
        },
        info() {},
      },
    };
  }

  function mockGithub(existingComments = []) {
    const created = [];
    const updated = [];
    return {
      created,
      updated,
      github: {
        rest: {
          issues: {
            listComments: 'listComments',
            async createComment(params) {
              created.push(params);
              return { data: { id: 999 } };
            },
            async updateComment(params) {
              updated.push(params);
              return { data: { id: params.comment_id } };
            },
          },
        },
        async paginate() {
          return existingComments;
        },
      },
    };
  }

  const context = { repo: { owner: 'elastic', repo: 'terraform-provider-ec' } };

  async function withEnv(vars, fn) {
    const previous = {};
    for (const key of Object.keys(vars)) {
      previous[key] = process.env[key];
      if (vars[key] === undefined) delete process.env[key];
      else process.env[key] = vars[key];
    }
    try {
      return await fn();
    } finally {
      for (const key of Object.keys(previous)) {
        if (previous[key] === undefined) delete process.env[key];
        else process.env[key] = previous[key];
      }
    }
  }

  function artifactWith(draft) {
    const root = fs.mkdtempSync(path.join(os.tmpdir(), 'urc-run-'));
    const researchDir = path.join(root, 'gh-aw', 'safe-jobs', 'agent', 'research');
    fs.mkdirSync(researchDir, { recursive: true });
    if (draft !== null) fs.writeFileSync(path.join(researchDir, 'draft-final.md'), draft);
    return root;
  }

  const cleanEnv = (root, overrides = {}) => ({
    RUNNER_TEMP: root,
    GH_AW_AGENT_OUTPUT: undefined,
    DETECTION_RESULT: 'success',
    GH_AW_DETECTION_CONCLUSION: 'success',
    INPUT_ISSUE_NUMBER: '42',
    GH_AW_BOT_LOGIN: undefined,
    ...overrides,
  });

  test('creates the sticky comment when none exists', async () => {
    const root = artifactWith(makeDraft());
    const { core, failed, outputs } = mockCore();
    const { github, created, updated } = mockGithub([]);

    await withEnv(cleanEnv(root), () => handler({ github, context, core }));

    assert.deepEqual(failed, []);
    assert.equal(updated.length, 0);
    assert.equal(created.length, 1);
    assert.equal(created[0].issue_number, 42);
    assert.ok(created[0].body.startsWith(RESEARCH_COMMENT_MARKER));
    assert.equal(outputs.research_comment_published, 'true');
  });

  test('updates the existing sticky comment instead of posting a second one', async () => {
    const root = artifactWith(makeDraft());
    const { core, failed } = mockCore();
    const { github, created, updated } = mockGithub([
      { id: 7, user: { login: 'github-actions[bot]' }, body: `${RESEARCH_COMMENT_MARKER}\nold research` },
    ]);

    await withEnv(cleanEnv(root), () => handler({ github, context, core }));

    assert.deepEqual(failed, []);
    assert.equal(created.length, 0);
    assert.equal(updated.length, 1);
    assert.equal(updated[0].comment_id, 7);
  });

  test('falls back to the default bot login when GH_AW_BOT_LOGIN is set but empty', async () => {
    const root = artifactWith(makeDraft());
    const { core } = mockCore();
    const { github, created, updated } = mockGithub([
      { id: 7, user: { login: 'github-actions[bot]' }, body: `${RESEARCH_COMMENT_MARKER}\nold research` },
    ]);

    await withEnv(cleanEnv(root, { GH_AW_BOT_LOGIN: '' }), () => handler({ github, context, core }));

    assert.equal(created.length, 0);
    assert.equal(updated.length, 1);
  });

  test('fails closed on a detection warning', async () => {
    const root = artifactWith(makeDraft());
    const { core, failed } = mockCore();
    const { github, created, updated } = mockGithub([]);

    await withEnv(cleanEnv(root, { GH_AW_DETECTION_CONCLUSION: 'warning' }), () =>
      handler({ github, context, core }),
    );

    assert.equal(created.length + updated.length, 0);
    assert.match(failed[0], /refusing to publish: detection/);
  });

  test('fails closed when the detection result is missing entirely', async () => {
    const root = artifactWith(makeDraft());
    const { core, failed } = mockCore();
    const { github, created } = mockGithub([]);

    await withEnv(cleanEnv(root, { DETECTION_RESULT: undefined, GH_AW_DETECTION_CONCLUSION: undefined }), () =>
      handler({ github, context, core }),
    );

    assert.equal(created.length, 0);
    assert.match(failed[0], /refusing to publish: detection/);
  });

  test('rejects a non-integer dispatch issue number instead of stamping #1', async () => {
    const root = artifactWith(makeDraft());
    const { core, failed } = mockCore();
    const { github, created } = mockGithub([]);

    await withEnv(cleanEnv(root, { INPUT_ISSUE_NUMBER: '1.5' }), () => handler({ github, context, core }));

    assert.equal(created.length, 0);
    assert.match(failed[0], /invalid issue number '1\.5'/);
  });

  test('fails with a redaction hint when the draft is absent', async () => {
    const root = artifactWith(null);
    const { core, failed } = mockCore();
    const { github, created } = mockGithub([]);

    await withEnv(cleanEnv(root), () => handler({ github, context, core }));

    assert.equal(created.length, 0);
    assert.match(failed[0], /no draft-final\.md in the agent artifact/);
    assert.match(failed[0], /secret redaction failed/);
  });

  test('does not publish a rejected draft and says a re-run will not fix it', async () => {
    const root = artifactWith(makeDraft({ approaches: 1 }));
    const { core, failed } = mockCore();
    const { github, created, updated } = mockGithub([]);

    await withEnv(cleanEnv(root), () => handler({ github, context, core }));

    assert.equal(created.length + updated.length, 0);
    assert.match(failed[0], /refusing to publish/);
    assert.match(failed[0], /re-apply the research-factory label/);
  });

  test('surfaces metadata problems as warnings and still publishes', async () => {
    const metadata = validMetadata();
    metadata.gate.critic.status = 'broken';
    const root = artifactWith(makeDraft({ metadata }));
    const { core, failed, warnings } = mockCore();
    const { github, created } = mockGithub([]);

    await withEnv(cleanEnv(root), () => handler({ github, context, core }));

    assert.deepEqual(failed, []);
    assert.equal(created.length, 1);
    assert.match(warnings.join('\n'), /does not satisfy schema_version 1\.1/);
  });

  test('fails when listing comments throws, before writing anything', async () => {
    const root = artifactWith(makeDraft());
    const { core, failed } = mockCore();
    const { github, created } = mockGithub([]);
    github.paginate = async () => {
      throw new Error('rate limited');
    };

    await withEnv(cleanEnv(root), () => handler({ github, context, core }));

    assert.equal(created.length, 0);
    assert.match(failed[0], /cannot list comments on #42: rate limited/);
  });

  test('fails when the publish call throws', async () => {
    const root = artifactWith(makeDraft());
    const { core, failed } = mockCore();
    const { github } = mockGithub([]);
    github.rest.issues.createComment = async () => {
      throw new Error('403');
    };

    await withEnv(cleanEnv(root), () => handler({ github, context, core }));

    assert.match(failed[0], /failed to publish the research comment on #42: 403/);
  });
});
