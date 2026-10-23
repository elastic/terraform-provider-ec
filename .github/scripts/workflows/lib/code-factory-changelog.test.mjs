import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
import { readFileSync, writeFileSync, mkdtempSync, rmSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { tmpdir } from 'node:os';
import { fileURLToPath } from 'node:url';
import fs from 'node:fs';

const require = createRequire(import.meta.url);
const {
  ALLOWED_TYPES,
  selectChangelogItem,
  validateChangelogItem,
  renderFragment,
  main,
} = require('./code-factory-changelog.js');

const here = dirname(fileURLToPath(import.meta.url));

// ---------------------------------------------------------------------------
// Drift guard: the JS ALLOWED_TYPES must match Go's allowedTypes verbatim.
// ---------------------------------------------------------------------------

test('ALLOWED_TYPES matches scripts/changelogcheck/validate.go allowedTypes', () => {
  const goPath = join(here, '../../../../scripts/changelogcheck/validate.go');
  const src = readFileSync(goPath, 'utf8');
  // Extract the block:  var allowedTypes = map[string]struct{}{\n ... \n}
  // The per-entry struct{} values also contain braces, so anchor the close on a
  // line-start brace (`\n}`) rather than the first `}`.
  const block = src.match(/allowedTypes\s*=\s*map\[string\]struct\{\}\{\n([\s\S]*?)\n\}/);
  assert.ok(block, 'could not locate allowedTypes map in validate.go');
  const goTypes = [...block[1].matchAll(/"([^"]+)"\s*:/g)].map((m) => m[1]).sort();
  assert.deepEqual([...ALLOWED_TYPES].sort(), goTypes,
    'JS ALLOWED_TYPES drifted from Go allowedTypes');
});

// ---------------------------------------------------------------------------
// selectChangelogItem
// ---------------------------------------------------------------------------

test('selectChangelogItem returns the single add_changelog item', () => {
  const item = selectChangelogItem({
    items: [
      { type: 'noise' },
      { type: 'add_changelog', release_note_type: 'none' },
    ],
  });
  assert.equal(item.release_note_type, 'none');
});

test('selectChangelogItem throws on zero or multiple items', () => {
  assert.throws(() => selectChangelogItem({ items: [] }), /exactly one/);
  assert.throws(() => selectChangelogItem({
    items: [{ type: 'add_changelog' }, { type: 'add_changelog' }],
  }), /found 2/);
});

test('selectChangelogItem throws when items is not an array', () => {
  assert.throws(() => selectChangelogItem({}), /missing an "items" array/);
  assert.throws(() => selectChangelogItem(null), /missing an "items" array/);
});

// ---------------------------------------------------------------------------
// validateChangelogItem
// ---------------------------------------------------------------------------

test('validateChangelogItem accepts a valid bug note', () => {
  const v = validateChangelogItem({
    release_note_type: 'bug', release_note_body: 'resource/ec_deployment: fix',
  });
  assert.deepEqual(v, { type: 'bug', body: 'resource/ec_deployment: fix' });
});

test('validateChangelogItem rejects unknown types', () => {
  assert.throws(() => validateChangelogItem({
    release_note_type: 'chore', release_note_body: 'x',
  }), /is not one of/);
});

test('validateChangelogItem enforces none => empty body', () => {
  assert.throws(() => validateChangelogItem({
    release_note_type: 'none', release_note_body: 'should be empty',
  }), /none must have an empty body/);
});

test('validateChangelogItem enforces non-none => non-empty body', () => {
  assert.throws(() => validateChangelogItem({
    release_note_type: 'enhancement', release_note_body: '   ',
  }), /must have a non-empty body/);
});

test('validateChangelogItem rejects a fence-only body that would collapse the block', () => {
  assert.throws(() => validateChangelogItem({
    release_note_type: 'bug', release_note_body: '```',
  }), /must not contain a code fence/);
});

test('validateChangelogItem rejects a multiline body', () => {
  assert.throws(() => validateChangelogItem({
    release_note_type: 'bug', release_note_body: 'line one\nline two',
  }), /must be a single line/);
});

test('validateChangelogItem rejects a body embedding an extra release-note fence', () => {
  assert.throws(() => validateChangelogItem({
    release_note_type: 'bug', release_note_body: 'ok\n```releasenote\nsneaky\n```',
  }), /single line|code fence/);
});

// ---------------------------------------------------------------------------
// renderFragment
// ---------------------------------------------------------------------------

test('renderFragment renders a typed note with a trailing newline', () => {
  assert.equal(
    renderFragment({ type: 'bug', body: 'resource/ec_deployment: fix' }),
    '```release-note:bug\nresource/ec_deployment: fix\n```\n',
  );
});

test('renderFragment renders none as an empty-bodied block', () => {
  assert.equal(renderFragment({ type: 'none', body: '' }), '```release-note:none\n```\n');
});

// ---------------------------------------------------------------------------
// main(--render-only) end to end
// ---------------------------------------------------------------------------

function withAgentOutput(obj, fn) {
  const dir = mkdtempSync(join(tmpdir(), 'cf-changelog-'));
  const path = join(dir, 'agent_output.json');
  writeFileSync(path, JSON.stringify(obj));
  try {
    return fn(path);
  } finally {
    rmSync(dir, { recursive: true, force: true });
  }
}

function capture() {
  let out = '', err = '';
  return { stdout: { write: (s) => { out += s; } }, stderr: { write: (s) => { err += s; } }, get out() { return out; }, get err() { return err; } };
}

test('main --render-only prints the fragment and exits 0', () => {
  withAgentOutput({ items: [{ type: 'add_changelog', release_note_type: 'bug', release_note_body: 'fix a thing' }] }, (path) => {
    const cap = capture();
    const code = main(['--render-only'], { GH_AW_AGENT_OUTPUT: path }, fs, cap.stdout, cap.stderr);
    assert.equal(code, 0);
    assert.equal(cap.out, '```release-note:bug\nfix a thing\n```\n');
  });
});

test('main --render-only exits 1 on a malformed item', () => {
  withAgentOutput({ items: [{ type: 'add_changelog', release_note_type: 'none', release_note_body: 'not empty' }] }, (path) => {
    const cap = capture();
    const code = main(['--render-only'], { GH_AW_AGENT_OUTPUT: path }, fs, cap.stdout, cap.stderr);
    assert.equal(code, 1);
    assert.match(cap.err, /none must have an empty body/);
  });
});

test('main without --render-only exits 2', () => {
  const cap = capture();
  const code = main([], {}, fs, cap.stdout, cap.stderr);
  assert.equal(code, 2);
});
