import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const fetchPriorResearchComment = require('./fetch-prior-research-comment.js');
const { RESEARCH_COMMENT_MARKER } = require('../lib/intake/research-factory-constants.js');

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

describe('fetch-prior-research-comment', { concurrency: false }, () => {
  test('returns the latest marked bot comment', async () => {
    const { core, outputs } = mockCore();
    const github = githubWith([
      restComment('alice', 'a human comment', 1),
      restComment('github-actions[bot]', `${RESEARCH_COMMENT_MARKER}\nold research`, 2),
      restComment('github-actions[bot]', 'a status comment with no marker', 3),
      restComment('github-actions[bot]', `${RESEARCH_COMMENT_MARKER}\nnewest research`, 4),
    ]);

    await withEnv({ INPUT_ISSUE_NUMBER: '42', GH_AW_BOT_LOGIN: undefined }, () =>
      fetchPriorResearchComment({ github, context, core }),
    );

    assert.match(outputs.prior_research_comment, /newest research/);
  });

  test('paginates without sort or direction so "last match" stays oldest-first', async () => {
    const capture = {};
    const { core } = mockCore();
    const github = githubWith([], capture);

    await withEnv({ INPUT_ISSUE_NUMBER: '42' }, () => fetchPriorResearchComment({ github, context, core }));

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
    const github = githubWith([restComment('attacker', `${RESEARCH_COMMENT_MARKER}\nnot ours`, 1)]);

    await withEnv({ INPUT_ISSUE_NUMBER: '42' }, () => fetchPriorResearchComment({ github, context, core }));

    assert.equal(outputs.prior_research_comment, '');
  });

  test('honours an overridden bot login', async () => {
    const { core, outputs } = mockCore();
    const github = githubWith([
      restComment('github-actions[bot]', `${RESEARCH_COMMENT_MARKER}\ndefault author`, 1),
      restComment('elastic-bot[bot]', `${RESEARCH_COMMENT_MARKER}\noverride author`, 2),
    ]);

    await withEnv({ INPUT_ISSUE_NUMBER: '42', GH_AW_BOT_LOGIN: 'elastic-bot[bot]' }, () =>
      fetchPriorResearchComment({ github, context, core }),
    );

    assert.match(outputs.prior_research_comment, /override author/);
  });

  test('falls back to the default login when GH_AW_BOT_LOGIN is set but empty', async () => {
    const { core, outputs } = mockCore();
    const github = githubWith([restComment('github-actions[bot]', `${RESEARCH_COMMENT_MARKER}\nfound`, 1)]);

    await withEnv({ INPUT_ISSUE_NUMBER: '42', GH_AW_BOT_LOGIN: '' }, () =>
      fetchPriorResearchComment({ github, context, core }),
    );

    assert.match(outputs.prior_research_comment, /found/);
  });

  test('returns empty for an invalid issue number without calling the API', async () => {
    const capture = {};
    const { core, outputs } = mockCore();
    const github = githubWith([], capture);

    await withEnv({ INPUT_ISSUE_NUMBER: '1.5' }, () => fetchPriorResearchComment({ github, context, core }));

    assert.equal(outputs.prior_research_comment, '');
    assert.equal(capture.params, undefined);
  });

  test('warns rather than fails when the lookup throws', async () => {
    const { core, outputs, warnings } = mockCore();
    const github = githubWith(new Error('rate limited'));

    await withEnv({ INPUT_ISSUE_NUMBER: '42' }, () => fetchPriorResearchComment({ github, context, core }));

    assert.equal(outputs.prior_research_comment, '');
    assert.equal(warnings.length, 1);
    assert.match(warnings[0], /Could not fetch prior research comment for issue #42: rate limited/);
  });

  test('a GraphQL-shaped comment list matches nothing, which is why the lookup is REST only', async () => {
    const { core, outputs } = mockCore();
    const graphqlShaped = [{ id: 1, author: { login: 'github-actions[bot]' }, body: `${RESEARCH_COMMENT_MARKER}\nresearch` }];
    const github = githubWith(graphqlShaped);

    await withEnv({ INPUT_ISSUE_NUMBER: '42' }, () => fetchPriorResearchComment({ github, context, core }));

    assert.equal(outputs.prior_research_comment, '');
  });
});
