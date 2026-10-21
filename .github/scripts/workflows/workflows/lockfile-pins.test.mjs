import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../../..');

test('compiled lock files pin gh-aw-actions to the SHA in actions-lock.json', () => {
  const lock = JSON.parse(
    fs.readFileSync(path.join(repoRoot, '.github/aw/actions-lock.json'), 'utf8'),
  );
  const setup = Object.values(lock.entries).find(
    (entry) => entry.repo === 'github/gh-aw-actions/setup',
  );
  assert.ok(setup?.sha, 'actions-lock.json must pin github/gh-aw-actions/setup');

  const workflowsDir = path.join(repoRoot, '.github/workflows');
  const lockFiles = fs.readdirSync(workflowsDir).filter((name) => name.endsWith('.lock.yml'));
  assert.ok(lockFiles.length > 0, 'expected at least one compiled workflow lock file');

  const usesRe = /uses:\s+github\/gh-aw-actions\/setup@([0-9a-f]{40})/g;
  for (const name of lockFiles) {
    const text = fs.readFileSync(path.join(workflowsDir, name), 'utf8');
    const shas = [...text.matchAll(usesRe)].map((m) => m[1]);
    assert.ok(shas.length > 0, `${name} must reference github/gh-aw-actions/setup`);
    for (const sha of shas) {
      assert.equal(
        sha,
        setup.sha,
        `${name} pins gh-aw-actions/setup@${sha}, expected ${setup.sha} from actions-lock.json`,
      );
    }
  }
});

test('renovate does not bump compiled lock files or gh-aw-actions independently', () => {
  const renovate = JSON.parse(fs.readFileSync(path.join(repoRoot, 'renovate.json'), 'utf8'));
  const rules = renovate.packageRules ?? [];

  const lockRule = rules.find((r) =>
    r.matchFileNames?.includes('.github/workflows/*.lock.yml'),
  );
  assert.equal(lockRule?.enabled, false);

  const pkgRule = rules.find((r) =>
    (r.matchPackageNames ?? []).some(
      (name) => name === 'github/gh-aw-actions' || name === 'github/gh-aw-actions/**',
    ),
  );
  assert.equal(pkgRule?.enabled, false);
});
