import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const checkDuplicatePR = require('./check-duplicate-pr.js');

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

function pull({ number, ref, label, body }) {
  return {
    number,
    state: 'open',
    head: { ref },
    labels: [{ name: label }],
    body,
    html_url: `https://example.com/pr/${number}`,
  };
}

async function run({ factory, eventName, payload, pulls }) {
  return withEnv({ FACTORY_NAME: factory }, async () => {
    const { outputs, core } = mockCore();
    const calls = [];
    const list = { marker: 'pulls.list' };
    const github = {
      rest: { pulls: { list } },
      paginate: async (fn, params) => {
        calls.push({ fn, params });
        return pulls;
      },
    };
    await checkDuplicatePR({
      github,
      context: {
        eventName,
        payload,
        repo: { owner: 'elastic', repo: 'terraform-provider-ec' },
      },
      core,
    });
    return { outputs, calls, list };
  });
}

describe('check-duplicate-pr', { concurrency: false }, () => {
  test('suppresses a re-run when an open code-factory PR uses Closes #N', async () => {
    const { outputs, calls, list } = await run({
      factory: 'code-factory',
      eventName: 'issues',
      payload: { issue: { number: 42 } },
      pulls: [pull({
        number: 101,
        ref: 'code-factory/issue-42',
        label: 'code-factory',
        body: 'Closes #42',
      })],
    });

    assert.equal(calls.length, 1);
    assert.equal(calls[0].fn, list);
    assert.deepEqual(calls[0].params, {
      owner: 'elastic',
      repo: 'terraform-provider-ec',
      state: 'open',
      head: 'elastic:code-factory/issue-42',
      per_page: 100,
    });
    assert.equal(outputs.duplicate_pr_found, 'true');
    assert.equal(outputs.duplicate_pr_url, 'https://example.com/pr/101');
    assert.match(outputs.gate_reason, /Closes #42/);
  });

  test('does not treat Related to #N as a code-factory duplicate', async () => {
    const { outputs } = await run({
      factory: 'code-factory',
      eventName: 'issues',
      payload: { issue: { number: 42 } },
      pulls: [pull({
        number: 102,
        ref: 'code-factory/issue-42',
        label: 'code-factory',
        body: 'Related to #42',
      })],
    });
    assert.equal(outputs.duplicate_pr_found, 'false');
  });

  test('suppresses a re-run when an open change-factory PR uses Related to #N', async () => {
    const { outputs, calls } = await run({
      factory: 'change-factory',
      eventName: 'issues',
      payload: { issue: { number: 7 } },
      pulls: [pull({
        number: 201,
        ref: 'change-factory/issue-7',
        label: 'change-factory',
        body: 'Related to #7',
      })],
    });

    assert.equal(calls[0].params.head, 'elastic:change-factory/issue-7');
    assert.equal(outputs.duplicate_pr_found, 'true');
    assert.match(outputs.gate_reason, /Related to #7/);
  });

  test('does not treat Closes #N as a change-factory duplicate', async () => {
    const { outputs } = await run({
      factory: 'change-factory',
      eventName: 'issues',
      payload: { issue: { number: 7 } },
      pulls: [pull({
        number: 202,
        ref: 'change-factory/issue-7',
        label: 'change-factory',
        body: 'Closes #7',
      })],
    });
    assert.equal(outputs.duplicate_pr_found, 'false');
  });

  test('reads the issue number from workflow_dispatch inputs', async () => {
    const { outputs, calls } = await run({
      factory: 'code-factory',
      eventName: 'workflow_dispatch',
      payload: { inputs: { issue_number: '42' } },
      pulls: [pull({
        number: 301,
        ref: 'code-factory/issue-42',
        label: 'code-factory',
        body: 'Closes #42',
      })],
    });
    assert.equal(calls[0].params.head, 'elastic:code-factory/issue-42');
    assert.equal(outputs.duplicate_pr_found, 'true');
  });

  test('ignores an open PR on the factory branch that lacks the factory label', async () => {
    const { outputs } = await run({
      factory: 'code-factory',
      eventName: 'issues',
      payload: { issue: { number: 42 } },
      pulls: [{
        number: 103,
        state: 'open',
        head: { ref: 'code-factory/issue-42' },
        labels: [{ name: 'bug' }],
        body: 'Closes #42',
        html_url: 'https://example.com/pr/103',
      }],
    });
    assert.equal(outputs.duplicate_pr_found, 'false');
  });

  test('skips the API when no issue number is available', async () => {
    const { outputs, calls } = await run({
      factory: 'code-factory',
      eventName: 'issues',
      payload: {},
      pulls: [],
    });
    assert.equal(calls.length, 0);
    assert.equal(outputs.duplicate_pr_found, 'false');
    assert.equal(outputs.duplicate_pr_url, '');
    assert.match(outputs.gate_reason, /No issue number/);
  });
});
