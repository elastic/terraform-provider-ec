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

func Test_readEnterpriseSearch(t *testing.T) {
	type args struct {
		in []*models.EnterpriseSearchResourceInfo
	}
	tests := []struct {
		name string
		args args
		want *EnterpriseSearch
	}{
		{
			name: "empty resource list returns empty list",
			args: args{in: []*models.EnterpriseSearchResourceInfo{}},
			want: nil,
		},
		{
			name: "empty current plan returns empty list",
			args: args{in: []*models.EnterpriseSearchResourceInfo{
				{
					Info: &models.EnterpriseSearchInfo{
						PlanInfo: &models.EnterpriseSearchPlansInfo{
							Pending: &models.EnterpriseSearchPlanInfo{},
						},
					},
				},
			}},
			want: nil,
		},
		{
			name: "parses the enterprisesearch resource",
			args: args{in: []*models.EnterpriseSearchResourceInfo{
				{
					Region:                    new("some-region"),
					RefID:                     new("main-enterprise_search"),
					ElasticsearchClusterRefID: new("main-elasticsearch"),
					Info: &models.EnterpriseSearchInfo{
						ID:     &mock.ValidClusterID,
						Name:   new("some-enterprisesearch-name"),
						Region: "some-region",
						Status: new("started"),
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
										Version:                  "7.7.0",
										UserSettingsYaml:         "some.setting: some value",
										UserSettingsOverrideYaml: "some.setting: some override",
										UserSettingsJSON: map[string]any{
											"some.setting": "some other value",
										},
										UserSettingsOverrideJSON: map[string]any{
											"some.setting": "some other override",
										},
									},
									ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
										EnterpriseSearch:             &models.EnterpriseSearchConfiguration{},
										ZoneCount:                    1,
										InstanceConfigurationID:      "aws.enterprisesearch.r4",
										InstanceConfigurationVersion: ec.Int32(5),
										Size: &models.TopologySize{
											Resource: new("memory"),
											Value:    ec.Int32(1024),
										},
										NodeType: &models.EnterpriseSearchNodeTypes{
											Appserver: new(true),
											Worker:    new(false),
										},
									}},
								},
							},
						},
					},
				},
				{
					Region:                    new("some-region"),
					RefID:                     new("main-enterprise_search"),
					ElasticsearchClusterRefID: new("main-elasticsearch"),
					Info: &models.EnterpriseSearchInfo{
						ID:     &mock.ValidClusterID,
						Name:   new("some-enterprisesearch-name"),
						Region: "some-region",
						Status: new("stopped"),
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
										Version:                  "7.7.0",
										UserSettingsYaml:         "some.setting: some value",
										UserSettingsOverrideYaml: "some.setting: some override",
										UserSettingsJSON: map[string]any{
											"some.setting": "some other value",
										},
										UserSettingsOverrideJSON: map[string]any{
											"some.setting": "some other override",
										},
									},
									ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
										EnterpriseSearch:             &models.EnterpriseSearchConfiguration{},
										ZoneCount:                    1,
										InstanceConfigurationID:      "aws.enterprisesearch.r4",
										InstanceConfigurationVersion: ec.Int32(5),
										Size: &models.TopologySize{
											Resource: new("memory"),
											Value:    ec.Int32(1024),
										},
										NodeType: &models.EnterpriseSearchNodeTypes{
											Appserver: new(true),
											Worker:    new(false),
										},
									}},
								},
							},
						},
					},
				},
			}},
			want: &EnterpriseSearch{
				ElasticsearchClusterRefId: new("main-elasticsearch"),
				RefId:                     new("main-enterprise_search"),
				ResourceId:                new(mock.ValidClusterID),
				Region:                    new("some-region"),
				HttpEndpoint:              new("http://enterprisesearchresource.cloud.elastic.co:9200"),
				HttpsEndpoint:             new("https://enterprisesearchresource.cloud.elastic.co:9243"),
				Config: &EnterpriseSearchConfig{
					UserSettingsJson:         new("{\"some.setting\":\"some other value\"}"),
					UserSettingsOverrideJson: new("{\"some.setting\":\"some other override\"}"),
					UserSettingsOverrideYaml: new("some.setting: some override"),
					UserSettingsYaml:         new("some.setting: some value"),
				},
				InstanceConfigurationId:      new("aws.enterprisesearch.r4"),
				InstanceConfigurationVersion: new(5),
				Size:                         new("1g"),
				SizeResource:                 new("memory"),
				ZoneCount:                    1,
				NodeTypeAppserver:            new(true),
				NodeTypeWorker:               new(false),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadEnterpriseSearches(tt.args.in)
			assert.Nil(t, err)
			assert.Equal(t, tt.want, got)

			var obj types.Object
			diags := tfsdk.ValueFrom(context.Background(), got, EnterpriseSearchSchema().GetType(), &obj)
			assert.Nil(t, diags)
		})
	}
}

func Test_IsEnterpriseSearchStopped(t *testing.T) {
	type args struct {
		res *models.EnterpriseSearchResourceInfo
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "started resource returns false",
			args: args{res: &models.EnterpriseSearchResourceInfo{Info: &models.EnterpriseSearchInfo{
				Status: new("started"),
			}}},
			want: false,
		},
		{
			name: "stopped resource returns true",
			args: args{res: &models.EnterpriseSearchResourceInfo{Info: &models.EnterpriseSearchInfo{
				Status: new("stopped"),
			}}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEnterpriseSearchStopped(tt.args.res)
			assert.Equal(t, tt.want, got)
		})
	}
}
