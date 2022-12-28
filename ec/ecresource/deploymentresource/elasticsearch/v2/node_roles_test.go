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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func Test_UseNodeRoles(t *testing.T) {
	type args struct {
		stateVersion string
		planVersion  string
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
				planVersion:  "7.0.0",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, diags := UseNodeRoles(types.String{Value: tt.args.stateVersion}, types.String{Value: tt.args.planVersion})

			if tt.expectedDiags == nil {
				assert.Nil(t, diags)
				assert.Equal(t, tt.expected, got)
			} else {
				assert.Equal(t, tt.expectedDiags, diags)
			}

		})
	}
}
