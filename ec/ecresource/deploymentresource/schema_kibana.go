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

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// NewSchema returns the schema for an "ec_deployment" resource.
func newKibanaResource() *schema.Resource {
	return &schema.Resource{
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
