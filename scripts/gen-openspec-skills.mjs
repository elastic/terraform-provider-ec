#!/usr/bin/env node
// ponytail: OpenSpec 1.12 has no project-local workflow pin; `openspec init`
// and `openspec update` use the global core profile (propose/explore/apply/
// update/sync/archive) and would delete new-change/continue-change. This imports
// dist/ internals of the pinned CLI so we emit exactly the seven named skills
// as `--tools agents`. If a later CLI grows `--workflows` or a project config
// field, delete this and call that instead. `.claude` is a symlink to `.agents`;
// never run `openspec init`/`update` in this repo.

import { mkdirSync, readdirSync, rmSync, writeFileSync } from 'node:fs';
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
const templates = getSkillTemplates(wanted);
if (templates.length !== wanted.length) {
  const got = new Set(templates.map((t) => t.workflowId));
  const missing = wanted.filter((id) => !got.has(id));
  console.error(
    `gen-openspec-skills: expected ${wanted.length} workflows, got ${templates.length}; missing: ${missing.join(', ')}`,
  );
  process.exit(1);
}

const transform = getSkillReferenceTransformer('agents');
const skillsRoot = join(root, '.agents/skills');
const keep = new Set(templates.map((t) => t.dirName));

mkdirSync(skillsRoot, { recursive: true });
for (const name of readdirSync(skillsRoot)) {
  if (!name.startsWith('openspec-') || keep.has(name)) continue;
  rmSync(join(skillsRoot, name), { recursive: true, force: true });
  console.log(`removed leftover .agents/skills/${name}`);
}

writeFileSync(join(skillsRoot, '.openspec-target'), 'agents\n');

for (const { template, dirName } of templates) {
  const dir = join(skillsRoot, dirName);
  mkdirSync(dir, { recursive: true });
  writeFileSync(join(dir, 'SKILL.md'), generateSkillContent(template, version, transform));
  console.log(`wrote .agents/skills/${dirName}/SKILL.md`);
}
