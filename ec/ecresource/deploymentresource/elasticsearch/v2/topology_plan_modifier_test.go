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

	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	deploymentv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/deployment/v2"
	v2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/testutil"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func Test_topologyPlanModifier(t *testing.T) {
	type args struct {
		// the actual attribute type doesn't matter
		attributePlan   types.String
		deploymentState deploymentv2.Deployment
		deploymentPlan  deploymentv2.Deployment
	}
	tests := []struct {
		name               string
		args               args
		expectedToUseState bool
	}{
		{
			name: "it should keep the current plan value if the plan is known",
			args: args{
				attributePlan: types.StringValue("plan value"),
			},
			expectedToUseState: false,
		},

		{
			name: "it should not use state if there is no such topology in the state",
			args: args{
				attributePlan: types.StringUnknown(),
				deploymentState: deploymentv2.Deployment{
					Elasticsearch: &v2.Elasticsearch{},
				},
			},
			expectedToUseState: false,
		},

		{
			name: "it should not use state if the plan changed the template attribute",
			args: args{
				attributePlan: types.StringUnknown(),
				deploymentState: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Elasticsearch: &v2.Elasticsearch{
						HotTier: v2.CreateTierForTest("hot_content", v2.ElasticsearchTopology{
							Autoscaling: &v2.ElasticsearchTopologyAutoscaling{
								MinSize: ec.String("1g"),
							},
						}),
					},
				},
				deploymentPlan: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-storage-optimized-v3",
					Elasticsearch: &v2.Elasticsearch{
						HotTier: v2.CreateTierForTest("hot_content", v2.ElasticsearchTopology{
							Autoscaling: &v2.ElasticsearchTopologyAutoscaling{},
						}),
					},
				},
			},
			expectedToUseState: false,
		},

		{
			name: "it should not use state if the migrate_to_latest_hardware is true and migration is available",
			args: args{
				attributePlan: types.StringUnknown(),
				deploymentState: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Elasticsearch: &v2.Elasticsearch{
						HotTier: v2.CreateTierForTest("hot_content", v2.ElasticsearchTopology{
							InstanceConfigurationVersion:       ec.Int(0),
							LatestInstanceConfigurationVersion: ec.Int(1),
							Autoscaling: &v2.ElasticsearchTopologyAutoscaling{
								MinSize: ec.String("1g"),
							},
						}),
					},
				},
				deploymentPlan: deploymentv2.Deployment{
					DeploymentTemplateId:    "aws-io-optimized-v2",
					MigrateToLatestHardware: ec.Bool(true),
					Elasticsearch: &v2.Elasticsearch{
						HotTier: v2.CreateTierForTest("hot_content", v2.ElasticsearchTopology{
							Autoscaling: &v2.ElasticsearchTopologyAutoscaling{},
						}),
					},
				},
			},
			expectedToUseState: false,
		},

		{
			name: "it should use state if the migrate_to_latest_hardware is true but migration is not available",
			args: args{
				attributePlan: types.StringUnknown(),
				deploymentState: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Elasticsearch: &v2.Elasticsearch{
						HotTier: v2.CreateTierForTest("hot_content", v2.ElasticsearchTopology{
							InstanceConfigurationId:            ec.String("aws.data.highio.i3"),
							LatestInstanceConfigurationId:      ec.String("aws.data.highio.i3"),
							InstanceConfigurationVersion:       ec.Int(0),
							LatestInstanceConfigurationVersion: ec.Int(0),
							Autoscaling: &v2.ElasticsearchTopologyAutoscaling{
								MinSize: ec.String("1g"),
							},
						}),
					},
				},
				deploymentPlan: deploymentv2.Deployment{
					DeploymentTemplateId:    "aws-io-optimized-v2",
					MigrateToLatestHardware: ec.Bool(true),
					Elasticsearch: &v2.Elasticsearch{
						HotTier: v2.CreateTierForTest("hot_content", v2.ElasticsearchTopology{
							Autoscaling: &v2.ElasticsearchTopologyAutoscaling{},
						}),
					},
				},
			},
			expectedToUseState: true,
		},

		{
			name: "it should use state if IC version is defined for the topology element, even if migration is available",
			args: args{
				attributePlan: types.StringUnknown(),
				deploymentState: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Elasticsearch: &v2.Elasticsearch{
						HotTier: v2.CreateTierForTest("hot_content", v2.ElasticsearchTopology{
							InstanceConfigurationVersion:       ec.Int(0),
							LatestInstanceConfigurationVersion: ec.Int(1),
							Autoscaling: &v2.ElasticsearchTopologyAutoscaling{
								MinSize: ec.String("1g"),
							},
						}),
					},
				},
				deploymentPlan: deploymentv2.Deployment{
					DeploymentTemplateId:    "aws-io-optimized-v2",
					MigrateToLatestHardware: ec.Bool(true),
					Elasticsearch: &v2.Elasticsearch{
						HotTier: v2.CreateTierForTest("hot_content", v2.ElasticsearchTopology{
							InstanceConfigurationVersion: ec.Int(1),
							Autoscaling:                  &v2.ElasticsearchTopologyAutoscaling{},
						}),
					},
				},
			},
			expectedToUseState: true,
		},

		{
			name: "it should use state if IC ID is defined for the topology element, even if migration is available",
			args: args{
				attributePlan: types.StringUnknown(),
				deploymentState: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Elasticsearch: &v2.Elasticsearch{
						HotTier: v2.CreateTierForTest("hot_content", v2.ElasticsearchTopology{
							InstanceConfigurationVersion:       ec.Int(0),
							LatestInstanceConfigurationVersion: ec.Int(1),
							Autoscaling: &v2.ElasticsearchTopologyAutoscaling{
								MinSize: ec.String("1g"),
							},
						}),
					},
				},
				deploymentPlan: deploymentv2.Deployment{
					DeploymentTemplateId:    "aws-io-optimized-v2",
					MigrateToLatestHardware: ec.Bool(true),
					Elasticsearch: &v2.Elasticsearch{
						HotTier: v2.CreateTierForTest("hot_content", v2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("aws.data.highio.c5d"),
							Autoscaling:             &v2.ElasticsearchTopologyAutoscaling{},
						}),
					},
				},
			},
			expectedToUseState: true,
		},

		{
			name: "it should use the current state if the topology is defined in the state, the template has not changed, and migrate_to_latest_hardware is undefined",
			args: args{
				attributePlan: types.StringUnknown(),
				deploymentState: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Elasticsearch: &v2.Elasticsearch{
						HotTier: v2.CreateTierForTest("hot_content", v2.ElasticsearchTopology{
							InstanceConfigurationId:       ec.String("aws.data.highio.i3"),
							LatestInstanceConfigurationId: ec.String("aws.data.highio.c5d"),
							Autoscaling: &v2.ElasticsearchTopologyAutoscaling{
								MaxSize: ec.String("1g"),
							},
						}),
					},
				},
				deploymentPlan: deploymentv2.Deployment{
					DeploymentTemplateId: "aws-io-optimized-v2",
					Elasticsearch: &v2.Elasticsearch{
						HotTier: v2.CreateTierForTest("hot_content", v2.ElasticsearchTopology{
							Autoscaling: &v2.ElasticsearchTopologyAutoscaling{},
						}),
					},
				},
			},
			expectedToUseState: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := v2.UseTopologyStateForUnknown("hot")

			deploymentStateValue := testutil.TfTypesValueFromGoTypeValue(t, tt.args.deploymentState, deploymentv2.DeploymentSchema().Type())

			deploymentPlanValue := testutil.TfTypesValueFromGoTypeValue(t, tt.args.deploymentPlan, deploymentv2.DeploymentSchema().Type())

			plan := tfsdk.Plan{
				Raw:    deploymentPlanValue,
				Schema: deploymentv2.DeploymentSchema(),
			}

			state := tfsdk.State{
				Raw:    deploymentStateValue,
				Schema: deploymentv2.DeploymentSchema(),
			}

			useState, diags := modifier.UseState(context.Background(), types.String{}, plan, state, tt.args.attributePlan)

			assert.Nil(t, diags)

			assert.Equal(t, tt.expectedToUseState, useState, func() string {
				if tt.expectedToUseState {
					return "it's expected to use state but it doesn't"
				}
				return "it's not expected to use state but it does"
			}())
		})
	}
}
