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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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

func (d *DataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"name": {
				Type:        types.StringType,
				Description: "The name we are filtering on.",
				Optional:    true,
			},
			"id": {
				Type:        types.StringType,
				Description: "The id we are filtering on.",
				Optional:    true,
			},
			"region": {
				Type:        types.StringType,
				Description: "The region we are filtering on.",
				Optional:    true,
			},

			// computed fields
			"rulesets": rulesetsListSchema(),
		},
	}, nil
}

func rulesetsListSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "List of all Rulesets for this user.",
		Computed:    true,
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"id": {
				Type:        types.StringType,
				Description: "The ID of the ruleset",
				Computed:    true,
			},
			"name": {
				Type:        types.StringType,
				Description: "The name of the ruleset.",
				Computed:    true,
			},
			"description": {
				Type:        types.StringType,
				Description: "The description of the ruleset.",
				Computed:    true,
			},
			"region": {
				Type:        types.StringType,
				Description: "The ruleset can be attached only to deployments in the specific region.",
				Computed:    true,
			},
			"include_by_default": {
				Type:        types.BoolType,
				Description: "Should the ruleset be automatically included in the new deployments.",
				Computed:    true,
			},
			"rules": rulesetListSchema(),
		}),
	}
}

func rulesetListSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "List of rules in a ruleset",
		Computed:    true,
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"id": {
				Type:        types.StringType,
				Description: "The ID of the rule",
				Computed:    true,
			},
			"source": {
				Type:        types.StringType,
				Description: "Allowed traffic filter source: IP address, CIDR mask, or VPC endpoint ID.",
				Computed:    true,
			},
			"description": {
				Type:        types.StringType,
				Description: "The description of the rule.",
				Computed:    true,
			},
		}),
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
	response.TypeName = request.ProviderTypeName + "_trafficfilter"
}

func (d *DataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	client, diags := internal.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	d.client = client
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
		if *ruleset.Name != state.Name.Value && *ruleset.ID != state.Id.Value && *ruleset.Region != state.Region.Value {
			continue
		}

		m := rulesetModelV0{
			Name:             types.String{Value: *ruleset.Name},
			Id:               types.String{Value: *ruleset.ID},
			Description:      types.String{Value: ruleset.Description},
			Region:           types.String{Value: *ruleset.Region},
			IncludeByDefault: types.Bool{Value: *ruleset.IncludeByDefault},
		}

		var ruleArray = make([]ruleModelV0, 0, len(ruleset.Rules))
		for _, rule := range ruleset.Rules {
			t := ruleModelV0{
				Id:          types.String{Value: rule.ID},
				Source:      types.String{Value: rule.Source},
				Description: types.String{Value: rule.Description},
			}
			ruleArray = append(ruleArray, t)
		}

		m.Rules = ruleArray

		result = append(result, m)
	}

	diags.Append(tfsdk.ValueFrom(ctx, result, types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: rulesetAttrTypes(),
		},
	}, &state.Rulesets)...)

	return diags
}

func rulesetAttrTypes() map[string]attr.Type {
	return rulesetsListSchema().Attributes.Type().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}
