import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { positiveIssueNumber } = require('./issue-number.js');

describe('positiveIssueNumber', () => {
  test('accepts a positive integer string', () => {
    assert.equal(positiveIssueNumber('42'), 42);
    assert.equal(positiveIssueNumber('1'), 1);
  });

  test('rejects a decimal that parseInt would truncate to a real issue', () => {
    assert.equal(positiveIssueNumber('1.5'), 0);
    assert.equal(positiveIssueNumber('12.0'), 0);
  });

  test('rejects trailing or leading junk', () => {
    assert.equal(positiveIssueNumber('12abc'), 0);
    assert.equal(positiveIssueNumber(' 12'), 0);
    assert.equal(positiveIssueNumber('12 '), 0);
    assert.equal(positiveIssueNumber('+12'), 0);
  });

  test('rejects zero, negatives and leading zeros', () => {
    assert.equal(positiveIssueNumber('0'), 0);
    assert.equal(positiveIssueNumber('-3'), 0);
    assert.equal(positiveIssueNumber('007'), 0);
  });

  test('rejects empty, missing and non-string inputs', () => {
    assert.equal(positiveIssueNumber(''), 0);
    assert.equal(positiveIssueNumber(undefined), 0);
    assert.equal(positiveIssueNumber(null), 0);
    assert.equal(positiveIssueNumber(42), 0);
  });

  test('rejects a value beyond the safe integer range', () => {
    assert.equal(positiveIssueNumber('9007199254740993'), 0);
  });
});
