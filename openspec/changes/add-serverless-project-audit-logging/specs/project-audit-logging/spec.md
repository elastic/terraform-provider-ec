# `project-audit-logging`

Audit log delivery on the serverless project resources. Resource implementations: `ec/ecresource/projectresource`.

## Purpose

Configure audit log delivery on Elastic Cloud serverless projects: enable or pause delivery, choose the destination project, and attach ignore-filter references.

**In scope:** the optional `monitoring.logging.audit` configuration on `ec_elasticsearch_project`, `ec_observability_project`, and `ec_security_project`.

**Out of scope:** hosted `ec_deployment`, any VectorDB project resource, the ignore-filter catalog (the monitoring API's list/get endpoints), and log categories other than `audit`.

## Schema

```hcl
resource "ec_elasticsearch_project" "example" {
  # same nested attributes on ec_observability_project and ec_security_project

  monitoring = { # optional; omit to leave audit logging unconfigured
    logging = { # optional
      audit = { # optional; omit to clear that category
        enabled = <optional+computed, bool> # default true when audit is set
        destination = { # required when audit is set
          project_id   = <required, string>
          project_type = <optional+computed, string> # elasticsearch | observability | security | vectordb
          status       = <computed, string> # enabled | suspended | deleted; not sent on write
        }
        ignore_filter_ids = <optional, set of string> # at most 10; null and empty are both absence
      }
    }
  }
}
```

## ADDED Requirements

### Requirement: Audit configuration on serverless project resources

`ec_elasticsearch_project`, `ec_observability_project`, and `ec_security_project` SHALL each accept the same optional `monitoring.logging.audit` nested attribute. The provider SHALL use the generated serverless project API client for create, read, and update of that attribute. Hosted deployment resources SHALL NOT gain it.

#### Scenario: Configuration on each serverless project type

- GIVEN a configuration for `ec_elasticsearch_project`, `ec_observability_project`, or `ec_security_project`
- WHEN `monitoring.logging.audit` is set
- THEN the provider SHALL accept that attribute on that resource

#### Scenario: Omitted configuration

- GIVEN a serverless project resource configuration with no `monitoring` attribute
- WHEN the resource is created
- THEN the provider SHALL NOT send `monitoring` on the create request
- AND SHALL store `monitoring` as null

### Requirement: Audit attributes

When `monitoring.logging.audit` is set, `destination.project_id` SHALL be required. `enabled` SHALL be optional and computed, and SHALL default to `true` when `audit` is set and `enabled` is omitted. `destination.project_type` SHALL be optional and computed. When a practitioner sets it, `project_type` SHALL be one of `elasticsearch`, `observability`, `security`, or `vectordb`. `destination` SHALL be required when `audit` is set. `ignore_filter_ids` SHALL be an optional set of ignore-filter id strings. Null and an empty set SHALL both mean no filters. The project API limits `ignore_filters` to 10 items (`maxItems`); the provider SHALL reject a longer set. The API field is an array of objects with an `id`; the provider SHALL map each Terraform string to one object and SHALL NOT expose that object as a nested attribute. `destination.status` SHALL be computed. The provider SHALL NOT send `destination.status` on create or update.

#### Scenario: Defaults inside audit

- GIVEN `monitoring.logging.audit` with `destination.project_id` set and `enabled` omitted
- WHEN the plan is built
- THEN the provider SHALL plan `enabled` as `true`

#### Scenario: Too many ignore filters

- GIVEN `ignore_filter_ids` with 11 ids
- WHEN the plan is validated
- THEN the provider SHALL return an error diagnostic
- AND SHALL NOT call the project API

### Requirement: Create sends the audit category

When `audit` is set, create SHALL send `monitoring.logging.audit` with the configured destination project id, `enabled`, and `project_type` only when the planned value is known. When `audit` is null, including `monitoring = {}` and `monitoring = { logging = {} }`, create SHALL omit `monitoring`. Each `ignore_filter_ids` entry SHALL be sent as an `ignore_filters` object `{ "id": "<id>" }`. When `ignore_filter_ids` is null or empty, create SHALL omit `ignore_filters`. Create SHALL NOT send `destination.status`.

#### Scenario: Create with destination and filters

- GIVEN `enabled` false, `destination.project_id` `proj-1`, `destination.project_type` `observability`, and `ignore_filter_ids` `["audit-ignore-data-reads"]`
- WHEN the project is created
- THEN the create request JSON SHALL set `monitoring.logging.audit.enabled` to false
- AND SHALL set `destination.project_id` to `proj-1` and `destination.project_type` to `observability`
- AND SHALL set `ignore_filters` to `[{"id":"audit-ignore-data-reads"}]`
- AND SHALL omit destination status

#### Scenario: Empty monitoring shell on create

- GIVEN `monitoring = { logging = {} }` with no `audit`
- WHEN the project is created
- THEN the create request JSON SHALL NOT contain `monitoring`

### Requirement: Read stores audit configuration and destination status

After a successful read, when the API returns `monitoring.logging.audit`, state SHALL reflect its `enabled`, `destination.project_id`, and `destination.project_type`. When the API returns that object and the plan or state `audit` is null, whether `monitoring` is null or a nested object with `audit` null, the provider SHALL store the API audit configuration, including on import and on refresh of logging configured outside Terraform. When the API `ignore_filters` array has ids, `ignore_filter_ids` SHALL be the set of those ids. When that array is absent or empty, the provider SHALL keep the null or empty set already in the plan or state; when the plan or state has no `audit`, or has a non-empty `ignore_filter_ids`, `ignore_filter_ids` SHALL be null. `destination.status` SHALL be the API destination `status` (`enabled`, `suspended`, or `deleted`). When the API response has no `monitoring.logging.audit` and the plan or state `audit` is already null, the provider SHALL keep the plan or state `monitoring` value unchanged, including a nested object whose `audit` is null. When the API response has no `monitoring.logging.audit` and `monitoring` in the plan or state is null, the provider SHALL store `monitoring` as null. When the plan or state has `audit` set and the API response has none, the provider SHALL store `monitoring` as null; on the read that completes create or update, Terraform then reports the dropped write as an inconsistent result.

#### Scenario: Read fills destination status

- GIVEN an existing project whose API audit destination status is `suspended`
- WHEN the resource is read
- THEN the provider SHALL set `destination.status` to `suspended`

#### Scenario: Read stores project type from the API

- GIVEN configuration that omits `destination.project_type`
- AND the API read returns `destination.project_type` `elasticsearch`
- WHEN the resource is read
- THEN the provider SHALL set `destination.project_type` to `elasticsearch`

#### Scenario: Read with no monitoring in plan or state

- GIVEN an API response with no `monitoring.logging.audit`
- AND the plan or state `monitoring` is null
- WHEN the resource is read
- THEN the provider SHALL store `monitoring` as null

#### Scenario: Empty monitoring parents stay as configured

- GIVEN the plan or state has `monitoring` set and `audit` null
- AND the API response has no `monitoring.logging.audit`
- WHEN the resource is read
- THEN the provider SHALL keep that `monitoring` value
- AND SHALL NOT store `monitoring` as null

#### Scenario: Refresh adopts audit configured outside Terraform

- GIVEN the plan or state `monitoring` is null
- AND the API response includes `monitoring.logging.audit` with no `ignore_filters`
- WHEN the resource is read
- THEN the provider SHALL store that API audit configuration
- AND SHALL store `ignore_filter_ids` as null

#### Scenario: Empty shell does not hide API audit

- GIVEN the plan or state has `monitoring` set and `audit` null
- AND the API response includes `monitoring.logging.audit`
- WHEN the resource is read
- THEN the provider SHALL store that API audit configuration
- AND SHALL NOT keep `audit` null

#### Scenario: Read drops audit the API does not have

- GIVEN the plan or prior state has `audit` set
- AND the API response has no `monitoring.logging.audit`
- WHEN the resource is read
- THEN the provider SHALL store `monitoring` as null

#### Scenario: Empty ignore filters keep the configured absence

- GIVEN an API audit object whose `ignore_filters` is absent or `[]`
- AND the plan or state has `ignore_filter_ids` null
- WHEN the resource is read
- THEN the provider SHALL store `ignore_filter_ids` as null

#### Scenario: Empty set stays an empty set

- GIVEN an API audit object whose `ignore_filters` is absent or `[]`
- AND the plan or state has `ignore_filter_ids` as an empty set
- WHEN the resource is read
- THEN the provider SHALL store `ignore_filter_ids` as an empty set

#### Scenario: API dropped the configured filters

- GIVEN an API audit object whose `ignore_filters` is absent or `[]`
- AND the plan or state has a non-empty `ignore_filter_ids`
- WHEN the resource is read
- THEN the provider SHALL store `ignore_filter_ids` as null

### Requirement: Update patches the audit category

When `audit` is set in the plan, including when it is added to a project that had none, update SHALL patch `monitoring.logging.audit` with the configured `enabled`, destination project id, `project_type` only when that planned value is known, and ignore-filter ids. Replacing a non-empty `ignore_filter_ids` with null or an empty set SHALL send `ignore_filters` as JSON null, which is how the project API removes all filters for the category. When `ignore_filter_ids` is null or empty in both the plan and the state, update SHALL omit `ignore_filters`. Update SHALL NOT send `destination.status`. `enabled` false SHALL pause delivery and SHALL keep the destination and filters that are still configured. When `audit` is unset in both configuration and state, update SHALL NOT send `monitoring`.

#### Scenario: Pause delivery

- GIVEN a project with audit delivery enabled
- WHEN `enabled` is changed to false and the destination is unchanged
- THEN the patch JSON SHALL set `monitoring.logging.audit.enabled` to false
- AND SHALL still send the configured destination project id

#### Scenario: Clear ignore filters

- GIVEN a project whose state has ignore-filter ids
- WHEN `audit` remains and `ignore_filter_ids` is null or an empty set
- THEN the patch JSON SHALL contain `"ignore_filters":null`

#### Scenario: No filters in plan or state omits the field

- GIVEN `ignore_filter_ids` is null or an empty set in both the plan and the state
- WHEN `audit` remains configured
- THEN the patch JSON SHALL omit `ignore_filters`

#### Scenario: Add audit on update

- GIVEN a project with no audit configuration in state
- WHEN `audit` is set in the plan
- THEN the patch JSON SHALL contain `monitoring.logging.audit` with the configured destination project id
- AND SHALL NOT contain `"audit":null`

#### Scenario: Unknown project type is omitted

- GIVEN an update whose planned `destination.project_type` is unknown
- WHEN the patch is built
- THEN the patch JSON SHALL omit `destination.project_type`

#### Scenario: Update with audit never configured

- GIVEN a project with no audit configuration in state or config
- WHEN another attribute is updated
- THEN the patch JSON SHALL NOT contain `monitoring`

### Requirement: Removing audit clears that category only

Removing `monitoring.logging.audit` from configuration SHALL send a patch whose JSON contains `"audit":null` and SHALL NOT contain `"monitoring":null`. A patch body that omits the `audit` key leaves the category unchanged on the API, so the provider SHALL send an explicit `"audit":null` rather than omitting the key. Other monitoring fields SHALL be left unchanged.

#### Scenario: Remove audit

- GIVEN a project with audit logging configured
- WHEN `audit` is removed from configuration
- THEN the patch JSON SHALL contain `"audit":null`
- AND SHALL NOT contain `"monitoring":null`
