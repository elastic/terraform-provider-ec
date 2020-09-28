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
	"github.com/stretchr/testify/assert"
)

func Test_flattenKibanaResources(t *testing.T) {
	type args struct {
		in   []*models.KibanaResourceInfo
		name string
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "empty resource list returns empty list",
			args: args{in: []*models.KibanaResourceInfo{}},
			want: []interface{}{},
		},
		{
			name: "empty current plan returns empty list",
			args: args{in: []*models.KibanaResourceInfo{
				{
					Info: &models.KibanaClusterInfo{
						PlanInfo: &models.KibanaClusterPlansInfo{
							Pending: &models.KibanaClusterPlanInfo{},
						},
					},
				},
			}},
			want: []interface{}{},
		},
		{
			name: "parses the kibana resource",
			args: args{in: []*models.KibanaResourceInfo{
				{
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("main-kibana"),
					ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
					Info: &models.KibanaClusterInfo{
						ClusterID:   &mock.ValidClusterID,
						ClusterName: ec.String("some-kibana-name"),
						Region:      "some-region",
						Status:      ec.String("started"),
						Metadata: &models.ClusterMetadataInfo{
							Endpoint: "kibanaresource.cloud.elastic.co",
							Ports: &models.ClusterMetadataPortInfo{
								HTTP:  ec.Int32(9200),
								HTTPS: ec.Int32(9243),
							},
						},
						PlanInfo: &models.KibanaClusterPlansInfo{
							Current: &models.KibanaClusterPlanInfo{
								Plan: &models.KibanaClusterPlan{
									Kibana: &models.KibanaConfiguration{
										Version: "7.7.0",
									},
									ClusterTopology: []*models.KibanaClusterTopologyElement{
										{
											ZoneCount:               1,
											InstanceConfigurationID: "aws.kibana.r4",
											Size: &models.TopologySize{
												Resource: ec.String("memory"),
												Value:    ec.Int32(1024),
											},
										},
									},
								},
							},
						},
					},
				},
				{
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("main-kibana"),
					ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
					Info: &models.KibanaClusterInfo{
						ClusterID:   &mock.ValidClusterID,
						ClusterName: ec.String("some-kibana-name"),
						Region:      "some-region",
						Status:      ec.String("stopped"),
						Metadata: &models.ClusterMetadataInfo{
							Endpoint: "kibanaresource.cloud.elastic.co",
							Ports: &models.ClusterMetadataPortInfo{
								HTTP:  ec.Int32(9200),
								HTTPS: ec.Int32(9243),
							},
						},
						PlanInfo: &models.KibanaClusterPlansInfo{
							Current: &models.KibanaClusterPlanInfo{
								Plan: &models.KibanaClusterPlan{
									Kibana: &models.KibanaConfiguration{
										Version: "7.7.0",
									},
									ClusterTopology: []*models.KibanaClusterTopologyElement{
										{
											ZoneCount:               1,
											InstanceConfigurationID: "aws.kibana.r4",
											Size: &models.TopologySize{
												Resource: ec.String("memory"),
												Value:    ec.Int32(1024),
											},
										},
									},
								},
							},
						},
					},
				},
				{
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("main-kibana"),
					ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
					Info: &models.KibanaClusterInfo{
						ClusterID:   &mock.ValidClusterID,
						ClusterName: ec.String("some-kibana-name"),
						Region:      "some-region",
						Status:      ec.String("started"),
						Metadata: &models.ClusterMetadataInfo{
							Endpoint: "kibanaresource.cloud.elastic.co",
							Ports: &models.ClusterMetadataPortInfo{
								HTTP:  ec.Int32(9200),
								HTTPS: ec.Int32(9243),
							},
						},
						PlanInfo: &models.KibanaClusterPlansInfo{
							Current: &models.KibanaClusterPlanInfo{
								Plan: &models.KibanaClusterPlan{
									Kibana: &models.KibanaConfiguration{
										Version:                  "7.7.0",
										UserSettingsYaml:         "some.setting: value",
										UserSettingsOverrideYaml: "some.setting: override",
										UserSettingsJSON:         "{\"some.setting\": \"value\"}",
										UserSettingsOverrideJSON: "{\"some.setting\": \"override\"}",
									},
									ClusterTopology: []*models.KibanaClusterTopologyElement{{
										Kibana: &models.KibanaConfiguration{
											UserSettingsYaml:         "some.setting: value",
											UserSettingsOverrideYaml: "some.setting: override",
											UserSettingsJSON:         "{\"some.setting\": \"value\"}",
											UserSettingsOverrideJSON: "{\"some.setting\": \"override\"}",
										},
										ZoneCount:               1,
										InstanceConfigurationID: "aws.kibana.r4",
										Size: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									}},
								},
							},
						},
					},
				},
			}},
			want: []interface{}{
				map[string]interface{}{
					"elasticsearch_cluster_ref_id": "main-elasticsearch",
					"ref_id":                       "main-kibana",
					"resource_id":                  mock.ValidClusterID,
					"version":                      "7.7.0",
					"region":                       "some-region",
					"http_endpoint":                "http://kibanaresource.cloud.elastic.co:9200",
					"https_endpoint":               "https://kibanaresource.cloud.elastic.co:9243",
					"topology": []interface{}{
						map[string]interface{}{
							"instance_configuration_id": "aws.kibana.r4",
							"memory_per_node":           "1g",
							"zone_count":                int32(1),
						},
					},
				},
				map[string]interface{}{
					"elasticsearch_cluster_ref_id": "main-elasticsearch",
					"ref_id":                       "main-kibana",
					"resource_id":                  mock.ValidClusterID,
					"version":                      "7.7.0",
					"region":                       "some-region",
					"http_endpoint":                "http://kibanaresource.cloud.elastic.co:9200",
					"https_endpoint":               "https://kibanaresource.cloud.elastic.co:9243",
					"config": []interface{}{map[string]interface{}{
						"user_settings_yaml":          "some.setting: value",
						"user_settings_override_yaml": "some.setting: override",
						"user_settings_json":          `{"some.setting": "value"}`,
						"user_settings_override_json": `{"some.setting": "override"}`,
					}},
					"topology": []interface{}{map[string]interface{}{
						"config": []interface{}{map[string]interface{}{
							"user_settings_yaml":          "some.setting: value",
							"user_settings_override_yaml": "some.setting: override",
							"user_settings_json":          `{"some.setting": "value"}`,
							"user_settings_override_json": `{"some.setting": "override"}`,
						}},
						"instance_configuration_id": "aws.kibana.r4",
						"memory_per_node":           "1g",
						"zone_count":                int32(1),
					}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := flattenKibanaResources(tt.args.in, tt.args.name)
			assert.Equal(t, tt.want, got)
		})
	}
}
