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
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

func ElasticsearchSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Description: "Elasticsearch cluster definition",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"autoscale": schema.BoolAttribute{
				Description: `Enable or disable autoscaling. Defaults to the setting coming from the deployment template.`,
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"ref_id": schema.StringAttribute{
				Description: "A human readable reference for the Elasticsearch resource. The default value `main-elasticsearch` is recommended.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					planmodifiers.StringDefaultValue("main-elasticsearch"),
				},
			},
			"resource_id": schema.StringAttribute{
				Description: "The Elasticsearch resource unique identifier",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": schema.StringAttribute{
				Description: "The Elasticsearch resource region",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cloud_id": schema.StringAttribute{
				Description: "The encoded Elasticsearch credentials to use in Beats or Logstash",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknownUnlessNameOrKibanaStateChanges(),
				},
			},
			"http_endpoint": schema.StringAttribute{
				Description: "The Elasticsearch resource HTTP endpoint",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"https_endpoint": schema.StringAttribute{
				Description: "The Elasticsearch resource HTTPs endpoint",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"hot": elasticsearchTopologySchema(topologySchemaOptions{
				tierName:                      "hot",
				required:                      true,
				nodeRolesImpactedBySizeChange: true,
			}),
			"coordinating": elasticsearchTopologySchema(topologySchemaOptions{
				tierName: "coordinating",
			}),
			"master": elasticsearchTopologySchema(topologySchemaOptions{
				tierName: "master",
			}),
			"warm": elasticsearchTopologySchema(topologySchemaOptions{
				tierName: "warm",
			}),
			"cold": elasticsearchTopologySchema(topologySchemaOptions{
				tierName: "cold",
			}),
			"frozen": elasticsearchTopologySchema(topologySchemaOptions{
				tierName: "frozen",
			}),
			"ml": elasticsearchTopologySchema(topologySchemaOptions{
				tierName: "ml",
			}),

			"trust_account": elasticsearchTrustAccountSchema(),

			"trust_external": elasticsearchTrustExternalSchema(),

			"config": elasticsearchConfigSchema(),

			"remote_cluster": ElasticsearchRemoteClusterSchema(),

			"snapshot": elasticsearchSnapshotSchema(),

			"snapshot_source": elasticsearchSnapshotSourceSchema(),

			"extension": elasticsearchExtensionSchema(),

			"strategy": schema.StringAttribute{
				Description: "Configuration strategy type " + strings.Join(strategiesList, ", "),
				Optional:    true,
				Validators:  []validator.String{stringvalidator.OneOf(strategiesList...)},
			},
		},
	}
}

func elasticsearchConfigSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Description: `Elasticsearch settings which will be applied to all topologies`,
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"docker_image": schema.StringAttribute{
				Description: "Overrides the docker image the Elasticsearch nodes will use. Note that this field will only work for internal users only.",
				Optional:    true,
			},
			"plugins": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "List of Elasticsearch supported plugins, which vary from version to version. Check the Stack Pack version to see which plugins are supported for each version. This is currently only available from the UI and [ecctl](https://www.elastic.co/guide/en/ecctl/master/ecctl_stack_list.html)",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
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
	}
}

func elasticsearchTopologyAutoscalingSchema(topologyAttributeName string) schema.Attribute {
	return schema.SingleNestedAttribute{
		Description: "Optional Elasticsearch autoscaling settings, such a maximum and minimum size and resources.",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"max_size_resource": schema.StringAttribute{
				Description: "Maximum resource type for the maximum autoscaling setting.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseTopologyStateForUnknown(topologyAttributeName),
				},
			},
			"max_size": schema.StringAttribute{
				Description: "Maximum size value for the maximum autoscaling setting.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseTopologyStateForUnknown(topologyAttributeName),
				},
			},
			"min_size_resource": schema.StringAttribute{
				Description: "Minimum resource type for the minimum autoscaling setting.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseTopologyStateForUnknown(topologyAttributeName),
				},
			},
			"min_size": schema.StringAttribute{
				Description: "Minimum size value for the minimum autoscaling setting.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseTopologyStateForUnknown(topologyAttributeName),
				},
			},
			"policy_override_json": schema.StringAttribute{
				Description: "Computed policy overrides set directly via the API or other clients.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseTopologyStateForUnknown(topologyAttributeName),
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
					Validators:  []validator.String{stringvalidator.LengthBetween(32, 32)},
					Required:    true,
				},
				"alias": schema.StringAttribute{
					Description: "Alias for this Cross Cluster Search binding",
					Validators:  []validator.String{stringvalidator.NoneOf("")},
					Required:    true,
				},
				"ref_id": schema.StringAttribute{
					Description: `Remote elasticsearch "ref_id", it is best left to the default value`,
					Computed:    true,
					PlanModifiers: []planmodifier.String{
						planmodifiers.StringDefaultValue("main-elasticsearch"),
					},
					Optional: true,
				},
				"skip_unavailable": schema.BoolAttribute{
					Description: "If true, skip the cluster during search when disconnected",
					PlanModifiers: []planmodifier.Bool{
						planmodifiers.BoolDefaultValue(false),
					},
					Computed: true,
					Optional: true,
				},
			},
		},
	}
}

func elasticsearchSnapshotSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Description: `(ECE only) Snapshot configuration settings for an Elasticsearch cluster.

For ESS please use the [elasticstack_elasticsearch_snapshot_repository](https://registry.terraform.io/providers/elastic/elasticstack/latest/docs/resources/elasticsearch_snapshot_repository) resource from the [Elastic Stack terraform provider](https://registry.terraform.io/providers/elastic/elasticstack/latest).`,
		Optional: true,
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Indicates if Snapshotting is enabled.",
				Required:    true,
			},
			"repository": elasticsearchSnapshotRepositorySchema(),
		},
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
	}
}

func elasticsearchSnapshotRepositorySchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Description: "Snapshot repository configuration",
		Optional:    true,
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"reference": elasticsearchSnapshotRepositoryReferenceSchema(),
		},
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
	}
}

func elasticsearchSnapshotRepositoryReferenceSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Description: "Cluster snapshot reference repository settings, containing the repository name in ECE fashion",
		Optional:    true,
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"repository_name": schema.StringAttribute{
				Description: "ECE snapshot repository name, from the '/platform/configuration/snapshots/repositories' endpoint",
				Required:    true,
			}},
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
	}
}

func elasticsearchSnapshotSourceSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Description: `Restores data from a snapshot of another deployment.

~> **Note on behavior** The <code>snapshot_source</code> block will not be saved in the Terraform state due to its transient nature. This means that whenever the <code>snapshot_source</code> block is set, a snapshot will **always be restored**, unless removed before running <code>terraform apply</code>.`,
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"source_elasticsearch_cluster_id": schema.StringAttribute{
				Description: "ID of the Elasticsearch cluster that will be used as the source of the snapshot",
				Required:    true,
			},
			"snapshot_name": schema.StringAttribute{
				Description: "Name of the snapshot to restore. Use '__latest_success__' to get the most recent successful snapshot.",
				PlanModifiers: []planmodifier.String{
					planmodifiers.StringDefaultValue("__latest_success__"),
				},
				Optional: true,
				Computed: true,
			},
		},
	}
}

func elasticsearchExtensionSchema() schema.Attribute {
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
					Validators:  []validator.String{stringvalidator.OneOf("bundle", "plugin")},
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

func elasticsearchTrustAccountSchema() schema.Attribute {
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
		PlanModifiers: []planmodifier.Set{
			setplanmodifier.UseStateForUnknown(),
		},
	}
}

func elasticsearchTrustExternalSchema() schema.Attribute {
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
		PlanModifiers: []planmodifier.Set{
			setplanmodifier.UseStateForUnknown(),
		},
	}
}

type topologySchemaOptions struct {
	required                      bool
	nodeRolesImpactedBySizeChange bool
	tierName                      string
}

func elasticsearchTopologySchema(options topologySchemaOptions) schema.Attribute {
	nodeRolesPlanModifiers := []planmodifier.Set{
		UseNodeRolesDefault(),
	}

	if options.nodeRolesImpactedBySizeChange {
		nodeRolesPlanModifiers = append(nodeRolesPlanModifiers, SetUnknownOnTopologySizeChange())
	}

	return schema.SingleNestedAttribute{
		Optional: !options.required,
		// it should be Computed but Computed triggers TF weird behaviour that leads to unempty plan for zero change config
		// Computed:    true,
		Required:    options.required,
		Description: fmt.Sprintf("'%s' topology element", options.tierName),
		Attributes: map[string]schema.Attribute{
			"instance_configuration_id": schema.StringAttribute{
				Description: `Computed Instance Configuration ID of the topology element`,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseTopologyStateForUnknown(options.tierName),
				},
			},
			"size": schema.StringAttribute{
				Description: `Amount of "size_resource" per node in the "<size in GB>g" notation`,
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					UseTopologyStateForUnknown(options.tierName),
				},
			},
			"size_resource": schema.StringAttribute{
				Description: `Size type, defaults to "memory".`,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					planmodifiers.StringDefaultValue("memory"),
				},
			},
			"zone_count": schema.Int64Attribute{
				Description: `Number of zones that the Elasticsearch cluster will span. This is used to set HA`,
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					UseTopologyStateForUnknown(options.tierName),
				},
			},
			"node_type_data": schema.StringAttribute{
				Description: `The node type for the Elasticsearch Topology element (data node)`,
				Computed:    true,
				Optional:    true,
				Validators: []validator.String{
					VersionSupportsNodeTypes(),
				},
				PlanModifiers: []planmodifier.String{
					UseNodeTypesDefault(),
				},
			},
			"node_type_master": schema.StringAttribute{
				Description: `The node type for the Elasticsearch Topology element (master node)`,
				Computed:    true,
				Optional:    true,
				Validators: []validator.String{
					versionSupportsNodeTypes{},
				},
				PlanModifiers: []planmodifier.String{
					UseNodeTypesDefault(),
				},
			},
			"node_type_ingest": schema.StringAttribute{
				Description: `The node type for the Elasticsearch Topology element (ingest node)`,
				Computed:    true,
				Optional:    true,
				Validators: []validator.String{
					VersionSupportsNodeTypes(),
				},
				PlanModifiers: []planmodifier.String{
					UseNodeTypesDefault(),
				},
			},
			"node_type_ml": schema.StringAttribute{
				Description: `The node type for the Elasticsearch Topology element (machine learning node)`,
				Computed:    true,
				Optional:    true,
				Validators: []validator.String{
					VersionSupportsNodeTypes(),
				},
				PlanModifiers: []planmodifier.String{
					UseNodeTypesDefault(),
				},
			},
			"node_roles": schema.SetAttribute{
				ElementType:   types.StringType,
				Description:   `The computed list of node roles for the current topology element`,
				Computed:      true,
				PlanModifiers: nodeRolesPlanModifiers,
				Validators: []validator.Set{
					VersionSupportsNodeRoles(),
				},
			},
			"autoscaling": elasticsearchTopologyAutoscalingSchema(options.tierName),
		},
	}
}
