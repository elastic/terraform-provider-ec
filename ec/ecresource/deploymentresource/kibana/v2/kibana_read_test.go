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
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("main-kibana"),
					ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
					Info: &models.KibanaClusterInfo{
						ClusterID:   &mock.ValidClusterID,
						ClusterName: ec.String("some-kibana-name"),
						Region:      "some-region",
						Status:      ec.String("stopped"),
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
									},
								},
							},
						},
					},
				},
				{
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("main-kibana"),
					ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
					Info: &models.KibanaClusterInfo{
						ClusterID:   &mock.ValidClusterID,
						ClusterName: ec.String("some-kibana-name"),
						Region:      "some-region",
						Status:      ec.String("started"),
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
										UserSettingsJSON: map[string]interface{}{
											"some.setting": "value",
										},
										UserSettingsOverrideJSON: map[string]interface{}{
											"some.setting": "override",
										},
									},
									ClusterTopology: []*models.KibanaClusterTopologyElement{{
										ZoneCount:               1,
										InstanceConfigurationID: "aws.kibana.r4",
										Size: &models.TopologySize{
											Resource: ec.String("memory"),
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
				ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
				RefId:                     ec.String("main-kibana"),
				ResourceId:                &mock.ValidClusterID,
				Region:                    ec.String("some-region"),
				HttpEndpoint:              ec.String("http://kibanaresource.cloud.elastic.co:9200"),
				HttpsEndpoint:             ec.String("https://kibanaresource.cloud.elastic.co:9243"),
				Config: &KibanaConfig{
					UserSettingsYaml:         ec.String("some.setting: value"),
					UserSettingsOverrideYaml: ec.String("some.setting: override"),
					UserSettingsJson:         ec.String(`{"some.setting":"value"}`),
					UserSettingsOverrideJson: ec.String(`{"some.setting":"override"}`),
				},
				InstanceConfigurationId: ec.String("aws.kibana.r4"),
				Size:                    ec.String("1g"),
				SizeResource:            ec.String("memory"),
				ZoneCount:               1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kibana, err := ReadKibanas(tt.args.in)
			assert.Nil(t, err)
			assert.Equal(t, tt.want, kibana)

			var kibanaTF types.Object
			diags := tfsdk.ValueFrom(context.Background(), kibana, KibanaSchema().FrameworkType(), &kibanaTF)
			assert.Nil(t, diags)
		})
	}
}
