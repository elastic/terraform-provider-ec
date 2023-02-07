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

package v1

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func ApmTopologySchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Optional: true,
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"instance_configuration_id": schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
				"size": schema.StringAttribute{
					Computed: true,
					Optional: true,
				},
				"size_resource": schema.StringAttribute{
					Description: `Optional size type, defaults to "memory".`,
					Optional:    true,
					Computed:    true,
				},
				"zone_count": schema.Int64Attribute{
					Computed: true,
					Optional: true,
				},
			},
		},
	}
}

func ApmConfigSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Description: `Optionally define the Apm configuration options for the APM Server`,
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"docker_image": schema.StringAttribute{
					Description: "Optionally override the docker image the APM nodes will use. This option will not work in ESS customers and should only be changed if you know what you're doing.",
					Optional:    true,
				},
				"debug_enabled": schema.BoolAttribute{
					Description: `Optionally enable debug mode for APM servers - defaults to false`,
					Optional:    true,
					Computed:    true,
				},
				"user_settings_json": schema.StringAttribute{
					Description: `An arbitrary JSON object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_yaml' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (This field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
					Optional:    true,
				},
				"user_settings_override_json": schema.StringAttribute{
					Description: `An arbitrary JSON object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_yaml' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
					Optional:    true,
				},
				"user_settings_yaml": schema.StringAttribute{
					Description: `An arbitrary YAML object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_json' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (These field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
					Optional:    true,
				},
				"user_settings_override_yaml": schema.StringAttribute{
					Description: `An arbitrary YAML object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_json' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
					Optional:    true,
				},
			},
		},
	}
}

func ApmSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Description: "Optional APM resource definition",
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"elasticsearch_cluster_ref_id": schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
				"ref_id": schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
				"resource_id": schema.StringAttribute{
					Computed: true,
				},
				"region": schema.StringAttribute{
					Computed: true,
				},
				"http_endpoint": schema.StringAttribute{
					Computed: true,
				},
				"https_endpoint": schema.StringAttribute{
					Computed: true,
				},
				"topology": ApmTopologySchema(),
				"config":   ApmConfigSchema(),
			},
		},
	}
}
