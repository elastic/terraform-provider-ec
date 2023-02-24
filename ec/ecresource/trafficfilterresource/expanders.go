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

	"github.com/elastic/cloud-sdk-go/pkg/models"
)

func expandModel(ctx context.Context, state modelV0) (*models.TrafficFilterRulesetRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	ruleSet := make([]trafficFilterRuleModelV0, 0, len(state.Rule.Elems))
	diags.Append(state.Rule.ElementsAs(ctx, &ruleSet, false)...)
	if diags.HasError() {
		return nil, diags
	}

	var request = models.TrafficFilterRulesetRequest{
		Name:             &state.Name.Value,
		Type:             &state.Type.Value,
		Region:           &state.Region.Value,
		Description:      state.Description.Value,
		IncludeByDefault: &state.IncludeByDefault.Value,
		Rules:            make([]*models.TrafficFilterRule, 0, len(ruleSet)),
	}

	for _, r := range ruleSet {
		var rule = models.TrafficFilterRule{
			Source: r.Source.Value,
		}

		if !r.ID.IsNull() && !r.ID.IsUnknown() {
			rule.ID = r.ID.Value
		}

		if !r.Description.IsNull() && !r.Description.IsUnknown() {
			rule.Description = r.Description.Value
		}

		if !r.AzureEndpointName.IsNull() && !r.AzureEndpointName.IsUnknown() {
			rule.AzureEndpointName = r.AzureEndpointName.Value
		}
		if !r.AzureEndpointGUID.IsNull() && !r.AzureEndpointGUID.IsUnknown() {
			rule.AzureEndpointGUID = r.AzureEndpointGUID.Value
		}

		request.Rules = append(request.Rules, &rule)
	}

	return &request, diags
}
