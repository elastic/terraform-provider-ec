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
	tplPath := "testdata/template-aws-io-optimized-v2.json"
	tp770 := func() *models.ElasticsearchPayload {
		return enrichElasticsearchTemplate(
			esResource(parseDeploymentTemplate(t, tplPath)),
			"aws-io-optimized-v2",
			"7.7.0",
			false,
		)
	}

	create710 := func() *models.ElasticsearchPayload {
		return enrichElasticsearchTemplate(
			esResource(parseDeploymentTemplate(t, tplPath)),
			"aws-io-optimized-v2",
			"7.10.0",
			true,
		)
	}

	update711 := func() *models.ElasticsearchPayload {
		return enrichElasticsearchTemplate(
			esResource(parseDeploymentTemplate(t, tplPath)),
			"aws-io-optimized-v2",
			"7.11.0",
			true,
		)
	}

	hotWarmTplPath := "testdata/template-aws-hot-warm-v2.json"
	hotWarmTpl770 := func() *models.ElasticsearchPayload {
		return enrichElasticsearchTemplate(
			esResource(parseDeploymentTemplate(t, hotWarmTplPath)),
			"aws-io-optimized-v2",
			"7.7.0",
			false,
		)
	}

	hotWarm7111Tpl := func() *models.ElasticsearchPayload {
		return enrichElasticsearchTemplate(
			esResource(parseDeploymentTemplate(t, hotWarmTplPath)),
			"aws-io-optimized-v2",
			"7.11.1",
			true,
		)
	}

	eceDefaultTplPath := "testdata/template-ece-3.0.0-default.json"
	eceDefaultTpl := func() *models.ElasticsearchPayload {
		return enrichElasticsearchTemplate(
			esResource(parseDeploymentTemplate(t, eceDefaultTplPath)),
			"aws-io-optimized-v2",
			"7.17.3",
			true,
		)
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
				dt: tp770(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":      "main-elasticsearch",
						"resource_id": mock.ValidClusterID,
						"version":     "7.7.0",
						"region":      "some-region",
						"topology": []interface{}{map[string]interface{}{
							"id":         "hot_content",
							"size":       "2g",
							"zone_count": 1,
						}},
					},
				},
			},
			want: enrichWithEmptyTopologies(tp770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version: "7.7.0",
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-io-optimized-v2"),
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
							},
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parses an ES resource with empty version (7.10.0) in state uses node_roles from the DT",
			args: args{
				dt: create710(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":      "main-elasticsearch",
						"resource_id": mock.ValidClusterID,
						"region":      "some-region",
						"topology": []interface{}{map[string]interface{}{
							"id":         "hot_content",
							"size":       "2g",
							"zone_count": 1,
						}},
					},
				},
			},
			want: enrichWithEmptyTopologies(create710(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version: "7.10.0",
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-io-optimized-v2"),
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
							NodeRoles: []string{
								"master",
								"ingest",
								"remote_cluster_client",
								"data_hot",
								"transform",
								"data_content",
							},
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parses an ES resource with version 7.11.0 has node_roles coming from the saved state",
			args: args{
				dt: update711(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":      "main-elasticsearch",
						"resource_id": mock.ValidClusterID,
						"region":      "some-region",
						"topology": []interface{}{map[string]interface{}{
							"id":         "hot_content",
							"size":       "2g",
							"zone_count": 1,
							"node_roles": schema.NewSet(schema.HashString, []interface{}{
								"a", "b", "c",
							}),
						}},
					},
				},
			},
			want: enrichWithEmptyTopologies(update711(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version: "7.11.0",
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-io-optimized-v2"),
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
							NodeRoles: []string{
								"a", "b", "c",
							},
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parses an ES resource with invalid id",
			args: args{
				dt: tp770(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":      "main-elasticsearch",
						"resource_id": mock.ValidClusterID,
						"version":     "7.7.0",
						"region":      "some-region",
						"topology": []interface{}{map[string]interface{}{
							"id":         "invalid",
							"size":       "2g",
							"zone_count": 1,
						}},
					},
				},
			},
			err: errors.New(`elasticsearch topology invalid: invalid id: valid topology IDs are "coordinating", "hot_content", "warm", "cold", "master", "ml"`),
		},
		{
			name: "parses an ES resource without a topology",
			args: args{
				dt: tp770(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":      "main-elasticsearch",
						"resource_id": mock.ValidClusterID,
						"version":     "7.7.0",
						"region":      "some-region",
					},
				},
			},
			want: enrichWithEmptyTopologies(tp770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version: "7.7.0",
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-io-optimized-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID:                      "hot_content",
							InstanceConfigurationID: "aws.data.highio.i3",
							ZoneCount:               2,
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(8192),
							},
							NodeType: &models.ElasticsearchNodeType{
								Data:   ec.Bool(true),
								Ingest: ec.Bool(true),
								Master: ec.Bool(true),
							},
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parses an ES resource (HotWarm)",
			args: args{
				dt: hotWarmTpl770(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":                 "main-elasticsearch",
						"resource_id":            mock.ValidClusterID,
						"region":                 "some-region",
						"deployment_template_id": "aws-hot-warm-v2",
						"topology": []interface{}{
							map[string]interface{}{
								"id":         "hot_content",
								"size":       "2g",
								"zone_count": 1,
							},
							map[string]interface{}{
								"id":         "warm",
								"size":       "2g",
								"zone_count": 1,
							},
						},
					},
				},
			},
			want: enrichWithEmptyTopologies(hotWarmTpl770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
					Curation:                  nil,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version:  "7.7.0",
						Curation: nil,
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-hot-warm-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID: "hot_content",
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
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "warm",
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
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parses an ES resource with config (HotWarm)",
			args: args{
				dt: hotWarmTpl770(),
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
								"id":         "hot_content",
								"size":       "2g",
								"zone_count": 1,
							},
							map[string]interface{}{
								"id":         "warm",
								"size":       "2g",
								"zone_count": 1,
							},
						},
					},
				},
			},
			want: enrichWithEmptyTopologies(hotWarmTpl770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
					Curation:                  nil,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
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
							ID: "hot_content",
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
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "warm",
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
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parses an ES resource with explicit nils",
			args: args{
				dt: hotWarmTpl770(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":                 "main-elasticsearch",
						"resource_id":            mock.ValidClusterID,
						"version":                "7.7.0",
						"region":                 "some-region",
						"deployment_template_id": "aws-hot-warm-v2",
						"config": []interface{}{map[string]interface{}{
							"user_settings_yaml": nil,
						}},
						"topology": []interface{}{
							map[string]interface{}{
								"id":         "hot_content",
								"size":       nil,
								"zone_count": 1,
							},
							map[string]interface{}{
								"id":         "warm",
								"size":       "2g",
								"zone_count": nil,
							},
						},
					},
				},
			},
			want: enrichWithEmptyTopologies(hotWarmTpl770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
					Curation:                  nil,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version:          "7.7.0",
						Curation:         nil,
						UserSettingsYaml: "",
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-hot-warm-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID: "hot_content",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
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
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "warm",
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
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parses an ES resource without a topology (HotWarm)",
			args: args{
				dt: hotWarmTpl770(),
				ess: []interface{}{map[string]interface{}{
					"ref_id":      "main-elasticsearch",
					"resource_id": mock.ValidClusterID,
					"region":      "some-region",
				}},
			},
			want: enrichWithEmptyTopologies(hotWarmTpl770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
					Curation:                  nil,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version:  "7.7.0",
						Curation: nil,
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-hot-warm-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID: "hot_content",
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
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "warm",
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
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parses an ES resource with node type overrides (HotWarm)",
			args: args{
				dt: hotWarmTpl770(),
				ess: []interface{}{map[string]interface{}{
					"ref_id":      "main-elasticsearch",
					"resource_id": mock.ValidClusterID,
					"region":      "some-region",
					"topology": []interface{}{
						map[string]interface{}{
							"id":               "hot_content",
							"node_type_data":   "false",
							"node_type_master": "false",
							"node_type_ingest": "false",
							"node_type_ml":     "true",
						},
						map[string]interface{}{
							"id":               "warm",
							"node_type_master": "true",
						},
					},
				}},
			},
			want: enrichWithEmptyTopologies(hotWarmTpl770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
					Curation:                  nil,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version:  "7.7.0",
						Curation: nil,
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-hot-warm-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID: "hot_content",
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
								Data:   ec.Bool(false),
								Ingest: ec.Bool(false),
								Master: ec.Bool(false),
								Ml:     ec.Bool(true),
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "warm",
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
								Master: ec.Bool(true),
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "migrates old node_type state to new node_roles payload when the cold tier is set",
			args: args{
				dt: hotWarm7111Tpl(),
				ess: []interface{}{map[string]interface{}{
					"ref_id":      "main-elasticsearch",
					"resource_id": mock.ValidClusterID,
					"region":      "some-region",
					"topology": []interface{}{
						map[string]interface{}{
							"id":               "hot_content",
							"node_type_data":   "false",
							"node_type_master": "false",
							"node_type_ingest": "false",
							"node_type_ml":     "true",
						},
						map[string]interface{}{
							"id":               "warm",
							"node_type_master": "true",
						},
						map[string]interface{}{
							"id":   "cold",
							"size": "2g",
						},
					},
				}},
			},
			want: enrichWithEmptyTopologies(hotWarm7111Tpl(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
					Curation:                  nil,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version:  "7.11.1",
						Curation: nil,
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-hot-warm-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID: "hot_content",
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
							NodeRoles: []string{
								"master",
								"ingest",
								"remote_cluster_client",
								"data_hot",
								"transform",
								"data_content",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "warm",
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
							NodeRoles: []string{
								"data_warm",
								"remote_cluster_client",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "cold",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "cold",
								},
							},
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highstorage.d2",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							NodeRoles: []string{
								"data_cold",
								"remote_cluster_client",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(59392),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "autoscaling enabled",
			args: args{
				dt: hotWarm7111Tpl(),
				ess: []interface{}{map[string]interface{}{
					"autoscale":   "true",
					"ref_id":      "main-elasticsearch",
					"resource_id": mock.ValidClusterID,
					"region":      "some-region",
					"topology": []interface{}{
						map[string]interface{}{
							"id": "hot_content",
						},
						map[string]interface{}{
							"id": "warm",
						},
						map[string]interface{}{
							"id":   "cold",
							"size": "2g",
						},
					},
				}},
			},
			want: enrichWithEmptyTopologies(hotWarm7111Tpl(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
					Curation:                  nil,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(true),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version:  "7.11.1",
						Curation: nil,
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-hot-warm-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID: "hot_content",
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
							NodeRoles: []string{
								"master",
								"ingest",
								"remote_cluster_client",
								"data_hot",
								"transform",
								"data_content",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "warm",
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
							NodeRoles: []string{
								"data_warm",
								"remote_cluster_client",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "cold",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "cold",
								},
							},
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highstorage.d2",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							NodeRoles: []string{
								"data_cold",
								"remote_cluster_client",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(59392),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "autoscaling enabled overriding the size with ml",
			args: args{
				dt: hotWarm7111Tpl(),
				ess: []interface{}{map[string]interface{}{
					"autoscale":   "true",
					"ref_id":      "main-elasticsearch",
					"resource_id": mock.ValidClusterID,
					"region":      "some-region",
					"topology": []interface{}{
						map[string]interface{}{
							"id": "hot_content",
							"autoscaling": []interface{}{
								map[string]interface{}{
									"max_size": "58g",
								},
							},
						},
						map[string]interface{}{
							"id": "warm",
							"autoscaling": []interface{}{
								map[string]interface{}{
									"max_size": "29g",
								},
							},
						},
						map[string]interface{}{
							"id":   "cold",
							"size": "2g",
							"autoscaling": []interface{}{
								map[string]interface{}{
									"max_size": "29g",
								},
							},
						},
						map[string]interface{}{
							"id":   "ml",
							"size": "1g",
							"autoscaling": []interface{}{
								map[string]interface{}{
									"max_size": "29g",
									"min_size": "1g",
								},
							},
						},
					},
				}},
			},
			want: enrichWithEmptyTopologies(hotWarm7111Tpl(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
					Curation:                  nil,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(true),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version:  "7.11.1",
						Curation: nil,
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-hot-warm-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID: "hot_content",
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
							NodeRoles: []string{
								"master",
								"ingest",
								"remote_cluster_client",
								"data_hot",
								"transform",
								"data_content",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(59392),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "warm",
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
							NodeRoles: []string{
								"data_warm",
								"remote_cluster_client",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(29696),
								Resource: ec.String("memory"),
							},
						},
						{
							ID: "cold",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "cold",
								},
							},
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highstorage.d2",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							NodeRoles: []string{
								"data_cold",
								"remote_cluster_client",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(29696),
								Resource: ec.String("memory"),
							},
						},
						{
							ID:                      "ml",
							ZoneCount:               1,
							InstanceConfigurationID: "aws.ml.m5d",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(1024),
							},
							NodeRoles: []string{
								"ml",
								"remote_cluster_client",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(29696),
								Resource: ec.String("memory"),
							},
							AutoscalingMin: &models.TopologySize{
								Value:    ec.Int32(1024),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "autoscaling enabled no dimension in template, default resource",
			args: args{
				dt: eceDefaultTpl(),
				ess: []interface{}{map[string]interface{}{
					"autoscale":   "true",
					"ref_id":      "main-elasticsearch",
					"resource_id": mock.ValidClusterID,
					"region":      "some-region",
					"topology": []interface{}{
						map[string]interface{}{
							"id": "hot_content",
							"autoscaling": []interface{}{
								map[string]interface{}{
									"max_size": "450g",
									"min_size": "2g",
								},
							},
						},
						map[string]interface{}{
							"id": "master",
							"autoscaling": []interface{}{
								map[string]interface{}{
									"max_size": "250g",
									"min_size": "1g",
								},
							},
						},
					},
				}},
			},
			want: enrichWithEmptyTopologies(eceDefaultTpl(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
					Curation:                  nil,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(true),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version:  "7.17.3",
						Curation: nil,
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-io-optimized-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID: "hot_content",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							ZoneCount:               1,
							InstanceConfigurationID: "data.default",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(4096),
							},
							NodeRoles: []string{
								"master",
								"ingest",
								"data_hot",
								"data_content",
								"remote_cluster_client",
								"transform",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(460800),
								Resource: ec.String("memory"),
							},
							AutoscalingMin: &models.TopologySize{
								Value:    ec.Int32(2048),
								Resource: ec.String("memory"),
							},
						},
						{
							ID:                      "master",
							ZoneCount:               1,
							InstanceConfigurationID: "master",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(0),
							},
							NodeRoles: []string{
								"master",
								"remote_cluster_client",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(256000),
								Resource: ec.String("memory"),
							},
							AutoscalingMin: &models.TopologySize{
								Value:    ec.Int32(1024),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "autoscaling enabled overriding the size and resources",
			args: args{
				dt: hotWarm7111Tpl(),
				ess: []interface{}{map[string]interface{}{
					"autoscale":   "true",
					"ref_id":      "main-elasticsearch",
					"resource_id": mock.ValidClusterID,
					"region":      "some-region",
					"topology": []interface{}{
						map[string]interface{}{
							"id": "hot_content",
							"autoscaling": []interface{}{
								map[string]interface{}{
									"max_size_resource": "storage",
									"max_size":          "450g",
								},
							},
						},
						map[string]interface{}{
							"id": "warm",
							"autoscaling": []interface{}{
								map[string]interface{}{
									"max_size_resource": "storage",
									"max_size":          "870g",
								},
							},
						},
						map[string]interface{}{
							"id":   "cold",
							"size": "4g",
							"autoscaling": []interface{}{
								map[string]interface{}{
									"max_size_resource": "storage",
									"max_size":          "1740g",

									"min_size_resource": "storage",
									"min_size":          "4g",
								},
							},
						},
					},
				}},
			},
			want: enrichWithEmptyTopologies(hotWarm7111Tpl(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
					Curation:                  nil,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(true),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version:  "7.11.1",
						Curation: nil,
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-hot-warm-v2"),
					},
					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
						{
							ID: "hot_content",
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
							NodeRoles: []string{
								"master",
								"ingest",
								"remote_cluster_client",
								"data_hot",
								"transform",
								"data_content",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(460800),
								Resource: ec.String("storage"),
							},
						},
						{
							ID: "warm",
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
							NodeRoles: []string{
								"data_warm",
								"remote_cluster_client",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(890880),
								Resource: ec.String("storage"),
							},
						},
						{
							ID: "cold",
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "cold",
								},
							},
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highstorage.d2",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(4096),
							},
							NodeRoles: []string{
								"data_cold",
								"remote_cluster_client",
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(0),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(1781760),
								Resource: ec.String("storage"),
							},
							AutoscalingMin: &models.TopologySize{
								Value:    ec.Int32(4096),
								Resource: ec.String("storage"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parses an ES resource with plugins",
			args: args{
				dt: tp770(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":      "main-elasticsearch",
						"resource_id": mock.ValidClusterID,
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
							"id":         "hot_content",
							"size":       "2g",
							"zone_count": 1,
						}},
					},
				},
			},
			want: enrichWithEmptyTopologies(tp770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
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
							ID:                      "hot_content",
							ZoneCount:               1,
							InstanceConfigurationID: "aws.data.highio.i3",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{"data": "hot"},
							},
							NodeType: &models.ElasticsearchNodeType{
								Data:   ec.Bool(true),
								Ingest: ec.Bool(true),
								Master: ec.Bool(true),
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parses an ES resource with snapshot settings",
			args: args{
				dt: tp770(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":      "main-elasticsearch",
						"resource_id": mock.ValidClusterID,
						"region":      "some-region",
						"snapshot_source": []interface{}{map[string]interface{}{
							"snapshot_name":                   "__latest_success__",
							"source_elasticsearch_cluster_id": mock.ValidClusterID,
						}},
						"topology": []interface{}{map[string]interface{}{
							"id":         "hot_content",
							"size":       "2g",
							"zone_count": 1,
						}},
					},
				},
			},
			want: enrichWithEmptyTopologies(tp770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
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
							},
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{"data": "hot"},
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
				},
			}),
		},
		{
			name: "parse autodetect configuration strategy",
			args: args{
				dt: tp770(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":      "main-elasticsearch",
						"resource_id": mock.ValidClusterID,
						"version":     "7.7.0",
						"region":      "some-region",
						"topology": []interface{}{map[string]interface{}{
							"id":         "hot_content",
							"size":       "2g",
							"zone_count": 1,
						}},
						"strategy": []interface{}{map[string]interface{}{
							"type": "autodetect",
						}},
					},
				},
			},

			want: enrichWithEmptyTopologies(tp770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version: "7.7.0",
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-io-optimized-v2"),
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
							},
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
					Transient: &models.TransientElasticsearchPlanConfiguration{
						Strategy: &models.PlanStrategy{
							Autodetect: new(models.AutodetectStrategyConfig),
						},
					},
				},
			}),
		},
		{
			name: "parse grow_and_shrink configuration strategy",
			args: args{
				dt: tp770(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":      "main-elasticsearch",
						"resource_id": mock.ValidClusterID,
						"version":     "7.7.0",
						"region":      "some-region",
						"topology": []interface{}{map[string]interface{}{
							"id":         "hot_content",
							"size":       "2g",
							"zone_count": 1,
						}},
						"strategy": []interface{}{map[string]interface{}{
							"type": "grow_and_shrink",
						}},
					},
				},
			},

			want: enrichWithEmptyTopologies(tp770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version: "7.7.0",
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-io-optimized-v2"),
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
							},
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
					Transient: &models.TransientElasticsearchPlanConfiguration{
						Strategy: &models.PlanStrategy{
							GrowAndShrink: new(models.GrowShrinkStrategyConfig),
						},
					},
				},
			}),
		},
		{
			name: "parse rolling_grow_and_shrink configuration strategy",
			args: args{
				dt: tp770(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":      "main-elasticsearch",
						"resource_id": mock.ValidClusterID,
						"version":     "7.7.0",
						"region":      "some-region",
						"topology": []interface{}{map[string]interface{}{
							"id":         "hot_content",
							"size":       "2g",
							"zone_count": 1,
						}},
						"strategy": []interface{}{map[string]interface{}{
							"type": "rolling_grow_and_shrink",
						}},
					},
				},
			},

			want: enrichWithEmptyTopologies(tp770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version: "7.7.0",
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-io-optimized-v2"),
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
							},
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
							},
						},
					},
					Transient: &models.TransientElasticsearchPlanConfiguration{
						Strategy: &models.PlanStrategy{
							RollingGrowAndShrink: new(models.RollingGrowShrinkStrategyConfig),
						},
					},
				},
			}),
		},
		{
			name: "parse rolling configuration strategy",
			args: args{
				dt: tp770(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":      "main-elasticsearch",
						"resource_id": mock.ValidClusterID,
						"version":     "7.7.0",
						"region":      "some-region",
						"topology": []interface{}{map[string]interface{}{
							"id":         "hot_content",
							"size":       "2g",
							"zone_count": 1,
						}},
						"strategy": []interface{}{map[string]interface{}{
							"type": "rolling_all",
						}},
					},
				},
			},

			want: enrichWithEmptyTopologies(tp770(), &models.ElasticsearchPayload{
				Region: ec.String("some-region"),
				RefID:  ec.String("main-elasticsearch"),
				Settings: &models.ElasticsearchClusterSettings{
					DedicatedMastersThreshold: 6,
				},
				Plan: &models.ElasticsearchClusterPlan{
					AutoscalingEnabled: ec.Bool(false),
					Elasticsearch: &models.ElasticsearchConfiguration{
						Version: "7.7.0",
					},
					DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("aws-io-optimized-v2"),
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
							},
							Elasticsearch: &models.ElasticsearchConfiguration{
								NodeAttributes: map[string]string{
									"data": "hot",
								},
							},
							TopologyElementControl: &models.TopologyElementControl{
								Min: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							},
							AutoscalingMax: &models.TopologySize{
								Value:    ec.Int32(118784),
								Resource: ec.String("memory"),
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
			}),
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
