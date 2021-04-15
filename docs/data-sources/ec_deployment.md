---
page_title: "Elastic Cloud: ec_deployment"
description: |-
  Retrieves information about an Elastic Cloud deployment.
---

# Data Source: ec_deployment

Use this data source to retrieve information about an existing Elastic Cloud deployment.

## Example Usage

```hcl
data "ec_deployment" "example" {
  id = "f759065e5e64e9f3546f6c44f2743893"
}
```

## Argument Reference

* `id` - The ID of an existing Elastic Cloud deployment.

## Attributes Reference

~> **NOTE:** Depending on the deployment definition, some values may not be set.
These will not be available for interpolation.

* `alias` - Deployment alias.
* `healthy` - Overall health status of the deployment.
* `id` - The unique ID of the deployment.
* `name` - The name of the deployment.
* `region` - Region where the deployment can be found.
* `deployment_template_id` - ID of the deployment template used to create the deployment.
* `traffic_filter` - Traffic filter block, which contains a list of traffic filter rule identifiers.
* `tags` Key value map of arbitrary string tags.
* `observability` Observability settings. Information about logs and metrics shipped to a dedicated deployment.
  * `observability.#.deployment_id` - Destination deployment ID for the shipped logs and monitoring metrics.
  * `observability.#.ref_id` - Elasticsearch resource kind ref_id of the destination deployment.
  * `observability.#.logs` - Defines whether logs are enabled or disabled.
  * `observability.#.metrics` - Defines whether metrics are enabled or disabled.
* `elasticsearch` - Instance configuration of the Elasticsearch resource kind.
  * `elasticsearch.#.autoscale` - Whether or not Elasticsearch autoscaling is enabled.
  * `elasticsearch.#.healthy` - Resource kind health status.
  * `elasticsearch.#.cloud_id` - The encoded Elasticsearch credentials to use in Beats or Logstash. See [Configure Beats and Logstash with Cloud ID](https://www.elastic.co/guide/en/cloud/current/ec-cloud-id.html) for more information.
  * `elasticsearch.#.http_endpoint` - HTTP endpoint for the resource kind.
  * `elasticsearch.#.https_endpoint` - HTTPS endpoint for the resource kind.
  * `elasticsearch.#.ref_id` - User specified ref_id for the resource kind.
  * `elasticsearch.#.resource_id` - The resource unique identifier.
  * `elasticsearch.#.status` - Resource kind status (for example, "started", "stopped", etc).
  * `elasticsearch.#.version` - Elastic stack version.
  * `elasticsearch.#.topology` - Topology element definition.
    * `elasticsearch.#.topology.#.instance_configuration_id` - Controls the allocation of this topology element as well as allowed sizes and node_types. It needs to match the ID of an existing instance configuration.
    * `elasticsearch.#.topology.#.size` - Amount of memory (RAM) per topology element in the "<size in GB>g" notation.
    * `elasticsearch.#.topology.#.zone_count` - Number of zones in which nodes will be placed.
    * `elasticsearch.#.topology.#.node_roles` - Defines the list of Elasticsearch node roles assigned to the topology element (>=7.10.0).
    * `elasticsearch.#.topology.#.node_type_data` - Defines whether this node can hold data (<7.10.0).
    * `elasticsearch.#.topology.#.node_type_master` - Defines whether this node can be elected master (<7.10.0).
    * `elasticsearch.#.topology.#.node_type_ingest` - Defines whether this node can run an ingest pipeline (<7.10.0).
    * `elasticsearch.#.topology.#.node_type_ml` - Defines whether this node can run ML jobs (<7.10.0).
    * `elasticsearch.#.topology.#.autoscaling.#.max_size` - The maximum size for the scale up policy.
    * `elasticsearch.#.topology.#.autoscaling.#.max_size_resource` - The maximum size resource for the scale up policy.
    * `elasticsearch.#.topology.#.autoscaling.#.min_size` - The minimum size for the scale down policy.
    * `elasticsearch.#.topology.#.autoscaling.#.min_size_resource` - The minimum size for the scale down policy.
    * `elasticsearch.#.topology.#.autoscaling.#.policy_override_json` - The advanced policy overrides for the autoscaling policy.
* `kibana` - Instance configuration of the Kibana type.
  * `kibana.#.elasticsearch_cluster_ref_id` - The user-specified ID of the Elasticsearch cluster to which this resource kind will link.
  * `kibana.#.healthy` - Resource kind health status.
  * `kibana.#.http_endpoint` - HTTP endpoint for the resource kind.
  * `kibana.#.https_endpoint` - HTTPS endpoint for the resource kind.
  * `kibana.#.ref_id` - User specified ref_id for the resource kind.
  * `kibana.#.resource_id` - The resource unique identifier.
  * `kibana.#.status` - Resource kind status (for example, "started", "stopped", etc).
  * `kibana.#.version` - Elastic stack version.
  * `kibana.#.topology` - Node topology element definition.
    * `kibana.#.topology.#.instance_configuration_id` - Controls the allocation of this topology element as well as allowed sizes and node_types. It needs to match the ID of an existing instance configuration.
    * `kibana.#.topology.#.size` - Amount of memory (RAM) per topology element in the "<size in GB>g" notation.
    * `kibana.#.topology.#.zone_count` - Number of zones in which nodes will be placed.
* `apm` - Instance configuration of the APM type.
  * `apm.#.elasticsearch_cluster_ref_id` - The user-specified ID of the Elasticsearch cluster to which this resource kind will link.
  * `apm.#.healthy` - Resource kind health status.
  * `apm.#.http_endpoint` - HTTP endpoint for the resource kind.
  * `apm.#.https_endpoint` - HTTPS endpoint for the resource kind.
  * `apm.#.ref_id` - User specified ref_id for the resource kind.
  * `apm.#.resource_id` - The resource unique identifier.
  * `apm.#.status` - Resource kind status (for example, "started", "stopped", etc).
  * `apm.#.version` - Elastic stack version.
  * `apm.#.topology` - Node topology element definition.
    * `apm.#.topology.#.instance_configuration_id` - Controls the allocation of this topology element as well as allowed sizes and node_types. It needs to match the ID of an existing instance configuration.
    * `apm.#.topology.#.size` - Amount of memory (RAM) per topology element in the "<size in GB>g" notation.
    * `apm.#.topology.#.zone_count` - Number of zones in which nodes will be placed.
* `enterprise_search` - Instance configuration of the Enterprise Search type.
  * `enterprise_search.#.elasticsearch_cluster_ref_id` - The user-specified ID of the Elasticsearch cluster to which this resource kind will link.
  * `enterprise_search.#.healthy` - Resource kind health status.
  * `enterprise_search.#.http_endpoint` - HTTP endpoint for the resource kind.
  * `enterprise_search.#.https_endpoint` - HTTPS endpoint for the resource kind.
  * `enterprise_search.#.ref_id` - User specified ref_id for the resource kind.
  * `enterprise_search.#.resource_id` - The resource unique identifier.
  * `enterprise_search.#.status` - Resource kind status (for example, "started", "stopped", etc).
  * `enterprise_search.#.version` - Elastic stack version.
  * `enterprise_search.#.topology` - Node topology element definition.
    * `enterprise_search.#.topology.#.instance_configuration_id` - Controls the allocation of this topology element as well as allowed sizes and node_types. It needs to match the ID of an existing instance configuration.
    * `enterprise_search.#.topology.#.size` - Amount of memory (RAM) per topology element in the "<size in GB>g" notation.
    * `enterprise_search.#.topology.#.zone_count` - Number of zones in which nodes will be placed.
    * `enterprise_search.#.topology.#.node_type_appserver` - Defines whether this instance should run as application/API server.
    * `enterprise_search.#.topology.#.node_type_connector` - Defines whether this instance should run as connector.
    * `enterprise_search.#.topology.#.node_type_worker` - Defines whether this instance should run as background worker.
