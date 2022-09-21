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

func Test_flattenApmResource(t *testing.T) {
	type args struct {
		in []*models.ApmResourceInfo
	}
	tests := []struct {
		name string
		args args
		want []apmResourceInfoModelV0
	}{
		{
			name: "empty resource list returns empty list",
			args: args{in: []*models.ApmResourceInfo{}},
			want: []apmResourceInfoModelV0{},
		},
		{
			name: "parses the apm resource",
			args: args{in: []*models.ApmResourceInfo{
				{
					RefID:                     ec.String("main-apm"),
					ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
					Info: &models.ApmInfo{
						Healthy: ec.Bool(true),
						Status:  ec.String("started"),
						ID:      &mock.ValidClusterID,
						Metadata: &models.ClusterMetadataInfo{
							Endpoint: "apmresource.cloud.elastic.co",
							Ports: &models.ClusterMetadataPortInfo{
								HTTP:  ec.Int32(9200),
								HTTPS: ec.Int32(9243),
							},
						},
						PlanInfo: &models.ApmPlansInfo{Current: &models.ApmPlanInfo{
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{
									Version: "7.7.0",
								},
								ClusterTopology: []*models.ApmTopologyElement{
									{
										ZoneCount:               1,
										InstanceConfigurationID: "aws.apm.r4",
										Size: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									{
										ZoneCount:               1,
										InstanceConfigurationID: "aws.apm.m5d",
										Size: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
								},
							},
						}},
					},
				},
			}},
			want: []apmResourceInfoModelV0{{
				ElasticsearchClusterRefID: types.String{Value: "main-elasticsearch"},
				RefID:                     types.String{Value: "main-apm"},
				ResourceID:                types.String{Value: mock.ValidClusterID},
				Version:                   types.String{Value: "7.7.0"},
				HttpEndpoint:              types.String{Value: "http://apmresource.cloud.elastic.co:9200"},
				HttpsEndpoint:             types.String{Value: "https://apmresource.cloud.elastic.co:9243"},
				Healthy:                   types.Bool{Value: true},
				Status:                    types.String{Value: "started"},
				Topology: types.List{ElemType: types.ObjectType{AttrTypes: apmTopologyAttrTypes()},
					Elems: []attr.Value{types.Object{
						AttrTypes: apmTopologyAttrTypes(),
						Attrs: map[string]attr.Value{
							"instance_configuration_id": types.String{Value: "aws.apm.r4"},
							"size":                      types.String{Value: "1g"},
							"size_resource":             types.String{Value: "memory"},
							"zone_count":                types.Int64{Value: 1},
						},
					}},
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var newState modelV0
			diags := flattenApmResources(context.Background(), tt.args.in, &newState.Apm)
			assert.Empty(t, diags)

			var got []apmResourceInfoModelV0
			newState.Apm.ElementsAs(context.Background(), &got, false)
			assert.Equal(t, tt.want, got)
		})
	}
}
