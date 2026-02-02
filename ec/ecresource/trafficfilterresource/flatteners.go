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
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/models"
)

func modelToState(ctx context.Context, res *models.TrafficFilterRulesetInfo, state *modelV0) diag.Diagnostics {
	state.Name = types.StringValue(*res.Name)
	state.Region = types.StringValue(*res.Region)
	state.Type = types.StringValue(*res.Type)
	state.IncludeByDefault = types.BoolValue(*res.IncludeByDefault)

	var diags diag.Diagnostics
	state.Rule, diags = flattenRules(ctx, res.Rules)

	if res.Description == "" {
		state.Description = types.StringNull()
	} else {
		state.Description = types.StringValue(res.Description)
	}

	return diags
}

func flattenRules(ctx context.Context, rules []*models.TrafficFilterRule) (types.Set, diag.Diagnostics) {
	var result = make([]trafficFilterRuleModelV0, 0, len(rules))
	for _, rule := range rules {
		model := trafficFilterRuleModelV0{
			ID:                 types.StringValue(rule.ID),
			Source:             types.StringNull(),
			Description:        types.StringNull(),
			AzureEndpointGUID:  types.StringNull(),
			AzureEndpointName:  types.StringNull(),
			RemoteClusterId:    types.StringNull(),
			RemoteClusterOrgId: types.StringNull(),
		}

		if rule.Source != "" {
			model.Source = types.StringValue(rule.Source)
		}

		if rule.Description != "" {
			model.Description = types.StringValue(rule.Description)
		}

		if rule.AzureEndpointGUID != "" {
			model.AzureEndpointGUID = types.StringValue(rule.AzureEndpointGUID)
		}

		if rule.AzureEndpointName != "" {
			model.AzureEndpointName = types.StringValue(rule.AzureEndpointName)
		}

		if rule.RemoteClusterID != "" {
			model.RemoteClusterId = types.StringValue(rule.RemoteClusterID)
		}

		if rule.RemoteClusterOrgID != "" {
			model.RemoteClusterOrgId = types.StringValue(rule.RemoteClusterOrgID)
		}

		result = append(result, model)
	}

	target, diags := types.SetValueFrom(ctx, trafficFilterRuleSetType().(types.SetType).ElementType(), result)

	return target, diags
}
