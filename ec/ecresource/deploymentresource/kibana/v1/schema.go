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
	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func KibanaSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Optional Kibana resource definition",
		Optional:    true,
		Validators:  []tfsdk.AttributeValidator{listvalidator.SizeAtMost(1)},
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"elasticsearch_cluster_ref_id": {
				Type: types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.String{Value: "main-elasticsearch"}),
				},
				Computed: true,
				Optional: true,
			},
			"ref_id": {
				Type: types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.String{Value: "main-kibana"}),
				},
				Computed: true,
				Optional: true,
			},
			"resource_id": {
				Type:     types.StringType,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
			"region": {
				Type:     types.StringType,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
			"http_endpoint": {
				Type:     types.StringType,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
			"https_endpoint": {
				Type:     types.StringType,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
			"topology": {
				Description: `Optional topology element`,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"instance_configuration_id": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							resource.UseStateForUnknown(),
						},
					},
					"size": {
						Type:     types.StringType,
						Computed: true,
						Optional: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							resource.UseStateForUnknown(),
						},
					},
					"size_resource": {
						Type:        types.StringType,
						Description: `Optional size type, defaults to "memory".`,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							planmodifier.DefaultValue(types.String{Value: "memory"}),
						},
						Computed: true,
						Optional: true,
					},
					"zone_count": {
						Type:     types.Int64Type,
						Computed: true,
						Optional: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							resource.UseStateForUnknown(),
						},
					},
				}),
			},
			"config": {
				Optional:    true,
				Description: `Optionally define the Kibana configuration options for the Kibana Server`,
				Validators:  []tfsdk.AttributeValidator{listvalidator.SizeAtMost(1)},
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"docker_image": {
						Type:        types.StringType,
						Description: "Optionally override the docker image the Kibana nodes will use. Note that this field will only work for internal users only.",
						Optional:    true,
					},
					"user_settings_json": {
						Type:        types.StringType,
						Description: `An arbitrary JSON object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_yaml' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (This field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
						Optional:    true,
					},
					"user_settings_override_json": {
						Type:        types.StringType,
						Description: `An arbitrary JSON object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_yaml' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
						Optional:    true,
					},
					"user_settings_yaml": {
						Type:        types.StringType,
						Description: `An arbitrary YAML object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_json' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (These field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
						Optional:    true,
					},
					"user_settings_override_yaml": {
						Type:        types.StringType,
						Description: `An arbitrary YAML object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_json' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
						Optional:    true,
					},
				}),
			},
		}),
	}
}
