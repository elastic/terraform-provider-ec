import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const require = createRequire(import.meta.url);

const classifyChanges = require(path.resolve(__dirname, 'classify-changes.js'));

function mockCore() {
  const outputs = {};
  const warnings = [];
  const infos = [];
  return {
    outputs,
    warnings,
    infos,
    core: {
      setOutput(key, value) {
        outputs[key] = value;
      },
      warning(msg) {
        warnings.push(msg);
      },
      info(msg) {
        infos.push(msg);
      },
    },
  };
}

function baseContext(eventName, overrides = {}) {
  return {
    eventName,
    repo: { owner: 'elastic', repo: 'terraform-provider-ec' },
    payload: {
      pull_request: { number: 42 },
      ...overrides.payload,
    },
    ...overrides,
  };
}

test('pull_request with a .github/ path → workflow_changes=true', async () => {
  const { outputs, core } = mockCore();
  await classifyChanges({
    github: {
      paginate: async () => [{ filename: '.github/workflows/workflows.yml' }],
      rest: { pulls: { listFiles: {} } },
    },
    context: baseContext('pull_request'),
    core,
  });
  assert.equal(outputs.workflow_changes, 'true');
});

test('pull_request with only non-.github paths → workflow_changes=false', async () => {
  const { outputs, core } = mockCore();
  await classifyChanges({
    github: {
      paginate: async () => [
        { filename: 'ec/internal/provider.go' },
        { filename: 'README.md' },
      ],
      rest: { pulls: { listFiles: {} } },
    },
    context: baseContext('pull_request'),
    core,
  });
  assert.equal(outputs.workflow_changes, 'false');
});

test('pull_request with empty file list → workflow_changes=true (fail-open)', async () => {
  const { outputs, core } = mockCore();
  await classifyChanges({
    github: {
      paginate: async () => [],
      rest: { pulls: { listFiles: {} } },
    },
    context: baseContext('pull_request'),
    core,
  });
  assert.equal(outputs.workflow_changes, 'true');
});

test('pull_request listFiles failure → workflow_changes=true and warning', async () => {
  const { outputs, warnings, core } = mockCore();
  await classifyChanges({
    github: {
      paginate: async () => {
        throw new Error('API down');
      },
      rest: { pulls: { listFiles: {} } },
    },
    context: baseContext('pull_request'),
    core,
  });
  assert.equal(outputs.workflow_changes, 'true');
  assert.equal(warnings.length, 1);
  assert.match(warnings[0], /Failed to list PR files/);
});

test('push → workflow_changes=true unconditionally', async () => {
  const { outputs, infos, core } = mockCore();
  await classifyChanges({
    github: {},
    context: baseContext('push'),
    core,
  });
  assert.equal(outputs.workflow_changes, 'true');
  assert.match(infos.join('\n'), /Push event/);
});

test('workflow_dispatch → workflow_changes=true', async () => {
  const { outputs, infos, core } = mockCore();
  await classifyChanges({
    github: {},
    context: baseContext('workflow_dispatch'),
    core,
  });
  assert.equal(outputs.workflow_changes, 'true');
  assert.match(infos.join('\n'), /workflow_dispatch/);
});

test('pull_request mixed paths including .github/scripts → workflow_changes=true', async () => {
  const { outputs, core } = mockCore();
  await classifyChanges({
    github: {
      paginate: async () => [
        { filename: 'ec/acc/foo_test.go' },
        { filename: '.github/scripts/workflows/lib/gate-workflows.js' },
      ],
      rest: { pulls: { listFiles: {} } },
    },
    context: baseContext('pull_request'),
    core,
  });
  assert.equal(outputs.workflow_changes, 'true');
});
