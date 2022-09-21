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

func Test_flattenKibanaResources(t *testing.T) {
	type args struct {
		in []*models.KibanaResourceInfo
	}
	tests := []struct {
		name string
		args args
		want []kibanaResourceInfoModelV0
	}{
		{
			name: "empty resource list returns empty list",
			args: args{in: []*models.KibanaResourceInfo{}},
			want: []kibanaResourceInfoModelV0{},
		},
		{
			name: "parses the kibana resource",
			args: args{in: []*models.KibanaResourceInfo{
				{
					RefID:                     ec.String("main-kibana"),
					ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
					Info: &models.KibanaClusterInfo{
						Healthy:   ec.Bool(true),
						Status:    ec.String("started"),
						ClusterID: &mock.ValidClusterID,
						Metadata: &models.ClusterMetadataInfo{
							Endpoint: "kibanaresource.cloud.elastic.co",
							Ports: &models.ClusterMetadataPortInfo{
								HTTP:  ec.Int32(9200),
								HTTPS: ec.Int32(9243),
							},
						},
						PlanInfo: &models.KibanaClusterPlansInfo{
							Current: &models.KibanaClusterPlanInfo{
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
												Value:    ec.Int32(1024),
											},
										},
										{
											ZoneCount:               1,
											InstanceConfigurationID: "aws.kibana.m5d",
											Size: &models.TopologySize{
												Resource: ec.String("memory"),
												Value:    ec.Int32(0),
											},
										},
									},
								},
							},
						},
					},
				},
			}},
			want: []kibanaResourceInfoModelV0{{
				ElasticsearchClusterRefID: types.String{Value: "main-elasticsearch"},
				RefID:                     types.String{Value: "main-kibana"},
				ResourceID:                types.String{Value: mock.ValidClusterID},
				Version:                   types.String{Value: "7.7.0"},
				HttpEndpoint:              types.String{Value: "http://kibanaresource.cloud.elastic.co:9200"},
				HttpsEndpoint:             types.String{Value: "https://kibanaresource.cloud.elastic.co:9243"},
				Healthy:                   types.Bool{Value: true},
				Status:                    types.String{Value: "started"},
				Topology: types.List{ElemType: types.ObjectType{AttrTypes: kibanaTopologyAttrTypes()},
					Elems: []attr.Value{types.Object{
						AttrTypes: kibanaTopologyAttrTypes(),
						Attrs: map[string]attr.Value{
							"instance_configuration_id": types.String{Value: "aws.kibana.r4"},
							"size":                      types.String{Value: "1g"},
							"size_resource":             types.String{Value: "memory"},
							"zone_count":                types.Int64{Value: 1},
						},
					}}},
			},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var model modelV0
			diags := flattenKibanaResources(context.Background(), tt.args.in, &model.Kibana)
			assert.Empty(t, diags)
			var got []kibanaResourceInfoModelV0
			model.Kibana.ElementsAs(context.Background(), &got, false)
			assert.Equal(t, tt.want, got)
		})
	}
}
