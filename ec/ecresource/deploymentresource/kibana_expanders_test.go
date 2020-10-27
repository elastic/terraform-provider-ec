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
	"errors"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/stretchr/testify/assert"
)

func Test_expandKibanaResources(t *testing.T) {
	tplPath := "testdata/aws-io-optimized-v2.json"
	tpl := func() *models.KibanaPayload {
		return kibanaResource(parseDeploymentTemplate(t,
			tplPath,
		))
	}
	type args struct {
		ess []interface{}
		tpl *models.KibanaPayload
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
			name: "parses a kibana resource with topology",
			args: args{
				tpl: tpl(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":                       "main-kibana",
						"resource_id":                  mock.ValidClusterID,
						"version":                      "7.7.0",
						"region":                       "some-region",
						"elasticsearch_cluster_ref_id": "somerefid",
						"topology": []interface{}{map[string]interface{}{
							"instance_configuration_id": "aws.kibana.r5d",
							"size":                      "2g",
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
					Plan: &models.KibanaClusterPlan{
						Kibana: &models.KibanaConfiguration{
							Version: "7.7.0",
						},
						ClusterTopology: []*models.KibanaClusterTopologyElement{
							{
								ZoneCount:               1,
								InstanceConfigurationID: "aws.kibana.r5d",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(2048),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "parses a kibana resource with incorrect instance_configuration_id",
			args: args{
				tpl: tpl(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":                       "main-kibana",
						"resource_id":                  mock.ValidClusterID,
						"version":                      "7.7.0",
						"region":                       "some-region",
						"elasticsearch_cluster_ref_id": "somerefid",
						"topology": []interface{}{map[string]interface{}{
							"instance_configuration_id": "gcp.some.config",
							"size":                      "2g",
							"zone_count":                1,
						}},
					},
				},
			},
			err: errors.New(`kibana topology: invalid instance_configuration_id: "gcp.some.config" doesn't match any of the deployment template instance configurations`),
		},
		{
			name: "parses a kibana resource without topology",
			args: args{
				tpl: tpl(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":                       "main-kibana",
						"resource_id":                  mock.ValidClusterID,
						"version":                      "7.7.0",
						"region":                       "some-region",
						"elasticsearch_cluster_ref_id": "somerefid",
					},
				},
			},
			want: []*models.KibanaPayload{
				{
					ElasticsearchClusterRefID: ec.String("somerefid"),
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("main-kibana"),
					Plan: &models.KibanaClusterPlan{
						Kibana: &models.KibanaConfiguration{
							Version: "7.7.0",
						},
						ClusterTopology: []*models.KibanaClusterTopologyElement{
							{
								ZoneCount:               1,
								InstanceConfigurationID: "aws.kibana.r5d",
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
		{
			name: "parses a kibana resource with a topology but no instance_configuration_id",
			args: args{
				tpl: tpl(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":                       "main-kibana",
						"resource_id":                  mock.ValidClusterID,
						"version":                      "7.7.0",
						"region":                       "some-region",
						"elasticsearch_cluster_ref_id": "somerefid",
						"topology": []interface{}{map[string]interface{}{
							"size": "4g",
						}},
					},
				},
			},
			want: []*models.KibanaPayload{
				{
					ElasticsearchClusterRefID: ec.String("somerefid"),
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("main-kibana"),
					Plan: &models.KibanaClusterPlan{
						Kibana: &models.KibanaConfiguration{
							Version: "7.7.0",
						},
						ClusterTopology: []*models.KibanaClusterTopologyElement{
							{
								ZoneCount:               1,
								InstanceConfigurationID: "aws.kibana.r5d",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(4096),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "parses a kibana resource with topology and settings",
			args: args{
				tpl: tpl(),
				ess: []interface{}{map[string]interface{}{
					"ref_id":                       "secondary-kibana",
					"elasticsearch_cluster_ref_id": "somerefid",
					"resource_id":                  mock.ValidClusterID,
					"version":                      "7.8.0",
					"region":                       "some-region",
					"config": []interface{}{map[string]interface{}{
						"user_settings_yaml":          "some.setting: value",
						"user_settings_override_yaml": "some.setting: override",
						"user_settings_json":          `{"some.setting":"value"}`,
						"user_settings_override_json": `{"some.setting":"override"}`,
					}},
					"topology": []interface{}{map[string]interface{}{
						"config": []interface{}{map[string]interface{}{
							"user_settings_yaml":          "some.setting: value",
							"user_settings_override_yaml": "some.setting: override",
							"user_settings_json":          `{"some.setting":"value"}`,
							"user_settings_override_json": `{"some.setting":"override"}`,
						}},
						"instance_configuration_id": "aws.kibana.r5d",
						"size":                      "4g",
						"zone_count":                1,
					}},
				}},
			},
			want: []*models.KibanaPayload{
				{
					ElasticsearchClusterRefID: ec.String("somerefid"),
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("secondary-kibana"),
					Plan: &models.KibanaClusterPlan{
						Kibana: &models.KibanaConfiguration{
							Version:                  "7.8.0",
							UserSettingsYaml:         "some.setting: value",
							UserSettingsOverrideYaml: "some.setting: override",
							UserSettingsJSON: map[string]interface{}{
								"some.setting": "value",
							},
							UserSettingsOverrideJSON: map[string]interface{}{
								"some.setting": "override",
							},
						},
						ClusterTopology: []*models.KibanaClusterTopologyElement{{
							Kibana: &models.KibanaConfiguration{
								UserSettingsYaml:         "some.setting: value",
								UserSettingsOverrideYaml: "some.setting: override",
								UserSettingsJSON: map[string]interface{}{
									"some.setting": "value",
								},
								UserSettingsOverrideJSON: map[string]interface{}{
									"some.setting": "override",
								},
							},
							ZoneCount:               1,
							InstanceConfigurationID: "aws.kibana.r5d",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(4096),
							},
						}},
					},
				},
			},
		},
		{
			name: "tries to parse an kibana resource when the template doesn't have a kibana instance set.",
			args: args{
				tpl: nil,
				ess: []interface{}{map[string]interface{}{
					"ref_id":                       "tertiary-kibana",
					"elasticsearch_cluster_ref_id": "somerefid",
					"resource_id":                  mock.ValidClusterID,
					"version":                      "7.8.0",
					"region":                       "some-region",
					"topology": []interface{}{map[string]interface{}{
						"instance_configuration_id": "aws.kibana.r5d",
						"size":                      "1g",
						"size_resource":             "memory",
						"zone_count":                1,
						"config": []interface{}{map[string]interface{}{
							"user_settings_yaml":          "some.setting: value",
							"user_settings_override_yaml": "some.setting: value2",
							"user_settings_json":          "{\"some.setting\": \"value\"}",
							"user_settings_override_json": "{\"some.setting\": \"value2\"}",
						}},
					}},
				}},
			},
			err: errors.New("kibana specified but deployment template does not allow it"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandKibanaResources(tt.args.ess, tt.args.tpl)
			if !assert.Equal(t, tt.err, err) {
				t.Error(err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
