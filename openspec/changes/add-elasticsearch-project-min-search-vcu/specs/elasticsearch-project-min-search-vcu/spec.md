# `elasticsearch-project-min-search-vcu`

The Minimum Search VCU floor on `ec_elasticsearch_project`. Resource implementation: `ec/ecresource/projectresource`.

## Purpose

Let practitioners pin a lower bound on Search tier autoscaling for an Elastic Cloud serverless Elasticsearch project, change it, and remove it, through the `search_lake.min_search_vcu` field of the serverless project API.

**In scope:** the optional `search_lake.min_search_vcu` attribute on `ec_elasticsearch_project`: schema and validation, create, read and import, update, and unset.

**Out of scope:** `ec_observability_project` (no `search_lake` attribute) and `ec_security_project` (its `search_lake` has only `data_retention`), hosted `ec_deployment`, the sibling `search_power` and `boost_window` attributes beyond how they interact with unsetting the floor, and client-side enforcement of the API's upper bound.

## Schema

```hcl
resource "ec_elasticsearch_project" "example" {
  name      = "my-project"
  region_id = "aws-us-east-1"

  search_lake = { # optional+computed; omit to leave the whole block, floor included, unmanaged
    search_power   = <optional+computed, number> # unchanged
    boost_window   = <optional+computed, number> # unchanged
    min_search_vcu = <optional, number>          # whole number >= 2; omit or null to unset the floor
  }
}
```

## ADDED Requirements

### Requirement: Minimum Search VCU attribute

`ec_elasticsearch_project` SHALL accept an optional integer attribute `min_search_vcu` inside its `search_lake` nested attribute. The attribute SHALL be optional and SHALL NOT be computed, so that a configuration without it plans the floor as null. The provider SHALL use the generated serverless project API client for create, read, and update of that attribute. `ec_observability_project`, `ec_security_project`, and `ec_deployment` SHALL NOT gain it.

#### Scenario: Attribute accepted on the Elasticsearch project resource

- GIVEN an `ec_elasticsearch_project` configuration whose `search_lake` block sets `min_search_vcu = 64`
- WHEN the configuration is validated
- THEN the provider SHALL accept it without a diagnostic

#### Scenario: Attribute absent on other project resources

- GIVEN the schemas of `ec_observability_project` and `ec_security_project`
- WHEN they are inspected
- THEN neither SHALL contain a `min_search_vcu` attribute

### Requirement: Lower bound validated at plan time, API maximum left to the API

The provider SHALL reject `min_search_vcu` values below 2 with an error diagnostic at plan time and SHALL NOT call the project API for such a configuration. The provider SHALL reject values above 2147483647 with an error diagnostic at plan time, because a larger value cannot be sent unchanged by 32-bit builds. Below that, the provider SHALL NOT enforce the API's maximum: values above the API's current maximum (4992 in the API specification, or a lower limit the API applies to an organization) SHALL be sent to the project API, and the API's rejection SHALL be returned to the practitioner as an error diagnostic at apply time that includes the API's status and body.

#### Scenario: Value below the minimum

- GIVEN `min_search_vcu = 1`
- WHEN the plan is validated
- THEN the provider SHALL return an error diagnostic
- AND SHALL NOT call the project API

#### Scenario: Value too large for a 32-bit build

- GIVEN `min_search_vcu = 4294967298`
- WHEN the plan is validated
- THEN the provider SHALL return an error diagnostic
- AND SHALL NOT call the project API

#### Scenario: Value above the API maximum

- GIVEN `min_search_vcu = 5000` on an organization whose maximum is 4992
- WHEN the project is created or updated
- THEN the provider SHALL send the value to the project API
- AND SHALL return an error diagnostic that includes the API's status and response body

### Requirement: Create sends the configured floor

When `search_lake.min_search_vcu` is set, create SHALL send `search_lake.min_search_vcu` with that value alongside any configured `search_power` and `boost_window`. When `min_search_vcu` is null, create SHALL omit the `min_search_vcu` key, so the project is created with no floor and the API applies no default.

#### Scenario: Create with a floor

- GIVEN `search_lake = { search_power = 100, min_search_vcu = 64 }`
- WHEN the project is created
- THEN the create request JSON SHALL contain `search_lake.search_power` 100 and `search_lake.min_search_vcu` 64

#### Scenario: Create without a floor

- GIVEN `search_lake = { search_power = 100 }`
- WHEN the project is created
- THEN the create request JSON SHALL contain `search_lake.search_power` 100
- AND SHALL NOT contain the key `min_search_vcu`

### Requirement: Read stores the floor the API reports

After a successful read, including import and refresh, `search_lake.min_search_vcu` in state SHALL be the value the API returns in `search_lake.min_search_vcu`, and SHALL be null when the API response has no `min_search_vcu` key. `search_power` and `boost_window` SHALL continue to be stored from the same response.

#### Scenario: Read with a floor

- GIVEN an existing project whose API response has `search_lake.min_search_vcu` 64
- WHEN the resource is read
- THEN the provider SHALL store `search_lake.min_search_vcu` as 64

#### Scenario: Read without a floor

- GIVEN an existing project whose API response has `search_lake` without a `min_search_vcu` key
- WHEN the resource is read
- THEN the provider SHALL store `search_lake.min_search_vcu` as null

#### Scenario: Import adopts a floor configured outside Terraform

- GIVEN a project whose floor was set in the Cloud UI and no prior Terraform state
- WHEN the resource is imported
- THEN the provider SHALL store `search_lake.min_search_vcu` as the API value

### Requirement: Update patches the floor

When the planned `search_lake.min_search_vcu` is non-null, update SHALL send `search_lake.min_search_vcu` with the planned value, whether the floor is new or changed. The patch SHALL include only the `search_lake` keys the plan knows: an unknown or null sibling SHALL be omitted so the API leaves it unchanged. When `min_search_vcu` is null in both the plan and the prior state, update SHALL NOT send the `min_search_vcu` key.

#### Scenario: Raise the floor

- GIVEN state `search_lake.min_search_vcu` 64 and plan 128
- WHEN the project is updated
- THEN the patch request JSON SHALL contain `search_lake.min_search_vcu` 128

#### Scenario: Add a floor to a project that had none

- GIVEN state `search_lake.min_search_vcu` null and plan 64
- WHEN the project is updated
- THEN the patch request JSON SHALL contain `search_lake.min_search_vcu` 64
- AND SHALL NOT contain `"min_search_vcu":null`

#### Scenario: Floor never configured

- GIVEN `search_lake.min_search_vcu` null in both plan and state
- WHEN another attribute is updated
- THEN the patch request JSON SHALL NOT contain the key `min_search_vcu`

### Requirement: Removing the floor sends an explicit null

When the prior state has a non-null `search_lake.min_search_vcu` and the plan has `search_lake` set with `min_search_vcu` null — because the practitioner removed the line or set it to `null` — update SHALL send a patch whose JSON contains `"search_lake":{...,"min_search_vcu":null}`. The patch SHALL NOT send `"search_lake":null`, which would also reset `search_power` and `boost_window` to their defaults, and SHALL NOT merely omit the key, which the API treats as unchanged. Other `search_lake` keys the plan knows SHALL be sent in the same request with their planned values.

#### Scenario: Remove the floor, keep siblings

- GIVEN state `search_lake = { search_power = 100, boost_window = 7, min_search_vcu = 64 }` and a configuration whose `search_lake` block sets `search_power = 100` and no `min_search_vcu`
- WHEN the project is updated
- THEN the patch request JSON SHALL contain `"min_search_vcu":null` inside `search_lake`
- AND SHALL contain `search_lake.search_power` 100
- AND SHALL NOT contain `"search_lake":null`

#### Scenario: Floor set outside Terraform is removed on apply

- GIVEN a configuration whose `search_lake` block has no `min_search_vcu`
- AND a refresh stored `search_lake.min_search_vcu` 64 that was set in the Cloud UI
- WHEN the practitioner plans and applies
- THEN the plan SHALL show `min_search_vcu` changing from 64 to null
- AND the apply SHALL send `"min_search_vcu":null` inside `search_lake`

### Requirement: Omitting the whole search_lake block leaves the floor unmanaged

When the configuration has no `search_lake` block and the prior state has a `search_lake` value, the provider SHALL plan `search_lake` as the prior state value, so the plan shows no change to `search_lake` whether or not the prior state holds a floor; it SHALL NOT send `search_lake` on update; and after read it SHALL store the `search_lake` values the API returns, including any `min_search_vcu` set outside Terraform. On create with no `search_lake` block the provider SHALL plan the block as unknown, as today. Practitioners who configure `search_lake` but want the floor managed elsewhere SHALL be able to use `lifecycle { ignore_changes = [search_lake.min_search_vcu] }`; the provider SHALL NOT need any special handling for that.

#### Scenario: No search_lake block with a floor set outside Terraform

- GIVEN a configuration with no `search_lake` block
- AND the prior state has `search_lake.min_search_vcu` 64, refreshed from a floor set in the Cloud UI
- WHEN the practitioner plans and applies a change to another attribute
- THEN the plan SHALL show no change to `search_lake`
- AND the patch request JSON SHALL NOT contain `search_lake`
- AND state SHALL store `search_lake.min_search_vcu` as 64 after the read

#### Scenario: Deleting the whole block keeps the floor

- GIVEN state `search_lake = { search_power = 100, boost_window = 7, min_search_vcu = 4 }` written by a configuration that set `min_search_vcu = 4`
- WHEN the practitioner deletes the whole `search_lake` block and plans
- THEN the plan SHALL show no change
- AND a later read SHALL still store `search_lake.min_search_vcu` as 4

#### Scenario: No search_lake block on create

- GIVEN a configuration with no `search_lake` block and no prior state
- WHEN the project is created
- THEN the create request JSON SHALL NOT contain `search_lake`
- AND state SHALL store the `search_lake` values the API returns, with `min_search_vcu` null

### Requirement: Documentation points at the public Minimum Search VCUs docs

The attribute description in the schema and the generated registry documentation SHALL describe `min_search_vcu` as the minimum number of search VCUs allocated to the project and SHALL link to the public Elastic Cloud serverless project settings documentation for the Minimum Search VCUs setting, rather than restating its pricing or inactivity-metering consequences. The documentation SHALL state that removing the attribute unsets the floor.

#### Scenario: Generated documentation

- GIVEN the regenerated `docs/resources/elasticsearch_project.md`
- WHEN a reader looks up `search_lake`
- THEN the nested schema SHALL list `min_search_vcu` as optional with a description that links to the public documentation
