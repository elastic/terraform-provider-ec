---
page_title: "Elastic Cloud: ec_deployment"
description: |-
  Provides an Elastic Cloud deployment resource, which allows deployments to be created, updated, and deleted.
---

# Resource: ec_deployment

Provides an Elastic Cloud deployment resource, which allows deployments to be created, updated, and deleted.

~> **Note on Elastic Stack versions** Using a version prior to `6.6.0` is not supported.

~> **Note on traffic filters** If you use `traffic_filter` on an `ec_deployment`, Terraform will manage the full set of traffic rules for the deployment, and treat additional traffic filters as drift. For this reason, `traffic_filter` cannot be mixed with the `ec_deployment_traffic_filter_association` resource for a given deployment.

-> **Note on regions and deployment templates** Before you start, you might want to read about [Elastic Cloud deployments](https://www.elastic.co/guide/en/cloud/current/ec-create-deployment.html) and check the [full list](https://www.elastic.co/guide/en/cloud/current/ec-regions-templates-instances.html) of regions and deployment templates available in Elasticsearch Service (ESS).

-> **Note on Elasticsearch topology IDs** Since the addition of data tiers, each Elasticsearch topology block requires the `"id"` field to be set. The accepted values are set in the deployment template that you have chosen, but values are closely related to the Elasticsearch data tiers. [Learn more abut Elasticsearch data tiers](https://www.elastic.co/guide/en/elasticsearch/reference/current/data-tiers.html). For a complete list of all the supported values, refer to the deployment template definition used by your deployment.

## Example Usage

### Basic

```hcl
data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }

  kibana = {}

  integrations_server = {}

  enterprise_search = {}
}
```

### Tiered deployment with Autoscaling enabled

```hcl
data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

resource "ec_deployment" "example_minimal" {
  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch = {

    autoscale = "true"

    # If `autoscale` is set, all topology elements that
    # - either set `size` in the plan or
    # - have non-zero default `max_size` (that is read from the deployment templates's `autoscaling_max` value)
    # have to be listed even if their blocks don't specify other fields beside `id`

    cold = {
      autoscaling = {}
    }

    frozen = {
      autoscaling = {}
    }

    hot = {
      size = "8g"

      autoscaling = {
        max_size          = "128g"
        max_size_resource = "memory"
      }
    }

    ml = {
      autoscaling = {}
    }

    warm = {
      autoscaling = {}
    }
  }

  # Initial size for `hot_content` tier is set to 8g
  # so `hot_content`'s size has to be added to the `ignore_changes` meta-argument to ignore future modifications that can be made by the autoscaler
  lifecycle {
    ignore_changes = [
      elasticsearch.hot.size
    ]
  }

  kibana = {}

  integrations_server = {}

  enterprise_search = {}
}
```

### With observability settings

```hcl
data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

resource "ec_deployment" "example_observability" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }

  kibana = {}

  # Optional observability settings
  observability = {
    deployment_id = ec_deployment.example_minimal.id
  }
}
```

It is possible to enable observability without using a second deployment, by storing the observability data in the current deployment. To enable this, set `deployment_id` to `self`.
```hcl
observability = {
  deployment_id = "self"
}
```

### With Cross Cluster Search settings

```hcl
data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

resource "ec_deployment" "source_deployment" {
  name = "my_ccs_source"

  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch = {
    hot = {
      size        = "1g"
      autoscaling = {}
    }
  }
}

resource "ec_deployment" "ccs" {
  name = "ccs deployment"

  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-cross-cluster-search-v2"

  elasticsearch = {
    hot = {
      autoscalign = {}
    }
    remote_cluster = [{
      deployment_id = ec_deployment.source_deployment.id
      alias         = ec_deployment.source_deployment.name
      ref_id        = ec_deployment.source_deployment.elasticsearch.0.ref_id
    }]
  }

  kibana = {}
}
```

### With tags

```hcl
data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

resource "ec_deployment" "with_tags" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }

  tags = {
    owner     = "elastic cloud"
    component = "search"
  }
}
```

### With configuration strategy

```hcl
data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

resource "ec_deployment" "with_tags" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch = {
    hot = {
      autoscaling = {}
    }
    strategy = [{
      type = "rolling_all"
    }]
  }

  tags = {
    owner     = "elastic cloud"
    component = "search"
  }
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Required) Elasticsearch Service (ESS) region where to create the deployment. For Elastic Cloud Enterprise (ECE) installations, set `"ece-region"`.

-> If you change the `region`, the resource will be destroyed and re-created.

* `deployment_template_id` - (Required) Deployment template identifier to create the deployment from. See the [full list](https://www.elastic.co/guide/en/cloud/current/ec-regions-templates-instances.html) of regions and deployment templates available in ESS.
* `version` - (Required) Elastic Stack version to use for all the deployment resources.

-> Read the [ESS stack version policy](https://www.elastic.co/guide/en/cloud/current/ec-version-policy.html#ec-version-policy-available) to understand which versions are available.

* `name` - (Optional) Name of the deployment.
* `alias` - (Optional) Deployment alias, affects the format of the resource URLs.
* `request_id` - (Optional) Request ID to set when you create the deployment. Use it only when previous attempts return an error and `request_id` is returned as part of the error.
* `elasticsearch` (Required) Elasticsearch cluster definition, can only be specified once. For multi-node Elasticsearch clusters, use multiple `topology` blocks.
* `kibana` (Optional) Kibana instance definition, can only be specified once.

-> **Note on disabling Kibana** While optional it is recommended deployments specify a Kibana block, since not doing so might cause issues when modifying or upgrading the deployment.

* `integrations_server` (Optional) Integrations Server instance definition, can only be specified once. It has replaced `apm` in stack version 8.0.0.
* `enterprise_search` (Optional) Enterprise Search server definition, can only be specified once. For multi-node Enterprise Search deployments, use multiple `topology` blocks.
* `apm` **DEPRECATED** (Optional) APM instance definition, can only be specified once. It should only be used with deployments with a version prior to 8.0.0.
* `traffic_filter` (Optional) List of traffic filter rule identifiers that will be applied to the deployment.
* `observability` (Optional) Observability settings that you can set to ship logs and metrics to a deployment. The target deployment can also be the current deployment itself.
* `tags` (Optional) Key value map of arbitrary string tags.

### Resources

!> **Warning on removing explicit topology objects** Due to current limitations, if a topology object is removed from the configuration, the removal won't trigger any changes since the field is optional and computed. There is no way to determine if the block was removed, which results in a _"sticky"_ topology configuration. To disable a topology element, set the `topology.size` to `"0g"`.

To create a valid deployment, you must specify at least the resource type `elasticsearch`. The supported resources are listed below.

A default topology from the deployment template is used for empty blocks: `elasticsearch {}`, `kibana {}`, `integrations_server {}`, `enterprise_search {}`. When a block is not set, the resource kind is not enabled in the deployment.

The `ec_deployment` resource will opt-out all the resources except Elasticsearch, which inherits the default topology from the deployment template. For example, the [I/O Optimized template includes an Elasticsearch cluster 8 GB memory x 2 availability zones](https://www.elastic.co/guide/en/cloud/current/ec-getting-started-profiles.html#ec-getting-started-profiles-io).

To customize the size or settings of the deployment resource, use the `topology` block within each resource kind block. The `topology` blocks are ordered lists and should be defined in the Terraform configuration in an ascending manner by alphabetical order of the `id` field.

#### Elasticsearch

The required `elasticsearch` block supports the following arguments:

* `topology` - (Optional) Can be set multiple times to compose complex topologies.
* `ref_id` - (Optional) Can be set on the Elasticsearch resource. The default value `main-elasticsearch` is recommended.
* `config` (Optional) Elasticsearch settings applied to all topologies unless overridden in the `topology` element.
* `remote_cluster` (Optional) Elasticsearch remote clusters to configure for the Elasticsearch resource. Can be set multiple times.
* `snapshot_source` (Optional) Restores data from a snapshot of another deployment.
* `extension` (Optional) Custom Elasticsearch bundles or plugins. Can be set multiple times.
* `autoscale` (Optional) Enable or disable autoscaling. Defaults to the setting coming from the deployment template. Accepted values are `"true"` or `"false"`.
* `trust_account` (Optional) The trust relationships with other ESS accounts.
* `trust_external` (Optional) The trust relationship with external entities (remote environments, remote accounts...).
* `strategy` (Optional) Choose the configuration strategy used to apply the changes.

##### Topology

To set up multi-node Elasticsearch clusters, you can set the topology block multiple times. Each block must specify the `id` field referencing the data tier name. This is particularly relevant for Elasticsearch clusters with multiple data tiers or Machine Learning.

-> To avoid infinite diff loops, topology blocks should be ordered alphabetically by the `topology.id` field. The order with the current data tiers at the time of this writing would be: "cold", "coordinating", "frozen", "hot_content", "master", "ml", "warm".

The optional `elasticsearch.topology` block supports the following arguments:

* `id` - (Required) Unique topology identifier. It generally refers to an Elasticsearch data tier, such as `hot_content`, `warm`, `cold`, `coordinating`, `frozen`, `ml` or `master`.
* `size` - (Optional) Amount in Gigabytes per topology element in the `"<size in GB>g"` notation. When omitted, it defaults to the deployment template value.
* `size_resource` - (Optional) Type of resource to which the size is assigned. Defaults to `"memory"`.
* `zone_count` - (Optional) Number of zones the instance type of the Elasticsearch cluster will span. This is used to set or unset HA on an Elasticsearch node type. When omitted, it defaults to the deployment template value.
* `node_type_data` - (Optional) The node type for the Elasticsearch cluster (data node).
* `node_type_master` - (Optional) The node type for the Elasticsearch cluster (master node).
* `node_type_ingest` - (Optional) The node type for the Elasticsearch cluster (ingest node).
* `node_type_ml` - (Optional) The node type for the Elasticsearch cluster (machine learning node).
* `autoscaling` - (Optional) Autoscaling policy defining the maximum and / or minimum total size for this topology element. For more information refer to the `autoscaling` block.

~> **Note when node_type_* fields set** After upgrading to a version that supports data tiers (7.10.0 or above), the `node_type_*` has no effect even if specified. The provider automatically migrates the `node_type_*` fields to the appropriate `node_roles` as set by the deployment template. After having upgraded to `7.10.0` or above, the fields should be removed from the terraform configuration, if explicitly configured.

##### Autoscaling

The optional `elasticsearch.autoscaling` block supports the following arguments:

* `min_size` - (Optional) Defines the minimum size the deployment will scale down to. When set, scale down will be enabled, please note that not all the tiers support this option.
* `min_size_resource` - (Optional) Defines the resource type the scale down will use (Defaults to `"memory"`).
* `max_size` - (Optional) Defines the maximum size the deployment will scale up to. When set, scaling up will be enabled. All tiers should support this option.
* `max_size_resource` - (Optional) Defines the resource type the scale up will use (Defaults to `"memory"`).

-> Note that none of these settings will take effect unless `elasticsearch.autoscale` is set to `"true"`.

Please refer to the [Deployment Autoscaling](https://www.elastic.co/guide/en/cloud/current/ec-autoscaling.html) documentation for an updated list of the Elasticsearch tiers supporting scale up and scale down.

##### Config

The optional `elasticsearch.config` block supports the following arguments:

* `plugins` - (Optional) List of Elasticsearch supported plugins. Check the Stack Pack version to see which plugins are supported for each version. This is currently only available from the UI and [ecctl](https://www.elastic.co/guide/en/ecctl/master/ecctl_stack_list.html).
* `user_settings_json` - (Optional) JSON-formatted user level `elasticsearch.yml` setting overrides.
* `user_settings_override_json` - (Optional) JSON-formatted admin (ECE) level `elasticsearch.yml` setting overrides.
* `user_settings_yaml` - (Optional) YAML-formatted user level `elasticsearch.yml` setting overrides.
* `user_settings_override_yaml` - (Optional) YAML-formatted admin (ECE) level `elasticsearch.yml` setting overrides.

##### Remote Cluster

The optional `elasticsearch.remote_cluster` block can be set multiple times. It represents one or multiple remote clusters to which the local Elasticsearch cluster connects for Cross Cluster Search and supports the following settings:

* `deployment_id` (Required) Remote deployment ID.
* `alias` (Required) Alias for the Cross Cluster Search binding.
* `ref_id` (Optional) Remote Elasticsearch `ref_id`. The default value `main-elasticsearch` is recommended.
* `skip_unavailable` (Optional) If true, skip the cluster during search when disconnected. Defaults to `false`.

##### Snapshot source

The optional `elasticsearch.snapshot_source` block, which restores data from a snapshot of another deployment, supports the following arguments:

* `source_elasticsearch_cluster_id` (Required) ID of the Elasticsearch cluster, not to be confused with the deployment ID, that will be used as the source of the snapshot. The Elasticsearch cluster must be in the same region and must have a compatible version of the Elastic Stack.
* `snapshot_name` (Optional) Name of the snapshot to restore. Use `__latest_success__` to get the most recent successful snapshot (Defaults to `__latest_success__`).

~> **Note on behavior** The `snapshot_source` block will not be saved in the Terraform state due to its transient nature. This means that whenever the `snapshot_source` block is set, a snapshot will **always be restored**, unless removed before running `terraform apply`.

##### Extension

The optional `elasticsearch.extension` block, allows custom plugins or bundles to be configured in the Elasticsearch cluster. It supports the following arguments:

* `name` (Required) Extension name.
* `type` (Required) Extension type, only `bundle` or `plugin` are supported.
* `version` (Required) Elasticsearch compatibility version. Bundles should specify major or minor versions with wildcards, such as `7.*` or `*` but **plugins must use full version notation down to the patch level**, such as `7.10.1` and wildcards are not allowed.
* `url` (Required) Bundle or plugin URL, the extension URL can be obtained from the `ec_deployment_extension.<name>.url` attribute or the API and cannot be a random HTTP address that is hosted elsewhere.

##### Trust Account

~> **Note on Computed Account Trusts** If your account has the default trust setting set to `Trust all my deployments (includes future deployments)`, the `trust_account` will always contain 1 element that sets up that trust. If that element is manually removed, the trust default will be unset for the cluster.

The optional `elasticsearch.trust_account` block, allows cross-account trust relationships to be set. It supports the following arguments:

* `account_id` (Required) The account identifier to establish the new trust with.
* `trust_all` (Optional) If true, all clusters in this account will by default be trusted and the `trust_allowlist` is ignored.
* `trust_allowlist` (Optional) The list of clusters to trust. Only used when `trust_all` is `false`.

##### Trust External

The optional `elasticsearch.trust_external` block, allows external trust relationships to be set. It supports the following arguments:

* `relationship_id` (Required) Identifier of the the trust relationship with external entities (remote environments, remote accounts...).
* `trust_all` (Optional) If true, all clusters in this external entity will be trusted and the `trust_allowlist` is ignored.
* `trust_allowlist` (Optional) The list of clusters to trust. Only used when `trust_all` is `false`.

##### Strategy

The optional `elasticsearch.strategy` allows you to choose the configuration strategy used to apply the changes. You do not need to change this setting unless you have a specific case where the `autodetect` does not cover your use case.

* `type` Set the type of configuration strategy [autodetect, grow_and_shrink, rolling_grow_and_shrink, rolling_all].
  * `autodetect` try to use the best associated with the type of change in the plan.
  * `grow_and_shrink` Add all nodes with the new changes before to stop any node.
  * `rolling_grow_and_shrink` Add nodes one by one replacing the existing ones when the new node is ready.
  * `rolling_all` Stop all nodes, perform the changes and start all nodes.

#### Kibana

The optional `kibana` block supports the following arguments:

* `topology` - (Optional) Can be set multiple times to compose complex topologies.
* `elasticsearch_cluster_ref_id` - (Optional) This field references the `ref_id` of the deployment Elasticsearch cluster. The default value `main-elasticsearch` is recommended.
* `ref_id` - (Optional) Can be set on the Kibana resource. The default value `main-kibana` is recommended.
* `config` (Optional) Kibana settings applied to all topologies unless overridden in the `topology` element.
* `instance_configuration_id` - (Optional) Default instance configuration of the deployment template. No need to change this value since Kibana has only one _instance type_.
* `size` - (Optional) Amount of memory (RAM) per topology element in the "<size in GB>g" notation. When omitted, it defaults to the deployment template value.
* `size_resource` - (Optional) Type of resource to which the size is assigned. Defaults to `"memory"`.
* `zone_count` - (Optional) Number of zones that the Kibana deployment will span. This is used to set HA. When omitted, it defaults to the deployment template value.

##### Config

The optional `kibana.config` block supports the following arguments:

* `user_settings_json` - (Optional) JSON-formatted user level `kibana.yml` setting overrides.
* `user_settings_override_json` - (Optional) JSON-formatted admin (ECE) level `kibana.yml` setting overrides.
* `user_settings_yaml` - (Optional) YAML-formatted user level `kibana.yml` setting overrides.
* `user_settings_override_yaml` - (Optional) YAML-formatted admin (ECE) level `kibana.yml` setting overrides.

#### Integrations Server

The optional `integrations_server` block supports the following arguments:

* `topology` - (Optional) Can be set multiple times to compose complex topologies.
* `elasticsearch_cluster_ref_id` - (Optional) This field references the `ref_id` of the deployment Elasticsearch cluster. The default value `main-elasticsearch` is recommended.
* `ref_id` - (Optional) Can be set on the Integrations Server resource. The default value `main-integrations_server` is recommended.
* `config` (Optional) Integrations Server settings applied to all topologies unless overridden in the `topology` element.
* `instance_configuration_id` - (Optional) Default instance configuration of the deployment template. No need to change this value since Integrations Server has only one _instance type_.
* `size` - (Optional) Amount of memory (RAM) per topology element in the "<size in GB>g" notation. When omitted, it defaults to the deployment template value.
* `size_resource` - (Optional) Type of resource to which the size is assigned. Defaults to `"memory"`.
* `zone_count` - (Optional) Number of zones that the Integrations Server deployment will span. This is used to set HA. When omitted, it defaults to the deployment template value.

##### Config

The optional `integrations_server.config` block supports the following arguments:

* `debug_enabled` - (Optional) Enable debug mode for the component. Defaults to `false`.

#### APM

The optional `apm` block supports the following arguments:

* `topology` - (Optional) Can be set multiple times to compose complex topologies.
* `elasticsearch_cluster_ref_id` - (Optional) This field references the `ref_id` of the deployment Elasticsearch cluster. The default value `main-elasticsearch` is recommended.
* `ref_id` - (Optional) Can be set on the APM resource. The default value `main-apm` is recommended.
* `config` (Optional) APM settings applied to all topologies unless overridden in the `topology` element.
* `instance_configuration_id` - (Optional) Default instance configuration of the deployment template. No need to change this value since APM has only one _instance type_.
* `size` - (Optional) Amount of memory (RAM) per topology element in the "<size in GB>g" notation. When omitted, it defaults to the deployment template value.
* `size_resource` - (Optional) Type of resource to which the size is assigned. Defaults to `"memory"`.
* `zone_count` - (Optional) Number of zones that the APM deployment will span. This is used to set HA. When omitted, it defaults to the deployment template value.

##### Config

The optional `apm.config` block supports the following arguments:

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
* `instance_configuration_id` - (Optional) Default instance configuration of the deployment template. To change it, use the [full list](https://www.elastic.co/guide/en/cloud/current/ec-regions-templates-instances.html) of regions and deployment templates available in ESS.
* `size` - (Optional) Amount of memory (RAM) per `topology` element in the "<size in GB>g" notation. When omitted, it defaults to the deployment template value.
* `size_resource` - (Optional) Type of resource to which the size is assigned. Defaults to `"memory"`.
* `zone_count` - (Optional) Number of zones that the Enterprise Search deployment will span. This is used to set HA. When omitted, it defaults to the deployment template value.

##### Config

The optional `enterprise_search.config` block supports the following arguments:

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
* `elasticsearch.#.region` - Elasticsearch region.
* `elasticsearch.#.cloud_id` - Encoded Elasticsearch credentials to use in Beats or Logstash. For more information, see [Configure Beats and Logstash with Cloud ID](https://www.elastic.co/guide/en/cloud/current/ec-cloud-id.html).
* `elasticsearch.#.http_endpoint` - Elasticsearch resource HTTP endpoint.
* `elasticsearch.#.https_endpoint` - Elasticsearch resource HTTPs endpoint.
* `elasticsearch.#.topology.#.instance_configuration_id` - instance configuration of the deployment topology element.
* `elasticsearch.#.topology.#.node_type_data` - Node type (data) for the Elasticsearch topology element.
* `elasticsearch.#.topology.#.node_type_master` - Node type (master) for the Elasticsearch topology element.
* `elasticsearch.#.topology.#.node_type_ingest` - Node type (ingest) for the Elasticsearch topology element.
* `elasticsearch.#.topology.#.node_type_ml` - Node type (machine learning) for the Elasticsearch topology element.
* `elasticsearch.#.topology.#.node_roles` - List of roles for the topology element. They are inferred from the deployment template.
* `elasticsearch.#.topology.#.autoscaling.#.policy_override_json` - Computed policy overrides set directly via the API or other clients.
* `elasticsearch.#.snapshot_source.#.source_elasticsearch_cluster_id` - ID of the Elasticsearch cluster that will be used as the source of the snapshot.
* `elasticsearch.#.snapshot_source.#.snapshot_name` - Name of the snapshot to restore.
* `kibana.#.resource_id` - Kibana resource unique identifier.
* `kibana.#.region` - Kibana region.
* `kibana.#.http_endpoint` - Kibana resource HTTP endpoint.
* `kibana.#.https_endpoint` - Kibana resource HTTPs endpoint.
* `integrations_server.#.resource_id` - Integrations Server resource unique identifier.
* `integrations_server.#.region` - Integrations Server region.
* `integrations_server.#.http_endpoint` - Integrations Server resource HTTP endpoint.
* `integrations_server.#.https_endpoint` - Integrations Server resource HTTPs endpoint.
* `integrations_server.#.fleet_https_endpoint` - HTTPs endpoint for Fleet Server.
* `integrations_server.#.apm_https_endpoint` - HTTPs endpoint for APM Server.
* `apm.#.resource_id` - APM resource unique identifier.
* `apm.#.region` - APM region.
* `apm.#.http_endpoint` - APM resource HTTP endpoint.
* `apm.#.https_endpoint` - APM resource HTTPs endpoint.
* `enterprise_search.#.resource_id` - Enterprise Search resource unique identifier.
* `enterprise_search.#.region` - Enterprise Search region.
* `enterprise_search.#.http_endpoint` - Enterprise Search resource HTTP endpoint.
* `enterprise_search.#.https_endpoint` - Enterprise Search resource HTTPs endpoint.
* `enterprise_search.#.topology.#.node_type_appserver` - Node type (Appserver) for the Enterprise Search topology element.
* `enterprise_search.#.topology.#.node_type_connector` - Node type (Connector) for the Enterprise Search topology element.
* `enterprise_search.#.topology.#.node_type_worker` - Node type (worker) for the Enterprise Search topology element.
* `observability.#.deployment_id` - Destination deployment ID for the shipped logs and monitoring metrics. Use `self` as destination deployment ID to target the current deployment.
* `observability.#.ref_id` - (Optional) Elasticsearch resource kind ref_id of the destination deployment.
* `observability.#.logs` - Enables or disables shipping logs. Defaults to true.
* `observability.#.metrics` - Enables or disables shipping metrics. Defaults to true.

## Import

~> **Note on deployment credentials** The `elastic` user credentials are only available whilst creating a deployment. Importing a deployment will not import the `elasticsearch_username` or `elasticsearch_password` attributes.

~> **Note on legacy (pre-slider) deployments** Importing deployments created prior to the addition of sliders in ECE or ESS, without being migrated to use sliders, is not supported.

~> **Note on pre 6.6.0 deployments** Importing deployments with a version lower than `6.6.0` is not supported.

~> **Note on deployments with topology user settings** Only deployments with global user settings (config) are supported. Make sure to migrate to global settings before importing.

Deployments can be imported using the `id`, for example:

```
$ terraform import ec_deployment.search 320b7b540dfc967a7a649c18e2fce4ed
```
