import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const setPhaseLabelRunner = require('./set-phase-label.js');

function mockCore() {
  const outputs = {};
  return {
    outputs,
    core: {
      setOutput(key, value) { outputs[key] = value; },
      info() {},
      warning() {},
    },
  };
}

// github mock recording addLabels / removeLabel calls; listLabelsOnIssue returns `existing`.
function mockGithub(existing = []) {
  const calls = { added: [], removed: [] };
  return {
    calls,
    github: {
      paginate: async (_fn, _params) => existing.map((name) => ({ name })),
      rest: {
        issues: {
          addLabels: async ({ labels }) => { calls.added.push(...labels); },
          listLabelsOnIssue: 'listLabelsOnIssue',
          removeLabel: async ({ name }) => { calls.removed.push(name); },
        },
      },
    },
  };
}

const context = { repo: { owner: 'o', repo: 'r' }, payload: {} };

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

test('set-phase-label applies phase-coding and strips other phase-* labels', async () => {
  const { github, calls } = mockGithub(['phase-coding', 'phase-specification', 'bug']);
  await withEnv({ INPUT_ISSUE_NUMBER: '42', PHASE_LABEL_NAME: 'phase-coding' }, async () => {
    const { outputs, core } = mockCore();
    await setPhaseLabelRunner({ github, context, core });
    assert.equal(outputs.phase_label_set, 'true');
    assert.equal(outputs.phase_label_name, 'phase-coding');
    assert.deepEqual(calls.added, ['phase-coding']);
    assert.deepEqual(calls.removed, ['phase-specification']);
  });
});

test('set-phase-label falls back to the issue payload number', async () => {
  const { github, calls } = mockGithub(['phase-coding']);
  const ctx = { repo: { owner: 'o', repo: 'r' }, payload: { issue: { number: 7 } } };
  await withEnv({ INPUT_ISSUE_NUMBER: undefined, PHASE_LABEL_NAME: 'phase-coding' }, async () => {
    const { outputs, core } = mockCore();
    await setPhaseLabelRunner({ github, context: ctx, core });
    assert.equal(outputs.phase_label_set, 'true');
    assert.deepEqual(calls.added, ['phase-coding']);
  });
});

test('set-phase-label reports not-set when no issue number is available', async () => {
  const { github } = mockGithub([]);
  await withEnv({ INPUT_ISSUE_NUMBER: undefined, PHASE_LABEL_NAME: 'phase-coding' }, async () => {
    const { outputs, core } = mockCore();
    await setPhaseLabelRunner({ github, context, core });
    assert.equal(outputs.phase_label_set, 'false');
  });
});
