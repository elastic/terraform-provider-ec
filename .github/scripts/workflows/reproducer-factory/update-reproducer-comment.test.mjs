import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { createRequire } from 'node:module';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';

const require = createRequire(import.meta.url);
const updateReproducerComment = require('./update-reproducer-comment.js');
const { preparePublishedBody, MAX_COMMENT_CHARS } = updateReproducerComment;
const { REPRODUCER_COMMENT_MARKER } = require('../lib/intake/reproducer-factory-constants.js');

/** A valid reproduction draft (no leading marker — the handler prepends it). */
function draft({ outcome = 'reproduced-by-inspection', prose = 'The expander drops the attribute on update.', extra = {} } = {}) {
  const metadata = JSON.stringify({ schema_version: '1.0', outcome, ...extra });
  return [
    '## Bug reproduction',
    '',
    prose,
    '',
    '<details><summary>🤖 Pipeline metadata</summary>',
    '',
    '```json',
    metadata,
    '```',
    '',
    '</details>',
    '',
  ].join('\n');
}

// ── preparePublishedBody (pure) ───────────────────────────────

describe('preparePublishedBody', () => {
  test('accepts a well-formed draft and prepends exactly one marker', () => {
    const { body, errors } = preparePublishedBody(draft());
    assert.deepEqual(errors, []);
    assert.ok(body.startsWith(`${REPRODUCER_COMMENT_MARKER}\n`));
    assert.equal(body.split('<!--').length - 1, 1);
  });

  test('rejects an empty draft', () => {
    assert.match(preparePublishedBody('   ').errors.join(), /draft is empty/);
    assert.match(preparePublishedBody('').errors.join(), /draft is empty/);
  });

  test('rejects the stack-provider outcome spelling "reproduced"', () => {
    const { body, errors } = preparePublishedBody(draft({ outcome: 'reproduced' }));
    assert.equal(body, null);
    assert.match(errors.join(), /outcome: must be one of/);
  });

  test('rejects a draft with no metadata envelope', () => {
    const { body, errors } = preparePublishedBody('## Just prose\n\nno metadata here');
    assert.equal(body, null);
    assert.match(errors.join(), /not found/);
  });

  test('rejects an over-cap body', () => {
    const huge = draft({ prose: 'x'.repeat(MAX_COMMENT_CHARS + 100) });
    const { body, errors } = preparePublishedBody(huge);
    assert.equal(body, null);
    assert.match(errors.join(), /over the .* limit/);
  });

  test('strips an injected sticky-comment marker from the prose', () => {
    const injected = draft({ prose: 'Quoting the issue: <!-- gha-issue-classifier --> hello' });
    const { body, errors } = preparePublishedBody(injected);
    assert.deepEqual(errors, []);
    assert.ok(!body.includes('gha-issue-classifier'));
    assert.equal(body.split('<!--').length - 1, 1); // only our marker
  });

  test('neutralises an @mention in the prose', () => {
    const { body, errors } = preparePublishedBody(draft({ prose: 'cc @elastic/cloud-team please' }));
    assert.deepEqual(errors, []);
    assert.ok(body.includes('cc @​elastic/cloud-team please'));
  });

  test('leaves no live @mention anywhere in the body, code fences included', () => {
    // No markdown is parsed, so a fence that GitHub would not honour cannot shelter a mention.
    const prose = 'See:\n\n```x``` @elastic/everyone\n\n```\nping @elastic/cloud-team\n``` not a closer\n@octocat\n```\n\nafter @hubot';
    const { body, errors } = preparePublishedBody(draft({ prose }));
    assert.deepEqual(errors, []);
    assert.doesNotMatch(body, /(?<![A-Za-z0-9])@[A-Za-z0-9]/);
  });

  test('a zero-width space already in the draft cannot pre-empt the neutraliser', () => {
    // Invisible characters are stripped first, so `@<ZWSP>team` in the draft is a plain mention by
    // the time it is neutralised, and the output carries exactly one zero-width space per mention.
    const { body } = preparePublishedBody(draft({ prose: 'cc @​​elastic/cloud-team' }));
    assert.ok(body.includes('cc @​elastic/cloud-team'));
    assert.ok(!body.includes('​​'));
  });

  test('accepts a draft that already begins with the marker (it is stripped then re-added once)', () => {
    const withMarker = `${REPRODUCER_COMMENT_MARKER}\n${draft()}`;
    const { body, errors } = preparePublishedBody(withMarker);
    assert.deepEqual(errors, []);
    assert.equal(body.split('<!--').length - 1, 1);
  });

  test('accepts a pr_url on the reproduced path', () => {
    const { body, errors } = preparePublishedBody(
      draft({ extra: { pr_url: 'https://github.com/elastic/terraform-provider-ec/pull/7' } }),
    );
    assert.deepEqual(errors, []);
    assert.ok(body);
  });
});

// ── the handler (fake octokit + temp artifact) ────────────────

function mockCore() {
  const outputs = {};
  const failures = [];
  const warnings = [];
  return {
    outputs,
    failures,
    warnings,
    core: {
      setOutput(k, v) {
        outputs[k] = v;
      },
      setFailed(m) {
        failures.push(m);
      },
      warning(m) {
        warnings.push(m);
      },
      info() {},
    },
  };
}

function mockGithub({ comments = [], isPullRequest = false } = {}) {
  const calls = { created: [], updated: [], got: [] };
  return {
    calls,
    github: {
      rest: {
        issues: {
          listComments: 'listComments',
          async get({ issue_number }) {
            calls.got.push(issue_number);
            return { data: { pull_request: isPullRequest ? { url: 'x' } : undefined } };
          },
          async createComment({ issue_number, body }) {
            calls.created.push({ issue_number, body });
            return { data: { id: 999 } };
          },
          async updateComment({ comment_id, body }) {
            calls.updated.push({ comment_id, body });
            return { data: { id: comment_id } };
          },
        },
      },
      async paginate(route, params) {
        assert.equal(route, 'listComments');
        void params;
        return comments;
      },
    },
  };
}

const context = { repo: { owner: 'elastic', repo: 'terraform-provider-ec' } };

/** Writes a draft into a temp artifact tree and returns the env that points the handler at it. */
function stageDraft(content) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'repro-handler-'));
  const reproDir = path.join(dir, 'agent', 'repro');
  fs.mkdirSync(reproDir, { recursive: true });
  fs.writeFileSync(path.join(reproDir, 'draft-final.md'), content);
  return { GH_AW_AGENT_OUTPUT: path.join(dir, 'agent-output.json') };
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

const cleanEnv = {
  DETECTION_RESULT: 'success',
  GH_AW_DETECTION_CONCLUSION: 'success',
  INPUT_ISSUE_NUMBER: '42',
  GH_AW_BOT_LOGIN: 'github-actions[bot]',
};

describe('update-reproducer-comment handler', { concurrency: false }, () => {
  test('creates a sticky comment on the happy path with no prior comment', async () => {
    const { core, outputs, failures } = mockCore();
    const { github, calls } = mockGithub();
    const env = { ...cleanEnv, ...stageDraft(draft()) };

    await withEnv(env, () => updateReproducerComment({ github, context, core }));

    assert.deepEqual(failures, []);
    assert.equal(calls.created.length, 1);
    assert.equal(calls.updated.length, 0);
    assert.ok(calls.created[0].body.startsWith(REPRODUCER_COMMENT_MARKER));
    assert.equal(outputs.reproducer_comment_published, 'true');
  });

  test('updates the existing marker comment instead of creating a second', async () => {
    const { core, failures } = mockCore();
    const { github, calls } = mockGithub({
      comments: [{ id: 7, user: { login: 'github-actions[bot]' }, body: `${REPRODUCER_COMMENT_MARKER}\nold` }],
    });
    const env = { ...cleanEnv, ...stageDraft(draft()) };

    await withEnv(env, () => updateReproducerComment({ github, context, core }));

    assert.deepEqual(failures, []);
    assert.equal(calls.created.length, 0);
    assert.equal(calls.updated.length, 1);
    assert.equal(calls.updated[0].comment_id, 7);
  });

  test('ignores a marker comment authored by a human and creates a fresh one', async () => {
    const { core, failures } = mockCore();
    const { github, calls } = mockGithub({
      comments: [{ id: 7, user: { login: 'attacker' }, body: `${REPRODUCER_COMMENT_MARKER}\nnot ours` }],
    });
    const env = { ...cleanEnv, ...stageDraft(draft()) };

    await withEnv(env, () => updateReproducerComment({ github, context, core }));

    assert.deepEqual(failures, []);
    assert.equal(calls.created.length, 1);
    assert.equal(calls.updated.length, 0);
  });

  test('fails closed when detection did not succeed, posting nothing', async () => {
    const { core, failures } = mockCore();
    const { github, calls } = mockGithub();
    const env = { ...cleanEnv, DETECTION_RESULT: 'failure', ...stageDraft(draft()) };

    await withEnv(env, () => updateReproducerComment({ github, context, core }));

    assert.equal(failures.length, 1);
    assert.match(failures[0], /refusing to publish: detection/);
    assert.equal(calls.created.length + calls.updated.length, 0);
  });

  test('fails closed when the detection conclusion is a warning, posting nothing', async () => {
    const { core, failures } = mockCore();
    const { github, calls } = mockGithub();
    const env = { ...cleanEnv, GH_AW_DETECTION_CONCLUSION: 'warning', ...stageDraft(draft()) };

    await withEnv(env, () => updateReproducerComment({ github, context, core }));

    assert.equal(failures.length, 1);
    assert.equal(calls.created.length + calls.updated.length, 0);
  });

  test('fails closed on an out-of-enum outcome, posting nothing', async () => {
    const { core, failures } = mockCore();
    const { github, calls } = mockGithub();
    const env = { ...cleanEnv, ...stageDraft(draft({ outcome: 'reproduced' })) };

    await withEnv(env, () => updateReproducerComment({ github, context, core }));

    assert.equal(failures.length, 1);
    assert.match(failures[0], /outcome: must be one of/);
    assert.equal(calls.created.length + calls.updated.length, 0);
  });

  test('fails closed when the target is a pull request, posting nothing', async () => {
    const { core, failures } = mockCore();
    const { github, calls } = mockGithub({ isPullRequest: true });
    const env = { ...cleanEnv, ...stageDraft(draft()) };

    await withEnv(env, () => updateReproducerComment({ github, context, core }));

    assert.equal(failures.length, 1);
    assert.match(failures[0], /is a pull request/);
    assert.equal(calls.created.length + calls.updated.length, 0);
  });

  test('fails closed on an invalid issue number, posting nothing', async () => {
    const { core, failures } = mockCore();
    const { github, calls } = mockGithub();
    const env = { ...cleanEnv, INPUT_ISSUE_NUMBER: '1.5', ...stageDraft(draft()) };

    await withEnv(env, () => updateReproducerComment({ github, context, core }));

    assert.equal(failures.length, 1);
    assert.equal(calls.created.length + calls.updated.length, 0);
  });

  test('fails with an actionable message when the draft artifact is missing', async () => {
    const { core, failures } = mockCore();
    const { github, calls } = mockGithub();
    const emptyDir = fs.mkdtempSync(path.join(os.tmpdir(), 'repro-empty-'));
    const env = { ...cleanEnv, GH_AW_AGENT_OUTPUT: path.join(emptyDir, 'agent-output.json') };

    await withEnv(env, () => updateReproducerComment({ github, context, core }));

    assert.equal(failures.length, 1);
    assert.match(failures[0], /no draft-final\.md/);
    assert.equal(calls.created.length + calls.updated.length, 0);
  });
});
