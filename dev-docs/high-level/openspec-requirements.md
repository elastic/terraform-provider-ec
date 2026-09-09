# OpenSpec requirements

This repository uses [OpenSpec](https://openspec.dev/) for **living functional requirements**.
Canonical specs live under [`openspec/specs/`](../../openspec/specs/) and describe behavior that is
already in the provider. Proposed new or changed behavior lives under `openspec/changes/` until the
implementation is archived into `openspec/specs/`.

This page is the authoring contract: file layout, requirement phrasing, and when to write a spec.
Create, apply, and archive a change with the skills in [`.agents/skills/`](../../.agents/skills/)
— which to reach for is in [`openspec-workflows.md`](./openspec-workflows.md). GitHub Agentic
Workflows (`change-factory`, `verify-openspec`) land in later phases.

## Layout

- **`openspec/specs/<capability>/spec.md`** — One capability per directory. Each file includes:
  - **`## Purpose`** — Scope in plain language (what is in, what is out).
  - **`## Schema`** (optional) — HCL sketch of the Terraform resource or data source; appendix-style,
    not a substitute for the Go schema.
  - **`## Requirements`** — `### Requirement: …` blocks using **SHALL** / **MUST** (RFC 2119).
  - **`#### Scenario: …`** — Given / When / Then checks reviewers and agents can trace to code or tests.
- **`openspec/changes/<name>/`** — An in-progress **change**: proposal, design, tasks, and **delta
  specs**. This is the default path for any work that adds or updates requirements. After the code
  matches the change, **openspec-sync-specs** / **openspec-archive-change** fold the deltas into
  `openspec/specs/` and move the change under `openspec/changes/archive/`.
- **`openspec/config.yaml`** — Project OpenSpec configuration (context, per-artifact rules, apply/archive
  guidance). The generated skills read this; do not fork the `SKILL.md` files to inject conventions.

Copy from the seed exemplars rather than inventing a new shape:

- [`openspec/specs/deployment-traffic-filter/spec.md`](../../openspec/specs/deployment-traffic-filter/spec.md)
  — small **resource** (`ec_deployment_traffic_filter`).
- [`openspec/specs/stack/spec.md`](../../openspec/specs/stack/spec.md) — small **data source**
  (`ec_stack`).

## Capability ids

Pick a stable kebab-case id **without** the `ec_` Terraform type prefix. It names both the canonical
directory `openspec/specs/<id>/` and any delta spec under a change.

| Terraform type | Capability id |
| --- | --- |
| `ec_deployment_traffic_filter` | `deployment-traffic-filter` |
| `ec_stack` | `stack` |

For a slice of a large resource, name the slice, not the monolith — e.g. a future
`deployment-integrations-server` spec for Integrations Server endpoints, **not** a single
`deployment` spec covering all of `ec_deployment`.

Automation that we author later uses a `ci-*` prefix (for example `ci-aw-openspec-verification`).
Do not retro-spec existing CI; only new or changed automation needs a spec.

## When to write a spec

**Going forward, only new or changed behavior needs a spec.** Do not backfill the rest of the
provider. The two seed specs above exist so authors have something to copy; they are not a mandate
to spec every existing resource.

Write a spec (via a change) when any of these is true:

- New Terraform resource or data source.
- New or changed user-visible attributes, validation, or CRUD behavior.
- A bug fix that changes observable behavior (the spec states the intended behavior).
- New CI / automation behavior that we choose to pin (`ci-*`).

Skip a spec for mechanical work with no spec impact (lint, dependency bumps, typo, generated-docs
refresh, changelog-only).

## Where to put it

| Situation | Location |
| --- | --- |
| New or changed behavior (the default) | `openspec/changes/<id>/` — delta specs plus proposal / design / tasks. After implement + verify, archive into `openspec/specs/`. |
| Behavior already in the product, captured as a copyable example (this seed) | `openspec/specs/<capability>/spec.md` directly. |
| Tiny follow-up on an existing canonical spec (typo, link, wording) | Edit `openspec/specs/` directly. |

Do **not** put unimplemented behavior into `openspec/specs/`. If the code does not do it yet, it
belongs in a change. Example: exposing an Integrations Server `ingest` endpoint is a change, not a
canonical requirement, until that code lands.

A change that extends an existing capability adds a delta against `openspec/specs/<capability>/`. A
change that introduces a brand-new capability ships its first `spec.md` as a delta; archive creates
the canonical directory.

## Authoring requirements

Each `### Requirement:` block states one observable rule with **SHALL** or **MUST**. Keep the
requirement at the Terraform / API-contract level (schema, validation, CRUD, error diagnostics),
not a line-by-line transcription of Go.

Each requirement has one or more `#### Scenario:` blocks:

```
#### Scenario: <short name>

- GIVEN <precondition>
- WHEN <action>
- THEN the provider SHALL <observable result>
```

Scenarios are the reviewable unit: a reader should be able to point at a test or a code path that
would fail if the requirement were violated. Prefer a few precise scenarios over exhaustive
combinatorics.

Optional `## Schema` sketches use HCL with `<required|optional|computed, type>` annotations. They
orient the reader; the Requirements section is normative.

When a current-behavior spec captures a known limitation, keep the `SHALL` as the observable rule
and add a non-normative **Known gaps** note. A later bug fix is a change that updates the `SHALL`.

## Cloud-provider constraints

These apply to every spec in this repo and to any later change that implements one:

- The provider is 100% [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) (no SDKv2).
- There is **no local Docker stack**. Hosted ESS calls go through `cloud-sdk-go`; serverless calls
  go through the generated client under `ec/internal/gen/serverless/`.
- **Agents never run acceptance tests** (`TF_ACC`). Acc hits the real, paid Elastic Cloud API and
  runs out-of-band on Buildkite. Specs may mention acc coverage as a *human* verification step; they
  MUST NOT require an agent to execute `make testacc`.
- User-facing implementation PRs add `.changelog/{PR}.txt` (not a PR-body changelog block). Spec /
  docs-only PRs skip it. See [`contributing.md`](./contributing.md).
- `make check-openspec` is **not** part of `make lint`. CI runs it in
  [`.github/workflows/openspec.yml`](../../.github/workflows/openspec.yml).

## Validation

`make check-openspec` installs the pinned OpenSpec CLI (`make setup-openspec`, Node.js 24) and runs
`openspec validate --all`. That checks structure and normative keywords. It does **not** prove the
Go implementation matches every requirement — that is code review (and, later, the
`openspec-verify-change` skill / `verify-openspec` label). Operational loop:
[`openspec-workflows.md`](./openspec-workflows.md).

Run it locally whenever you touch `openspec/`.

## References

- OpenSpec: [GitHub — OpenSpec](https://github.com/Fission-AI/OpenSpec)
- Contributor setup (Node 24, `make setup-openspec`): [`CONTRIBUTING.md`](../../CONTRIBUTING.md)
- Day-to-day make targets: [`development-workflow.md`](./development-workflow.md)
- Coding conventions: [`coding-standards.md`](./coding-standards.md)
