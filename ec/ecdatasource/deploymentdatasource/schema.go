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

func newSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"alias": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"healthy": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"region": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"deployment_template_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"traffic_filter": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"observability": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     newObservabilitySettings(),
		},
		"tags": {
			Type:     schema.TypeMap,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},

		// Deployment resources
		"elasticsearch": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     newElasticsearchResourceInfo(),
		},
		"kibana": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     newKibanaResourceInfo(),
		},
		"apm": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     newApmResourceInfo(),
		},
		"enterprise_search": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     newEnterpriseSearchResourceInfo(),
		},
	}
}

func newObservabilitySettings() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"deployment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ref_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"logs": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"metrics": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}
