import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { fenceOf, locateMetadata, extractMetadata } = require('./locate-metadata.js');

const lines = (...parts) => parts.join('\n');

function envelope(json, summary = '🤖 Pipeline metadata') {
  return lines(
    `<details><summary>${summary}</summary>`,
    '',
    '```json',
    json,
    '```',
    '',
    '</details>',
    '',
  );
}

describe('fenceOf', () => {
  test('recognises backtick and tilde fences with an info string', () => {
    assert.deepEqual(fenceOf('```json'), { char: '`', length: 3, info: 'json' });
    assert.deepEqual(fenceOf('~~~~json\n'), { char: '~', length: 4, info: 'json' });
    assert.deepEqual(fenceOf('   ```'), { char: '`', length: 3, info: '' });
  });

  test('rejects non-fence lines and inline backtick spans', () => {
    assert.equal(fenceOf('plain text'), null);
    assert.equal(fenceOf('`` inline ``'), null);
    assert.equal(fenceOf('```js `x` ```'), null);
    assert.equal(fenceOf('    ```json'), null, 'four-space indent is an indented code block');
  });
});

describe('locateMetadata', () => {
  test('returns the content range of the metadata fence', () => {
    const body = lines('### References', '- x', '', envelope('{"schema_version": "1.1"}'));
    const located = locateMetadata(body);

    assert.equal(located.error, undefined);
    assert.equal(located.text, '{"schema_version": "1.1"}');
    assert.equal(body.slice(located.start, located.end), '{"schema_version": "1.1"}');
  });

  test('ignores a decoy collapsed details block before the metadata envelope', () => {
    const body = lines(
      '#### Approach A',
      '<details><summary>Pros and cons</summary>',
      '',
      '```json',
      '{"not": "metadata"}',
      '```',
      '',
      '</details>',
      '',
      envelope('{"schema_version": "1.1", "real": true}'),
    );
    const { metadata, error } = extractMetadata(body);

    assert.equal(error, undefined);
    assert.deepEqual(metadata, { schema_version: '1.1', real: true });
  });

  test('accepts a second json fence elsewhere in the body', () => {
    const body = lines(
      '### Approaches considered',
      '```json',
      '{"example": "API response"}',
      '```',
      '',
      envelope('{"schema_version": "1.1"}'),
    );
    const { metadata, error } = extractMetadata(body);

    assert.equal(error, undefined);
    assert.deepEqual(metadata, { schema_version: '1.1' });
  });

  test('ignores details and summary tags that appear inside a code fence', () => {
    const body = lines(
      '### Approaches considered',
      '````markdown',
      '<details><summary>🤖 Pipeline metadata</summary>',
      '',
      '```json',
      '{"quoted": "example"}',
      '```',
      '',
      '</details>',
      '````',
      '',
      envelope('{"schema_version": "1.1"}'),
    );
    const { metadata, error } = extractMetadata(body);

    assert.equal(error, undefined);
    assert.deepEqual(metadata, { schema_version: '1.1' });
  });

  test('an info-bearing fence line does not close an open fence', () => {
    // CommonMark: a closing fence carries no info string. Without that rule the inner ```json line
    // of a quoted example closes the outer fence, scanning goes live again inside the example, and
    // its `<details>`/`</details>` tags are read as real envelopes.
    // Same fence length throughout, so the length rule cannot mask the info-string rule: only
    // `parsed.info === ''` stops the ```json line from closing the ```markdown example.
    const body = lines(
      '#### Approach A',
      '```markdown',
      '```json',
      '{"example": true}',
      '<details><summary>🤖 Pipeline metadata</summary>',
      '</details>',
      '```',
      '',
      envelope('{"schema_version": "1.1", "real": true}'),
    );
    const { metadata, error } = extractMetadata(body);

    assert.equal(error, undefined);
    assert.deepEqual(metadata, { schema_version: '1.1', real: true });
  });

  test('a longer closing fence closes a shorter opener, a shorter one does not', () => {
    const closedByLonger = lines('```json', '{"a": 1}', '`````', '', envelope('{"schema_version": "1.1"}'));
    assert.equal(extractMetadata(closedByLonger).error, undefined);

    const notClosed = lines('````json', '{"a": 1}', '```', '', envelope('{"schema_version": "1.1"}'));
    assert.match(locateMetadata(notClosed).error, /envelope not found/);
  });

  test('reports a missing envelope', () => {
    const { error } = locateMetadata('### References\n- none\n');
    assert.match(error, /envelope not found/);
  });

  test('reports two metadata envelopes', () => {
    const body = lines(envelope('{"a": 1}'), envelope('{"b": 2}'));
    const { error } = locateMetadata(body);
    assert.match(error, /multiple Pipeline metadata/);
  });

  test('reports an envelope with no json fence', () => {
    const body = lines('<details><summary>🤖 Pipeline metadata</summary>', '', 'no fence here', '', '</details>');
    const { error } = locateMetadata(body);
    assert.match(error, /exactly one json block, found 0/);
  });

  test('reports an envelope with two json fences', () => {
    const body = lines(
      '<details><summary>🤖 Pipeline metadata</summary>',
      '',
      '```json',
      '{"a": 1}',
      '```',
      '',
      '```json',
      '{"b": 2}',
      '```',
      '',
      '</details>',
    );
    const { error } = locateMetadata(body);
    assert.match(error, /exactly one json block, found 2/);
  });

  test('anchors on the summary text, not on the first details element', () => {
    const body = lines(envelope('{"schema_version": "1.1"}', 'Raw API payload'));
    const { error } = locateMetadata(body);
    assert.match(error, /envelope not found/);
  });

  test('matches the summary case-insensitively and with surrounding text', () => {
    const body = envelope('{"schema_version": "1.1"}', '🤖 PIPELINE METADATA (machine-readable)');
    const { metadata, error } = extractMetadata(body);
    assert.equal(error, undefined);
    assert.deepEqual(metadata, { schema_version: '1.1' });
  });

  test('handles CRLF line endings', () => {
    const body = lines('### References', '', envelope('{"schema_version": "1.1"}')).replace(/\n/g, '\r\n');
    const { metadata, error } = extractMetadata(body);

    assert.equal(error, undefined);
    assert.deepEqual(metadata, { schema_version: '1.1' });
  });

  test('tolerates a non-string body', () => {
    assert.match(locateMetadata(undefined).error, /envelope not found/);
    assert.match(locateMetadata(null).error, /envelope not found/);
  });
});

describe('extractMetadata', () => {
  test('reports unparseable JSON without throwing', () => {
    const { error, metadata } = extractMetadata(envelope('{"schema_version": "1.1",}'));
    assert.equal(metadata, undefined);
    assert.match(error, /could not be parsed/);
  });

  test('propagates the locate error', () => {
    const { error } = extractMetadata('nothing here');
    assert.match(error, /envelope not found/);
  });
});
