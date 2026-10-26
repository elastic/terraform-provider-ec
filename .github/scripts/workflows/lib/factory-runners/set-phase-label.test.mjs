import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const setPhaseLabelRunner = require('./set-phase-label.js');

function mockCore() {
  const outputs = {};
  const failed = [];
  const warnings = [];
  return {
    outputs,
    failed,
    warnings,
    core: {
      setOutput(key, value) {
        outputs[key] = value;
      },
      setFailed(message) {
        failed.push(message);
      },
      warning(message) {
        warnings.push(message);
      },
      info() {},
    },
  };
}

const context = { repo: { owner: 'elastic', repo: 'terraform-provider-ec' } };

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
 * @param {{ labels?: string[], addThrows?: boolean, listThrows?: boolean, removeThrows?: string[] }} opts
 */
function mockGithub(opts = {}) {
  const added = [];
  const removed = [];
  const removeThrows = new Set(opts.removeThrows || []);
  return {
    added,
    removed,
    github: {
      async paginate(_route, _params) {
        if (opts.listThrows) throw new Error('list failed');
        return (opts.labels || []).map((name) => ({ name }));
      },
      rest: {
        issues: {
          listLabelsOnIssue: 'listLabelsOnIssue',
          async addLabels({ labels }) {
            if (opts.addThrows) throw new Error('add failed');
            added.push(...labels);
          },
          async removeLabel({ name }) {
            if (removeThrows.has(name)) {
              const err = new Error(`remove ${name} failed`);
              err.status = 500;
              throw err;
            }
            removed.push(name);
          },
        },
      },
    },
  };
}

describe('set-phase-label runner', { concurrency: false }, () => {
  test('applies the label and removes stale phase labels', async () => {
    const { core, outputs, failed } = mockCore();
    const { github, added, removed } = mockGithub({
      labels: ['phase-research', 'phase-specification', 'triaged'],
    });

    await withEnv({ INPUT_ISSUE_NUMBER: '42', PHASE_LABEL_NAME: 'phase-research' }, () =>
      setPhaseLabelRunner({ github, context, core }),
    );

    assert.deepEqual(failed, []);
    assert.equal(outputs.phase_label_set, 'true');
    assert.equal(outputs.phase_label_name, 'phase-research');
    assert.equal(outputs.stale_labels_removed, 'phase-specification');
    assert.deepEqual(added, ['phase-research']);
    assert.deepEqual(removed, ['phase-specification']);
  });

  test('fails when a stale label could not be removed, even though the label was added', async () => {
    const { core, outputs, failed } = mockCore();
    const { github, added } = mockGithub({
      labels: ['phase-research', 'phase-specification'],
      removeThrows: ['phase-specification'],
    });

    await withEnv({ INPUT_ISSUE_NUMBER: '42', PHASE_LABEL_NAME: 'phase-research' }, () =>
      setPhaseLabelRunner({ github, context, core }),
    );

    assert.deepEqual(added, ['phase-research']);
    assert.equal(outputs.phase_label_set, 'false');
    assert.equal(failed.length, 1);
    assert.match(failed[0], /not fully applied to #42/);
  });

  test('fails when the label list could not be read after the add', async () => {
    const { core, outputs, failed } = mockCore();
    const { github } = mockGithub({ listThrows: true });

    await withEnv({ INPUT_ISSUE_NUMBER: '42', PHASE_LABEL_NAME: 'phase-research' }, () =>
      setPhaseLabelRunner({ github, context, core }),
    );

    assert.equal(outputs.phase_label_set, 'false');
    assert.equal(failed.length, 1);
    assert.match(failed[0], /failed to list current labels/);
  });

  test('fails when the label could not be added', async () => {
    const { core, outputs, failed } = mockCore();
    const { github } = mockGithub({ addThrows: true });

    await withEnv({ INPUT_ISSUE_NUMBER: '42', PHASE_LABEL_NAME: 'phase-research' }, () =>
      setPhaseLabelRunner({ github, context, core }),
    );

    assert.equal(outputs.phase_label_set, 'false');
    assert.equal(failed.length, 1);
    assert.match(failed[0], /Failed to add label/);
  });

  test('rejects a non-integer dispatch issue number instead of stamping #1', async () => {
    const { core, outputs, failed } = mockCore();
    const { github, added } = mockGithub({ labels: [] });

    await withEnv({ INPUT_ISSUE_NUMBER: '1.5', PHASE_LABEL_NAME: 'phase-research' }, () =>
      setPhaseLabelRunner({ github, context, core }),
    );

    assert.deepEqual(added, []);
    assert.equal(outputs.phase_label_set, 'false');
    assert.equal(failed.length, 1);
    assert.match(failed[0], /invalid issue number '1\.5'/);
  });

  test('fails on a missing phase label name', async () => {
    const { core, outputs, failed } = mockCore();
    const { github, added } = mockGithub({ labels: [] });

    await withEnv({ INPUT_ISSUE_NUMBER: '42', PHASE_LABEL_NAME: '' }, () =>
      setPhaseLabelRunner({ github, context, core }),
    );

    assert.deepEqual(added, []);
    assert.equal(outputs.phase_label_set, 'false');
    assert.equal(failed.length, 1);
    assert.match(failed[0], /No phase label name provided/);
  });

  test('reports success when there is no stale phase label to remove', async () => {
    const { core, outputs, failed } = mockCore();
    const { github, removed } = mockGithub({ labels: ['phase-research'] });

    await withEnv({ INPUT_ISSUE_NUMBER: '42', PHASE_LABEL_NAME: 'phase-research' }, () =>
      setPhaseLabelRunner({ github, context, core }),
    );

    assert.deepEqual(failed, []);
    assert.deepEqual(removed, []);
    assert.equal(outputs.phase_label_set, 'true');
    assert.equal(outputs.stale_labels_removed, '');
  });
});
