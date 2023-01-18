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

/*
import (
	"context"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	deploymentv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/deployment/v2"
	v2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

func Test_topologyPlanModifier(t *testing.T) {
	type args struct {
		// the actual attribute type doesn't matter
		attributeState  types.String
		attributePlan   types.String
		deploymentState deploymentv2.Deployment
		deploymentPlan  deploymentv2.Deployment
	}
	tests := []struct {
		name     string
		args     args
		expected types.String
	}{
		{
			name: "it should keep the current plan value if the plan is known",
			args: args{
				attributeState: types.String{Value: "state value"},
				attributePlan:  types.String{Value: "plan value"},
			},
			expected: types.String{Value: "plan value"},
		},

		{
			name: "it should not use state if there is no such topology in the state",
			args: args{
				attributeState: types.String{Null: true},
				attributePlan:  types.String{Unknown: true},
				deploymentState: deploymentv2.Deployment{
					Elasticsearch: &v2.Elasticsearch{},
				},
			},
			expected: types.String{Unknown: true},
		},

		{
			name: "it should not use state if the plan changed the template attribute",
			args: args{
				attributeState: types.String{Value: "1g"},
				attributePlan:  types.String{Unknown: true},
				deploymentState: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Elasticsearch: &v2.Elasticsearch{
						Topology: v2.ElasticsearchTopologies{
							*v2.CreateTierForTest("hot_content", v2.ElasticsearchTopology{
								Autoscaling: &v2.ElasticsearchTopologyAutoscaling{
									MinSize: ec.String("1g"),
								},
							}),
						},
					},
				},
				deploymentPlan: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-storage-optimized-v3",
					Elasticsearch: &v2.Elasticsearch{
						Topology: v2.ElasticsearchTopologies{
							*v2.CreateTierForTest("hot_content", v2.ElasticsearchTopology{
								Autoscaling: &v2.ElasticsearchTopologyAutoscaling{},
							}),
						},
					},
				},
			},
			expected: types.String{Unknown: true},
		},

		{
			name: "it should use the current state if the state is null, the topology is defined in the state and the template has not changed",
			args: args{
				attributeState: types.String{Null: true},
				attributePlan:  types.String{Unknown: true},
				deploymentState: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Elasticsearch: &v2.Elasticsearch{
						Topology: v2.ElasticsearchTopologies{
							*v2.CreateTierForTest("hot_content", v2.ElasticsearchTopology{
								Autoscaling: &v2.ElasticsearchTopologyAutoscaling{},
							}),
						},
					},
				},
				deploymentPlan: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Elasticsearch: &v2.Elasticsearch{
						Topology: v2.ElasticsearchTopologies{
							*v2.CreateTierForTest("hot_content", v2.ElasticsearchTopology{
								Autoscaling: &v2.ElasticsearchTopologyAutoscaling{},
							}),
						},
					},
				},
			},
			expected: types.String{Null: true},
		},

		{
			name: "it should use the current state if the topology is defined in the state and the template has not changed",
			args: args{
				attributeState: types.String{Value: "1g"},
				attributePlan:  types.String{Unknown: true},
				deploymentState: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Elasticsearch: &v2.Elasticsearch{
						Topology: v2.ElasticsearchTopologies{
							*v2.CreateTierForTest("hot_content", v2.ElasticsearchTopology{
								Autoscaling: &v2.ElasticsearchTopologyAutoscaling{
									MaxSize: ec.String("1g"),
								},
							}),
						},
					},
				},
				deploymentPlan: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Elasticsearch: &v2.Elasticsearch{
						Topology: v2.ElasticsearchTopologies{
							*v2.CreateTierForTest("hot_content", v2.ElasticsearchTopology{
								Autoscaling: &v2.ElasticsearchTopologyAutoscaling{},
							}),
						},
					},
				},
			},
			expected: types.String{Value: "1g"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := v2.UseTopologyStateForUnknown("hot")

			deploymentStateValue := tftypesValueFromGoTypeValue(t, tt.args.deploymentState, deploymentv2.DeploymentSchema().Type())

			deploymentPlanValue := tftypesValueFromGoTypeValue(t, tt.args.deploymentPlan, deploymentv2.DeploymentSchema().Type())

			req := tfsdk.ModifyAttributePlanRequest{
				// attributeConfig value is not used in the plan modifer
				// it just should be known
				AttributeConfig: types.String{},
				AttributeState:  tt.args.attributeState,
				State: tfsdk.State{
					Raw:    deploymentStateValue,
					Schema: deploymentv2.DeploymentSchema(),
				},
				Plan: tfsdk.Plan{
					Raw:    deploymentPlanValue,
					Schema: deploymentv2.DeploymentSchema(),
				},
			}

			resp := tfsdk.ModifyAttributePlanResponse{AttributePlan: tt.args.attributePlan}

			modifier.Modify(context.Background(), req, &resp)

			assert.Nil(t, resp.Diagnostics)

			assert.Equal(t, tt.expected, resp.AttributePlan)
		})
	}
}

func attrValueFromGoTypeValue(t *testing.T, goValue any, attributeType attr.Type) attr.Value {
	var attrValue attr.Value
	diags := tfsdk.ValueFrom(context.Background(), goValue, attributeType, &attrValue)
	assert.Nil(t, diags)
	return attrValue
}

func tftypesValueFromGoTypeValue(t *testing.T, goValue any, attributeType attr.Type) tftypes.Value {
	attrValue := attrValueFromGoTypeValue(t, goValue, attributeType)
	tftypesValue, err := attrValue.ToTerraformValue(context.Background())
	assert.Nil(t, err)
	return tftypesValue
}

func unknownValueFromAttrType(t *testing.T, attributeType attr.Type) attr.Value {
	tfVal := tftypes.NewValue(attributeType.TerraformType(context.Background()), tftypes.UnknownValue)
	val, err := attributeType.ValueFromTerraform(context.Background(), tfVal)
	assert.Nil(t, err)
	return val
}
*/
