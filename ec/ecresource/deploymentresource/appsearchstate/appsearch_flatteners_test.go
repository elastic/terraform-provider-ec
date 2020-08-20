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

package appsearchstate

import (
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/stretchr/testify/assert"
)

func TestFlattenResource(t *testing.T) {
	type args struct {
		in   []*models.AppSearchResourceInfo
		name string
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "empty resource list returns empty list",
			args: args{in: []*models.AppSearchResourceInfo{}},
			want: []interface{}{},
		},
		{
			name: "empty current plan returns empty list",
			args: args{in: []*models.AppSearchResourceInfo{{
				Info: &models.AppSearchInfo{
					PlanInfo: &models.AppSearchPlansInfo{
						Pending: &models.AppSearchPlanInfo{},
					},
				},
			}}},
			want: []interface{}{},
		},
		{
			name: "parses the appsearch resource",
			args: args{in: []*models.AppSearchResourceInfo{
				{
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("main-appsearch"),
					ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
					Info: &models.AppSearchInfo{
						ID:     &mock.ValidClusterID,
						Name:   ec.String("some-appsearch-name"),
						Region: "some-region",
						Metadata: &models.ClusterMetadataInfo{
							Endpoint: "appsearchresource.cloud.elastic.co",
							Ports: &models.ClusterMetadataPortInfo{
								HTTP:  ec.Int32(9200),
								HTTPS: ec.Int32(9243),
							},
						},
						PlanInfo: &models.AppSearchPlansInfo{
							Current: &models.AppSearchPlanInfo{
								Plan: &models.AppSearchPlan{
									Appsearch: &models.AppSearchConfiguration{
										Version:                  "7.6.2",
										UserSettingsYaml:         "some.setting: some value",
										UserSettingsOverrideYaml: "some.setting: some override",
										UserSettingsJSON:         `{"some.setting": "some other value"}`,
										UserSettingsOverrideJSON: `{"some.setting": "some other override"}`,
									},
									ClusterTopology: []*models.AppSearchTopologyElement{{
										Appsearch: &models.AppSearchConfiguration{
											SystemSettings: &models.AppSearchSystemSettings{
												SecretSessionKey: "somekey secret key",
											},
										},
										ZoneCount:               1,
										InstanceConfigurationID: "aws.appsearch.r4",
										Size: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
										NodeType: &models.AppSearchNodeTypes{
											Appserver: ec.Bool(true),
											Worker:    ec.Bool(false),
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
					"display_name":                 "some-appsearch-name",
					"ref_id":                       "main-appsearch",
					"resource_id":                  mock.ValidClusterID,
					"version":                      "7.6.2",
					"region":                       "some-region",
					"http_endpoint":                "http://appsearchresource.cloud.elastic.co:9200",
					"https_endpoint":               "https://appsearchresource.cloud.elastic.co:9243",
					"config": []interface{}{map[string]interface{}{
						"user_settings_json":          "{\"some.setting\": \"some other value\"}",
						"user_settings_override_json": "{\"some.setting\": \"some other override\"}",
						"user_settings_override_yaml": "some.setting: some override",
						"user_settings_yaml":          "some.setting: some value",
					}},
					"topology": []interface{}{map[string]interface{}{
						"instance_configuration_id": "aws.appsearch.r4",
						"memory_per_node":           "1g",
						"zone_count":                int32(1),
						"node_type_appserver":       true,
						"node_type_worker":          false,
						"config": []interface{}{map[string]interface{}{
							"secret_session_key": "somekey secret key",
						}},
					}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FlattenResources(tt.args.in, tt.args.name)
			assert.Equal(t, tt.want, got)
		})
	}
}
