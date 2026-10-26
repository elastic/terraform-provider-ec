import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { validateMetadata, validateCommentBody } = require('./validate-metadata.js');

const here = path.dirname(fileURLToPath(import.meta.url));
const cli = path.join(here, 'validate-metadata.js');

function validMetadata() {
  return {
    schema_version: '1.1',
    recommendation: { spine: 'extend-ec-deployment-schema', confidence: 'medium', approach_index: 0 },
    open_questions: [{ id: 'q1', text: 'Is the field exposed on ECE?', blocking: false }],
    affected_capabilities: ['ec_deployment'],
    estimated_scope: 'medium',
    references: [
      { type: 'repo-path', path: 'ec/ecresource/deploymentresource/schema.go' },
      { type: 'elastic-docs', url: 'https://www.elastic.co/docs/api/doc/cloud/operation/operation-get-deployment' },
    ],
    gate: {
      outcome: 'ready-for-change-factory',
      checklist: {
        grounded: true,
        mapped: true,
        compatible: true,
        versioned: true,
        testable: true,
        idiomatic: true,
      },
      score: 88,
      scores: [82, 88],
      converged: true,
      rounds: 2,
      outstanding_feedback: [],
      author_model: 'anthropic/claude-sonnet-5',
      critic: { model: 'openai/gpt-6.1-sol', status: 'ok' },
    },
  };
}

function errorsFor(mutate) {
  const metadata = validMetadata();
  mutate(metadata);
  return validateMetadata(metadata).errors;
}

function bodyWith(metadata) {
  return [
    '## Implementation research',
    '',
    '### References',
    '- none',
    '',
    '<details><summary>🤖 Pipeline metadata</summary>',
    '',
    '```json',
    JSON.stringify(metadata, null, 2),
    '```',
    '',
    '</details>',
    '',
  ].join('\n');
}

describe('validateMetadata', () => {
  test('accepts a conformant 1.1 block', () => {
    assert.deepEqual(validateMetadata(validMetadata()).errors, []);
  });

  test('accepts the minimal block: optional members omitted', () => {
    const metadata = validMetadata();
    delete metadata.open_questions;
    delete metadata.affected_capabilities;
    delete metadata.references;
    delete metadata.recommendation.confidence;
    assert.deepEqual(validateMetadata(metadata).errors, []);
  });

  test('rejects a non-object', () => {
    assert.deepEqual(validateMetadata(null).errors, ['metadata: must be a JSON object']);
    assert.deepEqual(validateMetadata([]).errors, ['metadata: must be a JSON object']);
  });

  test('rejects a schema_version other than 1.1', () => {
    assert.match(errorsFor((m) => (m.schema_version = '1.0')).join('\n'), /schema_version: must be "1\.1"/);
    assert.match(errorsFor((m) => delete m.schema_version).join('\n'), /schema_version/);
  });

  test('rejects a non-kebab-case spine', () => {
    for (const spine of ['Extend_Schema', 'extend schema', '-leading', 'trailing-', '']) {
      assert.match(errorsFor((m) => (m.recommendation.spine = spine)).join('\n'), /recommendation\.spine/);
    }
  });

  test('rejects a stringified or negative approach_index', () => {
    assert.match(errorsFor((m) => (m.recommendation.approach_index = '0')).join('\n'), /approach_index/);
    assert.match(errorsFor((m) => (m.recommendation.approach_index = 1.5)).join('\n'), /approach_index/);
    assert.match(errorsFor((m) => (m.recommendation.approach_index = -1)).join('\n'), /approach_index/);
    assert.match(errorsFor((m) => delete m.recommendation.approach_index).join('\n'), /approach_index/);
  });

  test('accepts approach_index 0 and rejects an unknown confidence', () => {
    assert.deepEqual(errorsFor((m) => (m.recommendation.approach_index = 0)), []);
    assert.match(errorsFor((m) => (m.recommendation.confidence = 'certain')).join('\n'), /confidence/);
  });

  test('rejects an unknown estimated_scope', () => {
    assert.match(errorsFor((m) => (m.estimated_scope = 'xl')).join('\n'), /estimated_scope/);
    assert.match(errorsFor((m) => delete m.estimated_scope).join('\n'), /estimated_scope/);
  });

  test('rejects an unknown reference type and a reference with no location', () => {
    assert.match(errorsFor((m) => (m.references[0].type = 'url')).join('\n'), /references\[0\]\.type/);
    assert.match(errorsFor((m) => (m.references[0].type = 'file')).join('\n'), /references\[0\]\.type/);
    assert.match(
      errorsFor((m) => (m.references[0] = { type: 'repo-path', path: '  ' })).join('\n'),
      /references\[0\]: must carry a non-empty url or path/,
    );
  });

  test('rejects malformed open questions', () => {
    assert.match(errorsFor((m) => (m.open_questions[0].blocking = 'false')).join('\n'), /open_questions\[0\]\.blocking/);
    assert.match(errorsFor((m) => delete m.open_questions[0].id).join('\n'), /open_questions\[0\]\.id/);
    assert.match(errorsFor((m) => (m.open_questions[0].text = '')).join('\n'), /open_questions\[0\]\.text/);
    assert.match(errorsFor((m) => (m.open_questions = {})).join('\n'), /open_questions: must be an array/);
  });

  test('rejects a capitalised checklist key', () => {
    const errors = errorsFor((m) => {
      delete m.gate.checklist.grounded;
      m.gate.checklist.Grounded = true;
    });
    assert.match(errors.join('\n'), /gate\.checklist\.grounded: must be a boolean/);
    assert.match(errors.join('\n'), /unknown keys Grounded/);
  });

  test('rejects a non-boolean checklist value and a missing checklist', () => {
    assert.match(errorsFor((m) => (m.gate.checklist.mapped = 'true')).join('\n'), /gate\.checklist\.mapped/);
    assert.match(errorsFor((m) => delete m.gate.checklist).join('\n'), /gate\.checklist: must be an object/);
  });

  test('rejects rounds out of step with scores.length', () => {
    assert.match(errorsFor((m) => (m.gate.rounds = 3)).join('\n'), /does not equal gate\.scores\.length \(2\)/);
    assert.match(errorsFor((m) => m.gate.scores.push(91)).join('\n'), /does not equal gate\.scores\.length \(3\)/);
  });

  test('rejects rounds beyond the 5-round cap and non-integer rounds', () => {
    assert.match(errorsFor((m) => (m.gate.rounds = 6)).join('\n'), /gate\.rounds: must be an integer 0-5/);
    assert.match(errorsFor((m) => (m.gate.rounds = 2.0001)).join('\n'), /gate\.rounds: must be an integer 0-5/);
  });

  test('accepts a zero-round gate with a null score', () => {
    const errors = errorsFor((m) => {
      m.gate.rounds = 0;
      m.gate.scores = [];
      m.gate.score = null;
      m.gate.converged = false;
      m.gate.outcome = 'research-needs-human';
      m.gate.critic.status = 'unavailable';
    });
    assert.deepEqual(errors, []);
  });

  test('rejects out-of-range scores', () => {
    assert.match(errorsFor((m) => (m.gate.scores = [82, 101])).join('\n'), /gate\.scores\[1\]/);
    assert.match(errorsFor((m) => (m.gate.scores = [82, '88'])).join('\n'), /gate\.scores\[1\]/);
    assert.match(errorsFor((m) => (m.gate.score = -1)).join('\n'), /gate\.score/);
    assert.match(errorsFor((m) => (m.gate.score = 'n/a')).join('\n'), /gate\.score/);
  });

  test('rejects an unknown outcome and critic status', () => {
    assert.match(errorsFor((m) => (m.gate.outcome = 'ready')).join('\n'), /gate\.outcome/);
    assert.match(errorsFor((m) => (m.gate.critic.status = 'OK')).join('\n'), /gate\.critic\.status/);
    assert.match(errorsFor((m) => delete m.gate.critic).join('\n'), /gate\.critic: must be an object/);
  });

  test('rejects an empty author_model and non-string outstanding feedback', () => {
    assert.match(errorsFor((m) => (m.gate.author_model = '')).join('\n'), /gate\.author_model/);
    assert.match(errorsFor((m) => (m.gate.outstanding_feedback = [{ text: 'x' }])).join('\n'), /outstanding_feedback\[0\]/);
    assert.match(errorsFor((m) => (m.gate.outstanding_feedback = 'none')).join('\n'), /outstanding_feedback: must be an array/);
  });

  test('rejects a non-boolean converged flag', () => {
    assert.match(errorsFor((m) => (m.gate.converged = 'true')).join('\n'), /gate\.converged/);
  });

  test('rejects non-string affected capabilities', () => {
    assert.match(errorsFor((m) => (m.affected_capabilities = [1])).join('\n'), /affected_capabilities\[0\]/);
    assert.match(errorsFor((m) => (m.affected_capabilities = 'ec_deployment')).join('\n'), /affected_capabilities: must be an array/);
  });
});

describe('validateCommentBody', () => {
  test('accepts a conformant draft body', () => {
    const { errors, metadata } = validateCommentBody(bodyWith(validMetadata()));
    assert.deepEqual(errors, []);
    assert.equal(metadata.schema_version, '1.1');
  });

  test('reports a missing metadata envelope', () => {
    const { errors } = validateCommentBody('## Implementation research\n');
    assert.match(errors.join('\n'), /envelope not found/);
  });
});

describe('validate-metadata CLI', { concurrency: false }, () => {
  function run(args) {
    try {
      const stdout = execFileSync(process.execPath, [cli, ...args], { encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] });
      return { status: 0, stdout, stderr: '' };
    } catch (err) {
      return { status: err.status, stdout: err.stdout ?? '', stderr: err.stderr ?? '' };
    }
  }

  function draftFile(metadata) {
    const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'validate-metadata-'));
    const file = path.join(dir, 'draft-final.md');
    fs.writeFileSync(file, bodyWith(metadata));
    return file;
  }

  test('exits 0 on a conformant draft', () => {
    const { status, stdout } = run([draftFile(validMetadata())]);
    assert.equal(status, 0);
    assert.match(stdout, /pipeline metadata OK/);
  });

  test('exits 1 and names each violated constraint', () => {
    const metadata = validMetadata();
    metadata.recommendation.approach_index = '0';
    metadata.gate.rounds = 4;
    const { status, stderr } = run([draftFile(metadata)]);

    assert.equal(status, 1);
    assert.match(stderr, /approach_index/);
    assert.match(stderr, /gate\.rounds/);
  });

  test('exits 2 with no argument and on an unreadable path', () => {
    assert.equal(run([]).status, 2);
    assert.equal(run([path.join(os.tmpdir(), 'no-such-draft-4381.md')]).status, 2);
  });
});
