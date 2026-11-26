## Why

`ec_deployment`'s automatic dedicated-master-tier management (`UpdateDedicatedMasterTier`)
decides whether a plan needs a dedicated master tier by counting the nodes each data/coordinating
tier would have. That count always uses the **deployment template's** instance configuration for
the tier — its largest discrete node size and its storage multiplier — even when the tier is
actually pinned to (or currently running on) a different instance configuration ID or version,
for example one kept after a deployment-template change.

Because the template's instance configuration can allow a different maximum node size or a
different storage multiplier than the tier's real instance configuration, the node count can be
wrong in either direction:

- **Undercount**: the tier's real instance configuration allows fewer resources per node than the
  template's, so the real deployment needs more nodes per zone than the template-based count
  implies. The provider can plan removal of a dedicated master tier that is still required.
- **Overcount**: for `size_resource = "storage"` tiers, the fix for #1108 only converts the
  storage size to memory when the tier runs on the template's instance configuration
  (`tierUsesInstanceConfig`); otherwise it skips the conversion and treats the raw storage GB
  figure as memory GB, inflating the apparent node count. The provider can add a dedicated master
  tier that isn't needed.

Both cases produce an incorrect Terraform plan for `master.size` / `master.zone_count` that does
not reflect what the Elastic Cloud API would actually require, which is a regression-prone
diagnosis surface (see #1015, #1108).

## What Changes

- `countNodesInCluster` looks up **each tier's own instance configuration** — from the private
  state list of instance configurations used by the deployment, keyed by instance configuration
  ID and version — instead of always using the deployment template's instance configuration for
  that tier.
- The tier's own instance configuration (when found) is used for both the largest discrete node
  size (`getMaxSize`) and the storage-to-memory multiplier (`convertSizeToInstanceConfigResource`)
  for that tier's node-count calculation.
- `countNodesInCluster` falls back to the deployment template's instance configuration for a tier
  when: there is no private state yet (resource creation), the tier's instance configuration
  cannot be found in private state, the tier's instance configuration ID or version is unknown
  (new tier, deployment-template change in progress), or `migrate_to_latest_hardware` is `true`
  (the tier is moving to the template's latest instance configuration).
- Because the per-tier instance configuration is now always resolved from the right source, the
  storage-size conversion no longer needs to be skipped when the tier is not on the template's
  instance configuration; `tierUsesInstanceConfig` becomes unnecessary for this path and is
  removed.
- `UpdateDedicatedMasterTier` threads `privateState` and the plan's `migrate_to_latest_hardware`
  value into `countNodesInCluster` so it can do this per-tier lookup the same way
  `enableMasterTier` already does for the master tier itself.

This only changes a `terraform plan`-time computation used to decide `master.size` /
`master.zone_count`; it does not change the Elastic Cloud API payload format or any other
attribute.

## Out of Scope

- Reusing `selectInstanceConfigurationToUse` as-is for data/coordinating tiers: it hardcodes the
  `"master"` tier id and returns `nil` when `DiscreteSizes` is `nil`, which is not the fallback
  behavior `countNodesInCluster` needs (it must still fall back to the template, not bail out,
  when a tier has no discrete sizes).
- Any change to how the master tier itself is sized once enabled (`enableMasterTier`,
  `selectInstanceConfigurationToUse`) — only the *decision* of whether to enable or disable it is
  in scope.
- Any change to `ec_deployment`'s serverless project counterparts; dedicated master management is
  specific to hosted Elasticsearch deployments.

## Capabilities

### New Capabilities

- `deployment-dedicated-master-tier`: Automatic management of the `ec_deployment` Elasticsearch
  dedicated master tier during plan.

### Modified Capabilities

## Impact

- Affects `ec/ecresource/deploymentresource/topology_dedicated_masters.go`
  (`UpdateDedicatedMasterTier`, `countNodesInCluster`, `tierUsesInstanceConfig` removal) and its
  unit tests (`TestUpdateDedicatedMasterTier`).
- No Terraform schema, state-storage, or API-payload changes; only the plan-time node-count
  heuristic changes.
- No Terraform type is new — `ec_deployment`'s existing automatic dedicated-master behavior is
  made more accurate.
