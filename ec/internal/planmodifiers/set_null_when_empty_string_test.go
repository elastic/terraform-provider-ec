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

func TestSetNullWhenEmptyString(t *testing.T) {
	tests := []struct {
		name              string
		planValue         types.String
		expectedPlanValue types.String
	}{
		{
			name:              "should do nothing when the plan value is null",
			planValue:         types.StringNull(),
			expectedPlanValue: types.StringNull(),
		},
		{
			name:              "should do nothing when the plan value is unknown",
			planValue:         types.StringUnknown(),
			expectedPlanValue: types.StringUnknown(),
		},
		{
			name:              "should do nothing when the plan value is not empty",
			planValue:         types.StringValue("hello"),
			expectedPlanValue: types.StringValue("hello"),
		},
		{
			name:              "should set the plan value to null when the plan value is empty",
			planValue:         types.StringValue(""),
			expectedPlanValue: types.StringNull(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := planmodifiers.SetNullWhenEmptyString()

			resp := planmodifier.StringResponse{
				PlanValue: tt.planValue,
			}
			modifier.PlanModifyString(context.Background(), planmodifier.StringRequest{
				PlanValue: tt.planValue,
			}, &resp)

			require.Equal(t, tt.expectedPlanValue, resp.PlanValue)
		})
	}
}
