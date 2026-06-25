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

package serverlesstrafficfilterresource

import (
	"context"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_serverless_traffic_filter"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// toCreateRequest converts the Terraform model to the API create request
func toCreateRequest(ctx context.Context, model resource_serverless_traffic_filter.ServerlessTrafficFilterModel, diags *diag.Diagnostics) serverless.CreateTrafficFilterRequest {
	req := serverless.CreateTrafficFilterRequest{
		Name:   model.Name.ValueString(),
		Region: model.Region.ValueString(),
		Type:   serverless.TrafficFilterType(model.Type.ValueString()),
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		desc := model.Description.ValueString()
		req.Description = &desc
	}

	if !model.IncludeByDefault.IsNull() && !model.IncludeByDefault.IsUnknown() {
		includeByDefault := model.IncludeByDefault.ValueBool()
		req.IncludeByDefault = &includeByDefault
	}

	if !model.Rules.IsNull() && !model.Rules.IsUnknown() {
		rules := expandRules(ctx, model.Rules, diags)
		if len(rules) > 0 {
			req.Rules = &rules
		}
	}

	return req
}

// toPatchRequest converts the Terraform model to the API patch request
func toPatchRequest(ctx context.Context, model resource_serverless_traffic_filter.ServerlessTrafficFilterModel, diags *diag.Diagnostics) serverless.PatchTrafficFilterRequest {
	req := serverless.PatchTrafficFilterRequest{}

	if !model.Name.IsNull() && !model.Name.IsUnknown() {
		name := model.Name.ValueString()
		req.Name = &name
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		desc := model.Description.ValueString()
		req.Description = &desc
	}

	if !model.IncludeByDefault.IsNull() && !model.IncludeByDefault.IsUnknown() {
		includeByDefault := model.IncludeByDefault.ValueBool()
		req.IncludeByDefault = &includeByDefault
	}

	if !model.Rules.IsNull() && !model.Rules.IsUnknown() {
		rules := expandRules(ctx, model.Rules, diags)
		if len(rules) > 0 {
			req.Rules = &rules
		}
	}

	return req
}

// expandRules converts Terraform rules list to API rules slice using the generated RulesValue type
func expandRules(ctx context.Context, rulesList types.List, diags *diag.Diagnostics) []serverless.TrafficFilterRule {
	if rulesList.IsNull() || rulesList.IsUnknown() {
		return nil
	}

	var rulesValues []resource_serverless_traffic_filter.RulesValue
	d := rulesList.ElementsAs(ctx, &rulesValues, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	rules := make([]serverless.TrafficFilterRule, 0, len(rulesValues))
	for _, rv := range rulesValues {
		rule := serverless.TrafficFilterRule{
			Source: rv.Source.ValueString(),
		}
		if !rv.Description.IsNull() && !rv.Description.IsUnknown() {
			desc := rv.Description.ValueString()
			rule.Description = &desc
		}
		rules = append(rules, rule)
	}

	return rules
}

// fromTrafficFilterInfo converts API response to Terraform model
func fromTrafficFilterInfo(ctx context.Context, info serverless.TrafficFilterInfo, diags *diag.Diagnostics) resource_serverless_traffic_filter.ServerlessTrafficFilterModel {
	model := resource_serverless_traffic_filter.ServerlessTrafficFilterModel{
		Id:               types.StringValue(info.Id),
		Name:             types.StringValue(info.Name),
		Type:             types.StringValue(string(info.Type)),
		Region:           types.StringValue(info.Region),
		IncludeByDefault: types.BoolValue(info.IncludeByDefault),
	}

	if info.Description != nil {
		model.Description = types.StringValue(*info.Description)
	} else {
		model.Description = types.StringNull()
	}

	model.Rules = flattenRules(ctx, info.Rules, diags)

	return model
}

// flattenRules converts API rules slice to Terraform rules list using the generated RulesValue type
func flattenRules(ctx context.Context, rules []serverless.TrafficFilterRule, diags *diag.Diagnostics) types.List {
	elemType := resource_serverless_traffic_filter.RulesValue{}.Type(ctx)

	if rules == nil {
		return types.ListNull(elemType)
	}

	ruleObjects := make([]resource_serverless_traffic_filter.RulesValue, 0, len(rules))
	for _, rule := range rules {
		desc := basetypes.NewStringNull()
		if rule.Description != nil {
			desc = basetypes.NewStringValue(*rule.Description)
		}
		rv, d := resource_serverless_traffic_filter.NewRulesValue(
			resource_serverless_traffic_filter.RulesValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"source":      basetypes.NewStringValue(rule.Source),
				"description": desc,
			},
		)
		diags.Append(d...)
		ruleObjects = append(ruleObjects, rv)
	}

	listVal, d := types.ListValueFrom(ctx, elemType, ruleObjects)
	diags.Append(d...)
	return listVal
}
