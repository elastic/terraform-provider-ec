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
)

func Test_readElasticsearch(t *testing.T) {
	type args struct {
		in      []*models.ElasticsearchResourceInfo
		remotes models.RemoteResources
	}
	tests := []struct {
		name  string
		args  args
		want  *Elasticsearch
		diags diag.Diagnostics
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
					Region: ec.String("some-region"),
					RefID:  ec.String("main-elasticsearch"),
					Info: &models.ElasticsearchClusterInfo{
						ClusterID: &mock.ValidClusterID,
						Region:    "some-region",
						Status:    ec.String("started"),
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
												Resource: ec.String("memory"),
												Value:    ec.Int32(2048),
											},
											NodeType: &models.ElasticsearchNodeType{
												Data:   ec.Bool(true),
												Ingest: ec.Bool(true),
												Master: ec.Bool(true),
												Ml:     ec.Bool(false),
											},
										},
									},
								},
							},
						},
					},
				},
				{
					Region: ec.String("some-region"),
					RefID:  ec.String("main-elasticsearch"),
					Info: &models.ElasticsearchClusterInfo{
						ClusterID: &mock.ValidClusterID,
						Region:    "some-region",
						Status:    ec.String("stopped"),
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
												Resource: ec.String("memory"),
												Value:    ec.Int32(2048),
											},
											NodeType: &models.ElasticsearchNodeType{
												Data:   ec.Bool(true),
												Ingest: ec.Bool(true),
												Master: ec.Bool(true),
												Ml:     ec.Bool(false),
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
				RefId:         ec.String("main-elasticsearch"),
				ResourceId:    ec.String(mock.ValidClusterID),
				Region:        ec.String("some-region"),
				CloudID:       ec.String("some CLOUD ID"),
				HttpEndpoint:  ec.String("http://somecluster.cloud.elastic.co:9200"),
				HttpsEndpoint: ec.String("https://somecluster.cloud.elastic.co:9243"),
				Config:        &ElasticsearchConfig{},
				HotTier: &ElasticsearchTopology{
					id:                      "hot_content",
					InstanceConfigurationId: ec.String("aws.data.highio.i3"),
					Size:                    ec.String("2g"),
					SizeResource:            ec.String("memory"),
					NodeTypeData:            ec.String("true"),
					NodeTypeIngest:          ec.String("true"),
					NodeTypeMaster:          ec.String("true"),
					NodeTypeMl:              ec.String("false"),
					ZoneCount:               1,
					Autoscaling:             &ElasticsearchTopologyAutoscaling{},
				},
			},
		},
		{
			name: "resource with a config object",
			args: args{in: []*models.ElasticsearchResourceInfo{
				{
					Region: ec.String("some-region"),
					RefID:  ec.String("main-elasticsearch"),
					Info: &models.ElasticsearchClusterInfo{
						ClusterID:   &mock.ValidClusterID,
						ClusterName: ec.String("some-name"),
						Region:      "some-region",
						Status:      ec.String("started"),
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
										UserSettingsJSON: map[string]interface{}{
											"some.setting": "value",
										},
										UserSettingsOverrideJSON: map[string]interface{}{
											"some.setting": "value2",
										},
									},
									ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
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
											Ml:     ec.Bool(false),
										},
									}},
								},
							},
						},
					},
				},
			}},
			want: &Elasticsearch{
				RefId:         ec.String("main-elasticsearch"),
				ResourceId:    ec.String(mock.ValidClusterID),
				Region:        ec.String("some-region"),
				HttpEndpoint:  ec.String("http://othercluster.cloud.elastic.co:9200"),
				HttpsEndpoint: ec.String("https://othercluster.cloud.elastic.co:9243"),
				Config: &ElasticsearchConfig{
					UserSettingsYaml:         ec.String("some.setting: value"),
					UserSettingsOverrideYaml: ec.String("some.setting: value2"),
					UserSettingsJson:         ec.String("{\"some.setting\":\"value\"}"),
					UserSettingsOverrideJson: ec.String("{\"some.setting\":\"value2\"}"),
				},
				HotTier: &ElasticsearchTopology{
					id:                      "hot_content",
					InstanceConfigurationId: ec.String("aws.data.highio.i3"),
					Size:                    ec.String("2g"),
					SizeResource:            ec.String("memory"),
					NodeTypeData:            ec.String("true"),
					NodeTypeIngest:          ec.String("true"),
					NodeTypeMaster:          ec.String("true"),
					NodeTypeMl:              ec.String("false"),
					ZoneCount:               1,
					Autoscaling:             &ElasticsearchTopologyAutoscaling{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadElasticsearches(tt.args.in, &tt.args.remotes)
			assert.Nil(t, err)
			assert.Equal(t, tt.want, got)

			var esObj types.Object
			diags := tfsdk.ValueFrom(context.Background(), got, ElasticsearchSchema().FrameworkType(), &esObj)
			if tt.diags.HasError() {
				assert.Equal(t, tt.diags, diags)
			}
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
							Value: ec.Int32(4096), Resource: ec.String("memory"),
						},
						NodeType: &models.ElasticsearchNodeType{
							Data:   ec.Bool(true),
							Ingest: ec.Bool(true),
							Master: ec.Bool(true),
						},
					},
					{
						ID:                      "coordinating",
						ZoneCount:               2,
						InstanceConfigurationID: "aws.coordinating.m5",
						Size: &models.TopologySize{
							Value: ec.Int32(0), Resource: ec.String("memory"),
						},
					},
				},
			}},
			want: ElasticsearchTopologies{
				{
					id:                      "hot_content",
					InstanceConfigurationId: ec.String("aws.data.highio.i3"),
					Size:                    ec.String("4g"),
					SizeResource:            ec.String("memory"),
					ZoneCount:               1,
					NodeTypeData:            ec.String("true"),
					NodeTypeIngest:          ec.String("true"),
					NodeTypeMaster:          ec.String("true"),
					Autoscaling:             &ElasticsearchTopologyAutoscaling{},
				},
				{
					id:                      "coordinating",
					InstanceConfigurationId: ec.String("aws.coordinating.m5"),
					Size:                    ec.String("0g"),
					SizeResource:            ec.String("memory"),
					ZoneCount:               2,
					Autoscaling:             &ElasticsearchTopologyAutoscaling{},
				},
			},
		},
		{
			name: "includes unsized autoscaling topologies",
			args: args{plan: &models.ElasticsearchClusterPlan{
				AutoscalingEnabled: ec.Bool(true),
				ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
					{
						ID:                      "hot_content",
						ZoneCount:               1,
						InstanceConfigurationID: "aws.data.highio.i3",
						Size: &models.TopologySize{
							Value: ec.Int32(4096), Resource: ec.String("memory"),
						},
						NodeType: &models.ElasticsearchNodeType{
							Data:   ec.Bool(true),
							Ingest: ec.Bool(true),
							Master: ec.Bool(true),
						},
					},
					{
						ID:                      "ml",
						ZoneCount:               1,
						InstanceConfigurationID: "aws.ml.m5",
						Size: &models.TopologySize{
							Value: ec.Int32(0), Resource: ec.String("memory"),
						},
						AutoscalingMax: &models.TopologySize{
							Value: ec.Int32(8192), Resource: ec.String("memory"),
						},
						AutoscalingMin: &models.TopologySize{
							Value: ec.Int32(0), Resource: ec.String("memory"),
						},
					},
				},
			}},
			want: ElasticsearchTopologies{
				{
					id:                      "hot_content",
					InstanceConfigurationId: ec.String("aws.data.highio.i3"),
					Size:                    ec.String("4g"),
					SizeResource:            ec.String("memory"),
					ZoneCount:               1,
					NodeTypeData:            ec.String("true"),
					NodeTypeIngest:          ec.String("true"),
					NodeTypeMaster:          ec.String("true"),
					Autoscaling:             &ElasticsearchTopologyAutoscaling{},
				},
				{
					id:                      "ml",
					InstanceConfigurationId: ec.String("aws.ml.m5"),
					Size:                    ec.String("0g"),
					SizeResource:            ec.String("memory"),
					ZoneCount:               1,
					Autoscaling: &ElasticsearchTopologyAutoscaling{
						MaxSize:         ec.String("8g"),
						MaxSizeResource: ec.String("memory"),
						MinSize:         ec.String("0g"),
						MinSizeResource: ec.String("memory"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadElasticsearchTopologies(tt.args.plan)
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
			got, err := ReadElasticsearchConfig(tt.args.cfg)
			assert.Nil(t, err)
			assert.Equal(t, tt.want, got)

			var config types.Object
			diags := tfsdk.ValueFrom(context.Background(), got, ElasticsearchConfigSchema().FrameworkType(), &config)
			assert.Nil(t, diags)
		})
	}
}
