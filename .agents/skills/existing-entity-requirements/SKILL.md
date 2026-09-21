---
name: existing-entity-requirements
description: Capture current OpenSpec requirements for one named Terraform resource or data source that is about to change and has no spec yet. Use when the user names a single existing entity (type or package) and needs a baseline to express a delta. Do not use to backfill the provider surface, to spec ec_deployment as a whole, or to author a brand-new unimplemented entity (use new-entity-requirements).
---

# Existing entity requirements

Produce an **OpenSpec spec** (`spec.md`) for **one** existing Terraform resource or data source by reading its implementation. Capture current behavior only. Follow [`dev-docs/high-level/openspec-requirements.md`](../../../dev-docs/high-level/openspec-requirements.md): `## Purpose`, optional `## Schema`, `## Requirements` with `### Requirement:` / `#### Scenario:`; use **SHALL** / **MUST** in requirement bodies.

This skill is **on demand**: the entity is about to change and has no spec yet, so you need a baseline for the delta. It is **not** a mandate to spec every resource.

## Input

- **Entity**: the user names one Terraform type (e.g. `ec_deployment_traffic_filter`) or one implementation package (e.g. `ec/ecresource/trafficfilterresource`). If they name more than one entity, or ask to spec “everything”, refuse and ask them to pick a single type that is about to change.
- Resolve the type via `Metadata()` (`request.ProviderTypeName + "_..."`) and the registry in `ec/provider.go` (`Resources()` / `DataSources()`). There is no SDKv2 list.

## Hard stops

- **`ec_deployment` as a whole**: refuse. Ask which **slice** is changing (e.g. Integrations Server endpoints → capability id `deployment-integrations-server`). See the authoring guide. Locate the package if useful, then stop.
- **Surface-wide backfill**: refuse.
- **Unimplemented behavior**: do not write it into `openspec/specs/`. That is `new-entity-requirements` / an OpenSpec change.

## Workflow

1. **Locate implementation**
   - **Resource**: `ec/ecresource/<name>resource/` — `schema.go` (`Schema`, `Metadata`, `Configure`, `ImportState`, `ValidateConfig`), `create.go` / `read.go` / `update.go` / `delete.go`, validators, plan modifiers, and the API↔Terraform mappers (filenames vary; see [reference.md](reference.md) §8).
   - **Data source**: `ec/ecdatasource/<name>datasource/` — `Schema`, `Read`, `Metadata`, `Configure`.
   - One Go package may register **several** Terraform types (`projectresource` → `ec_elasticsearch_project` / `ec_observability_project` / `ec_security_project`; `privatelinkdatasource` → three endpoint data sources). Spec **one type**, not the package.
   - Use the checklist in [reference.md](reference.md).

2. **Examine the code path** (do **not** open an existing `openspec/specs/<id>/spec.md` yet — drafting from the exemplar makes “consistent with 1.3” tautological):
   - Schema (attributes, blocks, required/optional/computed, plan modifiers, validators).
   - Metadata: type name, import.
   - CRUD: which client and API, how `id` is set, create/update-then-read, “not found” on read, delete (including association teardown if any).
   - Client: `internal.ConvertProviderData` → `Stateful` (`cloud-sdk-go`) vs `Serverless` (`ec/internal/gen/serverless/`). Unconfigured-client guard (`Unconfigured API Client`).
   - Mapping: API → state and plan → payload (names vary: `modelToState` / expanders, or `ReadDeployment` / `ReadElasticsearches` / `ElasticsearchPayload`), empty vs null, unknown-in-plan.
   - Lifecycle: `RequiresReplace` vs in-place update.
   - Type-gated validation (`ValidateConfig`) when present.
   - State upgrade: `schema.Schema` `Version`, `UpgradeState` / `ResourceWithUpgradeState`. If Version is non-zero and there is no upgrader, that is a known gap (do not invent an upgrader). See [reference.md](reference.md).

3. **Write policy**
   - Capability id: kebab-case **without** the `ec_` prefix (e.g. `deployment-traffic-filter`). One directory per Terraform type.
   - **No canonical spec** (`openspec/specs/<id>/` missing): write `openspec/specs/<id>/spec.md` as a current-behavior baseline. Then run `make check-openspec`. Tell the user to **review accuracy and freeze requirement headings** before the follow-up change — OpenSpec deltas match on heading text; unstable names make the delta fail to fold in.
   - **Spec already exists**: **do not write**. Finish the draft from code, *then* read the canonical spec and report a comparison in the conversation. If they want new/changed behavior, that is an OpenSpec change (`openspec-propose` / `openspec-new-change`), not an overwrite.
   - Never invent a scratch path under the repo for the draft.

4. **Spec shape** (copy the 1.3 exemplars, not a new layout):
   - H1 title: Terraform type name in backticks (e.g. `` # `ec_deployment_traffic_filter` ``).
   - **Required line** (load-bearing for `requirements-verification`): `Resource implementation:` or `Data source implementation:` followed by the Go package path in backticks.
   - **Purpose**: in scope / out of scope in plain language.
   - **Schema**: optional HCL with `<required|optional|optional+computed|computed>`, types, brief notes.
   - **Requirements**: named `### Requirement:` headings. Do **not** invent `REQ-xxx` ids (the 1.3 exemplars do not use them). Each body MUST contain **SHALL** or **MUST**. Each requirement has `#### Scenario:` Given / When / Then blocks.
   - Optional **Known gaps** for current-behavior limitations (non-normative). Keep the `SHALL` as what the code does.
   - Style references (read **after** drafting, or when this entity has no spec yet): `openspec/specs/deployment-traffic-filter/spec.md` (resource), `openspec/specs/stack/spec.md` (data source).

5. **Quality**
   - Every requirement traces to a file/function. Do not invent behavior.
   - Schema and requirements must agree (optional vs required, computed vs configured).
   - Agents never run `make testacc` / `TF_ACC`. Acc coverage may be named as a *human/Buildkite* step (`ec/acc/`, `TestAcc…`).

## Output format

~~~~markdown
# `ec_example`

Resource implementation: `ec/ecresource/exampleresource`

## Purpose
...

## Schema
```hcl
...
```

## Requirements

### Requirement: Short name

The resource SHALL ...

#### Scenario: ...
- GIVEN ...
- WHEN ...
- THEN the provider SHALL ...
~~~~

## Reference

- Authoring: `dev-docs/high-level/openspec-requirements.md`
- Resource exemplar (after drafting): `openspec/specs/deployment-traffic-filter/spec.md`
- Data-source exemplar (after drafting): `openspec/specs/stack/spec.md`
- Code-path checklist and categories: [reference.md](reference.md)
