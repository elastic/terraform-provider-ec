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

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
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
											// NodeRoles cannot be used simultaneously with NodeType
											// but let's have it here for testing purposes
											NodeRoles: []string{"data_content", "data_hot"},
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
				Autoscale:     types.StringValue("true"),
				RefID:         types.StringValue("main-elasticsearch"),
				ResourceID:    types.StringValue(mock.ValidClusterID),
				Version:       types.StringValue("7.7.0"),
				CloudID:       types.StringValue("some CLOUD ID"),
				HttpEndpoint:  types.StringValue("http://somecluster.cloud.elastic.co:9200"),
				HttpsEndpoint: types.StringValue("https://somecluster.cloud.elastic.co:9243"),
				Healthy:       types.BoolValue(true),
				Status:        types.StringValue("started"),
				Topology: func() types.List {
					nodeRoles, diags := types.SetValueFrom(
						context.Background(),
						types.StringType,
						[]string{"data_content", "data_hot"},
					)
					assert.Nil(t, diags)

					autoscalingList, diags := types.ListValueFrom(
						context.Background(),
						elasticsearchAutoscalingElemType(),
						[]elasticsearchAutoscalingModel{
							{
								MaxSize:            types.StringValue("15g"),
								MaxSizeResource:    types.StringValue("memory"),
								MinSize:            types.StringValue("1g"),
								MinSizeResource:    types.StringValue("memory"),
								PolicyOverrideJson: types.StringValue("{\"proactive_storage\":{\"forecast_window\":\"3 h\"}}"),
							},
						},
					)
					assert.Nil(t, diags)

					res, diags := types.ListValueFrom(
						context.Background(),
						types.ObjectType{AttrTypes: elasticsearchTopologyAttrTypes()},
						[]elasticsearchTopologyModelV0{
							{
								InstanceConfigurationID: types.StringValue("aws.data.highio.i3"),
								Size:                    types.StringValue("2g"),
								SizeResource:            types.StringValue("memory"),
								NodeTypeData:            types.BoolValue(true),
								NodeTypeIngest:          types.BoolValue(true),
								NodeTypeMaster:          types.BoolValue(true),
								NodeTypeMl:              types.BoolValue(false),
								NodeRoles:               nodeRoles,
								ZoneCount:               types.Int64Value(1),
								Autoscaling:             autoscalingList,
							},
						},
					)
					assert.Nil(t, diags)

					return res
				}(),
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elasticsearch, diags := flattenElasticsearchResources(context.Background(), tt.args.in)
			assert.Empty(t, diags)
			var got []elasticsearchResourceInfoModelV0
			elasticsearch.ElementsAs(context.Background(), &got, false)
			assert.Equal(t, tt.want, got)
			util.CheckConverionToAttrValue(t, &DataSource{}, "elasticsearch", elasticsearch)
		})
	}
}
