import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { utf8Len, sliceToByteBudget, clampUtf8Bytes } = require('./utf8.js');

test('utf8Len counts bytes, not code units', () => {
  assert.equal(utf8Len('abc'), 3);
  assert.equal(utf8Len('中'), 3); // CJK: 3 bytes
  assert.equal(utf8Len('\u{1f600}'), 4); // emoji: 4 bytes
});

test('sliceToByteBudget never splits a multi-byte character', () => {
  const cjk = '中'.repeat(10); // 30 bytes
  // Ask for 7 bytes → 2 whole chars (6 bytes), never a partial 3rd.
  const out = sliceToByteBudget(cjk, 7);
  assert.equal(utf8Len(out) <= 7, true);
  assert.doesNotMatch(out, /�/);
  assert.equal(out, '中中');
});

test('sliceToByteBudget handles emoji (4-byte) boundaries', () => {
  const emoji = '\u{1f600}'.repeat(5); // 20 bytes
  const out = sliceToByteBudget(emoji, 10); // 2 whole emoji = 8 bytes
  assert.equal(out, '\u{1f600}\u{1f600}');
  assert.doesNotMatch(out, /�/);
});

test('sliceToByteBudget returns the whole string when it fits', () => {
  assert.equal(sliceToByteBudget('hello', 100), 'hello');
});

test('sliceToByteBudget returns empty for a non-positive budget', () => {
  assert.equal(sliceToByteBudget('hello', 0), '');
});

test('clampUtf8Bytes passes short strings through untouched', () => {
  const r = clampUtf8Bytes('short', 1000);
  assert.deepEqual(r, { text: 'short', truncated: false });
});

test('clampUtf8Bytes clamps a large CJK body within the byte budget and marks it', () => {
  const big = '中'.repeat(50_000); // 150 KB
  const r = clampUtf8Bytes(big, 96_000);
  assert.equal(r.truncated, true);
  assert.ok(utf8Len(r.text) <= 96_000, `budget exceeded: ${utf8Len(r.text)}`);
  assert.ok(utf8Len(r.text) < 128 * 1024);
  assert.match(r.text, /truncated to fit/);
  assert.doesNotMatch(r.text, /�/);
});

test('clampUtf8Bytes tolerates non-string input', () => {
  assert.deepEqual(clampUtf8Bytes(undefined, 10), { text: '', truncated: false });
  assert.deepEqual(clampUtf8Bytes(null, 10), { text: '', truncated: false });
});

test('clampUtf8Bytes never exceeds the budget even when it is smaller than the marker', () => {
  const r = clampUtf8Bytes('abcdefghij', 5);
  assert.equal(r.truncated, true);
  assert.ok(utf8Len(r.text) <= 5, `budget exceeded: ${utf8Len(r.text)}`);
  assert.doesNotMatch(r.text, /�/);
});
