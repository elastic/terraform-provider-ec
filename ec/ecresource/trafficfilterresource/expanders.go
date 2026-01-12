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
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
)

func expandModel(ctx context.Context, state modelV0) (*models.TrafficFilterRulesetRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	ruleSet := make([]trafficFilterRuleModelV0, 0, len(state.Rule.Elements()))
	diags.Append(state.Rule.ElementsAs(ctx, &ruleSet, false)...)
	if diags.HasError() {
		return nil, diags
	}

	var request = models.TrafficFilterRulesetRequest{
		Name:             ec.String(state.Name.ValueString()),
		Type:             ec.String(state.Type.ValueString()),
		Region:           ec.String(state.Region.ValueString()),
		Description:      *ec.String(state.Description.ValueString()),
		IncludeByDefault: ec.Bool(state.IncludeByDefault.ValueBool()),
		Rules:            make([]*models.TrafficFilterRule, 0, len(ruleSet)),
	}

	for _, r := range ruleSet {
		var rule = models.TrafficFilterRule{
			Source: r.Source.ValueString(),
		}

		if !r.ID.IsNull() && !r.ID.IsUnknown() {
			rule.ID = r.ID.ValueString()
		}

		if !r.Description.IsNull() && !r.Description.IsUnknown() {
			rule.Description = r.Description.ValueString()
		}

		if !r.AzureEndpointName.IsNull() && !r.AzureEndpointName.IsUnknown() {
			rule.AzureEndpointName = r.AzureEndpointName.ValueString()
		}
		if !r.AzureEndpointGUID.IsNull() && !r.AzureEndpointGUID.IsUnknown() {
			rule.AzureEndpointGUID = r.AzureEndpointGUID.ValueString()
		}

		if !r.RemoteClusterId.IsNull() && !r.RemoteClusterId.IsUnknown() {
			rule.RemoteClusterID = r.RemoteClusterId.ValueString()
		}
		if !r.RemoteClusterOrgId.IsNull() && !r.RemoteClusterOrgId.IsUnknown() {
			rule.RemoteClusterOrgID = r.RemoteClusterOrgId.ValueString()
		}

		request.Rules = append(request.Rules, &rule)
	}

	return &request, diags
}
