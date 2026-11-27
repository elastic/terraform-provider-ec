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
  cannot be resolved (create, unknown ID or version, not found in private state), so the heuristic
  never regresses to failing outright or crashing on a `nil` instance configuration.
- Keep the `tierUsesInstanceConfig` storage-conversion skip on the fallback path only, where the
  template's multiplier may not be the tier's.

**Non-Goals:**

- Changing how the master tier is sized once the decision to enable it is made
  (`enableMasterTier` / `selectInstanceConfigurationToUse` stay as they are).
- Introducing new Terraform schema attributes or changing `PrivateState` storage format.
- Handling serverless projects; dedicated master orchestration only applies to
  `ec_deployment` Elasticsearch topologies.

## Decisions

- **Thread `privateState` into `countNodesInCluster`.** `UpdateDedicatedMasterTier` already
  receives `privateState` (passed to `enableMasterTier`); pass it down into `countNodesInCluster`,
  which decodes it and passes the instance configurations to a new per-tier resolver.
  - *Why not thread `migrate_to_latest_hardware` too?* It is not needed to pick the right instance
    configuration. When a migration applies, the topology plan modifier (`useTopologyState`)
    leaves `instance_configuration_id` unknown, so the resolver already falls back to the
    template. A tier with `instance_configuration_id` or `instance_configuration_version` set in
    its configuration is never migrated (`planmodifiers.CheckAvailableMigration`), so it stays
    on its own instance configuration and must be sized against it, not the template's.
  - *Why not read private state once per tier inside the loop?* `ReadPrivateStateInstanceConfigurations`
    deserializes the full private-state list; call it once per `countNodesInCluster` invocation
    (not once per tier) and pass the decoded slice into the per-tier resolver, mirroring how
    `selectInstanceConfigurationToUse` already does one decode for the master tier.
- **New `resolveTierInstanceConfiguration` helper** (name chosen to avoid colliding with
  `selectInstanceConfigurationToUse`, which stays master-tier-specific and keeps its own
  `nil`-on-no-`DiscreteSizes` behavior for `enableMasterTier`'s needs). It returns the instance
  configuration and whether it was resolved from private state:
  1. If the tier's `instance_configuration_version` is unknown, skip the lookup and fall back
     (step 3). `getInstanceConfiguration` only checks for an unknown or null ID; for an unknown
     version `ValueInt64()` returns 0, which can match an unrelated instance configuration with
     `ConfigVersion` 0 in private state. This happens when the ID is set in configuration and the
     version is left unknown (new tier, template change).
  2. Otherwise, call `getInstanceConfiguration(ctx, rawTopology, instanceConfigurations)` for the
     tier. If it returns a non-nil result with non-nil `DiscreteSizes`, use it (resolved from
     private state).
  3. Otherwise (no private state yet on create, tier not found, ID or version unknown, or the
     found instance configuration has no `DiscreteSizes`), fall back to the template's instance
     configuration for the tier id, matching current behavior (not resolved from private state).
  - Unlike `selectInstanceConfigurationToUse`, this resolver does not special-case `"master"` and
    does not bail out: when the tier's own instance configuration is unusable it falls back to the
    template's. That can still be nil when the template defines none for the tier;
    `getMaxSize`, `convertSizeToInstanceConfigResource` and `tierUsesInstanceConfig` already
    handle nil (one node per zone, no conversion), as today.
  - *Why fall back when `DiscreteSizes` is nil?* `getMaxSize` returns 0 for it, and the counting
    loop then counts one node per zone whatever the size, which can undercount and remove a master
    tier that is still needed. The template fallback keeps today's count; if the tier is not on
    the template's instance configuration, the `tierUsesInstanceConfig` guard below then skips the
    storage conversion, so the count errs high.
- **Use the resolved instance configuration for both `getMaxSize` and
  `convertSizeToInstanceConfigResource`.** When it was resolved from private state, convert
  unconditionally: it is the tier's own instance configuration, so its multiplier is the right
  one. When it is the template fallback, keep the existing `tierUsesInstanceConfig` guard: if the
  tier's ID or version differs from the template's (for example an ID that has no match in private
  state, or a null/unknown ID with a version pinned to a different version than the template's),
  the template's multiplier is not the tier's, so the conversion is skipped and the count errs
  high, keeping the master tier as #1108 intended.
- **Read `privateState` once per `countNodesInCluster` call, reusing the existing
  `ReadPrivateStateInstanceConfigurations` + diagnostics pattern.** On a read error, log at debug
  (as `selectInstanceConfigurationToUse` does) and fall back to the template's instance
  configuration for every tier in that call, rather than failing the plan.

## Risks / Trade-offs

- A private-state read failure now affects every tier's node count, not just the master tier's
  sizing → mitigated by falling back to today's template-only behavior on that path, so the worst
  case is the previously-existing heuristic, not a new failure mode.
- On the fallback path, a tier whose ID or version differs from the template's is still counted
  without the storage conversion, so it can overcount and keep a master tier that is not needed →
  accepted: this is today's behavior, and erring high avoids removing a master tier that is still
  needed. Converting with the template's multiplier there could undercount whenever that multiplier
  is larger than the tier's real one.

## Migration Plan

No state migration. This only changes a `ModifyPlan`-time calculation; existing `ec_deployment`
state and prior applies are unaffected. Rollback is reverting the code change — there is no schema
or private-state format to unwind.
