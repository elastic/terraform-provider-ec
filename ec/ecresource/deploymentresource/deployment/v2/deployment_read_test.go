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
	"errors"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	apmv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/apm/v2"
	elasticsearchv1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v1"
	elasticsearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
	enterprisesearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/enterprisesearch/v2"
	kibanav2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/kibana/v2"
	observabilityv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/observability/v2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_readDeployment(t *testing.T) {
	type args struct {
		res     *models.DeploymentGetResponse
		remotes models.RemoteResources
	}
	tests := []struct {
		name string
		args args
		want Deployment
		err  error
	}{
		{
			name: "flattens deployment resources",
			want: Deployment{
				Id:                   mock.ValidClusterID,
				Alias:                "my-deployment",
				Name:                 "my_deployment_name",
				DeploymentTemplateId: "aws-io-optimized-v2",
				Region:               "us-east-1",
				Version:              "7.7.0",
				Elasticsearch: &elasticsearchv2.Elasticsearch{
					RefId:         ec.String("main-elasticsearch"),
					ResourceId:    &mock.ValidClusterID,
					Region:        ec.String("us-east-1"),
					TrustAccount:  elasticsearchv2.ElasticsearchTrustAccounts{},
					TrustExternal: elasticsearchv2.ElasticsearchTrustExternals{},
					Config: &elasticsearchv2.ElasticsearchConfig{
						Plugins:                  []string{},
						UserSettingsYaml:         ec.String("some.setting: value"),
						UserSettingsOverrideYaml: ec.String("some.setting: value2"),
						UserSettingsJson:         ec.String("{\"some.setting\":\"value\"}"),
						UserSettingsOverrideJson: ec.String("{\"some.setting\":\"value2\"}"),
					},
					HotTier: elasticsearchv2.CreateTierForTest(
						"hot_content",
						elasticsearchv2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("aws.data.highio.i3"),
							Size:                    ec.String("2g"),
							SizeResource:            ec.String("memory"),
							NodeTypeData:            ec.String("true"),
							NodeTypeIngest:          ec.String("true"),
							NodeTypeMaster:          ec.String("true"),
							NodeTypeMl:              ec.String("false"),
							ZoneCount:               1,
							Autoscaling:             &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
				},
				Kibana: &kibanav2.Kibana{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-kibana"),
					ResourceId:                ec.String(mock.ValidClusterID),
					Region:                    ec.String("us-east-1"),
					InstanceConfigurationId:   ec.String("aws.kibana.r5d"),
					Size:                      ec.String("1g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
				Apm: &apmv2.Apm{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-apm"),
					ResourceId:                ec.String(mock.ValidClusterID),
					Region:                    ec.String("us-east-1"),
					Config: &apmv2.ApmConfig{
						DebugEnabled: ec.Bool(false),
					},
					InstanceConfigurationId: ec.String("aws.apm.r5d"),
					Size:                    ec.String("0.5g"),
					SizeResource:            ec.String("memory"),
					ZoneCount:               1,
				},
				EnterpriseSearch: &enterprisesearchv2.EnterpriseSearch{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-enterprise_search"),
					ResourceId:                ec.String(mock.ValidClusterID),
					Region:                    ec.String("us-east-1"),
					InstanceConfigurationId:   ec.String("aws.enterprisesearch.m5d"),
					Size:                      ec.String("2g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
					NodeTypeAppserver:         ec.Bool(true),
					NodeTypeConnector:         ec.Bool(true),
					NodeTypeWorker:            ec.Bool(true),
				},
				Observability: &observabilityv2.Observability{
					DeploymentId: ec.String(mock.ValidClusterID),
					RefId:        ec.String("main-elasticsearch"),
					Logs:         true,
					Metrics:      true,
				},
				TrafficFilter: []string{"0.0.0.0/0", "192.168.10.0/24"},
			},
			args: args{
				res: &models.DeploymentGetResponse{
					ID:    &mock.ValidClusterID,
					Alias: "my-deployment",
					Name:  ec.String("my_deployment_name"),
					Settings: &models.DeploymentSettings{
						TrafficFilterSettings: &models.TrafficFilterSettings{
							Rulesets: []string{"0.0.0.0/0", "192.168.10.0/24"},
						},
						Observability: &models.DeploymentObservabilitySettings{
							Logging: &models.DeploymentLoggingSettings{
								Destination: &models.ObservabilityAbsoluteDeployment{
									DeploymentID: &mock.ValidClusterID,
									RefID:        "main-elasticsearch",
								},
							},
							Metrics: &models.DeploymentMetricsSettings{
								Destination: &models.ObservabilityAbsoluteDeployment{
									DeploymentID: &mock.ValidClusterID,
									RefID:        "main-elasticsearch",
								},
							},
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
													Version:                  "7.7.0",
													UserSettingsYaml:         `some.setting: value`,
													UserSettingsOverrideYaml: `some.setting: value2`,
													UserSettingsJSON: map[string]interface{}{
														"some.setting": "value",
													},
													UserSettingsOverrideJSON: map[string]interface{}{
														"some.setting": "value2",
													},
												},
												DeploymentTemplate: &models.DeploymentTemplateReference{
													ID: ec.String("aws-io-optimized-v2"),
												},
												ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
													ID: "hot_content",
													Elasticsearch: &models.ElasticsearchConfiguration{
														NodeAttributes: map[string]string{"data": "hot"},
													},
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
													TopologyElementControl: &models.TopologyElementControl{
														Min: &models.TopologySize{
															Resource: ec.String("memory"),
															Value:    ec.Int32(1024),
														},
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
			name: "sets the global version to the lesser version",
			args: args{
				res: &models.DeploymentGetResponse{
					ID:    &mock.ValidClusterID,
					Alias: "my-deployment",
					Name:  ec.String("my_deployment_name"),
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
													Version:                  "7.7.0",
													UserSettingsYaml:         `some.setting: value`,
													UserSettingsOverrideYaml: `some.setting: value2`,
													UserSettingsJSON: map[string]interface{}{
														"some.setting": "value",
													},
													UserSettingsOverrideJSON: map[string]interface{}{
														"some.setting": "value2",
													},
												},
												DeploymentTemplate: &models.DeploymentTemplateReference{
													ID: ec.String("aws-io-optimized-v2"),
												},
												ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
													ID: "hot_content",
													Elasticsearch: &models.ElasticsearchConfiguration{
														NodeAttributes: map[string]string{"data": "hot"},
													},
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
													TopologyElementControl: &models.TopologyElementControl{
														Min: &models.TopologySize{
															Resource: ec.String("memory"),
															Value:    ec.Int32(1024),
														},
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
													Version: "7.6.2",
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
					},
				},
			},
			want: Deployment{
				Id:                   mock.ValidClusterID,
				Alias:                "my-deployment",
				Name:                 "my_deployment_name",
				DeploymentTemplateId: "aws-io-optimized-v2",
				Region:               "us-east-1",
				Version:              "7.6.2",
				Elasticsearch: &elasticsearchv2.Elasticsearch{
					RefId:         ec.String("main-elasticsearch"),
					ResourceId:    &mock.ValidClusterID,
					Region:        ec.String("us-east-1"),
					TrustAccount:  elasticsearchv2.ElasticsearchTrustAccounts{},
					TrustExternal: elasticsearchv2.ElasticsearchTrustExternals{},
					Config: &elasticsearchv2.ElasticsearchConfig{
						Plugins:                  []string{},
						UserSettingsYaml:         ec.String("some.setting: value"),
						UserSettingsOverrideYaml: ec.String("some.setting: value2"),
						UserSettingsJson:         ec.String("{\"some.setting\":\"value\"}"),
						UserSettingsOverrideJson: ec.String("{\"some.setting\":\"value2\"}"),
					},
					HotTier: elasticsearchv2.CreateTierForTest(
						"hot_content",
						elasticsearchv2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("aws.data.highio.i3"),
							Size:                    ec.String("2g"),
							SizeResource:            ec.String("memory"),
							NodeTypeData:            ec.String("true"),
							NodeTypeIngest:          ec.String("true"),
							NodeTypeMaster:          ec.String("true"),
							NodeTypeMl:              ec.String("false"),
							ZoneCount:               1,
							Autoscaling:             &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
				},
				Kibana: &kibanav2.Kibana{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-kibana"),
					ResourceId:                ec.String(mock.ValidClusterID),
					Region:                    ec.String("us-east-1"),
					InstanceConfigurationId:   ec.String("aws.kibana.r5d"),
					Size:                      ec.String("1g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
				TrafficFilter: []string{"0.0.0.0/0", "192.168.10.0/24"},
			},
		},

		{
			name: "flattens an azure plan (io-optimized)",
			args: args{
				res: deploymentGetResponseFromFile(t, "../../testdata/deployment-azure-io-optimized.json"),
			},
			want: Deployment{
				Id:                   "123e79d8109c4a0790b0b333110bf715",
				Alias:                "my-deployment",
				Name:                 "up2d",
				DeploymentTemplateId: "azure-io-optimized",
				Region:               "azure-eastus2",
				Version:              "7.9.2",
				Elasticsearch: &elasticsearchv2.Elasticsearch{
					RefId:         ec.String("main-elasticsearch"),
					ResourceId:    ec.String("1238f19957874af69306787dca662154"),
					Region:        ec.String("azure-eastus2"),
					Autoscale:     ec.Bool(false),
					CloudID:       ec.String("up2d:somecloudID"),
					TrustAccount:  elasticsearchv2.ElasticsearchTrustAccounts{},
					TrustExternal: elasticsearchv2.ElasticsearchTrustExternals{},
					HttpEndpoint:  ec.String("http://1238f19957874af69306787dca662154.eastus2.azure.elastic-cloud.com:9200"),
					HttpsEndpoint: ec.String("https://1238f19957874af69306787dca662154.eastus2.azure.elastic-cloud.com:9243"),
					HotTier: elasticsearchv2.CreateTierForTest(
						"hot_content",
						elasticsearchv2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("azure.data.highio.l32sv2"),
							Size:                    ec.String("4g"),
							SizeResource:            ec.String("memory"),
							NodeTypeData:            ec.String("true"),
							NodeTypeIngest:          ec.String("true"),
							NodeTypeMaster:          ec.String("true"),
							NodeTypeMl:              ec.String("false"),
							ZoneCount:               2,
							Autoscaling:             &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
					Config: &elasticsearchv2.ElasticsearchConfig{
						Plugins: []string{},
					},
				},
				Kibana: &kibanav2.Kibana{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-kibana"),
					ResourceId:                ec.String("1235cd4a4c7f464bbcfd795f3638b769"),
					Region:                    ec.String("azure-eastus2"),
					HttpEndpoint:              ec.String("http://1235cd4a4c7f464bbcfd795f3638b769.eastus2.azure.elastic-cloud.com:9200"),
					HttpsEndpoint:             ec.String("https://1235cd4a4c7f464bbcfd795f3638b769.eastus2.azure.elastic-cloud.com:9243"),
					InstanceConfigurationId:   ec.String("azure.kibana.e32sv3"),
					Size:                      ec.String("1g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
				Apm: &apmv2.Apm{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-apm"),
					ResourceId:                ec.String("1235d8c911b74dd6a03c2a7b37fd68ab"),
					Region:                    ec.String("azure-eastus2"),
					HttpEndpoint:              ec.String("http://1235d8c911b74dd6a03c2a7b37fd68ab.apm.eastus2.azure.elastic-cloud.com:9200"),
					HttpsEndpoint:             ec.String("https://1235d8c911b74dd6a03c2a7b37fd68ab.apm.eastus2.azure.elastic-cloud.com:443"),
					InstanceConfigurationId:   ec.String("azure.apm.e32sv3"),
					Size:                      ec.String("0.5g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
			},
		},

		{
			name: "flattens an aws plan (io-optimized)",
			args: args{res: deploymentGetResponseFromFile(t, "../../testdata/deployment-aws-io-optimized.json")},
			want: Deployment{
				Id:                   "123365f2805e46808d40849b1c0b266b",
				Alias:                "my-deployment",
				Name:                 "up2d",
				DeploymentTemplateId: "aws-io-optimized-v2",
				Region:               "aws-eu-central-1",
				Version:              "7.9.2",
				Elasticsearch: &elasticsearchv2.Elasticsearch{
					RefId:         ec.String("main-elasticsearch"),
					ResourceId:    ec.String("1239f7ee7196439ba2d105319ac5eba7"),
					Region:        ec.String("aws-eu-central-1"),
					Autoscale:     ec.Bool(false),
					CloudID:       ec.String("up2d:someCloudID"),
					TrustAccount:  elasticsearchv2.ElasticsearchTrustAccounts{},
					TrustExternal: elasticsearchv2.ElasticsearchTrustExternals{},
					HttpEndpoint:  ec.String("http://1239f7ee7196439ba2d105319ac5eba7.eu-central-1.aws.cloud.es.io:9200"),
					HttpsEndpoint: ec.String("https://1239f7ee7196439ba2d105319ac5eba7.eu-central-1.aws.cloud.es.io:9243"),
					Config: &elasticsearchv2.ElasticsearchConfig{
						Plugins: []string{},
					},
					HotTier: elasticsearchv2.CreateTierForTest(
						"hot_content",
						elasticsearchv2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("aws.data.highio.i3"),
							Size:                    ec.String("8g"),
							SizeResource:            ec.String("memory"),
							NodeTypeData:            ec.String("true"),
							NodeTypeIngest:          ec.String("true"),
							NodeTypeMaster:          ec.String("true"),
							NodeTypeMl:              ec.String("false"),
							ZoneCount:               2,
							Autoscaling:             &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
				},
				Kibana: &kibanav2.Kibana{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-kibana"),
					ResourceId:                ec.String("123dcfda06254ca789eb287e8b73ff4c"),
					Region:                    ec.String("aws-eu-central-1"),
					HttpEndpoint:              ec.String("http://123dcfda06254ca789eb287e8b73ff4c.eu-central-1.aws.cloud.es.io:9200"),
					HttpsEndpoint:             ec.String("https://123dcfda06254ca789eb287e8b73ff4c.eu-central-1.aws.cloud.es.io:9243"),
					InstanceConfigurationId:   ec.String("aws.kibana.r5d"),
					Size:                      ec.String("1g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
				Apm: &apmv2.Apm{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-apm"),
					ResourceId:                ec.String("12328579b3bf40c8b58c1a0ed5a4bd8b"),
					Region:                    ec.String("aws-eu-central-1"),
					HttpEndpoint:              ec.String("http://12328579b3bf40c8b58c1a0ed5a4bd8b.apm.eu-central-1.aws.cloud.es.io:80"),
					HttpsEndpoint:             ec.String("https://12328579b3bf40c8b58c1a0ed5a4bd8b.apm.eu-central-1.aws.cloud.es.io:443"),
					InstanceConfigurationId:   ec.String("aws.apm.r5d"),
					Size:                      ec.String("0.5g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
			},
		},

		{
			name: "flattens an aws plan with extensions (io-optimized)",
			args: args{
				res: deploymentGetResponseFromFile(t, "../../testdata/deployment-aws-io-optimized-extension.json"),
			},
			want: Deployment{
				Id:                   "123365f2805e46808d40849b1c0b266b",
				Alias:                "my-deployment",
				Name:                 "up2d",
				DeploymentTemplateId: "aws-io-optimized-v2",
				Region:               "aws-eu-central-1",
				Version:              "7.9.2",
				Elasticsearch: &elasticsearchv2.Elasticsearch{
					RefId:         ec.String("main-elasticsearch"),
					ResourceId:    ec.String("1239f7ee7196439ba2d105319ac5eba7"),
					Region:        ec.String("aws-eu-central-1"),
					Autoscale:     ec.Bool(false),
					CloudID:       ec.String("up2d:someCloudID"),
					TrustAccount:  elasticsearchv2.ElasticsearchTrustAccounts{},
					TrustExternal: elasticsearchv2.ElasticsearchTrustExternals{},
					HttpEndpoint:  ec.String("http://1239f7ee7196439ba2d105319ac5eba7.eu-central-1.aws.cloud.es.io:9200"),
					HttpsEndpoint: ec.String("https://1239f7ee7196439ba2d105319ac5eba7.eu-central-1.aws.cloud.es.io:9243"),
					Config: &elasticsearchv2.ElasticsearchConfig{
						Plugins: []string{},
					},
					HotTier: elasticsearchv2.CreateTierForTest(
						"hot_content",
						elasticsearchv2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("aws.data.highio.i3"),
							Size:                    ec.String("8g"),
							SizeResource:            ec.String("memory"),
							NodeTypeData:            ec.String("true"),
							NodeTypeIngest:          ec.String("true"),
							NodeTypeMaster:          ec.String("true"),
							NodeTypeMl:              ec.String("false"),
							ZoneCount:               2,
							Autoscaling:             &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
					Extension: elasticsearchv2.ElasticsearchExtensions{
						{
							Name:    "custom-bundle",
							Version: "7.9.2",
							Url:     "http://12345",
							Type:    "bundle",
						},
						{
							Name:    "custom-bundle2",
							Version: "7.9.2",
							Url:     "http://123456",
							Type:    "bundle",
						},
						{
							Name:    "custom-plugin",
							Version: "7.9.2",
							Url:     "http://12345",
							Type:    "plugin",
						},
						{
							Name:    "custom-plugin2",
							Version: "7.9.2",
							Url:     "http://123456",
							Type:    "plugin",
						},
					},
				},
				Kibana: &kibanav2.Kibana{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-kibana"),
					ResourceId:                ec.String("123dcfda06254ca789eb287e8b73ff4c"),
					Region:                    ec.String("aws-eu-central-1"),
					HttpEndpoint:              ec.String("http://123dcfda06254ca789eb287e8b73ff4c.eu-central-1.aws.cloud.es.io:9200"),
					HttpsEndpoint:             ec.String("https://123dcfda06254ca789eb287e8b73ff4c.eu-central-1.aws.cloud.es.io:9243"),
					InstanceConfigurationId:   ec.String("aws.kibana.r5d"),
					Size:                      ec.String("1g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
				Apm: &apmv2.Apm{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-apm"),
					ResourceId:                ec.String("12328579b3bf40c8b58c1a0ed5a4bd8b"),
					Region:                    ec.String("aws-eu-central-1"),
					HttpEndpoint:              ec.String("http://12328579b3bf40c8b58c1a0ed5a4bd8b.apm.eu-central-1.aws.cloud.es.io:80"),
					HttpsEndpoint:             ec.String("https://12328579b3bf40c8b58c1a0ed5a4bd8b.apm.eu-central-1.aws.cloud.es.io:443"),
					InstanceConfigurationId:   ec.String("aws.apm.r5d"),
					Size:                      ec.String("0.5g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
			},
		},

		{
			name: "flattens an aws plan with trusts",
			args: args{
				res: &models.DeploymentGetResponse{
					ID:    ec.String("123b7b540dfc967a7a649c18e2fce4ed"),
					Alias: "OH",
					Name:  ec.String("up2d"),
					Resources: &models.DeploymentResources{
						Elasticsearch: []*models.ElasticsearchResourceInfo{{
							RefID:  ec.String("main-elasticsearch"),
							Region: ec.String("aws-eu-central-1"),
							Info: &models.ElasticsearchClusterInfo{
								Status: ec.String("running"),
								PlanInfo: &models.ElasticsearchClusterPlansInfo{
									Current: &models.ElasticsearchClusterPlanInfo{
										Plan: &models.ElasticsearchClusterPlan{
											DeploymentTemplate: &models.DeploymentTemplateReference{
												ID: ec.String("aws-io-optimized-v2"),
											},
											Elasticsearch: &models.ElasticsearchConfiguration{
												Version: "7.13.1",
											},
											ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
												ID: "hot_content",
												Size: &models.TopologySize{
													Value:    ec.Int32(4096),
													Resource: ec.String("memory"),
												},
											}},
										},
									},
								},
								Settings: &models.ElasticsearchClusterSettings{
									Trust: &models.ElasticsearchClusterTrustSettings{
										Accounts: []*models.AccountTrustRelationship{
											{
												AccountID: ec.String("ANID"),
												TrustAll:  ec.Bool(true),
											},
											{
												AccountID: ec.String("anotherID"),
												TrustAll:  ec.Bool(false),
												TrustAllowlist: []string{
													"abc", "dfg", "hij",
												},
											},
										},
										External: []*models.ExternalTrustRelationship{
											{
												TrustRelationshipID: ec.String("external_id"),
												TrustAll:            ec.Bool(true),
											},
											{
												TrustRelationshipID: ec.String("another_external_id"),
												TrustAll:            ec.Bool(false),
												TrustAllowlist: []string{
													"abc", "dfg",
												},
											},
										},
									},
								},
							},
						}},
					},
				},
			},
			want: Deployment{
				Id:                   "123b7b540dfc967a7a649c18e2fce4ed",
				Alias:                "OH",
				Name:                 "up2d",
				DeploymentTemplateId: "aws-io-optimized-v2",
				Region:               "aws-eu-central-1",
				Version:              "7.13.1",
				Elasticsearch: &elasticsearchv2.Elasticsearch{
					RefId:  ec.String("main-elasticsearch"),
					Region: ec.String("aws-eu-central-1"),
					Config: &elasticsearchv2.ElasticsearchConfig{
						Plugins: []string{},
					},
					HotTier: elasticsearchv2.CreateTierForTest(
						"hot_content",
						elasticsearchv2.ElasticsearchTopology{
							Size:         ec.String("4g"),
							SizeResource: ec.String("memory"),
							Autoscaling:  &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
					TrustAccount: elasticsearchv2.ElasticsearchTrustAccounts{
						{
							AccountId: ec.String("ANID"),
							TrustAll:  ec.Bool(true),
						},
						{
							AccountId:      ec.String("anotherID"),
							TrustAll:       ec.Bool(false),
							TrustAllowlist: []string{"abc", "dfg", "hij"},
						},
					},
					TrustExternal: elasticsearchv2.ElasticsearchTrustExternals{
						{
							RelationshipId: ec.String("external_id"),
							TrustAll:       ec.Bool(true),
						},
						{
							RelationshipId: ec.String("another_external_id"),
							TrustAll:       ec.Bool(false),
							TrustAllowlist: []string{"abc", "dfg"},
						},
					},
				},
			},
		},

		{
			name: "flattens an aws plan with topology.config set",
			args: args{
				res: &models.DeploymentGetResponse{
					ID:    ec.String("123b7b540dfc967a7a649c18e2fce4ed"),
					Alias: "OH",
					Name:  ec.String("up2d"),
					Resources: &models.DeploymentResources{
						Elasticsearch: []*models.ElasticsearchResourceInfo{{
							RefID:  ec.String("main-elasticsearch"),
							Region: ec.String("aws-eu-central-1"),
							Info: &models.ElasticsearchClusterInfo{
								Status: ec.String("running"),
								PlanInfo: &models.ElasticsearchClusterPlansInfo{
									Current: &models.ElasticsearchClusterPlanInfo{
										Plan: &models.ElasticsearchClusterPlan{
											DeploymentTemplate: &models.DeploymentTemplateReference{
												ID: ec.String("aws-io-optimized-v2"),
											},
											Elasticsearch: &models.ElasticsearchConfiguration{
												Version: "7.13.1",
											},
											ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
												ID: "hot_content",
												Size: &models.TopologySize{
													Value:    ec.Int32(4096),
													Resource: ec.String("memory"),
												},
												Elasticsearch: &models.ElasticsearchConfiguration{
													UserSettingsYaml: "a.setting: true",
												},
											}},
										},
									},
								},
								Settings: &models.ElasticsearchClusterSettings{},
							},
						}},
					},
				},
			},
			want: Deployment{
				Id:                   "123b7b540dfc967a7a649c18e2fce4ed",
				Alias:                "OH",
				Name:                 "up2d",
				DeploymentTemplateId: "aws-io-optimized-v2",
				Region:               "aws-eu-central-1",
				Version:              "7.13.1",
				Elasticsearch: &elasticsearchv2.Elasticsearch{
					RefId:  ec.String("main-elasticsearch"),
					Region: ec.String("aws-eu-central-1"),
					Config: &elasticsearchv2.ElasticsearchConfig{
						Plugins: []string{},
					},
					TrustAccount:  elasticsearchv2.ElasticsearchTrustAccounts{},
					TrustExternal: elasticsearchv2.ElasticsearchTrustExternals{},
					HotTier: elasticsearchv2.CreateTierForTest(
						"hot_content",
						elasticsearchv2.ElasticsearchTopology{
							Size:         ec.String("4g"),
							SizeResource: ec.String("memory"),
							Autoscaling:  &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
				},
			},
		},

		{
			name: "flattens an plan with config.docker_image set",
			args: args{
				res: &models.DeploymentGetResponse{
					ID:    ec.String("123b7b540dfc967a7a649c18e2fce4ed"),
					Alias: "OH",
					Name:  ec.String("up2d"),
					Resources: &models.DeploymentResources{
						Elasticsearch: []*models.ElasticsearchResourceInfo{{
							RefID:  ec.String("main-elasticsearch"),
							Region: ec.String("aws-eu-central-1"),
							Info: &models.ElasticsearchClusterInfo{
								Status: ec.String("running"),
								PlanInfo: &models.ElasticsearchClusterPlansInfo{
									Current: &models.ElasticsearchClusterPlanInfo{
										Plan: &models.ElasticsearchClusterPlan{
											DeploymentTemplate: &models.DeploymentTemplateReference{
												ID: ec.String("aws-io-optimized-v2"),
											},
											Elasticsearch: &models.ElasticsearchConfiguration{
												Version:     "7.14.1",
												DockerImage: "docker.elastic.com/elasticsearch/cloud:7.14.1-hash",
											},
											ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
												ID: "hot_content",
												Size: &models.TopologySize{
													Value:    ec.Int32(4096),
													Resource: ec.String("memory"),
												},
												Elasticsearch: &models.ElasticsearchConfiguration{
													UserSettingsYaml: "a.setting: true",
												},
												ZoneCount: 1,
											}},
										},
									},
								},
								Settings: &models.ElasticsearchClusterSettings{},
							},
						}},
						Apm: []*models.ApmResourceInfo{{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							RefID:                     ec.String("main-apm"),
							Region:                    ec.String("aws-eu-central-1"),
							Info: &models.ApmInfo{
								Status: ec.String("running"),
								PlanInfo: &models.ApmPlansInfo{Current: &models.ApmPlanInfo{
									Plan: &models.ApmPlan{
										Apm: &models.ApmConfiguration{
											Version:     "7.14.1",
											DockerImage: "docker.elastic.com/apm/cloud:7.14.1-hash",
											SystemSettings: &models.ApmSystemSettings{
												DebugEnabled: ec.Bool(false),
											},
										},
										ClusterTopology: []*models.ApmTopologyElement{{
											InstanceConfigurationID: "aws.apm.r5d",
											Size: &models.TopologySize{
												Resource: ec.String("memory"),
												Value:    ec.Int32(512),
											},
											ZoneCount: 1,
										}},
									},
								}},
							},
						}},
						Kibana: []*models.KibanaResourceInfo{{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							RefID:                     ec.String("main-kibana"),
							Region:                    ec.String("aws-eu-central-1"),
							Info: &models.KibanaClusterInfo{
								Status: ec.String("running"),
								PlanInfo: &models.KibanaClusterPlansInfo{Current: &models.KibanaClusterPlanInfo{
									Plan: &models.KibanaClusterPlan{
										Kibana: &models.KibanaConfiguration{
											Version:     "7.14.1",
											DockerImage: "docker.elastic.com/kibana/cloud:7.14.1-hash",
										},
										ClusterTopology: []*models.KibanaClusterTopologyElement{{
											InstanceConfigurationID: "aws.kibana.r5d",
											Size: &models.TopologySize{
												Resource: ec.String("memory"),
												Value:    ec.Int32(1024),
											},
											ZoneCount: 1,
										}},
									},
								}},
							},
						}},
						EnterpriseSearch: []*models.EnterpriseSearchResourceInfo{{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							RefID:                     ec.String("main-enterprise_search"),
							Region:                    ec.String("aws-eu-central-1"),
							Info: &models.EnterpriseSearchInfo{
								Status: ec.String("running"),
								PlanInfo: &models.EnterpriseSearchPlansInfo{Current: &models.EnterpriseSearchPlanInfo{
									Plan: &models.EnterpriseSearchPlan{
										EnterpriseSearch: &models.EnterpriseSearchConfiguration{
											Version:     "7.14.1",
											DockerImage: "docker.elastic.com/enterprise_search/cloud:7.14.1-hash",
										},
										ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
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
											ZoneCount: 2,
										}},
									},
								}},
							},
						}},
					},
				},
			},
			want: Deployment{
				Id:                   "123b7b540dfc967a7a649c18e2fce4ed",
				Alias:                "OH",
				Name:                 "up2d",
				DeploymentTemplateId: "aws-io-optimized-v2",
				Region:               "aws-eu-central-1",
				Version:              "7.14.1",
				Elasticsearch: &elasticsearchv2.Elasticsearch{
					RefId:  ec.String("main-elasticsearch"),
					Region: ec.String("aws-eu-central-1"),
					Config: &elasticsearchv2.ElasticsearchConfig{
						Plugins:     []string{},
						DockerImage: ec.String("docker.elastic.com/elasticsearch/cloud:7.14.1-hash"),
					},
					TrustAccount:  elasticsearchv2.ElasticsearchTrustAccounts{},
					TrustExternal: elasticsearchv2.ElasticsearchTrustExternals{},
					HotTier: elasticsearchv2.CreateTierForTest(
						"hot_content",
						elasticsearchv2.ElasticsearchTopology{
							Size:         ec.String("4g"),
							SizeResource: ec.String("memory"),
							ZoneCount:    1,
							Autoscaling:  &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
				},
				Kibana: &kibanav2.Kibana{
					RefId:                     ec.String("main-kibana"),
					Region:                    ec.String("aws-eu-central-1"),
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					Config: &kibanav2.KibanaConfig{
						DockerImage: ec.String("docker.elastic.com/kibana/cloud:7.14.1-hash"),
					},
					InstanceConfigurationId: ec.String("aws.kibana.r5d"),
					Size:                    ec.String("1g"),
					SizeResource:            ec.String("memory"),
					ZoneCount:               1,
				},
				Apm: &apmv2.Apm{
					RefId:                     ec.String("main-apm"),
					Region:                    ec.String("aws-eu-central-1"),
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					Config: &apmv2.ApmConfig{
						DockerImage:  ec.String("docker.elastic.com/apm/cloud:7.14.1-hash"),
						DebugEnabled: ec.Bool(false),
					},
					InstanceConfigurationId: ec.String("aws.apm.r5d"),
					Size:                    ec.String("0.5g"),
					SizeResource:            ec.String("memory"),
					ZoneCount:               1,
				},
				EnterpriseSearch: &enterprisesearchv2.EnterpriseSearch{
					RefId:                     ec.String("main-enterprise_search"),
					Region:                    ec.String("aws-eu-central-1"),
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					Config: &enterprisesearchv2.EnterpriseSearchConfig{
						DockerImage: ec.String("docker.elastic.com/enterprise_search/cloud:7.14.1-hash"),
					},
					InstanceConfigurationId: ec.String("aws.enterprisesearch.m5d"),
					Size:                    ec.String("2g"),
					SizeResource:            ec.String("memory"),
					ZoneCount:               2,
					NodeTypeAppserver:       ec.Bool(true),
					NodeTypeConnector:       ec.Bool(true),
					NodeTypeWorker:          ec.Bool(true),
				},
			},
		},

		{
			name: "flattens an aws plan (io-optimized) with tags",
			args: args{res: deploymentGetResponseFromFile(t, "../../testdata/deployment-aws-io-optimized-tags.json")},
			want: Deployment{
				Id:                   "123365f2805e46808d40849b1c0b266b",
				Alias:                "my-deployment",
				Name:                 "up2d",
				DeploymentTemplateId: "aws-io-optimized-v2",
				Region:               "aws-eu-central-1",
				Version:              "7.9.2",
				Tags: map[string]string{
					"aaa":   "bbb",
					"cost":  "rnd",
					"owner": "elastic",
				},
				Elasticsearch: &elasticsearchv2.Elasticsearch{
					RefId:         ec.String("main-elasticsearch"),
					ResourceId:    ec.String("1239f7ee7196439ba2d105319ac5eba7"),
					Region:        ec.String("aws-eu-central-1"),
					Autoscale:     ec.Bool(false),
					CloudID:       ec.String("up2d:someCloudID"),
					TrustAccount:  elasticsearchv2.ElasticsearchTrustAccounts{},
					TrustExternal: elasticsearchv2.ElasticsearchTrustExternals{},
					HttpEndpoint:  ec.String("http://1239f7ee7196439ba2d105319ac5eba7.eu-central-1.aws.cloud.es.io:9200"),
					HttpsEndpoint: ec.String("https://1239f7ee7196439ba2d105319ac5eba7.eu-central-1.aws.cloud.es.io:9243"),
					Config: &elasticsearchv2.ElasticsearchConfig{
						Plugins: []string{},
					},
					HotTier: elasticsearchv2.CreateTierForTest(
						"hot_content",
						elasticsearchv2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("aws.data.highio.i3"),
							Size:                    ec.String("8g"),
							SizeResource:            ec.String("memory"),
							NodeTypeData:            ec.String("true"),
							NodeTypeIngest:          ec.String("true"),
							NodeTypeMaster:          ec.String("true"),
							NodeTypeMl:              ec.String("false"),
							ZoneCount:               2,
							Autoscaling:             &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
				},
				Kibana: &kibanav2.Kibana{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-kibana"),
					ResourceId:                ec.String("123dcfda06254ca789eb287e8b73ff4c"),
					Region:                    ec.String("aws-eu-central-1"),
					HttpEndpoint:              ec.String("http://123dcfda06254ca789eb287e8b73ff4c.eu-central-1.aws.cloud.es.io:9200"),
					HttpsEndpoint:             ec.String("https://123dcfda06254ca789eb287e8b73ff4c.eu-central-1.aws.cloud.es.io:9243"),
					InstanceConfigurationId:   ec.String("aws.kibana.r5d"),
					Size:                      ec.String("1g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
				Apm: &apmv2.Apm{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-apm"),
					ResourceId:                ec.String("12328579b3bf40c8b58c1a0ed5a4bd8b"),
					Region:                    ec.String("aws-eu-central-1"),
					HttpEndpoint:              ec.String("http://12328579b3bf40c8b58c1a0ed5a4bd8b.apm.eu-central-1.aws.cloud.es.io:80"),
					HttpsEndpoint:             ec.String("https://12328579b3bf40c8b58c1a0ed5a4bd8b.apm.eu-central-1.aws.cloud.es.io:443"),
					InstanceConfigurationId:   ec.String("aws.apm.r5d"),
					Size:                      ec.String("0.5g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
			},
		},

		{
			name: "flattens a gcp plan (io-optimized)",
			args: args{res: deploymentGetResponseFromFile(t, "../../testdata/deployment-gcp-io-optimized.json")},
			want: Deployment{
				Id:                   "1239e402d6df471ea374bd68e3f91cc5",
				Alias:                "my-deployment",
				Name:                 "up2d",
				DeploymentTemplateId: "gcp-io-optimized",
				Region:               "gcp-asia-east1",
				Version:              "7.9.2",
				Elasticsearch: &elasticsearchv2.Elasticsearch{
					RefId:         ec.String("main-elasticsearch"),
					ResourceId:    ec.String("123695e76d914005bf90b717e668ad4b"),
					Region:        ec.String("gcp-asia-east1"),
					Autoscale:     ec.Bool(false),
					CloudID:       ec.String("up2d:someCloudID"),
					TrustAccount:  elasticsearchv2.ElasticsearchTrustAccounts{},
					TrustExternal: elasticsearchv2.ElasticsearchTrustExternals{},
					HttpEndpoint:  ec.String("http://123695e76d914005bf90b717e668ad4b.asia-east1.gcp.elastic-cloud.com:9200"),
					HttpsEndpoint: ec.String("https://123695e76d914005bf90b717e668ad4b.asia-east1.gcp.elastic-cloud.com:9243"),
					Config: &elasticsearchv2.ElasticsearchConfig{
						Plugins: []string{},
					},
					HotTier: elasticsearchv2.CreateTierForTest(
						"hot_content",
						elasticsearchv2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("gcp.data.highio.1"),
							Size:                    ec.String("8g"),
							SizeResource:            ec.String("memory"),
							NodeTypeData:            ec.String("true"),
							NodeTypeIngest:          ec.String("true"),
							NodeTypeMaster:          ec.String("true"),
							NodeTypeMl:              ec.String("false"),
							ZoneCount:               2,
							Autoscaling:             &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
				},
				Kibana: &kibanav2.Kibana{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-kibana"),
					ResourceId:                ec.String("12365046781e4d729a07df64fe67c8c6"),
					Region:                    ec.String("gcp-asia-east1"),
					HttpEndpoint:              ec.String("http://12365046781e4d729a07df64fe67c8c6.asia-east1.gcp.elastic-cloud.com:9200"),
					HttpsEndpoint:             ec.String("https://12365046781e4d729a07df64fe67c8c6.asia-east1.gcp.elastic-cloud.com:9243"),
					InstanceConfigurationId:   ec.String("gcp.kibana.1"),
					Size:                      ec.String("1g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
				Apm: &apmv2.Apm{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-apm"),
					ResourceId:                ec.String("12307c6c304949b8a9f3682b80900879"),
					Region:                    ec.String("gcp-asia-east1"),
					HttpEndpoint:              ec.String("http://12307c6c304949b8a9f3682b80900879.apm.asia-east1.gcp.elastic-cloud.com:80"),
					HttpsEndpoint:             ec.String("https://12307c6c304949b8a9f3682b80900879.apm.asia-east1.gcp.elastic-cloud.com:443"),
					InstanceConfigurationId:   ec.String("gcp.apm.1"),
					Size:                      ec.String("0.5g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
			},
		},

		{
			name: "flattens a gcp plan with autoscale set (io-optimized)",
			args: args{res: deploymentGetResponseFromFile(t, "../../testdata/deployment-gcp-io-optimized-autoscale.json")},
			want: Deployment{
				Id:                   "1239e402d6df471ea374bd68e3f91cc5",
				Alias:                "",
				Name:                 "up2d",
				DeploymentTemplateId: "gcp-io-optimized",
				Region:               "gcp-asia-east1",
				Version:              "7.9.2",
				Elasticsearch: &elasticsearchv2.Elasticsearch{
					RefId:         ec.String("main-elasticsearch"),
					ResourceId:    ec.String("123695e76d914005bf90b717e668ad4b"),
					Region:        ec.String("gcp-asia-east1"),
					Autoscale:     ec.Bool(true),
					CloudID:       ec.String("up2d:someCloudID"),
					TrustAccount:  elasticsearchv2.ElasticsearchTrustAccounts{},
					TrustExternal: elasticsearchv2.ElasticsearchTrustExternals{},
					HttpEndpoint:  ec.String("http://123695e76d914005bf90b717e668ad4b.asia-east1.gcp.elastic-cloud.com:9200"),
					HttpsEndpoint: ec.String("https://123695e76d914005bf90b717e668ad4b.asia-east1.gcp.elastic-cloud.com:9243"),
					Config: &elasticsearchv2.ElasticsearchConfig{
						Plugins: []string{},
					},
					HotTier: elasticsearchv2.CreateTierForTest(
						"hot_content",
						elasticsearchv2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("gcp.data.highio.1"),
							Size:                    ec.String("8g"),
							SizeResource:            ec.String("memory"),
							NodeTypeData:            ec.String("true"),
							NodeTypeIngest:          ec.String("true"),
							NodeTypeMaster:          ec.String("true"),
							NodeTypeMl:              ec.String("false"),
							ZoneCount:               2,
							Autoscaling: &elasticsearchv2.ElasticsearchTopologyAutoscaling{
								MaxSize:            ec.String("29g"),
								MaxSizeResource:    ec.String("memory"),
								PolicyOverrideJson: ec.String(`{"proactive_storage":{"forecast_window":"3 h"}}`),
							},
						},
					),
					MlTier: elasticsearchv2.CreateTierForTest(
						"ml",
						elasticsearchv2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("gcp.ml.1"),
							Size:                    ec.String("1g"),
							SizeResource:            ec.String("memory"),
							NodeTypeData:            ec.String("false"),
							NodeTypeIngest:          ec.String("false"),
							NodeTypeMaster:          ec.String("false"),
							NodeTypeMl:              ec.String("true"),
							ZoneCount:               1,
							Autoscaling: &elasticsearchv2.ElasticsearchTopologyAutoscaling{
								MaxSize:         ec.String("30g"),
								MaxSizeResource: ec.String("memory"),
								MinSize:         ec.String("1g"),
								MinSizeResource: ec.String("memory"),
							},
						},
					),
					MasterTier: elasticsearchv2.CreateTierForTest(
						"master",
						elasticsearchv2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("gcp.master.1"),
							Size:                    ec.String("0g"),
							SizeResource:            ec.String("memory"),
							NodeTypeData:            ec.String("false"),
							NodeTypeIngest:          ec.String("false"),
							NodeTypeMaster:          ec.String("true"),
							NodeTypeMl:              ec.String("false"),
							ZoneCount:               3,
							Autoscaling:             &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
					CoordinatingTier: elasticsearchv2.CreateTierForTest(
						"coordinating",
						elasticsearchv2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("gcp.coordinating.1"),
							Size:                    ec.String("0g"),
							SizeResource:            ec.String("memory"),
							NodeTypeData:            ec.String("false"),
							NodeTypeIngest:          ec.String("true"),
							NodeTypeMaster:          ec.String("false"),
							NodeTypeMl:              ec.String("false"),
							ZoneCount:               2,
							Autoscaling:             &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
				},
				Kibana: &kibanav2.Kibana{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-kibana"),
					ResourceId:                ec.String("12365046781e4d729a07df64fe67c8c6"),
					Region:                    ec.String("gcp-asia-east1"),
					HttpEndpoint:              ec.String("http://12365046781e4d729a07df64fe67c8c6.asia-east1.gcp.elastic-cloud.com:9200"),
					HttpsEndpoint:             ec.String("https://12365046781e4d729a07df64fe67c8c6.asia-east1.gcp.elastic-cloud.com:9243"),
					InstanceConfigurationId:   ec.String("gcp.kibana.1"),
					Size:                      ec.String("1g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
				Apm: &apmv2.Apm{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-apm"),
					ResourceId:                ec.String("12307c6c304949b8a9f3682b80900879"),
					Region:                    ec.String("gcp-asia-east1"),
					HttpEndpoint:              ec.String("http://12307c6c304949b8a9f3682b80900879.apm.asia-east1.gcp.elastic-cloud.com:80"),
					HttpsEndpoint:             ec.String("https://12307c6c304949b8a9f3682b80900879.apm.asia-east1.gcp.elastic-cloud.com:443"),
					InstanceConfigurationId:   ec.String("gcp.apm.1"),
					Size:                      ec.String("0.5g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
			},
		},

		{
			name: "flattens a gcp plan (hot-warm)",
			args: args{res: deploymentGetResponseFromFile(t, "../../testdata/deployment-gcp-hot-warm.json")},
			want: Deployment{
				Id:                   "123d148423864552aa57b59929d4bf4d",
				Name:                 "up2d-hot-warm",
				DeploymentTemplateId: "gcp-hot-warm",
				Region:               "gcp-us-central1",
				Version:              "7.9.2",
				Elasticsearch: &elasticsearchv2.Elasticsearch{
					RefId:         ec.String("main-elasticsearch"),
					ResourceId:    ec.String("123e837db6ee4391bb74887be35a7a91"),
					Region:        ec.String("gcp-us-central1"),
					Autoscale:     ec.Bool(false),
					CloudID:       ec.String("up2d-hot-warm:someCloudID"),
					TrustAccount:  elasticsearchv2.ElasticsearchTrustAccounts{},
					TrustExternal: elasticsearchv2.ElasticsearchTrustExternals{},
					HttpEndpoint:  ec.String("http://123e837db6ee4391bb74887be35a7a91.us-central1.gcp.cloud.es.io:9200"),
					HttpsEndpoint: ec.String("https://123e837db6ee4391bb74887be35a7a91.us-central1.gcp.cloud.es.io:9243"),
					Config: &elasticsearchv2.ElasticsearchConfig{
						Plugins: []string{},
					},
					HotTier: elasticsearchv2.CreateTierForTest(
						"hot_content",
						elasticsearchv2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("gcp.data.highio.1"),
							Size:                    ec.String("4g"),
							SizeResource:            ec.String("memory"),
							NodeTypeData:            ec.String("true"),
							NodeTypeIngest:          ec.String("true"),
							NodeTypeMaster:          ec.String("true"),
							NodeTypeMl:              ec.String("false"),
							ZoneCount:               2,
							Autoscaling:             &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
					WarmTier: elasticsearchv2.CreateTierForTest(
						"warm",
						elasticsearchv2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("gcp.data.highstorage.1"),
							Size:                    ec.String("4g"),
							SizeResource:            ec.String("memory"),
							NodeTypeData:            ec.String("true"),
							NodeTypeIngest:          ec.String("true"),
							NodeTypeMaster:          ec.String("false"),
							NodeTypeMl:              ec.String("false"),
							ZoneCount:               2,
							Autoscaling:             &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
					CoordinatingTier: elasticsearchv2.CreateTierForTest(
						"coordinating",
						elasticsearchv2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("gcp.coordinating.1"),
							Size:                    ec.String("0g"),
							SizeResource:            ec.String("memory"),
							NodeTypeData:            ec.String("false"),
							NodeTypeIngest:          ec.String("true"),
							NodeTypeMaster:          ec.String("false"),
							NodeTypeMl:              ec.String("false"),
							ZoneCount:               2,
							Autoscaling:             &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
				},
				Kibana: &kibanav2.Kibana{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-kibana"),
					ResourceId:                ec.String("12372cc60d284e7e96b95ad14727c23d"),
					Region:                    ec.String("gcp-us-central1"),
					HttpEndpoint:              ec.String("http://12372cc60d284e7e96b95ad14727c23d.us-central1.gcp.cloud.es.io:9200"),
					HttpsEndpoint:             ec.String("https://12372cc60d284e7e96b95ad14727c23d.us-central1.gcp.cloud.es.io:9243"),
					InstanceConfigurationId:   ec.String("gcp.kibana.1"),
					Size:                      ec.String("1g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
				Apm: &apmv2.Apm{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-apm"),
					ResourceId:                ec.String("1234b68b0b9347f1b49b1e01b33bf4a4"),
					Region:                    ec.String("gcp-us-central1"),
					HttpEndpoint:              ec.String("http://1234b68b0b9347f1b49b1e01b33bf4a4.apm.us-central1.gcp.cloud.es.io:80"),
					HttpsEndpoint:             ec.String("https://1234b68b0b9347f1b49b1e01b33bf4a4.apm.us-central1.gcp.cloud.es.io:443"),
					InstanceConfigurationId:   ec.String("gcp.apm.1"),
					Size:                      ec.String("0.5g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
			},
		},

		{
			name: "flattens a gcp plan (hot-warm) with node_roles",
			args: args{res: deploymentGetResponseFromFile(t, "../../testdata/deployment-gcp-hot-warm-node_roles.json")},
			want: Deployment{
				Id:                   "123d148423864552aa57b59929d4bf4d",
				Name:                 "up2d-hot-warm",
				DeploymentTemplateId: "gcp-hot-warm",
				Region:               "gcp-us-central1",
				Version:              "7.11.0",
				Elasticsearch: &elasticsearchv2.Elasticsearch{
					RefId:         ec.String("main-elasticsearch"),
					ResourceId:    ec.String("123e837db6ee4391bb74887be35a7a91"),
					Region:        ec.String("gcp-us-central1"),
					Autoscale:     ec.Bool(false),
					CloudID:       ec.String("up2d-hot-warm:someCloudID"),
					TrustAccount:  elasticsearchv2.ElasticsearchTrustAccounts{},
					TrustExternal: elasticsearchv2.ElasticsearchTrustExternals{},
					HttpEndpoint:  ec.String("http://123e837db6ee4391bb74887be35a7a91.us-central1.gcp.cloud.es.io:9200"),
					HttpsEndpoint: ec.String("https://123e837db6ee4391bb74887be35a7a91.us-central1.gcp.cloud.es.io:9243"),
					Config: &elasticsearchv2.ElasticsearchConfig{
						Plugins: []string{},
					},
					HotTier: elasticsearchv2.CreateTierForTest(
						"hot_content",
						elasticsearchv2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("gcp.data.highio.1"),
							Size:                    ec.String("4g"),
							SizeResource:            ec.String("memory"),
							ZoneCount:               2,
							NodeRoles: []string{
								"master",
								"ingest",
								"remote_cluster_client",
								"data_hot",
								"transform",
								"data_content",
							},
							Autoscaling: &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
					WarmTier: elasticsearchv2.CreateTierForTest(
						"warm",
						elasticsearchv2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("gcp.data.highstorage.1"),
							Size:                    ec.String("4g"),
							SizeResource:            ec.String("memory"),
							ZoneCount:               2,
							NodeRoles: []string{
								"data_warm",
								"remote_cluster_client",
							},
							Autoscaling: &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
					MlTier: elasticsearchv2.CreateTierForTest(
						"ml",
						elasticsearchv2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("gcp.ml.1"),
							Size:                    ec.String("0g"),
							SizeResource:            ec.String("memory"),
							ZoneCount:               1,
							NodeRoles:               []string{"ml", "remote_cluster_client"},
							Autoscaling:             &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
					MasterTier: elasticsearchv2.CreateTierForTest(
						"master",
						elasticsearchv2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("gcp.master.1"),
							Size:                    ec.String("0g"),
							SizeResource:            ec.String("memory"),
							ZoneCount:               3,
							NodeRoles:               []string{"master"},
							Autoscaling:             &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
					CoordinatingTier: elasticsearchv2.CreateTierForTest(
						"coordinating",
						elasticsearchv2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("gcp.coordinating.1"),
							Size:                    ec.String("0g"),
							SizeResource:            ec.String("memory"),
							ZoneCount:               2,
							NodeRoles:               []string{"ingest", "remote_cluster_client"},
							Autoscaling:             &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
				},
				Kibana: &kibanav2.Kibana{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-kibana"),
					ResourceId:                ec.String("12372cc60d284e7e96b95ad14727c23d"),
					Region:                    ec.String("gcp-us-central1"),
					HttpEndpoint:              ec.String("http://12372cc60d284e7e96b95ad14727c23d.us-central1.gcp.cloud.es.io:9200"),
					HttpsEndpoint:             ec.String("https://12372cc60d284e7e96b95ad14727c23d.us-central1.gcp.cloud.es.io:9243"),
					InstanceConfigurationId:   ec.String("gcp.kibana.1"),
					Size:                      ec.String("1g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
				Apm: &apmv2.Apm{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-apm"),
					ResourceId:                ec.String("1234b68b0b9347f1b49b1e01b33bf4a4"),
					Region:                    ec.String("gcp-us-central1"),
					HttpEndpoint:              ec.String("http://1234b68b0b9347f1b49b1e01b33bf4a4.apm.us-central1.gcp.cloud.es.io:80"),
					HttpsEndpoint:             ec.String("https://1234b68b0b9347f1b49b1e01b33bf4a4.apm.us-central1.gcp.cloud.es.io:443"),
					InstanceConfigurationId:   ec.String("gcp.apm.1"),
					Size:                      ec.String("0.5g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
			},
		},

		{
			name: "flattens an aws plan (Cross Cluster Search)",
			args: args{
				res: deploymentGetResponseFromFile(t, "../../testdata/deployment-aws-ccs.json"),
				remotes: models.RemoteResources{Resources: []*models.RemoteResourceRef{
					{
						Alias:              ec.String("alias"),
						DeploymentID:       ec.String("someid"),
						ElasticsearchRefID: ec.String("main-elasticsearch"),
						SkipUnavailable:    ec.Bool(true),
					},
					{
						DeploymentID:       ec.String("some other id"),
						ElasticsearchRefID: ec.String("main-elasticsearch"),
					},
				}},
			},
			want: Deployment{
				Id:                   "123987dee8d54505974295e07fc7d13e",
				Name:                 "ccs",
				DeploymentTemplateId: "aws-cross-cluster-search-v2",
				Region:               "eu-west-1",
				Version:              "7.9.2",
				Elasticsearch: &elasticsearchv2.Elasticsearch{
					RefId:         ec.String("main-elasticsearch"),
					ResourceId:    ec.String("1230b3ae633b4f51a432d50971f7f1c1"),
					Region:        ec.String("eu-west-1"),
					Autoscale:     ec.Bool(false),
					CloudID:       ec.String("ccs:someCloudID"),
					TrustAccount:  elasticsearchv2.ElasticsearchTrustAccounts{},
					TrustExternal: elasticsearchv2.ElasticsearchTrustExternals{},
					HttpEndpoint:  ec.String("http://1230b3ae633b4f51a432d50971f7f1c1.eu-west-1.aws.found.io:9200"),
					HttpsEndpoint: ec.String("https://1230b3ae633b4f51a432d50971f7f1c1.eu-west-1.aws.found.io:9243"),
					Config: &elasticsearchv2.ElasticsearchConfig{
						Plugins: []string{},
					},
					RemoteCluster: elasticsearchv2.ElasticsearchRemoteClusters{
						{
							Alias:           ec.String("alias"),
							DeploymentId:    ec.String("someid"),
							RefId:           ec.String("main-elasticsearch"),
							SkipUnavailable: ec.Bool(true),
						},
						{
							DeploymentId: ec.String("some other id"),
							RefId:        ec.String("main-elasticsearch"),
						},
					},
					HotTier: elasticsearchv2.CreateTierForTest(
						"hot_content",
						elasticsearchv2.ElasticsearchTopology{
							InstanceConfigurationId: ec.String("aws.ccs.r5d"),
							Size:                    ec.String("1g"),
							SizeResource:            ec.String("memory"),
							ZoneCount:               1,
							NodeTypeData:            ec.String("true"),
							NodeTypeIngest:          ec.String("true"),
							NodeTypeMaster:          ec.String("true"),
							NodeTypeMl:              ec.String("false"),
							Autoscaling:             &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
						},
					),
				},
				Kibana: &kibanav2.Kibana{
					ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					RefId:                     ec.String("main-kibana"),
					ResourceId:                ec.String("12317425e9e14491b74ee043db3402eb"),
					Region:                    ec.String("eu-west-1"),
					HttpEndpoint:              ec.String("http://12317425e9e14491b74ee043db3402eb.eu-west-1.aws.found.io:9200"),
					HttpsEndpoint:             ec.String("https://12317425e9e14491b74ee043db3402eb.eu-west-1.aws.found.io:9243"),
					InstanceConfigurationId:   ec.String("aws.kibana.r5d"),
					Size:                      ec.String("1g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep, err := ReadDeployment(tt.args.res, &tt.args.remotes, nil)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, dep)
				assert.Equal(t, tt.want, *dep)
			}
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

func Test_ProcessSelfInObservability(t *testing.T) {
	tests := []struct {
		name                              string
		deployment                        *Deployment
		baseObservability                 *observabilityv2.Observability
		observabilityIsUnknown            bool
		expectedObservabilityDeploymentID *string
		expectNonNilDiags                 bool
	}{
		{
			name: "should noop if deployment is nil",
		},
		{name: "should noop if observability section is nil",
			deployment: &Deployment{},
		},
		{
			name: "should noop if observability deployment id is nil",
			deployment: &Deployment{
				Observability: &observabilityv2.Observability{},
			},
		},
		{
			name: "should not change the observability deployment id to self if it equals the deployment id and the configured value is not self",
			deployment: &Deployment{
				Id: "deployment-id",
				Observability: &observabilityv2.Observability{
					DeploymentId: ec.String("deployment-id"),
				},
			},
			baseObservability: &observabilityv2.Observability{
				DeploymentId: ec.String("deployment-id"),
			},
			expectedObservabilityDeploymentID: ec.String("deployment-id"),
		},
		{
			name: "should set observability deployment id to self if it equals the deployment id and the configured value is self",
			deployment: &Deployment{
				Id: "deployment-id",
				Observability: &observabilityv2.Observability{
					DeploymentId: ec.String("deployment-id"),
				},
			},
			baseObservability: &observabilityv2.Observability{
				DeploymentId: ec.String("self"),
			},
			expectedObservabilityDeploymentID: ec.String("self"),
		},
		{
			name: "should set observability deployment id to self if it equals the deployment id and the configured value is not set",
			deployment: &Deployment{
				Id: "deployment-id",
				Observability: &observabilityv2.Observability{
					DeploymentId: ec.String("deployment-id"),
				},
			},
			baseObservability:                 &observabilityv2.Observability{},
			expectedObservabilityDeploymentID: ec.String("self"),
		},
		{
			name: "should not change the observability deployment id if it does not equal the deployment id",
			deployment: &Deployment{
				Id: "deployment-id",
				Observability: &observabilityv2.Observability{
					DeploymentId: ec.String("another-deployment-id"),
				},
			},
			expectedObservabilityDeploymentID: ec.String("another-deployment-id"),
		},
		{
			name: "should set observability deployment id to self if it equals the deployment id and no observability is configured",
			deployment: &Deployment{
				Id: "deployment-id",
				Observability: &observabilityv2.Observability{
					DeploymentId: ec.String("deployment-id"),
				},
			},
			expectedObservabilityDeploymentID: ec.String("self"),
		},
		{
			name: "should set observability deployment id to self if it equals the deployment id and the configured value is unknown",
			deployment: &Deployment{
				Id: "deployment-id",
				Observability: &observabilityv2.Observability{
					DeploymentId: ec.String("deployment-id"),
				},
			},
			observabilityIsUnknown:            true,
			expectedObservabilityDeploymentID: ec.String("self"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var obj types.Object
			observabilitySchema := observabilityv2.ObservabilitySchema().GetType()
			schemaWithAttrs, ok := observabilitySchema.(attr.TypeWithAttributeTypes)
			require.True(t, ok)

			if tt.baseObservability != nil {
				diags := tfsdk.ValueFrom(context.Background(), tt.baseObservability, observabilitySchema, &obj)
				require.Nil(t, diags)
			} else if tt.observabilityIsUnknown {
				obj = types.ObjectUnknown(schemaWithAttrs.AttributeTypes())
			} else {
				obj = types.ObjectNull(schemaWithAttrs.AttributeTypes())
			}

			baseDeployment := DeploymentTF{Observability: obj}

			diags := tt.deployment.ProcessSelfInObservability(context.Background(), baseDeployment)
			if tt.expectNonNilDiags {
				require.NotNil(t, diags)
			} else {
				require.Nil(t, diags)
			}

			var finalObservabilityDeploymentID *string
			if tt.deployment != nil && tt.deployment.Observability != nil {
				finalObservabilityDeploymentID = tt.deployment.Observability.DeploymentId
			}

			require.Equal(t, tt.expectedObservabilityDeploymentID, finalObservabilityDeploymentID)
		})
	}
}

func Test_PersistSnapshotSource(t *testing.T) {
	tests := []struct {
		name                                 string
		deployment                           *Deployment
		snapshotSource                       *elasticsearchv1.ElasticsearchSnapshotSource
		expectedSourceElasticsearchClusterId string
		expectedSnapshotName                 string
	}{
		{
			name: "should noop if deployment is nil",
		},
		{
			name:       "should noop if the esplan snapshot source is null",
			deployment: &Deployment{},
		},
		{
			name: "should set the snapshot source cluster and snapshot name if specified in the plan",
			deployment: &Deployment{
				Elasticsearch: &elasticsearchv2.Elasticsearch{},
			},
			snapshotSource: &elasticsearchv1.ElasticsearchSnapshotSource{
				SourceElasticsearchClusterId: "source-cluster-id",
				SnapshotName:                 "snapshot-name",
			},
			expectedSourceElasticsearchClusterId: "source-cluster-id",
			expectedSnapshotName:                 "snapshot-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var obj types.Object
			if tt.snapshotSource != nil {
				diags := tfsdk.ValueFrom(context.Background(), tt.snapshotSource, elasticsearchv2.ElasticsearchSnapshotSourceSchema().GetType(), &obj)
				require.Nil(t, diags)
			}

			esPlan := elasticsearchv2.ElasticsearchTF{
				SnapshotSource: obj,
			}

			diags := tt.deployment.PersistSnapshotSource(context.Background(), &esPlan)
			require.Nil(t, diags)

			var snapshotName, sourceESClusterID string
			if tt.deployment != nil && tt.deployment.Elasticsearch != nil && tt.deployment.Elasticsearch.SnapshotSource != nil {
				snapshotName = tt.deployment.Elasticsearch.SnapshotSource.SnapshotName
				sourceESClusterID = tt.deployment.Elasticsearch.SnapshotSource.SourceElasticsearchClusterId
			}

			require.Equal(t, tt.expectedSnapshotName, snapshotName)
			require.Equal(t, tt.expectedSourceElasticsearchClusterId, sourceESClusterID)
		})
	}
}
