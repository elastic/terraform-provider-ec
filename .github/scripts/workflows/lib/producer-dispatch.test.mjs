import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
import { writeFileSync, mkdtempSync, rmSync } from 'node:fs';
import { join } from 'node:path';
import { tmpdir } from 'node:os';

const require = createRequire(import.meta.url);
const { parseTemporaryIdMap, dispatchCodeFactory } = require('./producer-dispatch.js');

function withTempFile(name, content) {
  const dir = mkdtempSync(join(tmpdir(), 'producer-dispatch-test-'));
  const path = join(dir, name);
  writeFileSync(path, content);
  return { path, cleanup: () => rmSync(dir, { recursive: true, force: true }) };
}

test('parseTemporaryIdMap returns entries for valid map', () => {
  const { path, cleanup } = withTempFile(
    'map.json',
    JSON.stringify({
      'issue-1': { repo: 'elastic/terraform-provider-ec', number: 42 },
      'issue-2': { repo: 'elastic/terraform-provider-ec', number: 43 },
    })
  );
  const entries = parseTemporaryIdMap(path);
  assert.equal(entries.length, 2);
  assert.deepStrictEqual(entries[0], {
    repo: 'elastic/terraform-provider-ec',
    number: 42,
  });
  assert.deepStrictEqual(entries[1], {
    repo: 'elastic/terraform-provider-ec',
    number: 43,
  });
  cleanup();
});

test('parseTemporaryIdMap returns empty array for empty map', () => {
  const { path, cleanup } = withTempFile('map.json', '{}');
  const entries = parseTemporaryIdMap(path);
  assert.equal(entries.length, 0);
  cleanup();
});

test('parseTemporaryIdMap throws when file is missing', () => {
  assert.throws(() => parseTemporaryIdMap('/nonexistent/path.json'), /not found/);
});

test('parseTemporaryIdMap throws for malformed JSON', () => {
  const { path, cleanup } = withTempFile('map.json', 'not json');
  assert.throws(() => parseTemporaryIdMap(path), /Failed to parse/);
  cleanup();
});

test('parseTemporaryIdMap throws for non-object root (array)', () => {
  const { path, cleanup } = withTempFile('map.json', '[]');
  assert.throws(() => parseTemporaryIdMap(path), /must be a JSON object/);
  cleanup();
});

test('parseTemporaryIdMap throws for non-object root (null)', () => {
  const { path, cleanup } = withTempFile('map.json', 'null');
  assert.throws(() => parseTemporaryIdMap(path), /must be a JSON object/);
  cleanup();
});

test('parseTemporaryIdMap throws for entry missing repo', () => {
  const { path, cleanup } = withTempFile(
    'map.json',
    JSON.stringify({ 'issue-1': { number: 42 } })
  );
  assert.throws(() => parseTemporaryIdMap(path), /invalid repo/);
  cleanup();
});

test('parseTemporaryIdMap throws for entry with invalid repo format', () => {
  const { path, cleanup } = withTempFile(
    'map.json',
    JSON.stringify({ 'issue-1': { repo: 'no-slash', number: 42 } })
  );
  assert.throws(() => parseTemporaryIdMap(path), /invalid repo/);
  cleanup();
});

test('parseTemporaryIdMap throws for entry with non-numeric number', () => {
  const { path, cleanup } = withTempFile(
    'map.json',
    JSON.stringify({
      'issue-1': { repo: 'elastic/terraform-provider-ec', number: 'abc' },
    })
  );
  assert.throws(() => parseTemporaryIdMap(path), /invalid number/);
  cleanup();
});

test('parseTemporaryIdMap throws for entry with zero number', () => {
  const { path, cleanup } = withTempFile(
    'map.json',
    JSON.stringify({
      'issue-1': { repo: 'elastic/terraform-provider-ec', number: 0 },
    })
  );
  assert.throws(() => parseTemporaryIdMap(path), /invalid number/);
  cleanup();
});

test('parseTemporaryIdMap throws for entry with negative number', () => {
  const { path, cleanup } = withTempFile(
    'map.json',
    JSON.stringify({
      'issue-1': { repo: 'elastic/terraform-provider-ec', number: -5 },
    })
  );
  assert.throws(() => parseTemporaryIdMap(path), /invalid number/);
  cleanup();
});

test('parseTemporaryIdMap throws for entry with decimal number', () => {
  const { path, cleanup } = withTempFile(
    'map.json',
    JSON.stringify({
      'issue-1': { repo: 'elastic/terraform-provider-ec', number: 3.14 },
    })
  );
  assert.throws(() => parseTemporaryIdMap(path), /invalid number/);
  cleanup();
});

test('parseTemporaryIdMap throws for entry that is a primitive', () => {
  const { path, cleanup } = withTempFile(
    'map.json',
    JSON.stringify({ 'issue-1': 42 })
  );
  assert.throws(() => parseTemporaryIdMap(path), /Entry "issue-1" must be an object/);
  cleanup();
});

// ---------------------------------------------------------------------------
// dispatchCodeFactory
// ---------------------------------------------------------------------------

function withEnv(overrides, fn) {
  const saved = {};
  for (const key of Object.keys(overrides)) {
    saved[key] = process.env[key];
    if (overrides[key] === undefined) {
      delete process.env[key];
    } else {
      process.env[key] = overrides[key];
    }
  }
  try {
    return fn();
  } finally {
    for (const key of Object.keys(overrides)) {
      if (saved[key] === undefined) {
        delete process.env[key];
      } else {
        process.env[key] = saved[key];
      }
    }
  }
}

test('dispatchCodeFactory throws when GH_TOKEN and GITHUB_TOKEN are absent with zero spawns', () => {
  const calls = [];
  withEnv({ GH_TOKEN: undefined, GITHUB_TOKEN: undefined }, () => {
    assert.throws(
      () =>
        dispatchCodeFactory(
          [{ repo: 'elastic/terraform-provider-ec', number: 1 }],
          'test',
          undefined,
          (...args) => {
            calls.push(args);
            return { status: 0 };
          },
        ),
      /GH_TOKEN or GITHUB_TOKEN/,
    );
  });
  assert.equal(calls.length, 0);
});

test('dispatchCodeFactory throws for cross-repo dispatch with zero spawns', () => {
  const calls = [];
  withEnv(
    { GITHUB_REPOSITORY: 'elastic/allowed', GH_TOKEN: 'test-token' },
    () => {
      assert.throws(
        () =>
          dispatchCodeFactory(
            [{ repo: 'elastic/different', number: 1 }],
            'test',
            undefined,
            (...args) => {
              calls.push(args);
              return { status: 0 };
            },
          ),
        /not the current repository/,
      );
    },
  );
  assert.equal(calls.length, 0);
});

test('dispatchCodeFactory refuses a mixed map before any spawn when one entry is cross-repo', () => {
  const calls = [];
  withEnv(
    {
      GITHUB_REPOSITORY: 'elastic/terraform-provider-ec',
      GH_TOKEN: 'test-token',
    },
    () => {
      assert.throws(
        () =>
          dispatchCodeFactory(
            [
              { repo: 'elastic/terraform-provider-ec', number: 42 },
              { repo: 'elastic/different', number: 43 },
            ],
            'test',
            undefined,
            (...args) => {
              calls.push(args);
              return { status: 0 };
            },
          ),
        /not the current repository/,
      );
    },
  );
  assert.equal(calls.length, 0);
});

test('dispatchCodeFactory fans out one spawn per entry with correct args', () => {
  const calls = [];
  withEnv(
    {
      GITHUB_REPOSITORY: 'elastic/terraform-provider-ec',
      GH_TOKEN: 'test-token',
    },
    () => {
      dispatchCodeFactory(
        [
          { repo: 'elastic/terraform-provider-ec', number: 42 },
          { repo: 'elastic/terraform-provider-ec', number: 43 },
        ],
        'schema-coverage',
        'code-factory-issue.lock.yml',
        (cmd, args, opts) => {
          calls.push({ cmd, args, opts });
          return { status: 0 };
        },
      );
    },
  );

  assert.equal(calls.length, 2);
  for (const call of calls) {
    assert.equal(call.cmd, 'gh');
    assert.equal(call.args[0], 'workflow');
    assert.equal(call.args[1], 'run');
    assert.equal(call.args[2], 'code-factory-issue.lock.yml');
    assert.equal(call.args[3], '--repo');
    assert.equal(call.args[4], 'elastic/terraform-provider-ec');
    assert.ok(call.args.includes('--field'));
    assert.ok(call.args.includes('source_workflow=schema-coverage'));
    assert.equal(call.opts.env.GH_TOKEN, 'test-token');
  }
  assert.ok(calls[0].args.includes('issue_number=42'));
  assert.ok(calls[1].args.includes('issue_number=43'));
});

test('dispatchCodeFactory uses default workflow file when omitted', () => {
  const calls = [];
  withEnv(
    {
      GITHUB_REPOSITORY: 'elastic/terraform-provider-ec',
      GH_TOKEN: 'test-token',
    },
    () => {
      dispatchCodeFactory(
        [{ repo: 'elastic/terraform-provider-ec', number: 7 }],
        'schema-coverage',
        undefined,
        (cmd, args) => {
          calls.push({ cmd, args });
          return { status: 0 };
        },
      );
    },
  );
  assert.equal(calls.length, 1);
  assert.equal(calls[0].args[2], 'code-factory-issue.lock.yml');
});

test('dispatchCodeFactory with empty entries performs zero spawns', () => {
  const calls = [];
  withEnv({ GH_TOKEN: 'test-token' }, () => {
    dispatchCodeFactory([], 'test', undefined, (...args) => {
      calls.push(args);
      return { status: 0 };
    });
  });
  assert.equal(calls.length, 0);
});

test('dispatchCodeFactory throws when spawn reports failure', () => {
  withEnv(
    {
      GITHUB_REPOSITORY: 'elastic/terraform-provider-ec',
      GH_TOKEN: 'test-token',
    },
    () => {
      assert.throws(
        () =>
          dispatchCodeFactory(
            [{ repo: 'elastic/terraform-provider-ec', number: 1 }],
            'test',
            undefined,
            () => ({ status: 1, stderr: Buffer.from('boom') }),
          ),
        /Failed to dispatch for issue #1/,
      );
    },
  );
});
