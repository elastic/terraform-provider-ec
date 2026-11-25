import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const fetchPriorReproducerComment = require('./fetch-prior-reproducer-comment.js');
const { REPRODUCER_COMMENT_MARKER } = require('../lib/intake/reproducer-factory-constants.js');

function mockCore() {
  const outputs = {};
  const warnings = [];
  return {
    outputs,
    warnings,
    core: {
      setOutput(key, value) {
        outputs[key] = value;
      },
      warning(message) {
        warnings.push(message);
      },
      info() {},
      setFailed(message) {
        warnings.push(`FAILED: ${message}`);
      },
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

function githubWith(comments, capture = {}) {
  return {
    rest: { issues: { listComments: 'listComments' } },
    async paginate(route, params) {
      capture.route = route;
      capture.params = params;
      if (comments instanceof Error) throw comments;
      return comments;
    },
  };
}

const restComment = (login, body, id = 1) => ({ id, user: { login }, body });

describe('fetch-prior-reproducer-comment', { concurrency: false }, () => {
  test('returns the latest marked bot comment', async () => {
    const { core, outputs } = mockCore();
    const github = githubWith([
      restComment('alice', 'a human comment', 1),
      restComment('github-actions[bot]', `${REPRODUCER_COMMENT_MARKER}\nold repro`, 2),
      restComment('github-actions[bot]', 'a status comment with no marker', 3),
      restComment('github-actions[bot]', `${REPRODUCER_COMMENT_MARKER}\nnewest repro`, 4),
    ]);

    await withEnv({ INPUT_ISSUE_NUMBER: '42', GH_AW_BOT_LOGIN: undefined }, () =>
      fetchPriorReproducerComment({ github, context, core }),
    );

    assert.match(outputs.prior_reproducer_comment, /newest repro/);
  });

  test('paginates without sort or direction so "last match" stays oldest-first', async () => {
    const capture = {};
    const { core } = mockCore();
    const github = githubWith([], capture);

    await withEnv({ INPUT_ISSUE_NUMBER: '42' }, () => fetchPriorReproducerComment({ github, context, core }));

    assert.equal(capture.route, 'listComments');
    assert.deepEqual(capture.params, {
      owner: 'elastic',
      repo: 'terraform-provider-ec',
      issue_number: 42,
      per_page: 100,
    });
    assert.equal('sort' in capture.params, false);
    assert.equal('direction' in capture.params, false);
  });

  test('ignores a marker-bearing comment from a human', async () => {
    const { core, outputs } = mockCore();
    const github = githubWith([restComment('attacker', `${REPRODUCER_COMMENT_MARKER}\nnot ours`, 1)]);

    await withEnv({ INPUT_ISSUE_NUMBER: '42' }, () => fetchPriorReproducerComment({ github, context, core }));

    assert.equal(outputs.prior_reproducer_comment, '');
  });

  test('falls back to the default login when GH_AW_BOT_LOGIN is set but empty', async () => {
    const { core, outputs } = mockCore();
    const github = githubWith([restComment('github-actions[bot]', `${REPRODUCER_COMMENT_MARKER}\nfound`, 1)]);

    await withEnv({ INPUT_ISSUE_NUMBER: '42', GH_AW_BOT_LOGIN: '' }, () =>
      fetchPriorReproducerComment({ github, context, core }),
    );

    assert.match(outputs.prior_reproducer_comment, /found/);
  });

  test('returns empty for an invalid issue number without calling the API', async () => {
    const capture = {};
    const { core, outputs } = mockCore();
    const github = githubWith([], capture);

    await withEnv({ INPUT_ISSUE_NUMBER: '1.5' }, () => fetchPriorReproducerComment({ github, context, core }));

    assert.equal(outputs.prior_reproducer_comment, '');
    assert.equal(capture.params, undefined);
  });

  test('warns rather than fails when the lookup throws', async () => {
    const { core, outputs, warnings } = mockCore();
    const github = githubWith(new Error('rate limited'));

    await withEnv({ INPUT_ISSUE_NUMBER: '42' }, () => fetchPriorReproducerComment({ github, context, core }));

    assert.equal(outputs.prior_reproducer_comment, '');
    assert.equal(warnings.length, 1);
    assert.match(warnings[0], /Could not fetch prior reproducer comment for issue #42: rate limited/);
  });

  test('bounds the prior comment in bytes before it becomes an environment variable', async () => {
    const { core, outputs, warnings } = mockCore();
    const big = `${REPRODUCER_COMMENT_MARKER}\n${'漢'.repeat(55_000)}`;
    const github = githubWith([restComment('github-actions[bot]', big, 1)]);

    await withEnv({ INPUT_ISSUE_NUMBER: '42' }, () => fetchPriorReproducerComment({ github, context, core }));

    assert.deepEqual(warnings, []);
    assert.ok(Buffer.byteLength(outputs.prior_reproducer_comment) < 128 * 1024);
    assert.ok(outputs.prior_reproducer_comment.startsWith(REPRODUCER_COMMENT_MARKER), 'still recognisable as ours');
    assert.ok(outputs.prior_reproducer_comment.isWellFormed());
  });

  test('a GraphQL-shaped comment list matches nothing, which is why the lookup is REST only', async () => {
    const { core, outputs } = mockCore();
    const graphqlShaped = [{ id: 1, author: { login: 'github-actions[bot]' }, body: `${REPRODUCER_COMMENT_MARKER}\nrepro` }];
    const github = githubWith(graphqlShaped);

    await withEnv({ INPUT_ISSUE_NUMBER: '42' }, () => fetchPriorReproducerComment({ github, context, core }));

    assert.equal(outputs.prior_reproducer_comment, '');
  });
});
