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

func newApmResource() *schema.Resource {
	return &schema.Resource{
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
			"topology": apmTopologySchema(),

			"config": apmConfig(),

			// TODO: Implement settings field.
			// "settings": interface{}
		},
	}
}

func apmTopologySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MaxItems: 1,
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
	}
}

func apmConfig() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeList,
		Optional:         true,
		MaxItems:         1,
		DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
		Description:      `Optionally define the Apm configuration options for the APM Server`,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"docker_image": {
					Type:        schema.TypeString,
					Description: "Optionally override the docker image the APM nodes will use. Note that this field will only work for internal users only.",
					Optional:    true,
				},
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
	}
}
