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

func Test_readIntegrationsServer(t *testing.T) {
	type args struct {
		in []*models.IntegrationsServerResourceInfo
	}
	tests := []struct {
		name string
		args args
		want *IntegrationsServer
	}{
		{
			name: "empty resource list returns empty list",
			args: args{in: []*models.IntegrationsServerResourceInfo{}},
			want: nil,
		},
		{
			name: "empty current plan returns empty list",
			args: args{in: []*models.IntegrationsServerResourceInfo{
				{
					Info: &models.IntegrationsServerInfo{
						PlanInfo: &models.IntegrationsServerPlansInfo{
							Pending: &models.IntegrationsServerPlanInfo{},
						},
					},
				},
			}},
			want: nil,
		},
		{
			name: "parses the integrations_server resource",
			args: args{in: []*models.IntegrationsServerResourceInfo{
				{
					Region:                    new("some-region"),
					RefID:                     new("main-integrations_server"),
					ElasticsearchClusterRefID: new("main-elasticsearch"),
					Info: &models.IntegrationsServerInfo{
						ID:     &mock.ValidClusterID,
						Name:   new("some-integrations_server-name"),
						Region: "some-region",
						Status: new("started"),
						Metadata: &models.ClusterMetadataInfo{
							Endpoint: "integrations_serverresource.cloud.elastic.co",
							Ports: &models.ClusterMetadataPortInfo{
								HTTP:  ec.Int32(9200),
								HTTPS: ec.Int32(9243),
							},
							ServicesUrls: []*models.ServiceURL{
								{
									Service: new("apm"),
									URL:     new("https://apm_endpoint.cloud.elastic.co"),
								},
								{
									Service: new("fleet"),
									URL:     new("https://fleet_endpoint.cloud.elastic.co"),
								},
								{
									Service: new("symbols"),
									URL:     new("https://symbols_endpoint.cloud.elastic.co"),
								},
								{
									Service: new("profiling"),
									URL:     new("https://profiling_endpoint.cloud.elastic.co"),
								},
							},
						},
						PlanInfo: &models.IntegrationsServerPlansInfo{Current: &models.IntegrationsServerPlanInfo{
							Plan: &models.IntegrationsServerPlan{
								IntegrationsServer: &models.IntegrationsServerConfiguration{
									Version: "7.7.0",
								},
								ClusterTopology: []*models.IntegrationsServerTopologyElement{
									{
										ZoneCount:                    1,
										InstanceConfigurationID:      "aws.integrations_server.r4",
										InstanceConfigurationVersion: ec.Int32(5),
										Size: &models.TopologySize{
											Resource: new("memory"),
											Value:    ec.Int32(1024),
										},
									},
								},
							},
						}},
					},
				},
			}},
			want: &IntegrationsServer{
				ElasticsearchClusterRefId: new("main-elasticsearch"),
				RefId:                     new("main-integrations_server"),
				ResourceId:                &mock.ValidClusterID,
				Region:                    new("some-region"),
				HttpEndpoint:              new("http://integrations_serverresource.cloud.elastic.co:9200"),
				HttpsEndpoint:             new("https://integrations_serverresource.cloud.elastic.co:9243"),
				Endpoints: &Endpoints{
					Fleet:     new("https://fleet_endpoint.cloud.elastic.co"),
					APM:       new("https://apm_endpoint.cloud.elastic.co"),
					Symbols:   new("https://symbols_endpoint.cloud.elastic.co"),
					Profiling: new("https://profiling_endpoint.cloud.elastic.co"),
				},
				InstanceConfigurationId:      new("aws.integrations_server.r4"),
				InstanceConfigurationVersion: new(5),
				Size:                         new("1g"),
				SizeResource:                 new("memory"),
				ZoneCount:                    1,
			},
		},
		{
			name: "parses the integrations_server resource with config overrides, ignoring a stopped resource",
			args: args{
				in: []*models.IntegrationsServerResourceInfo{
					{
						Region:                    new("some-region"),
						RefID:                     new("main-integrations_server"),
						ElasticsearchClusterRefID: new("main-elasticsearch"),
						Info: &models.IntegrationsServerInfo{
							ID:     &mock.ValidClusterID,
							Name:   new("some-integrations_server-name"),
							Region: "some-region",
							Status: new("started"),
							Metadata: &models.ClusterMetadataInfo{
								Endpoint: "integrations_serverresource.cloud.elastic.co",
								Ports: &models.ClusterMetadataPortInfo{
									HTTP:  ec.Int32(9200),
									HTTPS: ec.Int32(9243),
								},
								ServicesUrls: []*models.ServiceURL{
									{
										Service: new("apm"),
										URL:     new("https://apm_endpoint.cloud.elastic.co"),
									},
									{
										Service: new("fleet"),
										URL:     new("https://fleet_endpoint.cloud.elastic.co"),
									},
									{
										Service: new("symbols"),
										URL:     new("https://symbols_endpoint.cloud.elastic.co"),
									},
									{
										Service: new("profiling"),
										URL:     new("https://profiling_endpoint.cloud.elastic.co"),
									},
								},
							},
							PlanInfo: &models.IntegrationsServerPlansInfo{Current: &models.IntegrationsServerPlanInfo{
								Plan: &models.IntegrationsServerPlan{
									IntegrationsServer: &models.IntegrationsServerConfiguration{
										Version:                  "7.8.0",
										UserSettingsYaml:         `some.setting: value`,
										UserSettingsOverrideYaml: `some.setting: value2`,
										UserSettingsJSON: map[string]any{
											"some.setting": "value",
										},
										UserSettingsOverrideJSON: map[string]any{
											"some.setting": "value2",
										},
										SystemSettings: &models.IntegrationsServerSystemSettings{},
									},
									ClusterTopology: []*models.IntegrationsServerTopologyElement{
										{
											ZoneCount:               1,
											InstanceConfigurationID: "aws.integrations_server.r4",
											Size: &models.TopologySize{
												Resource: new("memory"),
												Value:    ec.Int32(1024),
											},
										},
									},
								},
							}},
						},
					},
					{
						Region:                    new("some-region"),
						RefID:                     new("main-integrations_server"),
						ElasticsearchClusterRefID: new("main-elasticsearch"),
						Info: &models.IntegrationsServerInfo{
							ID:     &mock.ValidClusterID,
							Name:   new("some-integrations_server-name"),
							Region: "some-region",
							Status: new("stopped"),
							Metadata: &models.ClusterMetadataInfo{
								Endpoint: "integrations_serverresource.cloud.elastic.co",
								Ports: &models.ClusterMetadataPortInfo{
									HTTP:  ec.Int32(9200),
									HTTPS: ec.Int32(9243),
								},
								ServicesUrls: []*models.ServiceURL{
									{
										Service: new("apm"),
										URL:     new("https://apm_endpoint.cloud.elastic.co"),
									},
									{
										Service: new("fleet"),
										URL:     new("https://fleet_endpoint.cloud.elastic.co"),
									},
									{
										Service: new("symbols"),
										URL:     new("https://symbols_endpoint.cloud.elastic.co"),
									},
									{
										Service: new("profiling"),
										URL:     new("https://profiling_endpoint.cloud.elastic.co"),
									},
								},
							},
							PlanInfo: &models.IntegrationsServerPlansInfo{Current: &models.IntegrationsServerPlanInfo{
								Plan: &models.IntegrationsServerPlan{
									IntegrationsServer: &models.IntegrationsServerConfiguration{
										Version:                  "7.8.0",
										UserSettingsYaml:         `some.setting: value`,
										UserSettingsOverrideYaml: `some.setting: value2`,
										UserSettingsJSON: map[string]any{
											"some.setting": "value",
										},
										UserSettingsOverrideJSON: map[string]any{
											"some.setting": "value2",
										},
										SystemSettings: &models.IntegrationsServerSystemSettings{},
									},
									ClusterTopology: []*models.IntegrationsServerTopologyElement{
										{
											ZoneCount:               1,
											InstanceConfigurationID: "aws.integrations_server.r4",
											Size: &models.TopologySize{
												Resource: new("memory"),
												Value:    ec.Int32(1024),
											},
										},
									},
								},
							}},
						},
					},
				},
			},
			want: &IntegrationsServer{
				ElasticsearchClusterRefId: new("main-elasticsearch"),
				RefId:                     new("main-integrations_server"),
				ResourceId:                &mock.ValidClusterID,
				Region:                    new("some-region"),
				HttpEndpoint:              new("http://integrations_serverresource.cloud.elastic.co:9200"),
				HttpsEndpoint:             new("https://integrations_serverresource.cloud.elastic.co:9243"),
				Endpoints: &Endpoints{
					Fleet:     new("https://fleet_endpoint.cloud.elastic.co"),
					APM:       new("https://apm_endpoint.cloud.elastic.co"),
					Symbols:   new("https://symbols_endpoint.cloud.elastic.co"),
					Profiling: new("https://profiling_endpoint.cloud.elastic.co"),
				},
				InstanceConfigurationId: new("aws.integrations_server.r4"),
				Size:                    new("1g"),
				SizeResource:            new("memory"),
				ZoneCount:               1,
				Config: &IntegrationsServerConfig{
					UserSettingsYaml:         new("some.setting: value"),
					UserSettingsOverrideYaml: new("some.setting: value2"),
					UserSettingsJson:         new("{\"some.setting\":\"value\"}"),
					UserSettingsOverrideJson: new("{\"some.setting\":\"value2\"}"),
				},
			},
		},
		{
			name: "parses the integrations_server resource with config overrides and system settings",
			args: args{in: []*models.IntegrationsServerResourceInfo{
				{
					Region:                    new("some-region"),
					RefID:                     new("main-integrations_server"),
					ElasticsearchClusterRefID: new("main-elasticsearch"),
					Info: &models.IntegrationsServerInfo{
						ID:     &mock.ValidClusterID,
						Name:   new("some-integrations_server-name"),
						Region: "some-region",
						Status: new("started"),
						Metadata: &models.ClusterMetadataInfo{
							Endpoint: "integrations_serverresource.cloud.elastic.co",
							Ports: &models.ClusterMetadataPortInfo{
								HTTP:  ec.Int32(9200),
								HTTPS: ec.Int32(9243),
							},
							ServicesUrls: []*models.ServiceURL{
								{
									Service: new("apm"),
									URL:     new("https://apm_endpoint.cloud.elastic.co"),
								},
								{
									Service: new("fleet"),
									URL:     new("https://fleet_endpoint.cloud.elastic.co"),
								},
								{
									Service: new("symbols"),
									URL:     new("https://symbols_endpoint.cloud.elastic.co"),
								},
								{
									Service: new("profiling"),
									URL:     new("https://profiling_endpoint.cloud.elastic.co"),
								},
							},
						},
						PlanInfo: &models.IntegrationsServerPlansInfo{Current: &models.IntegrationsServerPlanInfo{
							Plan: &models.IntegrationsServerPlan{
								IntegrationsServer: &models.IntegrationsServerConfiguration{
									Version:                  "7.8.0",
									UserSettingsYaml:         `some.setting: value`,
									UserSettingsOverrideYaml: `some.setting: value2`,
									UserSettingsJSON: map[string]any{
										"some.setting": "value",
									},
									UserSettingsOverrideJSON: map[string]any{
										"some.setting": "value2",
									},
									SystemSettings: &models.IntegrationsServerSystemSettings{
										DebugEnabled: new(true),
									},
								},
								ClusterTopology: []*models.IntegrationsServerTopologyElement{
									{
										ZoneCount:               1,
										InstanceConfigurationID: "aws.integrations_server.r4",
										Size: &models.TopologySize{
											Resource: new("memory"),
											Value:    ec.Int32(1024),
										},
									},
								},
							},
						}},
					},
				},
			}},
			want: &IntegrationsServer{
				ElasticsearchClusterRefId: new("main-elasticsearch"),
				RefId:                     new("main-integrations_server"),
				ResourceId:                &mock.ValidClusterID,
				Region:                    new("some-region"),
				HttpEndpoint:              new("http://integrations_serverresource.cloud.elastic.co:9200"),
				HttpsEndpoint:             new("https://integrations_serverresource.cloud.elastic.co:9243"),
				Endpoints: &Endpoints{
					Fleet:     new("https://fleet_endpoint.cloud.elastic.co"),
					APM:       new("https://apm_endpoint.cloud.elastic.co"),
					Symbols:   new("https://symbols_endpoint.cloud.elastic.co"),
					Profiling: new("https://profiling_endpoint.cloud.elastic.co"),
				},
				InstanceConfigurationId: new("aws.integrations_server.r4"),
				Size:                    new("1g"),
				SizeResource:            new("memory"),
				ZoneCount:               1,
				Config: &IntegrationsServerConfig{
					UserSettingsYaml:         new("some.setting: value"),
					UserSettingsOverrideYaml: new("some.setting: value2"),
					UserSettingsJson:         new("{\"some.setting\":\"value\"}"),
					UserSettingsOverrideJson: new("{\"some.setting\":\"value2\"}"),
					DebugEnabled:             new(true),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, err := ReadIntegrationsServers(tt.args.in)
			assert.Nil(t, err)
			assert.Equal(t, tt.want, srv)

			var obj types.Object
			diags := tfsdk.ValueFrom(context.Background(), srv, IntegrationsServerSchema().GetType(), &obj)
			assert.Nil(t, diags)
		})
	}
}

func Test_IsIntegrationsServerStopped(t *testing.T) {
	type args struct {
		res *models.IntegrationsServerResourceInfo
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "started resource returns false",
			args: args{res: &models.IntegrationsServerResourceInfo{Info: &models.IntegrationsServerInfo{
				Status: new("started"),
			}}},
			want: false,
		},
		{
			name: "stopped resource returns true",
			args: args{res: &models.IntegrationsServerResourceInfo{Info: &models.IntegrationsServerInfo{
				Status: new("stopped"),
			}}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsIntegrationsServerStopped(tt.args.res)
			assert.Equal(t, tt.want, got)
		})
	}
}
