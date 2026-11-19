# Existing entity requirements — code-path checklist

Use this checklist so the spec is complete and traceable to code. Package layout: [`coding-standards.md`](../../../dev-docs/high-level/coding-standards.md) and [`repo-structure.md`](../../../dev-docs/high-level/repo-structure.md).

## 1. Locate the implementation

| Entity type | Where to look |
| --- | --- |
| Resource | `ec/ecresource/<name>resource/` — typically `schema.go`, `create.go`, `read.go`, `update.go`, `delete.go`, plus the functions that map API ↔ Terraform (see §8). Larger packages split models (`projectresource/`, `deploymentresource/`). |
| Data source | `ec/ecdatasource/<name>datasource/` — `datasource.go` or `schema.go` + `read.go`. Exceptions exist (`ec/ecdatasource/deploymenttemplates/`); when the glob fails, use the `ec/provider.go` registry. |

**Type name**: `Metadata()` sets `response.TypeName = request.ProviderTypeName + "_..."` (sometimes `fmt.Sprintf`). Use that for the H1 title and HCL `resource` / `data` type.

**Registry**: `ec/provider.go` `Resources()` / `DataSources()` is the live list. One factory may return several types from one package.

**Implementation path**: the Go package used on the `Resource implementation:` / `Data source implementation:` line (e.g. `ec/ecresource/trafficfilterresource`).

**Multi-type packages** (spec one Terraform type, not the package):

| Package | Types |
| --- | --- |
| `ec/ecresource/projectresource` | `ec_elasticsearch_project`, `ec_observability_project`, `ec_security_project` |
| `ec/ecdatasource/privatelinkdatasource` | `ec_aws_privatelink_endpoint`, `ec_gcp_private_service_connect_endpoint`, `ec_azure_privatelink_endpoint` |
| `ec/ecresource/deploymentresource` | `ec_deployment` only — **do not spec as a whole**; name a slice |

## 2. Schema

| What to capture | Where |
| --- | --- |
| Attributes and blocks | `Schema()` in `schema.go` (or generated `*_resource_gen.go` under `ec/internal/gen/serverless/` for serverless types, plus wrappers). |
| Required / optional / computed | `Required`, `Optional`, `Computed`. Optional+computed = both. |
| Types | `StringAttribute`, `BoolAttribute`, `SetNestedBlock`, `ListNestedAttribute`, etc. |
| Plan modifiers | `UseStateForUnknown`, `RequiresReplace`, `BoolDefaultValue`, resource-local modifiers. |
| Validators | Schema `Validators` and `ValidateConfig`. |
| Defaults | Plan modifiers such as `planmodifiers.BoolDefaultValue(false)`. |
| Schema version | `schema.Schema{ Version: … }`. Unset or `0` is normal. A non-zero version without `UpgradeState` is a **known gap**, not “no StateUpgrade category”. |

Output: HCL with `<required|optional|optional+computed|computed>`, type, short notes.

## 3. Identity and import (resources)

| What to capture | Where |
| --- | --- |
| `id` format | Set in Create from the API; often `UseStateForUnknown()` so later plans keep it. |
| Import | `ImportState` (often `ImportStatePassthroughID`). Document the accepted id and that Read fills the rest. |

## 4. CRUD and API

Create and Update **read after the mutative call** (coding standard): persist state from GET, not from the create/update response alone. Some resources retry that GET via `util.ReadAfterMutate.UntilFound` (wired from `ProviderClients.ReadAfterMutate`) and may treat HTTP 403 as missing. Capture retry/403 behavior when the code has it.

| Operation | Where | What to capture |
| --- | --- | --- |
| Create | `create.go` | API (`cloud-sdk-go` package or serverless client method), expand plan → request, set `id`, internal `read`, error if missing after create. |
| Read | `read.go` | GET by id; not-found → remove from state with no error; other errors surfaced, state unchanged. |
| Update | `update.go` | Update API + read-after-update; include known nested ids when the code does. |
| Delete | `delete.go` | Identifier from state; association teardown if any; not-found treated as success when the code does that. |

One requirement: non-success API responses (except documented not-found cases) become diagnostics.

## 5. Client

| What to capture | Where |
| --- | --- |
| Configure | `internal.ConvertProviderData` → `ProviderClients.Stateful` (`*api.API` / cloud-sdk-go) or `.Serverless`. Optional `.ReadAfterMutate` for GET retries after create/update. |
| Ready guard | Missing client → error summary `Unconfigured API Client`, no API call. |

There is **no** resource-level `elasticsearch_connection` override. The provider config is the connection.

## 6. Lifecycle (resources)

`RequiresReplace()` on attributes → when X changes, the resource is replaced. Otherwise updates are in place.

## 7. State upgrade (resources)

| What to capture | Where |
| --- | --- |
| Schema version | `Version` on the registered `schema.Schema` (the one `Schema()` returns). Go package names `deployment/v1` vs `v2` are not the same thing unless that schema is actually registered. |
| Upgraders | `UpgradeState()` / `ResourceWithUpgradeState`. Per-version logic and upgrade-failure diagnostics. |

Most types have no `Version` and no upgrader — skip the category.

`ec_deployment` is the exception: live schema is `Version: 2` (`deployment/v2/schema.go`); `deployment/v1` still has `Version: 1`; the resource does **not** implement `UpgradeState`. README / changelog say users should re-import. Acc `TestAcc…_UpgradeFrom0_4_1` cases apply with published `elastic/ec` 0.4.1 then `PlanOnly` with current — all `t.Skip("skip until ec_deployment state upgrade is implemented")`. That is the TF **state** upgrade path. Tests named `post_node_roles` / `pre_node_roles_migration` bump Elastic Stack / `node_roles` on a live deployment; they are not `UpgradeState`.

If you spec a **slice** of `ec_deployment`, do not invent an upgrader. Record the gap in **Known gaps**: schema version 2, no `UpgradeState`, skipped `UpgradeFrom0_4_1` acc, import is the documented workaround.

## 8. Mapping (config ↔ API ↔ state)

| What to capture | Where |
| --- | --- |
| API → state (read) | Whatever function copies a backend response into the Terraform model. Small resources often use `flatteners.go` / `modelToState`. `ec_deployment` uses names like `ReadDeployment`, `ReadElasticsearches`. |
| Plan → API (write) | Whatever function builds the request payload. Small resources often use `expanders.go`. `ec_deployment` uses names like `ElasticsearchPayload`. |
| Unknown in plan | Plan modifiers that keep prior state or mark nested computed ids unknown when the set changes. |

Do not require the filenames `expanders.go` / `flatteners.go`. Search for the role.

## 9. Data sources only

Read only. Required lookup arguments vs computed results. Same unconfigured-client guard. No Create/Update/Delete, Import, RequiresReplace, or StateUpgrade.

## Requirement categories

Use these names in headings when they fit. Do **not** add stack-provider categories this repo does not have (resource-level connection override, Elasticsearch/Kibana server-version gates). Keep **StateUpgrade**: look for it; most types have none, `ec_deployment` has a version and no upgrader.

| Category | Use for |
| --- | --- |
| **Type / client** | Type name; Stateful vs Serverless client. |
| **Unconfigured client** | `Unconfigured API Client` on every CRUD/Read. |
| **Identity** | Computed `id` format. |
| **Import** | Import id and follow-up Read. |
| **Lifecycle** | RequiresReplace. |
| **Create/Update** | Mutate then read; missing-after-mutate errors. |
| **Read** | GET, not-found, other errors. |
| **Delete** | Destroy, associations, not-found-as-success. |
| **Validation** | `ValidateConfig` / schema validators, including type-gated fields. |
| **Mapping** | Expand/flatten, empty vs null. |
| **Plan/State** | Defaults, UseStateForUnknown, unknown nested ids. |
| **StateUpgrade** | Schema `Version` and `UpgradeState` (or the known gap when Version is set and there is no upgrader). |

## File layout

- Path: `openspec/specs/<capability>/spec.md`
- Capability id: kebab-case without `ec_`; slices of `ec_deployment` are named for the slice (`deployment-integrations-server`), never `deployment`.
- One spec directory per Terraform type.
