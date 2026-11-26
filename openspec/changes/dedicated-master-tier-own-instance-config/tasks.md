## 1. Thread private state and migration flag into node counting

- [ ] 1.1 In `UpdateDedicatedMasterTier` (`ec/ecresource/deploymentresource/topology_dedicated_masters.go`),
      read `migrate_to_latest_hardware` from `plan` once (same pattern as `enableMasterTier`) and
      pass it, together with the existing `privateState` parameter, into `countNodesInCluster`.

## 2. Resolve each tier's own instance configuration

- [ ] 2.1 Add a per-tier instance-configuration resolver (e.g. `resolveTierInstanceConfiguration`)
      that: returns the template's instance configuration for the tier when
      `migrate_to_latest_hardware` is `true`; otherwise looks up the tier's instance configuration
      via `getInstanceConfiguration` against the private-state instance-configuration list; falls
      back to the template's instance configuration for the tier id when the lookup finds nothing
      (no private state yet, tier not found, unknown ID/version).
- [ ] 2.2 In `countNodesInCluster`, read the private-state instance configurations once per call
      (reusing `ReadPrivateStateInstanceConfigurations`, logging and falling back to template-only
      behavior for all tiers on a read error, as `selectInstanceConfigurationToUse` already does),
      then call the new resolver per tier instead of unconditionally calling
      `getTemplateInstanceConfiguration`.
- [ ] 2.3 Use the resolved per-tier instance configuration for both `getMaxSize` and
      `convertSizeToInstanceConfigResource`, and remove the `tierUsesInstanceConfig` guard and
      function now that the conversion always uses the tier's own resolved instance configuration.

## 3. Unit tests

- [ ] 3.1 Add `TestUpdateDedicatedMasterTier` cases, each with a tier pinned (ID + version) to an
      instance configuration in private state that differs from the deployment template's:
      - Memory tier on a configuration with a smaller max node size than the template's enables
        the master tier (the undercount example from the issue).
      - Storage tier on a configuration whose multiplier matches its own max size correctly does
        **not** enable the master tier, where using the template's multiplier instead would have
        (the overcount example from the issue, replacing any existing `tierUsesInstanceConfig`
        regression test for this path).
      - Storage tier whose own multiplier still crosses the dedicated-masters threshold enables
        the master tier.
- [ ] 3.2 Add/extend `TestUpdateDedicatedMasterTier` cases for the fallback paths: resource
      creation with no private-state instance configurations yet, and
      `migrate_to_latest_hardware = true` with a tier pinned to a non-template instance
      configuration — both SHALL size against the template's instance configuration for the tier.
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
