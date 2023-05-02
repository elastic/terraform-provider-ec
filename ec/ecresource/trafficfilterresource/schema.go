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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api"

	"github.com/elastic/terraform-provider-ec/ec/internal"
	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifier"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

func (r *Resource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `Provides an Elastic Cloud traffic filter resource, which allows traffic filter rules to be created, updated, and deleted. Traffic filter rules are used to limit inbound traffic to deployment resources.`,
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:                types.StringType,
				MarkdownDescription: "Unique identifier of this resource.",
				Computed:            true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"name": {
				Type:        types.StringType,
				Description: "Name of the ruleset",
				Required:    true,
			},
			"type": {
				Type:        types.StringType,
				Description: "Type of the ruleset. It can be `ip`, `vpce`, `azure_private_endpoint`, or `gcp_private_service_connect_endpoint`",
				Required:    true,
			},
			"region": {
				Type:        types.StringType,
				Description: "Filter region, the ruleset can only be attached to deployments in the specific region",
				Required:    true,
			},
			"include_by_default": {
				Type:        types.BoolType,
				Description: "Indicates that the ruleset should be automatically included in new deployments (Defaults to false)",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.Bool{Value: false}),
				},
			},
			"description": {
				Type:        types.StringType,
				Description: "Ruleset description",
				Optional:    true,
			},
		},
		Blocks: map[string]tfsdk.Block{
			"rule": trafficFilterRuleSchema(),
		},
	}, nil
}

func trafficFilterRuleSchema() tfsdk.Block {
	return tfsdk.Block{
		Description: "Set of rules, which the ruleset is made of.",
		NestingMode: tfsdk.BlockNestingModeSet,
		MinItems:    1,
		Attributes: map[string]tfsdk.Attribute{
			"source": {
				Type:        types.StringType,
				Description: "Traffic filter source: IP address, CIDR mask, or VPC endpoint ID, **only required** when the type is not `azure_private_endpoint`",
				Optional:    true,
			},
			"description": {
				Type:        types.StringType,
				Description: "Description of this individual rule",
				Optional:    true,
			},
			"azure_endpoint_name": {
				Type:        types.StringType,
				Description: "Azure endpoint name. Only applicable when the ruleset type is set to `azure_private_endpoint`",
				Optional:    true,
			},
			"azure_endpoint_guid": {
				Type:        types.StringType,
				Description: "Azure endpoint GUID. Only applicable when the ruleset type is set to `azure_private_endpoint`",
				Optional:    true,
			},
			"id": {
				Type:        types.StringType,
				Description: "Computed rule ID",
				Computed:    true,
				// NOTE: The ID will change on update, so we intentionally do not use plan modifier resource.UseStateForUnknown() here!
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
	Rule             types.Set    `tfsdk:"rule"` //< trafficFilterRuleModelV0
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
