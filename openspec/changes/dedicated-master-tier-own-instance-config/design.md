## Context

See `proposal.md` for motivation and
`specs/deployment-dedicated-master-tier/spec.md` for the behavior contract. This is entirely a
Terraform Plugin Framework `ModifyPlan` concern on `ec_deployment` (this provider is 100% Plugin
Framework, no SDKv2); it does not call the Elastic Cloud API. The provider has already fetched the
data it needs before `UpdateDedicatedMasterTier` runs:

- The deployment template (`*models.DeploymentTemplateInfoV2`, from `cloud-sdk-go`), via
  `loadTemplate()`.
- The list of instance configurations the deployment actually uses, via
  `ReadPrivateStateInstanceConfigurations(ctx, privateState)` — private Terraform state populated
  from `cloud-sdk-go` responses on prior apply, keyed by instance configuration ID and
  `ConfigVersion`.

`getInstanceConfiguration(ctx, rawTopology, deploymentInstanceConfigs)` already does the ID+version
lookup against that list for the master tier (`selectInstanceConfigurationToUse`,
`topology_dedicated_masters.go:296`). The relevant discussion is
[#1108 PR review](https://github.com/elastic/terraform-provider-ec/pull/1108#discussion_r4503127521)
and issues #1015 / #1108.

## Goals / Non-Goals

**Goals:**

- Base each data/coordinating tier's node-count contribution on *that tier's own* instance
  configuration (max discrete node size + storage multiplier) when it is known, instead of always
  on the deployment template's instance configuration for that tier id.
- Preserve today's safe fallback to the template's instance configuration when the tier's own one
  cannot be resolved (create, unknown/new, `migrate_to_latest_hardware`), so the heuristic never
  regresses to failing outright or crashing on a `nil` instance configuration.
- Remove the now-redundant `tierUsesInstanceConfig` storage-conversion skip once every tier
  resolves its own instance configuration directly.

**Non-Goals:**

- Changing how the master tier is sized once the decision to enable it is made
  (`enableMasterTier` / `selectInstanceConfigurationToUse` stay as they are).
- Introducing new Terraform schema attributes or changing `PrivateState` storage format.
- Handling serverless projects; dedicated master orchestration only applies to
  `ec_deployment` Elasticsearch topologies.

## Decisions

- **Thread `privateState` and `migrateToLatestHw` into `countNodesInCluster`.**
  `UpdateDedicatedMasterTier` already reads `privateState` (passed to `enableMasterTier`) and
  `plan`'s `migrate_to_latest_hardware` is already read inside `enableMasterTier`. Read
  `migrate_to_latest_hardware` once in `UpdateDedicatedMasterTier` (via
  `plan.GetAttribute(ctx, path.Root("migrate_to_latest_hardware"), &migrateToLatestHw)`, the same
  pattern `enableMasterTier` uses) and pass both `privateState` and that bool down into
  `countNodesInCluster`, which in turn passes them to a new per-tier instance-configuration
  resolver.
  - *Why not read private state once per tier inside the loop?* `ReadPrivateStateInstanceConfigurations`
    deserializes the full private-state list; call it once per `countNodesInCluster` invocation
    (not once per tier) and pass the decoded slice into the per-tier resolver, mirroring how
    `selectInstanceConfigurationToUse` already does one decode for the master tier.
- **New `resolveTierInstanceConfiguration` helper** (name chosen to avoid colliding with
  `selectInstanceConfigurationToUse`, which stays master-tier-specific and keeps its own
  `nil`-on-no-`DiscreteSizes` behavior for `enableMasterTier`'s needs):
  1. If `migrateToLatestHw` is `true`, return the template's instance configuration for the tier
     (`getTemplateInstanceConfiguration(template, topologyId)`) — the tier is moving to the
     template's latest configuration, so that is the correct one to size against.
  2. Otherwise, call `getInstanceConfiguration(ctx, rawTopology, instanceConfigurations)` for the
     tier. If it returns a non-nil result, use it.
  3. Otherwise (no private state yet on create, tier not found, ID/version unknown), fall back to
     the template's instance configuration for the tier id, matching current behavior.
  - Unlike `selectInstanceConfigurationToUse`, this resolver does not special-case `"master"` and
    does not discard a found instance configuration merely because `DiscreteSizes` is `nil` —
    `getMaxSize` and `convertSizeToInstanceConfigResource` already treat a `nil`/zero
    `DiscreteSizes` or `StorageMultiplier` as "no adjustment", so the caller does not need the
    early-exit `selectInstanceConfigurationToUse` relies on for the master-tier enable path.
- **Use the resolved instance configuration for both `getMaxSize` and
  `convertSizeToInstanceConfigResource`**, unconditionally — drop the
  `tierUsesInstanceConfig` guard around the storage conversion. Once the instance configuration
  comes from the tier's own ID/version (or a deliberate template fallback), its storage multiplier
  is always the right one to convert with, so the "skip conversion, overcount instead" safety net
  from #1108 is no longer needed and `tierUsesInstanceConfig` is deleted.
- **Read `privateState` once per `countNodesInCluster` call, reusing the existing
  `ReadPrivateStateInstanceConfigurations` + diagnostics pattern.** On a read error, log at debug
  (as `selectInstanceConfigurationToUse` does) and fall back to the template's instance
  configuration for every tier in that call, rather than failing the plan.

## Risks / Trade-offs

- A private-state read failure now affects every tier's node count, not just the master tier's
  sizing → mitigated by falling back to today's template-only behavior on that path, so the worst
  case is the previously-existing heuristic, not a new failure mode.
- Removing `tierUsesInstanceConfig` changes behavior for any tier that previously benefited from
  the "skip storage conversion" safety net *and* still has no resolvable own instance
  configuration (template fallback) → the fallback path still resolves to the template's own
  instance configuration, so the conversion runs against the template's correct multiplier in that
  case too; the net effect is strictly more accurate, not less.

## Migration Plan

No state migration. This only changes a `ModifyPlan`-time calculation; existing `ec_deployment`
state and prior applies are unaffected. Rollback is reverting the code change — there is no schema
or private-state format to unwind.
