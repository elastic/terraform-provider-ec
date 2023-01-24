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

package deploymentresource

import (
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func Test_flattenEsResource(t *testing.T) {
	type args struct {
		in      []*models.ElasticsearchResourceInfo
		remotes models.RemoteResources
	}
	tests := []struct {
		name string
		args args
		want []interface{}
		err  string
	}{
		{
			name: "empty resource list returns empty list",
			args: args{in: []*models.ElasticsearchResourceInfo{}},
			want: []interface{}{},
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
			want: []interface{}{},
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
			want: []interface{}{
				map[string]interface{}{
					"ref_id":         "main-elasticsearch",
					"resource_id":    mock.ValidClusterID,
					"region":         "some-region",
					"cloud_id":       "some CLOUD ID",
					"http_endpoint":  "http://somecluster.cloud.elastic.co:9200",
					"https_endpoint": "https://somecluster.cloud.elastic.co:9243",
					"config":         func() []interface{} { return nil }(),
					"topology": []interface{}{
						map[string]interface{}{
							"config":                    func() []interface{} { return nil }(),
							"id":                        "hot_content",
							"instance_configuration_id": "aws.data.highio.i3",
							"size":                      "2g",
							"size_resource":             "memory",
							"node_type_data":            "true",
							"node_type_ingest":          "true",
							"node_type_master":          "true",
							"node_type_ml":              "false",
							"zone_count":                int32(1),
						},
					},
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
			want: []interface{}{map[string]interface{}{
				"ref_id":         "main-elasticsearch",
				"resource_id":    mock.ValidClusterID,
				"region":         "some-region",
				"http_endpoint":  "http://othercluster.cloud.elastic.co:9200",
				"https_endpoint": "https://othercluster.cloud.elastic.co:9243",
				"config": []interface{}{map[string]interface{}{
					"user_settings_yaml":          "some.setting: value",
					"user_settings_override_yaml": "some.setting: value2",
					"user_settings_json":          "{\"some.setting\":\"value\"}",
					"user_settings_override_json": "{\"some.setting\":\"value2\"}",
				}},
				"topology": []interface{}{map[string]interface{}{
					"config":                    func() []interface{} { return nil }(),
					"id":                        "hot_content",
					"instance_configuration_id": "aws.data.highio.i3",
					"size":                      "2g",
					"size_resource":             "memory",
					"node_type_data":            "true",
					"node_type_ingest":          "true",
					"node_type_master":          "true",
					"node_type_ml":              "false",
					"zone_count":                int32(1),
				}},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := flattenEsResources(tt.args.in, tt.args.remotes)
			if err != nil && !assert.EqualError(t, err, tt.err) {
				t.Error(err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_flattenEsTopology(t *testing.T) {
	type args struct {
		plan *models.ElasticsearchClusterPlan
	}
	tests := []struct {
		name string
		args args
		want []interface{}
		err  string
	}{
		{
			name: "no zombie topologies",
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
			want: []interface{}{map[string]interface{}{
				"config":                    func() []interface{} { return nil }(),
				"id":                        "hot_content",
				"instance_configuration_id": "aws.data.highio.i3",
				"size":                      "4g",
				"size_resource":             "memory",
				"zone_count":                int32(1),
				"node_type_data":            "true",
				"node_type_ingest":          "true",
				"node_type_master":          "true",
			}},
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
			want: []interface{}{
				map[string]interface{}{
					"config":                    func() []interface{} { return nil }(),
					"id":                        "hot_content",
					"instance_configuration_id": "aws.data.highio.i3",
					"size":                      "4g",
					"size_resource":             "memory",
					"zone_count":                int32(1),
					"node_type_data":            "true",
					"node_type_ingest":          "true",
					"node_type_master":          "true",
				},
				map[string]interface{}{
					"config":                    func() []interface{} { return nil }(),
					"id":                        "ml",
					"instance_configuration_id": "aws.ml.m5",
					"size":                      "0g",
					"size_resource":             "memory",
					"zone_count":                int32(1),
					"autoscaling": []interface{}{
						map[string]interface{}{
							"max_size":          "8g",
							"max_size_resource": "memory",
							"min_size":          "0g",
							"min_size_resource": "memory",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := flattenEsTopology(tt.args.plan)
			if err != nil && !assert.EqualError(t, err, tt.err) {
				t.Error(err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_flattenEsConfig(t *testing.T) {
	type args struct {
		cfg *models.ElasticsearchConfiguration
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "flattens plugins allowlist",
			args: args{cfg: &models.ElasticsearchConfiguration{
				EnabledBuiltInPlugins: []string{"some-allowed-plugin"},
			}},
			want: []interface{}{map[string]interface{}{
				"plugins": []interface{}{"some-allowed-plugin"},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := flattenEsConfig(tt.args.cfg)
			for _, g := range got {
				var rawVal []interface{}
				m := g.(map[string]interface{})
				if v, ok := m["plugins"]; ok {
					rawVal = v.(*schema.Set).List()
				}
				m["plugins"] = rawVal
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
