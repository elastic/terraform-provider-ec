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

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
)

func Test_readElasticsearch(t *testing.T) {
	type args struct {
		in      []*models.ElasticsearchResourceInfo
		remotes models.RemoteResources
	}
	tests := []struct {
		name string
		args args
		want *Elasticsearch
	}{
		{
			name: "empty resource list returns empty list",
			args: args{in: []*models.ElasticsearchResourceInfo{}},
			want: nil,
		},
		{
			name: "empty current plan returns empty list",
			args: args{in: []*models.ElasticsearchResourceInfo{
				{
					Info: &models.ElasticsearchClusterInfo{
						PlanInfo: &models.ElasticsearchClusterPlansInfo{
							Pending: &models.ElasticsearchClusterPlanInfo{},
						},
					},
				},
			}},
			want: nil,
		},
		{
			name: "parses an elasticsearch resource",
			args: args{in: []*models.ElasticsearchResourceInfo{
				{
					Region: new("some-region"),
					RefID:  new("main-elasticsearch"),
					Info: &models.ElasticsearchClusterInfo{
						ClusterID: &mock.ValidClusterID,
						Region:    "some-region",
						Status:    new("started"),
						Metadata: &models.ClusterMetadataInfo{
							CloudID:  "some CLOUD ID",
							Endpoint: "somecluster.cloud.elastic.co",
							Ports: &models.ClusterMetadataPortInfo{
								HTTP:  ec.Int32(9200),
								HTTPS: ec.Int32(9243),
							},
						},
						PlanInfo: &models.ElasticsearchClusterPlansInfo{
							Current: &models.ElasticsearchClusterPlanInfo{
								Plan: &models.ElasticsearchClusterPlan{
									Elasticsearch: &models.ElasticsearchConfiguration{
										Version: "7.7.0",
									},
									ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
										{
											ID:                           "hot_content",
											ZoneCount:                    1,
											InstanceConfigurationID:      "aws.data.highio.i3",
											InstanceConfigurationVersion: ec.Int32(1),
											Size: &models.TopologySize{
												Resource: new("memory"),
												Value:    ec.Int32(2048),
											},
											NodeType: &models.ElasticsearchNodeType{
												Data:   new(true),
												Ingest: new(true),
												Master: new(true),
												Ml:     new(false),
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
							},
						},
					},
				},
				{
					Region: new("some-region"),
					RefID:  new("main-elasticsearch"),
					Info: &models.ElasticsearchClusterInfo{
						ClusterID: &mock.ValidClusterID,
						Region:    "some-region",
						Status:    new("stopped"),
						Metadata: &models.ClusterMetadataInfo{
							CloudID:  "some CLOUD ID",
							Endpoint: "somecluster.cloud.elastic.co",
							Ports: &models.ClusterMetadataPortInfo{
								HTTP:  ec.Int32(9200),
								HTTPS: ec.Int32(9243),
							},
						},
						PlanInfo: &models.ElasticsearchClusterPlansInfo{
							Current: &models.ElasticsearchClusterPlanInfo{
								Plan: &models.ElasticsearchClusterPlan{
									Elasticsearch: &models.ElasticsearchConfiguration{
										Version: "7.7.0",
									},
									ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
										{
											ID:                           "hot_content",
											ZoneCount:                    1,
											InstanceConfigurationID:      "aws.data.highio.i3",
											InstanceConfigurationVersion: ec.Int32(1),
											Size: &models.TopologySize{
												Resource: new("memory"),
												Value:    ec.Int32(2048),
											},
											NodeType: &models.ElasticsearchNodeType{
												Data:   new(true),
												Ingest: new(true),
												Master: new(true),
												Ml:     new(false),
											},
										},
									},
								},
							},
						},
					},
				},
			}},
			want: &Elasticsearch{
				RefId:         new("main-elasticsearch"),
				ResourceId:    new(mock.ValidClusterID),
				Region:        new("some-region"),
				CloudID:       new("some CLOUD ID"),
				HttpEndpoint:  new("http://somecluster.cloud.elastic.co:9200"),
				HttpsEndpoint: new("https://somecluster.cloud.elastic.co:9243"),
				Config: &ElasticsearchConfig{
					Plugins: []string{},
				},
				HotTier: &ElasticsearchTopology{
					id:                           "hot_content",
					InstanceConfigurationId:      new("aws.data.highio.i3"),
					InstanceConfigurationVersion: new(1),
					Size:                         new("2g"),
					SizeResource:                 new("memory"),
					NodeTypeData:                 new("true"),
					NodeTypeIngest:               new("true"),
					NodeTypeMaster:               new("true"),
					NodeTypeMl:                   new("false"),
					ZoneCount:                    1,
					Autoscaling:                  &ElasticsearchTopologyAutoscaling{},
				},
				TrustAccount:  ElasticsearchTrustAccounts{},
				TrustExternal: ElasticsearchTrustExternals{},
			},
		},
		{
			name: "resource with a config object",
			args: args{in: []*models.ElasticsearchResourceInfo{
				{
					Region: new("some-region"),
					RefID:  new("main-elasticsearch"),
					Info: &models.ElasticsearchClusterInfo{
						ClusterID:   &mock.ValidClusterID,
						ClusterName: new("some-name"),
						Region:      "some-region",
						Status:      new("started"),
						Metadata: &models.ClusterMetadataInfo{
							Endpoint: "othercluster.cloud.elastic.co",
							Ports: &models.ClusterMetadataPortInfo{
								HTTP:  ec.Int32(9200),
								HTTPS: ec.Int32(9243),
							},
						},
						PlanInfo: &models.ElasticsearchClusterPlansInfo{
							Current: &models.ElasticsearchClusterPlanInfo{
								Plan: &models.ElasticsearchClusterPlan{
									Elasticsearch: &models.ElasticsearchConfiguration{
										Version:                  "7.7.0",
										UserSettingsYaml:         `some.setting: value`,
										UserSettingsOverrideYaml: `some.setting: value2`,
										UserSettingsJSON: map[string]any{
											"some.setting": "value",
										},
										UserSettingsOverrideJSON: map[string]any{
											"some.setting": "value2",
										},
									},
									ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
										ID:                      "hot_content",
										ZoneCount:               1,
										InstanceConfigurationID: "aws.data.highio.i3",
										Size: &models.TopologySize{
											Resource: new("memory"),
											Value:    ec.Int32(2048),
										},
										NodeType: &models.ElasticsearchNodeType{
											Data:   new(true),
											Ingest: new(true),
											Master: new(true),
											Ml:     new(false),
										},
									}},
									Transient: &models.TransientElasticsearchPlanConfiguration{
										Strategy: &models.PlanStrategy{
											GrowAndShrink: new(models.GrowShrinkStrategyConfig),
										},
									},
								},
							},
						},
					},
				},
			}},
			want: &Elasticsearch{
				RefId:         new("main-elasticsearch"),
				ResourceId:    new(mock.ValidClusterID),
				Region:        new("some-region"),
				HttpEndpoint:  new("http://othercluster.cloud.elastic.co:9200"),
				HttpsEndpoint: new("https://othercluster.cloud.elastic.co:9243"),
				Config: &ElasticsearchConfig{
					Plugins:                  []string{},
					UserSettingsYaml:         new("some.setting: value"),
					UserSettingsOverrideYaml: new("some.setting: value2"),
					UserSettingsJson:         new("{\"some.setting\":\"value\"}"),
					UserSettingsOverrideJson: new("{\"some.setting\":\"value2\"}"),
				},
				HotTier: &ElasticsearchTopology{
					id:                      "hot_content",
					InstanceConfigurationId: new("aws.data.highio.i3"),
					Size:                    new("2g"),
					SizeResource:            new("memory"),
					NodeTypeData:            new("true"),
					NodeTypeIngest:          new("true"),
					NodeTypeMaster:          new("true"),
					NodeTypeMl:              new("false"),
					ZoneCount:               1,
					Autoscaling:             &ElasticsearchTopologyAutoscaling{},
				},
				TrustAccount:  ElasticsearchTrustAccounts{},
				TrustExternal: ElasticsearchTrustExternals{},
			},
		},
		{
			name: "parses an elasticsearch resource with snapshot repository",
			args: args{in: []*models.ElasticsearchResourceInfo{
				{
					Region: new("some-region"),
					RefID:  new("main-elasticsearch"),
					Info: &models.ElasticsearchClusterInfo{
						ClusterID: &mock.ValidClusterID,
						Region:    "some-region",
						Status:    new("started"),
						Metadata: &models.ClusterMetadataInfo{
							CloudID:  "some CLOUD ID",
							Endpoint: "somecluster.cloud.elastic.co",
							Ports: &models.ClusterMetadataPortInfo{
								HTTP:  ec.Int32(9200),
								HTTPS: ec.Int32(9243),
							},
						},
						Settings: &models.ElasticsearchClusterSettings{
							Snapshot: &models.ClusterSnapshotSettings{
								Enabled: new(true),
								Repository: &models.ClusterSnapshotRepositoryInfo{
									Reference: &models.ClusterSnapshotRepositoryReference{
										RepositoryName: "my-snapshot-repository",
									},
								},
							},
						},
						PlanInfo: &models.ElasticsearchClusterPlansInfo{
							Current: &models.ElasticsearchClusterPlanInfo{
								Plan: &models.ElasticsearchClusterPlan{
									Elasticsearch: &models.ElasticsearchConfiguration{
										Version: "7.7.0",
									},
									ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
										{
											ID:                      "hot_content",
											ZoneCount:               1,
											InstanceConfigurationID: "aws.data.highio.i3",
											Size: &models.TopologySize{
												Resource: new("memory"),
												Value:    ec.Int32(2048),
											},
											NodeType: &models.ElasticsearchNodeType{
												Data:   new(true),
												Ingest: new(true),
												Master: new(true),
												Ml:     new(false),
											},
										},
									},
									Transient: &models.TransientElasticsearchPlanConfiguration{
										Strategy: &models.PlanStrategy{
											RollingGrowAndShrink: new(models.RollingGrowShrinkStrategyConfig),
										},
									},
								},
							},
						},
					},
				},
				{
					Region: new("some-region"),
					RefID:  new("main-elasticsearch"),
					Info: &models.ElasticsearchClusterInfo{
						ClusterID: &mock.ValidClusterID,
						Region:    "some-region",
						Status:    new("stopped"),
						Metadata: &models.ClusterMetadataInfo{
							CloudID:  "some CLOUD ID",
							Endpoint: "somecluster.cloud.elastic.co",
							Ports: &models.ClusterMetadataPortInfo{
								HTTP:  ec.Int32(9200),
								HTTPS: ec.Int32(9243),
							},
						},
						PlanInfo: &models.ElasticsearchClusterPlansInfo{
							Current: &models.ElasticsearchClusterPlanInfo{
								Plan: &models.ElasticsearchClusterPlan{
									Elasticsearch: &models.ElasticsearchConfiguration{
										Version: "7.7.0",
									},
									ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
										{
											ID:                      "hot_content",
											ZoneCount:               1,
											InstanceConfigurationID: "aws.data.highio.i3",
											Size: &models.TopologySize{
												Resource: new("memory"),
												Value:    ec.Int32(2048),
											},
											NodeType: &models.ElasticsearchNodeType{
												Data:   new(true),
												Ingest: new(true),
												Master: new(true),
												Ml:     new(false),
											},
										},
									},
								},
							},
						},
					},
				},
			}},
			want: &Elasticsearch{
				RefId:         new("main-elasticsearch"),
				ResourceId:    new(mock.ValidClusterID),
				Region:        new("some-region"),
				CloudID:       new("some CLOUD ID"),
				HttpEndpoint:  new("http://somecluster.cloud.elastic.co:9200"),
				HttpsEndpoint: new("https://somecluster.cloud.elastic.co:9243"),
				Config: &ElasticsearchConfig{
					Plugins: []string{},
				},
				HotTier: &ElasticsearchTopology{
					id:                      "hot_content",
					InstanceConfigurationId: new("aws.data.highio.i3"),
					Size:                    new("2g"),
					SizeResource:            new("memory"),
					NodeTypeData:            new("true"),
					NodeTypeIngest:          new("true"),
					NodeTypeMaster:          new("true"),
					NodeTypeMl:              new("false"),
					ZoneCount:               1,
					Autoscaling:             &ElasticsearchTopologyAutoscaling{},
				},
				Snapshot: &ElasticsearchSnapshot{
					Enabled: true,
					Repository: &ElasticsearchSnapshotRepositoryInfo{
						Reference: &ElasticsearchSnapshotRepositoryReference{
							RepositoryName: "my-snapshot-repository",
						},
					},
				},
				TrustAccount:  ElasticsearchTrustAccounts{},
				TrustExternal: ElasticsearchTrustExternals{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadElasticsearches(tt.args.in, &tt.args.remotes)
			assert.Nil(t, err)
			assert.Equal(t, tt.want, got)

			var esObj types.Object
			diags := tfsdk.ValueFrom(context.Background(), got, ElasticsearchSchema().GetType(), &esObj)
			assert.Nil(t, diags)
		})
	}
}

func Test_readElasticsearchTopology(t *testing.T) {
	type args struct {
		plan *models.ElasticsearchClusterPlan
	}
	tests := []struct {
		name string
		args args
		want ElasticsearchTopologies
		err  string
	}{
		{
			name: "all topologies (even with 0 size) are returned",
			args: args{plan: &models.ElasticsearchClusterPlan{
				ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
					{
						ID:                      "hot_content",
						ZoneCount:               1,
						InstanceConfigurationID: "aws.data.highio.i3",
						Size: &models.TopologySize{
							Value: ec.Int32(4096), Resource: new("memory"),
						},
						NodeType: &models.ElasticsearchNodeType{
							Data:   new(true),
							Ingest: new(true),
							Master: new(true),
						},
					},
					{
						ID:                      "coordinating",
						ZoneCount:               2,
						InstanceConfigurationID: "aws.coordinating.m5",
						Size: &models.TopologySize{
							Value: ec.Int32(0), Resource: new("memory"),
						},
					},
				},
			}},
			want: ElasticsearchTopologies{
				{
					id:                      "hot_content",
					InstanceConfigurationId: new("aws.data.highio.i3"),
					Size:                    new("4g"),
					SizeResource:            new("memory"),
					ZoneCount:               1,
					NodeTypeData:            new("true"),
					NodeTypeIngest:          new("true"),
					NodeTypeMaster:          new("true"),
					Autoscaling:             &ElasticsearchTopologyAutoscaling{},
				},
				{
					id:                      "coordinating",
					InstanceConfigurationId: new("aws.coordinating.m5"),
					Size:                    new("0g"),
					SizeResource:            new("memory"),
					ZoneCount:               2,
					Autoscaling:             &ElasticsearchTopologyAutoscaling{},
				},
			},
		},
		{
			name: "includes unsized autoscaling topologies",
			args: args{plan: &models.ElasticsearchClusterPlan{
				AutoscalingEnabled: new(true),
				ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
					{
						ID:                      "hot_content",
						ZoneCount:               1,
						InstanceConfigurationID: "aws.data.highio.i3",
						Size: &models.TopologySize{
							Value: ec.Int32(4096), Resource: new("memory"),
						},
						NodeType: &models.ElasticsearchNodeType{
							Data:   new(true),
							Ingest: new(true),
							Master: new(true),
						},
					},
					{
						ID:                           "ml",
						ZoneCount:                    1,
						InstanceConfigurationID:      "aws.ml.m5",
						InstanceConfigurationVersion: ec.Int32(2),
						Size: &models.TopologySize{
							Value: ec.Int32(0), Resource: new("memory"),
						},
						AutoscalingMax: &models.TopologySize{
							Value: ec.Int32(8192), Resource: new("memory"),
						},
						AutoscalingMin: &models.TopologySize{
							Value: ec.Int32(0), Resource: new("memory"),
						},
					},
				},
			}},
			want: ElasticsearchTopologies{
				{
					id:                      "hot_content",
					InstanceConfigurationId: new("aws.data.highio.i3"),
					Size:                    new("4g"),
					SizeResource:            new("memory"),
					ZoneCount:               1,
					NodeTypeData:            new("true"),
					NodeTypeIngest:          new("true"),
					NodeTypeMaster:          new("true"),
					Autoscaling:             &ElasticsearchTopologyAutoscaling{},
				},
				{
					id:                           "ml",
					InstanceConfigurationId:      new("aws.ml.m5"),
					InstanceConfigurationVersion: new(2),
					Size:                         new("0g"),
					SizeResource:                 new("memory"),
					ZoneCount:                    1,
					Autoscaling: &ElasticsearchTopologyAutoscaling{
						MaxSize:         new("8g"),
						MaxSizeResource: new("memory"),
						MinSize:         new("0g"),
						MinSizeResource: new("memory"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readElasticsearchTopologies(tt.args.plan)
			if err != nil && !assert.EqualError(t, err, tt.err) {
				t.Error(err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_readElasticsearchConfig(t *testing.T) {
	type args struct {
		cfg *models.ElasticsearchConfiguration
	}
	tests := []struct {
		name string
		args args
		want *ElasticsearchConfig
	}{
		{
			name: "read plugins allowlist",
			args: args{cfg: &models.ElasticsearchConfiguration{
				EnabledBuiltInPlugins: []string{"some-allowed-plugin"},
			}},
			want: &ElasticsearchConfig{
				Plugins: []string{"some-allowed-plugin"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readElasticsearchConfig(tt.args.cfg)
			assert.Nil(t, err)
			assert.Equal(t, tt.want, got)

			var config types.Object
			diags := tfsdk.ValueFrom(context.Background(), got, elasticsearchConfigSchema().GetType(), &config)
			assert.Nil(t, diags)
		})
	}
}

func Test_IsEsResourceStopped(t *testing.T) {
	type args struct {
		res *models.ElasticsearchResourceInfo
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "started resource returns false",
			args: args{res: &models.ElasticsearchResourceInfo{Info: &models.ElasticsearchClusterInfo{
				Status: new("started"),
			}}},
			want: false,
		},
		{
			name: "stopped resource returns true",
			args: args{res: &models.ElasticsearchResourceInfo{Info: &models.ElasticsearchClusterInfo{
				Status: new("stopped"),
			}}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsElasticsearchStopped(tt.args.res)
			assert.Equal(t, tt.want, got)
		})
	}
}
