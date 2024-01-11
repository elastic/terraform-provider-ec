---
page_title: "Elastic Cloud: ec_deployment Resource"
description: |-
  Provides an Elastic Cloud deployment resource, which allows deployments to be created, updated, and deleted.
---

# Resource: ec_deployment

Provides an Elastic Cloud deployment resource, which allows deployments to be created, updated, and deleted.

~> **Note on traffic filters** If you use `traffic_filter` on an `ec_deployment`, Terraform will manage the full set of traffic rules for the deployment, and treat additional traffic filters as drift. For this reason, `traffic_filter` cannot be mixed with the `ec_deployment_traffic_filter_association` resource for a given deployment.

~> **Note on Elastic Stack versions** Using a version prior to `6.6.0` is not supported.

~> **Note on regions and deployment templates** Before you start, you might want to read about [Elastic Cloud deployments](https://www.elastic.co/guide/en/cloud/current/ec-create-deployment.html) and check the [full list](https://www.elastic.co/guide/en/cloud/current/ec-regions-templates-instances.html) of regions and deployment templates available in Elasticsearch Service (ESS).

## Example Usage

### Basic

```terraform
# Retrieve the latest stack pack version
data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

# Create an Elastic Cloud deployment
resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "my_example_deployment"

  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }

  kibana = {}

  enterprise_search = {}

  integrations_server = {}
}
```

### With config

`es.yaml`
```yaml
# My example YAML configuration for elasicsearch nodes
repositories.url.allowed_urls: ["http://www.example.org/root/*", "https://*.mydomain.com/*?*#*"]
```

`deployment.tf`:
```terraform
# Retrieve the latest stack pack version
data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

# Create an Elastic Cloud deployment
resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "my_example_deployment"

  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch = {
    hot = {
      autoscaling = {}
    }
    config = {
      user_settings_yaml = file("./es.yaml")
    }
  }

  kibana = {}

  enterprise_search = {}

  integrations_server = {}
}
```

### With autoscaling

```terraform
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

### With observability

```terraform
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

  tags = {
    "monitoring" = "source"
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

```terraform
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

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `deployment_template_id` (String) Deployment template identifier to create the deployment from. See the [full list](https://www.elastic.co/guide/en/cloud/current/ec-regions-templates-instances.html) of regions and deployment templates available in ESS.
- `elasticsearch` (Attributes) Elasticsearch cluster definition (see [below for nested schema](#nestedatt--elasticsearch))
- `region` (String) Elasticsearch Service (ESS) region where the deployment should be hosted. For Elastic Cloud Enterprise (ECE) installations, set to `"ece-region".
- `version` (String) Elastic Stack version to use for all of the deployment resources.

-> Read the [ESS stack version policy](https://www.elastic.co/guide/en/cloud/current/ec-version-policy.html#ec-version-policy-available) to understand which versions are available.

### Optional

- `alias` (String) Deployment alias, affects the format of the resource URLs.
- `apm` (Attributes) **DEPRECATED** APM cluster definition. This should only be used for deployments running a version lower than 8.0 (see [below for nested schema](#nestedatt--apm))
- `enterprise_search` (Attributes) Enterprise Search cluster definition. (see [below for nested schema](#nestedatt--enterprise_search))
- `integrations_server` (Attributes) Integrations Server cluster definition. Integrations Server replaces `apm` in Stack versions > 8.0 (see [below for nested schema](#nestedatt--integrations_server))
- `kibana` (Attributes) Kibana cluster definition.

-> **Note on disabling Kibana** While optional it is recommended deployments specify a Kibana block, since not doing so might cause issues when modifying or upgrading the deployment. (see [below for nested schema](#nestedatt--kibana))
- `migrate_to_latest_hardware` (Boolean) When true, updates deployment according to the latest deployment template values.
- `name` (String) Name for the deployment
- `observability` (Attributes) Observability settings that you can set to ship logs and metrics to a deployment. The target deployment can also be the current deployment itself by setting observability.deployment_id to `self`. (see [below for nested schema](#nestedatt--observability))
- `request_id` (String) Request ID to set when you create the deployment. Use it only when previous attempts return an error and `request_id` is returned as part of the error.
- `reset_elasticsearch_password` (Boolean) Explicitly resets the elasticsearch_password when true
- `tags` (Map of String) Optional map of deployment tags
- `traffic_filter` (Set of String) List of traffic filters rule identifiers that will be applied to the deployment.

### Read-Only

- `apm_secret_token` (String, Sensitive)
- `elasticsearch_password` (String, Sensitive) Password for authenticating to the Elasticsearch resource.

~> **Note on deployment credentials** The <code>elastic</code> user credentials are only available whilst creating a deployment. Importing a deployment will not import the <code>elasticsearch_username</code> or <code>elasticsearch_password</code> attributes.
~> **Note on deployment credentials in state** The <code>elastic</code> user credentials are stored in the state file as plain text. Please follow the official Terraform recommendations regarding senstaive data in state.
- `elasticsearch_username` (String) Username for authenticating to the Elasticsearch resource.
- `id` (String) Unique identifier of this deployment.

<a id="nestedatt--elasticsearch"></a>
### Nested Schema for `elasticsearch`

Required:

- `hot` (Attributes) 'hot' topology element (see [below for nested schema](#nestedatt--elasticsearch--hot))

Optional:

- `autoscale` (Boolean) Enable or disable autoscaling. Defaults to the setting coming from the deployment template.
- `cold` (Attributes) 'cold' topology element (see [below for nested schema](#nestedatt--elasticsearch--cold))
- `config` (Attributes) Elasticsearch settings which will be applied to all topologies (see [below for nested schema](#nestedatt--elasticsearch--config))
- `coordinating` (Attributes) 'coordinating' topology element (see [below for nested schema](#nestedatt--elasticsearch--coordinating))
- `extension` (Attributes Set) Optional Elasticsearch extensions such as custom bundles or plugins. (see [below for nested schema](#nestedatt--elasticsearch--extension))
- `frozen` (Attributes) 'frozen' topology element (see [below for nested schema](#nestedatt--elasticsearch--frozen))
- `keystore_contents` (Attributes Map) Keystore contents that are controlled by the deployment resource. (see [below for nested schema](#nestedatt--elasticsearch--keystore_contents))
- `master` (Attributes) 'master' topology element (see [below for nested schema](#nestedatt--elasticsearch--master))
- `ml` (Attributes) 'ml' topology element (see [below for nested schema](#nestedatt--elasticsearch--ml))
- `ref_id` (String) A human readable reference for the Elasticsearch resource. The default value `main-elasticsearch` is recommended.
- `remote_cluster` (Attributes Set) Optional Elasticsearch remote clusters to configure for the Elasticsearch resource, can be set multiple times (see [below for nested schema](#nestedatt--elasticsearch--remote_cluster))
- `snapshot` (Attributes) (ECE only) Snapshot configuration settings for an Elasticsearch cluster.

For ESS please use the [elasticstack_elasticsearch_snapshot_repository](https://registry.terraform.io/providers/elastic/elasticstack/latest/docs/resources/elasticsearch_snapshot_repository) resource from the [Elastic Stack terraform provider](https://registry.terraform.io/providers/elastic/elasticstack/latest). (see [below for nested schema](#nestedatt--elasticsearch--snapshot))
- `snapshot_source` (Attributes) Restores data from a snapshot of another deployment.

~> **Note on behavior** The <code>snapshot_source</code> block will not be saved in the Terraform state due to its transient nature. This means that whenever the <code>snapshot_source</code> block is set, a snapshot will **always be restored**, unless removed before running <code>terraform apply</code>. (see [below for nested schema](#nestedatt--elasticsearch--snapshot_source))
- `strategy` (String) Configuration strategy type autodetect, grow_and_shrink, rolling_grow_and_shrink, rolling_all
- `trust_account` (Attributes Set) Optional Elasticsearch account trust settings. (see [below for nested schema](#nestedatt--elasticsearch--trust_account))
- `trust_external` (Attributes Set) Optional Elasticsearch external trust settings. (see [below for nested schema](#nestedatt--elasticsearch--trust_external))
- `warm` (Attributes) 'warm' topology element (see [below for nested schema](#nestedatt--elasticsearch--warm))

Read-Only:

- `cloud_id` (String) The encoded Elasticsearch credentials to use in Beats or Logstash
- `http_endpoint` (String) The Elasticsearch resource HTTP endpoint
- `https_endpoint` (String) The Elasticsearch resource HTTPs endpoint
- `region` (String) The Elasticsearch resource region
- `resource_id` (String) The Elasticsearch resource unique identifier

<a id="nestedatt--elasticsearch--hot"></a>
### Nested Schema for `elasticsearch.hot`

Required:

- `autoscaling` (Attributes) Optional Elasticsearch autoscaling settings, such a maximum and minimum size and resources. (see [below for nested schema](#nestedatt--elasticsearch--hot--autoscaling))

Optional:

- `instance_configuration_id` (String) Instance Configuration ID of the topology element
- `instance_configuration_version` (Number) Instance Configuration version of the topology element
- `node_type_data` (String) The node type for the Elasticsearch Topology element (data node)
- `node_type_ingest` (String) The node type for the Elasticsearch Topology element (ingest node)
- `node_type_master` (String) The node type for the Elasticsearch Topology element (master node)
- `node_type_ml` (String) The node type for the Elasticsearch Topology element (machine learning node)
- `size` (String) Amount of "size_resource" per node in the "<size in GB>g" notation
- `size_resource` (String) Size type, defaults to "memory".
- `zone_count` (Number) Number of zones that the Elasticsearch cluster will span. This is used to set HA

Read-Only:

- `latest_instance_configuration_id` (String) Latest Instance Configuration ID available on the deployment template for the topology element
- `latest_instance_configuration_version` (Number) Latest version available for the Instance Configuration with the latest_instance_configuration_id
- `node_roles` (Set of String) The computed list of node roles for the current topology element

<a id="nestedatt--elasticsearch--hot--autoscaling"></a>
### Nested Schema for `elasticsearch.hot.autoscaling`

Optional:

- `max_size` (String) Maximum size value for the maximum autoscaling setting.
- `max_size_resource` (String) Maximum resource type for the maximum autoscaling setting.
- `min_size` (String) Minimum size value for the minimum autoscaling setting.
- `min_size_resource` (String) Minimum resource type for the minimum autoscaling setting.

Read-Only:

- `policy_override_json` (String) Computed policy overrides set directly via the API or other clients.



<a id="nestedatt--elasticsearch--cold"></a>
### Nested Schema for `elasticsearch.cold`

Required:

- `autoscaling` (Attributes) Optional Elasticsearch autoscaling settings, such a maximum and minimum size and resources. (see [below for nested schema](#nestedatt--elasticsearch--cold--autoscaling))

Optional:

- `instance_configuration_id` (String) Instance Configuration ID of the topology element
- `instance_configuration_version` (Number) Instance Configuration version of the topology element
- `node_type_data` (String) The node type for the Elasticsearch Topology element (data node)
- `node_type_ingest` (String) The node type for the Elasticsearch Topology element (ingest node)
- `node_type_master` (String) The node type for the Elasticsearch Topology element (master node)
- `node_type_ml` (String) The node type for the Elasticsearch Topology element (machine learning node)
- `size` (String) Amount of "size_resource" per node in the "<size in GB>g" notation
- `size_resource` (String) Size type, defaults to "memory".
- `zone_count` (Number) Number of zones that the Elasticsearch cluster will span. This is used to set HA

Read-Only:

- `latest_instance_configuration_id` (String) Latest Instance Configuration ID available on the deployment template for the topology element
- `latest_instance_configuration_version` (Number) Latest version available for the Instance Configuration with the latest_instance_configuration_id
- `node_roles` (Set of String) The computed list of node roles for the current topology element

<a id="nestedatt--elasticsearch--cold--autoscaling"></a>
### Nested Schema for `elasticsearch.cold.autoscaling`

Optional:

- `max_size` (String) Maximum size value for the maximum autoscaling setting.
- `max_size_resource` (String) Maximum resource type for the maximum autoscaling setting.
- `min_size` (String) Minimum size value for the minimum autoscaling setting.
- `min_size_resource` (String) Minimum resource type for the minimum autoscaling setting.

Read-Only:

- `policy_override_json` (String) Computed policy overrides set directly via the API or other clients.



<a id="nestedatt--elasticsearch--config"></a>
### Nested Schema for `elasticsearch.config`

Optional:

- `docker_image` (String) Overrides the docker image the Elasticsearch nodes will use. Note that this field will only work for internal users only.
- `plugins` (Set of String) List of Elasticsearch supported plugins, which vary from version to version. Check the Stack Pack version to see which plugins are supported for each version. This is currently only available from the UI and [ecctl](https://www.elastic.co/guide/en/ecctl/master/ecctl_stack_list.html)
- `user_settings_json` (String) JSON-formatted user level "elasticsearch.yml" setting overrides
- `user_settings_override_json` (String) JSON-formatted admin (ECE) level "elasticsearch.yml" setting overrides
- `user_settings_override_yaml` (String) YAML-formatted admin (ECE) level "elasticsearch.yml" setting overrides
- `user_settings_yaml` (String) YAML-formatted user level "elasticsearch.yml" setting overrides


<a id="nestedatt--elasticsearch--coordinating"></a>
### Nested Schema for `elasticsearch.coordinating`

Required:

- `autoscaling` (Attributes) Optional Elasticsearch autoscaling settings, such a maximum and minimum size and resources. (see [below for nested schema](#nestedatt--elasticsearch--coordinating--autoscaling))

Optional:

- `instance_configuration_id` (String) Instance Configuration ID of the topology element
- `instance_configuration_version` (Number) Instance Configuration version of the topology element
- `node_type_data` (String) The node type for the Elasticsearch Topology element (data node)
- `node_type_ingest` (String) The node type for the Elasticsearch Topology element (ingest node)
- `node_type_master` (String) The node type for the Elasticsearch Topology element (master node)
- `node_type_ml` (String) The node type for the Elasticsearch Topology element (machine learning node)
- `size` (String) Amount of "size_resource" per node in the "<size in GB>g" notation
- `size_resource` (String) Size type, defaults to "memory".
- `zone_count` (Number) Number of zones that the Elasticsearch cluster will span. This is used to set HA

Read-Only:

- `latest_instance_configuration_id` (String) Latest Instance Configuration ID available on the deployment template for the topology element
- `latest_instance_configuration_version` (Number) Latest version available for the Instance Configuration with the latest_instance_configuration_id
- `node_roles` (Set of String) The computed list of node roles for the current topology element

<a id="nestedatt--elasticsearch--coordinating--autoscaling"></a>
### Nested Schema for `elasticsearch.coordinating.autoscaling`

Optional:

- `max_size` (String) Maximum size value for the maximum autoscaling setting.
- `max_size_resource` (String) Maximum resource type for the maximum autoscaling setting.
- `min_size` (String) Minimum size value for the minimum autoscaling setting.
- `min_size_resource` (String) Minimum resource type for the minimum autoscaling setting.

Read-Only:

- `policy_override_json` (String) Computed policy overrides set directly via the API or other clients.



<a id="nestedatt--elasticsearch--extension"></a>
### Nested Schema for `elasticsearch.extension`

Required:

- `name` (String) Extension name.
- `type` (String) Extension type, only `bundle` or `plugin` are supported.
- `url` (String) Bundle or plugin URL, the extension URL can be obtained from the `ec_deployment_extension.<name>.url` attribute or the API and cannot be a random HTTP address that is hosted elsewhere.
- `version` (String) Elasticsearch compatibility version. Bundles should specify major or minor versions with wildcards, such as `7.*` or `*` but **plugins must use full version notation down to the patch level**, such as `7.10.1` and wildcards are not allowed.


<a id="nestedatt--elasticsearch--frozen"></a>
### Nested Schema for `elasticsearch.frozen`

Required:

- `autoscaling` (Attributes) Optional Elasticsearch autoscaling settings, such a maximum and minimum size and resources. (see [below for nested schema](#nestedatt--elasticsearch--frozen--autoscaling))

Optional:

- `instance_configuration_id` (String) Instance Configuration ID of the topology element
- `instance_configuration_version` (Number) Instance Configuration version of the topology element
- `node_type_data` (String) The node type for the Elasticsearch Topology element (data node)
- `node_type_ingest` (String) The node type for the Elasticsearch Topology element (ingest node)
- `node_type_master` (String) The node type for the Elasticsearch Topology element (master node)
- `node_type_ml` (String) The node type for the Elasticsearch Topology element (machine learning node)
- `size` (String) Amount of "size_resource" per node in the "<size in GB>g" notation
- `size_resource` (String) Size type, defaults to "memory".
- `zone_count` (Number) Number of zones that the Elasticsearch cluster will span. This is used to set HA

Read-Only:

- `latest_instance_configuration_id` (String) Latest Instance Configuration ID available on the deployment template for the topology element
- `latest_instance_configuration_version` (Number) Latest version available for the Instance Configuration with the latest_instance_configuration_id
- `node_roles` (Set of String) The computed list of node roles for the current topology element

<a id="nestedatt--elasticsearch--frozen--autoscaling"></a>
### Nested Schema for `elasticsearch.frozen.autoscaling`

Optional:

- `max_size` (String) Maximum size value for the maximum autoscaling setting.
- `max_size_resource` (String) Maximum resource type for the maximum autoscaling setting.
- `min_size` (String) Minimum size value for the minimum autoscaling setting.
- `min_size_resource` (String) Minimum resource type for the minimum autoscaling setting.

Read-Only:

- `policy_override_json` (String) Computed policy overrides set directly via the API or other clients.



<a id="nestedatt--elasticsearch--keystore_contents"></a>
### Nested Schema for `elasticsearch.keystore_contents`

Required:

- `value` (String, Sensitive) Secret value. This can either be a string or a JSON object that is stored as a JSON string in the keystore.

Optional:

- `as_file` (Boolean) If true, the secret is handled as a file. Otherwise, it's handled as a plain string.


<a id="nestedatt--elasticsearch--master"></a>
### Nested Schema for `elasticsearch.master`

Required:

- `autoscaling` (Attributes) Optional Elasticsearch autoscaling settings, such a maximum and minimum size and resources. (see [below for nested schema](#nestedatt--elasticsearch--master--autoscaling))

Optional:

- `instance_configuration_id` (String) Instance Configuration ID of the topology element
- `instance_configuration_version` (Number) Instance Configuration version of the topology element
- `node_type_data` (String) The node type for the Elasticsearch Topology element (data node)
- `node_type_ingest` (String) The node type for the Elasticsearch Topology element (ingest node)
- `node_type_master` (String) The node type for the Elasticsearch Topology element (master node)
- `node_type_ml` (String) The node type for the Elasticsearch Topology element (machine learning node)
- `size` (String) Amount of "size_resource" per node in the "<size in GB>g" notation
- `size_resource` (String) Size type, defaults to "memory".
- `zone_count` (Number) Number of zones that the Elasticsearch cluster will span. This is used to set HA

Read-Only:

- `latest_instance_configuration_id` (String) Latest Instance Configuration ID available on the deployment template for the topology element
- `latest_instance_configuration_version` (Number) Latest version available for the Instance Configuration with the latest_instance_configuration_id
- `node_roles` (Set of String) The computed list of node roles for the current topology element

<a id="nestedatt--elasticsearch--master--autoscaling"></a>
### Nested Schema for `elasticsearch.master.autoscaling`

Optional:

- `max_size` (String) Maximum size value for the maximum autoscaling setting.
- `max_size_resource` (String) Maximum resource type for the maximum autoscaling setting.
- `min_size` (String) Minimum size value for the minimum autoscaling setting.
- `min_size_resource` (String) Minimum resource type for the minimum autoscaling setting.

Read-Only:

- `policy_override_json` (String) Computed policy overrides set directly via the API or other clients.



<a id="nestedatt--elasticsearch--ml"></a>
### Nested Schema for `elasticsearch.ml`

Required:

- `autoscaling` (Attributes) Optional Elasticsearch autoscaling settings, such a maximum and minimum size and resources. (see [below for nested schema](#nestedatt--elasticsearch--ml--autoscaling))

Optional:

- `instance_configuration_id` (String) Instance Configuration ID of the topology element
- `instance_configuration_version` (Number) Instance Configuration version of the topology element
- `node_type_data` (String) The node type for the Elasticsearch Topology element (data node)
- `node_type_ingest` (String) The node type for the Elasticsearch Topology element (ingest node)
- `node_type_master` (String) The node type for the Elasticsearch Topology element (master node)
- `node_type_ml` (String) The node type for the Elasticsearch Topology element (machine learning node)
- `size` (String) Amount of "size_resource" per node in the "<size in GB>g" notation
- `size_resource` (String) Size type, defaults to "memory".
- `zone_count` (Number) Number of zones that the Elasticsearch cluster will span. This is used to set HA

Read-Only:

- `latest_instance_configuration_id` (String) Latest Instance Configuration ID available on the deployment template for the topology element
- `latest_instance_configuration_version` (Number) Latest version available for the Instance Configuration with the latest_instance_configuration_id
- `node_roles` (Set of String) The computed list of node roles for the current topology element

<a id="nestedatt--elasticsearch--ml--autoscaling"></a>
### Nested Schema for `elasticsearch.ml.autoscaling`

Optional:

- `max_size` (String) Maximum size value for the maximum autoscaling setting.
- `max_size_resource` (String) Maximum resource type for the maximum autoscaling setting.
- `min_size` (String) Minimum size value for the minimum autoscaling setting.
- `min_size_resource` (String) Minimum resource type for the minimum autoscaling setting.

Read-Only:

- `policy_override_json` (String) Computed policy overrides set directly via the API or other clients.



<a id="nestedatt--elasticsearch--remote_cluster"></a>
### Nested Schema for `elasticsearch.remote_cluster`

Required:

- `alias` (String) Alias for this Cross Cluster Search binding
- `deployment_id` (String) Remote deployment ID

Optional:

- `ref_id` (String) Remote elasticsearch "ref_id", it is best left to the default value
- `skip_unavailable` (Boolean) If true, skip the cluster during search when disconnected


<a id="nestedatt--elasticsearch--snapshot"></a>
### Nested Schema for `elasticsearch.snapshot`

Required:

- `enabled` (Boolean) Indicates if Snapshotting is enabled.

Optional:

- `repository` (Attributes) Snapshot repository configuration (see [below for nested schema](#nestedatt--elasticsearch--snapshot--repository))

<a id="nestedatt--elasticsearch--snapshot--repository"></a>
### Nested Schema for `elasticsearch.snapshot.repository`

Optional:

- `reference` (Attributes) Cluster snapshot reference repository settings, containing the repository name in ECE fashion (see [below for nested schema](#nestedatt--elasticsearch--snapshot--repository--reference))

<a id="nestedatt--elasticsearch--snapshot--repository--reference"></a>
### Nested Schema for `elasticsearch.snapshot.repository.reference`

Required:

- `repository_name` (String) ECE snapshot repository name, from the '/platform/configuration/snapshots/repositories' endpoint




<a id="nestedatt--elasticsearch--snapshot_source"></a>
### Nested Schema for `elasticsearch.snapshot_source`

Required:

- `source_elasticsearch_cluster_id` (String) ID of the Elasticsearch cluster that will be used as the source of the snapshot

Optional:

- `snapshot_name` (String) Name of the snapshot to restore. Use '__latest_success__' to get the most recent successful snapshot.


<a id="nestedatt--elasticsearch--trust_account"></a>
### Nested Schema for `elasticsearch.trust_account`

Required:

- `account_id` (String) The ID of the Account.
- `trust_all` (Boolean) If true, all clusters in this account will by default be trusted and the `trust_allowlist` is ignored.

Optional:

- `trust_allowlist` (Set of String) The list of clusters to trust. Only used when `trust_all` is false.


<a id="nestedatt--elasticsearch--trust_external"></a>
### Nested Schema for `elasticsearch.trust_external`

Required:

- `relationship_id` (String) The ID of the external trust relationship.
- `trust_all` (Boolean) If true, all clusters in this account will by default be trusted and the `trust_allowlist` is ignored.

Optional:

- `trust_allowlist` (Set of String) The list of clusters to trust. Only used when `trust_all` is false.


<a id="nestedatt--elasticsearch--warm"></a>
### Nested Schema for `elasticsearch.warm`

Required:

- `autoscaling` (Attributes) Optional Elasticsearch autoscaling settings, such a maximum and minimum size and resources. (see [below for nested schema](#nestedatt--elasticsearch--warm--autoscaling))

Optional:

- `instance_configuration_id` (String) Instance Configuration ID of the topology element
- `instance_configuration_version` (Number) Instance Configuration version of the topology element
- `node_type_data` (String) The node type for the Elasticsearch Topology element (data node)
- `node_type_ingest` (String) The node type for the Elasticsearch Topology element (ingest node)
- `node_type_master` (String) The node type for the Elasticsearch Topology element (master node)
- `node_type_ml` (String) The node type for the Elasticsearch Topology element (machine learning node)
- `size` (String) Amount of "size_resource" per node in the "<size in GB>g" notation
- `size_resource` (String) Size type, defaults to "memory".
- `zone_count` (Number) Number of zones that the Elasticsearch cluster will span. This is used to set HA

Read-Only:

- `latest_instance_configuration_id` (String) Latest Instance Configuration ID available on the deployment template for the topology element
- `latest_instance_configuration_version` (Number) Latest version available for the Instance Configuration with the latest_instance_configuration_id
- `node_roles` (Set of String) The computed list of node roles for the current topology element

<a id="nestedatt--elasticsearch--warm--autoscaling"></a>
### Nested Schema for `elasticsearch.warm.autoscaling`

Optional:

- `max_size` (String) Maximum size value for the maximum autoscaling setting.
- `max_size_resource` (String) Maximum resource type for the maximum autoscaling setting.
- `min_size` (String) Minimum size value for the minimum autoscaling setting.
- `min_size_resource` (String) Minimum resource type for the minimum autoscaling setting.

Read-Only:

- `policy_override_json` (String) Computed policy overrides set directly via the API or other clients.




<a id="nestedatt--apm"></a>
### Nested Schema for `apm`

Optional:

- `config` (Attributes) Optionally define the Apm configuration options for the APM Server (see [below for nested schema](#nestedatt--apm--config))
- `elasticsearch_cluster_ref_id` (String)
- `instance_configuration_id` (String)
- `instance_configuration_version` (Number)
- `ref_id` (String)
- `size` (String)
- `size_resource` (String) Optional size type, defaults to "memory".
- `zone_count` (Number)

Read-Only:

- `http_endpoint` (String)
- `https_endpoint` (String)
- `latest_instance_configuration_id` (String)
- `latest_instance_configuration_version` (Number)
- `region` (String)
- `resource_id` (String)

<a id="nestedatt--apm--config"></a>
### Nested Schema for `apm.config`

Optional:

- `debug_enabled` (Boolean) Optionally enable debug mode for APM servers - defaults to false
- `docker_image` (String) Optionally override the docker image the APM nodes will use. This option will not work in ESS customers and should only be changed if you know what you're doing.
- `user_settings_json` (String) An arbitrary JSON object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_yaml' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (This field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)
- `user_settings_override_json` (String) An arbitrary JSON object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_yaml' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)
- `user_settings_override_yaml` (String) An arbitrary YAML object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_json' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)
- `user_settings_yaml` (String) An arbitrary YAML object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_json' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (These field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)



<a id="nestedatt--enterprise_search"></a>
### Nested Schema for `enterprise_search`

Optional:

- `config` (Attributes) Optionally define the Enterprise Search configuration options for the Enterprise Search Server (see [below for nested schema](#nestedatt--enterprise_search--config))
- `elasticsearch_cluster_ref_id` (String)
- `instance_configuration_id` (String)
- `instance_configuration_version` (Number)
- `ref_id` (String)
- `size` (String)
- `size_resource` (String) Optional size type, defaults to "memory".
- `zone_count` (Number)

Read-Only:

- `http_endpoint` (String)
- `https_endpoint` (String)
- `latest_instance_configuration_id` (String)
- `latest_instance_configuration_version` (Number)
- `node_type_appserver` (Boolean)
- `node_type_connector` (Boolean)
- `node_type_worker` (Boolean)
- `region` (String)
- `resource_id` (String)

<a id="nestedatt--enterprise_search--config"></a>
### Nested Schema for `enterprise_search.config`

Optional:

- `docker_image` (String) Optionally override the docker image the Enterprise Search nodes will use. Note that this field will only work for internal users only.
- `user_settings_json` (String) An arbitrary JSON object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_yaml' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (This field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)
- `user_settings_override_json` (String) An arbitrary JSON object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_yaml' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)
- `user_settings_override_yaml` (String) An arbitrary YAML object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_json' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)
- `user_settings_yaml` (String) An arbitrary YAML object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_json' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (These field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)



<a id="nestedatt--integrations_server"></a>
### Nested Schema for `integrations_server`

Optional:

- `config` (Attributes) Optionally define the Integrations Server configuration options for the IntegrationsServer Server (see [below for nested schema](#nestedatt--integrations_server--config))
- `elasticsearch_cluster_ref_id` (String)
- `endpoints` (Object) URLs for the accessing the Fleet and APM API's within this Integrations Server resource. (see [below for nested schema](#nestedatt--integrations_server--endpoints))
- `instance_configuration_id` (String)
- `instance_configuration_version` (Number)
- `ref_id` (String)
- `size` (String)
- `size_resource` (String) Optional size type, defaults to "memory".
- `zone_count` (Number)

Read-Only:

- `http_endpoint` (String)
- `https_endpoint` (String)
- `latest_instance_configuration_id` (String)
- `latest_instance_configuration_version` (Number)
- `region` (String)
- `resource_id` (String)

<a id="nestedatt--integrations_server--config"></a>
### Nested Schema for `integrations_server.config`

Optional:

- `debug_enabled` (Boolean) Optionally enable debug mode for Integrations Server instances - defaults to false
- `docker_image` (String) Optionally override the docker image the Integrations Server nodes will use. Note that this field will only work for internal users only.
- `user_settings_json` (String) An arbitrary JSON object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_yaml' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (This field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)
- `user_settings_override_json` (String) An arbitrary JSON object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_yaml' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)
- `user_settings_override_yaml` (String) An arbitrary YAML object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_json' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)
- `user_settings_yaml` (String) An arbitrary YAML object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_json' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (These field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)


<a id="nestedatt--integrations_server--endpoints"></a>
### Nested Schema for `integrations_server.endpoints`

Optional:

- `apm` (String)
- `fleet` (String)



<a id="nestedatt--kibana"></a>
### Nested Schema for `kibana`

Optional:

- `config` (Attributes) Optionally define the Kibana configuration options for the Kibana Server (see [below for nested schema](#nestedatt--kibana--config))
- `elasticsearch_cluster_ref_id` (String)
- `instance_configuration_id` (String)
- `instance_configuration_version` (Number)
- `ref_id` (String)
- `size` (String)
- `size_resource` (String) Optional size type, defaults to "memory".
- `zone_count` (Number)

Read-Only:

- `http_endpoint` (String)
- `https_endpoint` (String)
- `latest_instance_configuration_id` (String)
- `latest_instance_configuration_version` (Number)
- `region` (String)
- `resource_id` (String)

<a id="nestedatt--kibana--config"></a>
### Nested Schema for `kibana.config`

Optional:

- `docker_image` (String) Optionally override the docker image the Kibana nodes will use. Note that this field will only work for internal users only.
- `user_settings_json` (String) An arbitrary JSON object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_yaml' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (This field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)
- `user_settings_override_json` (String) An arbitrary JSON object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_yaml' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)
- `user_settings_override_yaml` (String) An arbitrary YAML object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_json' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)
- `user_settings_yaml` (String) An arbitrary YAML object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_json' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (These field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)



<a id="nestedatt--observability"></a>
### Nested Schema for `observability`

Required:

- `deployment_id` (String)

Optional:

- `logs` (Boolean)
- `metrics` (Boolean)
- `ref_id` (String)

## Import

~> **Note on deployment credentials** The `elastic` user credentials are only available whilst creating a deployment. Importing a deployment will not import the `elasticsearch_username` or `elasticsearch_password` attributes.

~> **Note on legacy (pre-slider) deployments** Importing deployments created prior to the addition of sliders in ECE or ESS, without being migrated to use sliders, is not supported.

~> **Note on pre 6.6.0 deployments** Importing deployments with a version lower than `6.6.0` is not supported.

~> **Note on deployments with topology user settings** Only deployments with global user settings (config) are supported. Make sure to migrate to global settings before importing.

Deployments can be imported using the `id`, for example:

```shell
terraform import ec_deployment.search 320b7b540dfc967a7a649c18e2fce4ed
```
