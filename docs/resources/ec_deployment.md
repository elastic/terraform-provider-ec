---
page_title: "Elastic Cloud: ec_deployment"
description: |-
  Provides an Elastic Cloud deployment resource. This allows deployments to be created, updated, and deleted.
---

# Resource: ec_deployment

Provides an Elastic Cloud deployment resource. This allows deployments to be created, updated, and deleted.

~> **Note on traffic filters** If you use `traffic_filter` on an `ec_deployment`, Terraform will assume management over the full set of traffic rules for the deployment, and treat additional traffic filters as drift. For this reason, `traffic_filter` cannot be mixed with the `ec_deployment_traffic_filter_association` resource for a given deployment.

-> **Note on regions and deployment templates** For a full list of regions and deployment templates available in the ESS, [please read this document](https://www.elastic.co/guide/en/cloud/current/ec-regions-templates-instances.html). You can also familiarize yourself with [Elastic Cloud deployments](https://www.elastic.co/guide/en/cloud/current/ec-create-deployment.html).

## Example Usage

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

* `region` - (Required) ESS region where to create the deployment. For ECE environments "ece-region" must be set.

-> Changing the `region` will cause the resource to be destroyed and re-created.

* `deployment_template_id` - (Required) Deployment Template identifier to create the deployment from. See the full list of available deployment templates on the top level note on regions and deployment templates.
* `version` - (Required) Elastic Stack version to use for all of the deployment resources.
* `name` - (Optional) Name for the deployment.
* `request_id` - (Optional) Request ID to set on the create operation. Only use when previous create attempts return with an error and a request_id is returned as part of the error.
* `elasticsearch` (Required) Elasticsearch cluster definition, can only be specified once. For multi-node Elasticsearch clusters, use multiple `topology` blocks.
* `kibana` (Optional) Kibana instance definition, can only be specified once.
* `apm` (Optional) APM instance definition, can only be specified once.
* `enterprise_search` (Optional) Enterprise Search server definition, can only be specified once. For multi-node Enterprise Search deployments, use multiple `topology` blocks.
* `traffic_filter` (Optional) Traffic Filters to apply to the deployment as a list of traffic filter rule identifiers.

### Resources

In order to create a valid deployment at least one resource type (`elasticsearch`) must be specified, below are the supported resources.

When empty blocks are specified (`elasticsearch {}`, `kibana {}`, `apm {}`, `enterprise_search {}`), a default topology from the deployment template will be used, for each of the blocks. When a block is not set, the resource kind will not be enabled in the deployment.

The `ec_deployment` resource, will opt-out all the resources except Elasticsearch, which as mentioned, will inherit the default topology from the deployment template, for example the [I/O Optimized template includes an Elasticsearch cluster 8 GB memory x 2 availability zones](https://www.elastic.co/guide/en/cloud/current/ec-getting-started-profiles.html#ec-getting-started-profiles-io).

To customize any of the deployment resource sizes or settings, use the `topology` block within.

#### Elasticsearch

The required `elasticsearch` block supports the following:

* `topology` - (Optional) Topology element which can be set multiple times to compose complex topologies.
* `ref_id` - (Optional) ref_id to set on the Elasticsearch resource, it is best left to the default value (Defaults to `main-elasticsearch`).
* `config` (Optional) Elasticsearch settings which will be applied to all topologies unless overridden on the topology element. 
* `remote_cluster` (Optional) Elasticsearch remote clusters to configure for the Elasticsearch resource, can be set multiple times.

##### Topology

To set up multi-node Elasticsearch clusters, single or multiple topology blocks can be set, each with a `instance_configuration_id` different set. This is particularly relevant for Elasticsearch clusters with Hot / Warm topologies, Machine Learning, etc.

The optional `elasticsearch.topology` block supports the following:

* `instance_configuration_id` - (Optional) Instance Configuration ID from the deployment template. By default, it will use the deployment template default instance configuration, but it can be changed. See top level note on regions and deployment templates.

-> Instance Configurations can be a little cryptic when getting started, make sure to read our documentation on the [ESS hardware and Instance Configurations](https://www.elastic.co/guide/en/cloud/current/ec-reference-hardware.html#ec-instance-configuration-names).

* `size` - (Optional) Amount of memory (RAM) per topology element in the "<size in GB>g" notation. When omitted, it will default to the deployment template value.
* `size_resource` - (Optional) Type of resource the size is being assigned to. (Defaults to `"memory"`).
* `zone_count` - (Optional) Number of zones that the instance type on the Elasticsearch cluster will span. This is used to set or unset HA on an Elasticsearch node type. When omitted, it will default to the deployment template value.
* `config` (Optional) Elasticsearch settings which will be applied at the topology level.

##### Config

The optional `elasticsearch.config` and `elasticsearch.topology.config` blocks support the following:

* `plugins` - (Optional) List of Elasticsearch supported plugins, which vary from version to version. Check the Stack Pack version to see which plugins are supported for each version. This is currently only available from the UI and [ecctl](https://www.elastic.co/guide/en/ecctl/master/ecctl_stack_list.html).
* `user_settings_json` - (Optional) JSON-formatted user level `elasticsearch.yml` setting overrides.
* `user_settings_override_json` - (Optional) JSON-formatted admin (ECE) level `elasticsearch.yml` setting overrides.
* `user_settings_yaml` - (Optional) YAML-formatted user level `elasticsearch.yml` setting overrides.
* `user_settings_override_yaml` - (Optional) YAML-formatted admin (ECE) level `elasticsearch.yml` setting overrides.

##### Remote Cluster

The optional `elasticsearch.remote_cluster` block can be set multiple times to represent multiple remote clusters the local Elasticsearch cluster connects to for Cross Cluster Search, supporting the following:

* `deployment_id` (Required) Remote deployment ID.
* `alias` (Optional) Alias for this Cross Cluster Search binding.
* `ref_id` (Optional) Remote elasticsearch `ref_id`, it is best left to the default value (Defaults to `main-elasticsearch`).
* `ignore_unavailable` (Optional) If true, skip the cluster during search when disconnected (Defaults to `false`).

#### Kibana

The optional `kibana` block supports the following:

* `topology` - (Optional) Topology element which can be set multiple times to compose complex topologies.
* `elasticsearch_cluster_ref_id` - (Optional) This field references the ref_id of the deployment Elasticsearch cluster, it is best left to the default value (Defaults to `main-elasticsearch`).
* `ref_id` - (Optional) ref_id to set on the Kibana resource. It is best left to the default value (Defaults to `main-kibana`).
* `config` (Optional) Kibana settings which will be applied to all topologies unless overridden on the topology element. 

##### Topology

The optional `kibana.topology` block supports the following:

* `instance_configuration_id` - (Optional) Instance Configuration ID from the deployment template. By default, it will use the deployment template default instance configuration, but it can be changed. See top level note on regions and deployment templates. There's no need to change this value since Kibana only has one _instance type_.
* `size` - (Optional) Amount of memory (RAM) per topology element in the "<size in GB>g" notation. When omitted, it will default to the deployment template value.
* `size_resource` - (Optional) Type of resource the size is being assigned to. (Defaults to `"memory"`).
* `zone_count` - (Optional) Number of zones that the Kibana deployment will span. This is used to set HA. When omitted, it will default to the deployment template value.
* `config` (Optional) Kibana settings which will be applied at the topology level. 

##### Config

The optional `kibana.config` and `kibana.topology.config` blocks support the following:

* `user_settings_json` - (Optional) JSON-formatted user level `kibana.yml` setting overrides.
* `user_settings_override_json` - (Optional) JSON-formatted admin (ECE) level `kibana.yml` setting overrides.
* `user_settings_yaml` - (Optional) YAML-formatted user level `kibana.yml` setting overrides.
* `user_settings_override_yaml` - (Optional) YAML-formatted admin (ECE) level `kibana.yml` setting overrides.

#### APM

The optional `apm` block supports the following:

* `topology` - (Optional) Topology element which can be set multiple times to compose complex topologies.
* `elasticsearch_cluster_ref_id` - (Optional) This field references the ref_id of the deployment Elasticsearch cluster, it is best left to the default value (Defaults to `main-elasticsearch`).
* `ref_id` - (Optional) ref_id to set on the APM resource. It is best left to the default value (Defaults to `main-apm`).
* `config` (Optional) APM settings which will be applied to all topologies unless overridden on the topology element. 

##### Topology

The optional `apm.topology` block supports the following:

* `instance_configuration_id` - (Optional) Instance Configuration ID from the deployment template. By default, it will use the deployment template default instance configuration, but it can be changed. See top level note on regions and deployment templates. There's no need to change this value since Kibana only has one _instance type_.
* `size` - (Optional) Amount of memory (RAM) per topology element in the "<size in GB>g" notation. When omitted, it will default to the deployment template value.
* `size_resource` - (Optional) Type of resource the size is being assigned to. (Defaults to `"memory"`).
* `zone_count` - (Optional) Number of zones that the APM deployment will span. This is used to set HA. When omitted, it will default to the deployment template value.
* `config` (Optional) APM settings which will be applied at the topology level. 

##### Config

The optional `apm.config` and `apm.topology.config` blocks support the following:

* `debug_enabled` - (Optional) Enable debug mode for APM servers (Defaults to `false`).
* `user_settings_json` - (Optional) JSON-formatted user level `apm.yml` setting overrides.
* `user_settings_override_json` - (Optional) JSON-formatted admin (ECE) level `apm.yml` setting overrides.
* `user_settings_yaml` - (Optional) YAML-formatted user level `apm.yml` setting overrides.
* `user_settings_override_yaml` - (Optional) YAML-formatted admin (ECE) level `apm.yml` setting overrides.

#### Enterprise Search

The optional `enterprise_search` block supports the following:

* `topology` - (Optional) Topology element which can be set multiple times to compose complex topologies.
* `elasticsearch_cluster_ref_id` - (Optional) This field references the ref_id of the deployment Elasticsearch cluster, it is best left to the default value (Defaults to `main-elasticsearch`).
* `ref_id` - (Optional) ref_id to set on the Enterprise Search resource. It is best left to the default value (Defaults to `main-enterprise_search`).
* `config` (Optional) Enterprise Search settings which will be applied to all topologies unless overridden on the topology element. 

##### Topology

The optional `enterprise_search.topology` block supports the following:

* `instance_configuration_id` - (Optional) Instance Configuration ID from the deployment template. By default, it will use the deployment template default instance configuration, but it can be changed. See top level note on regions and deployment templates.
* `size` - (Optional) Amount of memory (RAM) per topology element in the "<size in GB>g" notation. When omitted, it will default to the deployment template value.
* `size_resource` - (Optional) Type of resource the size is being assigned to. (Defaults to `"memory"`).
* `zone_count` - (Optional) Number of zones that the Enterprise Search deployment will span. This is used to set HA. When omitted, it will default to the deployment template value.
* `config` (Optional) Enterprise Search settings which will be applied at the topology level. 

##### Config

The optional `enterprise_search.config` and `enterprise_search.topology.config` blocks support the following:

* `user_settings_json` - (Optional) JSON-formatted user level `enterprise_search.yml` setting overrides.
* `user_settings_override_json` - (Optional) JSON-formatted admin (ECE) level `enterprise_search.yml` setting overrides.
* `user_settings_yaml` - (Optional) YAML-formatted user level `enterprise_search.yml` setting overrides.
* `user_settings_override_yaml` - (Optional) YAML-formatted admin (ECE) level `enterprise_search.yml` setting overrides.

### Timeouts

* Default: 40 minutes.
* Update: 60 minutes.
* Delete: 60 minutes.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The deployment identifier.
* `elasticsearch_username` - The auto-generated Elasticsearch username.
* `elasticsearch_password` - The auto-generated Elasticsearch password.
* `apm_secret_token` - The auto-generated APM secret_token, empty unless an `apm` resource is specified.
* `elasticsearch.#.resource_id` - The Elasticsearch resource unique identifier.
* `elasticsearch.#.version` - The Elasticsearch current version.
* `elasticsearch.#.region` - The Elasticsearch region.
* `elasticsearch.#.cloud_id` - The encoded Elasticsearch credentials to use in Beats or Logstash, [more information](https://www.elastic.co/guide/en/cloud/current/ec-cloud-id.html).
* `elasticsearch.#.http_endpoint` - The Elasticsearch resource HTTP endpoint.
* `elasticsearch.#.https_endpoint` - The Elasticsearch resource HTTPs endpoint.
* `elasticsearch.#.topology.#.node_type_data` - Node type (data) for the Elasticsearch Topology element.
* `elasticsearch.#.topology.#.node_type_master` - Node type (master) for the Elasticsearch Topology element.
* `elasticsearch.#.topology.#.node_type_ingest` - Node type (ingest) for the Elasticsearch Topology element.
* `elasticsearch.#.topology.#.node_type_ml` - Node type (machine learning) for the Elasticsearch Topology element.
* `kibana.#.resource_id` - The Kibana resource unique identifier.
* `kibana.#.version` - The Kibana current version.
* `kibana.#.region` - The Kibana region.
* `kibana.#.http_endpoint` - The Kibana resource HTTP endpoint.
* `kibana.#.https_endpoint` - The Kibana resource HTTPs endpoint.
* `apm.#.resource_id` - The APM resource unique identifier.
* `apm.#.version` - The APM current version.
* `apm.#.region` - The APM region.
* `apm.#.http_endpoint` - The APM resource HTTP endpoint.
* `apm.#.https_endpoint` - The APM resource HTTPs endpoint.
* `enterprise_search.#.resource_id` - The Enterprise Search resource unique identifier.
* `enterprise_search.#.version` - The Enterprise Search current version.
* `enterprise_search.#.region` - The Enterprise Search region.
* `enterprise_search.#.http_endpoint` - The Enterprise Search resource HTTP endpoint.
* `enterprise_search.#.https_endpoint` - The Enterprise Search resource HTTPs endpoint.
* `enterprise_search.#.topology.#.node_type_appserver` - Node type (Appserver) for the Enterprise Search Topology element.
* `enterprise_search.#.topology.#.node_type_connector` - Node type (Connector) for the Enterprise Search Topology element.
* `enterprise_search.#.topology.#.node_type_worker` - Node type (worker) for the Enterprise Search Topology element.

## Import

~> **Note on legacy (pre-slider) deployments** If your deployment was created prior to the addition of sliders in both ECE or ESS, and has not been migrated to use sliders, importing it is not supported and will result in erroneous behavior.

Deployments can be imported using the `id`, e.g.

```
$ terraform import ec_deployment.search 320b7b540dfc967a7a649c18e2fce4ed
```
