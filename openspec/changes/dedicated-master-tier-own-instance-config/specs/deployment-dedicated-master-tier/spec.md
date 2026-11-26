## Purpose

Defines how `ec_deployment` decides, during `terraform plan`, whether an Elasticsearch deployment
needs an automatically managed dedicated master tier, based on how many nodes the planned
data/coordinating topologies would require.

## ADDED Requirements

### Requirement: Node counting uses each tier's own instance configuration

When automatic dedicated-master-tier management counts the nodes a planned Elasticsearch tier
(hot content, coordinating, warm, cold, or frozen) would require, the provider SHALL compute the
tier's maximum node size and storage-to-memory conversion from the instance configuration the
tier itself runs on (resolved by instance configuration ID and version from the deployment's known
instance configurations), not unconditionally from the deployment template's instance
configuration for that tier.

#### Scenario: Tier pinned to a smaller-capacity instance configuration undercounts without the fix

- **GIVEN** a deployment template whose hot content tier's instance configuration allows up to 4g
  per node and a dedicated-masters-threshold of 6
- **AND** the hot content tier is actually running on a different instance configuration that
  allows at most 1g per node
- **WHEN** the plan requests a 4g memory hot content tier across 3 zones
- **THEN** the provider SHALL count that tier as 4 nodes per zone (12 nodes total) using the tier's
  own 1g-max instance configuration, and SHALL plan a dedicated master tier because the count
  reaches the threshold

#### Scenario: Storage-sized tier on a different multiplier is not overcounted

- **GIVEN** a deployment template whose hot content tier's instance configuration has a storage
  multiplier of 24 and a dedicated-masters-threshold of 6
- **AND** the hot content tier is actually running on a different instance configuration with a
  storage multiplier of 24 and a 4g per-node maximum
- **WHEN** the plan requests a 96g storage hot content tier across 3 zones
- **THEN** the provider SHALL convert the 96g storage size using the tier's own instance
  configuration's multiplier (96g storage / 24 = 4g memory), count 1 node per zone (3 nodes
  total), and SHALL NOT plan a dedicated master tier because the count stays below the threshold

#### Scenario: Storage-sized tier whose own multiplier still requires masters

- **GIVEN** a deployment template with a dedicated-masters-threshold of 6
- **AND** a tier running on an instance configuration with a storage multiplier of 8 and a 4g
  per-node maximum
- **WHEN** the plan requests a 96g storage tier across 3 zones
- **THEN** the provider SHALL convert using the tier's own multiplier (96g storage / 8 = 12g
  memory), count 3 nodes per zone (9 nodes total), and SHALL plan a dedicated master tier because
  the count reaches the threshold

### Requirement: Node counting falls back to the template's instance configuration

When a tier's own instance configuration cannot be determined, the provider SHALL use the
deployment template's instance configuration for that tier to compute its node count, so the
dedicated-master-tier decision never fails or silently skips a tier.

#### Scenario: Resource creation has no prior instance configuration state

- **GIVEN** a new `ec_deployment` resource being created, with no instance configurations yet
  recorded for the deployment
- **WHEN** the provider counts nodes for a planned tier during `terraform plan`
- **THEN** the provider SHALL use the deployment template's instance configuration for that tier

#### Scenario: Tier's instance configuration is unresolvable

- **GIVEN** a planned tier whose instance configuration ID or version is unknown (for example, a
  newly added tier, or a deployment template change in progress) or does not match any instance
  configuration known for the deployment
- **WHEN** the provider counts nodes for that tier during `terraform plan`
- **THEN** the provider SHALL use the deployment template's instance configuration for that tier

#### Scenario: Migrating to the latest hardware uses the template's configuration

- **GIVEN** a plan with `migrate_to_latest_hardware` set to `true`
- **WHEN** the provider counts nodes for a planned tier that will move to the deployment
  template's latest instance configuration
- **THEN** the provider SHALL use the deployment template's instance configuration for that tier
