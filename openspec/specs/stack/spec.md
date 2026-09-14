# `ec_stack` data source

Capability id: `stack`. Implementation: `ec/ecdatasource/stackdatasource`.

## Purpose

Capture the **current** behavior of the Terraform data source `ec_stack`, which looks up one Elastic Cloud stack pack (version plus per-kind metadata) for a region. This is a copyable canonical exemplar of behavior already in the provider, not a change proposal.

**In scope**

- Data source `ec_stack` only (Plugin Framework; stateful Elastic Cloud API).
- Lookup of a single stack from the region list: literal `latest` or a Go `version_regex`.
- The optional `lock` attribute as it actually behaves on Read (it does not pin a previously selected version).
- Flatten of computed stack metadata and per-kind config (denylist, capacity constraints, docker image, Elasticsearch plugins).
- Unconfigured-client guard and error diagnostics for list failure, no-match (regex path), and invalid regex.

**Out of scope**

- `ec_deployment` (or any other resource) version arguments.
- Changing the lookup algorithm (including any special-case for patch lines such as 7.17).
- Resources, serverless projects, or unimplemented fields (for example `node_types` on kind configs).

## Schema

```hcl
data "ec_stack" "example" {
  version_regex = <required, string>  # Go regexp, or the literal "latest"
  region        = <required, string>  # ESS region; ECE installations use "ece-region"
  lock          = <optional, bool>    # accepted in config; does not change lookup (see Requirements)

  # Computed
  id                  = <computed, string>         # same as version
  version             = <computed, string>
  accessible          = <computed, bool>           # ESS; not meaningful on ECE
  min_upgradable_from = <computed, string>
  upgradable_to       = <computed, list(string)>
  allowlisted         = <computed, bool>           # ESS; not meaningful on ECE

  # Computed lists, size at most 1
  apm = <computed, list(object)> [{
    denylist                 = <computed, list(string)>
    capacity_constraints_max = <computed, int64>
    capacity_constraints_min = <computed, int64>
    compatible_node_types    = <computed, list(string)>
    docker_image             = <computed, string>
  }]
  kibana            = <computed, list(object)>  # same object shape as apm
  enterprise_search = <computed, list(object)>  # same object shape as apm
  elasticsearch     = <computed, list(object)> [{
    denylist                 = <computed, list(string)>
    capacity_constraints_max = <computed, int64>
    capacity_constraints_min = <computed, int64>
    compatible_node_types    = <computed, list(string)>
    docker_image             = <computed, string>
    plugins                  = <computed, list(string)>
    default_plugins          = <computed, list(string)>
  }]
}
```

## Requirements

### Requirement: Data source schema

The provider SHALL register a data source of type `ec_stack`. `version_regex` and `region` SHALL be required strings. `lock` SHALL be an optional bool (unset is treated as false). The data source SHALL expose computed `id`, `version`, `accessible`, `min_upgradable_from`, `upgradable_to` (list of strings), `allowlisted`, and the nested lists `apm`, `kibana`, `enterprise_search`, and `elasticsearch`. Each nested list SHALL allow at most one object. Elasticsearch objects SHALL also include computed `plugins` and `default_plugins`; the other kinds SHALL NOT.

#### Scenario: Required version_regex enforced

- GIVEN a data source configuration that omits `version_regex`
- WHEN Terraform validates the configuration
- THEN Terraform SHALL reject the configuration with a validation error

#### Scenario: Required region enforced

- GIVEN a data source configuration that omits `region`
- WHEN Terraform validates the configuration
- THEN Terraform SHALL reject the configuration with a validation error

### Requirement: Unconfigured API client

On read, if the stateful Elastic Cloud API client has not been configured, the data source SHALL return an error diagnostic and SHALL NOT call the stack list API.

#### Scenario: Read with no client

- GIVEN the provider has not configured an API client on the data source
- WHEN the data source is read
- THEN the provider SHALL return an error diagnostic with summary `Unconfigured API Client`

### Requirement: List stacks for the configured region

The data source SHALL list stack packs for the configured `region` through the Elastic Cloud stack list API. The provider SHALL pass `region` through unchanged (Elastic Cloud Enterprise installations use `ece-region`) and SHALL NOT re-sort the returned list (the API returns latest-first). When the list call fails, the data source SHALL return an error diagnostic and SHALL NOT apply version filters.

#### Scenario: Successful list

- GIVEN a configured API client and `region = "us-east-1"`
- WHEN the data source is read
- THEN the provider SHALL list stacks for `us-east-1`

#### Scenario: ECE region passed through

- GIVEN a configured API client and `region = "ece-region"`
- WHEN the data source is read
- THEN the provider SHALL list stacks for `ece-region`

#### Scenario: List API failure

- GIVEN the stack list API returns an error
- WHEN the data source is read
- THEN the provider SHALL return an error diagnostic indicating it failed to retrieve the specified stack version

### Requirement: `latest` selects the first listed stack

When `version_regex` is the literal string `latest`, the data source SHALL return the first stack in the list.

#### Scenario: latest

- GIVEN the list `[7.9.1, 7.9.0, 7.8.1, 7.8.0]` (latest-first)
- AND `version_regex = "latest"`
- WHEN the data source resolves a stack
- THEN the provider SHALL select version `7.9.1`

### Requirement: `lock` does not change lookup on Read

`lock` SHALL be an optional configuration attribute. Read SHALL load **configuration only** (`version` is computed and is not in config), so a previously selected version is not available on a later Read. Therefore `lock = true` with `version_regex = "latest"` SHALL still select the first listed stack, the same as unlocked `latest`. `lock` SHALL NOT change lookup when `version_regex` is not `latest`.

#### Scenario: locked latest still takes the first stack

- GIVEN the list `[7.9.1, 7.9.0, 7.8.1, 7.8.0]`
- AND `version_regex = "latest"`
- AND `lock` is true
- AND a previous Read stored `version` `7.8.1` in state
- WHEN the data source is read again
- THEN the provider SHALL select version `7.9.1`

#### Scenario: lock ignored for a non-latest regex

- GIVEN the list `[7.9.1, 7.9.0, 7.8.1, 7.8.0]`
- AND `version_regex = "7.8.?"`
- AND `lock` is true
- WHEN the data source resolves a stack
- THEN the provider SHALL select version `7.8.1`

### Requirement: `version_regex` selects the first matching stack

When `version_regex` is not the literal `latest`, the provider SHALL compile `version_regex` as a Go regular expression and return the **first** stack whose `version` matches, in list order. Because the list is latest-first, the first match is the newest matching stack.

#### Scenario: Exact version string

- GIVEN the list `[7.9.1, 7.9.0, 7.8.1, 7.8.0]`
- AND `version_regex = "7.9.0"`
- WHEN the data source resolves a stack
- THEN the provider SHALL select version `7.9.0`

#### Scenario: Patch regex

- GIVEN the list `[7.9.1, 7.9.0, 7.8.1, 7.8.0]`
- AND `version_regex = "7.8.?"`
- WHEN the data source resolves a stack
- THEN the provider SHALL select version `7.8.1`

### Requirement: Lookup diagnostics

If `version_regex` is not valid Go regexp syntax, the data source SHALL return an error diagnostic that it failed to compile `version_regex`. If `version_regex` is not `latest` and no listed stack matches, the data source SHALL return an error diagnostic asking for a valid `version_regex` rather than producing empty state. These diagnostics do not apply to the `latest` path (see above).

#### Scenario: Invalid regex

- GIVEN `version_regex = "(?!"`
- WHEN the data source resolves a stack
- THEN the provider SHALL return an error diagnostic that it failed to compile `version_regex`

#### Scenario: No matching stack

- GIVEN the list `[{version: "7.8.0"}]`
- AND `version_regex = "7.9.1"`
- WHEN the data source resolves a stack
- THEN the provider SHALL return an error diagnostic matching `failed to obtain a stack version matching "7.9.1": please specify a valid version_regex`

### Requirement: Flatten selected stack into state

On a successful lookup, the data source SHALL write the selected stack into state: `id` and `version` SHALL equal the stack version string; `accessible`, `min_upgradable_from`, `upgradable_to`, and `allowlisted` SHALL be populated from the corresponding stack fields when present. Each of `apm`, `kibana`, `enterprise_search`, and `elasticsearch` SHALL be a list of one object when that kind has config (denylist from the API denylist/blacklist, capacity min/max, compatible node types, docker image; Elasticsearch also plugins and default plugins), and SHALL be empty/null when that kind is absent or empty.

#### Scenario: Metadata after a successful read

- GIVEN a selected stack with version `7.9.1`, accessible and allowlisted true, and `min_upgradable_from = "6.8.0"`
- WHEN the data source writes state
- THEN `id` and `version` SHALL equal `7.9.1`
- AND `accessible` and `allowlisted` SHALL be true
- AND `min_upgradable_from` SHALL equal `6.8.0`

#### Scenario: Elasticsearch kind config

- GIVEN the selected stack includes Elasticsearch denylist, capacity constraints, docker image, plugins, and default plugins
- WHEN the data source writes state
- THEN `elasticsearch` SHALL contain one object with those fields populated (`denylist`, `capacity_constraints_max`, `capacity_constraints_min`, `docker_image`, `plugins`, `default_plugins`)

#### Scenario: Empty kind omitted

- GIVEN the selected stack has no APM config (nil or empty)
- WHEN the data source writes state
- THEN `apm` SHALL be empty/null rather than a list of one empty object

## Known gaps (non-normative)

These record limitations of the current implementation. A later fix is a change that
updates the `SHALL`s above; do not treat this section as the desired contract.

### `lock` does not pin

Schema and registry docs describe `lock` as pinning the `latest` result so a new
stack release does not cascade. `stackFromFilters` still implements that pin when
`version` is non-empty, and the unit test `"returns the latest stackpack with a
locked version"` expects it. Read loads configuration only, so `version` is empty
and the pin is not Terraform-observable. A change that makes `lock` work should
replace the Read requirement above.

### `latest` with an empty list

The `latest` path indexes the first list element with no length check. An empty
list panics; it does not use the regex no-match diagnostic. A change should add
an explicit error (or documented empty handling) and a scenario with a positive
THEN.
