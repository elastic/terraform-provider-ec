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

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/models"
)

func modelToState(ctx context.Context, res *models.TrafficFilterRulesetInfo, state *modelV0) diag.Diagnostics {
	var diags diag.Diagnostics

	state.Name = types.String{Value: *res.Name}
	state.Region = types.String{Value: *res.Region}
	state.Type = types.String{Value: *res.Type}
	state.IncludeByDefault = types.Bool{Value: *res.IncludeByDefault}

	diags.Append(flattenRules(ctx, res.Rules, &state.Rule)...)

	if res.Description == "" {
		state.Description = types.String{Null: true}
	} else {
		state.Description = types.String{Value: res.Description}
	}

	return diags
}

func flattenRules(ctx context.Context, rules []*models.TrafficFilterRule, target interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var result = make([]trafficFilterRuleModelV0, 0, len(rules))
	for _, rule := range rules {
		model := trafficFilterRuleModelV0{
			ID:                types.String{Value: rule.ID},
			Source:            types.String{Null: true},
			Description:       types.String{Null: true},
			AzureEndpointGUID: types.String{Null: true},
			AzureEndpointName: types.String{Null: true},
		}

		if rule.Source != "" {
			model.Source = types.String{Value: rule.Source}
		}

		if rule.Description != "" {
			model.Description = types.String{Value: rule.Description}
		}

		if rule.AzureEndpointGUID != "" {
			model.AzureEndpointGUID = types.String{Value: rule.AzureEndpointGUID}
		}

		if rule.AzureEndpointName != "" {
			model.AzureEndpointName = types.String{Value: rule.AzureEndpointName}
		}

		result = append(result, model)
	}

	diags.Append(tfsdk.ValueFrom(ctx, result, trafficFilterRuleSetType(), target)...)

	return diags
}
