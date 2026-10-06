import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const fetchLiveIssue = require('./fetch-live-issue.js');

function mockCore() {
  const outputs = {};
  const failed = [];
  return {
    outputs,
    failed,
    core: {
      setOutput(key, value) {
        outputs[key] = value;
      },
      setFailed(message) {
        failed.push(message);
      },
      info() {},
    },
  };
}

function contextWithUnreadPayload() {
  let payloadReads = 0;
  const context = new Proxy(
    { repo: { owner: 'elastic', repo: 'terraform-provider-ec' } },
    {
      get(target, prop, receiver) {
        if (prop === 'payload') payloadReads += 1;
        return Reflect.get(target, prop, receiver);
      },
    },
  );
  return {
    context,
    payloadReads: () => payloadReads,
  };
}

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

describe('fetch-live-issue', { concurrency: false }, () => {
  test('reads the live API issue and ignores the webhook payload', async () => {
    const { context, payloadReads } = contextWithUnreadPayload();
    const { outputs, failed, core } = mockCore();
    let captured;
    const github = {
      rest: {
        issues: {
          get: async (args) => {
            captured = args;
            return { data: { number: 4372, title: 'Live title', body: 'Live body' } };
          },
        },
      },
    };

    await withEnv({ INPUT_ISSUE_NUMBER: '4372' }, () => fetchLiveIssue({ github, context, core }));

    assert.deepEqual(captured, { owner: 'elastic', repo: 'terraform-provider-ec', issue_number: 4372 });
    assert.equal(payloadReads(), 0);
    assert.equal(outputs.issue_number, '4372');
    assert.equal(outputs.issue_title, 'Live title');
    assert.equal(outputs.issue_body, 'Live body');
    assert.equal(outputs.fetch_error, '');
    assert.deepEqual(failed, []);
  });

  test('maps a null title and body to empty strings', async () => {
    const { context, payloadReads } = contextWithUnreadPayload();
    const { outputs, core } = mockCore();
    const github = {
      rest: {
        issues: {
          get: async () => ({ data: { number: 9, title: null, body: null } }),
        },
      },
    };

    await withEnv({ INPUT_ISSUE_NUMBER: '9' }, () => fetchLiveIssue({ github, context, core }));

    assert.equal(payloadReads(), 0);
    assert.equal(outputs.issue_title, '');
    assert.equal(outputs.issue_body, '');
    assert.equal(outputs.issue_number, '9');
  });

  test('rejects issue numbers that are not a positive integer', async () => {
    const cases = [undefined, '', '0', '-3', 'nope', '15oops', '15.9', '01', '1e2', ' 15', '9007199254740993'];
    for (const issueNumber of cases) {
      const { context, payloadReads } = contextWithUnreadPayload();
      const { outputs, failed, core } = mockCore();
      let called = false;
      const github = {
        rest: { issues: { get: async () => { called = true; return { data: {} }; } } },
      };

      await withEnv({ INPUT_ISSUE_NUMBER: issueNumber }, () => fetchLiveIssue({ github, context, core }));

      assert.equal(called, false, `API called for ${JSON.stringify(issueNumber)}`);
      assert.equal(payloadReads(), 0);
      assert.equal(outputs.issue_number, '');
      assert.equal(outputs.issue_title, '');
      assert.equal(outputs.issue_body, '');
      assert.equal(outputs.fetch_error, 'Invalid issue number in dispatch inputs.');
      assert.equal(failed.length, 1);
    }
  });

  test('fails closed when the API request throws', async () => {
    const { context, payloadReads } = contextWithUnreadPayload();
    const { outputs, failed, core } = mockCore();
    const github = {
      rest: { issues: { get: async () => { throw new Error('Not Found'); } } },
    };

    await withEnv({ INPUT_ISSUE_NUMBER: '15' }, () => fetchLiveIssue({ github, context, core }));

    assert.equal(payloadReads(), 0);
    assert.equal(outputs.issue_number, '');
    assert.equal(outputs.issue_title, '');
    assert.equal(outputs.issue_body, '');
    assert.equal(outputs.fetch_error, 'Not Found');
    assert.match(failed[0], /Failed to fetch issue #15: Not Found/);
  });
});
