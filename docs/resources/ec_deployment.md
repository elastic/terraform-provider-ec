---
page_title: "Elastic Cloud: ec_deployment"
description: |-
  Provides an Elastic Cloud deployment resource, which allows deployments to be created, updated, and deleted.
---

# Resource: ec_deployment

Provides an Elastic Cloud deployment resource, which allows deployments to be created, updated, and deleted.

~> **Note on traffic filters** If you use `traffic_filter` on an `ec_deployment`, Terraform will manage the full set of traffic rules for the deployment, and treat additional traffic filters as drift. For this reason, `traffic_filter` cannot be mixed with the `ec_deployment_traffic_filter_association` resource for a given deployment.

-> **Note on regions and deployment templates** Before you start, you might want to read about [Elastic Cloud deployments](https://www.elastic.co/guide/en/cloud/current/ec-create-deployment.html) and check the [full list](https://www.elastic.co/guide/en/cloud/current/ec-regions-templates-instances.html) of regions and deployment templates available in Elasticsearch Service (ESS).

## Example Usage

### Basic

```hcl
resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = "7.9.2"
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch {}

  kibana {}

  apm {}

  enterprise_search {}
}
```

### With observability settings

```hcl
resource "ec_deployment" "example_observability" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = "7.9.2"
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch {}

  kibana {}

  # Optional observability settings
  observability {
    deployment_id = ec_deployment.example_minimal.id
  }
}
```

### With Cross Cluster Search settings

```hcl
resource "ec_deployment" "source_deployment" {
  name = "my_ccs_source"

  region                 = "us-east-1"
  version                = "7.9.2"
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch {
    topology {
      size = "1g"
    }
  }
}

resource "ec_deployment" "ccs" {
  name = "ccs deployment"

  region                 = "us-east-1"
  version                = "7.9.2"
  deployment_template_id = "aws-cross-cluster-search-v2"

  elasticsearch {
    remote_cluster {
      deployment_id = ec_deployment.source_deployment.id
      alias         = ec_deployment.source_deployment.name
      ref_id        = ec_deployment.source_deployment.elasticsearch.0.ref_id
    }
  }

  kibana {}
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Required) Elasticsearch Service (ESS) region where to create the deployment. For Elastic Cloud Enterprise (ECE) installations, set `"ece-region"`.

-> If you change the `region`, the resource is destroyed and re-created.

* `deployment_template_id` - (Required) Deployment template identifier to create the deployment from. See the [full list](https://www.elastic.co/guide/en/cloud/current/ec-regions-templates-instances.html) of regions and deployment templates available in ESS.
* `version` - (Required) Elastic Stack version to use for all the deployment resources.
* `name` - (Optional) Name of the deployment.
* `request_id` - (Optional) Request ID to set when you create the deployment. Use it only when previous attempts return an error and `request_id` is returned as part of the error.
* `elasticsearch` (Required) Elasticsearch cluster definition, can only be specified once. For multi-node Elasticsearch clusters, use multiple `topology` blocks.
* `kibana` (Optional) Kibana instance definition, can only be specified once.
* `apm` (Optional) APM instance definition, can only be specified once.
* `enterprise_search` (Optional) Enterprise Search server definition, can only be specified once. For multi-node Enterprise Search deployments, use multiple `topology` blocks.
* `traffic_filter` (Optional) List of traffic filter rule identifiers that will be applied to the deployment.
* `observability` (Optional) Observability settings that you can set to ship logs and metrics to a separate deployment.

### Resources

!> **Warning on removing explicit topology objects** Due to current limitations, if a topology object is removed from the configuration, the removal won't trigger any changes since the field is optional and computed. There is no way to determine if the block was removed, which results in a _"sticky"_ topology configuration.

To create a valid deployment, you must specify at least the resource type `elasticsearch`. The supported resources are listed below.

A default topology from the deployment template is used for empty blocks: `elasticsearch {}`, `kibana {}`, `apm {}`, `enterprise_search {}`. When a block is not set, the resource kind is not enabled in the deployment.

The `ec_deployment` resource will opt-out all the resources except Elasticsearch, which inherits the default topology from the deployment template. For example, the [I/O Optimized template includes an Elasticsearch cluster 8 GB memory x 2 availability zones](https://www.elastic.co/guide/en/cloud/current/ec-getting-started-profiles.html#ec-getting-started-profiles-io).

To customize the size or settings of the deployment resource, use the `topology` block within each resource kind block.

#### Elasticsearch

The required `elasticsearch` block supports the following arguments:

* `topology` - (Optional) Can be set multiple times to compose complex topologies.
* `ref_id` - (Optional) Can be set on the Elasticsearch resource. The default value `main-elasticsearch` is recommended.
* `config` (Optional) Applied Elasticsearch settings to all topologies unless overridden in the `topology` element. 
* `remote_cluster` (Optional) Elasticsearch remote clusters that can be set multiple times.

##### Topology

To set up multi-node Elasticsearch clusters, you can set single or multiple topology blocks, each one with a different `instance_configuration_id`. This is particularly relevant for Elasticsearch clusters with hot-warm topologies or Machine Learning.

The optional `elasticsearch.topology` block supports the following arguments:

* `instance_configuration_id` - (Optional) Default instance configuration of the deployment template. To change it, use the [full list](https://www.elastic.co/guide/en/cloud/current/ec-regions-templates-instances.html) of regions and deployment templates available in ESS.

-> Before you get started with instance configurations, read the [ESS hardware and Instance Configurations](https://www.elastic.co/guide/en/cloud/current/ec-reference-hardware.html#ec-instance-configuration-names) documentation.

* `size` - (Optional) Amount of memory (RAM) per topology element in the `"<size in GB>g"` notation. When omitted, it defaults to the deployment template value.
* `size_resource` - (Optional) Type of resource to which the size is assigned. Defaults to `"memory"`.
* `zone_count` - (Optional) Number of zones the instance type of the Elasticsearch cluster will span. This is used to set or unset HA on an Elasticsearch node type. When omitted, it defaults to the deployment template value.
* `config` (Optional) Elasticsearch settings applied at the topology level.

##### Config

The optional `elasticsearch.config` and `elasticsearch.topology.config` blocks support the following arguments:

* `plugins` - (Optional) List of Elasticsearch supported plugins. Check the Stack Pack version to see which plugins are supported for each version. This is currently only available from the UI and [ecctl](https://www.elastic.co/guide/en/ecctl/master/ecctl_stack_list.html).
* `user_settings_json` - (Optional) JSON-formatted user level `elasticsearch.yml` setting overrides.
* `user_settings_override_json` - (Optional) JSON-formatted admin (ECE) level `elasticsearch.yml` setting overrides.
* `user_settings_yaml` - (Optional) YAML-formatted user level `elasticsearch.yml` setting overrides.
* `user_settings_override_yaml` - (Optional) YAML-formatted admin (ECE) level `elasticsearch.yml` setting overrides.

##### Remote Cluster

The optional `elasticsearch.remote_cluster` block can be set multiple times. It represents one or multiple remote clusters to which the local Elasticsearch cluster connects for Cross Cluster Search and supports the following settings:

* `deployment_id` (Required) Remote deployment ID.
* `alias` (Optional) Alias for the Cross Cluster Search binding.
* `ref_id` (Optional) Remote Elasticsearch `ref_id`. The default value `main-elasticsearch` is recommended.
* `ignore_unavailable` (Optional) If true, skip the cluster during search when disconnected. Defaults to `false`.

#### Kibana

The optional `kibana` block supports the following arguments:

* `topology` - (Optional) Can be set multiple times to compose complex topologies.
* `elasticsearch_cluster_ref_id` - (Optional) This field references the `ref_id` of the deployment Elasticsearch cluster. The default value `main-elasticsearch` is recommended.
* `ref_id` - (Optional) Can be set on the Kibana resource. The default value `main-kibana` is recommended.
* `config` (Optional) Kibana settings applied to all topologies unless overridden in the `topology` element. 

##### Topology

The optional `kibana.topology` block supports the following arguments:

* `instance_configuration_id` - (Optional) Default instance configuration of the deployment template. No need to change this value since Kibana has only one _instance type_.
* `size` - (Optional) Amount of memory (RAM) per topology element in the "<size in GB>g" notation. When omitted, it defaults to the deployment template value.
* `size_resource` - (Optional) Type of resource to which the size is assigned. Defaults to `"memory"`.
* `zone_count` - (Optional) Number of zones that the Kibana deployment will span. This is used to set HA. When omitted, it defaults to the deployment template value.
* `config` (Optional) Kibana settings applied at the topology level. 

##### Config

The optional `kibana.config` and `kibana.topology.config` blocks support the following arguments:

* `user_settings_json` - (Optional) JSON-formatted user level `kibana.yml` setting overrides.
* `user_settings_override_json` - (Optional) JSON-formatted admin (ECE) level `kibana.yml` setting overrides.
* `user_settings_yaml` - (Optional) YAML-formatted user level `kibana.yml` setting overrides.
* `user_settings_override_yaml` - (Optional) YAML-formatted admin (ECE) level `kibana.yml` setting overrides.

#### APM

The optional `apm` block supports the following arguments:

* `topology` - (Optional) Can be set multiple times to compose complex topologies.
* `elasticsearch_cluster_ref_id` - (Optional) This field references the `ref_id` of the deployment Elasticsearch cluster. The default value `main-elasticsearch` is recommended.
* `ref_id` - (Optional) Can be set on the APM resource. The default value `main-apm` is recommended.
* `config` (Optional) APM settings applied to all topologies unless overridden in the `topology` element. 

##### Topology

The optional `apm.topology` block supports the following arguments:

* `instance_configuration_id` - (Optional) Default instance configuration of the deployment template. No need to change this value since APM has only one _instance type_.
* `size` - (Optional) Amount of memory (RAM) per topology element in the "<size in GB>g" notation. When omitted, it defaults to the deployment template value.
* `size_resource` - (Optional) Type of resource to which the size is assigned. Defaults to `"memory"`.
* `zone_count` - (Optional) Number of zones that the APM deployment will span. This is used to set HA. When omitted, it defaults to the deployment template value.
* `config` (Optional) APM settings applied at the topology level. 

##### Config

The optional `apm.config` and `apm.topology.config` blocks support the following arguments:

* `debug_enabled` - (Optional) Enable debug mode for APM servers. Defaults to `false`.
* `user_settings_json` - (Optional) JSON-formatted user level `apm.yml` setting overrides.
* `user_settings_override_json` - (Optional) JSON-formatted admin (ECE) level `apm.yml` setting overrides.
* `user_settings_yaml` - (Optional) YAML-formatted user level `apm.yml` setting overrides.
* `user_settings_override_yaml` - (Optional) YAML-formatted admin (ECE) level `apm.yml` setting overrides.

#### Enterprise Search

The optional `enterprise_search` block supports the following arguments:

* `topology` - (Optional) Can be set multiple times to compose complex topologies.
* `elasticsearch_cluster_ref_id` - (Optional) This field references the `ref_id` of the deployment Elasticsearch cluster. The default value `main-elasticsearch` is recommended.
* `ref_id` - (Optional) Can be set on the Enterprise Search resource. The default value `main-enterprise_search` is recommended.
* `config` (Optional) Enterprise Search settings applied to all topologies unless overridden in the `topology` element. 

##### Topology

The optional `enterprise_search.topology` block supports the following settings:

* `instance_configuration_id` - (Optional) Default instance configuration of the deployment template. To change it, use the [full list](https://www.elastic.co/guide/en/cloud/current/ec-regions-templates-instances.html) of regions and deployment templates available in ESS.
* `size` - (Optional) Amount of memory (RAM) per `topology` element in the "<size in GB>g" notation. When omitted, it defaults to the deployment template value.
* `size_resource` - (Optional) Type of resource to which the size is assigned. Defaults to `"memory"`.
* `zone_count` - (Optional) Number of zones that the Enterprise Search deployment will span. This is used to set HA. When omitted, it defaults to the deployment template value.
* `config` (Optional) Enterprise Search settings applied at the topology level. 

##### Config

The optional `enterprise_search.config` and `enterprise_search.topology.config` blocks support the following arguments:

* `user_settings_json` - (Optional) JSON-formatted user level `enterprise_search.yml` setting overrides.
* `user_settings_override_json` - (Optional) JSON-formatted admin (ECE) level `enterprise_search.yml` setting overrides.
* `user_settings_yaml` - (Optional) YAML-formatted user level `enterprise_search.yml` setting overrides.
* `user_settings_override_yaml` - (Optional) YAML-formatted admin (ECE) level `enterprise_search.yml` setting overrides.

### Timeouts

* Default: 40 minutes.
* Update: 60 minutes.
* Delete: 60 minutes.

## Attributes Reference

In addition to all the arguments above, the following attributes are exported:

* `id` - Deployment identifier.
* `elasticsearch_username` - Auto-generated Elasticsearch username.
* `elasticsearch_password` - Auto-generated Elasticsearch password.
* `apm_secret_token` - Auto-generated APM secret_token, empty unless an `apm` resource is specified.
* `elasticsearch.#.resource_id` - Elasticsearch resource unique identifier.
* `elasticsearch.#.version` - Elasticsearch current version.
* `elasticsearch.#.region` - Elasticsearch region.
* `elasticsearch.#.cloud_id` - Encoded Elasticsearch credentials to use in Beats or Logstash. For more information, see [Configure Beats and Logstash with Cloud ID](https://www.elastic.co/guide/en/cloud/current/ec-cloud-id.html).
* `elasticsearch.#.http_endpoint` - Elasticsearch resource HTTP endpoint.
* `elasticsearch.#.https_endpoint` - Elasticsearch resource HTTPs endpoint.
* `elasticsearch.#.topology.#.node_type_data` - Node type (data) for the Elasticsearch topology element.
* `elasticsearch.#.topology.#.node_type_master` - Node type (master) for the Elasticsearch topology element.
* `elasticsearch.#.topology.#.node_type_ingest` - Node type (ingest) for the Elasticsearch topology element.
* `elasticsearch.#.topology.#.node_type_ml` - Node type (machine learning) for the Elasticsearch topology element.
* `kibana.#.resource_id` - Kibana resource unique identifier.
* `kibana.#.version` - Kibana current version.
* `kibana.#.region` - Kibana region.
* `kibana.#.http_endpoint` - Kibana resource HTTP endpoint.
* `kibana.#.https_endpoint` - Kibana resource HTTPs endpoint.
* `apm.#.resource_id` - APM resource unique identifier.
* `apm.#.version` - APM current version.
* `apm.#.region` - APM region.
* `apm.#.http_endpoint` - APM resource HTTP endpoint.
* `apm.#.https_endpoint` - APM resource HTTPs endpoint.
* `enterprise_search.#.resource_id` - Enterprise Search resource unique identifier.
* `enterprise_search.#.version` - Enterprise Search current version.
* `enterprise_search.#.region` - Enterprise Search region.
* `enterprise_search.#.http_endpoint` - Enterprise Search resource HTTP endpoint.
* `enterprise_search.#.https_endpoint` - Enterprise Search resource HTTPs endpoint.
* `enterprise_search.#.topology.#.node_type_appserver` - Node type (Appserver) for the Enterprise Search topology element.
* `enterprise_search.#.topology.#.node_type_connector` - Node type (Connector) for the Enterprise Search topology element.
* `enterprise_search.#.topology.#.node_type_worker` - Node type (worker) for the Enterprise Search topology element.
* `observability.#.deployment_id` - Destination deployment ID for the shipped logs and monitoring metrics.
* `observability.#.ref_id` - (Optional) Elasticsearch resource kind ref_id of the destination deployment.
* `observability.#.logs` - Enables or disables shipping logs. Defaults to true.
* `observability.#.metrics` - Enables or disables shipping metrics. Defaults to true.


## Import

~> **Note on legacy (pre-slider) deployments** Importing deployments created prior to the addition of sliders in ECE or ESS, without being migrated to use sliders, is not supported.

Deployments can be imported using the `id`, for example:

```
$ terraform import ec_deployment.search 320b7b540dfc967a7a649c18e2fce4ed
```
