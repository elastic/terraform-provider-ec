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

  test('getFactoryModule fails until factory-issue-shared.js exists', async () => {
    await withEnv({ FACTORY_NAME: 'code-factory' }, async () => {
      assert.throws(() => getFactoryModule(), (err) => {
        assert.equal(err.code, 'MODULE_NOT_FOUND');
        assert.match(err.message, /factory-issue-shared/);
        return true;
      });
    });
  });
});
