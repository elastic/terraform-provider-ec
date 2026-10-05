import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const removeTriggerLabel = require('./remove-trigger-label.js');

const FACTORIES = ['change-factory', 'code-factory', 'research-factory', 'reproducer-factory'];

async function withEnv(vars, fn) {
  const previous = {};
  for (const key of Object.keys(vars)) {
    previous[key] = process.env[key];
    if (vars[key] === undefined) delete process.env[key];
    else process.env[key] = vars[key];
  }
  try {
    return await fn();
  } finally {
    for (const key of Object.keys(previous)) {
      if (previous[key] === undefined) delete process.env[key];
      else process.env[key] = previous[key];
    }
  }
}

function mockCore() {
  const outputs = {};
  return {
    outputs,
    core: {
      setOutput(key, value) {
        outputs[key] = value;
      },
      info() {},
    },
  };
}

async function run({ factory = 'code-factory', issueNumber, removeLabel }) {
  return withEnv({ FACTORY_NAME: factory }, async () => {
    const { outputs, core } = mockCore();
    const calls = [];
    await removeTriggerLabel({
      github: {
        rest: {
          issues: {
            removeLabel: async (args) => {
              calls.push(args);
              return removeLabel(args);
            },
          },
        },
      },
      context: {
        payload: issueNumber == null ? {} : { issue: { number: issueNumber } },
        repo: { owner: 'elastic', repo: 'terraform-provider-ec' },
      },
      core,
    });
    return { outputs, calls };
  });
}

describe('remove-trigger-label', { concurrency: false }, () => {
  for (const factory of FACTORIES) {
    test(`removes the ${factory} label`, async () => {
      const { outputs, calls } = await run({
        factory,
        issueNumber: 42,
        removeLabel: async () => ({}),
      });
      assert.equal(outputs.trigger_label_removed, 'true');
      assert.deepEqual(calls, [{
        owner: 'elastic',
        repo: 'terraform-provider-ec',
        issue_number: 42,
        name: factory,
      }]);
    });
  }

  test('treats a 404 as success when the label is already absent', async () => {
    const err = new Error('Not Found');
    err.status = 404;
    const { outputs } = await run({
      issueNumber: 42,
      removeLabel: async () => {
        throw err;
      },
    });
    assert.equal(outputs.trigger_label_removed, 'true');
    assert.match(outputs.trigger_label_removed_reason, /code-factory/);
  });

  test('reports failure on a non-404 API error', async () => {
    const err = new Error('Internal Server Error');
    err.status = 500;
    const { outputs } = await run({
      issueNumber: 42,
      removeLabel: async () => {
        throw err;
      },
    });
    assert.equal(outputs.trigger_label_removed, 'false');
    assert.match(outputs.trigger_label_removed_reason, /Internal Server Error/);
  });

  test('skips removal when the event has no issue number', async () => {
    const { outputs, calls } = await run({
      issueNumber: null,
      removeLabel: async () => ({}),
    });
    assert.equal(calls.length, 0);
    assert.equal(outputs.trigger_label_removed, 'false');
    assert.match(outputs.trigger_label_removed_reason, /No issue number/);
  });
});
