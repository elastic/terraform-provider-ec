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

package planmodifiers_test

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestUseStateForUnknownOrUnknown(t *testing.T) {
	tests := []struct {
		name              string
		configValue       types.String
		stateValue        types.String
		planValue         types.String
		expectedPlanValue types.String
	}{
		{
			name:              "keeps configured value",
			configValue:       types.StringValue("configured"),
			stateValue:        types.StringValue("state"),
			planValue:         types.StringValue("configured"),
			expectedPlanValue: types.StringValue("configured"),
		},
		{
			name:              "uses prior state when plan is unknown",
			configValue:       types.StringNull(),
			stateValue:        types.StringValue("enabled"),
			planValue:         types.StringUnknown(),
			expectedPlanValue: types.StringValue("enabled"),
		},
		{
			name:              "uses prior state even when plan is null",
			configValue:       types.StringNull(),
			stateValue:        types.StringValue("enabled"),
			planValue:         types.StringNull(),
			expectedPlanValue: types.StringValue("enabled"),
		},
		{
			name:              "leaves plan unknown when no prior state and no config",
			configValue:       types.StringNull(),
			stateValue:        types.StringNull(),
			planValue:         types.StringNull(),
			expectedPlanValue: types.StringUnknown(),
		},
		{
			name:              "leaves plan unknown when prior state is unknown",
			configValue:       types.StringNull(),
			stateValue:        types.StringUnknown(),
			planValue:         types.StringNull(),
			expectedPlanValue: types.StringUnknown(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := planmodifiers.UseStateForUnknownOrUnknown()

			resp := planmodifier.StringResponse{
				PlanValue: tt.planValue,
			}
			modifier.PlanModifyString(context.Background(), planmodifier.StringRequest{
				ConfigValue: tt.configValue,
				StateValue:  tt.stateValue,
				PlanValue:   tt.planValue,
			}, &resp)

			require.Equal(t, tt.expectedPlanValue, resp.PlanValue)
		})
	}
}
