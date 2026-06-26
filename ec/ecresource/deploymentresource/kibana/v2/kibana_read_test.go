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

package v2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
)

func Test_ReadKibana(t *testing.T) {
	type args struct {
		in []*models.KibanaResourceInfo
	}
	tests := []struct {
		name string
		args args
		want *Kibana
	}{
		{
			name: "empty resource list returns empty list",
			args: args{in: []*models.KibanaResourceInfo{}},
			want: nil,
		},
		{
			name: "empty current plan returns empty list",
			args: args{in: []*models.KibanaResourceInfo{
				{
					Info: &models.KibanaClusterInfo{
						PlanInfo: &models.KibanaClusterPlansInfo{
							Pending: &models.KibanaClusterPlanInfo{},
						},
					},
				},
			}},
			want: nil,
		},
		{
			name: "parses the kibana resource",
			args: args{in: []*models.KibanaResourceInfo{
				{
					Region:                    new("some-region"),
					RefID:                     new("main-kibana"),
					ElasticsearchClusterRefID: new("main-elasticsearch"),
					Info: &models.KibanaClusterInfo{
						ClusterID:   &mock.ValidClusterID,
						ClusterName: new("some-kibana-name"),
						Region:      "some-region",
						Status:      new("stopped"),
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
											ZoneCount:                    1,
											InstanceConfigurationID:      "aws.kibana.r4",
											InstanceConfigurationVersion: ec.Int32(5),
											Size: &models.TopologySize{
												Resource: new("memory"),
												Value:    ec.Int32(1024),
											},
										},
									},
								},
							},
						},
					},
				},
				{
					Region:                    new("some-region"),
					RefID:                     new("main-kibana"),
					ElasticsearchClusterRefID: new("main-elasticsearch"),
					Info: &models.KibanaClusterInfo{
						ClusterID:   &mock.ValidClusterID,
						ClusterName: new("some-kibana-name"),
						Region:      "some-region",
						Status:      new("started"),
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
										Version:                  "7.7.0",
										UserSettingsYaml:         "some.setting: value",
										UserSettingsOverrideYaml: "some.setting: override",
										UserSettingsJSON: map[string]any{
											"some.setting": "value",
										},
										UserSettingsOverrideJSON: map[string]any{
											"some.setting": "override",
										},
									},
									ClusterTopology: []*models.KibanaClusterTopologyElement{{
										ZoneCount:                    1,
										InstanceConfigurationID:      "aws.kibana.r4",
										InstanceConfigurationVersion: ec.Int32(5),
										Size: &models.TopologySize{
											Resource: new("memory"),
											Value:    ec.Int32(1024),
										},
									}},
								},
							},
						},
					},
				},
			}},
			want: &Kibana{
				ElasticsearchClusterRefId: new("main-elasticsearch"),
				RefId:                     new("main-kibana"),
				ResourceId:                &mock.ValidClusterID,
				Region:                    new("some-region"),
				HttpEndpoint:              new("http://kibanaresource.cloud.elastic.co:9200"),
				HttpsEndpoint:             new("https://kibanaresource.cloud.elastic.co:9243"),
				Config: &KibanaConfig{
					UserSettingsYaml:         new("some.setting: value"),
					UserSettingsOverrideYaml: new("some.setting: override"),
					UserSettingsJson:         new(`{"some.setting":"value"}`),
					UserSettingsOverrideJson: new(`{"some.setting":"override"}`),
				},
				InstanceConfigurationId:      new("aws.kibana.r4"),
				InstanceConfigurationVersion: new(5),
				Size:                         new("1g"),
				SizeResource:                 new("memory"),
				ZoneCount:                    1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kibana, err := ReadKibanas(tt.args.in)
			assert.Nil(t, err)
			assert.Equal(t, tt.want, kibana)

			var obj types.Object
			diags := tfsdk.ValueFrom(context.Background(), kibana, KibanaSchema().GetType(), &obj)
			assert.Nil(t, diags)
		})
	}
}

func Test_IsKibanaResourceStopped(t *testing.T) {
	type args struct {
		res *models.KibanaResourceInfo
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "started resource returns false",
			args: args{res: &models.KibanaResourceInfo{Info: &models.KibanaClusterInfo{
				Status: new("started"),
			}}},
			want: false,
		},
		{
			name: "stopped resource returns true",
			args: args{res: &models.KibanaResourceInfo{Info: &models.KibanaClusterInfo{
				Status: new("stopped"),
			}}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsKibanaStopped(tt.args.res)
			assert.Equal(t, tt.want, got)
		})
	}
}
