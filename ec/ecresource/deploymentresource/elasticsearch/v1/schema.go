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
	"context"
	"strings"

	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifier"
	"github.com/elastic/terraform-provider-ec/ec/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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

func ElasticsearchSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Required Elasticsearch resource definition",
		Required:    true,
		Validators:  []tfsdk.AttributeValidator{listvalidator.SizeAtMost(1)},
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"autoscale": {
				Type:        types.StringType,
				Description: `Enable or disable autoscaling. Defaults to the setting coming from the deployment template. Accepted values are "true" or "false".`,
				Computed:    true,
				Optional:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"ref_id": {
				Type:        types.StringType,
				Description: "Optional ref_id to set on the Elasticsearch resource",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.String{Value: "main-elasticsearch"}),
					resource.UseStateForUnknown(),
				},
			},
			"resource_id": {
				Type:        types.StringType,
				Description: "The Elasticsearch resource unique identifier",
				Computed:    true,
				// PlanModifiers: tfsdk.AttributePlanModifiers{
				// 	resource.UseStateForUnknown(),
				// },
			},
			"region": {
				Type:        types.StringType,
				Description: "The Elasticsearch resource region",
				Computed:    true,
				// PlanModifiers: tfsdk.AttributePlanModifiers{
				// 	resource.UseStateForUnknown(),
				// },
			},
			"cloud_id": {
				Type:        types.StringType,
				Description: "The encoded Elasticsearch credentials to use in Beats or Logstash",
				Computed:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
					resource.RequiresReplaceIf(func(ctx context.Context, state, config attr.Value, path path.Path) (bool, diag.Diagnostics) {
						return true, nil
					}, "", ""),
				},
			},
			"http_endpoint": {
				Type:        types.StringType,
				Description: "The Elasticsearch resource HTTP endpoint",
				Computed:    true,
				// PlanModifiers: tfsdk.AttributePlanModifiers{
				// 	resource.UseStateForUnknown(),
				// },
			},
			"https_endpoint": {
				Type:        types.StringType,
				Description: "The Elasticsearch resource HTTPs endpoint",
				Computed:    true,
				// PlanModifiers: tfsdk.AttributePlanModifiers{
				// 	resource.UseStateForUnknown(),
				// },
			},
			"topology": ElasticsearchTopologySchema(),

			"trust_account": ElasticsearchTrustAccountSchema(),

			"trust_external": ElasticsearchTrustExternalSchema(),

			"config": ElasticsearchConfigSchema(),

			"remote_cluster": ElasticsearchRemoteClusterSchema(),

			"snapshot_source": ElasticsearchSnapshotSourceSchema(),

			"extension": ElasticsearchExtensionSchema(),

			"strategy": ElasticsearchStrategySchema(),
		}),
	}
}

func ElasticsearchConfigSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: `Optional Elasticsearch settings which will be applied to all topologies unless overridden on the topology element`,
		Optional:    true,
		Validators:  []tfsdk.AttributeValidator{listvalidator.SizeAtMost(1)},
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			// TODO
			// DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
			"docker_image": {
				Type:        types.StringType,
				Description: "Optionally override the docker image the Elasticsearch nodes will use. Note that this field will only work for internal users only.",
				Optional:    true,
			},
			"plugins": {
				Type: types.SetType{
					ElemType: types.StringType,
				},
				Description: "List of Elasticsearch supported plugins, which vary from version to version. Check the Stack Pack version to see which plugins are supported for each version. This is currently only available from the UI and [ecctl](https://www.elastic.co/guide/en/ecctl/master/ecctl_stack_list.html)",
				Optional:    true,
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

func ElasticsearchTopologySchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Computed:    true,
		Optional:    true,
		Description: `Optional topology element which must be set once but can be set multiple times to compose complex topologies`,
		PlanModifiers: tfsdk.AttributePlanModifiers{
			resource.UseStateForUnknown(),
		},
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"id": {
				Type:        types.StringType,
				Description: `Required topology ID from the deployment template`,
				Required:    true,
			},
			"instance_configuration_id": {
				Type:        types.StringType,
				Description: `Computed Instance Configuration ID of the topology element`,
				Computed:    true,
			},
			"size": {
				Type:        types.StringType,
				Description: `Optional amount of memory per node in the "<size in GB>g" notation`,
				Computed:    true,
				Optional:    true,
			},
			"size_resource": {
				Type:        types.StringType,
				Description: `Optional size type, defaults to "memory".`,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.String{Value: "memory"}),
				},
			},
			"zone_count": {
				Type:        types.Int64Type,
				Description: `Optional number of zones that the Elasticsearch cluster will span. This is used to set HA`,
				Computed:    true,
				Optional:    true,
			},
			"node_type_data": {
				Type:        types.StringType,
				Description: `The node type for the Elasticsearch Topology element (data node)`,
				Computed:    true,
				Optional:    true,
			},
			"node_type_master": {
				Type:        types.StringType,
				Description: `The node type for the Elasticsearch Topology element (master node)`,
				Computed:    true,
				Optional:    true,
			},
			"node_type_ingest": {
				Type:        types.StringType,
				Description: `The node type for the Elasticsearch Topology element (ingest node)`,
				Computed:    true,
				Optional:    true,
			},
			"node_type_ml": {
				Type:        types.StringType,
				Description: `The node type for the Elasticsearch Topology element (machine learning node)`,
				Computed:    true,
				Optional:    true,
			},
			"node_roles": {
				Type: types.SetType{
					ElemType: types.StringType,
				},
				Description: `The computed list of node roles for the current topology element`,
				Computed:    true,
			},
			"autoscaling": ElasticsearchTopologyAutoscalingSchema(),
			"config":      ElasticsearchTopologyConfigSchema(),
		}),
	}
}

func ElasticsearchTopologyAutoscalingSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Optional Elasticsearch autoscaling settings, such a maximum and minimum size and resources.",
		Optional:    true,
		Computed:    true,
		Validators:  []tfsdk.AttributeValidator{listvalidator.SizeAtMost(1)},
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"max_size_resource": {
				Description: "Maximum resource type for the maximum autoscaling setting.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"max_size": {
				Description: "Maximum size value for the maximum autoscaling setting.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"min_size_resource": {
				Description: "Minimum resource type for the minimum autoscaling setting.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"min_size": {
				Description: "Minimum size value for the minimum autoscaling setting.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"policy_override_json": {
				Type:        types.StringType,
				Description: "Computed policy overrides set directly via the API or other clients.",
				Computed:    true,
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
				// TODO fix examples/deployment_css/deployment.tf#61
				// Validators:  []tfsdk.AttributeValidator{validators.Length(32, 32)},
				Required: true,
			},
			"alias": {
				Description: "Alias for this Cross Cluster Search binding",
				Type:        types.StringType,
				// TODO fix examples/deployment_css/deployment.tf#62
				// Validators:  []tfsdk.AttributeValidator{validators.NotEmpty()},
				Required: true,
			},
			"ref_id": {
				Description: `Remote elasticsearch "ref_id", it is best left to the default value`,
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.String{Value: "main-elasticsearch"}),
					resource.UseStateForUnknown(),
				},
				Optional: true,
			},
			"skip_unavailable": {
				Description: "If true, skip the cluster during search when disconnected",
				Type:        types.BoolType,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.Bool{Value: false}),
				},
				Optional: true,
			},
		}),
	}
}

func ElasticsearchSnapshotSourceSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Optional snapshot source settings. Restore data from a snapshot of another deployment.",
		Optional:    true,
		Validators:  []tfsdk.AttributeValidator{listvalidator.SizeAtMost(1)},
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
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
					resource.UseStateForUnknown(),
				},
				Optional: true,
				Computed: true,
			},
		}),
	}
}

func ElasticsearchExtensionSchema() tfsdk.Attribute {
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
				Validators:  []tfsdk.AttributeValidator{validators.OneOf([]string{`"bundle"`, `"plugin"`})},
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

func ElasticsearchTrustAccountSchema() tfsdk.Attribute {
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

func ElasticsearchTrustExternalSchema() tfsdk.Attribute {
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

func ElasticsearchStrategySchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Configuration strategy settings.",
		Optional:    true,
		Validators:  []tfsdk.AttributeValidator{listvalidator.SizeAtMost(1)},
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"type": {
				Description: "Configuration strategy type " + strings.Join(strategiesList, ", "),
				Type:        types.StringType,
				Required:    true,
				Validators:  []tfsdk.AttributeValidator{validators.OneOf(strategiesList)},
				// TODO
				// changes on this setting do not change the plan.
				// DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				// 	return true
				// },
			},
		}),
	}
}

func ElasticsearchTopologyConfigSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: `Computed read-only configuration to avoid unsetting plan settings from 'topology.elasticsearch'`,
		Computed:    true,
		PlanModifiers: tfsdk.AttributePlanModifiers{
			resource.UseStateForUnknown(),
			planmodifier.DefaultValue(types.List{
				Null: true,
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"plugins": types.SetType{
							ElemType: types.StringType,
						},
						"user_settings_json":          types.StringType,
						"user_settings_override_json": types.StringType,
						"user_settings_yaml":          types.StringType,
						"user_settings_override_yaml": types.StringType,
					},
				},
			}),
		},
		Validators: []tfsdk.AttributeValidator{listvalidator.SizeAtMost(1)},
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"plugins": {
				Type: types.SetType{
					ElemType: types.StringType,
				},
				Description: "List of Elasticsearch supported plugins, which vary from version to version. Check the Stack Pack version to see which plugins are supported for each version. This is currently only available from the UI and [ecctl](https://www.elastic.co/guide/en/ecctl/master/ecctl_stack_list.html)",
				Computed:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"user_settings_json": {
				Type:        types.StringType,
				Description: `JSON-formatted user level "elasticsearch.yml" setting overrides`,
				Computed:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"user_settings_override_json": {
				Type:        types.StringType,
				Description: `JSON-formatted admin (ECE) level "elasticsearch.yml" setting overrides`,
				Computed:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"user_settings_yaml": {
				Type:        types.StringType,
				Description: `YAML-formatted user level "elasticsearch.yml" setting overrides`,
				Computed:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"user_settings_override_yaml": {
				Type:        types.StringType,
				Description: `YAML-formatted admin (ECE) level "elasticsearch.yml" setting overrides`,
				Computed:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
		}),
	}
}
