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

func Test_expandKibanaResources(t *testing.T) {
	type args struct {
		ess []interface{}
	}
	tests := []struct {
		name string
		args args
		want []*models.KibanaPayload
		err  error
	}{
		{
			name: "returns nil when there's no resources",
		},
		{
			name: "parses multiple resources",
			args: args{
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":                       "main-kibana",
						"resource_id":                  mock.ValidClusterID,
						"version":                      "7.7.0",
						"region":                       "some-region",
						"elasticsearch_cluster_ref_id": "somerefid",
						"topology": []interface{}{map[string]interface{}{
							"instance_configuration_id": "aws.kibana.r4",
							"memory_per_node":           "2g",
							"zone_count":                1,
						}},
					},
					map[string]interface{}{
						"ref_id":                       "secondary-kibana",
						"elasticsearch_cluster_ref_id": "somerefid",
						"resource_id":                  mock.ValidClusterID,
						"version":                      "7.6.0",
						"region":                       "some-region",
						"topology": []interface{}{map[string]interface{}{
							"instance_configuration_id": "aws.kibana.r4",
							"memory_per_node":           "4g",
							"zone_count":                1,
						}},
					},
					map[string]interface{}{
						"ref_id":                       "secondary-kibana",
						"elasticsearch_cluster_ref_id": "somerefid",
						"resource_id":                  mock.ValidClusterID,
						"version":                      "7.8.0",
						"region":                       "some-region",
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
							"memory_per_node":           "4g",
							"zone_count":                1,
						}},
					},
				},
			},
			want: []*models.KibanaPayload{
				{
					ElasticsearchClusterRefID: ec.String("somerefid"),
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("main-kibana"),
					Settings:                  &models.KibanaClusterSettings{},
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
									Value:    ec.Int32(2048),
								},
							},
						},
					},
				},
				{
					ElasticsearchClusterRefID: ec.String("somerefid"),
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("secondary-kibana"),
					Settings:                  &models.KibanaClusterSettings{},
					Plan: &models.KibanaClusterPlan{
						Kibana: &models.KibanaConfiguration{
							Version: "7.6.0",
						},
						ClusterTopology: []*models.KibanaClusterTopologyElement{{
							ZoneCount:               1,
							InstanceConfigurationID: "aws.kibana.r4",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(4096),
							},
						}},
					},
				},
				{
					ElasticsearchClusterRefID: ec.String("somerefid"),
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("secondary-kibana"),
					Settings:                  &models.KibanaClusterSettings{},
					Plan: &models.KibanaClusterPlan{
						Kibana: &models.KibanaConfiguration{
							Version:                  "7.8.0",
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
								Value:    ec.Int32(4096),
							},
						}},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandKibanaResources(tt.args.ess)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
