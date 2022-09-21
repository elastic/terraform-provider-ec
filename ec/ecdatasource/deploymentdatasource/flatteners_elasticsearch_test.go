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

package deploymentdatasource

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
)

func Test_flattenElasticsearchResources(t *testing.T) {
	type args struct {
		in []*models.ElasticsearchResourceInfo
	}
	tests := []struct {
		name string
		args args
		want []elasticsearchResourceInfoModelV0
		err  string
	}{
		{
			name: "empty resource list returns empty list",
			args: args{in: []*models.ElasticsearchResourceInfo{}},
			want: []elasticsearchResourceInfoModelV0{},
		},
		{
			name: "parses elasticsearch resource",
			args: args{in: []*models.ElasticsearchResourceInfo{
				{
					Region: ec.String("some-region"),
					RefID:  ec.String("main-elasticsearch"),
					Info: &models.ElasticsearchClusterInfo{
						Healthy:   ec.Bool(true),
						Status:    ec.String("started"),
						ClusterID: &mock.ValidClusterID,
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
									AutoscalingEnabled: ec.Bool(true),
									Elasticsearch: &models.ElasticsearchConfiguration{
										Version: "7.7.0",
									},
									ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
										{
											NodeCountPerZone:        1,
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
											AutoscalingMax: &models.TopologySize{
												Resource: ec.String("memory"),
												Value:    ec.Int32(15360),
											},
											AutoscalingMin: &models.TopologySize{
												Resource: ec.String("memory"),
												Value:    ec.Int32(1024),
											},
											AutoscalingPolicyOverrideJSON: map[string]interface{}{
												"proactive_storage": map[string]interface{}{
													"forecast_window": "3 h",
												},
											},
										},
										{
											NodeCountPerZone:        1,
											ZoneCount:               1,
											InstanceConfigurationID: "aws.coordinating.m5d",
											Size: &models.TopologySize{
												Resource: ec.String("memory"),
												Value:    ec.Int32(0),
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
			want: []elasticsearchResourceInfoModelV0{{
				Autoscale:     types.String{Value: "true"},
				RefID:         types.String{Value: "main-elasticsearch"},
				ResourceID:    types.String{Value: mock.ValidClusterID},
				Version:       types.String{Value: "7.7.0"},
				CloudID:       types.String{Value: "some CLOUD ID"},
				HttpEndpoint:  types.String{Value: "http://somecluster.cloud.elastic.co:9200"},
				HttpsEndpoint: types.String{Value: "https://somecluster.cloud.elastic.co:9243"},
				Healthy:       types.Bool{Value: true},
				Status:        types.String{Value: "started"},
				Topology: types.List{ElemType: types.ObjectType{AttrTypes: elasticsearchTopologyAttrTypes()},
					Elems: []attr.Value{types.Object{
						AttrTypes: elasticsearchTopologyAttrTypes(),
						Attrs: map[string]attr.Value{
							"instance_configuration_id": types.String{Value: "aws.data.highio.i3"},
							"size":                      types.String{Value: "2g"},
							"size_resource":             types.String{Value: "memory"},
							"node_type_data":            types.Bool{Value: true},
							"node_type_ingest":          types.Bool{Value: true},
							"node_type_master":          types.Bool{Value: true},
							"node_type_ml":              types.Bool{Value: false},
							"node_roles":                types.Set{ElemType: types.StringType, Elems: []attr.Value{}},
							"zone_count":                types.Int64{Value: 1},
							"autoscaling": types.List{ElemType: types.ObjectType{AttrTypes: elasticsearchAutoscalingAttrTypes()},
								Elems: []attr.Value{types.Object{
									AttrTypes: elasticsearchAutoscalingAttrTypes(),
									Attrs: map[string]attr.Value{
										"max_size":             types.String{Value: "15g"},
										"max_size_resource":    types.String{Value: "memory"},
										"min_size":             types.String{Value: "1g"},
										"min_size_resource":    types.String{Value: "memory"},
										"policy_override_json": types.String{Value: "{\"proactive_storage\":{\"forecast_window\":\"3 h\"}}"},
									}},
								},
							},
						}},
					},
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var model modelV0
			diags := flattenElasticsearchResources(context.Background(), tt.args.in, &model.Elasticsearch)
			assert.Empty(t, diags)

			var got []elasticsearchResourceInfoModelV0
			model.Elasticsearch.ElementsAs(context.Background(), &got, false)
			assert.Equal(t, tt.want, got)
		})
	}
}
