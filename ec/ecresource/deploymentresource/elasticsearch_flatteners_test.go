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
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func Test_flattenEsTopology(t *testing.T) {
	type args struct {
		plan *models.ElasticsearchClusterPlan
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "no zombie topologies",
			args: args{plan: &models.ElasticsearchClusterPlan{
				ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
					{
						ID:                      "hot_content",
						ZoneCount:               1,
						InstanceConfigurationID: "aws.data.highio.i3",
						Size: &models.TopologySize{
							Value: ec.Int32(4096), Resource: ec.String("memory"),
						},
						NodeType: &models.ElasticsearchNodeType{
							Data:   ec.Bool(true),
							Ingest: ec.Bool(true),
							Master: ec.Bool(true),
						},
					},
					{
						ID:                      "coordinating",
						ZoneCount:               2,
						InstanceConfigurationID: "aws.coordinating.m5",
						Size: &models.TopologySize{
							Value: ec.Int32(0), Resource: ec.String("memory"),
						},
					},
				},
			}},
			want: []interface{}{
				map[string]interface{}{
					"id":                        "hot_content",
					"instance_configuration_id": "aws.data.highio.i3",
					"size":                      "4g",
					"size_resource":             "memory",
					"zone_count":                int32(1),
					"node_type_data":            "true",
					"node_type_ingest":          "true",
					"node_type_master":          "true",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := flattenEsTopology(tt.args.plan)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_flattenEsConfig(t *testing.T) {
	type args struct {
		cfg *models.ElasticsearchConfiguration
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "flattens plugins allowlist",
			args: args{cfg: &models.ElasticsearchConfiguration{
				EnabledBuiltInPlugins: []string{"some-allowed-plugin"},
			}},
			want: []interface{}{map[string]interface{}{
				"plugins": []interface{}{"some-allowed-plugin"},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := flattenEsConfig(tt.args.cfg)
			for _, g := range got {
				var rawVal []interface{}
				m := g.(map[string]interface{})
				if v, ok := m["plugins"]; ok {
					rawVal = v.(*schema.Set).List()
				}
				m["plugins"] = rawVal
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
