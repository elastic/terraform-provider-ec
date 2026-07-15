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

package v2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func Test_UseNodeRoles(t *testing.T) {
	type args struct {
		stateVersion  string
		planVersion   string
		elasticsearch Elasticsearch
	}
	tests := []struct {
		name          string
		args          args
		expected      bool
		expectedDiags diag.Diagnostics
	}{

		{
			name: "it should fail when plan version is invalid",
			args: args{
				stateVersion: "7.0.0",
				planVersion:  "invalid_plan_version",
			},
			expected: true,
			expectedDiags: func() diag.Diagnostics {
				var diags diag.Diagnostics
				diags.AddError("Failed to determine whether to use node_roles", "failed to parse Elasticsearch version: No Major.Minor.Patch elements found")
				return diags
			}(),
		},

		{
			name: "it should fail when state version is invalid",
			args: args{
				stateVersion: "invalid.state.version",
				planVersion:  "7.10.0",
			},
			expected: true,
			expectedDiags: func() diag.Diagnostics {
				var diags diag.Diagnostics
				diags.AddError("Failed to parse previous Elasticsearch version", `Invalid character(s) found in major number "invalid"`)
				return diags
			}(),
		},

		{
			name: "it should instruct to use node_types if both version are prior to 7.10.0",
			args: args{
				stateVersion: "7.9.0",
				planVersion:  "7.9.1",
			},
			expected: false,
		},

		{
			name: "it should instruct to use node_types if plan version is 7.10.0 and state version is prior to 7.10.0",
			args: args{
				stateVersion: "7.9.0",
				planVersion:  "7.10.0",
			},
			expected: false,
		},

		{
			name: "it should instruct to use node_types if plan version is after 7.10.0 and state version is prior to 7.10.0",
			args: args{
				stateVersion: "7.9.2",
				planVersion:  "7.10.1",
			},
			expected: false,
		},

		{
			name: "it should instruct to use node_types if plan version is after 7.10.0 and state version is prior to 7.10.0",
			args: args{
				stateVersion: "7.9.2",
				planVersion:  "7.10.1",
			},
			expected: false,
		},

		{
			name: "it should instruct to use node_roles if plan version is equal to state version and both is 7.10.0",
			args: args{
				stateVersion: "7.10.0",
				planVersion:  "7.10.0",
			},
			expected: true,
		},

		{
			name: "it should instruct to use node_roles if plan version is equal to state version and both is after 7.10.0",
			args: args{
				stateVersion: "7.10.2",
				planVersion:  "7.10.2",
			},
			expected: true,
		},

		{
			name: "it should instruct to use node_types if both plan version and state version are after 7.10.0 and plan uses node_types",
			args: args{
				stateVersion: "7.11.1",
				planVersion:  "7.12.0",
				elasticsearch: Elasticsearch{
					HotTier: &ElasticsearchTopology{
						id:             "hot_content",
						NodeTypeData:   new("true"),
						NodeTypeMaster: new("true"),
						NodeTypeIngest: new("true"),
						NodeTypeMl:     new("false"),
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var elasticsearchObject types.Object

			diags := tfsdk.ValueFrom(context.Background(), tt.args.elasticsearch, ElasticsearchSchema().GetType(), &elasticsearchObject)

			assert.Nil(t, diags)

			got, diags := UseNodeRoles(context.Background(), types.StringValue(tt.args.stateVersion), types.StringValue(tt.args.planVersion), elasticsearchObject)

			if tt.expectedDiags == nil {
				assert.Nil(t, diags)
				assert.Equal(t, tt.expected, got)
			} else {
				assert.Equal(t, tt.expectedDiags, diags)
			}

		})
	}
}

func Test_ValidateRollingZoneUpgrade(t *testing.T) {
	makeObj := func(strategy *string) types.Object {
		var obj types.Object
		diags := tfsdk.ValueFrom(context.Background(), Elasticsearch{Strategy: strategy}, ElasticsearchSchema().GetType(), &obj)
		if diags.HasError() {
			t.Fatalf("failed to build elasticsearch object: %v", diags)
		}
		return obj
	}

	rollingZone := strategyRollingZone
	rollingAll := strategyRollingAll

	tests := []struct {
		name         string
		stateVersion string
		planVersion  string
		strategy     *string
		wantError    bool
	}{
		{
			name:         "rolling_zone on minor upgrade is allowed",
			stateVersion: "8.14.0", planVersion: "8.15.0",
			strategy: &rollingZone, wantError: false,
		},
		{
			name:         "rolling_zone on major upgrade is rejected",
			stateVersion: "8.15.0", planVersion: "9.0.0",
			strategy: &rollingZone, wantError: true,
		},
		{
			name:         "rolling_all on major upgrade is allowed",
			stateVersion: "8.15.0", planVersion: "9.0.0",
			strategy: &rollingAll, wantError: false,
		},
		{
			name:         "nil strategy on major upgrade is allowed",
			stateVersion: "8.15.0", planVersion: "9.0.0",
			strategy: nil, wantError: false,
		},
		{
			name:         "empty stateVersion (new deployment) is a no-op",
			stateVersion: "", planVersion: "9.0.0",
			strategy: &rollingZone, wantError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := ValidateRollingZoneUpgrade(
				context.Background(),
				types.StringValue(tt.stateVersion),
				types.StringValue(tt.planVersion),
				makeObj(tt.strategy),
			)
			if tt.wantError {
				assert.True(t, diags.HasError(), "expected an error diagnostic")
			} else {
				assert.False(t, diags.HasError(), "expected no error diagnostic, got: %v", diags)
			}
		})
	}
}
