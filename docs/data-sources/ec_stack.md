---
page_title: "Elastic Cloud: ec_stack"
description: |-
  Retrieves information of an Elastic Cloud stack.
---

# Data Source: ec_deployment

Use this data source to retrieve information about an existing Elastic Cloud stack.

## Example Usage

```hcl
data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
  lock          = true
}

data "ec_stack" "latest_patch" {
  version_regex = "7.9.?"
  region        = "us-east-1"
}
```

## Argument Reference
s
* `version_regex` (Required) - Regex to filter the available stacks. Can be any valid regex expression, when multiple stacks are matched through a regex, the latest version is returned. `"latest"` is also accepted to obtain the latest available Stack version.
* `region` (Required) - Region where the stack pack is, `"ece-region` needs to be used for ece.
* `lock` (Optional) - Lock the `"latest"` `version_regex` obtained so that when there's a new stack release, the new release doesn't cascade the changes down to the deployments, can be changed at any time.

## Attributes Reference

~> **NOTE:** Depending on the stack version, some values may not be set. These will not be available for interpolation.

* `version` - The stack version.
* `accessible` - Whether or not this version is accessible by the calling user. This is only relevant in EC (SaaS) and is not sent in ECE.
* `min_upgradable_from` - The minimum version recommended to upgrade to this version.
* `upgradable_to` - The Stack versions that this version can upgrade to.
* `allowlisted` - Whether or not this version is in the allowlist. This is only relevant in EC (SaaS) and is not sent in ECE.
* `apm` - Information for APM workloads on this Stack version.
  * `apm.#.denylist` - List of configuration options that cannot be overridden by user settings.
  * `apm.#.capacity_constraints_min` - Minimum size of the instances.
  * `apm.#.capacity_constraints_max` - Maximum size of the instances.
  * `apm.#.compatible_node_types` - List of node types compatible with this one.
  * `apm.#.docker_image` - Docker image to use for the APM instance.
* `elasticsearch` - Information for Elasticsearch workloads on this Stack version.
  * `elasticsearch.#.denylist` - List of configuration options that cannot be overridden by user settings.
  * `elasticsearch.#.capacity_constraints_min` - Minimum size of the instances.
  * `elasticsearch.#.capacity_constraints_max` - Maximum size of the instances.
  * `elasticsearch.#.compatible_node_types` - List of node types compatible with this one.
  * `elasticsearch.#.default_plugins` - List of default plugins which are included in all Elasticsearch cluster instances.
  * `elasticsearch.#.docker_image` - Docker image to use for the Elasticsearch cluster instances.
  * `elasticsearch.#.plugins` - List of available plugins to be specified by users in Elasticsearch cluster instances.
* `enterprise_search` - Information for Enterprise Search workloads on this Stack version.
  * `enterprise_search.#.denylist` - List of configuration options that cannot be overridden by user settings.
  * `enterprise_search.#.capacity_constraints_min` - Minimum size of the instances.
  * `enterprise_search.#.capacity_constraints_max` - Maximum size of the instances.
  * `enterprise_search.#.compatible_node_types` - List of node types compatible with this one.
  * `enterprise_search.#.docker_image` - Docker image to use for the Enterprise Search instance.
* `kibana` - Information for Kibana workloads on this Stack version.
  * `kibana.#.denylist` - List of configuration options that cannot be overridden by user settings.
  * `kibana.#.capacity_constraints_min` - Minimum size of the instances.
  * `kibana.#.capacity_constraints_max` - Maximum size of the instances.
  * `kibana.#.compatible_node_types` - List of node types compatible with this one.
  * `kibana.#.docker_image` - Docker image to use for the Kibana instance.
