/**
 * Which acceptance tests a Buildkite build runs, tested without running any.
 *
 * `.buildkite/acceptance-test-selection.sh` narrows a reproducer-factory PR to its one repro test.
 * The Buildkite context is a required check on master, so the dangerous direction is narrowing a
 * build that should have run the full suite: a fix pushed onto the repro branch would then merge
 * having run a single test. Most cases below are therefore "prints nothing" cases.
 *
 * Each case builds a throwaway repository with a local bare `origin`, sources the script in bash and
 * reads what `acceptance_test_name` prints. It uses local git only: no network, no Go, no Terraform,
 * and nothing here executes an acceptance test.
 */
import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import { execFileSync } from 'node:child_process';
import { mkdtempSync, mkdirSync, readFileSync, rmSync, symlinkSync, writeFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import os from 'node:os';
import path from 'node:path';

const require = createRequire(import.meta.url);
const { ISSUE_BRANCH_PREFIX } = require('./intake/reproducer-factory-constants.js');

const here = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(here, '../../../..');
const SCRIPT = path.join(repoRoot, '.buildkite/acceptance-test-selection.sh');
const ACCEPTANCE = path.join(repoRoot, '.buildkite/acceptance.sh');
const PROMPT = path.join(repoRoot, '.github/workflows/reproducer-factory-issue.md');

const BOT = 'github-actions[bot]';
const HUMAN = 'A Maintainer';

function git(cwd, args, author = HUMAN, extraEnv = {}) {
  return execFileSync('git', args, {
    cwd,
    encoding: 'utf8',
    stdio: ['ignore', 'pipe', 'pipe'],
    env: {
      ...process.env,
      ...extraEnv,
      GIT_AUTHOR_NAME: author,
      GIT_AUTHOR_EMAIL: 'author@example.invalid',
      GIT_COMMITTER_NAME: 'committer',
      GIT_COMMITTER_EMAIL: 'committer@example.invalid',
      GIT_CONFIG_GLOBAL: '/dev/null',
      GIT_CONFIG_SYSTEM: '/dev/null',
    },
  });
}

function reproTest(issue, name = `TestAccReproduceIssue${issue}`) {
  return `package acc\n\nimport "testing"\n\nfunc ${name}(t *testing.T) {\n\tt.Fatal("bug")\n}\n`;
}

/**
 * A repository whose `origin/master` has one commit, checked out on `branch`.
 * `commits` are applied in order; each is `{ files: { path: content }, author }`; a `null` content
 * deletes the file. `baseFiles` are extra files on `origin/master`.
 */
function withRepo({ branch, commits, baseFiles = {} }, fn) {
  const root = mkdtempSync(path.join(os.tmpdir(), 'acc-selection-'));
  try {
    const origin = path.join(root, 'origin.git');
    const work = path.join(root, 'work');
    git(root, ['init', '--quiet', '--bare', '--initial-branch=master', origin]);
    git(root, ['init', '--quiet', '--initial-branch=master', work]);
    git(work, ['remote', 'add', 'origin', origin]);
    mkdirSync(path.join(work, 'ec/acc'), { recursive: true });
    writeFileSync(path.join(work, 'ec/acc/existing_test.go'), 'package acc\n');
    for (const [file, content] of Object.entries(baseFiles)) {
      mkdirSync(path.dirname(path.join(work, file)), { recursive: true });
      writeFileSync(path.join(work, file), content);
    }
    git(work, ['add', '-A']);
    git(work, ['commit', '--quiet', '-m', 'base']);
    git(work, ['push', '--quiet', 'origin', 'master']);
    git(work, ['checkout', '--quiet', '-b', branch]);
    for (const { files, author } of commits) {
      for (const [file, content] of Object.entries(files)) {
        if (content === null) {
          rmSync(path.join(work, file));
          continue;
        }
        mkdirSync(path.dirname(path.join(work, file)), { recursive: true });
        writeFileSync(path.join(work, file), content);
      }
      git(work, ['add', '-A']);
      git(work, ['commit', '--quiet', '-m', 'change'], author);
    }
    return fn(work);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
}

/** Sources the script the way acceptance.sh does (errexit, nounset, pipefail) and prints the name. */
function select(work, env) {
  const buildkite = {
    BUILDKITE_PULL_REQUEST: '123',
    BUILDKITE_PULL_REQUEST_BASE_BRANCH: 'master',
    ...env,
  };
  for (const key of Object.keys(buildkite)) {
    if (buildkite[key] === undefined) delete buildkite[key];
  }
  return execFileSync(
    'bash',
    ['-c', 'set -euo pipefail; source "$1"; name="$(acceptance_test_name)"; printf "%s" "$name"', 'bash', SCRIPT],
    {
      cwd: work,
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'pipe'],
      env: {
        PATH: process.env.PATH,
        HOME: process.env.HOME ?? '',
        GIT_CONFIG_GLOBAL: '/dev/null',
        GIT_CONFIG_SYSTEM: '/dev/null',
        ...buildkite,
      },
    },
  );
}

/** Runs docs_only_pr in `work` the way acceptance.sh does; true when it returns 0 (suite skipped). */
function runDocsOnly(work, env = {}) {
  const buildkite = { BUILDKITE_PULL_REQUEST: '123', BUILDKITE_PULL_REQUEST_BASE_BRANCH: 'master', ...env };
  for (const key of Object.keys(buildkite)) {
    if (buildkite[key] === undefined) delete buildkite[key];
  }
  try {
    execFileSync(
      'bash',
      ['-c', 'set -euo pipefail; source "$1"; if docs_only_pr; then exit 0; else exit 3; fi', 'bash', SCRIPT],
      {
        cwd: work,
        stdio: 'ignore',
        env: {
          PATH: process.env.PATH,
          HOME: process.env.HOME ?? '',
          GIT_CONFIG_GLOBAL: '/dev/null',
          GIT_CONFIG_SYSTEM: '/dev/null',
          ...buildkite,
        },
      },
    );
    return true;
  } catch (error) {
    assert.equal(error.status, 3, `docs_only_pr did not return cleanly: ${error.message}`);
    return false;
  }
}

function docsOnly({ commits, baseFiles, env = {} }) {
  return withRepo({ branch: 'some-pr', commits, baseFiles }, (work) => runDocsOnly(work, env));
}

const branchFor = (issue) => `${ISSUE_BRANCH_PREFIX}${issue}`;
const fileFor = (issue) => `ec/acc/issue_${issue}_acc_test.go`;
const factoryCommit = (issue) => ({ files: { [fileFor(issue)]: reproTest(issue) }, author: BOT });

function selectFor({ branch, commits, baseFiles, env = {} }) {
  return withRepo({ branch, commits, baseFiles }, (work) => select(work, { BUILDKITE_BRANCH: branch, ...env }));
}

/** Runs narrowed_run_passed_test on a captured `go test -v` output; true when it returns 0. */
function ranAndPassed(pattern, output) {
  const dir = mkdtempSync(path.join(os.tmpdir(), 'acc-selection-log-'));
  try {
    const log = path.join(dir, 'test.log');
    writeFileSync(log, output);
    try {
      execFileSync(
        'bash',
        ['-c', 'set -euo pipefail; source "$1"; narrowed_run_passed_test "$2" "$3"', 'bash', SCRIPT, pattern, log],
        { stdio: 'ignore', env: { PATH: process.env.PATH } },
      );
      return true;
    } catch {
      return false;
    }
  } finally {
    rmSync(dir, { recursive: true, force: true });
  }
}

describe('acceptance test selection', () => {
  // ── the one case that narrows ───────────────────────────────

  test('a PR exactly as the factory opened it runs only its repro test', () => {
    assert.equal(
      selectFor({ branch: branchFor(42), commits: [factoryCommit(42)] }),
      '^TestAccReproduceIssue42$',
    );
  });

  test('the branch is built from ISSUE_BRANCH_PREFIX, so the two cannot drift apart', () => {
    // If the prefix in reproducer-factory-constants.js changes and the script does not follow, the
    // case above stops narrowing and fails. This states the dependency out loud.
    assert.equal(ISSUE_BRANCH_PREFIX, 'reproducer-factory/issue-');
    assert.match(readFileSync(SCRIPT, 'utf8'), /\^reproducer-factory\/issue-\(\[1-9\]\[0-9\]\*\)\$/);
  });

  test('several factory-authored commits to the one file still narrow', () => {
    assert.equal(
      selectFor({
        branch: branchFor(7),
        commits: [
          { files: { [fileFor(7)]: '// draft\n' + reproTest(7) }, author: BOT },
          factoryCommit(7),
        ],
      }),
      '^TestAccReproduceIssue7$',
    );
  });

  // ── a fix on the branch: back to the full suite ─────────────

  test('a human commit that adds a fix falls back to the full suite', () => {
    assert.equal(
      selectFor({
        branch: branchFor(42),
        commits: [factoryCommit(42), { files: { 'ec/provider.go': 'package ec\n' }, author: HUMAN }],
      }),
      '',
    );
  });

  test('another changed file falls back even when the bot authored it', () => {
    // The diff check carries the safety; the author check is only a tripwire. A second file means
    // the PR is no longer "one repro test", whoever wrote it.
    assert.equal(
      selectFor({
        branch: branchFor(42),
        commits: [factoryCommit(42), { files: { 'ec/provider.go': 'package ec\n' }, author: BOT }],
      }),
      '',
    );
  });

  test('a human commit that only edits the repro test falls back', () => {
    assert.equal(
      selectFor({
        branch: branchFor(42),
        commits: [factoryCommit(42), { files: { [fileFor(42)]: reproTest(42) + '// tweak\n' }, author: HUMAN }],
      }),
      '',
    );
  });

  test('a hand-made branch with the factory name and a human author falls back', () => {
    assert.equal(
      selectFor({
        branch: branchFor(42),
        commits: [{ files: { [fileFor(42)]: reproTest(42) }, author: HUMAN }],
      }),
      '',
    );
  });

  // ── the two sources of builds name the branch differently ───

  /** The variables Elastic's PR bot sets on a build it creates for a same-repository branch. */
  const botBuild = (branch, extra = {}) => ({
    BUILDKITE_SOURCE: 'api',
    BUILDKITE_BRANCH: `elastic:${branch}`,
    GITHUB_PR_BRANCH: branch,
    GITHUB_PR_OWNER: 'elastic',
    GITHUB_PR_BASE_OWNER: 'elastic',
    GITHUB_PR_TARGET_BRANCH: 'master',
    GITHUB_PR_USER: BOT,
    GITHUB_PR_TRIGGER_USER: 'a-maintainer',
    ...extra,
  });

  const selectWith = (env) =>
    withRepo({ branch: branchFor(42), commits: [factoryCommit(42)] }, (work) => select(work, env));

  test('a build the PR bot created for the repro branch narrows', () => {
    // The `buildkite test this` path: the bot names the branch `elastic:<branch>`.
    assert.equal(selectWith(botBuild(branchFor(42))), '^TestAccReproduceIssue42$');
  });

  test('a fork branch with the factory name runs the full suite', () => {
    assert.equal(
      selectWith(botBuild(branchFor(42), { BUILDKITE_BRANCH: `someone:${branchFor(42)}`, GITHUB_PR_OWNER: 'someone' })),
      '',
    );
  });

  test('PR-bot variables that disagree with each other run the full suite', () => {
    assert.equal(selectWith(botBuild(branchFor(42), { BUILDKITE_BRANCH: 'elastic:other-branch' })), '');
    assert.equal(selectWith(botBuild(branchFor(42), { GITHUB_PR_OWNER: undefined })), '');
    assert.equal(selectWith(botBuild(branchFor(42), { GITHUB_PR_BASE_OWNER: undefined })), '');
    assert.equal(selectWith(botBuild(branchFor(42), { GITHUB_PR_BRANCH: branchFor(43) })), '');
  });

  test('a pull request into a branch other than master is never narrowed', () => {
    // The pipeline does not rebuild when a pull request is retargeted, so a narrowed green earned
    // against another base would survive the move to master.
    assert.equal(selectWith(botBuild(branchFor(42), { GITHUB_PR_TARGET_BRANCH: 'release' })), '');
    assert.equal(
      withRepo({ branch: branchFor(42), commits: [factoryCommit(42)] }, (work) => {
        git(work, ['push', '--quiet', 'origin', 'master:release']);
        return select(work, { BUILDKITE_BRANCH: branchFor(42), BUILDKITE_PULL_REQUEST_BASE_BRANCH: 'release' });
      }),
      '',
    );
    assert.equal(selectWith({ BUILDKITE_BRANCH: branchFor(42), BUILDKITE_PULL_REQUEST_BASE_BRANCH: undefined }), '');
  });

  // ── nothing runs on a repro branch unless a person asked ────

  /** True when reproducer_factory_build_refused returns 0 for this environment. */
  function refused(env) {
    try {
      execFileSync(
        'bash',
        ['-c', 'set -euo pipefail; source "$1"; if reproducer_factory_build_refused; then exit 0; else exit 3; fi', 'bash', SCRIPT],
        { stdio: 'ignore', env: { PATH: process.env.PATH, ...Object.fromEntries(Object.entries(env).filter(([, v]) => v !== undefined)) } },
      );
      return true;
    } catch (error) {
      assert.equal(error.status, 3, `reproducer_factory_build_refused did not return cleanly: ${error.message}`);
      return false;
    }
  }

  test('the webhook build of a repro pull request is refused', () => {
    // What Buildkite starts by itself two seconds after the factory opens the pull request.
    assert.equal(refused({ BUILDKITE_SOURCE: 'webhook', BUILDKITE_BRANCH: branchFor(42), BUILDKITE_PULL_REQUEST: '7' }), true);
    assert.equal(refused({ BUILDKITE_SOURCE: 'webhook', BUILDKITE_BRANCH: branchFor(42), BUILDKITE_PULL_REQUEST: 'false' }), true);
  });

  test('a build a maintainer asked the PR bot for is not refused', () => {
    assert.equal(refused(botBuild(branchFor(42))), false);
    assert.equal(refused(botBuild(branchFor(42), { BUILDKITE_SOURCE: 'ui' })), false);
  });

  test('a repro branch build with no person behind it is refused, whatever its source', () => {
    for (const env of [
      botBuild(branchFor(42), { GITHUB_PR_TRIGGER_USER: BOT }),
      botBuild(branchFor(42), { GITHUB_PR_TRIGGER_USER: 'renovate[bot]' }),
      botBuild(branchFor(42), { GITHUB_PR_TRIGGER_USER: '' }),
      botBuild(branchFor(42), { GITHUB_PR_TRIGGER_USER: undefined }),
      botBuild(branchFor(42), { BUILDKITE_SOURCE: 'webhook' }),
      botBuild(branchFor(42), { BUILDKITE_SOURCE: 'schedule' }),
      botBuild(branchFor(42), { BUILDKITE_SOURCE: 'trigger_job' }),
      botBuild(branchFor(42), { BUILDKITE_SOURCE: undefined }),
      // A person starting a build from the Buildkite UI on the plain branch: no PR-bot variables.
      { BUILDKITE_SOURCE: 'ui', BUILDKITE_BRANCH: branchFor(42) },
      { BUILDKITE_SOURCE: 'api', BUILDKITE_BRANCH: branchFor(42) },
    ]) {
      assert.equal(refused(env), true, JSON.stringify(env));
    }
  });

  test('the repro branch is recognised broadly when deciding to refuse', () => {
    // Refusing errs the other way from narrowing: any owner, any suffix, either variable.
    for (const env of [
      { BUILDKITE_SOURCE: 'webhook', BUILDKITE_BRANCH: 'reproducer-factory/issue-42-fix' },
      { BUILDKITE_SOURCE: 'webhook', BUILDKITE_BRANCH: 'reproducer-factory/anything' },
      { BUILDKITE_SOURCE: 'webhook', BUILDKITE_BRANCH: `someone:${branchFor(42)}` },
      { BUILDKITE_SOURCE: 'webhook', BUILDKITE_BRANCH: 'elastic:other', GITHUB_PR_BRANCH: branchFor(42) },
    ]) {
      assert.equal(refused(env), true, JSON.stringify(env));
    }
  });

  test('builds on other branches are never refused', () => {
    for (const env of [
      { BUILDKITE_SOURCE: 'webhook', BUILDKITE_BRANCH: 'master' },
      { BUILDKITE_SOURCE: 'schedule', BUILDKITE_BRANCH: 'master' },
      { BUILDKITE_SOURCE: 'webhook', BUILDKITE_BRANCH: 'change-factory/issue-42' },
      { BUILDKITE_SOURCE: 'webhook', BUILDKITE_BRANCH: 'code-factory/issue-42' },
      { BUILDKITE_SOURCE: 'webhook', BUILDKITE_BRANCH: 'feature/reproducer-factory/issue-42' },
      { BUILDKITE_SOURCE: 'webhook', BUILDKITE_BRANCH: 'reproducer-factory' },
      { BUILDKITE_SOURCE: 'webhook' },
      botBuild('renovate/actions-checkout-7.x', { GITHUB_PR_TRIGGER_USER: 'renovate[bot]' }),
    ]) {
      assert.equal(refused(env), false, JSON.stringify(env));
    }
  });

  // ── bringing the pull request up to date ────────────────────

  /** The factory PR, then `master` moves on, then `who` merges master into the branch. */
  function afterMerge({ who, alsoInMerge = {}, commits = [factoryCommit(42)], masterFiles = { 'ec/other.go': 'package ec\n' } }) {
    const branch = branchFor(42);
    return withRepo({ branch, commits }, (work) => {
      git(work, ['checkout', '--quiet', 'master']);
      for (const [file, content] of Object.entries(masterFiles)) {
        mkdirSync(path.dirname(path.join(work, file)), { recursive: true });
        writeFileSync(path.join(work, file), content);
      }
      git(work, ['add', '-A']);
      git(work, ['commit', '--quiet', '-m', 'master moves']);
      git(work, ['push', '--quiet', 'origin', 'master']);
      git(work, ['checkout', '--quiet', branch]);
      git(work, ['merge', '--quiet', '--no-ff', '--no-commit', 'master'], who);
      for (const [file, content] of Object.entries(alsoInMerge)) {
        mkdirSync(path.dirname(path.join(work, file)), { recursive: true });
        writeFileSync(path.join(work, file), content);
      }
      git(work, ['add', '-A']);
      git(work, ['commit', '--quiet', '-m', 'Merge branch master'], who);
      return select(work, { BUILDKITE_BRANCH: branch });
    });
  }

  test('a maintainer merging the base branch in ("Update branch") keeps the one-test run', () => {
    assert.equal(afterMerge({ who: HUMAN }), '^TestAccReproduceIssue42$');
  });

  test('a merge that also changes another file falls back', () => {
    assert.equal(afterMerge({ who: HUMAN, alsoInMerge: { 'ec/provider.go': 'package ec\n' } }), '');
    assert.equal(afterMerge({ who: BOT, alsoInMerge: { 'ec/provider.go': 'package ec\n' } }), '');
  });

  test('a merge that removes the test declaration falls back', () => {
    assert.equal(afterMerge({ who: HUMAN, alsoInMerge: { [fileFor(42)]: 'package acc\n' } }), '');
  });

  test('a human commit next to a merge still falls back', () => {
    const branch = branchFor(42);
    assert.equal(
      withRepo({ branch, commits: [factoryCommit(42)] }, (work) => {
        git(work, ['checkout', '--quiet', 'master']);
        writeFileSync(path.join(work, 'other.txt'), 'x\n');
        git(work, ['add', '-A']);
        git(work, ['commit', '--quiet', '-m', 'master moves']);
        git(work, ['push', '--quiet', 'origin', 'master']);
        git(work, ['checkout', '--quiet', branch]);
        git(work, ['merge', '--quiet', '--no-ff', '-m', 'Merge branch master', 'master'], HUMAN);
        writeFileSync(path.join(work, fileFor(42)), reproTest(42) + '// tweak\n');
        git(work, ['add', '-A']);
        git(work, ['commit', '--quiet', '-m', 'tweak'], HUMAN);
        return select(work, { BUILDKITE_BRANCH: branch });
      }),
      '',
    );
  });

  test('a merge commit that itself adds the repro test falls back', () => {
    // No commit on the branch other than the merge: nothing on it was authored by the factory.
    assert.equal(
      afterMerge({ who: BOT, commits: [], alsoInMerge: { [fileFor(42)]: reproTest(42) } }),
      '',
    );
  });

  test('a merge that edits the repro test falls back', () => {
    // The diff is still the one added file, and merge commits are exempt from the author check, so
    // without the per-merge check this would be a way to change the test under a human name.
    assert.equal(
      afterMerge({ who: HUMAN, alsoInMerge: { [fileFor(42)]: reproTest(42) + '\nfunc init() {}\n' } }),
      '',
    );
  });

  /** Replaces HEAD with a hand-made commit that has `parents` and the current index as its tree. */
  function commitWithParents(work, parents, author) {
    git(work, ['add', '-A']);
    const tree = git(work, ['write-tree']).trim();
    const args = ['commit-tree', tree, '-m', 'hand-made merge'];
    for (const parent of parents) args.push('-p', parent);
    const sha = git(work, args, author).trim();
    git(work, ['reset', '--quiet', '--hard', sha]);
  }

  test('an edit to the repro test dressed up as a merge commit falls back', () => {
    // The base has not moved: the "merge" has the branch head and the base as parents and exists
    // only to carry a human edit past the author check.
    const branch = branchFor(42);
    assert.equal(
      withRepo({ branch, commits: [factoryCommit(42)] }, (work) => {
        writeFileSync(path.join(work, fileFor(42)), reproTest(42) + '// tweak\n');
        commitWithParents(work, ['HEAD', 'origin/master'], HUMAN);
        return select(work, { BUILDKITE_BRANCH: branch });
      }),
      '',
    );
  });

  /** Commits `files` on the current branch of `work` as `author`. */
  function commit(work, files, author, message = 'change') {
    for (const [file, content] of Object.entries(files)) {
      mkdirSync(path.dirname(path.join(work, file)), { recursive: true });
      writeFileSync(path.join(work, file), content);
    }
    git(work, ['add', '-A']);
    git(work, ['commit', '--quiet', '--allow-empty', '-m', message], author);
  }

  test('a pull request far behind the base keeps the one-test run after "Update branch"', () => {
    // More than 50 commits behind: a fetch that cut the clone to 50 commits would disconnect the
    // branch from the base and send this to the full suite.
    const branch = branchFor(42);
    assert.equal(
      withRepo({ branch, commits: [factoryCommit(42)] }, (work) => {
        git(work, ['checkout', '--quiet', 'master']);
        for (let i = 0; i < 60; i++) commit(work, { 'ec/other.go': `package ec // ${i}\n` }, HUMAN);
        git(work, ['push', '--quiet', 'origin', 'master']);
        git(work, ['checkout', '--quiet', branch]);
        git(work, ['merge', '--quiet', '--no-ff', '-m', 'Merge branch master', 'master'], HUMAN);
        return select(work, { BUILDKITE_BRANCH: branch });
      }),
      '^TestAccReproduceIssue42$',
    );
  });

  test('merging a side branch falls back, even when its commits carry the bot name and net to nothing', () => {
    const branch = branchFor(42);
    assert.equal(
      withRepo({ branch, commits: [factoryCommit(42)] }, (work) => {
        git(work, ['checkout', '--quiet', '-b', 'side']);
        commit(work, { 'ec/side.go': 'package ec\n' }, BOT);
        git(work, ['rm', '--quiet', 'ec/side.go']);
        git(work, ['commit', '--quiet', '-m', 'undo'], BOT);
        git(work, ['checkout', '--quiet', branch]);
        git(work, ['merge', '--quiet', '--no-ff', '-m', 'Merge branch side', 'side'], HUMAN);
        return select(work, { BUILDKITE_BRANCH: branch });
      }),
      '',
    );
  });

  test('an octopus merge falls back', () => {
    const branch = branchFor(42);
    assert.equal(
      withRepo({ branch, commits: [factoryCommit(42)] }, (work) => {
        git(work, ['checkout', '--quiet', 'master']);
        commit(work, { 'ec/a.go': 'package ec\n' }, HUMAN);
        git(work, ['branch', 'older']);
        commit(work, { 'ec/b.go': 'package ec\n' }, HUMAN);
        git(work, ['push', '--quiet', 'origin', 'master']);
        git(work, ['checkout', '--quiet', branch]);
        // git folds a real octopus of two commits on one line of history into a plain merge, so
        // build the three-parent commit by hand: the tree of a merge of master, three parents.
        git(work, ['merge', '--quiet', '--no-ff', '--no-commit', 'master'], HUMAN);
        const head = git(work, ['rev-parse', 'HEAD']).trim();
        git(work, ['add', '-A']);
        const tree = git(work, ['write-tree']).trim();
        const sha = git(work, ['commit-tree', tree, '-m', 'octopus', '-p', head, '-p', 'older', '-p', 'master'], HUMAN).trim();
        git(work, ['merge', '--abort']);
        git(work, ['reset', '--quiet', '--hard', sha]);
        assert.equal(git(work, ['rev-list', '--no-walk', '--parents', 'HEAD']).trim().split(' ').length, 4);
        return select(work, { BUILDKITE_BRANCH: branch });
      }),
      '',
    );
  });

  /**
   * A base branch with two real merges (p, then q), and a pull request branch that starts at
   * `start`, merges the other side with `-s ours` (so it silently drops that side's change), and
   * then gets `prCommit`. The pull request has two merge bases. Returns what `fn` returns.
   * p and q get distinct dates so the order of the merge bases does not depend on commit hashes.
   */
  function withTwoMergeBases({ start, branch, prCommit }, fn) {
    const other = start === 'p' ? 'q' : 'p';
    const at = (iso) => ({ GIT_AUTHOR_DATE: iso, GIT_COMMITTER_DATE: iso });
    return withRepo({ branch: 'scratch', commits: [] }, (work) => {
      for (const [name, date] of [['p', '2020-01-01T00:00:00Z'], ['q', '2020-02-01T00:00:00Z']]) {
        git(work, ['checkout', '--quiet', '-b', name, 'master']);
        writeFileSync(path.join(work, `ec/${name}.go`), 'package ec // new\n');
        git(work, ['add', '-A']);
        git(work, ['commit', '--quiet', '-m', name], HUMAN, at(date));
      }
      git(work, ['checkout', '--quiet', 'master']);
      git(work, ['merge', '--quiet', '--no-ff', '-m', 'Merge p', 'p'], HUMAN, at('2020-03-01T00:00:00Z'));
      git(work, ['merge', '--quiet', '--no-ff', '-m', 'Merge q', 'q'], HUMAN, at('2020-03-02T00:00:00Z'));
      git(work, ['push', '--quiet', 'origin', 'master']);
      git(work, ['checkout', '--quiet', '-b', branch, start]);
      git(work, ['merge', '--quiet', '-s', 'ours', '-m', 'Merge, keeping ours', other], HUMAN, at('2020-03-03T00:00:00Z'));
      commit(work, prCommit.files, prCommit.author);
      assert.equal(git(work, ['merge-base', '--all', 'origin/master', 'HEAD']).trim().split('\n').length, 2);
      // Merging this pull request would really change two files, whatever the three-dot diff says.
      assert.equal(git(work, ['diff', '--name-only', 'origin/master', 'HEAD']).trim().split('\n').length, 2);
      return fn(work);
    });
  }

  test('two merge bases fall back', () => {
    // The three-dot diff is taken against only one of the two merge bases, and for one orientation
    // it shows the repro test alone while merging the pull request would also revert a provider
    // file. Both orientations must fall back, and the setup must really produce the hiding diff for
    // at least one of them, or this test would pass without the guard.
    const branch = branchFor(42);
    let hidden = 0;
    for (const start of ['p', 'q']) {
      const selected = withTwoMergeBases({ start, branch, prCommit: factoryCommit(42) }, (work) => {
        const threeDot = git(work, ['diff', '--no-renames', '--name-status', 'origin/master...HEAD']).trim();
        if (threeDot === `A\t${fileFor(42)}`) hidden += 1;
        return select(work, { BUILDKITE_BRANCH: branch });
      });
      assert.equal(selected, '', `branch starting at ${start}`);
    }
    assert.ok(hidden >= 1, 'neither orientation produced the diff that hides the revert');
  });

  test('a text-only pull request with two merge bases runs the suite', () => {
    // Same shape for the docs-only skip: the three-dot diff shows docs/a.md alone while merging the
    // pull request would also revert a provider file.
    let hidden = 0;
    for (const start of ['p', 'q']) {
      const skipped = withTwoMergeBases(
        { start, branch: 'some-pr', prCommit: { files: { 'docs/a.md': 'a\n' }, author: HUMAN } },
        (work) => {
          const threeDot = git(work, ['diff', '--no-renames', '--name-only', 'origin/master...HEAD']).trim();
          if (threeDot === 'docs/a.md') hidden += 1;
          return runDocsOnly(work);
        },
      );
      assert.equal(skipped, false, `branch starting at ${start}`);
    }
    assert.ok(hidden >= 1, 'neither orientation produced the diff that hides the revert');
  });

  // ── renames and deletions hidden behind the one path ────────

  const manyTests = Array.from({ length: 40 }, (_, i) => `func TestAccExisting${i}(t *testing.T) {}\n`).join('');
  const bigTestFile = `package acc\n\nimport "testing"\n\n${manyTests}`;

  test('moving an existing test file onto the repro path falls back', () => {
    // With rename detection on, git reports this as the single destination path, and the forty
    // tests that disappear with the old file would never run.
    assert.equal(
      selectFor({
        branch: branchFor(42),
        baseFiles: { 'ec/acc/big_test.go': bigTestFile },
        commits: [{
          files: {
            'ec/acc/big_test.go': null,
            [fileFor(42)]: `${bigTestFile}\nfunc TestAccReproduceIssue42(t *testing.T) {\n\tt.Fatal("bug")\n}\n`,
          },
          author: BOT,
        }],
      }),
      '',
    );
  });

  test('moving a non-Go file onto the repro path falls back', () => {
    const owners = Array.from({ length: 30 }, (_, i) => `/path/${i} @elastic/team-${i}`).join('\n');
    assert.equal(
      selectFor({
        branch: branchFor(42),
        baseFiles: { '.github/CODEOWNERS': `${owners}\n` },
        commits: [{
          files: {
            '.github/CODEOWNERS': null,
            [fileFor(42)]: `package acc\n\nimport "testing"\n\n/*\n${owners}\n*/\n\nfunc TestAccReproduceIssue42(t *testing.T) {\n\tt.Fatal("bug")\n}\n`,
          },
          author: BOT,
        }],
      }),
      '',
    );
  });

  test('deleting another file alongside the repro test falls back', () => {
    assert.equal(
      selectFor({
        branch: branchFor(42),
        commits: [{ files: { 'ec/acc/existing_test.go': null, [fileFor(42)]: reproTest(42) }, author: BOT }],
      }),
      '',
    );
  });

  test('a repro file that already exists on the base falls back', () => {
    // Only an added file narrows: a modification means the base already carries this path, which
    // is not what the factory opens.
    assert.equal(
      selectFor({
        branch: branchFor(42),
        baseFiles: { [fileFor(42)]: 'package acc\n' },
        commits: [factoryCommit(42)],
      }),
      '',
    );
  });

  test('a symlink at the repro path falls back', () => {
    assert.equal(
      withRepo({ branch: branchFor(42), commits: [] }, (work) => {
        const target = path.join(work, '..', 'elsewhere.go');
        writeFileSync(target, reproTest(42));
        symlinkSync(target, path.join(work, fileFor(42)));
        git(work, ['add', '-A']);
        git(work, ['commit', '--quiet', '-m', 'change'], BOT);
        return select(work, { BUILDKITE_BRANCH: branchFor(42) });
      }),
      '',
    );
  });

  // ── a narrowed run that would test nothing ──────────────────

  test('a narrowed run counts only when the selected test ran and passed', () => {
    const pattern = '^TestAccReproduceIssue42$';
    assert.equal(ranAndPassed(pattern, '=== RUN   TestAccReproduceIssue42\n--- PASS: TestAccReproduceIssue42 (12.30s)\nPASS\nok  \tpkg\t12.4s\n'), true);
    // t.Skip: go test exits 0.
    assert.equal(ranAndPassed(pattern, '=== RUN   TestAccReproduceIssue42\n--- SKIP: TestAccReproduceIssue42 (0.00s)\nPASS\nok  \tpkg\t0.1s\n'), false);
    // A build constraint, or a declaration that only exists in a comment: nothing matches, exit 0.
    assert.equal(ranAndPassed(pattern, 'testing: warning: no tests to run\nPASS\nok  \tpkg\t0.1s [no tests to run]\n'), false);
    // Another test passing does not count, nor does a longer name or an indented (sub-test) line.
    assert.equal(ranAndPassed(pattern, '--- PASS: TestAccReproduceIssue421 (1.00s)\n'), false);
    assert.equal(ranAndPassed(pattern, '    --- PASS: TestAccReproduceIssue42 (1.00s)\n'), false);
    assert.equal(ranAndPassed(pattern, ''), false);
    // A skip is a skip even when the output also carries a PASS line for the same name.
    assert.equal(ranAndPassed(pattern, '--- SKIP: TestAccReproduceIssue42 (0.00s)\n--- PASS: TestAccReproduceIssue42 (0.00s)\n'), false);
    // A pattern that is not a plain test name is never accepted.
    assert.equal(ranAndPassed('^.*$', '--- PASS: TestAccReproduceIssue42 (1.00s)\n'), false);
  });

  test('a file that does not declare the expected test falls back', () => {
    // `go test -run` with no match exits 0: narrowing here would paint the required check green.
    assert.equal(
      selectFor({
        branch: branchFor(42),
        commits: [{ files: { [fileFor(42)]: reproTest(42, 'TestAccSomethingElse') }, author: BOT }],
      }),
      '',
    );
  });

  test('a test for a different issue number falls back', () => {
    assert.equal(
      selectFor({
        branch: branchFor(42),
        commits: [{ files: { [fileFor(42)]: reproTest(421) }, author: BOT }],
      }),
      '',
    );
  });

  test('the wrong file name for the branch falls back', () => {
    assert.equal(
      selectFor({ branch: branchFor(42), commits: [factoryCommit(43)] }),
      '',
    );
  });

  // ── branch shapes ───────────────────────────────────────────

  for (const branch of [
    'master',
    'feature/reproducer-factory/issue-42',
    'reproducer-factory/issue-42-fix',
    'reproducer-factory/issue-',
    'reproducer-factory/issue-042',
    'reproducer-factory/issue-4x',
    'someone:reproducer-factory/issue-42',
    'change-factory/issue-42',
  ]) {
    test(`branch ${JSON.stringify(branch)} runs the full suite`, () => {
      assert.equal(
        withRepo({ branch: 'scratch', commits: [factoryCommit(42)] }, (work) =>
          select(work, { BUILDKITE_BRANCH: branch })),
        '',
      );
    });
  }

  test('an unset branch runs the full suite', () => {
    assert.equal(
      withRepo({ branch: 'scratch', commits: [factoryCommit(42)] }, (work) =>
        select(work, { BUILDKITE_BRANCH: undefined })),
      '',
    );
  });

  // ── build shapes ────────────────────────────────────────────

  test('a branch build that is not a pull request runs the full suite', () => {
    assert.equal(
      selectFor({ branch: branchFor(42), commits: [factoryCommit(42)], env: { BUILDKITE_PULL_REQUEST: 'false' } }),
      '',
    );
    assert.equal(
      selectFor({ branch: branchFor(42), commits: [factoryCommit(42)], env: { BUILDKITE_PULL_REQUEST: undefined } }),
      '',
    );
  });

  test('a base branch that cannot be fetched runs the full suite', () => {
    assert.equal(
      selectFor({
        branch: branchFor(42),
        commits: [factoryCommit(42)],
        env: { BUILDKITE_PULL_REQUEST_BASE_BRANCH: 'no-such-branch' },
      }),
      '',
    );
  });

  test('a branch with no commits of its own runs the full suite', () => {
    assert.equal(selectFor({ branch: branchFor(42), commits: [] }), '');
  });

  // ── skipping the suite for text-only pull requests ──────────

  const providerFile = `package ec\n\n${Array.from({ length: 30 }, (_, i) => `func F${i}() {}\n`).join('')}`;

  test('a pull request that only touches text trees skips the suite', () => {
    assert.equal(
      docsOnly({
        commits: [{
          files: {
            'docs/a.md': 'a\n',
            'dev-docs/b.md': 'b\n',
            'openspec/c.md': 'c\n',
            '.changelog/1.txt': 'd\n',
          },
          author: HUMAN,
        }],
      }),
      true,
    );
  });

  test('a pull request that also touches provider code runs the suite', () => {
    assert.equal(
      docsOnly({ commits: [{ files: { 'docs/a.md': 'a\n', 'ec/provider.go': providerFile }, author: HUMAN }] }),
      false,
    );
  });

  test('moving a provider file into docs/ runs the suite', () => {
    // With rename detection on, git lists only docs/provider.go and the suite would be skipped for
    // a pull request that deletes provider code.
    assert.equal(
      docsOnly({
        baseFiles: { 'ec/provider.go': providerFile },
        commits: [{ files: { 'ec/provider.go': null, 'docs/provider.go': providerFile }, author: HUMAN }],
      }),
      false,
    );
  });

  test('deleting a provider file alongside a docs change runs the suite', () => {
    assert.equal(
      docsOnly({
        baseFiles: { 'ec/provider.go': providerFile },
        commits: [{ files: { 'ec/provider.go': null, 'docs/a.md': 'a\n' }, author: HUMAN }],
      }),
      false,
    );
  });

  test('a path that only looks like a text tree runs the suite', () => {
    assert.equal(docsOnly({ commits: [{ files: { 'ec/docs/a.md': 'a\n' }, author: HUMAN }] }), false);
    assert.equal(docsOnly({ commits: [{ files: { 'docs.go': 'package main\n' }, author: HUMAN }] }), false);
  });

  test('a text-only pull request far behind the base still skips the suite', () => {
    assert.equal(
      withRepo({ branch: 'some-pr', commits: [{ files: { 'docs/a.md': 'a\n' }, author: HUMAN }] }, (work) => {
        git(work, ['checkout', '--quiet', 'master']);
        for (let i = 0; i < 60; i++) {
          writeFileSync(path.join(work, 'ec/acc/existing_test.go'), `package acc // ${i}\n`);
          git(work, ['add', '-A']);
          git(work, ['commit', '--quiet', '-m', 'master moves']);
        }
        git(work, ['push', '--quiet', 'origin', 'master']);
        git(work, ['checkout', '--quiet', 'some-pr']);
        return runDocsOnly(work);
      }),
      true,
    );
  });

  test('a text-only pull request into a branch other than master still skips the suite', () => {
    // A backport of documentation must not run the paid suite. Only the narrowing rule is tied to
    // master; this one diffs against whatever base the pull request names.
    const commits = [{ files: { 'docs/a.md': 'a\n' }, author: HUMAN }];
    assert.equal(
      withRepo({ branch: 'some-pr', commits }, (work) => {
        git(work, ['push', '--quiet', 'origin', 'master:release']);
        return runDocsOnly(work, { BUILDKITE_PULL_REQUEST_BASE_BRANCH: 'release' });
      }),
      true,
    );
    // A code change against that base still runs it.
    assert.equal(
      withRepo({ branch: 'some-pr', commits: [{ files: { 'docs/a.md': 'a\n', 'ec/x.go': 'package ec\n' }, author: HUMAN }] }, (work) => {
        git(work, ['push', '--quiet', 'origin', 'master:release']);
        return runDocsOnly(work, { BUILDKITE_PULL_REQUEST_BASE_BRANCH: 'release' });
      }),
      false,
    );
  });

  test('the suite is never skipped when the build is not a pull request, the base is missing, or nothing changed', () => {
    const commits = [{ files: { 'docs/a.md': 'a\n' }, author: HUMAN }];
    assert.equal(docsOnly({ commits, env: { BUILDKITE_PULL_REQUEST: 'false' } }), false);
    assert.equal(docsOnly({ commits, env: { BUILDKITE_PULL_REQUEST: undefined } }), false);
    assert.equal(docsOnly({ commits, env: { BUILDKITE_PULL_REQUEST_BASE_BRANCH: 'no-such-branch' } }), false);
    assert.equal(docsOnly({ commits: [] }), false);
  });

  // ── wiring ──────────────────────────────────────────────────

  test('acceptance.sh refuses an unrequested repro build before anything from the branch is compiled', () => {
    const acceptance = readFileSync(ACCEPTANCE, 'utf8');
    const source = acceptance.indexOf('acceptance-test-selection.sh"');
    const refuse = acceptance.indexOf('if reproducer_factory_build_refused; then');
    assert.ok(source !== -1 && refuse > source);
    // Nothing that compiles or runs Go, and no other rule, comes before it.
    for (const later of ['if docs_only_pr; then', 'make vendor', 'go run', 'make testacc']) {
      const at = acceptance.indexOf(later);
      assert.ok(at > refuse, `${later} comes before the refusal`);
    }
    // It fails the step and keeps the pre-exit sweep, which compiles the same package, from running.
    const block = acceptance.slice(refuse, acceptance.indexOf('\nfi\n', refuse));
    assert.match(block, /touch \.buildkite\/\.skip-acceptance-sweep\n\s*exit 1$/);
    const hook = readFileSync(path.join(repoRoot, '.buildkite/hooks/pre-exit'), 'utf8');
    assert.match(hook, /if \[\[ -f \.buildkite\/\.skip-acceptance-sweep \]\]; then[\s\S]*?exit 0/);
  });

  /** Runs the pre-exit hook in a scratch checkout with a fake `make`; returns what it printed. */
  function runPreExit(env, { marker = false } = {}) {
    const dir = mkdtempSync(path.join(os.tmpdir(), 'acc-pre-exit-'));
    try {
      mkdirSync(path.join(dir, '.buildkite/hooks'), { recursive: true });
      mkdirSync(path.join(dir, 'bin-fake'));
      writeFileSync(path.join(dir, '.buildkite/acceptance-test-selection.sh'), readFileSync(SCRIPT));
      writeFileSync(path.join(dir, '.buildkite/hooks/pre-exit'), readFileSync(path.join(repoRoot, '.buildkite/hooks/pre-exit')));
      writeFileSync(path.join(dir, 'bin-fake/make'), '#!/bin/bash\necho "FAKE-MAKE $*"\n', { mode: 0o755 });
      if (marker) writeFileSync(path.join(dir, '.buildkite/.skip-acceptance-sweep'), '');
      return execFileSync('bash', ['.buildkite/hooks/pre-exit'], {
        cwd: dir,
        encoding: 'utf8',
        env: {
          PATH: `${path.join(dir, 'bin-fake')}:${process.env.PATH}`,
          BUILDKITE_STEP_KEY: 'acceptance-tests',
          TERRAFORM_PROVIDER_API_KEY_SECRET: 'not-a-real-key',
          ...Object.fromEntries(Object.entries(env).filter(([, v]) => v !== undefined)),
        },
      });
    } finally {
      rmSync(dir, { recursive: true, force: true });
    }
  }

  test('the pre-exit sweep is held back on a refused repro build even without the marker file', () => {
    // A job cancelled before acceptance.sh wrote the marker still reaches this hook, and the sweep
    // compiles the agent-written test with the live key.
    const webhook = { BUILDKITE_SOURCE: 'webhook', BUILDKITE_BRANCH: branchFor(42) };
    assert.doesNotMatch(runPreExit(webhook), /FAKE-MAKE/);
    assert.match(runPreExit(webhook), /Skip sweep/);
    // A build a person asked for sweeps as usual, and so does every other branch.
    assert.match(runPreExit(botBuild(branchFor(42))), /FAKE-MAKE sweep/);
    assert.match(runPreExit({ BUILDKITE_SOURCE: 'webhook', BUILDKITE_BRANCH: 'master' }), /FAKE-MAKE sweep/);
    assert.match(runPreExit({ BUILDKITE_SOURCE: 'schedule', BUILDKITE_BRANCH: 'master' }), /FAKE-MAKE sweep/);
    // The marker written by acceptance.sh still works on its own.
    assert.doesNotMatch(runPreExit({ BUILDKITE_SOURCE: 'webhook', BUILDKITE_BRANCH: 'master' }, { marker: true }), /FAKE-MAKE/);
    // Another step's pre-exit does nothing.
    assert.equal(runPreExit({ BUILDKITE_STEP_KEY: 'other', BUILDKITE_BRANCH: branchFor(42) }), '');
  });

  test('acceptance.sh takes the skip rule from the script, before anything paid runs', () => {
    const acceptance = readFileSync(ACCEPTANCE, 'utf8');
    assert.doesNotMatch(acceptance, /docs_only_pr\(\)/, 'acceptance.sh defines its own docs_only_pr again');
    const source = acceptance.indexOf('acceptance-test-selection.sh"');
    const skip = acceptance.indexOf('if docs_only_pr; then');
    const run = acceptance.indexOf('make testacc');
    assert.ok(source !== -1 && skip !== -1 && run !== -1);
    assert.ok(source < skip && skip < run);
    // No rename-detecting diff may decide what runs.
    const script = readFileSync(SCRIPT, 'utf8');
    for (const line of script.split('\n').filter((l) => /^\s*[a-z_]+=\$\(git diff /.test(l))) {
      assert.match(line, /--no-renames/, `rename detection is on in: ${line.trim()}`);
    }
    assert.doesNotMatch(acceptance, /git diff /);
  });

  test('acceptance.sh sources the script and exports the selection as TEST_NAME', () => {
    const acceptance = readFileSync(ACCEPTANCE, 'utf8');
    assert.match(acceptance, /source "\$\(dirname "\$\{BASH_SOURCE\[0\]\}"\)\/acceptance-test-selection\.sh"/);
    assert.match(acceptance, /selected_test_name="\$\(acceptance_test_name\)"/);
    assert.match(acceptance, /export TEST_NAME="\$selected_test_name"/);
    // The narrowed run's output is checked, and a run that did not pass the test fails the step.
    assert.match(
      acceptance,
      /make testacc 2>&1 \| tee "\$test_log"\n\s*if ! narrowed_run_passed_test "\$selected_test_name" "\$test_log"; then[\s\S]*?exit 1\n\s*fi/,
    );
    assert.match(acceptance, /^set -euo pipefail$/m);
    // TEST_NAME must only ever be set from the selection: a stray unconditional export would narrow
    // every build.
    assert.equal(acceptance.match(/TEST_NAME=/g).length, 1);
  });

  test('the prompt prescribes the file and test names the script matches', () => {
    const prompt = readFileSync(PROMPT, 'utf8');
    assert.match(prompt, /`ec\/acc\/issue_\$\{\{ needs\.pre_activation\.outputs\.issue_number \}\}_acc_test\.go`/);
    assert.match(prompt, /`TestAccReproduceIssue\$\{\{ needs\.pre_activation\.outputs\.issue_number \}\}`/);
    assert.match(prompt, /`reproducer-factory\/issue-\$\{\{ needs\.pre_activation\.outputs\.issue_number \}\}`/);
  });
});
