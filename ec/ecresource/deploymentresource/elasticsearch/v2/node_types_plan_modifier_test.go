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

	deploymentv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/deployment/v2"
	v2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/testutil"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func Test_nodeTypesPlanModifier(t *testing.T) {
	type args struct {
		attributeState  types.String
		attributePlan   types.String
		deploymentState *deploymentv2.Deployment
		deploymentPlan  deploymentv2.Deployment
	}
	tests := []struct {
		name     string
		args     args
		expected types.String
	}{
		{
			name: "it should keep current plan value if it's defined",
			args: args{
				attributePlan: types.StringValue("some value"),
			},
			expected: types.StringValue("some value"),
		},

		{
			name: "it should not use state if state doesn't have `version`",
			args: args{
				attributePlan: types.StringUnknown(),
			},
			expected: types.StringUnknown(),
		},

		{
			name: "it should not use state if plan changed deployment template`",
			args: args{
				attributePlan: types.StringUnknown(),
				deploymentState: &deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
				},
				deploymentPlan: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-storage-optimized-v3",
				},
			},
			expected: types.StringUnknown(),
		},

		{
			name: "it should not use state if plan version is less than 7.10.0 but the attribute state is null`",
			args: args{
				attributePlan:  types.StringUnknown(),
				attributeState: types.StringNull(),
				deploymentState: &deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
				},
				deploymentPlan: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Version:              "7.9.0",
				},
			},
			expected: types.StringUnknown(),
		},

		{
			name: "it should not use state if plan version is changed over 7.10.0, but the attribute state is null`",
			args: args{
				attributePlan:  types.StringUnknown(),
				attributeState: types.StringNull(),
				deploymentState: &deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Version:              "7.9.0",
				},
				deploymentPlan: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Version:              "7.10.1",
				},
			},
			expected: types.StringUnknown(),
		},

		{
			name: "it should not use state if both plan and state versions is or higher than 7.10.0, but the attribute state is not null`",
			args: args{
				attributePlan:  types.StringUnknown(),
				attributeState: types.StringValue("false"),
				deploymentState: &deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Version:              "7.10.0",
				},
				deploymentPlan: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Version:              "7.10.0",
				},
			},
			expected: types.StringUnknown(),
		},

		{
			name: "it should use state if both plan and state versions is or higher than 7.10.0 and the attribute state is null`",
			args: args{
				attributePlan:  types.StringUnknown(),
				attributeState: types.StringNull(),
				deploymentState: &deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Version:              "7.10.0",
				},
				deploymentPlan: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Version:              "7.10.0",
				},
			},
			expected: types.StringNull(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := v2.UseNodeTypesDefault()

			deploymentStateValue := testutil.TfTypesValueFromGoTypeValue(t, tt.args.deploymentState, deploymentv2.DeploymentSchema().Type())

			deploymentPlanValue := testutil.TfTypesValueFromGoTypeValue(t, tt.args.deploymentPlan, deploymentv2.DeploymentSchema().Type())

			req := planmodifier.StringRequest{
				// ConfigValue value is not used in the plan modifer,
				// it just should be known
				ConfigValue: types.StringValue(""),
				StateValue:  tt.args.attributeState,
				State: tfsdk.State{
					Raw:    deploymentStateValue,
					Schema: deploymentv2.DeploymentSchema(),
				},
				Plan: tfsdk.Plan{
					Raw:    deploymentPlanValue,
					Schema: deploymentv2.DeploymentSchema(),
				},
			}

			resp := planmodifier.StringResponse{PlanValue: tt.args.attributePlan}

			modifier.PlanModifyString(context.Background(), req, &resp)

			assert.Nil(t, resp.Diagnostics)

			assert.Equal(t, tt.expected, resp.PlanValue)
		})
	}
}
