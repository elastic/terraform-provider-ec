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
		ID:     mock.ValidClusterID,
		State:  newSampleDeployment(),
		Schema: newSchema(),
	})

	azureIOOptimizedRes := openDeploymentGet(t, "testdata/deployment-azure-io-optimized.json")
	azureIOOptimizedRD := schema.TestResourceDataRaw(t, newSchema(), nil)
	azureIOOptimizedRD.SetId(mock.ValidClusterID)
	wantAzureIOOptimizedDeployment := util.NewResourceData(t, util.ResDataParams{
		ID: mock.ValidClusterID,
		State: map[string]interface{}{
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
				"topology": []interface{}{map[string]interface{}{
					"instance_configuration_id": "azure.data.highio.l32sv2",
					"node_type_data":            "true",
					"node_type_ingest":          "true",
					"node_type_master":          "true",
					"node_type_ml":              "false",
					"size":                      "4g",
					"size_resource":             "memory",
					"zone_count":                2,
				}},
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

	awsIOOptimizedRes := openDeploymentGet(t, "testdata/deployment-aws-io-optimized.json")
	awsIOOptimizedRD := schema.TestResourceDataRaw(t, newSchema(), nil)
	awsIOOptimizedRD.SetId(mock.ValidClusterID)
	wantAwsIOOptimizedDeployment := util.NewResourceData(t, util.ResDataParams{
		ID: mock.ValidClusterID,
		State: map[string]interface{}{
			"deployment_template_id": "aws-io-optimized-v2",
			"id":                     "123b7b540dfc967a7a649c18e2fce4ed",
			"name":                   "up2d",
			"region":                 "aws-eu-central-1",
			"version":                "7.9.2",
			"apm": []interface{}{map[string]interface{}{
				"elasticsearch_cluster_ref_id": "main-elasticsearch",
				"ref_id":                       "main-apm",
				"region":                       "aws-eu-central-1",
				"resource_id":                  "12328579b3bf40c8b58c1a0ed5a4bd8b",
				"version":                      "7.9.2",
				"http_endpoint":                "http://12328579b3bf40c8b58c1a0ed5a4bd8b.apm.eu-central-1.aws.cloud.es.io:80",
				"https_endpoint":               "https://12328579b3bf40c8b58c1a0ed5a4bd8b.apm.eu-central-1.aws.cloud.es.io:443",
				"topology": []interface{}{map[string]interface{}{
					"instance_configuration_id": "aws.apm.r5d",
					"size":                      "0.5g",
					"size_resource":             "memory",
					"zone_count":                1,
				}},
			}},
			"elasticsearch": []interface{}{map[string]interface{}{
				"cloud_id":       "up2d:someCloudID",
				"http_endpoint":  "http://1239f7ee7196439ba2d105319ac5eba7.eu-central-1.aws.cloud.es.io:9200",
				"https_endpoint": "https://1239f7ee7196439ba2d105319ac5eba7.eu-central-1.aws.cloud.es.io:9243",
				"ref_id":         "main-elasticsearch",
				"region":         "aws-eu-central-1",
				"resource_id":    "1239f7ee7196439ba2d105319ac5eba7",
				"version":        "7.9.2",
				"topology": []interface{}{map[string]interface{}{
					"instance_configuration_id": "aws.data.highio.i3",
					"node_type_data":            "true",
					"node_type_ingest":          "true",
					"node_type_master":          "true",
					"node_type_ml":              "false",
					"size":                      "8g",
					"size_resource":             "memory",
					"zone_count":                2,
				}},
			}},
			"kibana": []interface{}{map[string]interface{}{
				"elasticsearch_cluster_ref_id": "main-elasticsearch",
				"ref_id":                       "main-kibana",
				"region":                       "aws-eu-central-1",
				"resource_id":                  "123dcfda06254ca789eb287e8b73ff4c",
				"version":                      "7.9.2",
				"http_endpoint":                "http://123dcfda06254ca789eb287e8b73ff4c.eu-central-1.aws.cloud.es.io:9200",
				"https_endpoint":               "https://123dcfda06254ca789eb287e8b73ff4c.eu-central-1.aws.cloud.es.io:9243",
				"topology": []interface{}{map[string]interface{}{
					"instance_configuration_id": "aws.kibana.r5d",
					"size":                      "1g",
					"size_resource":             "memory",
					"zone_count":                1,
				}},
			}},
		},
		Schema: newSchema(),
	})

	gcpIOOptimizedRes := openDeploymentGet(t, "testdata/deployment-gcp-io-optimized.json")
	gcpIOOptimizedRD := schema.TestResourceDataRaw(t, newSchema(), nil)
	gcpIOOptimizedRD.SetId(mock.ValidClusterID)
	wantGcpIOOptimizedDeployment := util.NewResourceData(t, util.ResDataParams{
		ID: mock.ValidClusterID,
		State: map[string]interface{}{
			"deployment_template_id": "gcp-io-optimized",
			"id":                     "123b7b540dfc967a7a649c18e2fce4ed",
			"name":                   "up2d",
			"region":                 "gcp-asia-east1",
			"version":                "7.9.2",
			"apm": []interface{}{map[string]interface{}{
				"elasticsearch_cluster_ref_id": "main-elasticsearch",
				"ref_id":                       "main-apm",
				"region":                       "gcp-asia-east1",
				"resource_id":                  "12307c6c304949b8a9f3682b80900879",
				"version":                      "7.9.2",
				"http_endpoint":                "http://12307c6c304949b8a9f3682b80900879.apm.asia-east1.gcp.elastic-cloud.com:80",
				"https_endpoint":               "https://12307c6c304949b8a9f3682b80900879.apm.asia-east1.gcp.elastic-cloud.com:443",
				"topology": []interface{}{map[string]interface{}{
					"instance_configuration_id": "gcp.apm.1",
					"size":                      "0.5g",
					"size_resource":             "memory",
					"zone_count":                1,
				}},
			}},
			"elasticsearch": []interface{}{map[string]interface{}{
				"cloud_id":       "up2d:someCloudID",
				"http_endpoint":  "http://123695e76d914005bf90b717e668ad4b.asia-east1.gcp.elastic-cloud.com:9200",
				"https_endpoint": "https://123695e76d914005bf90b717e668ad4b.asia-east1.gcp.elastic-cloud.com:9243",
				"ref_id":         "main-elasticsearch",
				"region":         "gcp-asia-east1",
				"resource_id":    "123695e76d914005bf90b717e668ad4b",
				"version":        "7.9.2",
				"topology": []interface{}{map[string]interface{}{
					"instance_configuration_id": "gcp.data.highio.1",
					"node_type_data":            "true",
					"node_type_ingest":          "true",
					"node_type_master":          "true",
					"node_type_ml":              "false",
					"size":                      "8g",
					"size_resource":             "memory",
					"zone_count":                2,
				}},
			}},
			"kibana": []interface{}{map[string]interface{}{
				"elasticsearch_cluster_ref_id": "main-elasticsearch",
				"ref_id":                       "main-kibana",
				"region":                       "gcp-asia-east1",
				"resource_id":                  "12365046781e4d729a07df64fe67c8c6",
				"version":                      "7.9.2",
				"http_endpoint":                "http://12365046781e4d729a07df64fe67c8c6.asia-east1.gcp.elastic-cloud.com:9200",
				"https_endpoint":               "https://12365046781e4d729a07df64fe67c8c6.asia-east1.gcp.elastic-cloud.com:9243",
				"topology": []interface{}{map[string]interface{}{
					"instance_configuration_id": "gcp.kibana.1",
					"size":                      "1g",
					"size_resource":             "memory",
					"zone_count":                1,
				}},
			}},
		},
		Schema: newSchema(),
	})

	gcpHotWarmRes := openDeploymentGet(t, "testdata/deployment-gcp-hot-warm.json")
	gcpHotWarmRD := schema.TestResourceDataRaw(t, newSchema(), nil)
	gcpHotWarmRD.SetId(mock.ValidClusterID)
	wantGcpHotWarmDeployment := util.NewResourceData(t, util.ResDataParams{
		ID: mock.ValidClusterID,
		State: map[string]interface{}{
			"deployment_template_id": "gcp-hot-warm",
			"id":                     "123b7b540dfc967a7a649c18e2fce4ed",
			"name":                   "up2d-hot-warm",
			"region":                 "gcp-us-central1",
			"version":                "7.9.2",
			"apm": []interface{}{map[string]interface{}{
				"elasticsearch_cluster_ref_id": "main-elasticsearch",
				"ref_id":                       "main-apm",
				"region":                       "gcp-us-central1",
				"resource_id":                  "1234b68b0b9347f1b49b1e01b33bf4a4",
				"version":                      "7.9.2",
				"http_endpoint":                "http://1234b68b0b9347f1b49b1e01b33bf4a4.apm.us-central1.gcp.cloud.es.io:80",
				"https_endpoint":               "https://1234b68b0b9347f1b49b1e01b33bf4a4.apm.us-central1.gcp.cloud.es.io:443",
				"topology": []interface{}{map[string]interface{}{
					"instance_configuration_id": "gcp.apm.1",
					"size":                      "0.5g",
					"size_resource":             "memory",
					"zone_count":                1,
				}},
			}},
			"elasticsearch": []interface{}{map[string]interface{}{
				"cloud_id":       "up2d-hot-warm:someCloudID",
				"http_endpoint":  "http://123e837db6ee4391bb74887be35a7a91.us-central1.gcp.cloud.es.io:9200",
				"https_endpoint": "https://123e837db6ee4391bb74887be35a7a91.us-central1.gcp.cloud.es.io:9243",
				"ref_id":         "main-elasticsearch",
				"region":         "gcp-us-central1",
				"resource_id":    "123e837db6ee4391bb74887be35a7a91",
				"version":        "7.9.2",
				"topology": []interface{}{
					map[string]interface{}{
						"instance_configuration_id": "gcp.data.highio.1",
						"node_type_data":            "true",
						"node_type_ingest":          "true",
						"node_type_master":          "true",
						"node_type_ml":              "false",
						"size":                      "4g",
						"size_resource":             "memory",
						"zone_count":                2,
					},
					map[string]interface{}{
						"instance_configuration_id": "gcp.data.highstorage.1",
						"node_type_data":            "true",
						"node_type_ingest":          "true",
						"node_type_master":          "false",
						"node_type_ml":              "false",
						"size":                      "4g",
						"size_resource":             "memory",
						"zone_count":                2,
					},
				},
			}},
			"kibana": []interface{}{map[string]interface{}{
				"elasticsearch_cluster_ref_id": "main-elasticsearch",
				"ref_id":                       "main-kibana",
				"region":                       "gcp-us-central1",
				"resource_id":                  "12372cc60d284e7e96b95ad14727c23d",
				"version":                      "7.9.2",
				"http_endpoint":                "http://12372cc60d284e7e96b95ad14727c23d.us-central1.gcp.cloud.es.io:9200",
				"https_endpoint":               "https://12372cc60d284e7e96b95ad14727c23d.us-central1.gcp.cloud.es.io:9243",
				"topology": []interface{}{map[string]interface{}{
					"instance_configuration_id": "gcp.kibana.1",
					"size":                      "1g",
					"size_resource":             "memory",
					"zone_count":                1,
				}},
			}},
		},
		Schema: newSchema(),
	})

	awsCCSRes := openDeploymentGet(t, "testdata/deployment-aws-ccs.json")
	awsCCSRD := schema.TestResourceDataRaw(t, newSchema(), nil)
	awsCCSRD.SetId(mock.ValidClusterID)
	wantAWSCCSDeployment := util.NewResourceData(t, util.ResDataParams{
		ID: mock.ValidClusterID,
		State: map[string]interface{}{
			"deployment_template_id": "aws-cross-cluster-search-v2",
			"id":                     "123b7b540dfc967a7a649c18e2fce4ed",
			"name":                   "ccs",
			"region":                 "eu-west-1",
			"version":                "7.9.2",
			"elasticsearch": []interface{}{map[string]interface{}{
				"cloud_id":       "ccs:someCloudID",
				"http_endpoint":  "http://1230b3ae633b4f51a432d50971f7f1c1.eu-west-1.aws.found.io:9200",
				"https_endpoint": "https://1230b3ae633b4f51a432d50971f7f1c1.eu-west-1.aws.found.io:9243",
				"ref_id":         "main-elasticsearch",
				"region":         "eu-west-1",
				"resource_id":    "1230b3ae633b4f51a432d50971f7f1c1",
				"version":        "7.9.2",
				"remote_cluster": []interface{}{
					map[string]interface{}{
						"alias":            "alias",
						"deployment_id":    "someid",
						"ref_id":           "main-elasticsearch",
						"skip_unavailable": true,
					},
					map[string]interface{}{
						"deployment_id": "some other id",
						"ref_id":        "main-elasticsearch",
					},
				},
				"topology": []interface{}{map[string]interface{}{
					"instance_configuration_id": "aws.ccs.r5d",
					"node_type_data":            "true",
					"node_type_ingest":          "true",
					"node_type_master":          "true",
					"node_type_ml":              "false",
					"size":                      "1g",
					"size_resource":             "memory",
					"zone_count":                1,
				}},
			}},
			"kibana": []interface{}{map[string]interface{}{
				"elasticsearch_cluster_ref_id": "main-elasticsearch",
				"ref_id":                       "main-kibana",
				"region":                       "eu-west-1",
				"resource_id":                  "12317425e9e14491b74ee043db3402eb",
				"version":                      "7.9.2",
				"http_endpoint":                "http://12317425e9e14491b74ee043db3402eb.eu-west-1.aws.found.io:9200",
				"https_endpoint":               "https://12317425e9e14491b74ee043db3402eb.eu-west-1.aws.found.io:9243",
				"topology": []interface{}{map[string]interface{}{
					"instance_configuration_id": "aws.kibana.r5d",
					"size":                      "1g",
					"size_resource":             "memory",
					"zone_count":                1,
				}},
			}},
		},
		Schema: newSchema(),
	})
	argCCSRemotes := models.RemoteResources{Resources: []*models.RemoteResourceRef{
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
	}}

	type args struct {
		d       *schema.ResourceData
		res     *models.DeploymentGetResponse
		remotes models.RemoteResources
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
						Observability: &models.DeploymentObservabilitySettings{
							Logging: &models.DeploymentLoggingSettings{
								Destination: &models.AbsoluteRefID{
									DeploymentID: &mock.ValidClusterID,
									RefID:        ec.String("main-elasticsearch"),
								},
							},
							Metrics: &models.DeploymentMetricsSettings{
								Destination: &models.AbsoluteRefID{
									DeploymentID: &mock.ValidClusterID,
									RefID:        ec.String("main-elasticsearch"),
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
			name: "flattens an azure plan (io-optimized)",
			args: args{d: azureIOOptimizedRD, res: azureIOOptimizedRes},
			want: wantAzureIOOptimizedDeployment,
		},
		{
			name: "flattens an aws plan (io-optimized)",
			args: args{d: awsIOOptimizedRD, res: awsIOOptimizedRes},
			want: wantAwsIOOptimizedDeployment,
		},
		{
			name: "flattens a gcp plan (io-optimized)",
			args: args{d: gcpIOOptimizedRD, res: gcpIOOptimizedRes},
			want: wantGcpIOOptimizedDeployment,
		},
		{
			name: "flattens a gcp plan (hot-warm)",
			args: args{d: gcpHotWarmRD, res: gcpHotWarmRes},
			want: wantGcpHotWarmDeployment,
		},
		{
			name: "flattens an aws plan (Cross Cluster Search)",
			args: args{d: awsCCSRD, res: awsCCSRes, remotes: argCCSRemotes},
			want: wantAWSCCSDeployment,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := modelToState(tt.args.d, tt.args.res, tt.args.remotes)
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
		ID:     mock.ValidClusterID,
		State:  newSampleDeployment(),
		Schema: newSchema(),
	})

	rawData := newSampleDeployment()
	rawData["elasticsearch_username"] = "my-username"
	rawData["elasticsearch_password"] = "my-password"
	rawData["apm_secret_token"] = "some-secret-token"

	wantDeploymentRD := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  rawData,
		Schema: newSchema(),
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
					ID:     mock.ValidClusterID,
					State:  rawData,
					Schema: newSchema(),
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

func Test_hasRunningResources(t *testing.T) {
	type args struct {
		res *models.DeploymentGetResponse
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "has all the resources stopped",
			args: args{res: &models.DeploymentGetResponse{Resources: &models.DeploymentResources{
				Elasticsearch: []*models.ElasticsearchResourceInfo{
					{Info: &models.ElasticsearchClusterInfo{Status: ec.String("stopped")}},
				},
				Kibana: []*models.KibanaResourceInfo{
					{Info: &models.KibanaClusterInfo{Status: ec.String("stopped")}},
				},
				Apm: []*models.ApmResourceInfo{
					{Info: &models.ApmInfo{Status: ec.String("stopped")}},
				},
				EnterpriseSearch: []*models.EnterpriseSearchResourceInfo{
					{Info: &models.EnterpriseSearchInfo{Status: ec.String("stopped")}},
				},
			}}},
			want: false,
		},
		{
			name: "has some resources stopped",
			args: args{res: &models.DeploymentGetResponse{Resources: &models.DeploymentResources{
				Elasticsearch: []*models.ElasticsearchResourceInfo{
					{Info: &models.ElasticsearchClusterInfo{Status: ec.String("running")}},
				},
				Kibana: []*models.KibanaResourceInfo{
					{Info: &models.KibanaClusterInfo{Status: ec.String("stopped")}},
				},
				Apm: []*models.ApmResourceInfo{
					{Info: &models.ApmInfo{Status: ec.String("running")}},
				},
				EnterpriseSearch: []*models.EnterpriseSearchResourceInfo{
					{Info: &models.EnterpriseSearchInfo{Status: ec.String("running")}},
				},
			}}},
			want: true,
		},
		{
			name: "has all resources running",
			args: args{res: &models.DeploymentGetResponse{Resources: &models.DeploymentResources{
				Elasticsearch: []*models.ElasticsearchResourceInfo{
					{Info: &models.ElasticsearchClusterInfo{Status: ec.String("running")}},
				},
				Kibana: []*models.KibanaResourceInfo{
					{Info: &models.KibanaClusterInfo{Status: ec.String("running")}},
				},
				Apm: []*models.ApmResourceInfo{
					{Info: &models.ApmInfo{Status: ec.String("running")}},
				},
				EnterpriseSearch: []*models.EnterpriseSearchResourceInfo{
					{Info: &models.EnterpriseSearchInfo{Status: ec.String("running")}},
				},
			}}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasRunningResources(tt.args.res); got != tt.want {
				t.Errorf("hasRunningResources() = %v, want %v", got, tt.want)
			}
		})
	}
}
