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

  size = 200

  tags = {
    "foo" = "bar"
  }

  elasticsearch {
    healthy = "true"
  }

  kibana {
    status = "started"
  }

  integrations_server {
    version = "8.0.0"
  }

  enterprise_search {
    healthy = "true"
  }
}
```

## Argument Reference

* `name_prefix` - Prefix that one or several deployment names have in common.
* `deployment_template_id` - ID of the deployment template used to create the deployment.
* `size` - The maximum number of deployments to return. Defaults to `100`.
* `tags` - Key value map of arbitrary string tags for the deployment.
* `healthy` - Overall health status of the deployment.
* `elasticsearch` - Filter by Elasticsearch resource kind status or configuration.
  * `elasticsearch.#.status` - Resource kind status (Available statuses are: initializing, stopping, stopped, rebooting, restarting, reconfiguring, and started).
  * `elasticsearch.#.version` - Elastic stack version.
  * `elasticsearch.#.healthy` - Overall health status of the Elasticsearch instances.
* `kibana` - Filter by Kibana resource kind status or configuration.
  * `kibana.#.status` - Resource kind status (Available statuses are: initializing, stopping, stopped, rebooting, restarting, reconfiguring, and started).
  * `kibana.#.version` - Elastic stack version.
  * `kibana.#.healthy` - Overall health status of the Kibana instances.
* `integrations_server` - Filter by Integrations Server resource kind status or configuration.
  * `integrations_server.#.status` - Resource kind status (Available statuses are: initializing, stopping, stopped, rebooting, restarting, reconfiguring, and started).
  * `integrations_server.#.version` - Elastic stack version.
  * `integrations_server.#.healthy` - Overall health status of the Integrations Server instances.
* `apm` - **DEPRECATED** Filter by APM resource kind status or configuration.
  * `apm.#.status` - Resource kind status (Available statuses are: initializing, stopping, stopped, rebooting, restarting, reconfiguring, and started).
  * `apm.#.version` - Elastic stack version.
  * `apm.#.healthy` - Overall health status of the APM instances.
* `enterprise_search` - Filter by Enterprise Search resource kind status or configuration.
  * `enterprise_search.#.status` - Resource kind status (Available statuses are: initializing, stopping, stopped, rebooting, restarting, reconfiguring, and started).
  * `enterprise_search.#.version` - Elastic stack version.
  * `enterprise_search.#.healthy` - Overall health status of the Enterprise Search instances.

~> **NOTE:** The `apm` resource has been deprecated starting on the Elastic Stack Version 8.0.0. New deployments  should use `integrations_server` instead.

## Attributes Reference

~> **NOTE:** Depending on the deployment definition, some values may not be set.
These will not be available for interpolation.

* `deployments` - List of deployments which match the specified query.
  * `deployments.#.deployment_id` - The deployment unique ID.
  * `deployments.#.alias` - Deployment alias.
  * `deployments.#.name` - The name of the deployment.
  * `deployments.#.elasticsearch_resource_id` - The Elasticsearch resource unique ID.
  * `deployments.#.elasticsearch_ref_id` - The Elasticsearch resource reference.
  * `deployments.#.kibana_resource_id` - The Kibana resource unique ID.
  * `deployments.#.kibana_ref_id` - The Kibana resource reference.
  * `deployments.#.integrations_server_resource_id` - The Integrations Server resource unique ID.
  * `deployments.#.integrations_server_ref_id` - The Integrations Server resource reference.
  * `deployments.#.apm_resource_id` - The APM resource unique ID.
  * `deployments.#.apm_ref_id` - The APM resource reference.
  * `deployments.#.enterprise_search_resource_id` - The Enterprise Search resource unique ID.
  * `deployments.#.enterprise_search_ref_id` - The Enterprise Search resource reference.
