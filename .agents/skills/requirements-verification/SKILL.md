---
name: requirements-verification
description: Analyze one OpenSpec capability spec (or a change delta) for internal consistency, implementation compliance, and test opportunities. Use when reviewing a single spec against its Go package. Do not use to gate an entire OpenSpec change before archive (that is openspec-verify-change, Phase 1.5). When a shell is available, run openspec validate first. Do not run acceptance tests (TF_ACC).
---

# Requirements verification

Analyze an OpenSpec spec (`openspec/specs/<capability>/spec.md` or a delta under `openspec/changes/**/specs/`) and produce three outputs:

1. **Internal consistency** — requirements vs each other and vs the schema.
2. **Implementation compliance** — whether the Go code meets each requirement.
3. **Test opportunities** — unit tests the agent can run; acceptance tests named for a **human/Buildkite** only.

## Input

- **Spec**: path or entity name. Resolve to `spec.md`.
- **Implementation**: from the spec’s `Resource implementation:` / `Data source implementation:` line. If that line is missing, treat it as incomplete (flag it) and locate the package via `Metadata()` / `ec/provider.go`.

## Workflow

### 0. Structural validation (when a shell is available)

`make check-openspec` or `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate --all`

Single spec (optional): `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate <capability> --type spec` (e.g. `deployment-traffic-filter`).

This checks **structure** (Requirement/Scenario shape, SHALL/MUST in bodies) — not semantics or code compliance.

If validation **fails**, report CLI errors first and fix (or stop) before sections 1–3. If there is no terminal, note that validate was not run.

A file under this skill’s `testdata/` is **not** part of `openspec/` and is not seen by `make check-openspec`. Use it only when asked to exercise this skill on an intentionally incomplete spec.

### 1. Parse the spec

Extract: title/type name, implementation path, Purpose, Schema HCL if present.

List every `### Requirement:` section by **heading text** (e.g. `Ruleset identity`). Do **not** invent `REQ-xxx` ids unless the heading already has them. Infer a category from content (see [reference.md](reference.md)).

Flag as incomplete before going deep:

- Missing `## Purpose`
- Missing `Resource implementation:` / `Data source implementation:`
- A requirement body with no **SHALL** / **MUST**
- A requirement with no `#### Scenario:`
- Schema and requirements that disagree on required/optional/computed

### 2. Internal consistency

Apply [reference.md](reference.md) (Consistency checks). Result: **Consistent** or a list of heading-vs-heading problems.

### 3. Implementation compliance

Resolve the package. Read `schema.go`, CRUD files, validators, and `Schema.Version` / `UpgradeState` if present. Mapping is **not** always `expanders.go` / `flatteners.go`: search for the functions that read API responses into Terraform models (e.g. `ReadDeployment`, `ReadElasticsearches`) and those that build API payloads (e.g. `ElasticsearchPayload`). For each requirement, mark **Met** / **Not met** / **Unclear** with evidence (file/function or “not found”). Mapping: [reference.md](reference.md).

### 4. Test opportunities

- **Unit**: `*_test.go` in the entity package (and helpers). These the agent may run via `make unit` (no `TF_ACC`).
- **Acceptance**: `ec/acc/` (`TestAcc…`). List gaps as **human/Buildkite** suggestions. **Never** run `make testacc` or set `TF_ACC=1`.

### 5. Report

```markdown
# Requirements analysis: <entity name>

**Document**: `openspec/specs/.../spec.md`
**Implementation**: `ec/ecresource/...` (or data source path)

## 1. Internal consistency

- **Result**: Consistent | Inconsistent
- **Inconsistencies** (if any): [Heading A] vs [Heading B]: ...
- **Incomplete** (if any): missing Purpose / implementation line / SHALL / scenarios / schema drift

## 2. Implementation compliance

| Requirement | Category | Status | Evidence |
| --- | --- | --- | --- |
| Ruleset identity | Identity | Met | create.go sets id from API |
...

**Summary**: X met, Y not met, Z unclear.

## 3. Test opportunities

| Requirement | Type | Suggested test | Verifies | Who runs |
| --- | --- | --- | --- | --- |
| Import by id | Unit | ImportState with id; state id set | Import passthrough | agent (`make unit`) |
| Create then read | Acceptance | `TestAccDeploymentTrafficFilter_basic` already covers create | Live create/read | human/Buildkite |
```

## Reference

- Authoring / CLI: [`dev-docs/high-level/openspec-requirements.md`](../../../dev-docs/high-level/openspec-requirements.md)
- Consistency checks and implementation mapping: [reference.md](reference.md)
- Code-path checklist: `.agents/skills/existing-entity-requirements/reference.md`
- Intentionally incomplete fixture (AC / skill exercise only): [testdata/incomplete-spec.md](testdata/incomplete-spec.md)
