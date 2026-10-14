import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const classifyIssues = require('./classify-issues.js');

function mockCore() {
  const outputs = {};
  const failures = [];
  return {
    outputs,
    failures,
    core: {
      setOutput(key, value) {
        outputs[key] = value;
      },
      setFailed(message) {
        failures.push(message);
      },
      info() {},
    },
  };
}

function issue(number, title, labels = []) {
  return { number, title, labels: labels.map((name) => ({ name })) };
}

async function run({ eventName, payload, github }) {
  const { outputs, failures, core } = mockCore();
  await classifyIssues({
    github: github ?? { paginate: async () => [], rest: { issues: {} } },
    context: {
      eventName,
      payload: payload ?? {},
      repo: { owner: 'elastic', repo: 'terraform-provider-ec' },
    },
    core,
  });
  return { outputs, failures };
}

describe('classify-issues', () => {
  test('skips an opened issue that already has triaged', async () => {
    const { outputs, failures } = await run({
      eventName: 'issues',
      payload: { issue: issue(12, 'scanner finding', ['triaged', 'flaky-test']) },
    });
    assert.deepEqual(failures, []);
    assert.equal(outputs.mode, 'event');
    assert.equal(outputs.issue_count, '0');
    assert.equal(outputs.issues_json, '[]');
    assert.match(outputs.gate_reason, /already has the triaged label/);
  });

  test('selects an opened untriaged issue', async () => {
    const { outputs } = await run({
      eventName: 'issues',
      payload: { issue: issue(7, 'Add traffic filter data source') },
    });
    assert.equal(outputs.mode, 'event');
    assert.equal(outputs.issue_count, '1');
    assert.deepEqual(JSON.parse(outputs.issues_json), [{ number: 7, title: 'Add traffic filter data source' }]);
    assert.match(outputs.gate_reason, /untriaged/);
  });

  test('skips an opened pull request', async () => {
    const pr = issue(9, 'some pr');
    pr.pull_request = { url: 'https://example.test/pulls/9' };
    const { outputs } = await run({
      eventName: 'issues',
      payload: { issue: pr },
    });
    assert.equal(outputs.issue_count, '0');
    assert.equal(outputs.issues_json, '[]');
    assert.match(outputs.gate_reason, /pull request/);
  });

  test('fails when the issues payload has no issue', async () => {
    const { outputs, failures } = await run({
      eventName: 'issues',
      payload: {},
    });
    assert.deepEqual(failures, ['issues event payload missing issue object']);
    assert.equal(outputs.issue_count, undefined);
  });

  test('fails dispatch when issue_number is not a positive integer', async () => {
    for (const issueNumber of ['0', '1.5', '007', '-3', 'abc', '1e2']) {
      const { failures, outputs } = await run({
        eventName: 'workflow_dispatch',
        payload: { inputs: { issue_number: issueNumber } },
      });
      assert.deepEqual(failures, [`Invalid issue_number input: "${issueNumber}"`]);
      assert.equal(outputs.issue_count, undefined);
    }
  });

  test('skips a dispatched issue that is already triaged', async () => {
    const { outputs } = await run({
      eventName: 'workflow_dispatch',
      payload: { inputs: { issue_number: 15 } },
      github: {
        rest: {
          issues: {
            get: async (params) => {
              assert.deepEqual(params, {
                owner: 'elastic',
                repo: 'terraform-provider-ec',
                issue_number: 15,
              });
              return { data: issue(15, 'done', ['triaged']) };
            },
          },
        },
      },
    });
    assert.equal(outputs.mode, 'dispatch');
    assert.equal(outputs.issue_count, '0');
    assert.match(outputs.gate_reason, /already has the triaged label/);
  });

  test('skips a dispatched pull request', async () => {
    const pr = issue(4, 'not an issue');
    pr.pull_request = {};
    const { outputs } = await run({
      eventName: 'workflow_dispatch',
      payload: { inputs: { issue_number: '4' } },
      github: {
        rest: { issues: { get: async () => ({ data: pr }) } },
      },
    });
    assert.equal(outputs.mode, 'dispatch');
    assert.equal(outputs.issue_count, '0');
    assert.match(outputs.gate_reason, /pull request/);
  });

  test('selects a dispatched untriaged issue', async () => {
    const { outputs } = await run({
      eventName: 'workflow_dispatch',
      payload: { inputs: { issue_number: '21' } },
      github: {
        rest: { issues: { get: async () => ({ data: issue(21, 'Bug with config') }) } },
      },
    });
    assert.equal(outputs.mode, 'dispatch');
    assert.equal(outputs.issue_count, '1');
    assert.deepEqual(JSON.parse(outputs.issues_json), [{ number: 21, title: 'Bug with config' }]);
  });

  test('dispatch without issue_number uses the scheduled selection and keeps mode dispatch', async () => {
    const listed = [];
    const { outputs } = await run({
      eventName: 'workflow_dispatch',
      payload: { inputs: {} },
      github: githubListing(
        [
          issue(3, 'newest'),
          issue(2, 'older', ['triaged']),
          issue(1, 'oldest'),
        ],
        listed,
      ),
    });
    assert.equal(outputs.mode, 'dispatch');
    assert.equal(outputs.issue_count, '2');
    assert.deepEqual(JSON.parse(outputs.issues_json), [
      { number: 3, title: 'newest' },
      { number: 1, title: 'oldest' },
    ]);
    assert.equal(listed.length, 1);
    assert.equal(listed[0].state, 'open');
    assert.equal(listed[0].sort, 'created');
    assert.equal(listed[0].direction, 'desc');
    assert.equal(listed[0].per_page, 100);
  });

  test('scheduled sweep drops pull requests and triaged issues and caps at 5 newest', async () => {
    const pr = issue(100, 'pr');
    pr.pull_request = {};
    const open = [pr, issue(9, 't', ['triaged'])];
    for (let n = 8; n >= 1; n -= 1) {
      open.push(issue(n, `issue ${n}`));
    }
    const { outputs } = await run({
      eventName: 'schedule',
      github: githubListing(open),
    });
    assert.equal(outputs.mode, 'scheduled');
    assert.equal(outputs.issue_count, '5');
    assert.deepEqual(
      JSON.parse(outputs.issues_json).map((item) => item.number),
      [8, 7, 6, 5, 4],
    );
    assert.match(outputs.gate_reason, /classifying up to 5/);
  });

  test('scheduled sweep with nothing untriaged sets count 0', async () => {
    const pr = issue(2, 'pr');
    pr.pull_request = {};
    const { outputs } = await run({
      eventName: 'schedule',
      github: githubListing([pr, issue(1, 'done', ['triaged'])]),
    });
    assert.equal(outputs.issue_count, '0');
    assert.equal(outputs.issues_json, '[]');
    assert.match(outputs.gate_reason, /No untriaged open issues/);
  });

  test('unknown events use the scheduled path', async () => {
    const { outputs } = await run({
      eventName: 'push',
      github: githubListing([issue(1, 'only')]),
    });
    assert.equal(outputs.mode, 'scheduled');
    assert.equal(outputs.issue_count, '1');
  });
});

function githubListing(issues, listed = []) {
  const listForRepo = () => {};
  return {
    rest: { issues: { listForRepo } },
    paginate: async (fn, params) => {
      assert.equal(fn, listForRepo);
      listed.push(params);
      return issues;
    },
  };
}
