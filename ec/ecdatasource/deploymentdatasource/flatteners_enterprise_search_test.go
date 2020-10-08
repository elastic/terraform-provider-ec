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
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/stretchr/testify/assert"
)

func Test_flattenEnterpriseSearchResource(t *testing.T) {
	type args struct {
		in []*models.EnterpriseSearchResourceInfo
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "empty resource list returns empty list",
			args: args{in: []*models.EnterpriseSearchResourceInfo{}},
			want: []interface{}{},
		},
		{
			name: "parses the enterprisesearch resource",
			args: args{in: []*models.EnterpriseSearchResourceInfo{
				{
					RefID:                     ec.String("main-enterprise_search"),
					ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
					Info: &models.EnterpriseSearchInfo{
						Healthy: ec.Bool(true),
						Status:  ec.String("started"),
						ID:      &mock.ValidClusterID,
						Name:    ec.String("some-enterprisesearch-name"),
						Metadata: &models.ClusterMetadataInfo{
							Endpoint: "enterprisesearchresource.cloud.elastic.co",
							Ports: &models.ClusterMetadataPortInfo{
								HTTP:  ec.Int32(9200),
								HTTPS: ec.Int32(9243),
							},
						},
						PlanInfo: &models.EnterpriseSearchPlansInfo{
							Current: &models.EnterpriseSearchPlanInfo{
								Plan: &models.EnterpriseSearchPlan{
									EnterpriseSearch: &models.EnterpriseSearchConfiguration{
										Version: "7.7.0",
									},
									ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
										ZoneCount:               1,
										InstanceConfigurationID: "aws.enterprisesearch.r4",
										Size: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
										NodeType: &models.EnterpriseSearchNodeTypes{
											Appserver: ec.Bool(true),
											Worker:    ec.Bool(false),
										},
									}},
								},
							},
						},
					},
				},
			}},
			want: []interface{}{
				map[string]interface{}{
					"elasticsearch_cluster_ref_id": "main-elasticsearch",
					"ref_id":                       "main-enterprise_search",
					"resource_id":                  mock.ValidClusterID,
					"version":                      "7.7.0",
					"http_endpoint":                "http://enterprisesearchresource.cloud.elastic.co:9200",
					"https_endpoint":               "https://enterprisesearchresource.cloud.elastic.co:9243",
					"healthy":                      true,
					"status":                       "started",
					"topology": []interface{}{map[string]interface{}{
						"instance_configuration_id": "aws.enterprisesearch.r4",
						"size":                      "1g",
						"size_resource":             "memory",
						"zone_count":                int32(1),
						"node_type_appserver":       true,
						"node_type_worker":          false,
					}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := flattenEnterpriseSearchResources(tt.args.in)
			assert.Equal(t, tt.want, got)
		})
	}
}
