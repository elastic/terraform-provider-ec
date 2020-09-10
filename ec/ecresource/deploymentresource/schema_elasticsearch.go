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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// NewSchema returns the schema for an "ec_deployment" resource.
func newElasticsearchResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			// This field is not very useful, it might be removed.
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
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
				Description: "The Elasticsearch resource identifier",
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
				Description: "The cloud_id credentials to use in Beats or Logstash",
				Computed:    true,
			},
			"http_endpoint": {
				Type:        schema.TypeString,
				Description: "The Elasticsearch resource HTTP endpoint to use to connect to the Elasticsearch cluster",
				Computed:    true,
			},
			"https_endpoint": {
				Type:        schema.TypeString,
				Description: "The Elasticsearch resource HTTPs endpoint to use to connect to the Elasticsearch cluster",
				Computed:    true,
			},

			// Sub-objects
			"topology": elasticsearchTopologySchema(),

			"config": elasticsearchConfig(),

			// This setting hasn't been implemented.
			"snapshot_settings": elasticsearchSnapshotSchema(),

			// This doesn't work properly. Deleting a monitoring setting doesn't work.
			"monitoring_settings": elasticsearchMonitoringSchema(),
		},
	}
}

func elasticsearchTopologySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MinItems: 1,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"instance_configuration_id": {
					Type:     schema.TypeString,
					Required: true,
				},
				"memory_per_node": {
					Type:     schema.TypeString,
					Default:  "4g",
					Optional: true,
				},
				"node_count_per_zone": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"zone_count": {
					Type:     schema.TypeInt,
					Default:  1,
					Optional: true,
				},

				// Node types

				"node_type_data": {
					Type:     schema.TypeBool,
					Default:  true,
					Optional: true,
				},
				"node_type_master": {
					Type:     schema.TypeBool,
					Default:  true,
					Optional: true,
				},
				"node_type_ingest": {
					Type:     schema.TypeBool,
					Default:  true,
					Optional: true,
				},
				"node_type_ml": {
					Type:     schema.TypeBool,
					Optional: true,
				},

				"config": elasticsearchConfig(),
			},
		},
	}
}

// TODO: This schema is missing quite a lot of properties compared to the API model.
func elasticsearchSnapshotSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enabled": {
					Type:     schema.TypeBool,
					Required: true,
				},
				"interval": {
					Type:     schema.TypeString,
					Required: true,
				},
				"retention_max_age": {
					Type:     schema.TypeString,
					Required: true,
				},
				"retention_snapshots": {
					Type:     schema.TypeInt,
					Required: true,
				},
			},
		},
	}
}

func elasticsearchMonitoringSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"target_cluster_id": {
					Type:     schema.TypeString,
					Required: true,
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
		Description:      `Optionally define the Elasticsearch configuration options for the Elasticsearch nodes`,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				// Settings

				// Ignored settings for now: [ user_bundles and user_bundles ].

				// plugins maps to the `enabled_built_in_plugins` API setting.
				"plugins": {
					Type:        schema.TypeSet,
					Set:         schema.HashString,
					Description: "A list of plugin names from the Elastic-supported subset that are bundled with the version images. NOTES: (Users should consult the Elastic stack objects to see what plugins are available, this is currently only available from the UI and ecctl)",
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
	}
}
