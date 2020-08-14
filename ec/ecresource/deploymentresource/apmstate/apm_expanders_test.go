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

package apmstate

import (
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/stretchr/testify/assert"
)

func TestExpandResources(t *testing.T) {
	type args struct {
		ess []interface{}
	}
	tests := []struct {
		name string
		args args
		want []*models.ApmPayload
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
						"ref_id":                       "main-apm",
						"resource_id":                  mock.ValidClusterID,
						"version":                      "7.7.0",
						"region":                       "some-region",
						"elasticsearch_cluster_ref_id": "somerefid",
						"topology": []interface{}{map[string]interface{}{
							"instance_configuration_id": "aws.apm.r4",
							"memory_per_node":           "2g",
							"zone_count":                1,
						}},
					},
					map[string]interface{}{
						"display_name":                 "somename",
						"ref_id":                       "secondary-apm",
						"elasticsearch_cluster_ref_id": "somerefid",
						"resource_id":                  mock.ValidClusterID,
						"version":                      "7.6.0",
						"region":                       "some-region",
						"topology": []interface{}{map[string]interface{}{
							"instance_configuration_id": "aws.apm.r4",
							"memory_per_node":           "4g",
							"zone_count":                1,
						}},
					},
					map[string]interface{}{
						"display_name":                 "somename",
						"ref_id":                       "tertiary-apm",
						"elasticsearch_cluster_ref_id": "somerefid",
						"resource_id":                  mock.ValidClusterID,
						"version":                      "7.8.0",
						"region":                       "some-region",
						"topology": []interface{}{map[string]interface{}{
							"instance_configuration_id": "aws.apm.r4",
							"memory_per_node":           "4g",
							"zone_count":                1,
							"config": []interface{}{map[string]interface{}{
								"docker_image":                "some-other-image",
								"user_settings_yaml":          "some.setting: value",
								"user_settings_override_yaml": "some.setting: value2",
								"user_settings_json":          "{\"some.setting\": \"value\"}",
								"user_settings_override_json": "{\"some.setting\": \"value2\"}",

								"debug_enabled":          true,
								"elasticsearch_password": "somepass",
								"elasticsearch_username": "someuser",
								"elasticsearch_url":      "someURL",
								"kibana_url":             "someKibanaURL",
								"secret_token":           "very_secret",
							}},
						}},
						"config": []interface{}{map[string]interface{}{
							"docker_image":  "some-docker-image:version",
							"debug_enabled": true,
						}},
					},
				},
			},
			want: []*models.ApmPayload{
				{
					ElasticsearchClusterRefID: ec.String("somerefid"),
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("main-apm"),
					Settings:                  &models.ApmSettings{},
					Plan: &models.ApmPlan{
						Apm: &models.ApmConfiguration{
							Version: "7.7.0",
						},
						ClusterTopology: []*models.ApmTopologyElement{{
							ZoneCount:               1,
							InstanceConfigurationID: "aws.apm.r4",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
						}},
					},
				},
				{
					ElasticsearchClusterRefID: ec.String("somerefid"),
					DisplayName:               "somename",
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("secondary-apm"),
					Settings:                  &models.ApmSettings{},
					Plan: &models.ApmPlan{
						Apm: &models.ApmConfiguration{
							Version: "7.6.0",
						},
						ClusterTopology: []*models.ApmTopologyElement{{
							ZoneCount:               1,
							InstanceConfigurationID: "aws.apm.r4",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(4096),
							},
						}},
					},
				},
				{
					ElasticsearchClusterRefID: ec.String("somerefid"),
					DisplayName:               "somename",
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("tertiary-apm"),
					Settings:                  &models.ApmSettings{},
					Plan: &models.ApmPlan{
						Apm: &models.ApmConfiguration{
							Version:     "7.8.0",
							DockerImage: "some-docker-image:version",
							SystemSettings: &models.ApmSystemSettings{
								DebugEnabled: ec.Bool(true),
							},
						},
						ClusterTopology: []*models.ApmTopologyElement{{
							Apm: &models.ApmConfiguration{
								DockerImage:              "some-other-image",
								UserSettingsYaml:         `some.setting: value`,
								UserSettingsOverrideYaml: `some.setting: value2`,
								UserSettingsJSON:         `{"some.setting": "value"}`,
								UserSettingsOverrideJSON: `{"some.setting": "value2"}`,
								SystemSettings: &models.ApmSystemSettings{
									DebugEnabled:          ec.Bool(true),
									ElasticsearchPassword: "somepass",
									ElasticsearchURL:      "someURL",
									ElasticsearchUsername: "someuser",
									KibanaURL:             "someKibanaURL",
									SecretToken:           "very_secret",
								},
							},
							ZoneCount:               1,
							InstanceConfigurationID: "aws.apm.r4",
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
			got, err := ExpandResources(tt.args.ess)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
