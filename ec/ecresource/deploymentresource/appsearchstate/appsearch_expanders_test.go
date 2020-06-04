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

package appsearchstate

import (
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/stretchr/testify/assert"
)

func TestExpandResources(t *testing.T) {
	type args struct {
		ess []interface{}
	}
	tests := []struct {
		name string
		args args
		want []*models.AppSearchPayload
		err  error
	}{
		{
			name: "returns nil when there's no resources",
		},
		{
			name: "parses multiple resources",
			args: args{
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":                       "main-appsearch",
						"resource_id":                  mock.ValidClusterID,
						"version":                      "7.7.0",
						"region":                       "some-region",
						"elasticsearch_cluster_ref_id": "somerefid",
						"topology": []interface{}{
							map[string]interface{}{
								"instance_configuration_id": "aws.appsearch.m5",
								"memory_per_node":           "2g",
								"zone_count":                1,
								"node_type_appserver":       true,
								"node_type_worker":          false,
							},
						},
					},
					map[string]interface{}{
						"display_name":                 "somename",
						"ref_id":                       "secondary-appsearch",
						"elasticsearch_cluster_ref_id": "somerefid",
						"resource_id":                  mock.ValidClusterID,
						"version":                      "7.6.0",
						"region":                       "some-region",
						"topology": []interface{}{
							map[string]interface{}{
								"instance_configuration_id": "aws.appsearch.m5",
								"memory_per_node":           "4g",
								"zone_count":                1,
								"node_type_appserver":       false,
								"node_type_worker":          true,
							}},
					},
				},
			},
			want: []*models.AppSearchPayload{
				{
					ElasticsearchClusterRefID: ec.String("somerefid"),
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("main-appsearch"),
					Settings:                  &models.AppSearchSettings{},
					Plan: &models.AppSearchPlan{
						Appsearch: &models.AppSearchConfiguration{
							Version: "7.7.0",
						},
						ClusterTopology: []*models.AppSearchTopologyElement{
							{
								ZoneCount:               1,
								InstanceConfigurationID: "aws.appsearch.m5",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(2048),
								},
								NodeType: &models.AppSearchNodeTypes{
									Appserver: ec.Bool(true),
									Worker:    ec.Bool(false),
								},
							},
						},
					},
				},
				{
					ElasticsearchClusterRefID: ec.String("somerefid"),
					DisplayName:               "somename",
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("secondary-appsearch"),
					Settings:                  &models.AppSearchSettings{},
					Plan: &models.AppSearchPlan{
						Appsearch: &models.AppSearchConfiguration{
							Version: "7.6.0",
						},
						ClusterTopology: []*models.AppSearchTopologyElement{
							{
								ZoneCount:               1,
								InstanceConfigurationID: "aws.appsearch.m5",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(4096),
								},
								NodeType: &models.AppSearchNodeTypes{
									Appserver: ec.Bool(false),
									Worker:    ec.Bool(true),
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
			got, err := ExpandResources(tt.args.ess)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
