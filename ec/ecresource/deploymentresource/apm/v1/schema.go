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

func ApmTopologySchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Optional topology attribute",
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
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.String{Value: "memory"}),
					resource.UseStateForUnknown(),
				},
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
	}
}

func ApmConfigSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: `Optionally define the Apm configuration options for the APM Server`,
		Optional:    true,
		Validators:  []tfsdk.AttributeValidator{listvalidator.SizeAtMost(1)},
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			// TODO
			// DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
			"docker_image": {
				Type:        types.StringType,
				Description: "Optionally override the docker image the APM nodes will use. Note that this field will only work for internal users only.",
				Optional:    true,
			},
			"debug_enabled": {
				Type:        types.BoolType,
				Description: `Optionally enable debug mode for APM servers - defaults to false`,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.Bool{Value: false}),
					resource.UseStateForUnknown(),
				},
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
				Description: `An arbitrary YAML object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_json' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
				Optional:    true,
			},
			"user_settings_override_yaml": {
				Type:        types.StringType,
				Description: `An arbitrary YAML object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_json' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (These field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
				Optional:    true,
			},
		}),
	}
}

func ApmSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Optional APM resource definition",
		Optional:    true,
		Validators:  []tfsdk.AttributeValidator{listvalidator.SizeAtMost(1)},
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"elasticsearch_cluster_ref_id": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.String{Value: "main-elasticsearch"}),
					// resource.UseStateForUnknown(),
					// planmodifier.UseStateForNoChange(),
				},
			},
			"ref_id": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.String{Value: "main-apm"}),
					// resource.UseStateForUnknown(),
					// planmodifier.UseStateForNoChange(),
				},
			},
			"resource_id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					// resource.UseStateForUnknown(),
					// planmodifier.UseStateForNoChange(),
				},
			},
			"region": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					// resource.UseStateForUnknown(),
					// planmodifier.UseStateForNoChange(),
				},
			},
			"http_endpoint": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					// resource.UseStateForUnknown(),
					// planmodifier.UseStateForNoChange(),
				},
			},
			"https_endpoint": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					// resource.UseStateForUnknown(),
					// planmodifier.UseStateForNoChange(),
				},
			},
			"topology": ApmTopologySchema(),
			"config":   ApmConfigSchema(),
		}),
	}
}
