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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func Test_expandEsResource(t *testing.T) {
	tplPath := "testdata/aws-io-optimized-v2.json"
	tpl := func() *models.ElasticsearchPayload {
		return esResource(parseDeploymentTemplate(t,
			tplPath,
		))
	}

	hotWarmTplPath := "testdata/aws-hot-warm-v2.json"
	hotWarmTpl := func() *models.ElasticsearchPayload {
		return esResource(parseDeploymentTemplate(t,
			hotWarmTplPath,
		))
	}

	type args struct {
		ess []interface{}
		dt  *models.ElasticsearchPayload
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
			name: "parses an ES resource",
			args: args{
				dt: tpl(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":      "main-elasticsearch",
						"resource_id": mock.ValidClusterID,
						"version":     "7.7.0",
						"region":      "some-region",
						"topology": []interface{}{map[string]interface{}{
							"instance_configuration_id": "aws.data.highio.i3",
							"size":                      "2g",
							"zone_count":                1,
						}},
					},
				},
			},
			want: []*models.ElasticsearchPayload{
				{
					Region: ec.String("some-region"),
					RefID:  ec.String("main-elasticsearch"),
					Settings: &models.ElasticsearchClusterSettings{
						DedicatedMastersThreshold: 6,
					},
					Plan: &models.ElasticsearchClusterPlan{
						Elasticsearch: &models.ElasticsearchConfiguration{
							Version: "7.7.0",
						},
						DeploymentTemplate: &models.DeploymentTemplateReference{
							ID: ec.String("aws-io-optimized-v2"),
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
								},
							},
						},
					},
				},
			},
		},
		{
			name: "parses an ES resource with invalid instance_configuration_id",
			args: args{
				dt: tpl(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":      "main-elasticsearch",
						"resource_id": mock.ValidClusterID,
						"version":     "7.7.0",
						"region":      "some-region",
						"topology": []interface{}{map[string]interface{}{
							"instance_configuration_id": "gcp.some.config",
							"size":                      "2g",
							"zone_count":                1,
						}},
					},
				},
			},
			err: errors.New(`elasticsearch topology: invalid instance_configuration_id: "gcp.some.config" doesn't match any of the deployment template instance configurations`),
		},
		{
			name: "parses an ES resource without a topology",
			args: args{
				dt: tpl(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":      "main-elasticsearch",
						"resource_id": mock.ValidClusterID,
						"version":     "7.7.0",
						"region":      "some-region",
					},
				},
			},
			want: []*models.ElasticsearchPayload{
				{
					Region: ec.String("some-region"),
					RefID:  ec.String("main-elasticsearch"),
					Settings: &models.ElasticsearchClusterSettings{
						DedicatedMastersThreshold: 6,
					},
					Plan: &models.ElasticsearchClusterPlan{
						Elasticsearch: &models.ElasticsearchConfiguration{
							Version: "7.7.0",
						},
						DeploymentTemplate: &models.DeploymentTemplateReference{
							ID: ec.String("aws-io-optimized-v2"),
						},
						ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
							{
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(8192),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   ec.Bool(true),
									Ingest: ec.Bool(true),
									Master: ec.Bool(true),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "parses an ES resource with topology but no instance_configuration_id",
			args: args{
				dt: tpl(),
				ess: []interface{}{map[string]interface{}{
					"ref_id":      "main-elasticsearch",
					"resource_id": mock.ValidClusterID,
					"version":     "7.7.0",
					"region":      "some-region",
					"topology": []interface{}{map[string]interface{}{
						"size": "1g",
					}},
				}},
			},
			want: []*models.ElasticsearchPayload{
				{
					Region: ec.String("some-region"),
					RefID:  ec.String("main-elasticsearch"),
					Settings: &models.ElasticsearchClusterSettings{
						DedicatedMastersThreshold: 6,
					},
					Plan: &models.ElasticsearchClusterPlan{
						Elasticsearch: &models.ElasticsearchConfiguration{
							Version: "7.7.0",
						},
						DeploymentTemplate: &models.DeploymentTemplateReference{
							ID: ec.String("aws-io-optimized-v2"),
						},
						ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
							{
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   ec.Bool(true),
									Ingest: ec.Bool(true),
									Master: ec.Bool(true),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "parses an ES resource (HotWarm)",
			args: args{
				dt: hotWarmTpl(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":                 "main-elasticsearch",
						"resource_id":            mock.ValidClusterID,
						"version":                "7.7.0",
						"region":                 "some-region",
						"deployment_template_id": "aws-hot-warm-v2",
						"topology": []interface{}{
							map[string]interface{}{
								"instance_configuration_id": "aws.data.highio.i3",
								"size":                      "2g",
								"zone_count":                1,
							},
							map[string]interface{}{
								"instance_configuration_id": "aws.data.highstorage.d2",
								"size":                      "2g",
								"zone_count":                1,
							},
						},
					},
				},
			},
			want: []*models.ElasticsearchPayload{
				{
					Region: ec.String("some-region"),
					RefID:  ec.String("main-elasticsearch"),
					Settings: &models.ElasticsearchClusterSettings{
						DedicatedMastersThreshold: 6,
						Curation:                  nil,
					},
					Plan: &models.ElasticsearchClusterPlan{
						Elasticsearch: &models.ElasticsearchConfiguration{
							Version:  "7.7.0",
							Curation: nil,
						},
						DeploymentTemplate: &models.DeploymentTemplateReference{
							ID: ec.String("aws-hot-warm-v2"),
						},
						ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
							{
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
							},
							{
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
							},
						},
					},
				},
			},
		},
		{
			name: "parses an ES resource with config (HotWarm)",
			args: args{
				dt: hotWarmTpl(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":                 "main-elasticsearch",
						"resource_id":            mock.ValidClusterID,
						"version":                "7.7.0",
						"region":                 "some-region",
						"deployment_template_id": "aws-hot-warm-v2",
						"config": []interface{}{map[string]interface{}{
							"user_settings_yaml": "somesetting: true",
						}},
						"topology": []interface{}{
							map[string]interface{}{
								"instance_configuration_id": "aws.data.highio.i3",
								"size":                      "2g",
								"zone_count":                1,
							},
							map[string]interface{}{
								"instance_configuration_id": "aws.data.highstorage.d2",
								"size":                      "2g",
								"zone_count":                1,
							},
						},
					},
				},
			},
			want: []*models.ElasticsearchPayload{
				{
					Region: ec.String("some-region"),
					RefID:  ec.String("main-elasticsearch"),
					Settings: &models.ElasticsearchClusterSettings{
						DedicatedMastersThreshold: 6,
						Curation:                  nil,
					},
					Plan: &models.ElasticsearchClusterPlan{
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
							},
							{
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
							},
						},
					},
				},
			},
		},
		{
			name: "parses an ES resource without a topology (HotWarm)",
			args: args{
				dt: hotWarmTpl(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":      "main-elasticsearch",
						"resource_id": mock.ValidClusterID,
						"version":     "7.7.0",
						"region":      "some-region",
					},
				},
			},
			want: []*models.ElasticsearchPayload{
				{
					Region: ec.String("some-region"),
					RefID:  ec.String("main-elasticsearch"),
					Settings: &models.ElasticsearchClusterSettings{
						DedicatedMastersThreshold: 6,
						Curation:                  nil,
					},
					Plan: &models.ElasticsearchClusterPlan{
						Elasticsearch: &models.ElasticsearchConfiguration{
							Version:  "7.7.0",
							Curation: nil,
						},
						DeploymentTemplate: &models.DeploymentTemplateReference{
							ID: ec.String("aws-hot-warm-v2"),
						},
						ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
							{
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
							},
							{
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
							},
						},
					},
				},
			},
		},
		{
			name: "parses an ES resource with a topology but no instance_configuration_id (HotWarm)",
			args: args{
				dt: hotWarmTpl(),
				ess: []interface{}{map[string]interface{}{
					"ref_id":      "main-elasticsearch",
					"resource_id": mock.ValidClusterID,
					"version":     "7.7.0",
					"region":      "some-region",
					"topology": []interface{}{
						map[string]interface{}{
							"size": "2g",
						},
						map[string]interface{}{
							"size": "2g",
						},
					},
				}},
			},
			want: []*models.ElasticsearchPayload{
				{
					Region: ec.String("some-region"),
					RefID:  ec.String("main-elasticsearch"),
					Settings: &models.ElasticsearchClusterSettings{
						DedicatedMastersThreshold: 6,
						Curation:                  nil,
					},
					Plan: &models.ElasticsearchClusterPlan{
						Elasticsearch: &models.ElasticsearchConfiguration{
							Version:  "7.7.0",
							Curation: nil,
						},
						DeploymentTemplate: &models.DeploymentTemplateReference{
							ID: ec.String("aws-hot-warm-v2"),
						},
						ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
							{
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{
										"data": "hot",
									},
								},
								ZoneCount:               2,
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
							},
							{
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{
										"data": "warm",
									},
								},
								ZoneCount:               2,
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
							},
						},
					},
				},
			},
		},
		{
			name: "parses an ES resource with config",
			args: args{
				dt: tpl(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":      "main-elasticsearch",
						"resource_id": mock.ValidClusterID,
						"version":     "7.7.0",
						"region":      "some-region",
						"config": []interface{}{map[string]interface{}{
							"user_settings_yaml":          "some.setting: value",
							"user_settings_override_yaml": "some.setting: value2",
							"user_settings_json":          "{\"some.setting\":\"value\"}",
							"user_settings_override_json": "{\"some.setting\":\"value2\"}",
							"plugins": schema.NewSet(schema.HashString, []interface{}{
								"plugin",
							}),
						}},
						"topology": []interface{}{map[string]interface{}{
							"instance_configuration_id": "aws.data.highio.i3",
							"size":                      "2g",
							"zone_count":                1,
						}},
					},
				},
			},
			want: []*models.ElasticsearchPayload{
				{
					Region: ec.String("some-region"),
					RefID:  ec.String("main-elasticsearch"),
					Settings: &models.ElasticsearchClusterSettings{
						DedicatedMastersThreshold: 6,
					},
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
							EnabledBuiltInPlugins: []string{"plugin"},
						},
						DeploymentTemplate: &models.DeploymentTemplateReference{
							ID: ec.String("aws-io-optimized-v2"),
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
								},
							},
						},
					},
				},
			},
		},
		{
			name: "parses an ES resource with snapshot settings",
			args: args{
				dt: tpl(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":      "main-elasticsearch",
						"resource_id": mock.ValidClusterID,
						"version":     "7.7.0",
						"region":      "some-region",
						"snapshot_source": []interface{}{map[string]interface{}{
							"snapshot_name":                   "__latest_success__",
							"source_elasticsearch_cluster_id": mock.ValidClusterID,
						}},
						"topology": []interface{}{map[string]interface{}{
							"instance_configuration_id": "aws.data.highio.i3",
							"size":                      "2g",
							"zone_count":                1,
						}},
					},
				},
			},
			want: []*models.ElasticsearchPayload{
				{
					Region: ec.String("some-region"),
					RefID:  ec.String("main-elasticsearch"),
					Settings: &models.ElasticsearchClusterSettings{
						DedicatedMastersThreshold: 6,
					},
					Plan: &models.ElasticsearchClusterPlan{
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
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandEsResources(tt.args.ess, tt.args.dt)
			if err != nil {
				var msg string
				if tt.err != nil {
					msg = tt.err.Error()
				}
				assert.EqualError(t, err, msg)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
