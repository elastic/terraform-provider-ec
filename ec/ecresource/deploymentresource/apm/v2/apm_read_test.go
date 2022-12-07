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

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	v1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/apm/v1"
)

func Test_readApm(t *testing.T) {
	type args struct {
		in []*models.ApmResourceInfo
	}

	tests := []struct {
		name  string
		args  args
		want  *Apm
		diags diag.Diagnostics
	}{
		{
			name:  "empty resource list returns empty list",
			args:  args{in: []*models.ApmResourceInfo{}},
			want:  nil,
			diags: nil,
		},
		{
			name: "empty current plan returns empty list",
			args: args{in: []*models.ApmResourceInfo{
				{
					Info: &models.ApmInfo{
						PlanInfo: &models.ApmPlansInfo{
							Pending: &models.ApmPlanInfo{},
						},
					},
				},
			}},
			want:  nil,
			diags: nil,
		},
		{
			name: "parses the apm resource",
			args: args{in: []*models.ApmResourceInfo{
				{
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("main-apm"),
					ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
					Info: &models.ApmInfo{
						ID:     &mock.ValidClusterID,
						Name:   ec.String("some-apm-name"),
						Region: "some-region",
						Status: ec.String("started"),
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
								},
							},
						}},
					},
				},
			}},
			want: &Apm{
				ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
				RefId:                     ec.String("main-apm"),
				ResourceId:                &mock.ValidClusterID,
				Region:                    ec.String("some-region"),
				HttpEndpoint:              ec.String("http://apmresource.cloud.elastic.co:9200"),
				HttpsEndpoint:             ec.String("https://apmresource.cloud.elastic.co:9243"),
				InstanceConfigurationId:   ec.String("aws.apm.r4"),
				Size:                      ec.String("1g"),
				SizeResource:              ec.String("memory"),
				ZoneCount:                 1,
			},
		},
		{
			name: "parses the apm resource with config overrides, ignoring a stopped resource",
			args: args{in: []*models.ApmResourceInfo{
				{
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("main-apm"),
					ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
					Info: &models.ApmInfo{
						ID:     &mock.ValidClusterID,
						Name:   ec.String("some-apm-name"),
						Region: "some-region",
						Status: ec.String("started"),
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
									Version:                  "7.8.0",
									UserSettingsYaml:         `some.setting: value`,
									UserSettingsOverrideYaml: `some.setting: value2`,
									UserSettingsJSON: map[string]interface{}{
										"some.setting": "value",
									},
									UserSettingsOverrideJSON: map[string]interface{}{
										"some.setting": "value2",
									},
									SystemSettings: &models.ApmSystemSettings{},
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
								},
							},
						}},
					},
				},
				{
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("main-apm"),
					ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
					Info: &models.ApmInfo{
						ID:     &mock.ValidClusterID,
						Name:   ec.String("some-apm-name"),
						Region: "some-region",
						Status: ec.String("stopped"),
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
									Version:                  "7.8.0",
									UserSettingsYaml:         `some.setting: value`,
									UserSettingsOverrideYaml: `some.setting: value2`,
									UserSettingsJSON: map[string]interface{}{
										"some.setting": "value",
									},
									UserSettingsOverrideJSON: map[string]interface{}{
										"some.setting": "value2",
									},
									SystemSettings: &models.ApmSystemSettings{},
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
								},
							},
						}},
					},
				},
			}},
			want: &Apm{
				ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
				RefId:                     ec.String("main-apm"),
				ResourceId:                &mock.ValidClusterID,
				Region:                    ec.String("some-region"),
				HttpEndpoint:              ec.String("http://apmresource.cloud.elastic.co:9200"),
				HttpsEndpoint:             ec.String("https://apmresource.cloud.elastic.co:9243"),
				InstanceConfigurationId:   ec.String("aws.apm.r4"),
				Size:                      ec.String("1g"),
				SizeResource:              ec.String("memory"),
				ZoneCount:                 1,
				Config: &v1.ApmConfig{
					UserSettingsYaml:         ec.String("some.setting: value"),
					UserSettingsOverrideYaml: ec.String("some.setting: value2"),
					UserSettingsJson:         ec.String("{\"some.setting\":\"value\"}"),
					UserSettingsOverrideJson: ec.String("{\"some.setting\":\"value2\"}"),
				},
			},
		},
		{
			name: "parses the apm resource with config overrides and system settings",
			args: args{in: []*models.ApmResourceInfo{
				{
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("main-apm"),
					ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
					Info: &models.ApmInfo{
						ID:     &mock.ValidClusterID,
						Name:   ec.String("some-apm-name"),
						Region: "some-region",
						Status: ec.String("started"),
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
									Version:                  "7.8.0",
									UserSettingsYaml:         `some.setting: value`,
									UserSettingsOverrideYaml: `some.setting: value2`,
									UserSettingsJSON: map[string]interface{}{
										"some.setting": "value",
									},
									UserSettingsOverrideJSON: map[string]interface{}{
										"some.setting": "value2",
									},
									SystemSettings: &models.ApmSystemSettings{
										DebugEnabled: ec.Bool(true),
									},
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
								},
							},
						}},
					},
				},
			}},
			want: &Apm{
				ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
				RefId:                     ec.String("main-apm"),
				ResourceId:                &mock.ValidClusterID,
				Region:                    ec.String("some-region"),
				HttpEndpoint:              ec.String("http://apmresource.cloud.elastic.co:9200"),
				HttpsEndpoint:             ec.String("https://apmresource.cloud.elastic.co:9243"),
				InstanceConfigurationId:   ec.String("aws.apm.r4"),
				Size:                      ec.String("1g"),
				SizeResource:              ec.String("memory"),
				ZoneCount:                 1,
				Config: &v1.ApmConfig{
					UserSettingsYaml:         ec.String("some.setting: value"),
					UserSettingsOverrideYaml: ec.String("some.setting: value2"),
					UserSettingsJson:         ec.String("{\"some.setting\":\"value\"}"),
					UserSettingsOverrideJson: ec.String("{\"some.setting\":\"value2\"}"),
					DebugEnabled:             ec.Bool(true),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apms, err := ReadApms(tt.args.in)
			assert.Nil(t, err)
			assert.Equal(t, tt.want, apms)

			var apmTF types.Object
			diags := tfsdk.ValueFrom(context.Background(), apms, ApmSchema().FrameworkType(), &apmTF)
			assert.Nil(t, diags)
		})
	}
}
