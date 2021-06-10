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

package deploymentresource

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/elastic/cloud-sdk-go/pkg/util/slice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func newElasticsearchResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"autoscale": {
				Type:        schema.TypeString,
				Description: `Enable or disable autoscaling. Defaults to the setting coming from the deployment template. Accepted values are "true" or "false".`,
				Computed:    true,
				Optional:    true,
				ValidateFunc: func(i interface{}, s string) ([]string, []error) {
					if _, err := strconv.ParseBool(i.(string)); err != nil {
						return nil, []error{
							fmt.Errorf("failed parsing autoscale value: %w", err),
						}
					}
					return nil, nil
				},
			},

			"ref_id": {
				Type:        schema.TypeString,
				Description: "Optional ref_id to set on the Elasticsearch resource",
				Default:     "main-elasticsearch",
				Optional:    true,
			},

			// Computed attributes
			"resource_id": {
				Type:        schema.TypeString,
				Description: "The Elasticsearch resource unique identifier",
				Computed:    true,
			},
			"region": {
				Type:        schema.TypeString,
				Description: "The Elasticsearch resource region",
				Computed:    true,
			},
			"cloud_id": {
				Type:        schema.TypeString,
				Description: "The encoded Elasticsearch credentials to use in Beats or Logstash",
				Computed:    true,
			},
			"http_endpoint": {
				Type:        schema.TypeString,
				Description: "The Elasticsearch resource HTTP endpoint",
				Computed:    true,
			},
			"https_endpoint": {
				Type:        schema.TypeString,
				Description: "The Elasticsearch resource HTTPs endpoint",
				Computed:    true,
			},

			// Sub-objects
			"topology": elasticsearchTopologySchema(),

			"config": elasticsearchConfig(),

			"remote_cluster": elasticsearchRemoteCluster(),

			"snapshot_source": newSnapshotSourceSettings(),

			"extension": newExtensionSchema(),

			"trust_account":  newTrustAccountSchema(),
			"trust_external": newTrustExternalSchema(),
		},
	}
}

func elasticsearchTopologySchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		MinItems:    1,
		Optional:    true,
		Computed:    true,
		Description: `Optional topology element which must be set once but can be set multiple times to compose complex topologies`,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Description: `Required topology ID from the deployment template`,
					Required:    true,
				},
				"instance_configuration_id": {
					Type:        schema.TypeString,
					Description: `Computed Instance Configuration ID of the topology element`,
					Computed:    true,
				},
				"size": {
					Type:        schema.TypeString,
					Description: `Optional amount of memory per node in the "<size in GB>g" notation`,
					Computed:    true,
					Optional:    true,
				},
				"size_resource": {
					Type:        schema.TypeString,
					Description: `Optional size type, defaults to "memory".`,
					Default:     "memory",
					Optional:    true,
				},
				"zone_count": {
					Type:        schema.TypeInt,
					Description: `Optional number of zones that the Elasticsearch cluster will span. This is used to set HA`,
					Computed:    true,
					Optional:    true,
				},
				"node_type_data": {
					Type:        schema.TypeString,
					Description: `The node type for the Elasticsearch Topology element (data node)`,
					Computed:    true,
					Optional:    true,
				},
				"node_type_master": {
					Type:        schema.TypeString,
					Description: `The node type for the Elasticsearch Topology element (master node)`,
					Computed:    true,
					Optional:    true,
				},
				"node_type_ingest": {
					Type:        schema.TypeString,
					Description: `The node type for the Elasticsearch Topology element (ingest node)`,
					Computed:    true,
					Optional:    true,
				},
				"node_type_ml": {
					Type:        schema.TypeString,
					Description: `The node type for the Elasticsearch Topology element (machine learning node)`,
					Computed:    true,
					Optional:    true,
				},
				"node_roles": {
					Type:        schema.TypeSet,
					Set:         schema.HashString,
					Description: `The computed list of node roles for the current topology element`,
					Computed:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},

				"autoscaling": {
					Type:        schema.TypeList,
					Description: "Optional Elasticsearch autoscaling settings, such a maximum and minimum size and resources.",
					Optional:    true,
					Computed:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"max_size_resource": {
								Description: "Maximum resource type for the maximum autoscaling setting.",
								Type:        schema.TypeString,
								Optional:    true,
								Computed:    true,
							},

							"max_size": {
								Description: "Maximum size value for the maximum autoscaling setting.",
								Type:        schema.TypeString,
								Optional:    true,
								Computed:    true,
							},

							"min_size_resource": {
								Description: "Minimum resource type for the minimum autoscaling setting.",
								Type:        schema.TypeString,
								Optional:    true,
								Computed:    true,
							},

							"min_size": {
								Description: "Minimum size value for the minimum autoscaling setting.",
								Type:        schema.TypeString,
								Optional:    true,
								Computed:    true,
							},

							"policy_override_json": {
								Type:        schema.TypeString,
								Description: "Computed policy overrides set directly via the API or other clients.",
								Computed:    true,
							},
						},
					},
				},
			},
		},
	}
}

func elasticsearchConfig() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeList,
		Optional:         true,
		MaxItems:         1,
		DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
		Description:      `Optional Elasticsearch settings which will be applied to all topologies unless overridden on the topology element`,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				// Settings

				// Ignored settings are: [ user_bundles and user_plugins ].
				// Adding support for them will allow users to specify
				// "Extensions" as it is possible in the UI today.
				// The implementation would differ between ECE and ESS.

				// plugins maps to the `enabled_built_in_plugins` API setting.
				"plugins": {
					Type:        schema.TypeSet,
					Set:         schema.HashString,
					Description: "List of Elasticsearch supported plugins, which vary from version to version. Check the Stack Pack version to see which plugins are supported for each version. This is currently only available from the UI and [ecctl](https://www.elastic.co/guide/en/ecctl/master/ecctl_stack_list.html)",
					Optional:    true,
					MinItems:    1,
					Elem: &schema.Schema{
						MinItems: 1,
						Type:     schema.TypeString,
					},
				},

				// User settings
				"user_settings_json": {
					Type:        schema.TypeString,
					Description: `JSON-formatted user level "elasticsearch.yml" setting overrides`,
					Optional:    true,
				},
				"user_settings_override_json": {
					Type:        schema.TypeString,
					Description: `JSON-formatted admin (ECE) level "elasticsearch.yml" setting overrides`,
					Optional:    true,
				},
				"user_settings_yaml": {
					Type:        schema.TypeString,
					Description: `YAML-formatted user level "elasticsearch.yml" setting overrides`,
					Optional:    true,
				},
				"user_settings_override_yaml": {
					Type:        schema.TypeString,
					Description: `YAML-formatted admin (ECE) level "elasticsearch.yml" setting overrides`,
					Optional:    true,
				},
			},
		},
	}
}

func elasticsearchRemoteCluster() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MinItems:    1,
		Description: "Optional Elasticsearch remote clusters to configure for the Elasticsearch resource, can be set multiple times",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"deployment_id": {
					Description:  "Remote deployment ID",
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(32, 32),
					Required:     true,
				},
				"alias": {
					Description:  "Alias for this Cross Cluster Search binding",
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
					Optional:     true,
				},
				"ref_id": {
					Description: `Remote elasticsearch "ref_id", it is best left to the default value`,
					Type:        schema.TypeString,
					Default:     "main-elasticsearch",
					Optional:    true,
				},
				"skip_unavailable": {
					Description: "If true, skip the cluster during search when disconnected",
					Type:        schema.TypeBool,
					Default:     false,
					Optional:    true,
				},
			},
		},
	}
}

func newSnapshotSourceSettings() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Optional snapshot source settings. Restore data from a snapshot of another deployment.",
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"source_elasticsearch_cluster_id": {
					Description: "ID of the Elasticsearch cluster that will be used as the source of the snapshot",
					Type:        schema.TypeString,
					Required:    true,
				},
				"snapshot_name": {
					Description: "Name of the snapshot to restore. Use '__latest_success__' to get the most recent successful snapshot.",
					Type:        schema.TypeString,
					Default:     "__latest_success__",
					Optional:    true,
				},
			},
		},
	}
}

func newExtensionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Set:         esExtensionHash,
		Description: "Optional Elasticsearch extensions such as custom bundles or plugins.",
		Optional:    true,
		MinItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Description: "Extension name.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"type": {
					Description: "Extension type, only `bundle` or `plugin` are supported.",
					Type:        schema.TypeString,
					Required:    true,
					ValidateFunc: func(val interface{}, _ string) ([]string, []error) {
						t := val.(string)
						if !slice.HasString([]string{"bundle", "plugin"}, t) {
							return nil, []error{fmt.Errorf(
								"invalid extension type %s: accepted values are bundle or plugin",
								t,
							)}
						}
						return nil, nil
					},
				},
				"version": {
					Description: "Elasticsearch compatibility version. Bundles should specify major or minor versions with wildcards, such as `7.*` or `*` but **plugins must use full version notation down to the patch level**, such as `7.10.1` and wildcards are not allowed.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"url": {
					Description: "Bundle or plugin URL, the extension URL can be obtained from the `ec_deployment_extension.<name>.url` attribute or the API and cannot be a random HTTP address that is hosted elsewhere.",
					Type:        schema.TypeString,
					Required:    true,
				},
			},
		},
	}
}

func esExtensionHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(m["type"].(string))
	buf.WriteString(m["version"].(string))
	buf.WriteString(m["url"].(string))
	buf.WriteString(m["name"].(string))
	return schema.HashString(buf.String())
}

func newTrustAccountSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Description: "Optional Elasticsearch account trust settings.",
		Optional:    true,
		Computed:    true,
		Elem:        accountResource(),
	}
}

func accountResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"account_id": {
				Description: "The ID of the Account.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"trust_all": {
				Description: "If true, all clusters in this account will by default be trusted and the `trust_allowlist` is ignored.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"trust_allowlist": {
				Description: "The list of clusters to trust. Only used when `trust_all` is false.",
				Type:        schema.TypeSet,
				Set:         schema.HashString,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func newTrustExternalSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Description: "Optional Elasticsearch external trust settings.",
		Optional:    true,
		Computed:    true,
		Elem:        externalResource(),
	}
}

func externalResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"relationship_id": {
				Description: "The ID of the external trust relationship.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"trust_all": {
				Description: "If true, all clusters in this account will by default be trusted and the `trust_allowlist` is ignored.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"trust_allowlist": {
				Description: "The list of clusters to trust. Only used when `trust_all` is false.",
				Type:        schema.TypeSet,
				Set:         schema.HashString,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}
