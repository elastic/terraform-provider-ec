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

package elasticsearchstate

import (
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestExpandResource(t *testing.T) {
	type args struct {
		ess []interface{}
		dt  string
	}
	tests := []struct {
		name string
		args args
		want []*models.ElasticsearchPayload
		err  error
	}{
		{
			name: "returns nil when there's no resources",
		},
		{
			name: "parses an ES resource with monitoring",
			args: args{
				dt: "deployment-template-id",
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":      "secondary-elasticsearch",
						"resource_id": mock.ValidClusterID,
						"version":     "7.6.0",
						"region":      "some-region",
						"monitoring_settings": []interface{}{
							map[string]interface{}{"target_cluster_id": "some"},
						},
						"topology": []interface{}{
							map[string]interface{}{
								"instance_configuration_id": "aws.data.highio.i3",
								"memory_per_node":           "4g",
								"node_type_data":            true,
								"node_type_ingest":          true,
								"node_type_master":          true,
								"node_type_ml":              false,
								"zone_count":                1,
							},
						},
					},
				},
			},
			want: []*models.ElasticsearchPayload{
				{
					Region: ec.String("some-region"),
					RefID:  ec.String("secondary-elasticsearch"),
					Settings: &models.ElasticsearchClusterSettings{
						Monitoring: &models.ManagedMonitoringSettings{
							TargetClusterID: ec.String("some"),
						},
					},
					Plan: &models.ElasticsearchClusterPlan{
						Elasticsearch: &models.ElasticsearchConfiguration{
							Version: "7.6.0",
						},
						DeploymentTemplate: &models.DeploymentTemplateReference{
							ID: ec.String("deployment-template-id"),
						},
						ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
							{
								ZoneCount:               1,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(4096),
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
		{
			name: "parses an ES resource without monitoring",
			args: args{
				dt: "deployment-template-id",
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":      "main-elasticsearch",
						"resource_id": mock.ValidClusterID,
						"version":     "7.7.0",
						"region":      "some-region",
						"topology": []interface{}{
							map[string]interface{}{
								"instance_configuration_id": "aws.data.highio.i3",
								"memory_per_node":           "2g",
								"node_type_data":            true,
								"node_type_ingest":          true,
								"node_type_master":          true,
								"node_type_ml":              false,
								"zone_count":                1,
							},
						},
					},
				},
			},
			want: []*models.ElasticsearchPayload{
				{
					Region:   ec.String("some-region"),
					RefID:    ec.String("main-elasticsearch"),
					Settings: &models.ElasticsearchClusterSettings{},
					Plan: &models.ElasticsearchClusterPlan{
						Elasticsearch: &models.ElasticsearchConfiguration{
							Version: "7.7.0",
						},
						DeploymentTemplate: &models.DeploymentTemplateReference{
							ID: ec.String("deployment-template-id"),
						},
						ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
							{
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
		{
			name: "parses an ES resource with config",
			args: args{
				dt: "deployment-template-id",
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":      "main-elasticsearch",
						"resource_id": mock.ValidClusterID,
						"version":     "7.7.0",
						"region":      "some-region",
						"topology": []interface{}{map[string]interface{}{
							"instance_configuration_id": "aws.data.highio.i3",
							"memory_per_node":           "2g",
							"node_type_data":            true,
							"node_type_ingest":          true,
							"node_type_master":          true,
							"node_type_ml":              false,
							"zone_count":                1,
							"config": []interface{}{map[string]interface{}{
								"user_settings_yaml":          "some.setting: value",
								"user_settings_override_yaml": "some.setting: value2",
								"user_settings_json":          "{\"some.setting\": \"value\"}",
								"user_settings_override_json": "{\"some.setting\": \"value2\"}",
								"plugins": schema.NewSet(schema.HashString, []interface{}{
									"plugin",
								}),
							}},
						}},
					},
				},
			},
			want: []*models.ElasticsearchPayload{
				{
					Region:   ec.String("some-region"),
					RefID:    ec.String("main-elasticsearch"),
					Settings: &models.ElasticsearchClusterSettings{},
					Plan: &models.ElasticsearchClusterPlan{
						Elasticsearch: &models.ElasticsearchConfiguration{
							Version: "7.7.0",
						},
						DeploymentTemplate: &models.DeploymentTemplateReference{
							ID: ec.String("deployment-template-id"),
						},
						ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
							{
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
								Elasticsearch: &models.ElasticsearchConfiguration{
									UserSettingsYaml:         `some.setting: value`,
									UserSettingsOverrideYaml: `some.setting: value2`,
									UserSettingsJSON:         `{"some.setting": "value"}`,
									UserSettingsOverrideJSON: `{"some.setting": "value2"}`,
									EnabledBuiltInPlugins:    []string{"plugin"},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandResources(tt.args.ess, tt.args.dt)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
