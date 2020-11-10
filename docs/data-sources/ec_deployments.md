---
page_title: "Elastic Cloud: ec_deployments"
description: |-
  Returns a list of deployments that match the specified query.
---

# Data Source: ec_deployments

Use this data source to retrieve a list of IDs for the deployment and resource kinds, based on the specified query.

## Example Usage

```hcl
data "ec_deployments" "example" {
  name_prefix            = "test"
  deployment_template_id = "azure-compute-optimized"

  elasticsearch {
    healthy = "true"
  }

  kibana {
    status = "started"
  }

  apm {
    version = "7.9.1"
  }

  enterprise_search {
    healthy = "true"
  }
}
```

## Argument Reference

* `name_prefix` - Prefix that one or several deployment names have in common.
* `deployment_template_id` - ID of the deployment template used to create the deployment.
* `healthy` - Overall health status of the deployment.
* `elasticsearch` - Filter by Elasticsearch resource kind status or configuration.
  * `elasticsearch.#.status` - Resource kind status (for example, "started", "stopped", etc).
  * `elasticsearch.#.version` - Elastic Stack version.
  * `elasticsearch.#.healthy` - Overall health status of the Elasticsearch instances.
* `kibana` - Filter by Kibana resource kind status or configuration.
  * `kibana.#.status` - Resource kind status (for example, "started", "stopped", etc).
  * `kibana.#.version` - Elastic Stack version.
  * `kibana.#.healthy` - Overall health status of the Kibana instances.
* `apm` - Filter by APM resource kind status or configuration.
  * `apm.#.status` - Resource kind status (for example, "started", "stopped", etc).
  * `apm.#.version` - Elastic Stack version.
  * `apm.#.healthy` - Overall health status of the APM instances.
* `enterprise_search` - Filter by Enterprise Search resource kind status or configuration.
  * `enterprise_search.#.status` - Resource kind status (for example, "started", "stopped", etc).
  * `enterprise_search.#.version` - Elastic Stack version.
  * `enterprise_search.#.healthy` - Overall health status of the Enterprise Search instances.

## Attributes Reference

~> **NOTE:** Depending on the deployment definition, some values may not be set.
These will not be available for interpolation.

* `deployments` - List of deployments which match the specified query.
  * `deployments.#.deployment_id` - The deployment unique ID.
  * `deployments.#.elasticsearch_resource_id` - The Elasticsearch resource unique ID.
  * `deployments.#.kibana_resource_id` - The Kibana resource unique ID.
  * `deployments.#.apm_resource_id` - The APM resource unique ID.
  * `deployments.#.enterprise_search_resource_id` - The Enterprise Search resource unique ID.
