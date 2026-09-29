# Reference: schema ↔ test matching (cloud provider)

This repo is 100% Terraform Plugin Framework. If you see `map[string]*schema.Schema` / `ForceNew`, you are in a vendored or unrelated directory — stop and return to `ec/`.

## Entity discovery

Do not hardcode the resource/data-source list. Canonical registry: `Resources()` and `DataSources()` in `ec/provider.go`. Look up **kind** first (`Resources()` vs `DataSources()`), then TypeName. `ec_deployment` is on **both** lists (`deploymentresource` vs `deploymentdatasource`) with different schemas and path conventions. A report must pick one kind and one acc population.

Schema packages almost never contain the literal `ec_…` string. They set `response.TypeName = request.ProviderTypeName + "_suffix"` (or `fmt.Sprintf("%s_%s_…", request.ProviderTypeName, …)`). To find the owning package, grep the **suffix** in the matching tree (`ec/ecresource` or `ec/ecdatasource`):

```
rg 'TypeName' ec/ecresource ec/ecdatasource
rg '"_deployment_traffic_filter"' ec/
```

Grep the full type name (`ec_deployment_traffic_filter`) under `ec/acc/` and `docs/` / testdata, **filtered by kind**. For `ec_deployment` resource use `deployment_*.go` / `testdata/deployment_*.tf` (not `datasource_*`). For the data source use `datasource_deployment_*.go` / `testdata/datasource_deployment_*.tf`.

Two packages register **three** TypeNames each:

- `ec/ecresource/projectresource` — `ec_{elasticsearch,observability,security}_project`. Shared `Resource[T].Schema` delegates to `elasticsearch.go` / `security.go` / `observability.go`.
- `ec/ecdatasource/privatelinkdatasource` — `ec_aws_privatelink_endpoint`, `ec_gcp_private_service_connect_endpoint`, `ec_azure_privatelink_endpoint`. One generic implementation; CSP + name come from `d.csp` / `d.privateLinkName` in `aws_datasource.go` / `gcp_datasource.go` / `azure_datasource.go`.

### `ec_deployment` **resource** schema graph

Current schema is **v2**. Entry point: `ec/ecresource/deploymentresource/resource.go` assigns `v2.DeploymentSchema()`, defined in `ec/ecresource/deploymentresource/deployment/v2/schema.go`. That composes `elasticsearch/v2`, `kibana/v2`, `apm/v2`, `integrationsserver/v2`, `enterprisesearch/v2`, `observability/v2`. Ignore sibling `*/v1/` packages (state upgrade only).

Scope a run to **one** of:

- `root` — the top-level attributes on `DeploymentSchema()` (`id`, `alias`, `version`, `region`, `deployment_template_id`, `name`, `request_id`, credentials, `traffic_filter`, `tags`, …). These belong to no component subtree.
- one nested component: `elasticsearch`, `kibana`, `apm`, `integrations_server`, `enterprise_search` (prefer these over `observability`, which is a small nested object whose name collides with the serverless project).

## Schema extraction (Plugin Framework)

Schemas live in:

- `Schema(ctx, req resource.SchemaRequest, resp *resource.SchemaResponse)`
- `Schema(ctx, req datasource.SchemaRequest, resp *datasource.SchemaResponse)`

Look for `resp.Schema = schema.Schema{ Attributes: …, Blocks: … }` or a helper that returns `schema.Schema`. Follow helpers until keys are concrete (`genericSchema()`, `trafficFilterRuleSchema()`, `v2.DeploymentSchema()`, …).

Capture **all** keys, including nested objects.

1. Top-level `Attributes: map[string]schema.Attribute{ … }`
2. `Blocks: map[string]schema.Block{ … }` — `ListNestedBlock`, `SetNestedBlock`, `SingleNestedBlock`
3. Nested object attributes — `SingleNestedAttribute`, `ListNestedAttribute`, `SetNestedAttribute`, `MapNestedAttribute`
4. Collection primitives — `ListAttribute` / `SetAttribute` / `MapAttribute` with `ElementType`

Record per path:

- `Required` / `Optional` / `Computed` (Optional+Computed is common for defaults / server-populated)
- Validators (`Validators: []validator.…`, also `setvalidator.SizeAtLeast`, resource-level `ConfigValidators` / `ValidateConfig`)
- Plan modifiers, including custom `ec/internal/planmodifiers`:
  - `RequiresReplace()`, `UseStateForUnknown()`
  - `BoolDefaultValue`, `StringDefaultValue`, `SetDefaultValue`
  - `UseStateIfNotNullForUnknown()` (file `use_state_if_known.go`)
  - `UseStateForUnknownUnlessMigrationIsRequired`
- Defaults (plan-modifier defaults count as `has-defaults`)
- Nesting depth (0 = top-level)

### `timeouts`

Do **not** flag `timeouts` unless it appears in that entity's `Attributes` or `Blocks` map. Import tests often list `"timeouts"` in `ImportStateVerifyIgnore` as a leftover; that is not a schema gap.

## Acceptance tests (`ec/acc/`)

Not under the entity package. Search `ec/acc/` for the type name **and kind** in `*_test.go` and `testdata/`. Do not mix resource and data-source tests when the TypeName is shared.

### Config shapes (follow the helper)

**Dominant:** `resource.TestStep{ Config: <helper>(…) }`. The helper is usually `os.ReadFile("testdata/<file>.tf")` then `fmt.Sprintf` with positional `%s` (region, template, names). Those `.tf` files are **not** standalone HCL — read the helper to know which placeholders map to which values, then treat the substituted keys as configured attributes.

Examples: `fixtureAccDeploymentResourceBasic`, `fixtureDeploymentDefaults`, `cfgF`. Shared readers live in files like `ec/acc/deployment_fixture_test.go`.

**Inline heredoc:** the helper returns a raw HCL string (`fmt.Sprintf(\`resource "ec_…" {…}\`)`). Parse that string. Example: `ec/acc/serverless_traffic_filter_test.go`.

**Rare:** `ConfigDirectory: config.StaticDirectory("testdata/…")` plus `ConfigVariables`. Directories of real HCL + variable interpolation. Today this is traffic-filter tests (`deployment_traffic_filter_test.go`, `datasource_traffic_filter_test.go`).

If you only look for `ConfigDirectory`, you will report near-total "no coverage" for most entities. Always follow `Config:` helpers.

### Check helpers (follow these too)

Baseline assertions are often wrapped in a package-level helper invoked from many steps, sometimes defined in another file (`checkBasicDeploymentResource` in `deployment_basic_test.go`, `testAccCheckDeploymentExists` in `deployment_checks_test.go`, `checkBasicDeploymentTrafficFilterResource` in `deployment_traffic_filter_test.go`). Follow `Check:` the same way as `Config:`. Missing this produces false "configured but never asserted".

### Assertions

- `resource.TestCheckResourceAttr(name, "path", "value")` — value-specific
- `resource.TestCheckResourceAttrPair(nameA, "pathA", nameB, "pathB")` — value-specific equality (counts as asserted for **both** paths; this is how `ec_deployment` data-source checks and `observability.deployment_id` work)
- `resource.TestCheckResourceAttrWith(name, "path", func)` — value-specific predicate (e.g. `ec/acc/deployment_add_dedicated_master_test.go`)
- `resource.TestCheckResourceAttrSet(name, "path")` — set-only
- `resource.TestCheckNoResourceAttr(name, "path")` — absence
- `resource.TestMatchResourceAttr(name, "path", regexp)`
- `resource.TestCheckTypeSetElemAttr` / `TestCheckTypeSetElemNestedAttrs(name, "block.*", map[string]string{…})`
- Length: `"block.#"` / `"tags.%"`

### Import

Import steps typically set `ImportState: true` with **no** `Config` — they reuse the previous included step's config.

`ImportStateVerify: true` round-trips attributes that are **non-null in the imported state**. That is weaker than a value assertion, stronger than nothing.

It does **not** mean every schema path is asserted:

- A never-configured optional stays null; the round-trip is vacuous. Leave it in **no coverage**.
- Import-verify evidence applies only to paths present in that imported state (configured in the inherited config, or computed-and-populated).
- Import-verify alone never upgrades a never-configured path out of `no-coverage`. At best, a configured-but-only-import-verified path is **poor**.

`ImportStateVerifyIgnore: []string{"a", "b"}` — each entry is a **named gap** (`import-ignored`), except `timeouts` as above. If the same step (or the next) has `ImportPlanChecks` / `ExpectEmptyPlan` that re-check the ignored path (e.g. `product_types` on `ec_security_project`), record that under **Observed** so you do not recommend a test that already exists.

### Update coverage (resources only)

Skip `no-update` when the attribute has `RequiresReplace()` (or equivalent). Those fields cannot change in place — `snapshot_repository.name`, keystore setting name/value, traffic-filter association ids, `encryption_key_path`. Do not suggest an update step for them.

An attribute has in-place update coverage only if:

- It is **not** replacement-only, AND
- Multiple **included** steps apply different configs for the same resource, AND
- The attribute's value meaningfully differs, AND
- A post-update check asserts the new value (prefer exact match over set-only).

Data sources: mark update as `n/a`. Do not flag "no update coverage" on a data source.

### Excluded steps (not coverage)

Do not count configured or asserted attributes from:

- `ExpectError` — the apply is expected to fail (e.g. `TestAccDeploymentTrafficFilter_azure`)
- Unconditional **test-level** `t.Skip("reason")` / `t.Skipf` — state-upgrade skips, issue-linked skips (e.g. #443 / #746)
- `PlanOnly: true` — no apply, no post-apply state
- `ExternalProviders` pinning an old `elastic/ec` (e.g. `0.4.1`) — old schema, not current

**Not an exclusion:** `requiresAPIConn` in `ec/acc/acc_prereq.go` calls `t.Skip()` when `TF_ACC != "1"`. That is the live-API gate. `fixtureDeploymentDefaults` and most deployment tests call it. Treating that skip as "test excluded" would drop almost all `ec_deployment` coverage. This skill never sets `TF_ACC`; it still **reads** those tests as included.

If an attribute is configured **only** in excluded steps, classify it as none (if never in an included step) or poor with reason `excluded-step-only`. Do not treat that as coverage.

### Env-derived values (do not flag as `single-value`)

`region` (`getRegion()` / `EC_REGION`), `version` (`latestStackVersion()`), `deployment_template_id` (`setDefaultTemplate` / `buildAwsTemplate`), and `*instance_configuration_id` are taken from env or the live API. Acc then asserts those same runtime values. That is not "only one literal was ever tested". Do not tag them `single-value` and do not recommend a second hardcoded value unless the user asked for multi-region coverage.

## Path normalization

State paths in this repo:

- Top-level: `name`
- **`ec_deployment` resource** (v2, single nested): `elasticsearch.hot.size`, `kibana.region`, `apm.config.%` — **no** `.0.` index. Do not emit `elasticsearch.0.hot.0.size`.
- **`ec_deployment` data source**: indexed paths (`elasticsearch.0.ref_id`, `kibana.0.topology.0.size`). Do not apply the resource convention here.
- List/set **blocks** or set attributes: `rule.0.source`, `rule.#`, `rule.*`, `traffic_filter.#`
- Map: `tags.%` / `tags.key`

Match schema children of lists/sets as `block[*].attr`. For sets, prefer `TypeSetElem…` over hard-coded indices.

`ec_deployment` v2 is almost entirely `SingleNestedAttribute` with named topology keys (`hot`, `warm`, …). Set-typed exceptions at root include `traffic_filter` (set of strings) and nested sets such as `elasticsearch.trust_account`.

## Coverage classes

### No coverage

Path never configured in an **included** step and never referenced by an **included** check (`TestCheckResourceAttr`, `…Pair`, `…With`, `…Set`, `TestCheckNoResourceAttr`, `…Match…`, type-set, `#`/`%`). Import-verify of a null path does not count. Omit + `TestCheckNoResourceAttr` is the optional-unset case, not no coverage (it may still be **poor** if the attribute is never set to a value).

### Poor coverage

Any of:

- Configured but never asserted (import-verify-only counts as weak assertion, still poor if that is all)
- Set-only where a deterministic value assertion is feasible
- Single value only (except env-derived values above)
- Optional never unset (no omit + absence/default assert)
- Collection never empty **and** empty is schema-valid (no `setvalidator.SizeAtLeast(1)` / equivalent — `trafficFilterRuleSchema` cannot be empty; do not tag `rule` `no-empty-collection` or suggest an empty case)
- Resource attribute never updated in place (skip when `RequiresReplace()`; skip on data sources)
- `import-ignored`
- `excluded-step-only` (configured only in a skipped/error/plan-only/old-provider step)

## Ranking tags (closed set)

Use these in `Tags:` / JSON `reasons`. Do **not** invent numeric scores.

| Tag | When |
| --- | --- |
| `no-coverage` | Class is none |
| `required` | `Required: true` |
| `writable` | Optional or required, not computed-only |
| `has-validators` | Schema, block, resource `ConfigValidators`, or `ValidateConfig` (`ResourceWithValidateConfig`) |
| `has-plan-modifiers` | Any plan modifier, including custom defaults |
| `has-defaults` | Default via plan modifier or framework default |
| `nested` | Nesting depth ≥ 1 |
| `set-only` | Only `TestCheckResourceAttrSet` (and value is deterministic) |
| `single-value` | Only one distinct configured/asserted value (not env-derived) |
| `no-unset` | Optional never omitted + asserted |
| `no-empty-collection` | Collection never empty **and** empty is allowed by validators |
| `no-update` | Resource; no in-place update (omit on data sources and on `RequiresReplace`) |
| `import-ignored` | In `ImportStateVerifyIgnore` |
| `excluded-step-only` | Only appears in excluded steps |
| `computed-only` | Computed, not optional/required — sort these **down** |

High-risk first: `no-coverage` + `required`/`writable` + validators/plan-modifiers + deeper nesting. Then poor writable. Computed-only ids last.

## Unit tests (annotation only)

Colocated `*_test.go` under the entity package (expand/flatten, validators, plan modifiers). They **never** change `coverage` from `none`/`poor` to covered.

Set `close_via` on every gap:

- `offline` — a table-driven unit test in that package could close the gap (validator, plan modifier, default, expand/flatten round-trip)
- `needs-live-api` — needs an acc config/assert/CRUD round-trip (human or Buildkite; this skill never runs it)

## Zero-acc entities

If `ec/acc/` has no match for the type name, do not hunt forever. Report:

- Acc files: none
- Entity wholly uncovered
- Top 5 = entry points only (required/writable top-level + the primary nested block), not every nested leaf

Known today (verify with grep; do not treat this list as source of truth): `ec_organization`, `ec_snapshot_repository`, the three PrivateLink data sources.

## Report checklist

- Headings from `SKILL.md` (entity / no coverage / poor coverage / top 5 / concrete test additions)
- Kind filled; `ec_deployment` resource and data source not mixed
- Implementation directory filled
- Excluded steps not counted as coverage; TF_ACC gate skips still counted
- Data sources and `RequiresReplace` fields: no `no-update` tag
- Empty-collection gap omitted when `SizeAtLeast(1)` (or equivalent) forbids empty
- `timeouts` omitted unless in the schema map
- Import-verify did not empty the no-coverage list
- `close_via` on every gap
- `ec_deployment` **resource** paths use single-nested form; **data source** paths stay indexed
- JSON fence matches the markdown list (no extra fields)
