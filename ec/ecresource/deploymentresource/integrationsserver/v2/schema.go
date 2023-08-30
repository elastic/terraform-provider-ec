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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func IntegrationsServerSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Description: "Integrations Server cluster definition. Integrations Server replaces `apm` in Stack versions > 8.0",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"elasticsearch_cluster_ref_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					planmodifiers.StringDefaultValue("main-elasticsearch"),
				},
			},
			"ref_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					planmodifiers.StringDefaultValue("main-integrations_server"),
				},
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
			"endpoints": schema.ObjectAttribute{
				Optional:    true,
				Computed:    true,
				Description: "URLs for the accessing the Fleet and APM API's within this Integrations Server resource.",
				AttributeTypes: map[string]attr.Type{
					"apm":   types.StringType,
					"fleet": types.StringType,
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			"instance_configuration_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					planmodifiers.UseStateForUnknownUnlessTemplateChanged(),
				},
			},
			"size": schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					planmodifiers.UseStateForUnknownUnlessTemplateChanged(),
				},
			},
			"size_resource": schema.StringAttribute{
				Description: `Optional size type, defaults to "memory".`,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					planmodifiers.StringDefaultValue("memory"),
				},
			},
			"zone_count": schema.Int64Attribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"config": schema.SingleNestedAttribute{
				Description: `Optionally define the Integrations Server configuration options for the IntegrationsServer Server`,
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"docker_image": schema.StringAttribute{
						Description: "Optionally override the docker image the Integrations Server nodes will use. Note that this field will only work for internal users only.",
						Optional:    true,
					},
					"debug_enabled": schema.BoolAttribute{
						Description: `Optionally enable debug mode for Integrations Server instances - defaults to false`,
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Bool{
							planmodifiers.BoolDefaultValue(false),
						},
					},
					"user_settings_json": schema.StringAttribute{
						Description: `An arbitrary JSON object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_yaml' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (This field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
						Optional:    true,
						PlanModifiers: []planmodifier.String{
							planmodifiers.SetNullWhenEmptyString(),
						},
					},
					"user_settings_override_json": schema.StringAttribute{
						Description: `An arbitrary JSON object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_yaml' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
						Optional:    true,
						PlanModifiers: []planmodifier.String{
							planmodifiers.SetNullWhenEmptyString(),
						},
					},
					"user_settings_yaml": schema.StringAttribute{
						Description: `An arbitrary YAML object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_json' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (These field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
						Optional:    true,
						PlanModifiers: []planmodifier.String{
							planmodifiers.SetNullWhenEmptyString(),
						},
					},
					"user_settings_override_yaml": schema.StringAttribute{
						Description: `An arbitrary YAML object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_json' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
						Optional:    true,
						PlanModifiers: []planmodifier.String{
							planmodifiers.SetNullWhenEmptyString(),
						},
					},
				},
			},
		},
	}
}
