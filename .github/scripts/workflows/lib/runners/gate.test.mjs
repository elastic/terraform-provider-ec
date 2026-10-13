import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const gate = require('./gate.js');

const GATE_ENV = [
  'GATE_NAME',
  'WORKFLOWS_GATE_CLASSIFY_RESULT',
  'WORKFLOWS_GATE_TEST_RESULT',
];

async function withEnv(vars, fn) {
  const previous = {};
  for (const key of GATE_ENV) {
    previous[key] = process.env[key];
    if (vars[key] === undefined) delete process.env[key];
    else process.env[key] = vars[key];
  }
  try {
    return await fn();
  } finally {
    for (const key of GATE_ENV) {
      if (previous[key] === undefined) delete process.env[key];
      else process.env[key] = previous[key];
    }
  }
}

function mockCore() {
  const infos = [];
  const failures = [];
  return {
    infos,
    failures,
    core: {
      info(msg) {
        infos.push(msg);
      },
      setFailed(msg) {
        failures.push(msg);
      },
    },
  };
}

async function run(env) {
  return withEnv(env, async () => {
    const { infos, failures, core } = mockCore();
    await gate({ github: {}, context: {}, core });
    return { infos, failures };
  });
}

describe('runners/gate', { concurrency: false }, () => {
  test('missing GATE_NAME → setFailed', async () => {
    const { infos, failures } = await run({});
    assert.deepEqual(infos, []);
    assert.equal(failures.length, 1);
    assert.match(failures[0], /GATE_NAME environment variable is required/);
  });

  test('unknown GATE_NAME → setFailed listing known gates', async () => {
    const { infos, failures } = await run({ GATE_NAME: 'provider' });
    assert.deepEqual(infos, []);
    assert.equal(failures.length, 1);
    assert.match(failures[0], /Unknown GATE_NAME: provider/);
    assert.match(failures[0], /Expected one of: workflows/);
  });

  test('workflows + classify=true + test=success → info (passed)', async () => {
    const { infos, failures } = await run({
      GATE_NAME: 'workflows',
      WORKFLOWS_GATE_CLASSIFY_RESULT: 'true',
      WORKFLOWS_GATE_TEST_RESULT: 'success',
    });
    assert.deepEqual(failures, []);
    assert.equal(infos.length, 1);
    assert.match(infos[0], /Workflow tests succeeded/);
  });

  test('workflows + classify=false + test=skipped → info (legitimate skip)', async () => {
    const { infos, failures } = await run({
      GATE_NAME: 'workflows',
      WORKFLOWS_GATE_CLASSIFY_RESULT: 'false',
      WORKFLOWS_GATE_TEST_RESULT: 'skipped',
    });
    assert.deepEqual(failures, []);
    assert.equal(infos.length, 1);
    assert.match(infos[0], /legitimately skipped/);
  });

  test('workflows + classify=true + test=failure → setFailed', async () => {
    const { infos, failures } = await run({
      GATE_NAME: 'workflows',
      WORKFLOWS_GATE_CLASSIFY_RESULT: 'true',
      WORKFLOWS_GATE_TEST_RESULT: 'failure',
    });
    assert.deepEqual(infos, []);
    assert.equal(failures.length, 1);
    assert.match(failures[0], /Workflow test failure\. Gate failed\./);
  });

  test('missing WORKFLOWS_GATE_* env vars map to empty strings → setFailed (invalid classify)', async () => {
    const { infos, failures } = await run({ GATE_NAME: 'workflows' });
    assert.deepEqual(infos, []);
    assert.equal(failures.length, 1);
    assert.match(failures[0], /Invalid classify result/);
  });
});
