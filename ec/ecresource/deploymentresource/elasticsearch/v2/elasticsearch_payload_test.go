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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/testutil"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/utils"
)

func Test_writeElasticsearch(t *testing.T) {
	tplPath := "../../testdata/template-aws-io-optimized-v2.json"
	tp770 := func() *models.ElasticsearchPayload {
		return utils.EnrichElasticsearchTemplate(
			utils.EsResource(testutil.ParseDeploymentTemplate(t, tplPath)),
			"aws-io-optimized-v2",
			"7.7.0",
			false,
		)
	}

	create710 := func() *models.ElasticsearchPayload {
		return utils.EnrichElasticsearchTemplate(
			utils.EsResource(testutil.ParseDeploymentTemplate(t, tplPath)),
			"aws-io-optimized-v2",
			"7.10.0",
			true,
		)
	}

	update711 := func() *models.ElasticsearchPayload {
		return utils.EnrichElasticsearchTemplate(
			utils.EsResource(testutil.ParseDeploymentTemplate(t, tplPath)),
			"aws-io-optimized-v2",
			"7.11.0",
			true,
		)
	}

	hotWarmTplPath := "../../testdata/template-aws-hot-warm-v2.json"
	hotWarmTpl770 := func() *models.ElasticsearchPayload {
		return utils.EnrichElasticsearchTemplate(
			utils.EsResource(testutil.ParseDeploymentTemplate(t, hotWarmTplPath)),
			"aws-io-optimized-v2",
			"7.7.0",
			false,
		)
	}

	hotWarm7111Tpl := func() *models.ElasticsearchPayload {
		return utils.EnrichElasticsearchTemplate(
			utils.EsResource(testutil.ParseDeploymentTemplate(t, hotWarmTplPath)),
			"aws-io-optimized-v2",
			"7.11.1",
			true,
		)
	}

	eceDefaultTplPath := "../../testdata/template-ece-3.0.0-default.json"
	eceDefaultTpl := func() *models.ElasticsearchPayload {
		return utils.EnrichElasticsearchTemplate(
			utils.EsResource(testutil.ParseDeploymentTemplate(t, eceDefaultTplPath)),
			"aws-io-optimized-v2",
			"7.17.3",
			true,
		)
	}

	type args struct {
		es           Elasticsearch
		template     *models.DeploymentTemplateInfoV2
		templateID   string
		version      string
		useNodeRoles bool
	}
	tests := []struct {
		name  string
		args  args
		want  *models.ElasticsearchPayload
		diags diag.Diagnostics
	}{
		{
			name: "parses an ES resource",
			args: args{
				es: Elasticsearch{
					RefId:      ec.String("main-elasticsearch"),
					ResourceId: ec.String(mock.ValidClusterID),
					Region:     ec.String("some-region"),
					HotTier: &ElasticsearchTopology{
						id:        "hot_content",
						Size:      ec.String("2g"),
						ZoneCount: 1,
					},
				},
				template:     testutil.ParseDeploymentTemplate(t, "../../testdata/template-aws-io-optimized-v2.json"),
				templateID:   "aws-io-optimized-v2",
				version:      "7.7.0",
				useNodeRoles: false,
			},
			want: testutil.EnrichWithEmptyTopologies(tp770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version: "7.7.0",
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-io-optimized-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID:                      "hot_content",
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highio.i3",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							NodeType: &models.ElasticsearchNodeType{
								Data:   ec.Bool(true),
								Ingest: ec.Bool(true),
								Master: ec.Bool(true),
							},
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parses an ES resource with empty version (7.10.0) in state uses node_roles from the DT",
			args: args{
				es: Elasticsearch{
					RefId:      ec.String("main-elasticsearch"),
					ResourceId: ec.String(mock.ValidClusterID),
					Region:     ec.String("some-region"),
					HotTier: &ElasticsearchTopology{
						id:        "hot_content",
						Size:      ec.String("2g"),
						ZoneCount: 1,
					},
				},
				template:     testutil.ParseDeploymentTemplate(t, "../../testdata/template-aws-io-optimized-v2.json"),
				templateID:   "aws-io-optimized-v2",
				version:      "7.10.0",
				useNodeRoles: true,
			},
			want: testutil.EnrichWithEmptyTopologies(create710(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version: "7.10.0",
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-io-optimized-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID:                      "hot_content",
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highio.i3",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							NodeRoles: []string{
								"master",
								"ingest",
								"remote_cluster_client",
								"data_hot",
								"transform",
								"data_content",
							},
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parses an ES resource with version 7.11.0 has node_roles coming from the saved state",
			args: args{
				es: Elasticsearch{
					RefId:      ec.String("main-elasticsearch"),
					ResourceId: ec.String(mock.ValidClusterID),
					Region:     ec.String("some-region"),
					HotTier: &ElasticsearchTopology{
						id:        "hot_content",
						Size:      ec.String("2g"),
						ZoneCount: 1,
						NodeRoles: []string{"a", "b", "c"},
					},
				},
				template:     testutil.ParseDeploymentTemplate(t, "../../testdata/template-aws-io-optimized-v2.json"),
				templateID:   "aws-io-optimized-v2",
				version:      "7.11.0",
				useNodeRoles: true,
			},
			want: testutil.EnrichWithEmptyTopologies(update711(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version: "7.11.0",
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-io-optimized-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID:                      "hot_content",
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highio.i3",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							NodeRoles: []string{
								"a", "b", "c",
							},
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parses an ES resource without a topology",
			args: args{
				es: Elasticsearch{
					RefId:      ec.String("main-elasticsearch"),
					ResourceId: ec.String(mock.ValidClusterID),
					Region:     ec.String("some-region"),
				},
				template:     testutil.ParseDeploymentTemplate(t, "../../testdata/template-aws-io-optimized-v2.json"),
				templateID:   "aws-io-optimized-v2",
				version:      "7.7.0",
				useNodeRoles: false,
			},
			want: testutil.EnrichWithEmptyTopologies(tp770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version: "7.7.0",
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-io-optimized-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID:                      "hot_content",
							InstanceConfigurationID: "aws.data.highio.i3",
							ZoneCount:               2,
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(8192),
							},
							NodeType: &models.ElasticsearchNodeType{
								Data:   ec.Bool(true),
								Ingest: ec.Bool(true),
								Master: ec.Bool(true),
							},
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parses an ES resource (HotWarm)",
			args: args{
				es: Elasticsearch{
					RefId:      ec.String("main-elasticsearch"),
					ResourceId: ec.String(mock.ValidClusterID),
					Region:     ec.String("some-region"),
					HotTier: &ElasticsearchTopology{
						id:        "hot_content",
						Size:      ec.String("2g"),
						ZoneCount: 1,
					},
					WarmTier: &ElasticsearchTopology{
						id:        "warm",
						Size:      ec.String("2g"),
						ZoneCount: 1,
					},
				},
				template:     testutil.ParseDeploymentTemplate(t, "../../testdata/template-aws-hot-warm-v2.json"),
				templateID:   "aws-hot-warm-v2",
				version:      "7.7.0",
				useNodeRoles: false,
			},
			want: testutil.EnrichWithEmptyTopologies(hotWarmTpl770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
					Curation:                  nil,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version:  "7.7.0",
						Curation: nil,
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-hot-warm-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID: "hot_content",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highio.i3",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							NodeType: &models.ElasticsearchNodeType{
								Data:   ec.Bool(true),
								Ingest: ec.Bool(true),
								Master: ec.Bool(true),
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "warm",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "warm",
								},
							},
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highstorage.d2",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							NodeType: &models.ElasticsearchNodeType{
								Data:   ec.Bool(true),
								Ingest: ec.Bool(true),
								Master: ec.Bool(false),
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parses an ES resource with config (HotWarm)",
			args: args{
				es: Elasticsearch{
					RefId:      ec.String("main-elasticsearch"),
					ResourceId: ec.String(mock.ValidClusterID),
					Region:     ec.String("some-region"),
					Config: &ElasticsearchConfig{
						UserSettingsYaml: ec.String("somesetting: true"),
					},
					HotTier: &ElasticsearchTopology{
						id:        "hot_content",
						Size:      ec.String("2g"),
						ZoneCount: 1,
					},
					WarmTier: &ElasticsearchTopology{
						id:        "warm",
						Size:      ec.String("2g"),
						ZoneCount: 1,
					},
				},
				template:     testutil.ParseDeploymentTemplate(t, "../../testdata/template-aws-hot-warm-v2.json"),
				templateID:   "aws-hot-warm-v2",
				version:      "7.7.0",
				useNodeRoles: false,
			},
			want: testutil.EnrichWithEmptyTopologies(hotWarmTpl770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
					Curation:                  nil,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version:          "7.7.0",
						Curation:         nil,
						UserSettingsYaml: "somesetting: true",
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-hot-warm-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID: "hot_content",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highio.i3",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							NodeType: &models.ElasticsearchNodeType{
								Data:   ec.Bool(true),
								Ingest: ec.Bool(true),
								Master: ec.Bool(true),
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "warm",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "warm",
								},
							},
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highstorage.d2",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							NodeType: &models.ElasticsearchNodeType{
								Data:   ec.Bool(true),
								Ingest: ec.Bool(true),
								Master: ec.Bool(false),
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parses an ES resource without a topology (HotWarm)",
			args: args{
				es: Elasticsearch{
					RefId:      ec.String("main-elasticsearch"),
					ResourceId: ec.String(mock.ValidClusterID),
					Region:     ec.String("some-region"),
				},
				template:     testutil.ParseDeploymentTemplate(t, "../../testdata/template-aws-hot-warm-v2.json"),
				templateID:   "aws-hot-warm-v2",
				version:      "7.7.0",
				useNodeRoles: false,
			},
			want: testutil.EnrichWithEmptyTopologies(hotWarmTpl770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
					Curation:                  nil,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version:  "7.7.0",
						Curation: nil,
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-hot-warm-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID: "hot_content",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							ZoneCount:               2,
							InstanceConfigurationID: "aws.data.highio.i3",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(4096),
							},
							NodeType: &models.ElasticsearchNodeType{
								Data:   ec.Bool(true),
								Ingest: ec.Bool(true),
								Master: ec.Bool(true),
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "warm",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "warm",
								},
							},
							ZoneCount:               2,
							InstanceConfigurationID: "aws.data.highstorage.d2",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(4096),
							},
							NodeType: &models.ElasticsearchNodeType{
								Data:   ec.Bool(true),
								Ingest: ec.Bool(true),
								Master: ec.Bool(false),
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parses an ES resource with node type overrides (HotWarm)",
			args: args{
				es: Elasticsearch{
					RefId:      ec.String("main-elasticsearch"),
					ResourceId: ec.String(mock.ValidClusterID),
					Region:     ec.String("some-region"),
					HotTier: &ElasticsearchTopology{
						id:             "hot_content",
						NodeTypeData:   ec.String("false"),
						NodeTypeMaster: ec.String("false"),
						NodeTypeIngest: ec.String("false"),
						NodeTypeMl:     ec.String("true"),
					},
					WarmTier: &ElasticsearchTopology{
						id:             "warm",
						NodeTypeMaster: ec.String("true"),
					},
				},
				template:     testutil.ParseDeploymentTemplate(t, "../../testdata/template-aws-hot-warm-v2.json"),
				templateID:   "aws-hot-warm-v2",
				version:      "7.7.0",
				useNodeRoles: false,
			},
			want: testutil.EnrichWithEmptyTopologies(hotWarmTpl770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
					Curation:                  nil,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version:  "7.7.0",
						Curation: nil,
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-hot-warm-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID: "hot_content",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							ZoneCount:               2,
							InstanceConfigurationID: "aws.data.highio.i3",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(4096),
							},
							NodeType: &models.ElasticsearchNodeType{
								Data:   ec.Bool(false),
								Ingest: ec.Bool(false),
								Master: ec.Bool(false),
								Ml:     ec.Bool(true),
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "warm",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "warm",
								},
							},
							ZoneCount:               2,
							InstanceConfigurationID: "aws.data.highstorage.d2",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(4096),
							},
							NodeType: &models.ElasticsearchNodeType{
								Data:   ec.Bool(true),
								Ingest: ec.Bool(true),
								Master: ec.Bool(true),
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "migrates old node_type state to new node_roles payload when the cold tier is set",
			args: args{
				es: Elasticsearch{
					RefId:      ec.String("main-elasticsearch"),
					ResourceId: ec.String(mock.ValidClusterID),
					Region:     ec.String("some-region"),
					HotTier: &ElasticsearchTopology{
						id:             "hot_content",
						NodeTypeData:   ec.String("false"),
						NodeTypeMaster: ec.String("false"),
						NodeTypeIngest: ec.String("false"),
						NodeTypeMl:     ec.String("true"),
					},
					WarmTier: &ElasticsearchTopology{
						id:             "warm",
						NodeTypeMaster: ec.String("true"),
					},
					ColdTier: &ElasticsearchTopology{
						id:   "cold",
						Size: ec.String("2g"),
					},
				},
				template:     testutil.ParseDeploymentTemplate(t, "../../testdata/template-aws-hot-warm-v2.json"),
				templateID:   "aws-io-optimized-v2",
				version:      "7.11.1",
				useNodeRoles: true,
			},
			want: testutil.EnrichWithEmptyTopologies(hotWarm7111Tpl(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
					Curation:                  nil,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version:  "7.11.1",
						Curation: nil,
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-hot-warm-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID: "hot_content",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							ZoneCount:               2,
							InstanceConfigurationID: "aws.data.highio.i3",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(4096),
							},
							NodeRoles: []string{
								"master",
								"ingest",
								"remote_cluster_client",
								"data_hot",
								"transform",
								"data_content",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "warm",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "warm",
								},
							},
							ZoneCount:               2,
							InstanceConfigurationID: "aws.data.highstorage.d2",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(4096),
							},
							NodeRoles: []string{
								"data_warm",
								"remote_cluster_client",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "cold",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "cold",
								},
							},
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highstorage.d2",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							NodeRoles: []string{
								"data_cold",
								"remote_cluster_client",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(59392),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "autoscaling enabled",
			args: args{
				es: Elasticsearch{
					Autoscale:  ec.String("true"),
					RefId:      ec.String("main-elasticsearch"),
					ResourceId: ec.String(mock.ValidClusterID),
					Region:     ec.String("some-region"),
					HotTier: &ElasticsearchTopology{
						id: "hot_content",
					},
					WarmTier: &ElasticsearchTopology{
						id: "warm",
					},
					ColdTier: &ElasticsearchTopology{
						id:   "cold",
						Size: ec.String("2g"),
					},
				},
				template:     testutil.ParseDeploymentTemplate(t, "../../testdata/template-aws-hot-warm-v2.json"),
				templateID:   "aws-io-optimized-v2",
				version:      "7.11.1",
				useNodeRoles: true,
			},
			want: testutil.EnrichWithEmptyTopologies(hotWarm7111Tpl(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
					Curation:                  nil,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(true),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version:  "7.11.1",
						Curation: nil,
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-hot-warm-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID: "hot_content",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							ZoneCount:               2,
							InstanceConfigurationID: "aws.data.highio.i3",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(4096),
							},
							NodeRoles: []string{
								"master",
								"ingest",
								"remote_cluster_client",
								"data_hot",
								"transform",
								"data_content",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "warm",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "warm",
								},
							},
							ZoneCount:               2,
							InstanceConfigurationID: "aws.data.highstorage.d2",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(4096),
							},
							NodeRoles: []string{
								"data_warm",
								"remote_cluster_client",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "cold",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "cold",
								},
							},
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highstorage.d2",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							NodeRoles: []string{
								"data_cold",
								"remote_cluster_client",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(59392),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "autoscaling enabled overriding the size with ml",
			args: args{
				es: Elasticsearch{
					Autoscale:  ec.String("true"),
					RefId:      ec.String("main-elasticsearch"),
					ResourceId: ec.String(mock.ValidClusterID),
					Region:     ec.String("some-region"),
					HotTier: &ElasticsearchTopology{
						id: "hot_content",
						Autoscaling: &ElasticsearchTopologyAutoscaling{
							MaxSize: ec.String("58g"),
						},
					},
					WarmTier: &ElasticsearchTopology{
						id: "warm",
						Autoscaling: &ElasticsearchTopologyAutoscaling{
							MaxSize: ec.String("29g"),
						},
					},
					ColdTier: &ElasticsearchTopology{
						id:   "cold",
						Size: ec.String("2g"),
						Autoscaling: &ElasticsearchTopologyAutoscaling{
							MaxSize: ec.String("29g"),
						},
					},
					MlTier: &ElasticsearchTopology{
						id:   "ml",
						Size: ec.String("1g"),
						Autoscaling: &ElasticsearchTopologyAutoscaling{
							MaxSize: ec.String("29g"),
							MinSize: ec.String("1g"),
						},
					},
				},
				template:     testutil.ParseDeploymentTemplate(t, "../../testdata/template-aws-hot-warm-v2.json"),
				templateID:   "aws-io-optimized-v2",
				version:      "7.11.1",
				useNodeRoles: true,
			},
			want: testutil.EnrichWithEmptyTopologies(hotWarm7111Tpl(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
					Curation:                  nil,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(true),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version:  "7.11.1",
						Curation: nil,
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-hot-warm-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID: "hot_content",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							ZoneCount:               2,
							InstanceConfigurationID: "aws.data.highio.i3",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(4096),
							},
							NodeRoles: []string{
								"master",
								"ingest",
								"remote_cluster_client",
								"data_hot",
								"transform",
								"data_content",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(59392),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "warm",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "warm",
								},
							},
							ZoneCount:               2,
							InstanceConfigurationID: "aws.data.highstorage.d2",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(4096),
							},
							NodeRoles: []string{
								"data_warm",
								"remote_cluster_client",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(29696),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "cold",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "cold",
								},
							},
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highstorage.d2",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							NodeRoles: []string{
								"data_cold",
								"remote_cluster_client",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(29696),
								Resource: ec.String("memory"),
							},
						},
						{
							ID:                      "ml",
							ZoneCount:               1,
							InstanceConfigurationID: "aws.ml.m5d",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(1024),
							},
							NodeRoles: []string{
								"ml",
								"remote_cluster_client",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(29696),
								Resource: ec.String("memory"),
							},
							AutoscalingMin: &models.TopologySize{
								Value:    ec.Int32(1024),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "autoscaling enabled no dimension in template, default resource",
			args: args{
				es: Elasticsearch{
					Autoscale:  ec.String("true"),
					RefId:      ec.String("main-elasticsearch"),
					ResourceId: ec.String(mock.ValidClusterID),
					Region:     ec.String("some-region"),
					HotTier: &ElasticsearchTopology{
						id: "hot_content",
						Autoscaling: &ElasticsearchTopologyAutoscaling{
							MaxSize: ec.String("450g"),
							MinSize: ec.String("2g"),
						},
					},
					MasterTier: &ElasticsearchTopology{
						id: "master",
						Autoscaling: &ElasticsearchTopologyAutoscaling{
							MaxSize: ec.String("250g"),
							MinSize: ec.String("1g"),
						},
					},
				},
				template:     testutil.ParseDeploymentTemplate(t, "../../testdata/template-ece-3.0.0-default.json"),
				templateID:   "aws-io-optimized-v2",
				version:      "7.17.3",
				useNodeRoles: true,
			},
			want: testutil.EnrichWithEmptyTopologies(eceDefaultTpl(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
					Curation:                  nil,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(true),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version:  "7.17.3",
						Curation: nil,
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-io-optimized-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID: "hot_content",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							ZoneCount:               1,
							InstanceConfigurationID: "data.default",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(4096),
							},
							NodeRoles: []string{
								"master",
								"ingest",
								"data_hot",
								"data_content",
								"remote_cluster_client",
								"transform",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(460800),
								Resource: ec.String("memory"),
							},
							AutoscalingMin: &models.TopologySize{
								Value:    ec.Int32(2048),
								Resource: ec.String("memory"),
							},
						},
						{
							ID:                      "master",
							ZoneCount:               1,
							InstanceConfigurationID: "master",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(0),
							},
							NodeRoles: []string{
								"master",
								"remote_cluster_client",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(256000),
								Resource: ec.String("memory"),
							},
							AutoscalingMin: &models.TopologySize{
								Value:    ec.Int32(1024),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "autoscaling enabled overriding the size and resources",
			args: args{
				es: Elasticsearch{
					Autoscale:  ec.String("true"),
					RefId:      ec.String("main-elasticsearch"),
					ResourceId: ec.String(mock.ValidClusterID),
					Region:     ec.String("some-region"),
					HotTier: &ElasticsearchTopology{
						id: "hot_content",
						Autoscaling: &ElasticsearchTopologyAutoscaling{
							MaxSize:         ec.String("450g"),
							MaxSizeResource: ec.String("storage"),
						},
					},
					WarmTier: &ElasticsearchTopology{
						id: "warm",
						Autoscaling: &ElasticsearchTopologyAutoscaling{
							MaxSize:         ec.String("870g"),
							MaxSizeResource: ec.String("storage"),
						},
					},
					ColdTier: &ElasticsearchTopology{
						id:   "cold",
						Size: ec.String("4g"),
						Autoscaling: &ElasticsearchTopologyAutoscaling{
							MaxSize:         ec.String("1740g"),
							MaxSizeResource: ec.String("storage"),
							MinSizeResource: ec.String("storage"),
							MinSize:         ec.String("4g"),
						},
					},
				},
				template:     testutil.ParseDeploymentTemplate(t, "../../testdata/template-aws-hot-warm-v2.json"),
				templateID:   "aws-io-optimized-v2",
				version:      "7.11.1",
				useNodeRoles: true,
			},
			want: testutil.EnrichWithEmptyTopologies(hotWarm7111Tpl(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
					Curation:                  nil,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(true),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version:  "7.11.1",
						Curation: nil,
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-hot-warm-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID: "hot_content",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							ZoneCount:               2,
							InstanceConfigurationID: "aws.data.highio.i3",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(4096),
							},
							NodeRoles: []string{
								"master",
								"ingest",
								"remote_cluster_client",
								"data_hot",
								"transform",
								"data_content",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(460800),
								Resource: ec.String("storage"),
							},
						},
						{
							ID: "warm",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "warm",
								},
							},
							ZoneCount:               2,
							InstanceConfigurationID: "aws.data.highstorage.d2",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(4096),
							},
							NodeRoles: []string{
								"data_warm",
								"remote_cluster_client",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(890880),
								Resource: ec.String("storage"),
							},
						},
						{
							ID: "cold",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "cold",
								},
							},
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highstorage.d2",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(4096),
							},
							NodeRoles: []string{
								"data_cold",
								"remote_cluster_client",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(1781760),
								Resource: ec.String("storage"),
							},
							AutoscalingMin: &models.TopologySize{
								Value:    ec.Int32(4096),
								Resource: ec.String("storage"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parses an ES resource with plugins",
			args: args{
				es: Elasticsearch{
					RefId:      ec.String("main-elasticsearch"),
					ResourceId: ec.String(mock.ValidClusterID),
					Region:     ec.String("some-region"),
					Config: &ElasticsearchConfig{
						UserSettingsYaml:         ec.String("some.setting: value"),
						UserSettingsOverrideYaml: ec.String("some.setting: value2"),
						UserSettingsJson:         ec.String("{\"some.setting\":\"value\"}"),
						UserSettingsOverrideJson: ec.String("{\"some.setting\":\"value2\"}"),
						Plugins:                  []string{"plugin"},
					},
					HotTier: &ElasticsearchTopology{
						id:        "hot_content",
						Size:      ec.String("2g"),
						ZoneCount: 1,
					},
				},
				template:     testutil.ParseDeploymentTemplate(t, "../../testdata/template-aws-io-optimized-v2.json"),
				templateID:   "aws-io-optimized-v2",
				version:      "7.7.0",
				useNodeRoles: false,
			},
			want: testutil.EnrichWithEmptyTopologies(tp770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version:                  "7.7.0",
						UserSettingsYaml:         `some.setting: value`,
						UserSettingsOverrideYaml: `some.setting: value2`,
						UserSettingsJSON: map[string]interface{}{
							"some.setting": "value",
						},
						UserSettingsOverrideJSON: map[string]interface{}{
							"some.setting": "value2",
						},
						EnabledBuiltInPlugins: []string{"plugin"},
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-io-optimized-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID:                      "hot_content",
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highio.i3",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{"data": "hot"},
							},
							NodeType: &models.ElasticsearchNodeType{
								Data:   ec.Bool(true),
								Ingest: ec.Bool(true),
								Master: ec.Bool(true),
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parses an ES resource with snapshot settings",
			args: args{
				es: Elasticsearch{
					RefId:      ec.String("main-elasticsearch"),
					ResourceId: ec.String(mock.ValidClusterID),
					Region:     ec.String("some-region"),
					SnapshotSource: &ElasticsearchSnapshotSource{
						SnapshotName:                 "__latest_success__",
						SourceElasticsearchClusterId: mock.ValidClusterID,
					},
					HotTier: &ElasticsearchTopology{
						id:        "hot_content",
						Size:      ec.String("2g"),
						ZoneCount: 1,
					},
				},
				template:     testutil.ParseDeploymentTemplate(t, "../../testdata/template-aws-io-optimized-v2.json"),
				templateID:   "aws-io-optimized-v2",
				version:      "7.7.0",
				useNodeRoles: false,
			},
			want: testutil.EnrichWithEmptyTopologies(tp770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version: "7.7.0",
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-io-optimized-v2"),
					},
					Transient: &models.TransientElasticsearchPlanConfiguration{
						RestoreSnapshot: &models.RestoreSnapshotConfiguration{
							SnapshotName:    ec.String("__latest_success__"),
							SourceClusterID: mock.ValidClusterID,
						},
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID:                      "hot_content",
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highio.i3",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							NodeType: &models.ElasticsearchNodeType{
								Data:   ec.Bool(true),
								Ingest: ec.Bool(true),
								Master: ec.Bool(true),
							},
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{"data": "hot"},
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parse autodetect configuration strategy",
			args: args{
				es: Elasticsearch{
					RefId:      ec.String("main-elasticsearch"),
					ResourceId: ec.String(mock.ValidClusterID),
					Region:     ec.String("some-region"),
					HotTier: &ElasticsearchTopology{
						id:        "hot_content",
						Size:      ec.String("2g"),
						ZoneCount: 1,
					},
					Strategy: ec.String("autodetect"),
				},
				template:     testutil.ParseDeploymentTemplate(t, "../../testdata/template-aws-io-optimized-v2.json"),
				templateID:   "aws-io-optimized-v2",
				version:      "7.7.0",
				useNodeRoles: false,
			},
			want: testutil.EnrichWithEmptyTopologies(tp770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version: "7.7.0",
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-io-optimized-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID:                      "hot_content",
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highio.i3",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							NodeType: &models.ElasticsearchNodeType{
								Data:   ec.Bool(true),
								Ingest: ec.Bool(true),
								Master: ec.Bool(true),
							},
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
					Transient: &models.TransientElasticsearchPlanConfiguration{
						Strategy: &models.PlanStrategy{
							Autodetect: new(models.AutodetectStrategyConfig),
						},
					},
				},
			}),
		},
		{
			name: "parse grow_and_shrink configuration strategy",
			args: args{
				es: Elasticsearch{
					RefId:      ec.String("main-elasticsearch"),
					ResourceId: ec.String(mock.ValidClusterID),
					Region:     ec.String("some-region"),
					HotTier: &ElasticsearchTopology{
						id:        "hot_content",
						Size:      ec.String("2g"),
						ZoneCount: 1,
					},
					Strategy: ec.String("grow_and_shrink"),
				},
				template:     testutil.ParseDeploymentTemplate(t, "../../testdata/template-aws-io-optimized-v2.json"),
				templateID:   "aws-io-optimized-v2",
				version:      "7.7.0",
				useNodeRoles: false,
			},
			want: testutil.EnrichWithEmptyTopologies(tp770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version: "7.7.0",
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-io-optimized-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID:                      "hot_content",
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highio.i3",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							NodeType: &models.ElasticsearchNodeType{
								Data:   ec.Bool(true),
								Ingest: ec.Bool(true),
								Master: ec.Bool(true),
							},
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
					Transient: &models.TransientElasticsearchPlanConfiguration{
						Strategy: &models.PlanStrategy{
							GrowAndShrink: new(models.GrowShrinkStrategyConfig),
						},
					},
				},
			}),
		},
		{
			name: "parse rolling_grow_and_shrink configuration strategy",
			args: args{
				es: Elasticsearch{
					RefId:      ec.String("main-elasticsearch"),
					ResourceId: ec.String(mock.ValidClusterID),
					Region:     ec.String("some-region"),
					HotTier: &ElasticsearchTopology{
						id:        "hot_content",
						Size:      ec.String("2g"),
						ZoneCount: 1,
					},
					Strategy: ec.String("rolling_grow_and_shrink"),
				},
				template:     testutil.ParseDeploymentTemplate(t, "../../testdata/template-aws-io-optimized-v2.json"),
				templateID:   "aws-io-optimized-v2",
				version:      "7.7.0",
				useNodeRoles: false,
			},
			want: testutil.EnrichWithEmptyTopologies(tp770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version: "7.7.0",
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-io-optimized-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID:                      "hot_content",
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highio.i3",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							NodeType: &models.ElasticsearchNodeType{
								Data:   ec.Bool(true),
								Ingest: ec.Bool(true),
								Master: ec.Bool(true),
							},
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
					Transient: &models.TransientElasticsearchPlanConfiguration{
						Strategy: &models.PlanStrategy{
							RollingGrowAndShrink: new(models.RollingGrowShrinkStrategyConfig),
						},
					},
				},
			}),
		},
		{
			name: "parse rolling configuration strategy",
			args: args{
				es: Elasticsearch{
					RefId:      ec.String("main-elasticsearch"),
					ResourceId: ec.String(mock.ValidClusterID),
					Region:     ec.String("some-region"),
					HotTier: &ElasticsearchTopology{
						id:        "hot_content",
						Size:      ec.String("2g"),
						ZoneCount: 1,
					},
					Strategy: ec.String("rolling_all"),
				},
				template:     testutil.ParseDeploymentTemplate(t, "../../testdata/template-aws-io-optimized-v2.json"),
				templateID:   "aws-io-optimized-v2",
				version:      "7.7.0",
				useNodeRoles: false,
			},

			want: testutil.EnrichWithEmptyTopologies(tp770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version: "7.7.0",
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-io-optimized-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID:                      "hot_content",
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highio.i3",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							NodeType: &models.ElasticsearchNodeType{
								Data:   ec.Bool(true),
								Ingest: ec.Bool(true),
								Master: ec.Bool(true),
							},
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
					Transient: &models.TransientElasticsearchPlanConfiguration{
						Strategy: &models.PlanStrategy{
							Rolling: &models.RollingStrategyConfig{
								GroupBy: "__all__",
							},
						},
					},
				},
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var elasticsearch types.Object
			diags := tfsdk.ValueFrom(context.Background(), tt.args.es, ElasticsearchSchema().FrameworkType(), &elasticsearch)
			assert.Nil(t, diags)

			got, diags := ElasticsearchPayload(context.Background(), elasticsearch, tt.args.template, tt.args.templateID, tt.args.version, tt.args.useNodeRoles, false)
			if tt.diags != nil {
				assert.Equal(t, tt.diags, diags)
			} else {
				assert.Nil(t, diags)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
