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
	"testing"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_serverless_traffic_filter"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:fix inline
func strPtr(s string) *string {
	return new(s)
}

func TestFromTrafficFilterInfo(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		info     serverless.TrafficFilterInfo
		validate func(t *testing.T, model resource_serverless_traffic_filter.ServerlessTrafficFilterModel)
	}{
		{
			name: "basic filter info",
			info: serverless.TrafficFilterInfo{
				Id:               "filter-123",
				Name:             "test-filter",
				Region:           "aws-us-east-1",
				Type:             serverless.Ip,
				Description:      new("Test description"),
				IncludeByDefault: false,
				Rules: []serverless.TrafficFilterRule{
					{
						Source:      "192.168.1.0/24",
						Description: new("Office network"),
					},
				},
			},
			validate: func(t *testing.T, model resource_serverless_traffic_filter.ServerlessTrafficFilterModel) {
				assert.Equal(t, "filter-123", model.Id.ValueString())
				assert.Equal(t, "test-filter", model.Name.ValueString())
				assert.Equal(t, "aws-us-east-1", model.Region.ValueString())
				assert.Equal(t, "ip", model.Type.ValueString())
				assert.Equal(t, "Test description", model.Description.ValueString())
				assert.Equal(t, false, model.IncludeByDefault.ValueBool())
				assert.False(t, model.Rules.IsNull())
				assert.Equal(t, 1, len(model.Rules.Elements()))
			},
		},
		{
			name: "filter without description",
			info: serverless.TrafficFilterInfo{
				Id:               "filter-456",
				Name:             "no-desc-filter",
				Region:           "aws-eu-west-1",
				Type:             serverless.Vpce,
				IncludeByDefault: true,
				Rules:            nil,
			},
			validate: func(t *testing.T, model resource_serverless_traffic_filter.ServerlessTrafficFilterModel) {
				assert.Equal(t, "filter-456", model.Id.ValueString())
				assert.True(t, model.Description.IsNull())
				assert.True(t, model.IncludeByDefault.ValueBool())
				assert.True(t, model.Rules.IsNull())
			},
		},
		{
			name: "filter with multiple rules",
			info: serverless.TrafficFilterInfo{
				Id:               "filter-789",
				Name:             "multi-rule",
				Region:           "aws-us-east-1",
				Type:             serverless.Ip,
				IncludeByDefault: false,
				Rules: []serverless.TrafficFilterRule{
					{Source: "10.0.0.0/8", Description: new("Internal")},
					{Source: "172.16.0.0/12"},
				},
			},
			validate: func(t *testing.T, model resource_serverless_traffic_filter.ServerlessTrafficFilterModel) {
				assert.Equal(t, 2, len(model.Rules.Elements()))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := fromTrafficFilterInfo(ctx, tt.info, &diags)

			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
			tt.validate(t, result)
		})
	}
}

func TestToCreateRequest(t *testing.T) {
	ctx := context.Background()

	t.Run("full model", func(t *testing.T) {
		rules := buildRulesList(ctx, t, []serverless.TrafficFilterRule{
			{Source: "10.0.0.0/8", Description: new("Internal")},
		})

		model := resource_serverless_traffic_filter.ServerlessTrafficFilterModel{
			Name:             types.StringValue("test-filter"),
			Region:           types.StringValue("aws-us-east-1"),
			Type:             types.StringValue("ip"),
			Description:      types.StringValue("A test filter"),
			IncludeByDefault: types.BoolValue(false),
			Rules:            rules,
		}

		var diags diag.Diagnostics
		req := toCreateRequest(ctx, model, &diags)
		require.False(t, diags.HasError())

		assert.Equal(t, "test-filter", req.Name)
		assert.Equal(t, "aws-us-east-1", req.Region)
		assert.Equal(t, serverless.Ip, req.Type)
		require.NotNil(t, req.Description)
		assert.Equal(t, "A test filter", *req.Description)
		require.NotNil(t, req.IncludeByDefault)
		assert.Equal(t, false, *req.IncludeByDefault)
		require.NotNil(t, req.Rules)
		assert.Len(t, *req.Rules, 1)
		assert.Equal(t, "10.0.0.0/8", (*req.Rules)[0].Source)
	})

	t.Run("minimal model with null optional fields", func(t *testing.T) {
		model := resource_serverless_traffic_filter.ServerlessTrafficFilterModel{
			Name:             types.StringValue("minimal"),
			Region:           types.StringValue("aws-us-east-1"),
			Type:             types.StringValue("vpce"),
			Description:      types.StringNull(),
			IncludeByDefault: types.BoolNull(),
			Rules:            types.ListNull(resource_serverless_traffic_filter.RulesValue{}.Type(ctx)),
		}

		var diags diag.Diagnostics
		req := toCreateRequest(ctx, model, &diags)
		require.False(t, diags.HasError())

		assert.Equal(t, "minimal", req.Name)
		assert.Nil(t, req.Description)
		assert.Nil(t, req.IncludeByDefault)
		assert.Nil(t, req.Rules)
	})
}

func TestToPatchRequest(t *testing.T) {
	ctx := context.Background()

	t.Run("sets all fields", func(t *testing.T) {
		model := resource_serverless_traffic_filter.ServerlessTrafficFilterModel{
			Name:             types.StringValue("updated-name"),
			Description:      types.StringValue("Updated desc"),
			IncludeByDefault: types.BoolValue(true),
			Rules:            types.ListNull(resource_serverless_traffic_filter.RulesValue{}.Type(ctx)),
		}

		var diags diag.Diagnostics
		req := toPatchRequest(ctx, model, &diags)
		require.False(t, diags.HasError())

		require.NotNil(t, req.Name)
		assert.Equal(t, "updated-name", *req.Name)
		require.NotNil(t, req.Description)
		assert.Equal(t, "Updated desc", *req.Description)
		require.NotNil(t, req.IncludeByDefault)
		assert.Equal(t, true, *req.IncludeByDefault)
		assert.Nil(t, req.Rules)
	})
}

func TestExpandRules(t *testing.T) {
	ctx := context.Background()

	t.Run("null list returns nil", func(t *testing.T) {
		var diags diag.Diagnostics
		result := expandRules(ctx, types.ListNull(resource_serverless_traffic_filter.RulesValue{}.Type(ctx)), &diags)
		require.False(t, diags.HasError())
		assert.Nil(t, result)
	})

	t.Run("converts rules with description", func(t *testing.T) {
		rules := buildRulesList(ctx, t, []serverless.TrafficFilterRule{
			{Source: "10.0.0.0/8", Description: new("Internal")},
			{Source: "172.16.0.0/12"},
		})

		var diags diag.Diagnostics
		result := expandRules(ctx, rules, &diags)
		require.False(t, diags.HasError())
		require.Len(t, result, 2)
		assert.Equal(t, "10.0.0.0/8", result[0].Source)
		require.NotNil(t, result[0].Description)
		assert.Equal(t, "Internal", *result[0].Description)
		assert.Equal(t, "172.16.0.0/12", result[1].Source)
		assert.Nil(t, result[1].Description)
	})
}

func TestFlattenRules(t *testing.T) {
	ctx := context.Background()

	t.Run("nil rules returns null list", func(t *testing.T) {
		var diags diag.Diagnostics
		result := flattenRules(ctx, nil, &diags)
		require.False(t, diags.HasError())
		assert.True(t, result.IsNull())
	})

	t.Run("empty rules returns empty list", func(t *testing.T) {
		var diags diag.Diagnostics
		result := flattenRules(ctx, []serverless.TrafficFilterRule{}, &diags)
		require.False(t, diags.HasError())
		assert.False(t, result.IsNull())
		assert.Equal(t, 0, len(result.Elements()))
	})

	t.Run("converts rules with and without description", func(t *testing.T) {
		var diags diag.Diagnostics
		result := flattenRules(ctx, []serverless.TrafficFilterRule{
			{Source: "10.0.0.0/8", Description: new("Internal")},
			{Source: "172.16.0.0/12"},
		}, &diags)
		require.False(t, diags.HasError())
		assert.Equal(t, 2, len(result.Elements()))
	})
}

// buildRulesList is a test helper that constructs a types.List of RulesValue from API rules.
func buildRulesList(ctx context.Context, t *testing.T, apiRules []serverless.TrafficFilterRule) types.List {
	t.Helper()
	elemType := resource_serverless_traffic_filter.RulesValue{}.Type(ctx)
	ruleObjects := make([]resource_serverless_traffic_filter.RulesValue, 0, len(apiRules))
	for _, rule := range apiRules {
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
		require.False(t, d.HasError())
		ruleObjects = append(ruleObjects, rv)
	}
	listVal, d := types.ListValueFrom(ctx, elemType, ruleObjects)
	require.False(t, d.HasError())
	return listVal
}
