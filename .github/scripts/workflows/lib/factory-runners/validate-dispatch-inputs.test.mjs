import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const validateDispatchInputs = require('./validate-dispatch-inputs.js');

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

async function run(issueNumber) {
  const { outputs, core } = mockCore();
  await validateDispatchInputs({
    github: {},
    context: {
      repo: { owner: 'elastic', repo: 'terraform-provider-ec' },
      payload: { inputs: { issue_number: issueNumber } },
    },
    core,
  });
  return outputs;
}

describe('validate-dispatch-inputs', () => {
  test('accepts a positive integer issue number', async () => {
    const outputs = await run('42');
    assert.equal(outputs.event_eligible, 'true');
    assert.equal(outputs.issue_number, '42');
    assert.match(outputs.event_eligible_reason, /issue #42/);
    assert.match(outputs.event_eligible_reason, /elastic\/terraform-provider-ec/);
  });

  test('rejects an issue number above the safe integer range', async () => {
    const outputs = await run('9007199254740992');
    assert.equal(outputs.event_eligible, 'false');
    assert.equal(outputs.issue_number, undefined);
    assert.match(outputs.event_eligible_reason, /not a valid positive integer/);
  });

  for (const issueNumber of ['', '0', '-1', 'abc', '3.14', '007']) {
    test(`rejects issue_number ${JSON.stringify(issueNumber)}`, async () => {
      const outputs = await run(issueNumber);
      assert.equal(outputs.event_eligible, 'false');
      assert.equal(outputs.issue_number, undefined);
      assert.match(outputs.event_eligible_reason, /not a valid positive integer/);
    });
  }
});
