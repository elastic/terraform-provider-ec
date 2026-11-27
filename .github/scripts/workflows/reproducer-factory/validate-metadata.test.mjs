import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { validateMetadata, validateCommentBody, SCHEMA_VERSION, OUTCOMES } = require('./validate-metadata.js');

/** A metadata object that conforms, with any overrides applied. */
function meta(overrides = {}) {
  return { schema_version: SCHEMA_VERSION, outcome: 'reproduced-by-inspection', ...overrides };
}

/** Wraps a JSON string in the standard Pipeline metadata envelope. */
function withEnvelope(json) {
  return `## Reproduction\n\n<details><summary>🤖 Pipeline metadata</summary>\n\n\`\`\`json\n${json}\n\`\`\`\n\n</details>\n`;
}

// ── the three outcomes ────────────────────────────────────────

for (const outcome of OUTCOMES) {
  test(`validateMetadata accepts outcome ${outcome}`, () => {
    assert.deepEqual(validateMetadata(meta({ outcome })).errors, []);
  });
}

test('OUTCOMES is exactly the three #4384 values', () => {
  assert.deepEqual([...OUTCOMES].sort(), ['appears-fixed', 'cannot-reproduce', 'reproduced-by-inspection']);
});

test('validateMetadata rejects the stack-provider spelling "reproduced"', () => {
  const { errors } = validateMetadata(meta({ outcome: 'reproduced' }));
  assert.equal(errors.length, 1);
  assert.match(errors[0], /outcome: must be one of/);
});

test('validateMetadata rejects a missing outcome', () => {
  const { errors } = validateMetadata({ schema_version: SCHEMA_VERSION });
  assert.match(errors.join(), /outcome: must be one of/);
});

// ── schema_version ────────────────────────────────────────────

test('validateMetadata rejects a wrong schema_version', () => {
  const { errors } = validateMetadata(meta({ schema_version: '1.1' }));
  assert.match(errors.join(), /schema_version: must be "1.0"/);
});

test('validateMetadata rejects a missing schema_version', () => {
  const { errors } = validateMetadata({ outcome: 'appears-fixed' });
  assert.match(errors.join(), /schema_version/);
});

// ── non-object ────────────────────────────────────────────────

test('validateMetadata rejects a non-object', () => {
  assert.match(validateMetadata(null).errors.join(), /must be a JSON object/);
  assert.match(validateMetadata([1, 2]).errors.join(), /must be a JSON object/);
  assert.match(validateMetadata('x').errors.join(), /must be a JSON object/);
});

// ── test_name (optional) ──────────────────────────────────────

test('validateMetadata accepts an absent test_name', () => {
  assert.deepEqual(validateMetadata(meta()).errors, []);
});

test('validateMetadata accepts a non-empty test_name', () => {
  assert.deepEqual(validateMetadata(meta({ test_name: 'TestIssue1234Expander' })).errors, []);
});

test('validateMetadata rejects an empty test_name', () => {
  assert.match(validateMetadata(meta({ test_name: '  ' })).errors.join(), /test_name: when present/);
});

test('validateMetadata rejects a non-string test_name', () => {
  assert.match(validateMetadata(meta({ test_name: 42 })).errors.join(), /test_name: when present/);
});

// ── pr_url (optional, S6/H) ───────────────────────────────────

test('validateMetadata accepts an absent pr_url', () => {
  assert.deepEqual(validateMetadata(meta()).errors, []);
});

test('validateMetadata accepts an https pr_url', () => {
  assert.deepEqual(
    validateMetadata(meta({ pr_url: 'https://github.com/elastic/terraform-provider-ec/pull/42' })).errors,
    [],
  );
});

test('validateMetadata rejects a non-https pr_url', () => {
  assert.match(validateMetadata(meta({ pr_url: 'http://example.com' })).errors.join(), /pr_url: when present/);
});

test('validateMetadata rejects an empty pr_url', () => {
  assert.match(validateMetadata(meta({ pr_url: '' })).errors.join(), /pr_url: when present/);
});

// ── validateCommentBody (integration with the envelope) ───────

test('validateCommentBody accepts a well-formed comment body', () => {
  const body = withEnvelope(JSON.stringify(meta({ test_name: 'TestFoo' })));
  const { errors, metadata } = validateCommentBody(body);
  assert.deepEqual(errors, []);
  assert.equal(metadata.outcome, 'reproduced-by-inspection');
});

test('validateCommentBody surfaces a missing envelope as an error', () => {
  const { errors } = validateCommentBody('no metadata at all');
  assert.match(errors.join(), /not found/);
});

test('validateCommentBody surfaces an outcome enum violation through the envelope', () => {
  const body = withEnvelope(JSON.stringify(meta({ outcome: 'reproduced' })));
  assert.match(validateCommentBody(body).errors.join(), /outcome: must be one of/);
});
