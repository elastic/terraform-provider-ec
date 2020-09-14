---
page_title: "Elastic Cloud: ec_deployment"
description: |-
  Provides an Elastic Cloud deployment resource. This allows deployments to be created, updated, and deleted.
---

# Resource: ec_deployment

Provides an Elastic Cloud deployment resource. This allows deployments to be created, updated, and deleted.

## Example Usage

```hcl
resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = "7.9.1"
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch {
    topology {
      instance_configuration_id = "aws.data.highio.i3"
    }
  }

  kibana {
    topology {
      instance_configuration_id = "aws.kibana.r5d"
    }
  }

  apm {
    topology {
      instance_configuration_id = "aws.apm.r5d"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Required) ESS region where to create the deployment. For ECE environments "ece-region" must be set.
* `deployment_template_id` - (Required) Deployment Template identifier to create the deployment from.
* `version` - (Required) Elastic Stack version to use for all of the deployment resources.
* `name` - (Optional) Name for the deployment.
* `request_id` - (Optional) Request ID to set on the create operation. only use when previous create attempts return with an error and a request_id is returned as part of the error.
* `elasticsearch` (Required) Elasticsearch cluster definition, can only be specified once.
* `kibana` (Optional) Kibana instance definition, can only be specified once.
* `apm` (Optional) APM instance definition, can only be specified once.
* `enterprise_search` (Optional) Enterprise Search server definition, can only be specified once.
* `traffic_filter` (Optional) Traffic Filter block, which contains a list of traffic filter rule identifiers.

### Resources

In order to be able to create a valid deployment at least one resource type must be specified, below are the supported resources.

#### Elasticsearch

The required `elasticsearch` block supports the following:

* `topology` - (Required) Topology element which must be set once but can be set multiple times to compose complex topologies.
* `ref_id` - (Optional) ref_id to set on the Elasticsearch resource (Defaults to `main-elasticsearch`).
* `config` (Optional) Elasticsearch settings which will be applied to all topologies unless overridden on the topology element. 

##### Topology

The required `elasticsearch.topology` block supports the following:

* `instance_configuration_id` - (Required) Instance Configuration ID from the deployment template.
* `memory_per_node` - (Optional) Amount of memory (RAM) per node in the "<size in GB>g" notation (Defaults to `4g`).
* `zone_count` - (Optional) Number of zones that the Elasticsearch cluster will span. This is used to set HA (Defaults to `1`).
* `node_type_data` - (Optional) Node type (data) for the Elasticsearch Topology element (Defaults to `true`) 
* `node_type_master` - (Optional) Node type (master) for the Elasticsearch Topology element (Defaults to `true`)
* `node_type_ingest` - (Optional) Node type (ingest) for the Elasticsearch Topology element (Defaults to `true`)
* `node_type_ml` - (Optional) Node type (machine learning) for the Elasticsearch Topology element (Defaults to `false`).
* `config` (Optional) Elasticsearch settings which will be applied at the topology level. 

##### Config

The optional `elasticsearch.config` and `elasticsearch.topology.config` blocks support the following:

* `plugins` - (Optional) List of Elasticsearch supported plugins, which vary from version to version. Check the Stack Pack version to see which plugins are supported for each version. This is currently only available from the UI and [ecctl](https://www.elastic.co/guide/en/ecctl/master/ecctl_stack_list.html).
* `user_settings_json` - (Optional) JSON-formatted user level `elasticsearch.yml` setting overrides.
* `user_settings_override_json` - (Optional) JSON-formatted admin (ECE) level `elasticsearch.yml` setting overrides.
* `user_settings_yaml` - (Optional) YAML-formatted user level `elasticsearch.yml` setting overrides.
* `user_settings_override_yaml` - (Optional) YAML-formatted admin (ECE) level `elasticsearch.yml` setting overrides.

### Timeouts

* Default: 40 minutes.
* Update: 60 minutes.
* Delete: 60 minutes.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The deployment identifier.
* `elasticsearch_username` - The auto-generated Elasticsearch username.
* `elasticsearch_password` - The auto-generated Elasticsearch password.
* `elasticsearch.#.resource_id` - The Elasticsearch resource unique identifier.
* `elasticsearch.#.version` - The Elasticsearch current version.
* `elasticsearch.#.region` - The Elasticsearch region.
* `elasticsearch.#.cloud_id` - The encoded Elasticsearch credentials to use in Beats or Logstash, [more information](https://www.elastic.co/guide/en/cloud/current/ec-cloud-id.html).
* `elasticsearch.#.http_endpoint` - The Elasticsearch resource HTTP endpoint.
* `elasticsearch.#.https_endpoint` - The Elasticsearch resource HTTPs endpoint.

## Import

Deployments can be imported using the `id`, e.g.

```
$ terraform import ec_deployment.search 320b7b540dfc967a7a649c18e2fce4ed
```
