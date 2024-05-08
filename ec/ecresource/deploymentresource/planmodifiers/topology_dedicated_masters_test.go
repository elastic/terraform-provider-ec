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

package planmodifiers

import (
	"context"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	depl "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/deployment/v2"
	es "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/testutil"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdateDedicatedMasterTier(t *testing.T) {
	type args struct {
		plan                 es.Elasticsearch
		config               es.Elasticsearch
		deploymentTemplate   models.DeploymentTemplateInfoV2
		masterInstanceConfig models.InstanceConfiguration
		migrateToLatestHw    bool
	}
	tests := []struct {
		name         string
		args         args
		expectedPlan es.Elasticsearch
	}{
		{
			name: "Should add master tier for 6 nodes",
			args: args{
				plan: es.Elasticsearch{
					HotTier: &es.ElasticsearchTopology{
						Size:      ec.String("1g"),
						ZoneCount: 3,
					},
					WarmTier: &es.ElasticsearchTopology{
						Size:      ec.String("1g"),
						ZoneCount: 3,
					},
				},
				config:               es.Elasticsearch{},
				deploymentTemplate:   deploymentTemplate(),
				masterInstanceConfig: masterInstanceConfig(),
			},
			expectedPlan: es.Elasticsearch{
				HotTier: &es.ElasticsearchTopology{
					Size:      ec.String("1g"),
					ZoneCount: 3,
				},
				WarmTier: &es.ElasticsearchTopology{
					Size:      ec.String("1g"),
					ZoneCount: 3,
				},
				MasterTier: &es.ElasticsearchTopology{
					Size:      ec.String("1g"),
					ZoneCount: 2,
				},
			},
		},
		{
			name: "Should remove master tier for 5 nodes",
			args: args{
				plan: es.Elasticsearch{
					HotTier: &es.ElasticsearchTopology{
						Size:      ec.String("1g"),
						ZoneCount: 3,
					},
					WarmTier: &es.ElasticsearchTopology{
						Size:      ec.String("1g"),
						ZoneCount: 2,
					},
					MasterTier: &es.ElasticsearchTopology{
						Size:      ec.String("4g"),
						ZoneCount: 3,
					},
				},
				config:               es.Elasticsearch{},
				deploymentTemplate:   deploymentTemplate(),
				masterInstanceConfig: masterInstanceConfig(),
			},
			expectedPlan: es.Elasticsearch{
				HotTier: &es.ElasticsearchTopology{
					Size:      ec.String("1g"),
					ZoneCount: 3,
				},
				WarmTier: &es.ElasticsearchTopology{
					Size:      ec.String("1g"),
					ZoneCount: 2,
				},
				MasterTier: &es.ElasticsearchTopology{
					Size:      ec.String("0g"),
					ZoneCount: 3,
				},
			},
		},
		{
			name: "Should ignore ML nodes for dedicated master count",
			args: args{
				plan: es.Elasticsearch{
					HotTier: &es.ElasticsearchTopology{
						Size:      ec.String("1g"),
						ZoneCount: 3,
					},
					WarmTier: &es.ElasticsearchTopology{
						Size:      ec.String("1g"),
						ZoneCount: 2,
					},
					MasterTier: &es.ElasticsearchTopology{
						Size:      ec.String("0g"),
						ZoneCount: 0,
					},
					MlTier: &es.ElasticsearchTopology{
						Size:      ec.String("1g"),
						ZoneCount: 3,
					},
				},
				config:               es.Elasticsearch{},
				deploymentTemplate:   deploymentTemplate(),
				masterInstanceConfig: masterInstanceConfig(),
			},
			expectedPlan: es.Elasticsearch{
				HotTier: &es.ElasticsearchTopology{
					Size:      ec.String("1g"),
					ZoneCount: 3,
				},
				WarmTier: &es.ElasticsearchTopology{
					Size:      ec.String("1g"),
					ZoneCount: 2,
				},
				MasterTier: &es.ElasticsearchTopology{
					Size:      ec.String("0g"),
					ZoneCount: 0,
				},
				MlTier: &es.ElasticsearchTopology{
					Size:      ec.String("1g"),
					ZoneCount: 3,
				},
			},
		},
		{
			name: "Should count multiple nodes for a tier in one zone (when size is > max size)",
			args: args{
				plan: es.Elasticsearch{
					HotTier: &es.ElasticsearchTopology{
						Size:      ec.String("8g"), // Max in template is 4g
						ZoneCount: 3,
					},
				},
				config:               es.Elasticsearch{},
				deploymentTemplate:   deploymentTemplate(),
				masterInstanceConfig: masterInstanceConfig(),
			},
			expectedPlan: es.Elasticsearch{
				HotTier: &es.ElasticsearchTopology{
					Size:      ec.String("8g"),
					ZoneCount: 3,
				},
				MasterTier: &es.ElasticsearchTopology{
					Size:      ec.String("1g"),
					ZoneCount: 2,
				},
			},
		},
		{
			name: "Should not override configured values",
			args: args{
				plan: es.Elasticsearch{
					HotTier: &es.ElasticsearchTopology{
						Size:      ec.String("1g"),
						ZoneCount: 3,
					},
					WarmTier: &es.ElasticsearchTopology{
						Size:      ec.String("1g"),
						ZoneCount: 3,
					},
					MasterTier: &es.ElasticsearchTopology{
						Size:      ec.String("2g"),
						ZoneCount: 2,
					},
				},
				config: es.Elasticsearch{
					HotTier: &es.ElasticsearchTopology{
						Size:      ec.String("1g"),
						ZoneCount: 3,
					},
					WarmTier: &es.ElasticsearchTopology{
						Size:      ec.String("1g"),
						ZoneCount: 3,
					},
					MasterTier: &es.ElasticsearchTopology{
						Size:      ec.String("2g"),
						ZoneCount: 2,
					},
				},
				deploymentTemplate:   deploymentTemplate(),
				masterInstanceConfig: masterInstanceConfig(),
			},
			expectedPlan: es.Elasticsearch{
				HotTier: &es.ElasticsearchTopology{
					Size:      ec.String("1g"),
					ZoneCount: 3,
				},
				WarmTier: &es.ElasticsearchTopology{
					Size:      ec.String("1g"),
					ZoneCount: 3,
				},
				MasterTier: &es.ElasticsearchTopology{
					Size:      ec.String("2g"),
					ZoneCount: 2,
				},
			},
		},
		{
			name: "Should use IC from state when enabling master tier",
			args: args{
				plan: es.Elasticsearch{
					HotTier: &es.ElasticsearchTopology{
						Size:      ec.String("1g"),
						ZoneCount: 3,
					},
					WarmTier: &es.ElasticsearchTopology{
						Size:      ec.String("1g"),
						ZoneCount: 3,
					},
					MasterTier: &es.ElasticsearchTopology{
						InstanceConfigurationId:      ec.String("master-ic"),
						InstanceConfigurationVersion: ec.Int(1),
						Size:                         ec.String("0g"),
						ZoneCount:                    0,
					},
				},
				config:               es.Elasticsearch{},
				deploymentTemplate:   deploymentTemplate(),
				masterInstanceConfig: masterInstanceConfig(),
			},
			expectedPlan: es.Elasticsearch{
				HotTier: &es.ElasticsearchTopology{
					Size:      ec.String("1g"),
					ZoneCount: 3,
				},
				WarmTier: &es.ElasticsearchTopology{
					Size:      ec.String("1g"),
					ZoneCount: 3,
				},
				MasterTier: &es.ElasticsearchTopology{
					InstanceConfigurationId:      ec.String("master-ic"),
					InstanceConfigurationVersion: ec.Int(1),
					Size:                         ec.String("4g"),
					ZoneCount:                    3,
				},
			},
		},
		{
			name: "Should not change IC when master is already enabled and stays enabled",
			args: args{
				plan: es.Elasticsearch{
					HotTier: &es.ElasticsearchTopology{
						Size:      ec.String("1g"),
						ZoneCount: 3,
					},
					WarmTier: &es.ElasticsearchTopology{
						Size:      ec.String("1g"),
						ZoneCount: 3,
					},
					MasterTier: &es.ElasticsearchTopology{
						InstanceConfigurationId:      ec.String("master-ic"),
						InstanceConfigurationVersion: ec.Int(1),
						Size:                         ec.String("4g"),
						ZoneCount:                    3,
					},
				},
				config:               es.Elasticsearch{},
				deploymentTemplate:   deploymentTemplate(),
				masterInstanceConfig: masterInstanceConfig(),
			},
			expectedPlan: es.Elasticsearch{
				HotTier: &es.ElasticsearchTopology{
					Size:      ec.String("1g"),
					ZoneCount: 3,
				},
				WarmTier: &es.ElasticsearchTopology{
					Size:      ec.String("1g"),
					ZoneCount: 3,
				},
				MasterTier: &es.ElasticsearchTopology{
					InstanceConfigurationId:      ec.String("master-ic"),
					InstanceConfigurationVersion: ec.Int(1),
					Size:                         ec.String("4g"),
					ZoneCount:                    3,
				},
			},
		},
		{
			name: "Should use the latest IC when migrate_to_latest_hardware is set",
			args: args{
				migrateToLatestHw: true,
				plan: es.Elasticsearch{
					HotTier: &es.ElasticsearchTopology{
						Size:      ec.String("1g"),
						ZoneCount: 3,
					},
					WarmTier: &es.ElasticsearchTopology{
						Size:      ec.String("1g"),
						ZoneCount: 3,
					},
					MasterTier: &es.ElasticsearchTopology{
						Size:      ec.String("4g"),
						ZoneCount: 3,
					},
				},
				config:               es.Elasticsearch{},
				deploymentTemplate:   deploymentTemplate(),
				masterInstanceConfig: masterInstanceConfig(),
			},
			expectedPlan: es.Elasticsearch{
				HotTier: &es.ElasticsearchTopology{
					Size:      ec.String("1g"),
					ZoneCount: 3,
				},
				WarmTier: &es.ElasticsearchTopology{
					Size:      ec.String("1g"),
					ZoneCount: 3,
				},
				MasterTier: &es.ElasticsearchTopology{
					Size:      ec.String("1g"),
					ZoneCount: 2,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			deploymentConfig := depl.Deployment{
				Elasticsearch:           &test.args.config,
				MigrateToLatestHardware: &test.args.migrateToLatestHw,
			}
			config := tfsdk.Config{
				Raw:    testutil.TfTypesValueFromGoTypeValue(t, deploymentConfig, depl.DeploymentSchema().Type()),
				Schema: depl.DeploymentSchema(),
			}
			deploymentPlan := depl.Deployment{
				Elasticsearch:           &test.args.plan,
				MigrateToLatestHardware: &test.args.migrateToLatestHw,
			}
			plan := tfsdk.Plan{
				Raw:    testutil.TfTypesValueFromGoTypeValue(t, deploymentPlan, depl.DeploymentSchema().Type()),
				Schema: depl.DeploymentSchema(),
			}
			request := resource.ModifyPlanRequest{
				Config: config,
				Plan:   plan,
			}
			response := resource.ModifyPlanResponse{
				Plan: plan,
			}
			loadTemplate := func() (*models.DeploymentTemplateInfoV2, error) {
				return &test.args.deploymentTemplate, nil
			}
			loadInstanceConfig := func(id string, version *int64) (*models.InstanceConfiguration, error) {
				return &test.args.masterInstanceConfig, nil
			}

			UpdateDedicatedMasterTier(ctx, request, &response, loadTemplate, loadInstanceConfig)

			assert.Empty(t, response.Diagnostics)
			var actualPlan es.Elasticsearch
			diags := response.Plan.GetAttribute(ctx, path.Root("elasticsearch"), &actualPlan)
			println(diags.Errors())
			assert.Equal(t, test.expectedPlan, actualPlan)
		})
	}
}

func deploymentTemplate() models.DeploymentTemplateInfoV2 {
	return models.DeploymentTemplateInfoV2{
		DeploymentTemplate: &models.DeploymentCreateRequest{
			Resources: &models.DeploymentCreateResources{
				Elasticsearch: []*models.ElasticsearchPayload{
					{
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "hot_content",
									InstanceConfigurationID: "hot-ic",
								},
								{
									ID:                      "warm",
									InstanceConfigurationID: "warm-ic",
								},
								{
									ID:                      "master",
									InstanceConfigurationID: "master-ic",
								},
							},
						},
					},
				},
			},
		},
		InstanceConfigurations: []*models.InstanceConfigurationInfo{
			{
				ID: "hot-ic",
				DiscreteSizes: &models.DiscreteSizes{
					DefaultSize: 1024,
					Sizes:       []int32{1024, 2048, 4096},
				},
				MaxZones: 3,
			},
			{
				ID: "warm-ic",
				DiscreteSizes: &models.DiscreteSizes{
					DefaultSize: 1024,
					Sizes:       []int32{1024, 2048, 4096},
				},
				MaxZones: 3,
			},
			{
				ID:            "master-ic",
				ConfigVersion: 2,
				DiscreteSizes: &models.DiscreteSizes{
					DefaultSize: 1024,
					Sizes:       []int32{1024},
				},
				MaxZones: 2,
			},
		},
	}
}

func masterInstanceConfig() models.InstanceConfiguration {
	return models.InstanceConfiguration{
		ID: "master-ic",
		DiscreteSizes: &models.DiscreteSizes{
			DefaultSize: 4096,
			Resource:    "memory",
			Sizes:       []int32{4096},
		},
		MaxZones: 3,
	}
}
