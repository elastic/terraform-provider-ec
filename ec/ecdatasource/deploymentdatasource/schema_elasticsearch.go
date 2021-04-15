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

package deploymentdatasource

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newElasticsearchResourceInfo() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"autoscale": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"healthy": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"cloud_id": {
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
			"ref_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"topology": elasticsearchTopologySchema(),
		},
	}
}

func elasticsearchTopologySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"instance_configuration_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"size": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"size_resource": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"zone_count": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"node_type_data": {
					Type:     schema.TypeBool,
					Computed: true,
				},
				"node_type_master": {
					Type:     schema.TypeBool,
					Computed: true,
				},
				"node_type_ingest": {
					Type:     schema.TypeBool,
					Computed: true,
				},
				"node_type_ml": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"node_roles": {
					Type:     schema.TypeSet,
					Set:      schema.HashString,
					Computed: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},

				"autoscaling": {
					Type:        schema.TypeList,
					Description: "Optional Elasticsearch autoscaling settings, such a maximum and minimum size and resources.",
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"max_size_resource": {
								Description: "Maximum resource type for the maximum autoscaling setting.",
								Type:        schema.TypeString,
								Computed:    true,
							},

							"max_size": {
								Description: "Maximum size value for the maximum autoscaling setting.",
								Type:        schema.TypeString,
								Computed:    true,
							},

							"min_size_resource": {
								Description: "Minimum resource type for the minimum autoscaling setting.",
								Type:        schema.TypeString,
								Computed:    true,
							},

							"min_size": {
								Description: "Minimum size value for the minimum autoscaling setting.",
								Type:        schema.TypeString,
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
