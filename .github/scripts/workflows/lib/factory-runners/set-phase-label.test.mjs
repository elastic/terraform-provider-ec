import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const setPhaseLabelRunner = require('./set-phase-label.js');

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

/**
 * @param {{ issueNumber?: string, labelName?: string, currentLabels?: string[] }} opts
 */
async function run({ issueNumber, labelName, currentLabels = [] }) {
  const calls = { addLabels: [], removeLabel: [], listLabelsOnIssue: [] };
  const listLabelsOnIssue = async (params) => {
    calls.listLabelsOnIssue.push(params);
    return { data: currentLabels.map(name => ({ name })) };
  };
  const github = {
    rest: {
      issues: {
        addLabels: async (params) => { calls.addLabels.push(params); },
        removeLabel: async (params) => { calls.removeLabel.push(params); },
        listLabelsOnIssue,
      },
    },
  };
  const outputs = {};
  const infos = [];
  const failures = [];
  await withEnv({ INPUT_ISSUE_NUMBER: issueNumber, PHASE_LABEL_NAME: labelName }, () =>
    setPhaseLabelRunner({
      github,
      context: { eventName: 'issues', payload: {}, repo: { owner: 'elastic', repo: 'terraform-provider-ec' } },
      core: {
        setOutput(key, value) { outputs[key] = value; },
        info(message) { infos.push(message); },
        warning() {},
        setFailed(message) { failures.push(message); },
      },
    }),
  );
  return { outputs, calls, infos, failures };
}

describe('set-phase-label runner', { concurrency: false }, () => {
  test('sets the phase label and reports the stale labels it removed', async () => {
    const { outputs, calls, failures } = await run({
      issueNumber: '42',
      labelName: 'phase-specification',
      currentLabels: ['bug', 'phase-research', 'phase-specification', 'phase-triage'],
    });
    assert.deepEqual(calls.addLabels, [{
      owner: 'elastic',
      repo: 'terraform-provider-ec',
      issue_number: 42,
      labels: ['phase-specification'],
    }]);
    assert.deepEqual(calls.removeLabel.map(c => c.name), ['phase-research', 'phase-triage']);
    assert.equal(outputs.phase_label_set, 'true');
    assert.equal(outputs.phase_label_name, 'phase-specification');
    assert.deepEqual(JSON.parse(outputs.stale_labels_removed), ['phase-research', 'phase-triage']);
    assert.match(outputs.reason, /Set phase label phase-specification/);
    assert.deepEqual(failures, []);
  });

  test('reports an empty stale list when there is nothing to clear', async () => {
    const { outputs } = await run({
      issueNumber: '42',
      labelName: 'phase-research',
      currentLabels: ['phase-research'],
    });
    assert.equal(outputs.phase_label_set, 'true');
    assert.deepEqual(JSON.parse(outputs.stale_labels_removed), []);
  });

  test('echoes the reason through core.info', async () => {
    const { outputs, infos } = await run({ issueNumber: '42', labelName: 'phase-specification' });
    assert.ok(infos.includes(outputs.reason));
  });

  for (const [label, raw] of [['missing', undefined], ['empty', ''], ['fractional', '1.5'], ['zero-padded', '007'], ['non-numeric', 'abc'], ['negative', '-3'], ['spaced', ' 42 ']]) {
    test(`refuses a ${label} issue number without throwing`, async () => {
      const { outputs, calls, failures } = await run({ issueNumber: raw, labelName: 'phase-specification' });
      assert.equal(outputs.phase_label_set, 'false');
      assert.deepEqual(calls.addLabels, []);
      assert.match(outputs.reason, /No usable issue number/);
      assert.deepEqual(failures, []);
    });
  }

  test('refuses a label name the phase-label helper rejects', async () => {
    const { outputs, calls, failures } = await run({ issueNumber: '42', labelName: 'bug' });
    assert.equal(outputs.phase_label_set, 'false');
    assert.equal(outputs.phase_label_name, 'bug');
    assert.deepEqual(calls.addLabels, []);
    assert.match(outputs.reason, /must start with "phase-"/);
    assert.deepEqual(failures, []);
  });

  test('refuses a missing label name', async () => {
    const { outputs, calls } = await run({ issueNumber: '42', labelName: undefined });
    assert.equal(outputs.phase_label_set, 'false');
    assert.deepEqual(calls.addLabels, []);
    assert.match(outputs.reason, /No phase label name provided/);
  });

  test('reports an addLabels failure instead of throwing', async () => {
    const outputs = {};
    const failures = [];
    await withEnv({ INPUT_ISSUE_NUMBER: '42', PHASE_LABEL_NAME: 'phase-specification' }, () =>
      setPhaseLabelRunner({
        github: {
          rest: {
            issues: {
              addLabels: async () => { throw new Error('boom'); },
              removeLabel: async () => {},
              listLabelsOnIssue: async () => ({ data: [] }),
            },
          },
        },
        context: { eventName: 'issues', payload: {}, repo: { owner: 'elastic', repo: 'terraform-provider-ec' } },
        core: {
          setOutput(key, value) { outputs[key] = value; },
          info() {},
          warning() {},
          setFailed(message) { failures.push(message); },
        },
      }),
    );
    assert.equal(outputs.phase_label_set, 'false');
    assert.match(outputs.reason, /Failed to add label: boom/);
    assert.deepEqual(failures, []);
  });
});
