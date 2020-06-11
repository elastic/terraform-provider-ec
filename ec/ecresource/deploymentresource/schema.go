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

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

// NewSchema returns the schema for an "ec_deployment" resource.
func NewSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"version": {
			Type:     schema.TypeString,
			Required: true,
		},
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"region": {
			Type:     schema.TypeString,
			Required: true,
		},
		"request_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"deployment_template_id": {
			Type:     schema.TypeString,
			Required: true,
		},

		// Workloads

		"elasticsearch": {
			Type:     schema.TypeList,
			MinItems: 1,
			MaxItems: 1,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"display_name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"ref_id": {
						Type:     schema.TypeString,
						Default:  "main-elasticsearch",
						Optional: true,
					},

					// Computed attributes
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
					"cloud_id": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"username": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"password": {
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

					// Sub-objects
					"topology": elasticsearchTopologySchema(),

					// This setting hasn't been implemented.
					"snapshot_settings": elasticsearchSnapshotSchema(),

					// This doesn't work properly. Deleting a monitoring setting doesn't work.
					"monitoring_settings": elasticsearchMonitoringSchema(),

					// TODO: Implement settings field.
					// "settings": interface{}
				},
			},
		},
		"kibana": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"elasticsearch_cluster_ref_id": {
						Type:     schema.TypeString,
						Default:  "main-elasticsearch",
						Optional: true,
					},
					"display_name": {
						Type:     schema.TypeString,
						Computed: true,
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
					"topology": kibanaTopologySchema(),

					// TODO: Implement settings field.
					// "settings": interface{}
				},
			},
		},
		"apm": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"elasticsearch_cluster_ref_id": {
						Type:     schema.TypeString,
						Default:  "main-elasticsearch",
						Optional: true,
					},
					"display_name": {
						Type:     schema.TypeString,
						Computed: true,
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
					"secret_token": {
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
					"topology": apmTopologySchema(),

					// TODO: Implement settings field.
					// "settings": interface{}
				},
			},
		},
		"appsearch": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"elasticsearch_cluster_ref_id": {
						Type:     schema.TypeString,
						Default:  "main-elasticsearch",
						Optional: true,
					},
					"display_name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"ref_id": {
						Type:     schema.TypeString,
						Default:  "main-appsearch",
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
					"topology": appsearchTopologySchema(),

					// TODO: Implement settings field.
					// "settings": interface{}
				},
			},
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

func kibanaTopologySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"instance_configuration_id": {
					Type:     schema.TypeString,
					Required: true,
				},
				"memory_per_node": {
					Type:     schema.TypeString,
					Default:  "1g",
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
			},
		},
	}
}

func apmTopologySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"instance_configuration_id": {
					Type:     schema.TypeString,
					Required: true,
				},
				"memory_per_node": {
					Type:     schema.TypeString,
					Default:  "0.5g",
					Optional: true,
				},
				"zone_count": {
					Type:     schema.TypeInt,
					Default:  1,
					Optional: true,
				},
			},
		},
	}
}

func appsearchTopologySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"instance_configuration_id": {
					Type:     schema.TypeString,
					Required: true,
				},
				"memory_per_node": {
					Type:     schema.TypeString,
					Default:  "2g",
					Optional: true,
				},
				"zone_count": {
					Type:     schema.TypeInt,
					Default:  1,
					Optional: true,
				},

				// Node types

				"node_type_appserver": {
					Type:     schema.TypeBool,
					Default:  true,
					Optional: true,
				},
				"node_type_worker": {
					Type:     schema.TypeBool,
					Default:  true,
					Optional: true,
				},
			},
		},
	}
}
