import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const qualifyTrigger = require('./qualify-trigger.js');

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

async function run(factory, { eventName, payload }) {
  return withEnv({ FACTORY_NAME: factory }, async () => {
    const { outputs, core } = mockCore();
    await qualifyTrigger({
      github: {},
      context: {
        eventName,
        payload,
        repo: { owner: 'elastic', repo: 'terraform-provider-ec' },
      },
      core,
    });
    return outputs;
  });
}

describe('qualify-trigger', { concurrency: false }, () => {
  for (const factory of FACTORIES) {
    const constants = require(`../intake/${factory}-constants.js`);

    test(`accepts issues.labeled when the applied label is ${factory}`, async () => {
      const outputs = await run(factory, {
        eventName: 'issues',
        payload: { action: 'labeled', label: { name: factory }, issue: { labels: [] } },
      });
      assert.equal(outputs.event_eligible, 'true');
      assert.match(outputs.event_eligible_reason, new RegExp(`applied label is ${factory}`));
    });

    test(`rejects issues.opened without the ${factory} label`, async () => {
      const outputs = await run(factory, {
        eventName: 'issues',
        payload: { action: 'opened', issue: { labels: [{ name: 'bug' }] } },
      });
      assert.equal(outputs.event_eligible, 'false');
      assert.equal(outputs.event_eligible_reason, constants.ISSUE_OPENED_NOT_ELIGIBLE_REASON);
      assert.match(outputs.event_eligible_reason, new RegExp(`without the ${factory} label`));
    });
  }

  test('rejects issues.labeled when a different label was applied', async () => {
    const outputs = await run('code-factory', {
      eventName: 'issues',
      payload: { action: 'labeled', label: { name: 'bug' }, issue: { labels: [{ name: 'code-factory' }] } },
    });
    assert.equal(outputs.event_eligible, 'false');
    assert.match(outputs.event_eligible_reason, /not 'code-factory'/);
  });

  test('accepts an issue_comment because the slash-command trigger routes there', async () => {
    const outputs = await run('code-factory', {
      eventName: 'issue_comment',
      payload: { action: 'created', comment: { body: '/code-factory' } },
    });
    assert.equal(outputs.event_eligible, 'true');
    assert.match(outputs.event_eligible_reason, /issue_comment/);
  });

  test('accepts issues.opened when the factory label is already present', async () => {
    const outputs = await run('code-factory', {
      eventName: 'issues',
      payload: { action: 'opened', issue: { labels: [{ name: 'code-factory' }] } },
    });
    assert.equal(outputs.event_eligible, 'true');
    assert.match(outputs.event_eligible_reason, /already has the code-factory label/);
  });

  test('rejects unsupported issue actions such as closed', async () => {
    const outputs = await run('code-factory', {
      eventName: 'issues',
      payload: { action: 'closed', issue: { labels: [{ name: 'code-factory' }] } },
    });
    assert.equal(outputs.event_eligible, 'false');
    assert.match(outputs.event_eligible_reason, /not eligible/);
  });

  test('rejects events other than issues and issue_comment', async () => {
    const outputs = await run('code-factory', {
      eventName: 'pull_request',
      payload: { action: 'opened' },
    });
    assert.equal(outputs.event_eligible, 'false');
    assert.match(outputs.event_eligible_reason, /expected 'issues'/);
  });
});
