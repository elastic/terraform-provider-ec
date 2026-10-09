import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { TRIGGER_LABEL, verifyTriggerLabel } = require('./verify-label.js');

test('TRIGGER_LABEL is verify-openspec', () => {
  assert.equal(TRIGGER_LABEL, 'verify-openspec');
});

test('verifyTriggerLabel accepts the trigger label', () => {
  const result = verifyTriggerLabel('verify-openspec');
  assert.equal(result.label_verified, true);
  assert.ok(result.label_verified_reason.includes('verify-openspec'));
});

test('verifyTriggerLabel rejects a different label', () => {
  const result = verifyTriggerLabel('code-factory');
  assert.equal(result.label_verified, false);
  assert.ok(result.label_verified_reason.includes('verify-openspec'));
  assert.ok(result.label_verified_reason.includes('code-factory'));
});

test('verifyTriggerLabel rejects empty label', () => {
  const result = verifyTriggerLabel('');
  assert.equal(result.label_verified, false);
  assert.ok(result.label_verified_reason.includes('(empty)'));
});

test('verifyTriggerLabel rejects undefined label', () => {
  const result = verifyTriggerLabel(undefined);
  assert.equal(result.label_verified, false);
  assert.ok(result.label_verified_reason.includes('(empty)'));
});
