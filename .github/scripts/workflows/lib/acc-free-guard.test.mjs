/**
 * AC #2 as a regression test: the *agent* never runs acceptance tests and holds no cloud credential.
 *
 * This scans BOTH compiled artifacts, because they fail differently:
 *
 * - `reproducer-factory-issue.lock.yml` — gh-aw v0.92.0 does NOT inline the markdown prompt into the
 *   lock; it emits a `{{#runtime-import …}}` reference and reads the `.md` at run time. So a lock-only
 *   guard would be blind to the single most likely place a live acceptance run would be reintroduced:
 *   the prompt. The lock is scanned for any acc *token* at all.
 * - `reproducer-factory-issue.md` — the prompt deliberately *names* `TF_ACC`, `TestAcc…` and
 *   `EC_API_KEY` as prohibitions, so a substring guard would trip on the workflow's own safety text.
 *   The `.md` is therefore scanned only for acc *invocation shapes* (an assignment or a run command),
 *   never bare mentions.
 *
 * Authoring constraint — for the workflow author, NOT for the prompt. When writing the `.md`, state
 * prohibitions as bare token names and never as a runnable invocation: no `TF_ACC` followed by `=`, no
 * `make` + `testacc`, no `-run` + `TestAcc`, no cloud-key name followed by `=`. A sentence spelling out
 * the forbidden shapes matches these very regexes and fails this guard. (The prompt's L2 line
 * `env -u TF_ACC go test ./<pkg> -run <Name> -count=1` does NOT match: `TF_ACC` there is an argument to
 * `env -u`, not an assignment, and `-run <Name>` is a placeholder.) An HTML comment is not an escape
 * hatch: this guard greps the raw file.
 *
 * The workflow opens a draft PR carrying an acceptance test the agent *wrote*, so the lock contains
 * `create_pull_request` and the `.md` names `TestAccReproduceIssue<N>`. Neither is caught here, on
 * purpose: a test *name* is not an invocation. What this guard asserts is that the agent job has no acc
 * *invocation* and no cloud credential. Buildkite runs the authored test, out of band, and only after a
 * maintainer asks for it on the PR; `.buildkite/` is deliberately not scanned, because running the
 * suite is that directory's job.
 */
import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const here = path.dirname(fileURLToPath(import.meta.url));
// From .github/scripts/workflows/lib/ up to .github/workflows/.
const workflowsDir = path.resolve(here, '../../../workflows');
const LOCK = path.join(workflowsDir, 'reproducer-factory-issue.lock.yml');
const MD = path.join(workflowsDir, 'reproducer-factory-issue.md');

const lock = readFileSync(LOCK, 'utf8');
const md = readFileSync(MD, 'utf8');

// ── lock: no acc token of any shape ───────────────────────────

const LOCK_FORBIDDEN = [
  { name: 'TF_ACC assignment or key', re: /TF_ACC\s*[:=]/ },
  { name: 'cloud API key name', re: /EC_API_KEY/ },
  { name: 'cloud API host', re: /elastic-cloud\.com|api\.elastic\.cloud/ },
  { name: 'acceptance target', re: /testacc/ },
];

for (const { name, re } of LOCK_FORBIDDEN) {
  test(`lock contains no ${name}`, () => {
    const match = re.exec(lock);
    assert.equal(match, null, match ? `found ${JSON.stringify(match[0])} in the compiled lock` : '');
  });
}

// ── prompt: no acc INVOCATION shape (bare mentions are allowed) ─

const MD_FORBIDDEN = [
  { name: 'TF_ACC assignment', re: /TF_ACC\s*=/ },
  { name: 'make testacc invocation', re: /\bmake\s+testacc\b/ },
  { name: '-run TestAcc invocation', re: /-run\s+['"]?TestAcc/ },
  { name: 'cloud API key assignment', re: /EC_API_KEY\s*=/ },
];

for (const { name, re } of MD_FORBIDDEN) {
  test(`prompt contains no ${name}`, () => {
    const match = re.exec(md);
    assert.equal(match, null, match ? `found ${JSON.stringify(match[0])} in the prompt .md` : '');
  });
}

// ── the guard is actually looking at the right files ──────────

test('both artifacts exist and are non-empty', () => {
  assert.ok(lock.length > 0, `empty or missing lock at ${LOCK}`);
  assert.ok(md.length > 0, `empty or missing prompt at ${MD}`);
});

test('the prompt names the repro test without ever passing it to a test runner', () => {
  // The name has to be there (the Buildkite narrowing matches on it), which is exactly why the
  // invocation regexes above must be able to tell a name from a command.
  assert.match(md, /TestAccReproduceIssue/);
  assert.equal(/go\s+test[^\n]*TestAcc/.test(md), false, 'the prompt puts a TestAcc name on a go test command line');
});

test('the prompt does name the prohibitions as bare tokens (guard is not vacuous)', () => {
  // If this ever fails, the prompt dropped its safety text — the guard above would then pass trivially.
  assert.match(md, /TF_ACC/);
  assert.match(md, /EC_API_KEY/);
});
