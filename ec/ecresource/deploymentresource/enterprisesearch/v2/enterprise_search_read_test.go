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
	"testing"

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
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("main-enterprise_search"),
					ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
					Info: &models.EnterpriseSearchInfo{
						ID:     &mock.ValidClusterID,
						Name:   ec.String("some-enterprisesearch-name"),
						Region: "some-region",
						Status: ec.String("started"),
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
										UserSettingsJSON: map[string]interface{}{
											"some.setting": "some other value",
										},
										UserSettingsOverrideJSON: map[string]interface{}{
											"some.setting": "some other override",
										},
									},
									ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
										EnterpriseSearch:        &models.EnterpriseSearchConfiguration{},
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
				{
					Region:                    ec.String("some-region"),
					RefID:                     ec.String("main-enterprise_search"),
					ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
					Info: &models.EnterpriseSearchInfo{
						ID:     &mock.ValidClusterID,
						Name:   ec.String("some-enterprisesearch-name"),
						Region: "some-region",
						Status: ec.String("stopped"),
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
										UserSettingsJSON: map[string]interface{}{
											"some.setting": "some other value",
										},
										UserSettingsOverrideJSON: map[string]interface{}{
											"some.setting": "some other override",
										},
									},
									ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
										EnterpriseSearch:        &models.EnterpriseSearchConfiguration{},
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
			want: &EnterpriseSearch{
				ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
				RefId:                     ec.String("main-enterprise_search"),
				ResourceId:                ec.String(mock.ValidClusterID),
				Region:                    ec.String("some-region"),
				HttpEndpoint:              ec.String("http://enterprisesearchresource.cloud.elastic.co:9200"),
				HttpsEndpoint:             ec.String("https://enterprisesearchresource.cloud.elastic.co:9243"),
				Config: &EnterpriseSearchConfig{
					UserSettingsJson:         ec.String("{\"some.setting\":\"some other value\"}"),
					UserSettingsOverrideJson: ec.String("{\"some.setting\":\"some other override\"}"),
					UserSettingsOverrideYaml: ec.String("some.setting: some override"),
					UserSettingsYaml:         ec.String("some.setting: some value"),
				},
				InstanceConfigurationId: ec.String("aws.enterprisesearch.r4"),
				Size:                    ec.String("1g"),
				SizeResource:            ec.String("memory"),
				ZoneCount:               1,
				NodeTypeAppserver:       ec.Bool(true),
				NodeTypeWorker:          ec.Bool(false),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadEnterpriseSearches(tt.args.in)
			assert.Nil(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
