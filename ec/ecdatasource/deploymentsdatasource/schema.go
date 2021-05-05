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

package deploymentsdatasource

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func newSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name_prefix": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"healthy": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"deployment_template_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"tags": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"size": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  100,
		},

		// Computed
		"return_count": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"deployments": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     newDeploymentList(),
		},

		// Deployment resources
		"elasticsearch": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem:     newResourceFilters(),
		},
		"kibana": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem:     newResourceFilters(),
		},
		"apm": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem:     newResourceFilters(),
		},
		"enterprise_search": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem:     newResourceFilters(),
		},
	}
}

func newDeploymentList() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"deployment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"elasticsearch_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kibana_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"apm_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enterprise_search_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func newResourceFilters() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"healthy": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}
