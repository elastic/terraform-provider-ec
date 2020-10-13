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
	"errors"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_modelToState(t *testing.T) {
	deploymentSchemaArg := schema.TestResourceDataRaw(t, newSchema(), nil)
	deploymentSchemaArg.SetId(mock.ValidClusterID)

	wantDeployment := util.NewResourceData(t, util.ResDataParams{
		ID:        mock.ValidClusterID,
		Resources: newSampleDeployment(),
		Schema:    newSchema(),
	})

	legacyRes := openDeploymentGet(t, "testdata/deployment_legacy_response.json")
	legacyRD := schema.TestResourceDataRaw(t, newSchema(), nil)
	legacyRD.SetId(mock.ValidClusterID)
	wantLegacyDeployment := util.NewResourceData(t, util.ResDataParams{
		ID: mock.ValidClusterID,
		Resources: map[string]interface{}{
			"deployment_template_id": "default",
			"id":                     "320b7b540dfc967a7a649c18e2fce4ed",
			"name":                   "ear",
			"region":                 "us-east-1",
			"version":                "2.4.5",
			"elasticsearch": []interface{}{map[string]interface{}{
				"cloud_id":       "ear:somecloudID",
				"http_endpoint":  "http://122c96c491b3d5e10e147463927a5349.us-east-1.aws.found.io:9200",
				"https_endpoint": "https://122c96c491b3d5e10e147463927a5349.us-east-1.aws.found.io:9243",
				"ref_id":         "elasticsearch",
				"region":         "us-east-1",
				"resource_id":    "122c96c491b3d5e10e147463927a5349",
				"version":        "2.4.5",
				"topology": []interface{}{map[string]interface{}{
					"instance_configuration_id": "",
					"node_type_data":            true,
					"node_type_ingest":          false,
					"node_type_master":          true,
					"size":                      "2g",
					"size_resource":             "memory",
					"zone_count":                2,
				}},
			}},
			"kibana": []interface{}{map[string]interface{}{
				"elasticsearch_cluster_ref_id": "elasticsearch",
				"ref_id":                       "kibana",
				"region":                       "us-east-1",
				"resource_id":                  "b211f9ef4af84d78851ddf79f439ad8d",
				"version":                      "4.6.6",
				"topology": []interface{}{map[string]interface{}{
					"size":          "1g",
					"size_resource": "memory",
					"zone_count":    1,
				}},
			}},
		},
		Schema: newSchema(),
	})

	legacyConvertedRes := openDeploymentGet(t, "testdata/deployment_legacy_converted_response.json")
	legacyConvertedRD := schema.TestResourceDataRaw(t, newSchema(), nil)
	legacyConvertedRD.SetId(mock.ValidClusterID)
	wantLegacyConvertedDeployment := util.NewResourceData(t, util.ResDataParams{
		ID: mock.ValidClusterID,
		Resources: map[string]interface{}{
			"deployment_template_id": "default",
			"id":                     "320b7b540dfc967a7a649c18e2fce4ed",
			"name":                   "ear",
			"region":                 "us-east-1",
			"version":                "2.4.5",
			"elasticsearch": []interface{}{map[string]interface{}{
				"cloud_id":       "ear:somecloudID",
				"http_endpoint":  "http://122c96c491b3d5e10e147463927a5349.us-east-1.aws.found.io:9200",
				"https_endpoint": "https://122c96c491b3d5e10e147463927a5349.us-east-1.aws.found.io:9243",
				"ref_id":         "elasticsearch",
				"region":         "us-east-1",
				"resource_id":    "122c96c491b3d5e10e147463927a5349",
				"version":        "2.4.5",
				"topology": []interface{}{
					map[string]interface{}{
						"instance_configuration_id": "aws.highio.classic",
						"node_type_data":            true,
						"node_type_ingest":          false,
						"node_type_master":          true,
						"size":                      "2g",
						"size_resource":             "memory",
						"zone_count":                2,
					},
					map[string]interface{}{
						"instance_configuration_id": "aws.master.classic",
						"node_type_data":            false,
						"node_type_ingest":          false,
						"node_type_master":          true,
						"size":                      "1g",
						"size_resource":             "memory",
						"zone_count":                1,
					},
				},
			}},
			"kibana": []interface{}{map[string]interface{}{
				"elasticsearch_cluster_ref_id": "elasticsearch",
				"ref_id":                       "kibana",
				"region":                       "us-east-1",
				"resource_id":                  "b211f9ef4af84d78851ddf79f439ad8d",
				"version":                      "4.6.6",
				"topology": []interface{}{map[string]interface{}{
					"instance_configuration_id": "aws.kibana.classic",
					"size":                      "1g",
					"size_resource":             "memory",
					"zone_count":                1,
				}},
			}},
		},
		Schema: newSchema(),
	})

	azureIOOptimizedRes := openDeploymentGet(t, "testdata/deployment-azure-io-optimized.json")
	azureIOOptimizedRD := schema.TestResourceDataRaw(t, newSchema(), nil)
	azureIOOptimizedRD.SetId(mock.ValidClusterID)
	wantAzureIOOptimizedDeployment := util.NewResourceData(t, util.ResDataParams{
		ID: mock.ValidClusterID,
		Resources: map[string]interface{}{
			"deployment_template_id": "azure-io-optimized",
			"id":                     "123b7b540dfc967a7a649c18e2fce4ed",
			"name":                   "up2d",
			"region":                 "azure-eastus2",
			"version":                "7.9.2",
			"apm": []interface{}{map[string]interface{}{
				"elasticsearch_cluster_ref_id": "main-elasticsearch",
				"ref_id":                       "main-apm",
				"region":                       "azure-eastus2",
				"resource_id":                  "1235d8c911b74dd6a03c2a7b37fd68ab",
				"version":                      "7.9.2",
				"http_endpoint":                "http://1235d8c911b74dd6a03c2a7b37fd68ab.apm.eastus2.azure.elastic-cloud.com:9200",
				"https_endpoint":               "https://1235d8c911b74dd6a03c2a7b37fd68ab.apm.eastus2.azure.elastic-cloud.com:443",
				"topology": []interface{}{map[string]interface{}{
					"instance_configuration_id": "azure.apm.e32sv3",
					"size":                      "0.5g",
					"size_resource":             "memory",
					"zone_count":                1,
					"config": []interface{}{map[string]interface{}{
						"debug_enabled":               false,
						"user_settings_json":          "",
						"user_settings_override_json": "",
						"user_settings_override_yaml": "",
						"user_settings_yaml":          "",
					}},
				}},
			}},
			"elasticsearch": []interface{}{map[string]interface{}{
				"cloud_id":       "up2d:somecloudID",
				"http_endpoint":  "http://1238f19957874af69306787dca662154.eastus2.azure.elastic-cloud.com:9200",
				"https_endpoint": "https://1238f19957874af69306787dca662154.eastus2.azure.elastic-cloud.com:9243",
				"ref_id":         "main-elasticsearch",
				"region":         "azure-eastus2",
				"resource_id":    "1238f19957874af69306787dca662154",
				"version":        "7.9.2",
				"topology": []interface{}{
					map[string]interface{}{
						"instance_configuration_id": "azure.data.highio.l32sv2",
						"node_type_data":            true,
						"node_type_ingest":          true,
						"node_type_master":          true,
						"size":                      "4g",
						"size_resource":             "memory",
						"zone_count":                2,
					},
				},
			}},
			"kibana": []interface{}{map[string]interface{}{
				"elasticsearch_cluster_ref_id": "main-elasticsearch",
				"ref_id":                       "main-kibana",
				"region":                       "azure-eastus2",
				"resource_id":                  "1235cd4a4c7f464bbcfd795f3638b769",
				"version":                      "7.9.2",
				"http_endpoint":                "http://1235cd4a4c7f464bbcfd795f3638b769.eastus2.azure.elastic-cloud.com:9200",
				"https_endpoint":               "https://1235cd4a4c7f464bbcfd795f3638b769.eastus2.azure.elastic-cloud.com:9243",
				"topology": []interface{}{map[string]interface{}{
					"instance_configuration_id": "azure.kibana.e32sv3",
					"size":                      "1g",
					"size_resource":             "memory",
					"zone_count":                1,
				}},
			}},
		},
		Schema: newSchema(),
	})

	type args struct {
		d   *schema.ResourceData
		res *models.DeploymentGetResponse
	}
	tests := []struct {
		name string
		args args
		want *schema.ResourceData
		err  error
	}{
		{
			name: "flattens deployment resources",
			want: wantDeployment,
			args: args{
				d: deploymentSchemaArg,
				res: &models.DeploymentGetResponse{
					Name: ec.String("my_deployment_name"),
					Settings: &models.DeploymentSettings{
						TrafficFilterSettings: &models.TrafficFilterSettings{
							Rulesets: []string{"0.0.0.0/0", "192.168.10.0/24"},
						},
					},
					Resources: &models.DeploymentResources{
						Elasticsearch: []*models.ElasticsearchResourceInfo{
							{
								Region: ec.String("us-east-1"),
								RefID:  ec.String("main-elasticsearch"),
								Info: &models.ElasticsearchClusterInfo{
									Status:      ec.String("started"),
									ClusterID:   &mock.ValidClusterID,
									ClusterName: ec.String("some-name"),
									Region:      "us-east-1",
									ElasticsearchMonitoringInfo: &models.ElasticsearchMonitoringInfo{
										DestinationClusterIds: []string{"some"},
									},
									PlanInfo: &models.ElasticsearchClusterPlansInfo{
										Current: &models.ElasticsearchClusterPlanInfo{
											Plan: &models.ElasticsearchClusterPlan{
												Elasticsearch: &models.ElasticsearchConfiguration{
													Version: "7.7.0",
												},
												DeploymentTemplate: &models.DeploymentTemplateReference{
													ID: ec.String("aws-io-optimized-v2"),
												},
												ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
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
													Elasticsearch: &models.ElasticsearchConfiguration{
														UserSettingsYaml:         `some.setting: value`,
														UserSettingsOverrideYaml: `some.setting: value2`,
														UserSettingsJSON:         `{"some.setting": "value"}`,
														UserSettingsOverrideJSON: `{"some.setting": "value2"}`,
													},
												}},
											},
										},
									},
								},
							},
						},
						Kibana: []*models.KibanaResourceInfo{
							{
								Region:                    ec.String("us-east-1"),
								RefID:                     ec.String("main-kibana"),
								ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
								Info: &models.KibanaClusterInfo{
									Status:      ec.String("started"),
									ClusterID:   &mock.ValidClusterID,
									ClusterName: ec.String("some-kibana-name"),
									Region:      "us-east-1",
									PlanInfo: &models.KibanaClusterPlansInfo{
										Current: &models.KibanaClusterPlanInfo{
											Plan: &models.KibanaClusterPlan{
												Kibana: &models.KibanaConfiguration{
													Version: "7.7.0",
												},
												ClusterTopology: []*models.KibanaClusterTopologyElement{
													{
														ZoneCount:               1,
														InstanceConfigurationID: "aws.kibana.r5d",
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
						},
						Apm: []*models.ApmResourceInfo{{
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-apm"),
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Info: &models.ApmInfo{
								Status: ec.String("started"),
								ID:     &mock.ValidClusterID,
								Name:   ec.String("some-apm-name"),
								Region: "us-east-1",
								PlanInfo: &models.ApmPlansInfo{
									Current: &models.ApmPlanInfo{
										Plan: &models.ApmPlan{
											Apm: &models.ApmConfiguration{
												Version: "7.7.0",
												SystemSettings: &models.ApmSystemSettings{
													DebugEnabled: ec.Bool(false),
												},
											},
											ClusterTopology: []*models.ApmTopologyElement{{
												ZoneCount:               1,
												InstanceConfigurationID: "aws.apm.r5d",
												Size: &models.TopologySize{
													Resource: ec.String("memory"),
													Value:    ec.Int32(512),
												},
												Apm: &models.ApmConfiguration{
													SystemSettings: &models.ApmSystemSettings{
														DebugEnabled: ec.Bool(false),
													},
												},
											}},
										},
									},
								},
							},
						}},
						EnterpriseSearch: []*models.EnterpriseSearchResourceInfo{
							{
								Region:                    ec.String("us-east-1"),
								RefID:                     ec.String("main-enterprise_search"),
								ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
								Info: &models.EnterpriseSearchInfo{
									Status: ec.String("started"),
									ID:     &mock.ValidClusterID,
									Name:   ec.String("some-enterprise_search-name"),
									Region: "us-east-1",
									PlanInfo: &models.EnterpriseSearchPlansInfo{
										Current: &models.EnterpriseSearchPlanInfo{
											Plan: &models.EnterpriseSearchPlan{
												EnterpriseSearch: &models.EnterpriseSearchConfiguration{
													Version: "7.7.0",
												},
												ClusterTopology: []*models.EnterpriseSearchTopologyElement{
													{
														ZoneCount:               1,
														InstanceConfigurationID: "aws.enterprisesearch.m5d",
														Size: &models.TopologySize{
															Resource: ec.String("memory"),
															Value:    ec.Int32(2048),
														},
														NodeType: &models.EnterpriseSearchNodeTypes{
															Appserver: ec.Bool(true),
															Connector: ec.Bool(true),
															Worker:    ec.Bool(true),
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "flattens a legacy plan (non converted)",
			args: args{d: legacyRD, res: legacyRes},
			want: wantLegacyDeployment,
		},
		{
			name: "flattens a legacy plan (converted)",
			args: args{d: legacyConvertedRD, res: legacyConvertedRes},
			want: wantLegacyConvertedDeployment,
		},
		{
			name: "flattens an azure plan (io-optimized)",
			args: args{d: legacyConvertedRD, res: azureIOOptimizedRes},
			want: wantAzureIOOptimizedDeployment,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := modelToState(tt.args.d, tt.args.res)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}

			var wantState interface{}
			if tt.want != nil {
				wantState = tt.want.State().Attributes
			}

			assert.Equal(t, wantState, tt.args.d.State().Attributes)
		})
	}
}

func Test_getDeploymentTemplateID(t *testing.T) {
	type args struct {
		res *models.DeploymentResources
	}
	tests := []struct {
		name string
		args args
		want string
		err  error
	}{
		{
			name: "empty resources returns an error",
			args: args{res: &models.DeploymentResources{}},
			err:  errors.New("failed to obtain the deployment template id"),
		},
		{
			name: "single empty current plan returns error",
			args: args{res: &models.DeploymentResources{
				Elasticsearch: []*models.ElasticsearchResourceInfo{
					{
						Info: &models.ElasticsearchClusterInfo{
							PlanInfo: &models.ElasticsearchClusterPlansInfo{
								Pending: &models.ElasticsearchClusterPlanInfo{
									Plan: &models.ElasticsearchClusterPlan{
										DeploymentTemplate: &models.DeploymentTemplateReference{
											ID: ec.String("aws-io-optimized"),
										},
									},
								},
							},
						},
					},
				},
			}},
			err: errors.New("failed to obtain the deployment template id"),
		},
		{
			name: "multiple deployment templates returns an error",
			args: args{res: &models.DeploymentResources{
				Elasticsearch: []*models.ElasticsearchResourceInfo{
					{
						Info: &models.ElasticsearchClusterInfo{
							PlanInfo: &models.ElasticsearchClusterPlansInfo{
								Current: &models.ElasticsearchClusterPlanInfo{
									Plan: &models.ElasticsearchClusterPlan{
										DeploymentTemplate: &models.DeploymentTemplateReference{
											ID: ec.String("someid"),
										},
									},
								},
							},
						},
					},
					{
						Info: &models.ElasticsearchClusterInfo{
							PlanInfo: &models.ElasticsearchClusterPlansInfo{
								Current: &models.ElasticsearchClusterPlanInfo{
									Plan: &models.ElasticsearchClusterPlan{
										DeploymentTemplate: &models.DeploymentTemplateReference{
											ID: ec.String("someotherid"),
										},
									},
								},
							},
						},
					},
				},
			}},
			err: errors.New("there are more than 1 deployment templates specified on the deployment: \"someid, someotherid\""),
		},
		{
			name: "single deployment template returns it",
			args: args{res: &models.DeploymentResources{
				Elasticsearch: []*models.ElasticsearchResourceInfo{
					{
						Info: &models.ElasticsearchClusterInfo{
							PlanInfo: &models.ElasticsearchClusterPlansInfo{
								Current: &models.ElasticsearchClusterPlanInfo{
									Plan: &models.ElasticsearchClusterPlan{
										DeploymentTemplate: &models.DeploymentTemplateReference{
											ID: ec.String("aws-io-optimized"),
										},
									},
								},
							},
						},
					},
				},
			}},
			want: "aws-io-optimized",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getDeploymentTemplateID(tt.args.res)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_parseCredentials(t *testing.T) {
	deploymentRD := util.NewResourceData(t, util.ResDataParams{
		ID:        mock.ValidClusterID,
		Resources: newSampleDeployment(),
		Schema:    newSchema(),
	})

	rawData := newSampleDeployment()
	rawData["elasticsearch_username"] = "my-username"
	rawData["elasticsearch_password"] = "my-password"
	rawData["apm_secret_token"] = "some-secret-token"

	wantDeploymentRD := util.NewResourceData(t, util.ResDataParams{
		ID:        mock.ValidClusterID,
		Resources: rawData,
		Schema:    newSchema(),
	})

	type args struct {
		d         *schema.ResourceData
		resources []*models.DeploymentResource
	}
	tests := []struct {
		name string
		args args
		want *schema.ResourceData
		err  error
	}{
		{
			name: "Parses credentials",
			args: args{
				d: deploymentRD,
				resources: []*models.DeploymentResource{{
					Credentials: &models.ClusterCredentials{
						Username: ec.String("my-username"),
						Password: ec.String("my-password"),
					},
					SecretToken: "some-secret-token",
				}},
			},
			want: wantDeploymentRD,
		},
		{
			name: "when no credentials are passed, it doesn't overwrite them",
			args: args{
				d: util.NewResourceData(t, util.ResDataParams{
					ID:        mock.ValidClusterID,
					Resources: rawData,
					Schema:    newSchema(),
				}),
				resources: []*models.DeploymentResource{
					{},
				},
			},
			want: wantDeploymentRD,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseCredentials(tt.args.d, tt.args.resources)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want.State().Attributes, tt.args.d.State().Attributes)
		})
	}
}
