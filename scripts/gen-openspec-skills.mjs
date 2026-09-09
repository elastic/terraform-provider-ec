#!/usr/bin/env node
// ponytail: OpenSpec 1.12 has no project-local workflow pin; `openspec update`
// uses the global core profile (propose/explore/apply/update/sync/archive) and
// would delete new-change/continue-change. This imports dist/ internals of the
// pinned CLI so we emit exactly the seven named skills as `--tools agents`.
// If a later CLI grows `--workflows` or a project config field, delete this
// and call that instead. Do not run `openspec init --tools claude`: `.claude`
// is a symlink to `.agents` and would fight this tree.

import { mkdirSync, writeFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = join(dirname(fileURLToPath(import.meta.url)), '..');
const pkgRoot = join(root, 'node_modules/@fission-ai/openspec');
const require = createRequire(import.meta.url);

const { getSkillTemplates, generateSkillContent } = await import(
  join(pkgRoot, 'dist/core/shared/skill-generation.js')
);
const { getSkillReferenceTransformer } = await import(
  join(pkgRoot, 'dist/utils/command-references.js')
);
const { version } = require(join(pkgRoot, 'package.json'));

const wanted = ['explore', 'propose', 'new', 'continue', 'apply', 'sync', 'archive'];
const transform = getSkillReferenceTransformer('agents');
const skillsRoot = join(root, '.agents/skills');

mkdirSync(skillsRoot, { recursive: true });
writeFileSync(join(skillsRoot, '.openspec-target'), 'agents\n');

for (const { template, dirName } of getSkillTemplates(wanted)) {
  const dir = join(skillsRoot, dirName);
  mkdirSync(dir, { recursive: true });
  writeFileSync(join(dir, 'SKILL.md'), generateSkillContent(template, version, transform));
  console.log(`wrote .agents/skills/${dirName}/SKILL.md`);
}
