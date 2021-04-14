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

func Test_flattenElasticsearchResources(t *testing.T) {
	type args struct {
		in []*models.ElasticsearchResourceInfo
	}
	tests := []struct {
		name string
		args args
		want []interface{}
		err  string
	}{
		{
			name: "empty resource list returns empty list",
			args: args{in: []*models.ElasticsearchResourceInfo{}},
			want: []interface{}{},
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
			want: []interface{}{map[string]interface{}{
				"autoscale":      "true",
				"ref_id":         "main-elasticsearch",
				"resource_id":    mock.ValidClusterID,
				"version":        "7.7.0",
				"cloud_id":       "some CLOUD ID",
				"http_endpoint":  "http://somecluster.cloud.elastic.co:9200",
				"https_endpoint": "https://somecluster.cloud.elastic.co:9243",
				"healthy":        true,
				"status":         "started",
				"topology": []interface{}{map[string]interface{}{
					"instance_configuration_id": "aws.data.highio.i3",
					"size":                      "2g",
					"size_resource":             "memory",
					"node_type_data":            true,
					"node_type_ingest":          true,
					"node_type_master":          true,
					"node_type_ml":              false,
					"zone_count":                int32(1),
					"autoscaling": []interface{}{map[string]interface{}{
						"max_size":          "15g",
						"max_size_resource": "memory",
						"min_size":          "1g",
						"min_size_resource": "memory",
					}},
				}},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := flattenElasticsearchResources(tt.args.in)
			if err != nil && assert.EqualError(t, err, tt.err) {
				t.Error(err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
