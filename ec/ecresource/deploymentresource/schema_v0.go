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
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceStateUpgradeV0(_ context.Context, raw map[string]interface{}, m interface{}) (map[string]interface{}, error) {
	for _, apm := range raw["apm"].([]interface{}) {
		rawApm := apm.(map[string]interface{})
		delete(rawApm, "version")
	}

	for _, es := range raw["elasticsearch"].([]interface{}) {
		rawEs := es.(map[string]interface{})
		delete(rawEs, "version")
	}

	for _, ess := range raw["enterprise_search"].([]interface{}) {
		rawEss := ess.(map[string]interface{})
		delete(rawEss, "version")
	}

	for _, kibana := range raw["kibana"].([]interface{}) {
		rawKibana := kibana.(map[string]interface{})
		delete(rawKibana, "version")
	}

	return raw, nil
}

// Copy of the revision 0 of the deployment schema.
func resourceSchemaV0() *schema.Resource {
	return &schema.Resource{Schema: map[string]*schema.Schema{
		"version": {
			Type:        schema.TypeString,
			Description: "Required Elastic Stack version to use for all of the deployment resources",
			Required:    true,
		},
		"region": {
			Type:        schema.TypeString,
			Description: `Required ESS region where to create the deployment, for ECE environments "ece-region" must be set`,
			Required:    true,
			ForceNew:    true,
		},
		"deployment_template_id": {
			Type:        schema.TypeString,
			Description: "Required Deployment Template identifier to create the deployment from",
			Required:    true,
		},
		"name": {
			Type:        schema.TypeString,
			Description: "Optional name for the deployment",
			Optional:    true,
		},
		"request_id": {
			Type:        schema.TypeString,
			Description: "Optional request_id to set on the create operation, only use when previous create attempts return with an error and a request_id is returned as part of the error",
			Optional:    true,
		},

		// Computed ES Creds
		"elasticsearch_username": {
			Type:        schema.TypeString,
			Description: "Computed username obtained upon creating the Elasticsearch resource",
			Computed:    true,
		},
		"elasticsearch_password": {
			Type:        schema.TypeString,
			Description: "Computed password obtained upon creating the Elasticsearch resource",
			Computed:    true,
			Sensitive:   true,
		},

		// APM secret_token
		"apm_secret_token": {
			Type:      schema.TypeString,
			Computed:  true,
			Sensitive: true,
		},

		// Resources
		"elasticsearch": {
			Type:        schema.TypeList,
			Description: "Required Elasticsearch resource definition",
			MaxItems:    1,
			Required:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
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
					"version": {
						Type:        schema.TypeString,
						Description: "The Elasticsearch resource current version",
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
					"topology": {
						Type:        schema.TypeList,
						MinItems:    1,
						Optional:    true,
						Computed:    true,
						Description: `Optional topology element which must be set once but can be set multiple times to compose complex topologies`,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"instance_configuration_id": {
									Type:        schema.TypeString,
									Description: `Computed Instance Configuration ID of the topology element`,
									Computed:    true,
									Optional:    true,
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
							},
						},
					},

					"config": {
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
					},

					"remote_cluster": {
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
					},

					"snapshot_source": {
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
					},
				},
			},
		},
		"kibana": {
			Type:        schema.TypeList,
			Description: "Optional Kibana resource definition",
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"elasticsearch_cluster_ref_id": {
						Type:     schema.TypeString,
						Default:  "main-elasticsearch",
						Optional: true,
					},
					"ref_id": {
						Type:     schema.TypeString,
						Default:  "main-kibana",
						Optional: true,
					},
					"resource_id": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"version": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"region": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"http_endpoint": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"https_endpoint": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"topology": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"instance_configuration_id": {
									Type:     schema.TypeString,
									Optional: true,
									Computed: true,
								},
								"size": {
									Type:     schema.TypeString,
									Computed: true,
									Optional: true,
								},
								"size_resource": {
									Type:        schema.TypeString,
									Description: `Optional size type, defaults to "memory".`,
									Default:     "memory",
									Optional:    true,
								},
								"zone_count": {
									Type:     schema.TypeInt,
									Computed: true,
									Optional: true,
								},
							},
						},
					},

					"config": {
						Type:             schema.TypeList,
						Optional:         true,
						MaxItems:         1,
						DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
						Description:      `Optionally define the Kibana configuration options for the Kibana Server`,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"user_settings_json": {
									Type:        schema.TypeString,
									Description: `An arbitrary JSON object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_yaml' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (This field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
									Optional:    true,
								},
								"user_settings_override_json": {
									Type:        schema.TypeString,
									Description: `An arbitrary JSON object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_yaml' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
									Optional:    true,
								},
								"user_settings_yaml": {
									Type:        schema.TypeString,
									Description: `An arbitrary YAML object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_json' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
									Optional:    true,
								},
								"user_settings_override_yaml": {
									Type:        schema.TypeString,
									Description: `An arbitrary YAML object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_json' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (These field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
									Optional:    true,
								},
							},
						},
					},
				},
			},
		},
		"apm": {
			Type:        schema.TypeList,
			Description: "Optional APM resource definition",
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"elasticsearch_cluster_ref_id": {
						Type:     schema.TypeString,
						Default:  "main-elasticsearch",
						Optional: true,
					},
					"ref_id": {
						Type:     schema.TypeString,
						Default:  "main-apm",
						Optional: true,
					},
					"resource_id": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"version": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"region": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"http_endpoint": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"https_endpoint": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"topology": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"instance_configuration_id": {
									Type:     schema.TypeString,
									Optional: true,
									Computed: true,
								},
								"size": {
									Type:     schema.TypeString,
									Computed: true,
									Optional: true,
								},
								"size_resource": {
									Type:        schema.TypeString,
									Description: `Optional size type, defaults to "memory".`,
									Default:     "memory",
									Optional:    true,
								},
								"zone_count": {
									Type:     schema.TypeInt,
									Computed: true,
									Optional: true,
								},
							},
						},
					},

					"config": {
						Type:             schema.TypeList,
						Optional:         true,
						MaxItems:         1,
						DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
						Description:      `Optionally define the Apm configuration options for the APM Server`,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								// APM System Settings
								"debug_enabled": {
									Type:        schema.TypeBool,
									Description: `Optionally enable debug mode for APM servers - defaults to false`,
									Optional:    true,
									Default:     false,
								},

								"user_settings_json": {
									Type:        schema.TypeString,
									Description: `An arbitrary JSON object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_yaml' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (This field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
									Optional:    true,
								},
								"user_settings_override_json": {
									Type:        schema.TypeString,
									Description: `An arbitrary JSON object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_yaml' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
									Optional:    true,
								},
								"user_settings_yaml": {
									Type:        schema.TypeString,
									Description: `An arbitrary YAML object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_json' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
									Optional:    true,
								},
								"user_settings_override_yaml": {
									Type:        schema.TypeString,
									Description: `An arbitrary YAML object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_json' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (These field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
									Optional:    true,
								},
							},
						},
					},
				},
			},
		},
		"enterprise_search": {
			Type:        schema.TypeList,
			Description: "Optional Enterprise Search resource definition",
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"elasticsearch_cluster_ref_id": {
						Type:     schema.TypeString,
						Default:  "main-elasticsearch",
						Optional: true,
					},
					"ref_id": {
						Type:     schema.TypeString,
						Default:  "main-enterprise_search",
						Optional: true,
					},
					"resource_id": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"version": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"region": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"http_endpoint": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"https_endpoint": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"topology": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"instance_configuration_id": {
									Type:     schema.TypeString,
									Optional: true,
									Computed: true,
								},
								"size": {
									Type:     schema.TypeString,
									Computed: true,
									Optional: true,
								},
								"size_resource": {
									Type:        schema.TypeString,
									Description: `Optional size type, defaults to "memory".`,
									Default:     "memory",
									Optional:    true,
								},
								"zone_count": {
									Type:     schema.TypeInt,
									Computed: true,
									Optional: true,
								},

								// Node types

								"node_type_appserver": {
									Type:     schema.TypeBool,
									Computed: true,
								},
								"node_type_connector": {
									Type:     schema.TypeBool,
									Computed: true,
								},
								"node_type_worker": {
									Type:     schema.TypeBool,
									Computed: true,
								},
							},
						},
					},

					"config": {
						Type:             schema.TypeList,
						Optional:         true,
						MaxItems:         1,
						DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
						Description:      `Optionally define the Enterprise Search configuration options for the Enterprise Search Server`,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"user_settings_json": {
									Type:        schema.TypeString,
									Description: `An arbitrary JSON object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_yaml' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (This field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
									Optional:    true,
								},
								"user_settings_override_json": {
									Type:        schema.TypeString,
									Description: `An arbitrary JSON object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_yaml' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
									Optional:    true,
								},
								"user_settings_yaml": {
									Type:        schema.TypeString,
									Description: `An arbitrary YAML object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_json' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
									Optional:    true,
								},
								"user_settings_override_yaml": {
									Type:        schema.TypeString,
									Description: `An arbitrary YAML object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_json' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (These field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
									Optional:    true,
								},
							},
						},
					},
				},
			},
		},

		// Settings
		"traffic_filter": {
			Description: "Optional list of traffic filters to apply to this deployment.",
			// This field is a TypeSet since the order of the items isn't
			// important, but the unique list is. This prevents infinite loops
			// for autogenerated IDs.
			Type:     schema.TypeSet,
			Set:      schema.HashString,
			Optional: true,
			MinItems: 1,
			Elem: &schema.Schema{
				MinItems: 1,
				Type:     schema.TypeString,
			},
		},
		"observability": {
			Type:        schema.TypeList,
			Description: "Optional observability settings. Ship logs and metrics to a dedicated deployment.",
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"deployment_id": {
						Type:     schema.TypeString,
						Required: true,
					},
					"ref_id": {
						Type:     schema.TypeString,
						Computed: true,
						Optional: true,
					},
					"logs": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  true,
					},
					"metrics": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  true,
					},
				},
			},
		},
		"tags": {
			Description: "Optional map of deployment tags",
			Type:        schema.TypeMap,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}}
}
