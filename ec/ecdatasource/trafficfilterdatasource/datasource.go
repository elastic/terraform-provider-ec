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

package trafficfilterdatasource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/trafficfilterapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/terraform-provider-ec/ec/internal"
)

type DataSource struct {
	client *api.API
}

var _ datasource.DataSource = &DataSource{}
var _ datasource.DataSourceWithConfigure = &DataSource{}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to filter for an existing traffic filter that has been created via one of the provided filters.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The exact name of the traffic filter to select.",
				Optional:    true,
			},
			"id": schema.StringAttribute{
				Description: "The id of the traffic filter to select.",
				Optional:    true,
			},
			"region": schema.StringAttribute{
				Description: "Region where the traffic filter is. For Elastic Cloud Enterprise (ECE) installations, use `ece-region`",
				Optional:    true,
			},

			// computed fields
			"rulesets": rulesetSchema(),
		},
	}
}

func rulesetSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Description: "An individual ruleset",
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Description: "The ID of the ruleset",
					Computed:    true,
				},
				"name": schema.StringAttribute{
					Description: "The name of the ruleset.",
					Computed:    true,
				},
				"description": schema.StringAttribute{
					Description: "The description of the ruleset.",
					Computed:    true,
				},
				"region": schema.StringAttribute{
					Description: "The ruleset can be attached only to deployments in the specific region.",
					Computed:    true,
				},
				"include_by_default": schema.BoolAttribute{
					Description: "Should the ruleset be automatically included in the new deployments.",
					Computed:    true,
				},
				"rules": ruleSchema(),
			},
		},
	}
}

func ruleSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Description: "An individual rule",
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Description: "The ID of the rule",
					Computed:    true,
				},
				"source": schema.StringAttribute{
					Description: "Allowed traffic filter source: IP address, CIDR mask, or VPC endpoint ID.",
					Computed:    true,
				},
				"description": schema.StringAttribute{
					Description: "The description of the rule.",
					Computed:    true,
				},
			},
		},
	}
}

func (d DataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	// Prevent panic if the provider has not been configured.
	if d.client == nil {
		response.Diagnostics.AddError(
			"Unconfigured API Client",
			"Expected configured API client. Please report this issue to the provider developers.",
		)

		return
	}

	var newState modelV0
	response.Diagnostics.Append(request.Config.Get(ctx, &newState)...)
	if response.Diagnostics.HasError() {
		return
	}

	res, err := trafficfilterapi.List(trafficfilterapi.ListParams{
		API: d.client,
	})

	if err != nil {
		response.Diagnostics.AddError(
			"Failed retrieving deployment information",
			fmt.Sprintf("Failed retrieving deployment information: %s", err),
		)
		return
	}

	response.Diagnostics.Append(modelToState(ctx, res, &newState)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Finally, set the state
	response.Diagnostics.Append(response.State.Set(ctx, newState)...)
}

func (d *DataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_traffic_filter"
}

func (d *DataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	clients, diags := internal.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	d.client = clients.Stateful
}

type modelV0 struct {
	Name     types.String `tfsdk:"name"`
	Id       types.String `tfsdk:"id"`
	Region   types.String `tfsdk:"region"`
	Rulesets types.List   `tfsdk:"rulesets"` //< rulesetModelV0
}

type rulesetModelV0 struct {
	Id               types.String  `tfsdk:"id"`
	Name             types.String  `tfsdk:"name"`
	Description      types.String  `tfsdk:"description"`
	Region           types.String  `tfsdk:"region"`
	IncludeByDefault types.Bool    `tfsdk:"include_by_default"`
	Rules            []ruleModelV0 `tfsdk:"rules"` //< ruleModelV0
}

type ruleModelV0 struct {
	Id          types.String `tfsdk:"id"`
	Source      types.String `tfsdk:"source"`
	Description types.String `tfsdk:"description"`
}

func modelToState(ctx context.Context, res *models.TrafficFilterRulesets, state *modelV0) diag.Diagnostics {
	var diags diag.Diagnostics
	var result = make([]rulesetModelV0, 0, len(res.Rulesets))

	for _, ruleset := range res.Rulesets {
		if *ruleset.Name != state.Name.ValueString() && *ruleset.ID != state.Id.ValueString() && *ruleset.Region != state.Region.ValueString() {
			continue
		}

		m := rulesetModelV0{
			Name:             types.StringValue(*ruleset.Name),
			Id:               types.StringValue(*ruleset.ID),
			Description:      types.StringValue(ruleset.Description),
			Region:           types.StringValue(*ruleset.Region),
			IncludeByDefault: types.BoolValue(*ruleset.IncludeByDefault),
		}

		var ruleArray = make([]ruleModelV0, 0, len(ruleset.Rules))
		for _, rule := range ruleset.Rules {
			t := ruleModelV0{
				Id:          types.StringValue(rule.ID),
				Source:      types.StringValue(rule.Source),
				Description: types.StringValue(rule.Description),
			}
			ruleArray = append(ruleArray, t)
		}
		if len(ruleArray) > 0 {
			m.Rules = ruleArray
		}

		result = append(result, m)
	}

	state.Rulesets, diags = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: rulesetAttrTypes()}, result)
	return diags
}

func rulesetAttrTypes() map[string]attr.Type {
	return rulesetSchema().GetType().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}

func rulesetElemType() attr.Type {
	return rulesetSchema().GetType().(types.ListType).ElemType
}

func ruleAttrTypes() map[string]attr.Type {
	return ruleSchema().GetType().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}

func ruleElemType() attr.Type {
	return ruleSchema().GetType().(types.ListType).ElemType
}
