import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const finalizeGate = require('./finalize-gate.js');

const GATE_ENV = [
  'FACTORY_NAME',
  'EVENT_ELIGIBLE',
  'EVENT_ELIGIBLE_REASON',
  'ACTOR_TRUSTED',
  'ACTOR_TRUSTED_REASON',
  'DUPLICATE_PR_FOUND',
  'DUPLICATE_PR_URL',
  'DUPLICATE_GATE_REASON',
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

async function run(env) {
  return withEnv(env, async () => {
    const { outputs, core } = mockCore();
    await finalizeGate({ github: {}, context: {}, core });
    return outputs;
  });
}

const eligibleTrusted = {
  EVENT_ELIGIBLE: 'true',
  EVENT_ELIGIBLE_REASON: 'Event is eligible.',
  ACTOR_TRUSTED: 'true',
  ACTOR_TRUSTED_REASON: 'Trigger actor is trusted.',
};

describe('finalize-gate', { concurrency: false }, () => {
  test('emits the eligibility failure and ignores later gates', async () => {
    const outputs = await run({
      FACTORY_NAME: 'code-factory',
      EVENT_ELIGIBLE: 'false',
      EVENT_ELIGIBLE_REASON: 'Issue labeled event does not qualify.',
      ACTOR_TRUSTED: 'false',
      DUPLICATE_PR_FOUND: 'true',
    });
    assert.deepEqual(Object.keys(outputs), ['gate_reason']);
    assert.equal(outputs.gate_reason, 'Issue labeled event does not qualify.');
  });

  test('emits the actor-trust failure when the event is eligible', async () => {
    const outputs = await run({
      FACTORY_NAME: 'code-factory',
      EVENT_ELIGIBLE: 'true',
      EVENT_ELIGIBLE_REASON: 'Event is eligible.',
      ACTOR_TRUSTED: 'false',
      ACTOR_TRUSTED_REASON: 'Trigger actor is not a member.',
      DUPLICATE_PR_FOUND: 'true',
      DUPLICATE_GATE_REASON: 'Found existing linked code-factory PR #101.',
    });
    assert.equal(outputs.gate_reason, 'Trigger actor is not a member.');
  });

  for (const factory of ['change-factory', 'code-factory', 'reproducer-factory']) {
    test(`emits the duplicate-PR reason for ${factory}`, async () => {
      const reason = `Found existing linked ${factory} PR #101.`;
      const outputs = await run({
        FACTORY_NAME: factory,
        ...eligibleTrusted,
        DUPLICATE_PR_FOUND: 'true',
        DUPLICATE_PR_URL: 'https://example.com/pr/101',
        DUPLICATE_GATE_REASON: reason,
      });
      assert.equal(outputs.gate_reason, reason);
    });
  }

  test('emits a single success reason when every gate passes', async () => {
    const outputs = await run({
      FACTORY_NAME: 'code-factory',
      ...eligibleTrusted,
      DUPLICATE_PR_FOUND: 'false',
      DUPLICATE_GATE_REASON: 'No open linked code-factory PR found.',
    });
    assert.deepEqual(Object.keys(outputs), ['gate_reason']);
    assert.equal(outputs.gate_reason, 'No open linked code-factory PR found.');
  });

  test('fails closed when actor trust was never determined', async () => {
    const outputs = await run({
      FACTORY_NAME: 'code-factory',
      EVENT_ELIGIBLE: 'true',
      EVENT_ELIGIBLE_REASON: 'Event is eligible.',
      DUPLICATE_PR_FOUND: 'false',
    });
    assert.match(outputs.gate_reason, /Actor trust could not be determined/);
  });

  test('fails closed when the duplicate-PR check produced no output', async () => {
    const outputs = await run({
      FACTORY_NAME: 'code-factory',
      ...eligibleTrusted,
    });
    assert.match(outputs.gate_reason, /Duplicate PR check did not complete/);
  });

  test('fails closed when the duplicate-PR output is not exactly true or false', async () => {
    const outputs = await run({
      FACTORY_NAME: 'code-factory',
      ...eligibleTrusted,
      DUPLICATE_PR_FOUND: 'garbage',
      DUPLICATE_GATE_REASON: 'No open linked code-factory PR found.',
    });
    assert.match(outputs.gate_reason, /Duplicate PR check did not complete/);
    assert.doesNotMatch(outputs.gate_reason, /All deterministic gates passed/);
  });

  test('research-factory forces the duplicate result off', async () => {
    const outputs = await run({
      FACTORY_NAME: 'research-factory',
      ...eligibleTrusted,
      DUPLICATE_PR_FOUND: 'true',
      DUPLICATE_PR_URL: 'https://example.com/pr/101',
      DUPLICATE_GATE_REASON: 'Found existing linked research-factory PR #101.',
    });
    assert.match(outputs.gate_reason, /All deterministic gates passed/);
    assert.doesNotMatch(outputs.gate_reason, /PR #101/);
  });
});
