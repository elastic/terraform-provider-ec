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

package deploymenttemplates

import (
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_mapResponseToModel(t *testing.T) {
	tests := []struct {
		name           string
		apiResponse    []*models.DeploymentTemplateInfoV2
		showDeprecated bool
		expected       []deploymentTemplateModel
	}{
		{
			name:           "should filter out any hidden templates for showDeprecated=false",
			showDeprecated: false,
			apiResponse: []*models.DeploymentTemplateInfoV2{
				{
					ID:          ec.String("id-nonhidden"),
					Name:        ec.String("name-nonhidden"),
					Description: "description nonhidden",
					MinVersion:  "",
					Metadata:    []*models.MetadataItem{},
				},
				{
					ID:          ec.String("id-hidden"),
					Name:        ec.String("name-hidden"),
					Description: "description hidden",
					MinVersion:  "7.17.0",
					Metadata: []*models.MetadataItem{
						{
							Key:   ec.String("anotherkey"),
							Value: ec.String("false"),
						},
						{
							Key:   ec.String("hidden"),
							Value: ec.String("true"),
						},
					},
				},
			},
			expected: []deploymentTemplateModel{
				{
					ID:              "id-nonhidden",
					Name:            "name-nonhidden",
					Description:     "description nonhidden",
					MinStackVersion: "",
					Deprecated:      false,
				},
			},
		},
		{
			name:           "should show all templates for showDeprecated=true",
			showDeprecated: true,
			apiResponse: []*models.DeploymentTemplateInfoV2{
				{
					ID:          ec.String("id-nonhidden"),
					Name:        ec.String("name-nonhidden"),
					Description: "description nonhidden",
					MinVersion:  "",
					Metadata:    []*models.MetadataItem{},
				},
				{
					ID:          ec.String("id-hidden"),
					Name:        ec.String("name-hidden"),
					Description: "description hidden",
					MinVersion:  "7.17.0",
					Metadata: []*models.MetadataItem{
						{
							Key:   ec.String("anotherkey"),
							Value: ec.String("false"),
						},
						{
							Key:   ec.String("hidden"),
							Value: ec.String("true"),
						},
					},
				},
			},
			expected: []deploymentTemplateModel{
				{
					ID:              "id-nonhidden",
					Name:            "name-nonhidden",
					Description:     "description nonhidden",
					MinStackVersion: "",
					Deprecated:      false,
				},
				{
					ID:              "id-hidden",
					Name:            "name-hidden",
					Description:     "description hidden",
					MinStackVersion: "7.17.0",
					Deprecated:      true,
				},
			},
		},
		{
			name:           "should correctly map the toplogy",
			showDeprecated: true,
			apiResponse: []*models.DeploymentTemplateInfoV2{
				{
					ID:   ec.String("id"),
					Name: ec.String("name"),
					DeploymentTemplate: &models.DeploymentCreateRequest{
						Resources: &models.DeploymentCreateResources{
							Elasticsearch: []*models.ElasticsearchPayload{
								{
									Plan: &models.ElasticsearchClusterPlan{
										AutoscalingEnabled: ec.Bool(true),
										ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
											buildRequestTopology("hot_content", "es-hot"),
											buildRequestTopology("coordinating", "es-coord"),
											buildRequestTopology("master", "es-master"),
											buildRequestTopology("warm", "es-warm"),
											buildRequestTopology("cold", "es-cold"),
											buildRequestTopology("frozen", "es-frozen"),
											buildRequestTopology("ml", "es-ml"),
										},
									},
								},
							},
							Kibana: []*models.KibanaPayload{
								{
									Plan: &models.KibanaClusterPlan{
										ClusterTopology: []*models.KibanaClusterTopologyElement{
											{
												InstanceConfigurationID:      "kibana-id",
												InstanceConfigurationVersion: ec.Int32(1),
												Size: &models.TopologySize{
													Resource: ec.String("memory"),
													Value:    ec.Int32(1024),
												},
											},
										},
									},
								},
							},
							EnterpriseSearch: []*models.EnterpriseSearchPayload{
								{
									Plan: &models.EnterpriseSearchPlan{
										ClusterTopology: []*models.EnterpriseSearchTopologyElement{
											{
												InstanceConfigurationID:      "enterprise-search-id",
												InstanceConfigurationVersion: ec.Int32(1),
												Size: &models.TopologySize{
													Resource: ec.String("memory"),
													Value:    ec.Int32(1024),
												},
											},
										},
									},
								},
							},
							Apm: []*models.ApmPayload{
								{
									Plan: &models.ApmPlan{
										ClusterTopology: []*models.ApmTopologyElement{
											{
												InstanceConfigurationID:      "apm-id",
												InstanceConfigurationVersion: ec.Int32(1),
												Size: &models.TopologySize{
													Resource: ec.String("memory"),
													Value:    ec.Int32(1024),
												},
											},
										},
									},
								},
							},
							IntegrationsServer: []*models.IntegrationsServerPayload{
								{
									Plan: &models.IntegrationsServerPlan{
										ClusterTopology: []*models.IntegrationsServerTopologyElement{
											{
												InstanceConfigurationID:      "integrations-server-id",
												InstanceConfigurationVersion: ec.Int32(1),
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
					InstanceConfigurations: []*models.InstanceConfigurationInfo{
						buildInstanceConfiguration("es-hot"),
						buildInstanceConfiguration("es-coord"),
						buildInstanceConfiguration("es-master"),
						buildInstanceConfiguration("es-warm"),
						buildInstanceConfiguration("es-cold"),
						buildInstanceConfiguration("es-frozen"),
						buildInstanceConfiguration("es-ml"),
						buildInstanceConfiguration("kibana-id"),
						buildInstanceConfiguration("enterprise-search-id"),
						buildInstanceConfiguration("apm-id"),
						buildInstanceConfiguration("integrations-server-id"),
					},
				},
			},
			expected: []deploymentTemplateModel{
				{
					ID:   "id",
					Name: "name",
					Elasticsearch: &elasticsearchModel{
						Autoscale:        ec.Bool(true),
						HotTier:          buildTopologyModel("es-hot"),
						CoordinatingTier: buildTopologyModel("es-coord"),
						MasterTier:       buildTopologyModel("es-master"),
						WarmTier:         buildTopologyModel("es-warm"),
						ColdTier:         buildTopologyModel("es-cold"),
						FrozenTier:       buildTopologyModel("es-frozen"),
						MlTier:           buildTopologyModel("es-ml"),
					},
					Kibana:             buildStatelessModel("kibana-id"),
					EnterpriseSearch:   buildStatelessModel("enterprise-search-id"),
					Apm:                buildStatelessModel("apm-id"),
					IntegrationsServer: buildStatelessModel("integrations-server-id"),
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := mapResponseToModel(test.apiResponse, test.showDeprecated, "")
			assert.EqualValues(t, test.expected, actual)
		})
	}
}

func buildRequestTopology(id string, instanceConfigurationId string) *models.ElasticsearchClusterTopologyElement {
	return &models.ElasticsearchClusterTopologyElement{
		ID:                           id,
		InstanceConfigurationID:      instanceConfigurationId,
		InstanceConfigurationVersion: ec.Int32(1),
		Size: &models.TopologySize{
			Resource: ec.String("memory"),
			Value:    ec.Int32(2048),
		},
		AutoscalingTierOverride: ec.Bool(true),
		AutoscalingMin: &models.TopologySize{
			Resource: ec.String("memory"),
			Value:    ec.Int32(0),
		},
		AutoscalingMax: &models.TopologySize{
			Resource: ec.String("memory"),
			Value:    ec.Int32(65536),
		},
	}
}

func buildInstanceConfiguration(instanceConfigurationId string) *models.InstanceConfigurationInfo {
	return &models.InstanceConfigurationInfo{
		ID:       instanceConfigurationId,
		MaxZones: 3,
		DiscreteSizes: &models.DiscreteSizes{
			Sizes: []int32{1024, 2048, 4096},
		},
	}
}

func buildTopologyModel(instanceConfigurationId string) *topologyModel {
	return &topologyModel{
		InstanceConfigurationId:      instanceConfigurationId,
		InstanceConfigurationVersion: ec.Int32(1),
		DefaultSize:                  ec.String("2g"),
		AvailableSizes:               []string{"1g", "2g", "4g"},
		SizeResource:                 ec.String("memory"),
		Autoscaling: autoscalingModel{
			Autoscale:       ec.Bool(true),
			MaxSizeResource: ec.String("memory"),
			MaxSize:         ec.String("64g"),
			MinSizeResource: ec.String("memory"),
			MinSize:         ec.String("0g"),
		},
	}
}

func buildStatelessModel(instanceConfigurationId string) *statelessModel {
	return &statelessModel{
		InstanceConfigurationId:      instanceConfigurationId,
		InstanceConfigurationVersion: ec.Int32(1),
		DefaultSize:                  ec.String("1g"),
		AvailableSizes:               []string{"1g", "2g", "4g"},
		SizeResource:                 ec.String("memory"),
	}
}
