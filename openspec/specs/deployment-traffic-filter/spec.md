# `ec_deployment_traffic_filter`

Resource implementation: `ec/ecresource/trafficfilterresource`

## Purpose

Define schema and behavior for the Elastic Cloud traffic filter **resource** `ec_deployment_traffic_filter` (capability id `deployment-traffic-filter`): ruleset identity and import, type-gated rule validation, CRUD against the hosted Elastic Cloud traffic-filter API (`cloud-sdk-go` `trafficfilterapi`), create/update-then-read, unconfigured-client guard, and mapping between Terraform state and a traffic filter ruleset.

**In scope:** this resource only — attributes, `rule` blocks, validation, lifecycle, and import-by-id.

**Out of scope:**

- Data source `ec_traffic_filter` (different Terraform type name)
- Resource `ec_deployment_traffic_filter_association`
- The `traffic_filter` argument on `ec_deployment`
- Attaching or detaching a ruleset from a deployment except for the association teardown this resource performs on destroy

## Schema

```hcl
resource "ec_deployment_traffic_filter" "example" {
  id                 = <computed, string>  # Elastic Cloud ruleset id; UseStateForUnknown
  name               = <required, string>
  type               = <required, string>  # ip | vpce | azure_private_endpoint | gcp_private_service_connect_endpoint | remote_cluster
  region             = <required, string>
  include_by_default = <optional+computed, bool>  # default false
  description        = <optional, string>

  rule {  # set nested block; at least one required
    source                = <optional, string>  # IP, CIDR, VPC endpoint id, or GCP PSC id; documented as required except when type is azure_private_endpoint
    description           = <optional, string>
    azure_endpoint_name   = <optional, string>  # only when type is azure_private_endpoint
    azure_endpoint_guid   = <optional, string>  # only when type is azure_private_endpoint
    remote_cluster_id     = <optional, string>  # required when type is remote_cluster
    remote_cluster_org_id = <optional, string>  # required when type is remote_cluster
    id                    = <computed, string>  # API rule id; unknown when rules change; do not UseStateForUnknown
  }
}
```

## Requirements

### Requirement: Resource type and API client

The resource type name SHALL be `<provider>_deployment_traffic_filter` (with the default provider name, `ec_deployment_traffic_filter`).

The resource SHALL use the configured stateful Elastic Cloud API client (`cloud-sdk-go`) for all CRUD calls. It SHALL NOT use the serverless generated client.

#### Scenario: Type name

- GIVEN the provider type name `ec`
- WHEN the resource metadata is registered
- THEN the provider SHALL expose the resource as `ec_deployment_traffic_filter`

### Requirement: Unconfigured API client

Every Create, Read, Update, and Delete call SHALL require a non-nil API client. When the client is missing, the resource SHALL add an error diagnostic with summary `Unconfigured API Client` and SHALL NOT call the traffic-filter API.

#### Scenario: CRUD with no client

- GIVEN the resource has not been configured with an API client
- WHEN Create, Read, Update, or Delete is invoked
- THEN the provider SHALL surface an `Unconfigured API Client` error
- AND SHALL NOT send a traffic-filter API request

### Requirement: Ruleset identity

The Terraform `id` SHALL be the Elastic Cloud traffic filter ruleset identifier returned by the create API. The attribute SHALL be computed and SHALL use `UseStateForUnknown` so the ruleset id is not shown as unknown on later plans.

#### Scenario: Id after create

- GIVEN a successful create whose API response includes ruleset id `abc123`
- WHEN state is persisted
- THEN the provider SHALL set `id` to `abc123`

### Requirement: Required ruleset attributes

`name`, `type`, and `region` SHALL be required. `description` SHALL be optional. Documented `type` values are `ip`, `vpce`, `azure_private_endpoint`, `gcp_private_service_connect_endpoint`, and `remote_cluster`.

`include_by_default` SHALL be optional and computed. When omitted, the plan SHALL default it to `false`.

#### Scenario: Omit include_by_default

- GIVEN a configuration that does not set `include_by_default`
- WHEN the configuration is planned
- THEN the provider SHALL plan `include_by_default` as `false`

#### Scenario: Set include_by_default true

- GIVEN a configuration with `include_by_default = true`
- WHEN Create is called
- THEN the provider SHALL send `include_by_default` as true in the create request

### Requirement: At least one rule

The `rule` block SHALL be a set of nested objects. The resource SHALL require at least one rule (`setvalidator.SizeAtLeast(1)`).

#### Scenario: No rule blocks

- GIVEN a configuration with `name`, `type`, and `region` but no `rule` blocks
- WHEN the configuration is validated
- THEN the provider SHALL reject the configuration because the rule set size is below 1

#### Scenario: One rule accepted

- GIVEN a configuration with a single valid `rule` block
- WHEN the configuration is validated
- THEN the provider SHALL accept the rule set

### Requirement: Computed rule id

Each rule's `id` SHALL be computed from the API (the Elastic Cloud rule identifier). The attribute SHALL NOT use `UseStateForUnknown`. When the planned set of rules differs from state, each rule `id` SHALL be unknown in the plan (`StringIsUnknownIfRulesChange`).

Because unknown rule ids are omitted from the update payload, an apply that changes rules MAY receive new rule ids from the API.

#### Scenario: Rule ids after create

- GIVEN a newly created ruleset whose GET response includes rule ids
- WHEN Create finishes
- THEN the provider SHALL store those rule `id` values in state

#### Scenario: Rules change marks rule ids unknown

- GIVEN an existing ruleset in state with computed rule ids
- AND the configuration adds, removes, or replaces a `rule` block
- WHEN a plan is computed
- THEN the provider SHALL set each rule `id` to unknown in the plan

### Requirement: Type-gated rule validation

`ValidateConfig` SHALL type-gate rule attributes. It SHALL skip that gating when `type` or `rule` is unknown (for example when those values come from variables that are not yet known).

When `type` is not `remote_cluster`, a rule SHALL NOT set `remote_cluster_id` or `remote_cluster_org_id`. When `type` is `remote_cluster` and both values are known, both `remote_cluster_id` and `remote_cluster_org_id` SHALL be set; otherwise the resource SHALL add a `Missing Required Attributes` error.

When `type` is not `azure_private_endpoint`, a rule SHALL NOT set `azure_endpoint_name` or `azure_endpoint_guid`.

The resource SHALL NOT Terraform-validate that `source` is present or absent for a given `type`; `source` is optional in the schema and is documented as required except when `type` is `azure_private_endpoint`. Azure endpoint fields are optional in the schema even when `type` is `azure_private_endpoint`.

#### Scenario: Skip gating while type is unknown

- GIVEN `type` is unknown at validate time
- WHEN `ValidateConfig` runs
- THEN the provider SHALL NOT emit type-gating diagnostics for rule attributes

#### Scenario: Remote cluster fields on an IP ruleset

- GIVEN `type = "ip"`
- AND a `rule` sets `remote_cluster_id` and/or `remote_cluster_org_id`
- WHEN the configuration is validated
- THEN the provider SHALL add an `Invalid Rule Configuration` error stating those attributes can only be specified when `type` is `remote_cluster`

#### Scenario: Remote cluster type missing a field

- GIVEN `type = "remote_cluster"`
- AND a `rule` sets `remote_cluster_id` but omits `remote_cluster_org_id`
- WHEN the configuration is validated
- THEN the provider SHALL add a `Missing Required Attributes` error stating both `remote_cluster_id` and `remote_cluster_org_id` are required

#### Scenario: Azure fields on an IP ruleset

- GIVEN `type = "ip"`
- AND a `rule` sets `azure_endpoint_name` and/or `azure_endpoint_guid`
- WHEN the configuration is validated
- THEN the provider SHALL add an `Invalid Rule Configuration` error stating those attributes can only be specified when `type` is `azure_private_endpoint`

#### Scenario: Azure private endpoint rule

- GIVEN `type = "azure_private_endpoint"`
- AND a `rule` sets `azure_endpoint_name` and `azure_endpoint_guid`
- WHEN the configuration is validated
- THEN the provider SHALL accept the configuration

#### Scenario: Remote cluster rule

- GIVEN `type = "remote_cluster"`
- AND a `rule` sets both `remote_cluster_id` and `remote_cluster_org_id`
- WHEN the configuration is validated
- THEN the provider SHALL accept the configuration

### Requirement: Create then read

Create SHALL expand the plan into a `TrafficFilterRulesetRequest` and call `trafficfilterapi.Create`. On API error, the resource SHALL surface the error and SHALL NOT persist state.

On success, the resource SHALL set `id` from the create response, then SHALL read the ruleset back (`trafficfilterapi.Get` with `include_associations=false`) and persist state from that read. If the ruleset is not found after create, the resource SHALL add the error `Failed to read deployment traffic filter ruleset after create.`, SHALL remove the resource from state, and SHALL NOT leave a partial resource.

#### Scenario: Create IP ruleset

- GIVEN a plan with `type = "ip"`, a name, a region, and one or more `rule` blocks with `source` set
- WHEN Create is called
- THEN the provider SHALL call `trafficfilterapi.Create` with that name, type, region, and rules
- AND SHALL GET the ruleset by the returned id
- AND SHALL persist attributes from the GET response

#### Scenario: Create remote cluster ruleset

- GIVEN a plan with `type = "remote_cluster"` and a rule with `remote_cluster_id` and `remote_cluster_org_id`
- WHEN Create is called
- THEN the provider SHALL include `remote_cluster_id` and `remote_cluster_org_id` on the rule in the create request

#### Scenario: Create Azure private endpoint ruleset

- GIVEN a plan with `type = "azure_private_endpoint"` and a rule with `azure_endpoint_name` and `azure_endpoint_guid`
- WHEN Create is called
- THEN the provider SHALL include those Azure fields on the rule in the create request

#### Scenario: Create API error

- GIVEN the create API returns an error
- WHEN Create is called
- THEN the provider SHALL surface the error
- AND SHALL NOT persist state

#### Scenario: Missing after create

- GIVEN create succeeds
- AND the subsequent GET reports the ruleset as not found
- WHEN Create finishes
- THEN the provider SHALL surface `Failed to read deployment traffic filter ruleset after create.`
- AND SHALL remove the resource from state

### Requirement: Read

Read SHALL call `trafficfilterapi.Get` with the state `id` and `include_associations=false`.

When the API reports the ruleset as not found, the resource SHALL be removed from state with no error. Any other API error SHALL be surfaced and state SHALL be left unchanged.

#### Scenario: Read existing ruleset

- GIVEN a ruleset exists in Elastic Cloud
- WHEN Read is called
- THEN the provider SHALL map name, type, region, description, `include_by_default`, and rules into state

#### Scenario: Read missing ruleset

- GIVEN the ruleset was deleted out of band
- WHEN Read is called
- THEN the provider SHALL remove the resource from state with no error

#### Scenario: Read API error

- GIVEN GET returns a non-not-found error
- WHEN Read is called
- THEN the provider SHALL surface the error
- AND SHALL NOT remove the resource from state as not found

### Requirement: Update then read

Update SHALL expand the plan into a `TrafficFilterRulesetRequest` and call `trafficfilterapi.Update` with the existing ruleset `id`. Known (non-null, non-unknown) rule `id` values SHALL be included on the corresponding rules in the request.

On API error, the resource SHALL surface the error. On success, the resource SHALL read the ruleset back and persist state from that read. If the ruleset is not found after update, the resource SHALL add the error `Failed to read deployment traffic filter ruleset after update.` and SHALL remove the resource from state.

#### Scenario: Update include_by_default

- GIVEN a ruleset exists
- AND the plan changes `include_by_default`
- WHEN Update is called
- THEN the provider SHALL call `trafficfilterapi.Update` with the planned ruleset
- AND SHALL GET the ruleset
- AND SHALL persist attributes from the GET response

#### Scenario: Update sends known rule ids

- GIVEN state has rule ids assigned by the API
- AND the planned rules are unchanged enough that those ids remain known
- WHEN Update is called
- THEN the provider SHALL include those rule ids in the update request

#### Scenario: Missing after update

- GIVEN update succeeds
- AND the subsequent GET reports the ruleset as not found
- WHEN Update finishes
- THEN the provider SHALL surface `Failed to read deployment traffic filter ruleset after update.`
- AND SHALL remove the resource from state

### Requirement: Delete disassociates then deletes

Delete SHALL GET the ruleset with `include_associations=true`. If that GET reports not found, delete SHALL succeed without further API calls.

For each association on the ruleset, Delete SHALL call `trafficfilterapi.DeleteAssociation`. A not-found error on an association SHALL be ignored; any other association-delete error SHALL be surfaced and SHALL stop deletion.

After associations are cleared, Delete SHALL call `trafficfilterapi.Delete`. A not-found error on the ruleset delete SHALL be treated as success. Any other delete error SHALL be surfaced.

#### Scenario: Delete ruleset with no associations

- GIVEN a ruleset exists with no associations
- WHEN Destroy is called
- THEN the provider SHALL GET the ruleset with associations
- AND SHALL call `trafficfilterapi.Delete`
- AND SHALL remove the resource from state

#### Scenario: Delete ruleset with associations

- GIVEN a ruleset is associated with one or more deployments
- WHEN Destroy is called
- THEN the provider SHALL delete each association
- AND SHALL then delete the ruleset

#### Scenario: Association already gone

- GIVEN an association delete returns not found
- WHEN Destroy is running
- THEN the provider SHALL ignore that error and continue deleting remaining associations and the ruleset

#### Scenario: Ruleset already gone on GET

- GIVEN GET-with-associations reports the ruleset as not found
- WHEN Destroy is called
- THEN the provider SHALL treat delete as success with no error

#### Scenario: Ruleset already gone on DELETE

- GIVEN GET succeeds
- AND `trafficfilterapi.Delete` reports not found
- WHEN Destroy is called
- THEN the provider SHALL treat delete as success with no error

#### Scenario: Association delete fails

- GIVEN an association delete returns a non-not-found error
- WHEN Destroy is called
- THEN the provider SHALL surface the error
- AND SHALL NOT proceed to delete the ruleset

### Requirement: Import by id

Import SHALL accept the Elastic Cloud ruleset id as the import ID and SHALL set Terraform `id` to that value. A subsequent Read SHALL populate the remaining attributes from the API.

#### Scenario: Import existing ruleset

- GIVEN a ruleset exists with id `320b7b540dfc967a7a649c18e2fce4ed`
- WHEN `terraform import ec_deployment_traffic_filter.name 320b7b540dfc967a7a649c18e2fce4ed` is run
- THEN the provider SHALL set `id` to `320b7b540dfc967a7a649c18e2fce4ed`
- AND SHALL read the ruleset and populate name, type, region, description, `include_by_default`, and rules

### Requirement: State mapping

Expand SHALL send `name`, `type`, `region`, `include_by_default`, and every planned rule. A rule `source`, `description`, Azure endpoint fields, and remote-cluster fields SHALL be sent when they are known and non-null. A known non-null rule `id` SHALL be sent; unknown or null rule ids SHALL be omitted.

Flatten SHALL map empty API strings for ruleset `description` and for each rule's `source`, `description`, Azure endpoint fields, and remote-cluster fields to null in Terraform state (not empty strings). Non-empty API values SHALL be stored as strings. Rule `id` SHALL be stored from the API rule id.

#### Scenario: Empty ruleset description

- GIVEN the API returns an empty ruleset description
- WHEN Read maps the response
- THEN `description` in state SHALL be null

#### Scenario: Flatten remote cluster rule

- GIVEN the API returns a rule with `remote_cluster_id` and `remote_cluster_org_id` and no `source`
- WHEN Read maps the response
- THEN the rule in state SHALL have those remote-cluster attributes set
- AND `source` SHALL be null

#### Scenario: Flatten Azure private endpoint rule

- GIVEN the API returns a rule with `azure_endpoint_name` and `azure_endpoint_guid` and no `source`
- WHEN Read maps the response
- THEN the rule in state SHALL have those Azure attributes set
- AND `source` SHALL be null

#### Scenario: Flatten IP rule

- GIVEN the API returns a rule with `source` `1.1.1.1` and a rule id
- WHEN Read maps the response
- THEN the rule in state SHALL have `source` `1.1.1.1` and that `id`
- AND Azure and remote-cluster attributes SHALL be null
