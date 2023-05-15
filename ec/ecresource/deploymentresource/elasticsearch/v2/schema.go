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

package v2

import (
	"strings"

	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// These constants are only used to determine whether or not a dedicated
// tier of masters or ingest (coordinating) nodes are set.
const (
	dataTierRolePrefix           = "data_"
	ingestDataTierRole           = "ingest"
	masterDataTierRole           = "master"
	strategyAutodetect           = "autodetect"
	strategyGrowAndShrink        = "grow_and_shrink"
	strategyRollingGrowAndShrink = "rolling_grow_and_shrink"
	strategyRollingAll           = "rolling_all"
)

// List of update strategies availables.
var strategiesList = []string{
	strategyAutodetect, strategyGrowAndShrink, strategyRollingGrowAndShrink, strategyRollingAll,
}

func ElasticsearchSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Elasticsearch cluster definition",
		Required:    true,
		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			"autoscale": {
				Type:        types.BoolType,
				Description: `Enable or disable autoscaling. Defaults to the setting coming from the deployment template.`,
				Computed:    true,
				Optional:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"ref_id": {
				Type:        types.StringType,
				Description: "A human readable reference for the Elasticsearch resource. The default value `main-elasticsearch` is recommended.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.String{Value: "main-elasticsearch"}),
				},
			},
			"resource_id": {
				Type:        types.StringType,
				Description: "The Elasticsearch resource unique identifier",
				Computed:    true,
			},
			"region": {
				Type:        types.StringType,
				Description: "The Elasticsearch resource region",
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
			"cloud_id": {
				Type:        types.StringType,
				Description: "The encoded Elasticsearch credentials to use in Beats or Logstash",
				Computed:    true,
			},
			"http_endpoint": {
				Type:        types.StringType,
				Description: "The Elasticsearch resource HTTP endpoint",
				Computed:    true,
			},
			"https_endpoint": {
				Type:        types.StringType,
				Description: "The Elasticsearch resource HTTPs endpoint",
				Computed:    true,
			},

			"hot":          elasticsearchTopologySchema("'hot' topology element", true, "hot"),
			"coordinating": elasticsearchTopologySchema("'coordinating' topology element", false, "coordinating"),
			"master":       elasticsearchTopologySchema("'master' topology element", false, "master"),
			"warm":         elasticsearchTopologySchema("'warm' topology element", false, "warm"),
			"cold":         elasticsearchTopologySchema("'cold' topology element", false, "cold"),
			"frozen":       elasticsearchTopologySchema("'frozen' topology element", false, "frozen"),
			"ml":           elasticsearchTopologySchema("'ml' topology element", false, "ml"),

			"trust_account": elasticsearchTrustAccountSchema(),

			"trust_external": elasticsearchTrustExternalSchema(),

			"config": elasticsearchConfigSchema(),

			"remote_cluster": ElasticsearchRemoteClusterSchema(),

			"snapshot": elasticsearchSnapshotSchema(),

			"snapshot_source": elasticsearchSnapshotSourceSchema(),

			"extension": elasticsearchExtensionSchema(),

			"strategy": {
				Description: "Configuration strategy type " + strings.Join(strategiesList, ", "),
				Type:        types.StringType,
				Optional:    true,
				Validators:  []tfsdk.AttributeValidator{stringvalidator.OneOf(strategiesList...)},
			},
		}),
	}
}

func elasticsearchConfigSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: `Elasticsearch settings which will be applied to all topologies`,
		Optional:    true,
		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			"docker_image": {
				Type:        types.StringType,
				Description: "Overrides the docker image the Elasticsearch nodes will use. Note that this field will only work for internal users only.",
				Optional:    true,
			},
			"plugins": {
				Type: types.SetType{
					ElemType: types.StringType,
				},
				Description: "List of Elasticsearch supported plugins, which vary from version to version. Check the Stack Pack version to see which plugins are supported for each version. This is currently only available from the UI and [ecctl](https://www.elastic.co/guide/en/ecctl/master/ecctl_stack_list.html)",
				Optional:    true,
				Computed:    true,
			},
			"user_settings_json": {
				Type:        types.StringType,
				Description: `JSON-formatted user level "elasticsearch.yml" setting overrides`,
				Optional:    true,
			},
			"user_settings_override_json": {
				Type:        types.StringType,
				Description: `JSON-formatted admin (ECE) level "elasticsearch.yml" setting overrides`,
				Optional:    true,
			},
			"user_settings_yaml": {
				Type:        types.StringType,
				Description: `YAML-formatted user level "elasticsearch.yml" setting overrides`,
				Optional:    true,
			},
			"user_settings_override_yaml": {
				Type:        types.StringType,
				Description: `YAML-formatted admin (ECE) level "elasticsearch.yml" setting overrides`,
				Optional:    true,
			},
		}),
	}
}

func elasticsearchTopologyAutoscalingSchema(topologyAttributeName string) tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Optional Elasticsearch autoscaling settings, such a maximum and minimum size and resources.",
		Required:    true,
		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			"max_size_resource": {
				Description: "Maximum resource type for the maximum autoscaling setting.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					UseTopologyStateForUnknown(topologyAttributeName),
				},
			},
			"max_size": {
				Description: "Maximum size value for the maximum autoscaling setting.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					UseTopologyStateForUnknown(topologyAttributeName),
				},
			},
			"min_size_resource": {
				Description: "Minimum resource type for the minimum autoscaling setting.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					UseTopologyStateForUnknown(topologyAttributeName),
				},
			},
			"min_size": {
				Description: "Minimum size value for the minimum autoscaling setting.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					UseTopologyStateForUnknown(topologyAttributeName),
				},
			},
			"policy_override_json": {
				Type:        types.StringType,
				Description: "Computed policy overrides set directly via the API or other clients.",
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					UseTopologyStateForUnknown(topologyAttributeName),
				},
			},
		}),
	}
}

func ElasticsearchRemoteClusterSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Optional Elasticsearch remote clusters to configure for the Elasticsearch resource, can be set multiple times",
		Optional:    true,
		Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			"deployment_id": {
				Description: "Remote deployment ID",
				Type:        types.StringType,
				Validators:  []tfsdk.AttributeValidator{stringvalidator.LengthBetween(32, 32)},
				Required:    true,
			},
			"alias": {
				Description: "Alias for this Cross Cluster Search binding",
				Type:        types.StringType,
				Validators:  []tfsdk.AttributeValidator{stringvalidator.NoneOf("")},
				Required:    true,
			},
			"ref_id": {
				Description: `Remote elasticsearch "ref_id", it is best left to the default value`,
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.String{Value: "main-elasticsearch"}),
				},
				Optional: true,
			},
			"skip_unavailable": {
				Description: "If true, skip the cluster during search when disconnected",
				Type:        types.BoolType,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.Bool{Value: false}),
				},
				Computed: true,
				Optional: true,
			},
		}),
	}
}

func elasticsearchSnapshotSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: `(ECE only) Snapshot configuration settings for an Elasticsearch cluster.

For ESS please use the [elasticstack_elasticsearch_snapshot_repository](https://registry.terraform.io/providers/elastic/elasticstack/latest/docs/resources/elasticsearch_snapshot_repository) resource from the [Elastic Stack terraform provider](https://registry.terraform.io/providers/elastic/elasticstack/latest).`,
		Optional: true,
		Computed: true,
		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			"enabled": {
				Description: "Indicates if Snapshotting is enabled.",
				Type:        types.BoolType,
				Required:    true,
			},
			"repository": elasticsearchSnapshotRepositorySchema(),
		}),
		PlanModifiers: []tfsdk.AttributePlanModifier{
			resource.UseStateForUnknown(),
		},
	}
}

func elasticsearchSnapshotRepositorySchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Snapshot repository configuration",
		Optional:    true,
		Computed:    true,
		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			"reference": elasticsearchSnapshotRepositoryReferenceSchema(),
		}),
		PlanModifiers: []tfsdk.AttributePlanModifier{
			resource.UseStateForUnknown(),
		},
	}
}

func elasticsearchSnapshotRepositoryReferenceSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Cluster snapshot reference repository settings, containing the repository name in ECE fashion",
		Optional:    true,
		Computed:    true,
		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			"repository_name": {
				Description: "ECE snapshot repository name, from the '/platform/configuration/snapshots/repositories' endpoint",
				Type:        types.StringType,
				Required:    true,
			}}),
		PlanModifiers: []tfsdk.AttributePlanModifier{
			resource.UseStateForUnknown(),
		},
	}
}

func elasticsearchSnapshotSourceSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: `Restores data from a snapshot of another deployment.

~> **Note on behavior** The <code>snapshot_source</code> block will not be saved in the Terraform state due to its transient nature. This means that whenever the <code>snapshot_source</code> block is set, a snapshot will **always be restored**, unless removed before running <code>terraform apply</code>.`,
		Optional: true,
		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			"source_elasticsearch_cluster_id": {
				Description: "ID of the Elasticsearch cluster that will be used as the source of the snapshot",
				Type:        types.StringType,
				Required:    true,
			},
			"snapshot_name": {
				Description: "Name of the snapshot to restore. Use '__latest_success__' to get the most recent successful snapshot.",
				Type:        types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.String{Value: "__latest_success__"}),
				},
				Optional: true,
				Computed: true,
			},
		}),
	}
}

func elasticsearchExtensionSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Optional Elasticsearch extensions such as custom bundles or plugins.",
		Optional:    true,
		Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			"name": {
				Description: "Extension name.",
				Type:        types.StringType,
				Required:    true,
			},
			"type": {
				Description: "Extension type, only `bundle` or `plugin` are supported.",
				Type:        types.StringType,
				Required:    true,
				Validators:  []tfsdk.AttributeValidator{stringvalidator.OneOf("bundle", "plugin")},
			},
			"version": {
				Description: "Elasticsearch compatibility version. Bundles should specify major or minor versions with wildcards, such as `7.*` or `*` but **plugins must use full version notation down to the patch level**, such as `7.10.1` and wildcards are not allowed.",
				Type:        types.StringType,
				Required:    true,
			},
			"url": {
				Description: "Bundle or plugin URL, the extension URL can be obtained from the `ec_deployment_extension.<name>.url` attribute or the API and cannot be a random HTTP address that is hosted elsewhere.",
				Type:        types.StringType,
				Required:    true,
			},
		}),
	}
}

func elasticsearchTrustAccountSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Optional Elasticsearch account trust settings.",
		Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			"account_id": {
				Description: "The ID of the Account.",
				Type:        types.StringType,
				Required:    true,
			},
			"trust_all": {
				Description: "If true, all clusters in this account will by default be trusted and the `trust_allowlist` is ignored.",
				Type:        types.BoolType,
				Required:    true,
			},
			"trust_allowlist": {
				Description: "The list of clusters to trust. Only used when `trust_all` is false.",
				Type: types.SetType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
		}),
		Computed: true,
		Optional: true,
		PlanModifiers: tfsdk.AttributePlanModifiers{
			resource.UseStateForUnknown(),
		},
	}
}

func elasticsearchTrustExternalSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Optional Elasticsearch external trust settings.",
		Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			"relationship_id": {
				Description: "The ID of the external trust relationship.",
				Type:        types.StringType,
				Required:    true,
			},
			"trust_all": {
				Description: "If true, all clusters in this account will by default be trusted and the `trust_allowlist` is ignored.",
				Type:        types.BoolType,
				Required:    true,
			},
			"trust_allowlist": {
				Description: "The list of clusters to trust. Only used when `trust_all` is false.",
				Type: types.SetType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
		}),
		Computed: true,
		Optional: true,
		PlanModifiers: tfsdk.AttributePlanModifiers{
			resource.UseStateForUnknown(),
		},
	}
}

func elasticsearchTopologySchema(description string, required bool, topologyAttributeName string) tfsdk.Attribute {
	return tfsdk.Attribute{
		Optional: !required,
		// it should be Computed but Computed triggers TF weird behaviour that leads to unempty plan for zero change config
		// Computed:    true,
		Required:    required,
		Description: description,
		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			"instance_configuration_id": {
				Type:        types.StringType,
				Description: `Computed Instance Configuration ID of the topology element`,
				Computed:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					UseTopologyStateForUnknown(topologyAttributeName),
				},
			},
			"size": {
				Type:        types.StringType,
				Description: `Amount of "size_resource" per node in the "<size in GB>g" notation`,
				Computed:    true,
				Optional:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					UseTopologyStateForUnknown(topologyAttributeName),
				},
			},
			"size_resource": {
				Type:        types.StringType,
				Description: `Size type, defaults to "memory".`,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.String{Value: "memory"}),
				},
			},
			"zone_count": {
				Type:        types.Int64Type,
				Description: `Number of zones that the Elasticsearch cluster will span. This is used to set HA`,
				Computed:    true,
				Optional:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					UseTopologyStateForUnknown(topologyAttributeName),
				},
			},
			"node_type_data": {
				Type:        types.StringType,
				Description: `The node type for the Elasticsearch Topology element (data node)`,
				Computed:    true,
				Optional:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					UseNodeTypesDefault(),
				},
			},
			"node_type_master": {
				Type:        types.StringType,
				Description: `The node type for the Elasticsearch Topology element (master node)`,
				Computed:    true,
				Optional:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					UseNodeTypesDefault(),
				},
			},
			"node_type_ingest": {
				Type:        types.StringType,
				Description: `The node type for the Elasticsearch Topology element (ingest node)`,
				Computed:    true,
				Optional:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					UseNodeTypesDefault(),
				},
			},
			"node_type_ml": {
				Type:        types.StringType,
				Description: `The node type for the Elasticsearch Topology element (machine learning node)`,
				Computed:    true,
				Optional:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					UseNodeTypesDefault(),
				},
			},
			"node_roles": {
				Type: types.SetType{
					ElemType: types.StringType,
				},
				Description: `The computed list of node roles for the current topology element`,
				Computed:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					UseNodeRolesDefault(),
				},
			},
			"autoscaling": elasticsearchTopologyAutoscalingSchema(topologyAttributeName),
		}),
	}
}
