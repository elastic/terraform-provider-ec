---
name: schema-coverage
description: Analyzes an Elastic Cloud Terraform resource or data source schema and compares it to attributes used in the acceptance test suite (configs + assertions). Produces a ranked (high-risk-first) coverage-gap report for ec_* entities. Use when the user asks about schema coverage, test coverage gaps, or improving Terraform acceptance tests for an Elastic Cloud deployment, serverless project, traffic filter, snapshot repository, organization, or other ec_* type.
user-invocable: true
argument-hint: "[ec_* entity name]"
---

# Schema Coverage (ec_* entity vs acceptance tests)

## Goal

Given **one** Terraform entity (`ec_*` resource or data source), compare its Plugin Framework schema to what the acceptance tests **configure** and **assert**, then emit a **ranked (high-risk-first)** coverage-gap report.

This skill is **read-only analysis**. It does not change provider code, write tests, or run them.

Read [reference.md](reference.md) before extracting schema or matching tests. Follow it for Framework shapes, acc-config discovery, excluded steps, import, path normalization, and ranking tags.

## Safety

- **Never** set `TF_ACC` or run `make testacc`. Acc hits the real, paid Elastic Cloud API.
- Do not execute tests to "see what is covered". Coverage is inferred from source (`ec/acc/` Go + testdata, plus colocated unit tests as an annotation only).
- For the `ec_deployment` **resource**, analyze **one scope per run** unless the user asks for the whole resource:
  - `root` — top-level attributes on `DeploymentSchema()` (`name`, `region`, `version`, `traffic_filter`, credentials, …)
  - one component subtree: `elasticsearch`, `kibana`, `apm`, `integrations_server`, or `enterprise_search`

## Inputs (infer if not provided)

Primary contract: **one entity**. Do not enumerate the whole provider unless the user asked for that.

- Entity name (e.g. `ec_deployment_traffic_filter`, `ec_stack`) **and kind** (`resource` or `data_source`)
- Or a schema file / package path
- Or an open acc test — infer from `resName := "ec_..."` / `resource.ParallelTest`

`ec_deployment` is registered as **both** a resource and a data source (`ec/provider.go`). They have different packages, schemas, and state paths. If the user names `ec_deployment` without a kind, **ask**. Never merge their acc files into one report.

If the user did not name an entity, ask. Fallback only: locate TypeNames via `ec/provider.go` `Resources()` / `DataSources()` (see reference for the suffix-grep recipe), then ask which one (and which kind when the name is on both lists).

## Workflow

### 1) Locate and parse the schema

Find the owning package (reference: suffix-grep `response.TypeName`). Parse the **current** Framework schema (`resp.Schema = …` or helper such as `v2.DeploymentSchema()`). Skip `*/v1/` packages — those are state upgraders.

For each attribute/block, record:

- Path (see reference path normalization — `ec_deployment` **resource** v2 is `elasticsearch.hot.size`; the **data source** uses indexed paths like `elasticsearch.0.…`. Do not mix them.)
- Required / Optional / Computed
- Type (String, Bool, Int64, List, Set, Map, SingleNested, …)
- Validators, plan modifiers, defaults
- Nesting depth

Exclude `timeouts` unless it is declared in that entity's `Attributes` or `Blocks` map.

### 2) Locate acceptance tests and collect usage

Acc lives in `ec/acc/`, not next to the entity package. Match the type name **and kind** in test files, testdata HCL, and helpers (`ec_deployment_traffic_filter` appears in acc/docs; schema packages set `request.ProviderTypeName + "_suffix"`). For `ec_deployment`, keep resource tests (`deployment_*.go`) separate from data-source tests (`datasource_deployment_*.go`).

Follow **both** `Config:` and `Check:` helpers to their definitions (they may live in another file in `ec/acc/`). Dominant config pattern is `Config:` + Go helper + `os.ReadFile` + `fmt.Sprintf`; `ConfigDirectory` is rare.

Collect two signals from **included** steps only:

- **Configured**: set in that step's HCL (including nested blocks). Import steps have no `Config` of their own — they inherit the previous included step's config.
- **Asserted**: `TestCheckResourceAttr`, `TestCheckResourceAttrPair`, `TestCheckResourceAttrWith`, `TestCheckResourceAttrSet`, `TestCheckNoResourceAttr`, `TestMatchResourceAttr`, type-set helpers, `"block.#"` / `"tags.%"` — including those inside `Check:` helpers.

Also record **import** (see reference). `ImportStateVerify: true` is **not** "every schema path is asserted". It only supports paths that were non-null in the imported state (i.e. configured in the inherited config, or computed-and-present). It **never** lifts a never-configured attribute out of no-coverage.

**Do not count** `ExpectError`, `PlanOnly: true`, `ExternalProviders` (old provider), or an **unconditional test-level** `t.Skip("reason")`. The `requiresAPIConn` / `TF_ACC != "1"` skip is the acc gate, **not** an exclusion.

Attributes that appear **only** in excluded steps get the reason `excluded-step-only`, not "covered".

### 3) Build a coverage matrix

For each schema path:

- Configured? Asserted? Assertion quality (value-specific / set-only / absence)
- Value diversity across included steps (do not flag env-derived `region` / `version` / template / instance-configuration IDs as `single-value` — see reference)
- Optional-unset, empty-collection, update coverage (update is **N/A** for data sources)
- Import-verify / import-ignore (and any `ImportPlanChecks` next to an ignore)
- Unit-test annotation (`offline` vs `needs-live-api`) — **never** reduces the acc coverage class

### 4) Classify and rank

**No coverage:** never configured in an included step **and** never referenced by an included check (`TestCheckResourceAttr`, `…Pair`, `…With`, `…Set`, `TestCheckNoResourceAttr`, `…Match…`, type-set, `#`/`%`). Import-verify of a null/absent path does not count as a check. An omit + `TestCheckNoResourceAttr` is **not** no coverage (that is the optional-unset case; it may still be poor if the attribute is never set to a value).

**Poor coverage:** configured and/or asserted, but weak (configured-never-asserted, set-only where a value is deterministic, single value only, optional never unset, collection never empty **only if empty is schema-valid**, no in-place update on a resource **except** `RequiresReplace` fields, import-ignored).

Rank **high-risk first**. Do not assign numeric scores. Sort using the tags in the reference. Computed-only ids/endpoints sort last unless they have validators or plan modifiers.

### 5) Zero-acc entities

If nothing under `ec/acc/` references the type name, say so once. Treat the entity as wholly uncovered. Rank only **entry-point** gaps worth a first test (required/writable top-level attributes and the main nested block), not every leaf.

### 6) Produce the report

Use the templates below. Markdown is the 4.4-shaped deliverable (stable headings). Append one fenced JSON block that mirrors the same fields — not a separate versioned API. No `generated_at`, no `risk_score`. Put `close_via` on **every** gap (`none` and `poor`).

## Report template (markdown)

```markdown
## Schema coverage report: <entity>

### Entity
- **Name**: <ec_…>
- **Kind**: resource | data_source   (required; `ec_deployment` is both)
- **Implementation directory**: <path from repo root>
- **Schema**: <file(s) or function(s)>
- **Acceptance tests**: <file(s), or "none">

### Attributes with no coverage
- `<attr_path>`: <Required/Optional/Computed>, <type>. Tags: <reason tags>. **Close via**: `offline` | `needs-live-api`. **Gap**: …

### Attributes with poor coverage
- `<attr_path>`: <schema flags/type>. Tags: <reason tags>
  - **Observed**: <configured/asserted/import, example values>
  - **Gaps**: <set-only | single value | no unset | no empty collection | no update | import-ignored | excluded-step-only | …>
  - **Close via**: `offline` | `needs-live-api`
  - **Suggested improvement**: <concrete step or assertion>

### Prioritized top 5
1. `<attr_path>` — <why this is first>
2. …

### Concrete test additions
- <smallest diffs that would close the top gaps; name TestAcc vs unit-test file>
```

If there are no acc files, keep the same headings; under no-coverage state that the entity has zero acceptance coverage, and keep top 5 / concrete additions to a first-test sketch.

## Report template (JSON, advisory)

Append after the markdown, same analysis:

```json
{
  "entity": {
    "name": "ec_example",
    "kind": "resource",
    "implementation_dir": "ec/ecresource/example",
    "schema_files": ["ec/ecresource/example/schema.go"],
    "acc_files": ["ec/acc/example_test.go"]
  },
  "gaps": [
    {
      "path": "name",
      "required": true,
      "optional": false,
      "computed": false,
      "type": "String",
      "coverage": "none",
      "reasons": ["no-coverage", "required", "writable"],
      "suggestion": "…",
      "close_via": "needs-live-api"
    }
  ],
  "top_5": ["name"]
}
```

`close_via` is `offline` or `needs-live-api` on every gap. `coverage` is `none` or `poor`. `reasons` uses the closed tag set in the reference.

## Rules of thumb

- Prefer **value-specific** assertions over “is set”.
- Optional: at least one included step that **omits** the attribute, plus absence/default assertion.
- Collections: empty (or omitted) case + `.#` / `.%` assertion **only if empty is valid** (no `SizeAtLeast(1)` / equivalent). `ec_deployment_traffic_filter.rule` cannot be empty — do not suggest that case.
- Resources: at least one multi-step **in-place** update with a post-update value assertion. Skip `no-update` when the attribute has `RequiresReplace()`. Data sources: skip update.
- Computed-only: set-only is acceptable when the value is non-deterministic.
- Configured-but-never-asserted is **poor**, not good.
- Unit tests never mark an acc gap as covered.
- Suggestions that need live CRUD are acc stubs for a human/Buildkite; never instruct anyone to run `TF_ACC` from this skill.
