## 1. Thread private state into node counting

- [ ] 1.1 In `UpdateDedicatedMasterTier` (`ec/ecresource/deploymentresource/topology_dedicated_masters.go`),
      pass the existing `privateState` parameter into `countNodesInCluster`.
      `migrate_to_latest_hardware` is not threaded (see design.md).

## 2. Resolve each tier's own instance configuration

- [ ] 2.1 Add a per-tier instance-configuration resolver (e.g. `resolveTierInstanceConfiguration`)
      that returns the instance configuration and whether it was resolved from private state. It
      skips the lookup when the tier's `instance_configuration_version` is unknown; otherwise it
      looks up the tier's instance configuration via `getInstanceConfiguration` against the
      private-state instance-configuration list; it falls back to the template's instance
      configuration for the tier id when the lookup is skipped or finds nothing (no private state
      yet, tier not found, unknown ID or version), or when the found instance configuration has
      nil `DiscreteSizes`.
- [ ] 2.2 In `countNodesInCluster`, read the private-state instance configurations once per call
      (reusing `ReadPrivateStateInstanceConfigurations`, logging and falling back to template-only
      behavior for all tiers on a read error, as `selectInstanceConfigurationToUse` already does),
      then call the new resolver per tier instead of unconditionally calling
      `getTemplateInstanceConfiguration`.
- [ ] 2.3 Use the resolved per-tier instance configuration for both `getMaxSize` and
      `convertSizeToInstanceConfigResource`. Convert unconditionally when it was resolved from
      private state; on the template fallback, keep the existing `tierUsesInstanceConfig` guard
      around the conversion.

## 3. Unit tests

- [ ] 3.1 Add `TestUpdateDedicatedMasterTier` cases, each with a tier pinned (ID + version) to an
      instance configuration in private state that differs from the deployment template's:
      - Memory tier on a configuration with a smaller max node size than the template's enables
        the master tier (the undercount example from the issue).
      - Storage tier on a configuration with multiplier 24 and a 4g maximum, with a template
        multiplier of 8, does **not** enable the master tier (3 nodes), where using the template's
        multiplier instead would have (9 nodes).
      - The existing `Should not convert storage based sizes when …` cases have no private-state
        instance configuration for the hot tier, so they exercise the template fallback and keep
        their current expectations.
      - Storage tier whose own multiplier still crosses the dedicated-masters threshold enables
        the master tier.
- [ ] 3.2 Add/extend `TestUpdateDedicatedMasterTier` cases for the fallback paths:
      - Resource creation with no private-state instance configurations yet sizes against the
        template's instance configuration.
      - A tier whose ID is set and whose version is unknown falls back to the template, even when
        private state has an instance configuration with that ID and `ConfigVersion` 0.
      - A storage tier whose ID is known, differs from the template's, and is missing from private
        state is counted without conversion and enables the master tier.
      - A tier whose instance configuration is in private state but has nil `DiscreteSizes` is
        sized against the template's maximum node size instead of one node per zone: for example
        8g memory with a 4g template maximum across 3 zones counts 6 nodes and enables the master
        tier, where one node per zone would count 3 and not.
      - `migrate_to_latest_hardware = true` with a migrating tier whose
        `instance_configuration_id` and `instance_configuration_version` are unknown (as the
        topology plan modifier leaves them) sizes against the template's instance configuration,
        even though private state still holds the tier's old instance configuration.
      - `migrate_to_latest_hardware = true` with a tier pinned to a non-template instance
        configuration that is in private state sizes against the tier's own instance
        configuration.
      - The test harness builds the plan from `es.Elasticsearch` structs, where nil becomes null,
        not unknown. Set unknown values explicitly (e.g. `types.StringUnknown()` /
        `types.Int64Unknown()` via `plan.SetAttribute`) for the unknown-version and migrating-tier
        cases, so they do not silently test null.
- [ ] 3.3 Run `env -u TF_ACC go test ./ec/ecresource/deploymentresource/...` and
      `env -u TF_ACC make unit` to confirm the new and existing cases pass.

## 4. Validation

- [ ] 4.1 Run `env -u TF_ACC make lint`.
- [ ] 4.2 Run `env -u TF_ACC make check-openspec` and fix any reported issues.

## 5. Out-of-band human / Buildkite verification

- [ ] 5.1 A human runs the relevant `TestAcc…` dedicated-master acceptance case(s) for
      `ec_deployment` locally (not run by an agent, not run with `TF_ACC` set in this workflow)
      before merge, and the full acceptance suite runs on Buildkite
      (`buildkite/terraform-provider-ec-acceptance`) as a required check on `master`.
