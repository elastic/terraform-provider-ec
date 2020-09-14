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
			// Pending removal.
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
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
		Type:        schema.TypeList,
		MinItems:    1,
		Required:    true,
		Description: `Required topology element which must be set once but can be set multiple times to compose complex topologies`,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"instance_configuration_id": {
					Type:        schema.TypeString,
					Description: `Required Instance Configuration ID from the deployment template`,
					Required:    true,
				},
				"memory_per_node": {
					Type:        schema.TypeString,
					Description: `Optional amount of memory per node in the "<size in GB>g" notation`,
					Default:     "4g",
					Optional:    true,
				},
				"node_count_per_zone": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"zone_count": {
					Type:        schema.TypeInt,
					Description: `Optional number of zones that the Elasticsearch cluster will span. This is used to set HA`,
					Default:     1,
					Optional:    true,
				},

				// Node types

				"node_type_data": {
					Type:        schema.TypeBool,
					Description: `Optional node type (data) for the Elasticsearch Topology element`,
					Default:     true,
					Optional:    true,
				},
				"node_type_master": {
					Type:        schema.TypeBool,
					Description: `Optional node type (master) for the Elasticsearch Topology element`,
					Default:     true,
					Optional:    true,
				},
				"node_type_ingest": {
					Type:        schema.TypeBool,
					Description: `Optional node type (ingest) for the Elasticsearch Topology element`,
					Default:     true,
					Optional:    true,
				},
				"node_type_ml": {
					Type:        schema.TypeBool,
					Description: `Optional node type (machine learning) for the Elasticsearch Topology element`,
					Optional:    true,
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
