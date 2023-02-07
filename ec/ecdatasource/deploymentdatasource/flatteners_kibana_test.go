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
			want: []kibanaResourceInfoModelV0{
				{
					ElasticsearchClusterRefID: types.StringValue("main-elasticsearch"),
					RefID:                     types.StringValue("main-kibana"),
					ResourceID:                types.StringValue(mock.ValidClusterID),
					Version:                   types.StringValue("7.7.0"),
					HttpEndpoint:              types.StringValue("http://kibanaresource.cloud.elastic.co:9200"),
					HttpsEndpoint:             types.StringValue("https://kibanaresource.cloud.elastic.co:9243"),
					Healthy:                   types.BoolValue(true),
					Status:                    types.StringValue("started"),
					Topology: func() types.List {
						res, diags := types.ListValueFrom(
							context.Background(),
							types.ObjectType{AttrTypes: kibanaTopologyAttrTypes()},
							[]kibanaTopologyModelV0{
								{
									InstanceConfigurationID: types.StringValue("aws.kibana.r4"),
									Size:                    types.StringValue("1g"),
									SizeResource:            types.StringValue("memory"),
									ZoneCount:               types.Int64Value(1),
								},
							},
						)
						assert.Nil(t, diags)

						return res
					}(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kibana, diags := flattenKibanaResources(context.Background(), tt.args.in)
			assert.Empty(t, diags)
			var got []kibanaResourceInfoModelV0
			kibana.ElementsAs(context.Background(), &got, false)
			assert.Equal(t, tt.want, got)
			util.CheckConverionToAttrValue(t, &DataSource{}, "kibana", kibana)
		})
	}
}
