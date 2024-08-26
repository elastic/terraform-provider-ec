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

package v2_test

import (
	"context"
	"testing"

	apmv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/apm/v2"
	v2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/deployment/v2"
	integrationsserverv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/integrationsserver/v2"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestUseNullUnlessAddingAPMOrIntegrationsServer_PlanModifyString(t *testing.T) {
	tests := []struct {
		name              string
		planValue         types.String
		state             *v2.Deployment
		plan              v2.Deployment
		expectedPlanValue types.String
	}{
		{
			name:              "should update the plan value to null if neither the plan nor state define either apm or integrations server",
			state:             &v2.Deployment{},
			plan:              v2.Deployment{},
			expectedPlanValue: types.StringNull(),
			planValue:         types.StringUnknown(),
		},
		{
			name: "should update the plan value to null if apm exists in both the plan and state",
			state: &v2.Deployment{
				Apm: &apmv2.Apm{},
			},
			plan: v2.Deployment{
				Apm: &apmv2.Apm{},
			},
			expectedPlanValue: types.StringNull(),
			planValue:         types.StringUnknown(),
		},
		{
			name: "should update the plan value to null if integrations server exists in both the plan and state",
			state: &v2.Deployment{
				IntegrationsServer: &integrationsserverv2.IntegrationsServer{},
			},
			plan: v2.Deployment{
				IntegrationsServer: &integrationsserverv2.IntegrationsServer{},
			},
			expectedPlanValue: types.StringNull(),
			planValue:         types.StringUnknown(),
		},
		{
			name:  "should do nothing if the plan value is known",
			state: &v2.Deployment{},
			plan: v2.Deployment{
				IntegrationsServer: &integrationsserverv2.IntegrationsServer{},
			},
			expectedPlanValue: types.StringValue("sekret"),
			planValue:         types.StringValue("sekret"),
		},
		{
			name:  "should do nothing if the plan value is null",
			state: &v2.Deployment{},
			plan: v2.Deployment{
				IntegrationsServer: &integrationsserverv2.IntegrationsServer{},
			},
			expectedPlanValue: types.StringNull(),
			planValue:         types.StringNull(),
		},
		{
			name:  "should do nothing if the plan value is unknown and the plan adds an apm resource",
			state: &v2.Deployment{},
			plan: v2.Deployment{
				Apm: &apmv2.Apm{},
			},
			expectedPlanValue: types.StringUnknown(),
			planValue:         types.StringUnknown(),
		},
		{
			name:  "should do nothing if the plan value is unknown and the plan adds an integrations server resource",
			state: &v2.Deployment{},
			plan: v2.Deployment{
				IntegrationsServer: &integrationsserverv2.IntegrationsServer{},
			},
			expectedPlanValue: types.StringUnknown(),
			planValue:         types.StringUnknown(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateValue := util.TfTypesValueFromGoTypeValue(t, tt.state, v2.DeploymentSchema().Type())
			planValue := util.TfTypesValueFromGoTypeValue(t, tt.plan, v2.DeploymentSchema().Type())
			req := planmodifier.StringRequest{
				PlanValue: tt.planValue,
				State: tfsdk.State{
					Raw:    stateValue,
					Schema: v2.DeploymentSchema(),
				},
				Plan: tfsdk.Plan{
					Raw:    planValue,
					Schema: v2.DeploymentSchema(),
				},
			}

			resp := planmodifier.StringResponse{
				PlanValue: tt.planValue,
			}
			modifier := v2.UseNullUnlessAddingAPMOrIntegrationsServer()

			modifier.PlanModifyString(context.Background(), req, &resp)

			require.Equal(t, tt.expectedPlanValue, resp.PlanValue)
		})
	}
}
