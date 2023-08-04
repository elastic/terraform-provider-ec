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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_nodeRolesPlanModifier(t *testing.T) {
	type args struct {
		attributeState  []string
		attributePlan   []string
		deploymentState *deploymentv2.Deployment
		deploymentPlan  deploymentv2.Deployment
	}
	tests := []struct {
		name            string
		args            args
		expected        []string
		expectedUnknown bool
	}{
		{
			name: "it should keep current plan value if it's defined",
			args: args{
				attributePlan: []string{
					"data_content",
					"data_hot",
					"ingest",
					"master",
				},
			},
			expected: []string{
				"data_content",
				"data_hot",
				"ingest",
				"master",
			},
		},

		{
			name:            "it should not use state if state doesn't have `version`",
			args:            args{},
			expectedUnknown: true,
		},

		{
			name: "it should not use state if plan changed deployment template`",
			args: args{
				deploymentState: &deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
				},
				deploymentPlan: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-storage-optimized-v3",
				},
			},
			expectedUnknown: true,
		},

		{
			name: "it should not use state if plan version is less than 7.10.0 but the attribute state is not null`",
			args: args{
				attributeState: []string{"data_hot"},
				deploymentState: &deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
				},
				deploymentPlan: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Version:              "7.9.0",
				},
			},
			expectedUnknown: true,
		},

		{
			name: "it should not use state if plan version is changed over 7.10.0 and the attribute state is not null`",
			args: args{
				attributeState: []string{"data_hot"},
				deploymentState: &deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Version:              "7.9.0",
				},
				deploymentPlan: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Version:              "7.10.1",
				},
			},
			expectedUnknown: true,
		},

		{
			name: "it should use state if plan version is changed over 7.10.0 and the attribute state is null`",
			args: args{
				attributeState: nil,
				deploymentState: &deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Version:              "7.9.0",
				},
				deploymentPlan: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Version:              "7.10.1",
				},
			},
			expected: nil,
		},

		{
			name: "it should use state if both plan and state versions is or higher than 7.10.0 and the attribute state is not null`",
			args: args{
				attributeState: []string{"data_hot"},
				deploymentState: &deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Version:              "7.10.0",
				},
				deploymentPlan: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Version:              "7.10.0",
				},
			},
			expected: []string{"data_hot"},
		},

		{
			name: "it should not use state if both plan and state versions is or higher than 7.10.0 and the attribute state is null`",
			args: args{
				attributeState: nil,
				deploymentState: &deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Version:              "7.10.0",
				},
				deploymentPlan: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Version:              "7.10.0",
				},
			},
			expectedUnknown: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := v2.UseNodeRolesDefault()

			// attributeConfigValue := attrValueFromGoTypeValue(t, []string{}, types.SetType{ElemType: types.StringType})

			// attributeStateValue := attrValueFromGoTypeValue(t, tt.args.attributeState, types.SetType{ElemType: types.StringType})

			stateValue, diags := types.SetValueFrom(context.Background(), types.StringType, tt.args.attributeState)
			assert.Nil(t, diags)

			deploymentStateValue := tftypesValueFromGoTypeValue(t, tt.args.deploymentState, deploymentv2.DeploymentSchema().Type())

			deploymentPlanValue := tftypesValueFromGoTypeValue(t, tt.args.deploymentPlan, deploymentv2.DeploymentSchema().Type())

			req := planmodifier.SetRequest{
				// AttributeState:  attributeStateValue,
				// ConfigValue value is not used in the plan modifer,
				// it just should be known
				ConfigValue: types.SetValueMust(types.StringType, []attr.Value{}),
				StateValue:  stateValue,
				State: tfsdk.State{
					Raw:    deploymentStateValue,
					Schema: deploymentv2.DeploymentSchema(),
				},
				Plan: tfsdk.Plan{
					Raw:    deploymentPlanValue,
					Schema: deploymentv2.DeploymentSchema(),
				},
			}

			// the default plan value is `Unknown` ("known after apply")
			// the plan modifier either keeps this value or uses the current state
			// if test doesn't specify plan value, let's use the default (`Unknown`) value that is used by TF during plan modifier execution
			planValue := types.SetUnknown(types.StringType)
			if tt.args.attributePlan != nil {
				planValue, diags = types.SetValueFrom(context.Background(), types.StringType, tt.args.attributePlan)
				assert.Nil(t, diags)
			}

			resp := planmodifier.SetResponse{PlanValue: planValue}

			modifier.PlanModifySet(context.Background(), req, &resp)

			assert.Nil(t, resp.Diagnostics)

			if tt.expectedUnknown {
				assert.True(t, resp.PlanValue.IsUnknown(), "attributePlan should be unknown")
				return
			}

			var attributePlan []string

			diags = resp.PlanValue.ElementsAs(context.Background(), &attributePlan, true)

			assert.Nil(t, diags)

			assert.Equal(t, tt.expected, attributePlan)
		})
	}
}

func ptr[T any](t T) *T {
	return &t
}

func TestSetUnknownOnTopologySizeChange_PlanModifySet(t *testing.T) {
	tests := []struct {
		name              string
		setSizesToUnknown bool
		plan              *deploymentv2.Deployment
		state             *deploymentv2.Deployment
		planValue         types.Set
		expectedPlanValue types.Set
	}{
		{
			name:              "should do nothing if the plan value is unknown",
			planValue:         types.SetUnknown(types.StringType),
			plan:              &deploymentv2.Deployment{},
			state:             &deploymentv2.Deployment{},
			expectedPlanValue: types.SetUnknown(types.StringType),
		},
		{
			name:              "should do nothing if the plan value is null",
			planValue:         types.SetNull(types.StringType),
			plan:              &deploymentv2.Deployment{},
			state:             &deploymentv2.Deployment{},
			expectedPlanValue: types.SetNull(types.StringType),
		},
		{
			name:      "should do nothing if the only deployment topology size is unchanged",
			planValue: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("hot")}),
			plan: &deploymentv2.Deployment{
				Elasticsearch: &v2.Elasticsearch{
					HotTier: &v2.ElasticsearchTopology{
						Size:      ptr("1g"),
						ZoneCount: 3,
					},
				},
			},
			state: &deploymentv2.Deployment{
				Elasticsearch: &v2.Elasticsearch{
					HotTier: &v2.ElasticsearchTopology{
						Size:      ptr("1g"),
						ZoneCount: 3,
					},
				},
			},
			expectedPlanValue: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("hot")}),
		},
		{
			name:      "should set the plan value to unknown if the only deployment topology size has changed",
			planValue: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("hot")}),
			plan: &deploymentv2.Deployment{
				Elasticsearch: &v2.Elasticsearch{
					HotTier: &v2.ElasticsearchTopology{
						Size:      ptr("1g"),
						ZoneCount: 3,
					},
				},
			},
			state: &deploymentv2.Deployment{
				Elasticsearch: &v2.Elasticsearch{
					HotTier: &v2.ElasticsearchTopology{
						Size:      ptr("2g"),
						ZoneCount: 3,
					},
				},
			},
			expectedPlanValue: types.SetUnknown(types.StringType),
		},
		{
			name:      "should set the plan value to unknown if the only deployment topology zone count has changed",
			planValue: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("hot")}),
			plan: &deploymentv2.Deployment{
				Elasticsearch: &v2.Elasticsearch{
					HotTier: &v2.ElasticsearchTopology{
						Size:      ptr("1g"),
						ZoneCount: 3,
					},
				},
			},
			state: &deploymentv2.Deployment{
				Elasticsearch: &v2.Elasticsearch{
					HotTier: &v2.ElasticsearchTopology{
						Size:      ptr("1g"),
						ZoneCount: 2,
					},
				},
			},
			expectedPlanValue: types.SetUnknown(types.StringType),
		},
		{
			name:      "should set the plan value to unknown if the another deployment topology is added",
			planValue: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("hot")}),
			plan: &deploymentv2.Deployment{
				Elasticsearch: &v2.Elasticsearch{
					HotTier: &v2.ElasticsearchTopology{
						Size:      ptr("1g"),
						ZoneCount: 3,
					},
					WarmTier: &v2.ElasticsearchTopology{
						Size:      ptr("1g"),
						ZoneCount: 3,
					},
				},
			},
			state: &deploymentv2.Deployment{
				Elasticsearch: &v2.Elasticsearch{
					HotTier: &v2.ElasticsearchTopology{
						Size:      ptr("1g"),
						ZoneCount: 3,
					},
				},
			},
			expectedPlanValue: types.SetUnknown(types.StringType),
		},
		{
			name:      "should set the plan value to unknown if the another deployment topology size has changed",
			planValue: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("hot")}),
			plan: &deploymentv2.Deployment{
				Elasticsearch: &v2.Elasticsearch{
					HotTier: &v2.ElasticsearchTopology{
						Size:      ptr("1g"),
						ZoneCount: 3,
					},
					WarmTier: &v2.ElasticsearchTopology{
						Size:      ptr("1g"),
						ZoneCount: 3,
					},
				},
			},
			state: &deploymentv2.Deployment{
				Elasticsearch: &v2.Elasticsearch{
					HotTier: &v2.ElasticsearchTopology{
						Size:      ptr("1g"),
						ZoneCount: 3,
					},
					WarmTier: &v2.ElasticsearchTopology{
						Size:      ptr("2g"),
						ZoneCount: 3,
					},
				},
			},
			expectedPlanValue: types.SetUnknown(types.StringType),
		},
		{
			name:      "should set the plan value to unknown if the another deployment topology zone count has changed",
			planValue: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("hot")}),
			plan: &deploymentv2.Deployment{
				Elasticsearch: &v2.Elasticsearch{
					HotTier: &v2.ElasticsearchTopology{
						Size:      ptr("1g"),
						ZoneCount: 3,
					},
					WarmTier: &v2.ElasticsearchTopology{
						Size:      ptr("1g"),
						ZoneCount: 3,
					},
				},
			},
			state: &deploymentv2.Deployment{
				Elasticsearch: &v2.Elasticsearch{
					HotTier: &v2.ElasticsearchTopology{
						Size:      ptr("1g"),
						ZoneCount: 3,
					},
					WarmTier: &v2.ElasticsearchTopology{
						Size:      ptr("1g"),
						ZoneCount: 2,
					},
				},
			},
			expectedPlanValue: types.SetUnknown(types.StringType),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateValue := tftypesValueFromGoTypeValue(t, tt.state, deploymentv2.DeploymentSchema().Type())
			planValue := tftypesValueFromGoTypeValue(t, tt.plan, deploymentv2.DeploymentSchema().Type())
			req := planmodifier.SetRequest{
				PlanValue: tt.planValue,
				State: tfsdk.State{
					Raw:    stateValue,
					Schema: deploymentv2.DeploymentSchema(),
				},
				Plan: tfsdk.Plan{
					Raw:    planValue,
					Schema: deploymentv2.DeploymentSchema(),
				},
			}

			resp := planmodifier.SetResponse{
				PlanValue: tt.planValue,
			}
			modifier := v2.SetUnknownOnTopologySizeChange()
			modifier.PlanModifySet(context.Background(), req, &resp)

			require.Equal(t, tt.expectedPlanValue, resp.PlanValue)
		})
	}
}
