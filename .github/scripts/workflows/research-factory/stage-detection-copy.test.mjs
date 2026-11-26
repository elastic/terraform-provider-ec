import assert from 'node:assert/strict';
import test, { after, describe } from 'node:test';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const stage = require('./stage-detection-copy.js');
const {
  DETECTION_COPY_MAX_BYTES,
  DETECTION_COPY_DIRNAME,
  DETECTION_COPY_FILENAME,
  PUBLISH_OUTPUT_TYPE,
  DEFAULT_PATHS,
  elisionNotice,
  sliceTailToBytes,
  boundDetectionCopy,
  stageDetectionCopy,
} = stage;
const { MAX_ENV_PAYLOAD_BYTES, byteLength } = require('../lib/env-payload.js');
const { preparePublishedBody, DRAFT_FINAL } = require('./update-research-comment.js');
const { RESEARCH_COMMENT_MARKER } = require('../lib/intake/research-factory-constants.js');

const NOTICE_PATTERN = /\n\n\[\.\.\. detection copy truncated: (\d+) of (\d+) bytes of the published comment omitted here \.\.\.\]\n\n/g;
const LONE_SURROGATE = /[\uD800-\uDBFF](?![\uDC00-\uDFFF])|(?<![\uD800-\uDBFF])[\uDC00-\uDFFF]/;

/**
 * A minimal draft the publish handler accepts unedited. `filler` is inserted into the problem
 * framing, `proseOutcome` replaces the outcome stated in the `### Quality gate` prose (the metadata
 * keeps `research-needs-human`).
 *
 * @param {{filler?: string, proseOutcome?: string}} [opts]
 */
function localDraft({ filler = '', proseOutcome = 'research-needs-human' } = {}) {
  return [
    '## Implementation research',
    '',
    'Run: local test',
    '',
    '> Edits inside this comment are read as input on the next run but are not preserved verbatim. For',
    '> durable feedback, post a comment or edit the issue body.',
    '',
    '### Problem framing',
    '',
    `The issue asks for a new optional attribute on an existing resource.${filler}`,
    '',
    '### Approaches considered',
    '',
    '#### Approach A - add the attribute to the existing resource',
    '',
    'Terraform shape: `example` (string, optional).',
    '',
    '#### Approach B - a separate resource',
    '',
    'Terraform shape: a new resource with `example` (string, required).',
    '',
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
    '### Quality gate',
    '',
    `**Outcome:** \`${proseOutcome}\` - no critic ran.`,
    '',
    '### References',
    '',
    '- `ec/provider.go`',
    '',
    '<details>',
    '<summary>🤖 Pipeline metadata</summary>',
    '',
    '```json',
    JSON.stringify(
      {
        schema_version: '1.1',
        recommendation: { spine: 'approach-a', confidence: 'low', approach_index: 0 },
        open_questions: [],
        affected_capabilities: ['example'],
        estimated_scope: 'small',
        references: [{ type: 'repo-path', path: 'ec/provider.go' }],
        gate: {
          outcome: 'research-needs-human',
          checklist: { grounded: false, mapped: false, compatible: false, versioned: false, testable: false, idiomatic: false },
          score: null,
          scores: [],
          converged: false,
          rounds: 0,
          outstanding_feedback: ['No critic ran.'],
          author_model: 'anthropic/claude-sonnet-5',
          critic: { model: 'openai/gpt-6.1-sol', status: 'unavailable' },
        },
      },
      null,
      2,
    ),
    '```',
    '',
    '</details>',
    '',
  ].join('\n');
}

const tempDirs = [];
function tempDir(prefix) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), prefix));
  tempDirs.push(dir);
  return dir;
}
after(() => {
  for (const dir of tempDirs) {
    fs.rmSync(dir, { recursive: true, force: true });
  }
});

/** A temp download root with the draft at the canonical path (or nowhere, for `null`), and a temp detection dir. */
function layout(draft) {
  const sourceRoot = tempDir('sdc-src-');
  const detectionDir = tempDir('sdc-det-');
  if (draft !== null) {
    const dir = path.join(sourceRoot, 'agent', 'research');
    fs.mkdirSync(dir, { recursive: true });
    fs.writeFileSync(path.join(dir, DRAFT_FINAL), draft);
  }
  return { sourceRoot, detectionDir };
}

const stagedDir = (detectionDir) => path.join(detectionDir, DETECTION_COPY_DIRNAME);
const stagedFile = (detectionDir) => path.join(stagedDir(detectionDir), DETECTION_COPY_FILENAME);

/** A conformant draft whose published body is over the byte budget but under the character limit. */
function overBudgetDraft() {
  return localDraft({ filler: ` ${'漢'.repeat(35_000)}` });
}

function mockCore() {
  const calls = { info: [], warning: [], setFailed: [], outputs: {} };
  return {
    calls,
    core: {
      info: (m) => calls.info.push(String(m)),
      warning: (m) => calls.warning.push(String(m)),
      setFailed: (m) => calls.setFailed.push(String(m)),
      setOutput: (k, v) => {
        calls.outputs[k] = v;
      },
    },
  };
}

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

describe('localDraft', () => {
  test('is a draft the publish handler accepts', () => {
    const prepared = preparePublishedBody(localDraft());
    assert.deepEqual(prepared.fatals, []);
    assert.equal(typeof prepared.body, 'string');
  });
});

describe('boundDetectionCopy', () => {
  test('the budget is MAX_ENV_PAYLOAD_BYTES', () => {
    assert.equal(DETECTION_COPY_MAX_BYTES, MAX_ENV_PAYLOAD_BYTES);
  });

  test('returns a body within budget unchanged', () => {
    const text = 'a'.repeat(60_000);
    const result = boundDetectionCopy(text);
    assert.equal(result.text, text);
    assert.equal(result.truncated, false);
    assert.equal(result.omittedBytes, 0);
    assert.equal(result.totalBytes, 60_000);
  });

  test('bounds the copy in bytes, not characters', () => {
    const text = '漢'.repeat(60_000);
    const result = boundDetectionCopy(text);
    assert.equal(result.truncated, true);
    assert.equal(result.totalBytes, 180_000);
    assert.ok(byteLength(result.text) <= DETECTION_COPY_MAX_BYTES, `${byteLength(result.text)} bytes`);
  });

  test('marks the omitted span with one notice that accounts for every byte', () => {
    const text = `${'a'.repeat(50_000)}${'漢'.repeat(30_000)}${'b'.repeat(20_000)}`;
    const result = boundDetectionCopy(text);
    const notices = [...result.text.matchAll(NOTICE_PATTERN)];
    assert.equal(notices.length, 1);

    const [notice] = notices;
    const prefix = result.text.slice(0, notice.index);
    const suffix = result.text.slice(notice.index + notice[0].length);
    assert.ok(text.startsWith(prefix));
    assert.ok(text.endsWith(suffix));
    assert.equal(Number(notice[1]), result.totalBytes - byteLength(prefix) - byteLength(suffix));
    assert.equal(Number(notice[1]), result.omittedBytes);
    assert.equal(Number(notice[2]), result.totalBytes);
  });

  test('never splits a code point', () => {
    const decoder = new TextDecoder('utf-8', { fatal: true });
    for (const text of ['😀'.repeat(40_000), `x${'😀漢é'.repeat(20_000)}y`]) {
      const result = boundDetectionCopy(text);
      assert.equal(result.truncated, true);
      assert.doesNotThrow(() => decoder.decode(Buffer.from(result.text, 'utf8')));
      assert.doesNotMatch(result.text, LONE_SURROGATE);
      assert.ok(byteLength(result.text) <= DETECTION_COPY_MAX_BYTES);
    }
  });

  test('keeps the end of the comment', () => {
    const sentinel = 'END-OF-COMMENT-SENTINEL-7f3a';
    const result = boundDetectionCopy(`${'漢'.repeat(40_000)}\n${sentinel}`);
    assert.equal(result.truncated, true);
    assert.ok(result.text.endsWith(sentinel));
  });

  test('honours a small explicit budget and refuses one smaller than its notice', () => {
    const text = 'z'.repeat(10_000);
    const small = boundDetectionCopy(text, 400);
    assert.equal(small.truncated, true);
    assert.ok(byteLength(small.text) <= 400);
    assert.throws(() => boundDetectionCopy(text, byteLength(elisionNotice(10_000, 10_000))), RangeError);
    assert.throws(() => boundDetectionCopy(text, 20), RangeError);
  });

  test('is deterministic', () => {
    const text = `${'漢'.repeat(30_000)}${'q'.repeat(30_000)}`;
    assert.deepEqual(boundDetectionCopy(text), boundDetectionCopy(text));
  });
});

describe('sliceTailToBytes', () => {
  test('returns the exact suffix of ASCII', () => {
    assert.equal(sliceTailToBytes('abcdef', 3), 'def');
    assert.equal(sliceTailToBytes('abc', 10), 'abc');
  });

  test('keeps whole code points of multi-byte text', () => {
    assert.equal(sliceTailToBytes('a漢字', 5), '字');
    assert.equal(sliceTailToBytes('a漢字', 6), '漢字');
    assert.equal(sliceTailToBytes('x😀😀', 7), '😀');
  });

  test('returns an empty string for a budget of zero', () => {
    assert.equal(sliceTailToBytes('abc', 0), '');
  });
});

describe('stageDetectionCopy', () => {
  test('stages exactly the string the publish handler posts', () => {
    const draft = localDraft();
    const { sourceRoot, detectionDir } = layout(draft);

    const result = stageDetectionCopy({ sourceRoot, detectionDir });

    assert.equal(result.status, 'staged');
    assert.equal(result.stagedPath, stagedFile(detectionDir));
    assert.equal(result.truncated, false);
    const staged = fs.readFileSync(result.stagedPath);
    assert.equal(Buffer.compare(staged, Buffer.from(preparePublishedBody(draft).body, 'utf8')), 0);
    assert.ok(staged.toString('utf8').startsWith(RESEARCH_COMMENT_MARKER));
    assert.equal(result.stagedBytes, staged.length);
  });

  test('stages the transformed body, not the draft', () => {
    const draft = localDraft({ proseOutcome: 'ready-for-change-factory' });
    const { sourceRoot, detectionDir } = layout(draft);

    const result = stageDetectionCopy({ sourceRoot, detectionDir });

    assert.equal(result.status, 'staged');
    const staged = fs.readFileSync(result.stagedPath, 'utf8');
    assert.match(staged, /\*\*Outcome:\*\* `research-needs-human` - forced by the comment handler/);
    assert.notEqual(staged, draft);
  });

  test('stages the body without invisible characters', () => {
    const draft = localDraft({ filler: ' zero​width' });
    assert.ok(draft.includes('​'));
    const { sourceRoot, detectionDir } = layout(draft);

    const result = stageDetectionCopy({ sourceRoot, detectionDir });

    assert.equal(result.status, 'staged');
    assert.ok(!fs.readFileSync(result.stagedPath, 'utf8').includes('​'));
  });

  test('stages nothing when there is no draft at the canonical path', () => {
    const { sourceRoot, detectionDir } = layout(null);

    const result = stageDetectionCopy({ sourceRoot, detectionDir });

    assert.equal(result.status, 'no-canonical-draft');
    assert.equal(result.canonicalPath, path.join(sourceRoot, 'agent', 'research', 'draft-final.md'));
    assert.equal(fs.existsSync(stagedDir(detectionDir)), false);
  });

  test('stages nothing for a draft the publish handler rejects', () => {
    const { sourceRoot, detectionDir } = layout(`${localDraft()}\n<!-- stray opener\n`);

    const result = stageDetectionCopy({ sourceRoot, detectionDir });

    assert.equal(result.status, 'rejected');
    assert.ok(result.fatals.length > 0);
    assert.equal(fs.existsSync(stagedDir(detectionDir)), false);
  });

  test('stages only the canonical path, never a draft found by searching', () => {
    const { sourceRoot, detectionDir } = layout(null);
    const elsewhere = path.join(sourceRoot, 'sandbox', 'firewall', 'logs');
    fs.mkdirSync(elsewhere, { recursive: true });
    fs.writeFileSync(path.join(elsewhere, DRAFT_FINAL), localDraft());

    const result = stageDetectionCopy({ sourceRoot, detectionDir });

    assert.equal(result.status, 'no-canonical-draft');
    assert.equal(typeof result.foundElsewhere, 'string');
    assert.ok(result.foundElsewhere.length > 0);
    assert.equal(fs.existsSync(stagedDir(detectionDir)), false);
  });

  test('stages a bounded copy with a notice when the body is over budget', () => {
    const draft = overBudgetDraft();
    const prepared = preparePublishedBody(draft);
    assert.equal(typeof prepared.body, 'string', prepared.fatals.join('; '));
    assert.ok(byteLength(prepared.body) > DETECTION_COPY_MAX_BYTES);
    const { sourceRoot, detectionDir } = layout(draft);

    const result = stageDetectionCopy({ sourceRoot, detectionDir });

    assert.equal(result.status, 'staged');
    assert.equal(result.truncated, true);
    assert.ok(fs.statSync(result.stagedPath).size <= DETECTION_COPY_MAX_BYTES);
    assert.match(fs.readFileSync(result.stagedPath, 'utf8'), /detection copy truncated/);
  });

  test('replaces an existing copy', () => {
    const draft = localDraft();
    const { sourceRoot, detectionDir } = layout(draft);
    fs.mkdirSync(stagedDir(detectionDir), { recursive: true });
    fs.writeFileSync(stagedFile(detectionDir), 'stale');

    const result = stageDetectionCopy({ sourceRoot, detectionDir });

    assert.equal(result.status, 'staged');
    assert.equal(fs.readFileSync(stagedFile(detectionDir), 'utf8'), preparePublishedBody(draft).body);
  });

  test('refuses a comment-memory directory that is a symlink', () => {
    const { sourceRoot, detectionDir } = layout(localDraft());
    const target = tempDir('sdc-link-');
    fs.symlinkSync(target, stagedDir(detectionDir));

    assert.throws(() => stageDetectionCopy({ sourceRoot, detectionDir }), /is not a directory/);
    assert.deepEqual(fs.readdirSync(target), []);
  });

  test('the staged file name is not the name the handler searches for', () => {
    assert.notEqual(DETECTION_COPY_FILENAME, DRAFT_FINAL);
  });
});

describe('stage-detection-copy entry', { concurrency: false }, () => {
  test('stages a publishable draft without logging its text', async () => {
    const sentinel = 'DRAFT-TEXT-SENTINEL-41c9';
    const { sourceRoot, detectionDir } = layout(localDraft({ filler: ` ${sentinel}` }));
    const { core, calls } = mockCore();

    await stage({ core }, { paths: { sourceRoot, detectionDir }, outputTypes: PUBLISH_OUTPUT_TYPE });

    assert.deepEqual(calls.setFailed, []);
    assert.equal(calls.outputs.detection_copy_status, 'staged');
    assert.equal(calls.info.length, 1);
    for (const line of [...calls.info, ...calls.warning, ...calls.setFailed]) {
      assert.ok(!line.includes(sentinel));
    }
    assert.ok(fs.readFileSync(stagedFile(detectionDir), 'utf8').includes(sentinel));
  });

  test('stages a publishable draft even when no comment was requested', async () => {
    const { sourceRoot, detectionDir } = layout(localDraft());
    const { core, calls } = mockCore();

    await stage({ core }, { paths: { sourceRoot, detectionDir }, outputTypes: 'noop' });

    assert.deepEqual(calls.setFailed, []);
    assert.equal(calls.outputs.detection_copy_status, 'staged');
    assert.ok(fs.existsSync(stagedFile(detectionDir)));
  });

  for (const [label, outputTypes] of [
    ['undefined', undefined],
    ['empty', ''],
    ['blank', '  '],
  ]) {
    test(`fails when the output types are ${label}`, async () => {
      const { sourceRoot, detectionDir } = layout(localDraft());
      const { core, calls } = mockCore();

      await withEnv({ AGENT_OUTPUT_TYPES: undefined }, () => stage({ core }, { paths: { sourceRoot, detectionDir }, outputTypes }));

      assert.equal(calls.setFailed.length, 1);
      assert.match(calls.setFailed[0], /AGENT_OUTPUT_TYPES is empty/);
      assert.equal(fs.existsSync(stagedDir(detectionDir)), false);
    });
  }

  test('reads the output types from AGENT_OUTPUT_TYPES', async () => {
    const { sourceRoot, detectionDir } = layout(localDraft());
    const { core, calls } = mockCore();

    await withEnv({ AGENT_OUTPUT_TYPES: PUBLISH_OUTPUT_TYPE }, () => stage({ core }, { paths: { sourceRoot, detectionDir } }));

    assert.deepEqual(calls.setFailed, []);
    assert.equal(calls.outputs.detection_copy_status, 'staged');
  });

  test('fails when a comment was requested and there is no canonical draft', async () => {
    const { sourceRoot, detectionDir } = layout(null);
    const { core, calls } = mockCore();

    await stage({ core }, { paths: { sourceRoot, detectionDir }, outputTypes: `noop,${PUBLISH_OUTPUT_TYPE}` });

    assert.equal(calls.setFailed.length, 1);
    assert.ok(calls.setFailed[0].includes(path.join(sourceRoot, 'agent', 'research', 'draft-final.md')));
    assert.equal(calls.outputs.detection_copy_status, 'no-canonical-draft');
  });

  test('fails when a comment was requested and the draft is only found by searching', async () => {
    const { sourceRoot, detectionDir } = layout(null);
    const elsewhere = path.join(sourceRoot, 'sandbox', 'firewall', 'logs');
    fs.mkdirSync(elsewhere, { recursive: true });
    fs.writeFileSync(path.join(elsewhere, DRAFT_FINAL), localDraft());
    const { core, calls } = mockCore();

    await stage({ core }, { paths: { sourceRoot, detectionDir }, outputTypes: PUBLISH_OUTPUT_TYPE });

    assert.equal(calls.setFailed.length, 1);
    assert.ok(calls.setFailed[0].includes(path.join(sourceRoot, 'agent', 'research', 'draft-final.md')));
    assert.equal(calls.outputs.detection_copy_status, 'no-canonical-draft');
    assert.equal(fs.existsSync(stagedDir(detectionDir)), false);
  });

  test('does not fail when there is no draft and no comment was requested', async () => {
    const { sourceRoot, detectionDir } = layout(null);
    const { core, calls } = mockCore();

    await stage({ core }, { paths: { sourceRoot, detectionDir }, outputTypes: 'noop' });

    assert.deepEqual(calls.setFailed, []);
    assert.equal(calls.outputs.detection_copy_status, 'no-canonical-draft');
  });

  test('does not fail for a draft the publish job will reject', async () => {
    const { sourceRoot, detectionDir } = layout(`${localDraft()}\n<!-- stray opener\n`);
    const { core, calls } = mockCore();

    await stage({ core }, { paths: { sourceRoot, detectionDir }, outputTypes: PUBLISH_OUTPUT_TYPE });

    assert.deepEqual(calls.setFailed, []);
    assert.equal(calls.outputs.detection_copy_status, 'rejected');
  });

  test('fails for an over-budget comment and still stages the bounded copy', async () => {
    const { sourceRoot, detectionDir } = layout(overBudgetDraft());
    const { core, calls } = mockCore();

    await stage({ core }, { paths: { sourceRoot, detectionDir }, outputTypes: PUBLISH_OUTPUT_TYPE });

    assert.equal(calls.setFailed.length, 1);
    assert.match(calls.setFailed[0], /over the 102400-byte detection budget; it will not be published/);
    assert.match(fs.readFileSync(stagedFile(detectionDir), 'utf8'), /detection copy truncated/);
  });

  test('fails without throwing when the copy cannot be written', async () => {
    const { sourceRoot, detectionDir } = layout(localDraft());
    fs.symlinkSync(tempDir('sdc-link-'), stagedDir(detectionDir));
    const { core, calls } = mockCore();

    await assert.doesNotReject(stage({ core }, { paths: { sourceRoot, detectionDir }, outputTypes: PUBLISH_OUTPUT_TYPE }));

    assert.equal(calls.setFailed.length, 1);
    assert.match(calls.setFailed[0], /cannot stage the research comment body for detection/);
  });

  test('defaults to the detection job paths', () => {
    assert.deepEqual({ ...DEFAULT_PATHS }, { sourceRoot: '/tmp/gh-aw', detectionDir: '/tmp/gh-aw/threat-detection' });
  });
});

describe('source guard', () => {
  test('neither the staging step nor the publish handler calls sanitizeUserContent', () => {
    for (const file of ['stage-detection-copy.js', 'update-research-comment.js']) {
      const source = fs.readFileSync(new URL(`./${file}`, import.meta.url), 'utf8');
      assert.doesNotMatch(source, /sanitizeUserContent\s*\(/, file);
    }
  });
});
