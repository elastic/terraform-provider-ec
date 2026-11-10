import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const notifyDuplicateBlocked = require('./notify-duplicate-blocked.js');

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

async function run({ factory = 'change-factory', url, contractMatch, issueNumber }) {
  const calls = [];
  const github = {
    rest: { issues: { createComment: async (params) => { calls.push(params); } } },
  };
  await withEnv(
    {
      FACTORY_NAME: factory,
      DUPLICATE_PR_URL: url,
      DUPLICATE_PR_CONTRACT_MATCH: contractMatch,
      ISSUE_NUMBER: issueNumber,
    },
    () => notifyDuplicateBlocked({
      github,
      context: { eventName: 'issues', payload: {}, repo: { owner: 'elastic', repo: 'terraform-provider-ec' } },
      core: { info() {} },
    }),
  );
  return calls;
}

describe('notify-duplicate-blocked', { concurrency: false }, () => {
  test('tells the maintainer to close the factory-owned PR and retry', async () => {
    const calls = await run({
      url: 'https://github.com/elastic/terraform-provider-ec/pull/101',
      contractMatch: 'true',
      issueNumber: '42',
    });
    assert.equal(calls.length, 1);
    assert.equal(calls[0].owner, 'elastic');
    assert.equal(calls[0].repo, 'terraform-provider-ec');
    assert.equal(calls[0].issue_number, 42);
    assert.match(calls[0].body, /\*\*change-factory skipped\*\* — PR #101 is already open for this issue/);
    assert.match(calls[0].body, /re-apply the `change-factory` label \(or comment `\/change-factory`\) to retry/);
    assert.equal(calls[0].body.includes('move it to another branch'), false);
  });

  // The blocker may be a human's own PR, so this wording must not assume it is disposable.
  test('offers the move-the-branch escape when the PR is not contract-matching', async () => {
    const calls = await run({
      url: 'https://github.com/elastic/terraform-provider-ec/pull/202',
      contractMatch: 'false',
      issueNumber: '7',
    });
    assert.equal(calls.length, 1);
    assert.match(calls[0].body, /PR #202 is already open on branch `change-factory\/issue-7`/);
    assert.match(calls[0].body, /missing the `change-factory` label or the `Related to #7` line/);
    assert.match(calls[0].body, /Close it \*\*or move it to another branch\*\*/);
  });

  test('treats any non-true contract-match value as the cautious wording', async () => {
    for (const contractMatch of ['', 'TRUE', 'yes', undefined]) {
      const calls = await run({
        url: 'https://github.com/elastic/terraform-provider-ec/pull/202',
        contractMatch,
        issueNumber: '7',
      });
      assert.match(calls[0].body, /move it to another branch/, `contractMatch=${contractMatch}`);
    }
  });

  test('builds the strings from FACTORY_NAME', async () => {
    const calls = await run({
      factory: 'code-factory',
      url: 'https://github.com/elastic/terraform-provider-ec/pull/303',
      contractMatch: 'false',
      issueNumber: '9',
    });
    assert.match(calls[0].body, /\*\*code-factory skipped\*\*/);
    assert.match(calls[0].body, /branch `code-factory\/issue-9`/);
    assert.match(calls[0].body, /`Closes #9` line/);
  });

  test('falls back to the raw URL when no PR number can be extracted', async () => {
    const calls = await run({ url: 'https://example.com/pulls', contractMatch: 'true', issueNumber: '42' });
    assert.match(calls[0].body, /PR #https:\/\/example\.com\/pulls is already open/);
  });

  test('posts nothing when DUPLICATE_PR_URL is empty', async () => {
    assert.deepEqual(await run({ url: '', contractMatch: 'true', issueNumber: '42' }), []);
    assert.deepEqual(await run({ url: undefined, contractMatch: 'true', issueNumber: '42' }), []);
  });

  // A failure here must not fail the job: the steps that follow consume the trigger label, and
  // skipping them would break the documented retry path.
  test('reports a createComment failure instead of throwing', async () => {
    const warnings = [];
    await withEnv(
      {
        FACTORY_NAME: 'change-factory',
        DUPLICATE_PR_URL: 'https://github.com/elastic/terraform-provider-ec/pull/101',
        DUPLICATE_PR_CONTRACT_MATCH: 'true',
        ISSUE_NUMBER: '42',
      },
      () => notifyDuplicateBlocked({
        github: { rest: { issues: { createComment: async () => { throw new Error('rate limited'); } } } },
        context: { eventName: 'issues', payload: {}, repo: { owner: 'elastic', repo: 'terraform-provider-ec' } },
        core: { info() {}, warning(message) { warnings.push(message); } },
      }),
    );
    assert.equal(warnings.length, 1);
    assert.match(warnings[0], /Could not post the duplicate-blocked comment on issue #42: rate limited/);
  });

  test('falls back to core.info when the core mock has no warning', async () => {
    const infos = [];
    await withEnv(
      {
        FACTORY_NAME: 'change-factory',
        DUPLICATE_PR_URL: 'https://github.com/elastic/terraform-provider-ec/pull/101',
        DUPLICATE_PR_CONTRACT_MATCH: 'true',
        ISSUE_NUMBER: '42',
      },
      () => notifyDuplicateBlocked({
        github: { rest: { issues: { createComment: async () => { throw new Error('boom'); } } } },
        context: { eventName: 'issues', payload: {}, repo: { owner: 'elastic', repo: 'terraform-provider-ec' } },
        core: { info(message) { infos.push(message); } },
      }),
    );
    assert.ok(infos.some(m => /Could not post the duplicate-blocked comment/.test(m)));
  });

  test('posts nothing when ISSUE_NUMBER is empty', async () => {
    const url = 'https://github.com/elastic/terraform-provider-ec/pull/101';
    assert.deepEqual(await run({ url, contractMatch: 'true', issueNumber: '' }), []);
    assert.deepEqual(await run({ url, contractMatch: 'true', issueNumber: undefined }), []);
  });
});
