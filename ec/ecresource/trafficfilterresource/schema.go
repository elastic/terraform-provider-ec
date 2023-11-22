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

package trafficfilterresource

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api"

	"github.com/elastic/terraform-provider-ec/ec/internal"
	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifiers"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Provides an Elastic Cloud traffic filter resource, which allows traffic filter rules to be created, updated, and deleted. Traffic filter rules are used to limit inbound traffic to deployment resources.

  ~> **Note on traffic filters** If you use traffic_filter on an ec_deployment, Terraform will manage the full set of traffic rules for the deployment, and treat additional traffic filters as drift. For this reason, traffic_filter cannot be mixed with the ec_deployment_traffic_filter_association resource for a given deployment.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the ruleset",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of the ruleset. It can be `ip`, `vpce`, `azure_private_endpoint`, or `gcp_private_service_connect_endpoint`",
				Required:    true,
			},
			"region": schema.StringAttribute{
				Description: "Filter region, the ruleset can only be attached to deployments in the specific region",
				Required:    true,
			},
			"include_by_default": schema.BoolAttribute{
				Description: "Indicates that the ruleset should be automatically included in new deployments (Defaults to false)",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					planmodifiers.BoolDefaultValue(false),
				},
			},
			"description": schema.StringAttribute{
				Description: "Ruleset description",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"rule": trafficFilterRuleSchema(),
		},
	}
}

func trafficFilterRuleSchema() schema.Block {
	return schema.SetNestedBlock{
		Description: "Set of rules, which the ruleset is made of.",
		Validators:  []validator.Set{setvalidator.SizeAtLeast(1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"source": schema.StringAttribute{
					Description: "Traffic filter source: IP address, CIDR mask, or VPC endpoint ID, **only required** when the type is not `azure_private_endpoint`",
					Optional:    true,
				},
				"description": schema.StringAttribute{
					Description: "Description of this individual rule",
					Optional:    true,
				},
				"azure_endpoint_name": schema.StringAttribute{
					Description: "Azure endpoint name. Only applicable when the ruleset type is set to `azure_private_endpoint`",
					Optional:    true,
				},
				"azure_endpoint_guid": schema.StringAttribute{
					Description: "Azure endpoint GUID. Only applicable when the ruleset type is set to `azure_private_endpoint`",
					Optional:    true,
				},
				"id": schema.StringAttribute{
					Description: "Computed rule ID",
					Computed:    true,
					PlanModifiers: []planmodifier.String{
						StringIsUnknownIfRulesChange(),
					},
					// NOTE: The ID will change on update, so we intentionally do not use plan modifier resource.UseStateForUnknown() here!
				},
			},
		},
	}
}

func trafficFilterRuleSetType() attr.Type {
	return trafficFilterRuleSchema().Type()
}

func trafficFilterRuleElemType() attr.Type {
	return trafficFilterRuleSchema().Type().(types.SetType).ElemType
}

func trafficFilterRuleAttrTypes() map[string]attr.Type {
	return trafficFilterRuleSchema().Type().(types.SetType).ElemType.(types.ObjectType).AttrTypes
}

/* TODO
Timeouts: &schema.ResourceTimeout{
	Default: schema.DefaultTimeout(10 * time.Minute),
},
*/

type Resource struct {
	client *api.API
}

func resourceReady(r Resource, dg *diag.Diagnostics) bool {
	if r.client == nil {
		dg.AddError(
			"Unconfigured API Client",
			"Expected configured API client. Please report this issue to the provider developers.",
		)

		return false
	}
	return true
}

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), request.ID)...)
}

func (r *Resource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client, diags := internal.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	r.client = client
}

func (r *Resource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_deployment_traffic_filter"
}

type modelV0 struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Type             types.String `tfsdk:"type"`
	Region           types.String `tfsdk:"region"`
	Rule             types.Set    `tfsdk:"rule"` //< trafficFilterRuleModelV0TF
	IncludeByDefault types.Bool   `tfsdk:"include_by_default"`
	Description      types.String `tfsdk:"description"`
}

type trafficFilterRuleModelV0 struct {
	ID                types.String `tfsdk:"id"`
	Source            types.String `tfsdk:"source"`
	Description       types.String `tfsdk:"description"`
	AzureEndpointName types.String `tfsdk:"azure_endpoint_name"`
	AzureEndpointGUID types.String `tfsdk:"azure_endpoint_guid"`
}

type stringIsUnknownIfRulesChange struct{}

func StringIsUnknownIfRulesChange() planmodifier.String {
	return &stringIsUnknownIfRulesChange{}
}

func (m *stringIsUnknownIfRulesChange) Description(ctx context.Context) string {
	return m.MarkdownDescription(ctx)
}

func (m *stringIsUnknownIfRulesChange) MarkdownDescription(ctx context.Context) string {
	return "Sets the plan to unknown if there are rule changes"
}

func (m *stringIsUnknownIfRulesChange) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.ConfigValue.IsNull() {
		return
	}

	var stateRules []trafficFilterRuleModelV0
	resp.Diagnostics = append(resp.Diagnostics, req.State.GetAttribute(ctx, path.Root("rule"), &stateRules)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var planRules []trafficFilterRuleModelV0
	resp.Diagnostics = append(resp.Diagnostics, req.Plan.GetAttribute(ctx, path.Root("rule"), &planRules)...)

	hasChanges := false
	for _, stateRule := range stateRules {
		if !slices.Contains(planRules, stateRule) {
			hasChanges = true
		}
	}

	for _, planRule := range planRules {
		if !slices.Contains(stateRules, planRule) {
			hasChanges = true
		}
	}

	if hasChanges {
		resp.PlanValue = types.StringUnknown()
	}
}
