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

	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	apmv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/apm/v2"
	v2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/deployment/v2"
	integrationsserverv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/integrationsserver/v2"
	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifiers"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestUseStateForUnknownUnlessMigrationIsRequired_Apm_PlanModifyInt64(t *testing.T) {
	ver := 1
	tests := []struct {
		name              string
		resourceKind      string
		nullable          bool
		state             *v2.Deployment
		plan              v2.Deployment
		planValue         types.Int64
		stateValue        types.Int64
		configValue       types.Int64
		expectedPlanValue types.Int64
	}{
		{
			name:         "should update the plan value to null if apm has state null",
			resourceKind: "apm",
			nullable:     true,
			state: &v2.Deployment{
				Apm: &apmv2.Apm{
					InstanceConfigurationVersion: nil,
				},
			},
			plan: v2.Deployment{
				Apm: &apmv2.Apm{},
			},
			expectedPlanValue: types.Int64Null(),
			planValue:         types.Int64Unknown(),
			stateValue:        types.Int64Null(),
			configValue:       types.Int64Null(),
		},
		{
			name:         "should update the plan value if apm has state value",
			resourceKind: "apm",
			nullable:     true,
			state: &v2.Deployment{
				Apm: &apmv2.Apm{
					InstanceConfigurationVersion: &ver,
				},
			},
			plan: v2.Deployment{
				Apm: &apmv2.Apm{},
			},
			expectedPlanValue: types.Int64Value(1),
			planValue:         types.Int64Unknown(),
			stateValue:        types.Int64Value(1),
			configValue:       types.Int64Null(),
		},
		{
			name:         "should keep apm instance_configuration_version unknown when the resource is being created",
			resourceKind: "apm",
			nullable:     true,
			state:        &v2.Deployment{},
			plan: v2.Deployment{
				Apm: &apmv2.Apm{
					Size: ec.String("2g"),
				},
			},
			expectedPlanValue: types.Int64Unknown(),
			planValue:         types.Int64Unknown(),
			stateValue:        types.Int64Null(),
			configValue:       types.Int64Null(),
		},
		{
			name:         "should update the plan value to null if apm has state null",
			resourceKind: "integrations_server",
			nullable:     true,
			state: &v2.Deployment{
				IntegrationsServer: &integrationsserverv2.IntegrationsServer{
					InstanceConfigurationVersion: nil,
				},
			},
			plan: v2.Deployment{
				IntegrationsServer: &integrationsserverv2.IntegrationsServer{},
			},
			expectedPlanValue: types.Int64Null(),
			planValue:         types.Int64Unknown(),
			stateValue:        types.Int64Null(),
			configValue:       types.Int64Null(),
		},
		{
			name:         "should update the plan value if apm has state value",
			resourceKind: "integrations_server",
			nullable:     true,
			state: &v2.Deployment{
				IntegrationsServer: &integrationsserverv2.IntegrationsServer{
					InstanceConfigurationVersion: &ver,
				},
			},
			plan: v2.Deployment{
				IntegrationsServer: &integrationsserverv2.IntegrationsServer{},
			},
			expectedPlanValue: types.Int64Value(1),
			planValue:         types.Int64Unknown(),
			stateValue:        types.Int64Value(1),
			configValue:       types.Int64Null(),
		},
		{
			name:         "should keep apm instance_configuration_version unknown when the resource is being created",
			resourceKind: "integrations_server",
			nullable:     true,
			state:        &v2.Deployment{},
			plan: v2.Deployment{
				IntegrationsServer: &integrationsserverv2.IntegrationsServer{
					Size: ec.String("2g"),
				},
			},
			expectedPlanValue: types.Int64Unknown(),
			planValue:         types.Int64Unknown(),
			stateValue:        types.Int64Null(),
			configValue:       types.Int64Null(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateRaw := util.TfTypesValueFromGoTypeValue(t, tt.state, v2.DeploymentSchema().Type())
			planRaw := util.TfTypesValueFromGoTypeValue(t, tt.plan, v2.DeploymentSchema().Type())
			req := planmodifier.Int64Request{
				PlanValue:   tt.planValue,
				StateValue:  tt.stateValue,
				ConfigValue: tt.configValue,
				State: tfsdk.State{
					Raw:    stateRaw,
					Schema: v2.DeploymentSchema(),
				},
				Plan: tfsdk.Plan{
					Raw:    planRaw,
					Schema: v2.DeploymentSchema(),
				},
			}

			resp := planmodifier.Int64Response{
				PlanValue: tt.planValue,
			}
			modifier := planmodifiers.UseStateForUnknownUnlessMigrationIsRequired(tt.resourceKind, tt.nullable)

			modifier.PlanModifyInt64(context.Background(), req, &resp)

			require.Equal(t, tt.expectedPlanValue, resp.PlanValue)
		})
	}
}
