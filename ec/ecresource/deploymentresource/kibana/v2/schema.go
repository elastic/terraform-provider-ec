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

package v2

import (
	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func KibanaSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Description: `Kibana cluster definition.

-> **Note on disabling Kibana** While optional it is recommended deployments specify a Kibana block, since not doing so might cause issues when modifying or upgrading the deployment.`,
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"elasticsearch_cluster_ref_id": schema.StringAttribute{
				PlanModifiers: []planmodifier.String{
					planmodifiers.StringDefaultValue("main-elasticsearch"),
				},
				Computed: true,
				Optional: true,
			},
			"ref_id": schema.StringAttribute{
				PlanModifiers: []planmodifier.String{
					planmodifiers.StringDefaultValue("main-kibana"),
				},
				Computed: true,
				Optional: true,
			},
			"resource_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"http_endpoint": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"https_endpoint": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"instance_configuration_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					planmodifier.UseStateForUnknownUnlessTemplateChanged(),
				},
			},
			"size": schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					planmodifier.UseStateForUnknownUnlessTemplateChanged(),
				},
			},
			"size_resource": schema.StringAttribute{
				Description: `Optional size type, defaults to "memory".`,
				PlanModifiers: []planmodifier.String{
					planmodifiers.StringDefaultValue("memory"),
				},
				Computed: true,
				Optional: true,
			},
			"zone_count": schema.Int64Attribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"config": schema.SingleNestedAttribute{
				Optional:    true,
				Description: `Optionally define the Kibana configuration options for the Kibana Server`,
				Attributes: map[string]schema.Attribute{
					"docker_image": schema.StringAttribute{
						Description: "Optionally override the docker image the Kibana nodes will use. Note that this field will only work for internal users only.",
						Optional:    true,
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
		},
	}
}
