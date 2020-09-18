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

package stackdatasource

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"version_regex": {
			Type:     schema.TypeString,
			Required: true,
		},
		"region": {
			Type:     schema.TypeString,
			Required: true,
		},
		"lock": {
			Type:     schema.TypeBool,
			Optional: true,
		},

		// Exported attributes
		"version": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"accessible": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"min_upgradable_from": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"upgradable_to": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"allowlisted": {
			Type:     schema.TypeBool,
			Computed: true,
		},

		"apm":               newKindResourceSchema(),
		"enterprise_search": newKindResourceSchema(),
		"elasticsearch":     newKindResourceSchema(),
		"kibana":            newKindResourceSchema(),
	}
}

func newKindResourceSchema() *schema.Schema {
	return &schema.Schema{
		Computed: true,
		Type:     schema.TypeList,
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"denylist": {
				Computed: true,
				Type:     schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"capacity_constraints_max": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"capacity_constraints_min": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"compatible_node_types": {
				Computed: true,
				Type:     schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"docker_image": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"plugins": {
				Computed: true,
				Type:     schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"default_plugins": {
				Computed: true,
				Type:     schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			// node_types not added. It is highly unlikely they will be used
			// for anything, and if they're needed in the future, then we can
			// invest on adding them.
		}},
	}
}
