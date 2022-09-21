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

func Test_flattenIntegrationsServerResource(t *testing.T) {
	type args struct {
		in []*models.IntegrationsServerResourceInfo
	}
	tests := []struct {
		name string
		args args
		want []integrationsServerResourceInfoModelV0
	}{
		{
			name: "empty resource list returns empty list",
			args: args{in: []*models.IntegrationsServerResourceInfo{}},
			want: []integrationsServerResourceInfoModelV0{},
		},
		{
			name: "parses the integrations_server resource",
			args: args{in: []*models.IntegrationsServerResourceInfo{
				{
					RefID:                     ec.String("main-integrations_server"),
					ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
					Info: &models.IntegrationsServerInfo{
						Healthy: ec.Bool(true),
						Status:  ec.String("started"),
						ID:      &mock.ValidClusterID,
						Metadata: &models.ClusterMetadataInfo{
							Endpoint: "integrations_serverresource.cloud.elastic.co",
							Ports: &models.ClusterMetadataPortInfo{
								HTTP:  ec.Int32(9200),
								HTTPS: ec.Int32(9243),
							},
						},
						PlanInfo: &models.IntegrationsServerPlansInfo{
							Current: &models.IntegrationsServerPlanInfo{
								Plan: &models.IntegrationsServerPlan{
									IntegrationsServer: &models.IntegrationsServerConfiguration{
										Version: "8.0.0",
									},
									ClusterTopology: []*models.IntegrationsServerTopologyElement{
										{
											ZoneCount:               1,
											InstanceConfigurationID: "aws.integrations_server.r4",
											Size: &models.TopologySize{
												Resource: ec.String("memory"),
												Value:    ec.Int32(1024),
											},
										},
										{
											ZoneCount:               1,
											InstanceConfigurationID: "aws.integrations_server.m5d",
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
			want: []integrationsServerResourceInfoModelV0{{
				ElasticsearchClusterRefID: types.String{Value: "main-elasticsearch"},
				RefID:                     types.String{Value: "main-integrations_server"},
				ResourceID:                types.String{Value: mock.ValidClusterID},
				Version:                   types.String{Value: "8.0.0"},
				HttpEndpoint:              types.String{Value: "http://integrations_serverresource.cloud.elastic.co:9200"},
				HttpsEndpoint:             types.String{Value: "https://integrations_serverresource.cloud.elastic.co:9243"},
				Healthy:                   types.Bool{Value: true},
				Status:                    types.String{Value: "started"},
				Topology: types.List{ElemType: types.ObjectType{AttrTypes: integrationsServerTopologyAttrTypes()},
					Elems: []attr.Value{types.Object{
						AttrTypes: integrationsServerTopologyAttrTypes(),
						Attrs: map[string]attr.Value{
							"instance_configuration_id": types.String{Value: "aws.integrations_server.r4"},
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
			diags := flattenIntegrationsServerResources(context.Background(), tt.args.in, &newState.IntegrationsServer)
			assert.Empty(t, diags)

			var got []integrationsServerResourceInfoModelV0
			newState.IntegrationsServer.ElementsAs(context.Background(), &got, false)
			assert.Equal(t, tt.want, got)
		})
	}
}
