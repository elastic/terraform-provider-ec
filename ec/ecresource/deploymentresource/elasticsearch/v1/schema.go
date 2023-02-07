// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package v1

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// These constants are only used to determine whether or not a dedicated
// tier of masters or ingest (coordinating) nodes are set.
const (
	dataTierRolePrefix   = "data_"
	ingestDataTierRole   = "ingest"
	masterDataTierRole   = "master"
	autodetect           = "autodetect"
	growAndShrink        = "grow_and_shrink"
	rollingGrowAndShrink = "rolling_grow_and_shrink"
	rollingAll           = "rolling_all"
)

// List of update strategies availables.
var strategiesList = []string{
	autodetect, growAndShrink, rollingGrowAndShrink, rollingAll,
}

func ElasticsearchSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Description: "Required Elasticsearch resource definition",
		Required:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"autoscale": schema.StringAttribute{
					Description: `Enable or disable autoscaling. Defaults to the setting coming from the deployment template. Accepted values are "true" or "false".`,
					Computed:    true,
					Optional:    true,
				},
				"ref_id": schema.StringAttribute{
					Description: "Optional ref_id to set on the Elasticsearch resource",
					Optional:    true,
					Computed:    true,
				},
				"resource_id": schema.StringAttribute{
					Description: "The Elasticsearch resource unique identifier",
					Computed:    true,
				},
				"region": schema.StringAttribute{
					Description: "The Elasticsearch resource region",
					Computed:    true,
				},
				"cloud_id": schema.StringAttribute{
					Description: "The encoded Elasticsearch credentials to use in Beats or Logstash",
					Computed:    true,
				},
				"http_endpoint": schema.StringAttribute{
					Description: "The Elasticsearch resource HTTP endpoint",
					Computed:    true,
				},
				"https_endpoint": schema.StringAttribute{
					Description: "The Elasticsearch resource HTTPs endpoint",
					Computed:    true,
				},
				"topology": ElasticsearchTopologySchema(),

				"trust_account": ElasticsearchTrustAccountSchema(),

				"trust_external": ElasticsearchTrustExternalSchema(),

				"config": ElasticsearchConfigSchema(),

				"remote_cluster": ElasticsearchRemoteClusterSchema(),

				"snapshot_source": ElasticsearchSnapshotSourceSchema(),

				"extension": ElasticsearchExtensionSchema(),

				"strategy": ElasticsearchStrategySchema(),
			},
		},
	}
}

func ElasticsearchConfigSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Description: `Optional Elasticsearch settings which will be applied to all topologies unless overridden on the topology element`,
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"docker_image": schema.StringAttribute{
					Description: "Optionally override the docker image the Elasticsearch nodes will use. This option will not work in ESS customers and should only be changed if you know what you're doing.",
					Optional:    true,
				},
				"plugins": schema.SetAttribute{
					ElementType: types.StringType,
					Description: "List of Elasticsearch supported plugins, which vary from version to version. Check the Stack Pack version to see which plugins are supported for each version. This is currently only available from the UI and [ecctl](https://www.elastic.co/guide/en/ecctl/master/ecctl_stack_list.html)",
					Optional:    true,
				},
				"user_settings_json": schema.StringAttribute{
					Description: `JSON-formatted user level "elasticsearch.yml" setting overrides`,
					Optional:    true,
				},
				"user_settings_override_json": schema.StringAttribute{
					Description: `JSON-formatted admin (ECE) level "elasticsearch.yml" setting overrides`,
					Optional:    true,
				},
				"user_settings_yaml": schema.StringAttribute{
					Description: `YAML-formatted user level "elasticsearch.yml" setting overrides`,
					Optional:    true,
				},
				"user_settings_override_yaml": schema.StringAttribute{
					Description: `YAML-formatted admin (ECE) level "elasticsearch.yml" setting overrides`,
					Optional:    true,
				},
			},
		},
	}
}

func ElasticsearchTopologySchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Computed:    true,
		Optional:    true,
		Description: `Optional topology element which must be set once but can be set multiple times to compose complex topologies`,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Description: `Required topology ID from the deployment template`,
					Required:    true,
				},
				"instance_configuration_id": schema.StringAttribute{
					Description: `Computed Instance Configuration ID of the topology element`,
					Computed:    true,
				},
				"size": schema.StringAttribute{
					Description: `Optional amount of memory per node in the "<size in GB>g" notation`,
					Computed:    true,
					Optional:    true,
				},
				"size_resource": schema.StringAttribute{
					Description: `Optional size type, defaults to "memory".`,
					Optional:    true,
					Computed:    true,
				},
				"zone_count": schema.Int64Attribute{
					Description: `Optional number of zones that the Elasticsearch cluster will span. This is used to set HA`,
					Computed:    true,
					Optional:    true,
				},
				"node_type_data": schema.StringAttribute{
					Description: `The node type for the Elasticsearch Topology element (data node)`,
					Computed:    true,
					Optional:    true,
				},
				"node_type_master": schema.StringAttribute{
					Description: `The node type for the Elasticsearch Topology element (master node)`,
					Computed:    true,
					Optional:    true,
				},
				"node_type_ingest": schema.StringAttribute{
					Description: `The node type for the Elasticsearch Topology element (ingest node)`,
					Computed:    true,
					Optional:    true,
				},
				"node_type_ml": schema.StringAttribute{
					Description: `The node type for the Elasticsearch Topology element (machine learning node)`,
					Computed:    true,
					Optional:    true,
				},
				"node_roles": schema.SetAttribute{
					ElementType: types.StringType,
					Description: `The computed list of node roles for the current topology element`,
					Computed:    true,
				},
				"autoscaling": ElasticsearchTopologyAutoscalingSchema(),
				"config":      ElasticsearchTopologyConfigSchema(),
			},
		},
	}
}

func ElasticsearchTopologyAutoscalingSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Description: "Optional Elasticsearch autoscaling settings, such a maximum and minimum size and resources.",
		Optional:    true,
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"max_size_resource": schema.StringAttribute{
					Description: "Maximum resource type for the maximum autoscaling setting.",
					Optional:    true,
					Computed:    true,
				},
				"max_size": schema.StringAttribute{
					Description: "Maximum size value for the maximum autoscaling setting.",
					Optional:    true,
					Computed:    true,
				},
				"min_size_resource": schema.StringAttribute{
					Description: "Minimum resource type for the minimum autoscaling setting.",
					Optional:    true,
					Computed:    true,
				},
				"min_size": schema.StringAttribute{
					Description: "Minimum size value for the minimum autoscaling setting.",
					Optional:    true,
					Computed:    true,
				},
				"policy_override_json": schema.StringAttribute{
					Description: "Computed policy overrides set directly via the API or other clients.",
					Computed:    true,
				},
			},
		},
	}
}

func ElasticsearchRemoteClusterSchema() schema.Attribute {
	return schema.SetNestedAttribute{
		Description: "Optional Elasticsearch remote clusters to configure for the Elasticsearch resource, can be set multiple times",
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"deployment_id": schema.StringAttribute{
					Description: "Remote deployment ID",
					Required:    true,
				},
				"alias": schema.StringAttribute{
					Description: "Alias for this Cross Cluster Search binding",
					Required:    true,
				},
				"ref_id": schema.StringAttribute{
					Description: `Remote elasticsearch "ref_id", it is best left to the default value`,
					Computed:    true,
					Optional:    true,
				},
				"skip_unavailable": schema.BoolAttribute{
					Description: "If true, skip the cluster during search when disconnected",
					Optional:    true,
				},
			},
		},
	}
}

func ElasticsearchSnapshotSourceSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Description: "Optional snapshot source settings. Restore data from a snapshot of another deployment.",
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"source_elasticsearch_cluster_id": schema.StringAttribute{
					Description: "ID of the Elasticsearch cluster that will be used as the source of the snapshot",
					Required:    true,
				},
				"snapshot_name": schema.StringAttribute{
					Description: "Name of the snapshot to restore. Use '__latest_success__' to get the most recent successful snapshot.",
					Optional:    true,
					Computed:    true,
				},
			},
		},
	}
}

func ElasticsearchExtensionSchema() schema.Attribute {
	return schema.SetNestedAttribute{
		Description: "Optional Elasticsearch extensions such as custom bundles or plugins.",
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "Extension name.",
					Required:    true,
				},
				"type": schema.StringAttribute{
					Description: "Extension type, only `bundle` or `plugin` are supported.",
					Required:    true,
				},
				"version": schema.StringAttribute{
					Description: "Elasticsearch compatibility version. Bundles should specify major or minor versions with wildcards, such as `7.*` or `*` but **plugins must use full version notation down to the patch level**, such as `7.10.1` and wildcards are not allowed.",
					Required:    true,
				},
				"url": schema.StringAttribute{
					Description: "Bundle or plugin URL, the extension URL can be obtained from the `ec_deployment_extension.<name>.url` attribute or the API and cannot be a random HTTP address that is hosted elsewhere.",
					Required:    true,
				},
			},
		},
	}
}

func ElasticsearchTrustAccountSchema() schema.Attribute {
	return schema.SetNestedAttribute{
		Description: "Optional Elasticsearch account trust settings.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"account_id": schema.StringAttribute{
					Description: "The ID of the Account.",
					Required:    true,
				},
				"trust_all": schema.BoolAttribute{
					Description: "If true, all clusters in this account will by default be trusted and the `trust_allowlist` is ignored.",
					Required:    true,
				},
				"trust_allowlist": schema.SetAttribute{
					Description: "The list of clusters to trust. Only used when `trust_all` is false.",
					ElementType: types.StringType,
					Optional:    true,
				},
			},
		},
		Computed: true,
		Optional: true,
	}
}

func ElasticsearchTrustExternalSchema() schema.Attribute {
	return schema.SetNestedAttribute{
		Description: "Optional Elasticsearch external trust settings.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"relationship_id": schema.StringAttribute{
					Description: "The ID of the external trust relationship.",
					Required:    true,
				},
				"trust_all": schema.BoolAttribute{
					Description: "If true, all clusters in this account will by default be trusted and the `trust_allowlist` is ignored.",
					Required:    true,
				},
				"trust_allowlist": schema.SetAttribute{
					Description: "The list of clusters to trust. Only used when `trust_all` is false.",
					ElementType: types.StringType,
					Optional:    true,
				},
			},
		},
		Computed: true,
		Optional: true,
	}
}

func ElasticsearchStrategySchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Description: "Configuration strategy settings.",
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"type": schema.StringAttribute{
					Description: "Configuration strategy type " + strings.Join(strategiesList, ", "),
					Required:    true,
				},
			},
		},
	}
}

func ElasticsearchTopologyConfigSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Description: `Computed read-only configuration to avoid unsetting plan settings from 'topology.elasticsearch'`,
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"plugins": schema.SetAttribute{
					ElementType: types.StringType,
					Description: "List of Elasticsearch supported plugins, which vary from version to version. Check the Stack Pack version to see which plugins are supported for each version. This is currently only available from the UI and [ecctl](https://www.elastic.co/guide/en/ecctl/master/ecctl_stack_list.html)",
					Computed:    true,
				},
				"user_settings_json": schema.StringAttribute{
					Description: `JSON-formatted user level "elasticsearch.yml" setting overrides`,
					Computed:    true,
				},
				"user_settings_override_json": schema.StringAttribute{
					Description: `JSON-formatted admin (ECE) level "elasticsearch.yml" setting overrides`,
					Computed:    true,
				},
				"user_settings_yaml": schema.StringAttribute{
					Description: `YAML-formatted user level "elasticsearch.yml" setting overrides`,
					Computed:    true,
				},
				"user_settings_override_yaml": schema.StringAttribute{
					Description: `YAML-formatted admin (ECE) level "elasticsearch.yml" setting overrides`,
					Computed:    true,
				},
			},
		},
	}
}
