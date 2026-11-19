import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const {
  getFactoryName,
  getFactoryConstants,
  getFactoryModule,
  getFactoryContextDir,
} = require('./_factory-context.js');

const FACTORY_NAMES = ['change-factory', 'code-factory', 'research-factory', 'reproducer-factory'];

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

describe('_factory-context', { concurrency: false }, () => {
  test('loads without factory-issue-shared.js', () => {
    assert.equal(typeof getFactoryModule, 'function');
  });

  test('getFactoryName rejects a missing or unknown FACTORY_NAME', async () => {
    await withEnv({ FACTORY_NAME: undefined }, async () => {
      assert.throws(() => getFactoryName(), /FACTORY_NAME environment variable is required/);
    });
    await withEnv({ FACTORY_NAME: '' }, async () => {
      assert.throws(() => getFactoryName(), /FACTORY_NAME environment variable is required/);
    });
    for (const name of ['not-a-factory', '../../etc', 'constructor', '__proto__']) {
      await withEnv({ FACTORY_NAME: name }, async () => {
        assert.throws(() => getFactoryName(), /Unknown FACTORY_NAME/);
      });
    }
  });

  test('context directory is /tmp/<factory>-context for each allowlisted factory', async () => {
    for (const name of FACTORY_NAMES) {
      await withEnv({ FACTORY_NAME: name }, async () => {
        assert.equal(getFactoryName(), name);
        assert.equal(getFactoryContextDir(), `/tmp/${name}-context`);
        assert.equal(getFactoryConstants().FACTORY_LABEL, name);
      });
    }
  });

  test('PR-opening factories bind branch prefix and linkage mode', async () => {
    const expected = {
      'change-factory': { prefix: 'change-factory/issue-', mode: 'related-literal' },
      'code-factory': { prefix: 'code-factory/issue-', mode: 'closes-literal' },
      'reproducer-factory': { prefix: 'reproducer-factory/issue-', mode: 'related-literal' },
    };
    for (const [name, spec] of Object.entries(expected)) {
      await withEnv({ FACTORY_NAME: name }, async () => {
        const constants = getFactoryConstants();
        assert.equal(constants.ISSUE_BRANCH_PREFIX, spec.prefix);
        assert.equal(constants.DUPLICATE_LINKAGE_MODE, spec.mode);
        assert.equal(getFactoryModule().issueBranchName(42), `${spec.prefix}42`);
      });
    }
  });

  test('getFactoryModule returns the gate functions', async () => {
    await withEnv({ FACTORY_NAME: 'code-factory' }, async () => {
      const mod = getFactoryModule();
      assert.equal(typeof mod.qualifyTriggerEvent, 'function');
      assert.equal(typeof mod.checkDuplicatePR, 'function');
      assert.equal(typeof mod.computeGateReason, 'function');
      assert.equal(typeof mod.parseFinalizeGateEnv, 'function');
      assert.equal(mod.issueBranchName(42), 'code-factory/issue-42');
    });
  });
});
