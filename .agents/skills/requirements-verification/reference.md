# Requirements verification — consistency, mapping, tests

## Consistency checks

### Identity and Import

Identity `id` format must match Import’s accepted id. Imported value must be what Create would store.

### Schema vs requirements

- “When [attribute] is configured” ⇒ that attribute is optional or optional+computed in the schema, not required-only.
- “The resource SHALL set [attribute] in state” ⇒ computed or optional+computed.
- Type-gated validation (“when type is X”) ⇒ `type` and the gated fields exist in the schema.

### Lifecycle

No pair that says attribute X both RequiresReplace and updates in place.

### Client

Create/Update/Read/Delete (or data-source Read) must agree on Stateful vs Serverless. Do not mix `cloud-sdk-go` and the generated serverless client for one type unless the code actually does.

### State and Plan/State

For one attribute, do not require both “preserve null” and “store empty list” when the API returns empty. `UseStateForUnknown` / defaults should match optional+computed or computed fields.

### API

Create/Update reference the APIs the code calls. Read and Delete match GET/DELETE (and association teardown if specified).

### StateUpgrade

- A requirement that the resource “supports upgrading prior state schema version N to M” implies `Schema.Version` ≥ M and an `UpgradeState` upgrader for N→M.
- If the registered schema has a non-zero `Version` and there is **no** `UpgradeState`, a spec that claims upgrades work is inconsistent with the code. The current `ec_deployment` contract is: Version 2, no upgrader, re-import (README). Acc `TestAcc…_UpgradeFrom0_4_1` exist and are skipped — do not treat them as coverage.
- Do not confuse those tests with `TestAccDeployment_post_node_roles` / pre-node-role migration (stack version / `node_roles`, not TF state).

## Requirement → implementation mapping

| Category | Typical location | What to check |
| --- | --- | --- |
| **Type / client** | `schema.go` `Metadata` / `Configure` | TypeName; `ConvertProviderData`; Stateful vs Serverless. |
| **Unconfigured client** | `resourceReady` / equivalent in schema or resource file | Diagnostic summary `Unconfigured API Client`; no API call. |
| **Identity** | `create.go` (id set); plan modifier on `id` | API id stored; kept on later plans. |
| **Import** | `ImportState` in `schema.go` | Passthrough or custom; error on bad id if validated. |
| **Lifecycle** | `schema.go` plan modifiers | `RequiresReplace` vs in-place. |
| **Create/Update** | `create.go`, `update.go` | API call; read-after-mutate; missing-after error text. |
| **Read** | `read.go` | GET; not-found removes state; other errors keep state. |
| **Delete** | `delete.go` | Associations; not-found as success when required. |
| **Validation** | `ValidateConfig`, schema validators | Type-gated attributes; skip when unknown. |
| **Mapping** | Functions that flatten API → state and expand plan → API. Names vary: `expanders.go` / `flatteners.go` / `modelToState` on small resources; `ReadDeployment`, `ReadElasticsearches`, `ElasticsearchPayload` (and siblings) under `deploymentresource/`. Search for the role, not the filename. | Empty string → null; omit unknown nested ids; which planned fields are sent. |
| **Plan/State** | plan modifiers in `schema.go` | Defaults; unknown nested computed ids when the set changes. |
| **StateUpgrade** | Registered `schema.Schema` `Version`; `UpgradeState()` if implemented | Version vs upgraders; if Version is set and there is no upgrader, status is **Not met** / known gap (not “category does not apply”). |

Data sources: Read, Type/client, Unconfigured client, Mapping, Plan/State only. No StateUpgrade.

## Test opportunity patterns

### Unit (no live API) — agent may run `make unit`

| Kind | Example | Verifies |
| --- | --- | --- |
| Expand/flatten | Empty API description → null in state | Mapping |
| ValidateConfig | Remote-cluster fields on `type=ip` | Type-gated validation |
| Import | ImportState sets id | Identity/Import |
| Unconfigured client | CRUD with nil client → diagnostic | Ready guard |
| Plan modifiers | Default `include_by_default` false | Plan/State |
| StateUpgrade | Table-driven prior-state JSON → upgraded state / error | Only if `UpgradeState` exists. Do not write one to “cover” `ec_deployment`. |

### Acceptance (live Elastic Cloud API) — human / Buildkite only

Tests live in `ec/acc/` as `TestAcc…`. **Never** run them from this skill.

| Kind | Example | Verifies |
| --- | --- | --- |
| CRUD round-trip | `TestAccDeploymentTrafficFilter_basic` | Create/read/update/destroy |
| Type variant | `TestAccDeploymentTrafficFilter_remoteCluster` | Type-gated payload |
| Import | Acc ImportState step | Import + read |
| State upgrade from an old provider | `TestAcc…_UpgradeFrom0_4_1` (0.4.1 then `PlanOnly`) | TF state load after schema bump. **Skipped** until `ec_deployment` `UpgradeState` exists — say skipped, do not un-skip from this skill. |

Suggest: requirement heading, type (unit / acceptance), description, who runs (agent vs human/Buildkite). If `ec/acc/` already covers the case, say so rather than inventing a duplicate.
