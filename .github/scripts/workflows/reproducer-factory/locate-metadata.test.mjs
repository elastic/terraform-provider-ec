import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { fenceOf, locateMetadata, linesOutsideFences, extractMetadata } = require('./locate-metadata.js');

/** Builds a comment body wrapping `json` in the standard Pipeline metadata envelope. */
function withEnvelope(json, { before = '', after = '' } = {}) {
  return `${before}<details><summary>🤖 Pipeline metadata</summary>\n\n\`\`\`json\n${json}\n\`\`\`\n\n</details>\n${after}`;
}

// ── fenceOf ───────────────────────────────────────────────────

test('fenceOf parses a json fence', () => {
  assert.deepEqual(fenceOf('```json'), { char: '`', length: 3, info: 'json' });
});

test('fenceOf parses a bare closing fence', () => {
  assert.deepEqual(fenceOf('```'), { char: '`', length: 3, info: '' });
});

test('fenceOf returns null for a non-fence line', () => {
  assert.equal(fenceOf('just text'), null);
});

// ── locateMetadata ────────────────────────────────────────────

test('locateMetadata finds the single json fence in the envelope', () => {
  const body = withEnvelope('{"outcome":"appears-fixed"}');
  const located = locateMetadata(body);
  assert.ok(!located.error, located.error);
  assert.equal(located.text.trim(), '{"outcome":"appears-fixed"}');
});

test('locateMetadata errors when there is no envelope', () => {
  const located = locateMetadata('## Summary\nno metadata here');
  assert.match(located.error, /not found/);
});

test('locateMetadata errors on multiple envelopes', () => {
  const body = `${withEnvelope('{"a":1}')}\n${withEnvelope('{"b":2}')}`;
  const located = locateMetadata(body);
  assert.match(located.error, /multiple/);
});

test('locateMetadata ignores a json fence outside the envelope', () => {
  const body = withEnvelope('{"outcome":"cannot-reproduce"}', {
    before: 'A Cloud API response:\n\n```json\n{"not":"metadata"}\n```\n\n',
  });
  const located = locateMetadata(body);
  assert.ok(!located.error, located.error);
  assert.equal(located.text.trim(), '{"outcome":"cannot-reproduce"}');
});

test('locateMetadata ignores <details>/<summary> quoted inside a code fence', () => {
  const body = withEnvelope('{"outcome":"reproduced-by-inspection"}', {
    before: '```markdown\n<details><summary>🤖 Pipeline metadata</summary>\nfake\n</details>\n```\n\n',
  });
  const located = locateMetadata(body);
  assert.ok(!located.error, located.error);
  assert.equal(located.text.trim(), '{"outcome":"reproduced-by-inspection"}');
});

// ── extractMetadata ───────────────────────────────────────────

test('extractMetadata parses the metadata object', () => {
  const extracted = extractMetadata(withEnvelope('{"schema_version":"1.0","outcome":"appears-fixed"}'));
  assert.ok(!extracted.error, extracted.error);
  assert.equal(extracted.metadata.outcome, 'appears-fixed');
});

test('extractMetadata reports a parse error for invalid JSON', () => {
  const extracted = extractMetadata(withEnvelope('{not json}'));
  assert.match(extracted.error, /could not be parsed/);
});

// ── linesOutsideFences ────────────────────────────────────────

test('linesOutsideFences skips lines inside a fence', () => {
  const body = 'outside1\n```\ninside\n```\noutside2';
  const outside = [...linesOutsideFences(body)].map((l) => l.line.replace(/\n$/, ''));
  assert.deepEqual(outside, ['outside1', 'outside2']);
});
