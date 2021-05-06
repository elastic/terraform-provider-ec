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
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func fileAsResponseBody(t *testing.T, name string) io.ReadCloser {
	t.Helper()
	f, err := os.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var buf = new(bytes.Buffer)
	if _, err := io.Copy(buf, f); err != nil {
		t.Fatal(err)
	}
	buf.WriteString("\n")

	return ioutil.NopCloser(buf)
}

func Test_createResourceToModel(t *testing.T) {
	deploymentRD := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleLegacyDeployment(),
		Schema: newSchema(),
	})
	deploymentNodeRolesRD := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleDeployment(),
		Schema: newSchema(),
	})
	ioOptimizedTpl := func() io.ReadCloser {
		return fileAsResponseBody(t, "testdata/template-aws-io-optimized-v2.json")
	}
	deploymentOverrideRd := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleDeploymentOverrides(),
		Schema: newSchema(),
	})
	deploymentOverrideICRd := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleDeploymentOverridesIC(),
		Schema: newSchema(),
	})
	hotWarmTpl := func() io.ReadCloser {
		return fileAsResponseBody(t, "testdata/template-aws-hot-warm-v2.json")
	}
	deploymentHotWarm := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		Schema: newSchema(),
		State: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-hot-warm-v2",
			"region":                 "us-east-1",
			"version":                "7.9.2",
			"elasticsearch":          []interface{}{map[string]interface{}{}},
			"kibana":                 []interface{}{map[string]interface{}{}},
		},
	})

	ccsTpl := func() io.ReadCloser {
		return fileAsResponseBody(t, "testdata/template-aws-cross-cluster-search-v2.json")
	}
	deploymentCCS := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		Schema: newSchema(),
		State: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-cross-cluster-search-v2",
			"region":                 "us-east-1",
			"version":                "7.9.2",
			"elasticsearch":          []interface{}{map[string]interface{}{}},
			"kibana":                 []interface{}{map[string]interface{}{}},
		},
	})

	emptyTpl := func() io.ReadCloser {
		return fileAsResponseBody(t, "testdata/template-empty.json")
	}
	deploymentEmptyTemplate := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		Schema: newSchema(),
		State: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "empty-deployment-template",
			"region":                 "us-east-1",
			"version":                "7.9.2",
			"elasticsearch":          []interface{}{map[string]interface{}{}},
			"kibana":                 []interface{}{map[string]interface{}{}},
			"apm":                    []interface{}{map[string]interface{}{}},
			"enterprise_search":      []interface{}{map[string]interface{}{}},
		},
	})

	deploymentWithTags := util.NewResourceData(t, util.ResDataParams{
		ID: mock.ValidClusterID,
		State: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-io-optimized-v2",
			"region":                 "us-east-1",
			"version":                "7.10.1",
			"elasticsearch": []interface{}{
				map[string]interface{}{
					"topology": []interface{}{map[string]interface{}{
						"id":   "hot_content",
						"size": "8g",
					}},
				},
			},
			"tags": map[string]interface{}{
				"aaa":         "bbb",
				"owner":       "elastic",
				"cost-center": "rnd",
			},
		},
		Schema: newSchema(),
	})

	type args struct {
		d      *schema.ResourceData
		client *api.API
	}
	tests := []struct {
		name string
		args args
		want *models.DeploymentCreateRequest
		err  error
	}{
		{
			name: "parses the resources",
			args: args{
				d: deploymentNodeRolesRD,
				client: api.NewMock(
					mock.New200Response(hotWarmTpl()),
					mock.New200Response(
						mock.NewStructBody(models.DeploymentGetResponse{
							Healthy: ec.Bool(true),
							ID:      ec.String(mock.ValidClusterID),
							Resources: &models.DeploymentResources{
								Elasticsearch: []*models.ElasticsearchResourceInfo{{
									ID:    ec.String(mock.ValidClusterID),
									RefID: ec.String("main-elasticsearch"),
								}},
							},
						}),
					),
				),
			},
			want: &models.DeploymentCreateRequest{
				Name:  "my_deployment_name",
				Alias: "my-deployment",
				Settings: &models.DeploymentCreateSettings{
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
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, hotWarmTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version:                  "7.11.1",
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
								ID: ec.String("aws-hot-warm-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "hot_content",
									ZoneCount:               1,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									NodeRoles: []string{
										"data_content",
										"data_hot",
										"ingest",
										"master",
										"remote_cluster_client",
										"transform",
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               1,
									InstanceConfigurationID: "aws.data.highstorage.d2",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "warm"},
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					}),
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-apm"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{
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
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-enterprise_search"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
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
		{
			name: "parses the legacy resources",
			args: args{
				d: deploymentRD,
				client: api.NewMock(
					mock.New200Response(ioOptimizedTpl()),
					mock.New200Response(
						mock.NewStructBody(models.DeploymentGetResponse{
							Healthy: ec.Bool(true),
							ID:      ec.String(mock.ValidClusterID),
							Resources: &models.DeploymentResources{
								Elasticsearch: []*models.ElasticsearchResourceInfo{{
									ID:    ec.String(mock.ValidClusterID),
									RefID: ec.String("main-elasticsearch"),
								}},
							},
						}),
					),
				),
			},
			want: &models.DeploymentCreateRequest{
				Name:  "my_deployment_name",
				Alias: "my-deployment",
				Settings: &models.DeploymentCreateSettings{
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
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
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
								ID:                      "hot_content",
								ZoneCount:               1,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(2048),
								},
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
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
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					}),
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-apm"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{
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
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-enterprise_search"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
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
		{
			name: "parses the resources with empty declarations (IO Optimized)",
			args: args{
				d: util.NewResourceData(t, util.ResDataParams{
					ID: mock.ValidClusterID,
					State: map[string]interface{}{
						"name":                   "my_deployment_name",
						"deployment_template_id": "aws-io-optimized-v2",
						"region":                 "us-east-1",
						"version":                "7.7.0",
						"elasticsearch":          []interface{}{map[string]interface{}{}},
						"kibana":                 []interface{}{map[string]interface{}{}},
						"apm":                    []interface{}{map[string]interface{}{}},
						"enterprise_search":      []interface{}{map[string]interface{}{}},
						"traffic_filter":         []interface{}{"0.0.0.0/0", "192.168.10.0/24"},
					},
					Schema: newSchema(),
				}),
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name: "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{
					TrafficFilterSettings: &models.TrafficFilterSettings{
						Rulesets: []string{"0.0.0.0/0", "192.168.10.0/24"},
					},
				},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.7.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(8192),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   ec.Bool(true),
									Ingest: ec.Bool(true),
									Master: ec.Bool(true),
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					}),
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-apm"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{},
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
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-enterprise_search"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
								ClusterTopology: []*models.EnterpriseSearchTopologyElement{
									{
										ZoneCount:               2,
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
		{
			name: "parses the resources with empty declarations (IO Optimized) with node_roles",
			args: args{
				d: util.NewResourceData(t, util.ResDataParams{
					ID: mock.ValidClusterID,
					State: map[string]interface{}{
						"name":                   "my_deployment_name",
						"deployment_template_id": "aws-io-optimized-v2",
						"region":                 "us-east-1",
						"version":                "7.11.0",
						"elasticsearch":          []interface{}{map[string]interface{}{}},
						"kibana":                 []interface{}{map[string]interface{}{}},
						"apm":                    []interface{}{map[string]interface{}{}},
						"enterprise_search":      []interface{}{map[string]interface{}{}},
						"traffic_filter":         []interface{}{"0.0.0.0/0", "192.168.10.0/24"},
					},
					Schema: newSchema(),
				}),
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name: "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{
					TrafficFilterSettings: &models.TrafficFilterSettings{
						Rulesets: []string{"0.0.0.0/0", "192.168.10.0/24"},
					},
				},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.11.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(8192),
								},
								NodeRoles: []string{
									"master",
									"ingest",
									"remote_cluster_client",
									"data_hot",
									"transform",
									"data_content",
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					}),
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-apm"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{},
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
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-enterprise_search"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
								ClusterTopology: []*models.EnterpriseSearchTopologyElement{
									{
										ZoneCount:               2,
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
		{
			name: "parses the resources with topology overrides (size)",
			args: args{
				d:      deploymentOverrideRd,
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:  "my_deployment_name",
				Alias: "my-deployment",
				Settings: &models.DeploymentCreateSettings{
					TrafficFilterSettings: &models.TrafficFilterSettings{
						Rulesets: []string{"0.0.0.0/0", "192.168.10.0/24"},
					},
				},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.7.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(4096),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   ec.Bool(true),
									Ingest: ec.Bool(true),
									Master: ec.Bool(true),
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					}),
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
								ClusterTopology: []*models.KibanaClusterTopologyElement{
									{
										ZoneCount:               1,
										InstanceConfigurationID: "aws.kibana.r5d",
										Size: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(2048),
										},
									},
								},
							},
						},
					},
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-apm"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{},
								ClusterTopology: []*models.ApmTopologyElement{{
									ZoneCount:               1,
									InstanceConfigurationID: "aws.apm.r5d",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								}},
							},
						},
					},
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-enterprise_search"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
								ClusterTopology: []*models.EnterpriseSearchTopologyElement{
									{
										ZoneCount:               2,
										InstanceConfigurationID: "aws.enterprisesearch.m5d",
										Size: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(4096),
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
		{
			name: "parses the resources with topology overrides (IC)",
			args: args{
				d:      deploymentOverrideICRd,
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:  "my_deployment_name",
				Alias: "my-deployment",
				Settings: &models.DeploymentCreateSettings{
					TrafficFilterSettings: &models.TrafficFilterSettings{
						Rulesets: []string{"0.0.0.0/0", "192.168.10.0/24"},
					},
				},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.7.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(8192),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   ec.Bool(true),
									Ingest: ec.Bool(true),
									Master: ec.Bool(true),
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					}),
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-apm"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{},
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
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-enterprise_search"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
								ClusterTopology: []*models.EnterpriseSearchTopologyElement{
									{
										ZoneCount:               2,
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
		{
			name: "parses the resources with empty declarations (Hot Warm)",
			args: args{
				d:      deploymentHotWarm,
				client: api.NewMock(mock.New200Response(hotWarmTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, hotWarmTpl(), false), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
							Curation:                  nil,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Curation: nil,
								Version:  "7.9.2",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-hot-warm-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeType: &models.ElasticsearchNodeType{
										Data:   ec.Bool(true),
										Ingest: ec.Bool(true),
										Master: ec.Bool(true),
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d2",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeType: &models.ElasticsearchNodeType{
										Data:   ec.Bool(true),
										Ingest: ec.Bool(true),
										Master: ec.Bool(false),
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "warm",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					}),
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
		{
			name: "parses the resources with empty declarations (Hot Warm) with node_roles",
			args: args{
				d: util.NewResourceData(t, util.ResDataParams{
					ID:     mock.ValidClusterID,
					Schema: newSchema(),
					State: map[string]interface{}{
						"name":                   "my_deployment_name",
						"deployment_template_id": "aws-hot-warm-v2",
						"region":                 "us-east-1",
						"version":                "7.12.0",
						"elasticsearch":          []interface{}{map[string]interface{}{}},
						"kibana":                 []interface{}{map[string]interface{}{}},
					},
				}),
				client: api.NewMock(mock.New200Response(hotWarmTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, hotWarmTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
							Curation:                  nil,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Curation: nil,
								Version:  "7.12.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-hot-warm-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeRoles: []string{
										"master",
										"ingest",
										"remote_cluster_client",
										"data_hot",
										"transform",
										"data_content",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d2",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "warm",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					}),
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
		{
			name: "parses the resources with empty declarations (Hot Warm) with node_roles and extensions",
			args: args{
				d: util.NewResourceData(t, util.ResDataParams{
					ID:     mock.ValidClusterID,
					Schema: newSchema(),
					State: map[string]interface{}{
						"name":                   "my_deployment_name",
						"deployment_template_id": "aws-hot-warm-v2",
						"region":                 "us-east-1",
						"version":                "7.12.0",
						"elasticsearch": []interface{}{map[string]interface{}{
							"extension": []interface{}{
								map[string]interface{}{
									"name":    "my-plugin",
									"type":    "plugin",
									"url":     "repo://12311234",
									"version": "7.7.0",
								},
								map[string]interface{}{
									"name":    "my-second-plugin",
									"type":    "plugin",
									"url":     "repo://12311235",
									"version": "7.7.0",
								},
								map[string]interface{}{
									"name":    "my-bundle",
									"type":    "bundle",
									"url":     "repo://1231122",
									"version": "7.7.0",
								},
								map[string]interface{}{
									"name":    "my-second-bundle",
									"type":    "bundle",
									"url":     "repo://1231123",
									"version": "7.7.0",
								},
							},
						}},
					},
				}),
				client: api.NewMock(mock.New200Response(hotWarmTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, hotWarmTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.12.0",
								UserBundles: []*models.ElasticsearchUserBundle{
									{
										URL:                  ec.String("repo://1231122"),
										Name:                 ec.String("my-bundle"),
										ElasticsearchVersion: ec.String("7.7.0"),
									},
									{
										URL:                  ec.String("repo://1231123"),
										Name:                 ec.String("my-second-bundle"),
										ElasticsearchVersion: ec.String("7.7.0"),
									},
								},
								UserPlugins: []*models.ElasticsearchUserPlugin{
									{
										URL:                  ec.String("repo://12311235"),
										Name:                 ec.String("my-second-plugin"),
										ElasticsearchVersion: ec.String("7.7.0"),
									},
									{
										URL:                  ec.String("repo://12311234"),
										Name:                 ec.String("my-plugin"),
										ElasticsearchVersion: ec.String("7.7.0"),
									},
								},
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-hot-warm-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeRoles: []string{
										"master",
										"ingest",
										"remote_cluster_client",
										"data_hot",
										"transform",
										"data_content",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d2",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "warm",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					}),
				},
			},
		},
		{
			name: "deployment with autoscaling enabled",
			args: args{
				d: util.NewResourceData(t, util.ResDataParams{
					ID:     mock.ValidClusterID,
					Schema: newSchema(),
					State: map[string]interface{}{
						"name":                   "my_deployment_name",
						"deployment_template_id": "aws-io-optimized-v2",
						"region":                 "us-east-1",
						"version":                "7.12.0",
						"elasticsearch": []interface{}{map[string]interface{}{
							"autoscale": "true",
							"topology": []interface{}{
								map[string]interface{}{
									"id":   "cold",
									"size": "2g",
								},
								map[string]interface{}{
									"id":   "hot_content",
									"size": "8g",
								},
								map[string]interface{}{
									"id":   "warm",
									"size": "4g",
								},
							},
						}},
					},
				}),
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(true),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.12.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "cold",
									ZoneCount:               1,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
									NodeRoles: []string{
										"data_cold",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "cold",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(59392),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(8192),
									},
									NodeRoles: []string{
										"master",
										"ingest",
										"remote_cluster_client",
										"data_hot",
										"transform",
										"data_content",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "warm",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					}),
				},
			},
		},
		{
			name: "deployment with autoscaling enabled and custom policies set",
			args: args{
				d: util.NewResourceData(t, util.ResDataParams{
					ID:     mock.ValidClusterID,
					Schema: newSchema(),
					State: map[string]interface{}{
						"name":                   "my_deployment_name",
						"deployment_template_id": "aws-io-optimized-v2",
						"region":                 "us-east-1",
						"version":                "7.12.0",
						"elasticsearch": []interface{}{map[string]interface{}{
							"autoscale": "true",
							"topology": []interface{}{
								map[string]interface{}{
									"id":   "cold",
									"size": "2g",
								},
								map[string]interface{}{
									"id":   "hot_content",
									"size": "8g",
									"autoscaling": []interface{}{map[string]interface{}{
										"max_size": "232g",
									}},
								},
								map[string]interface{}{
									"id":   "warm",
									"size": "4g",
									"autoscaling": []interface{}{map[string]interface{}{
										"max_size": "116g",
									}},
								},
							},
						}},
					},
				}),
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(true),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.12.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "cold",
									ZoneCount:               1,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
									NodeRoles: []string{
										"data_cold",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "cold",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(59392),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(8192),
									},
									NodeRoles: []string{
										"master",
										"ingest",
										"remote_cluster_client",
										"data_hot",
										"transform",
										"data_content",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(237568),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "warm",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					}),
				},
			},
		},
		{
			name: "deployment with dedicated master and cold tiers",
			args: args{
				d: util.NewResourceData(t, util.ResDataParams{
					ID:     mock.ValidClusterID,
					Schema: newSchema(),
					State: map[string]interface{}{
						"name":                   "my_deployment_name",
						"deployment_template_id": "aws-io-optimized-v2",
						"region":                 "us-east-1",
						"version":                "7.12.0",
						"elasticsearch": []interface{}{map[string]interface{}{
							"topology": []interface{}{
								map[string]interface{}{
									"id":   "cold",
									"size": "2g",
								},
								map[string]interface{}{
									"id":   "hot_content",
									"size": "8g",
								},
								map[string]interface{}{
									"id":         "master",
									"size":       "1g",
									"zone_count": 3,
								},
								map[string]interface{}{
									"id":   "warm",
									"size": "4g",
								},
							},
						}},
					},
				}),
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.12.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "cold",
									ZoneCount:               1,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
									NodeRoles: []string{
										"data_cold",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "cold",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(59392),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(8192),
									},
									NodeRoles: []string{
										"ingest",
										"remote_cluster_client",
										"data_hot",
										"transform",
										"data_content",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "master",
									ZoneCount:               3,
									InstanceConfigurationID: "aws.master.r5d",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
									NodeRoles: []string{
										"master",
										"remote_cluster_client",
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "warm",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					}),
				},
			},
		},
		{
			name: "deployment with dedicated coordinating and cold tiers",
			args: args{
				d: util.NewResourceData(t, util.ResDataParams{
					ID:     mock.ValidClusterID,
					Schema: newSchema(),
					State: map[string]interface{}{
						"name":                   "my_deployment_name",
						"deployment_template_id": "aws-io-optimized-v2",
						"region":                 "us-east-1",
						"version":                "7.12.0",
						"elasticsearch": []interface{}{map[string]interface{}{
							"topology": []interface{}{
								map[string]interface{}{
									"id":   "cold",
									"size": "2g",
								},
								map[string]interface{}{
									"id":         "coordinating",
									"size":       "2g",
									"zone_count": 2,
								},
								map[string]interface{}{
									"id":   "hot_content",
									"size": "8g",
								},
								map[string]interface{}{
									"id":   "warm",
									"size": "4g",
								},
							},
						}},
					},
				}),
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.12.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "cold",
									ZoneCount:               1,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
									NodeRoles: []string{
										"data_cold",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "cold",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(59392),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "coordinating",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.coordinating.m5d",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
									NodeRoles: []string{
										"ingest",
										"remote_cluster_client",
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
								},
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(8192),
									},
									NodeRoles: []string{
										"master",
										"remote_cluster_client",
										"data_hot",
										"transform",
										"data_content",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "warm",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					}),
				},
			},
		},
		{
			name: "deployment with dedicated coordinating, master and cold tiers",
			args: args{
				d: util.NewResourceData(t, util.ResDataParams{
					ID:     mock.ValidClusterID,
					Schema: newSchema(),
					State: map[string]interface{}{
						"name":                   "my_deployment_name",
						"deployment_template_id": "aws-io-optimized-v2",
						"region":                 "us-east-1",
						"version":                "7.12.0",
						"elasticsearch": []interface{}{map[string]interface{}{
							"topology": []interface{}{
								map[string]interface{}{
									"id":   "cold",
									"size": "2g",
								},
								map[string]interface{}{
									"id":         "coordinating",
									"size":       "2g",
									"zone_count": 2,
								},
								map[string]interface{}{
									"id":   "hot_content",
									"size": "8g",
								},
								map[string]interface{}{
									"id":         "master",
									"size":       "1g",
									"zone_count": 3,
								},
								map[string]interface{}{
									"id":   "warm",
									"size": "4g",
								},
							},
						}},
					},
				}),
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.12.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "cold",
									ZoneCount:               1,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
									NodeRoles: []string{
										"data_cold",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "cold",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(59392),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "coordinating",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.coordinating.m5d",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
									NodeRoles: []string{
										"ingest",
										"remote_cluster_client",
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
								},
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(8192),
									},
									NodeRoles: []string{
										"remote_cluster_client",
										"data_hot",
										"transform",
										"data_content",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "master",
									ZoneCount:               3,
									InstanceConfigurationID: "aws.master.r5d",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
									NodeRoles: []string{
										"master",
										"remote_cluster_client",
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "warm",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					}),
				},
			},
		},
		{
			name: "parses the resources with empty declarations (Cross Cluster Search)",
			args: args{
				d:      deploymentCCS,
				client: api.NewMock(mock.New200Response(ccsTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ccsTpl(), false), &models.ElasticsearchPayload{
						Region:   ec.String("us-east-1"),
						RefID:    ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{},
						Plan: &models.ElasticsearchClusterPlan{
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.9.2",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-cross-cluster-search-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "hot_content",
									ZoneCount:               1,
									InstanceConfigurationID: "aws.ccs.r5d",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
									NodeType: &models.ElasticsearchNodeType{
										Data:   ec.Bool(true),
										Ingest: ec.Bool(true),
										Master: ec.Bool(true),
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
								},
							},
						},
					}),
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
		{
			name: "parses the resources with tags",
			args: args{
				d:      deploymentWithTags,
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{Tags: []*models.MetadataItem{
					{Key: ec.String("aaa"), Value: ec.String("bbb")},
					{Key: ec.String("cost-center"), Value: ec.String("rnd")},
					{Key: ec.String("owner"), Value: ec.String("elastic")},
				}},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.10.1",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(8192),
								},
								NodeRoles: []string{
									"master",
									"ingest",
									"remote_cluster_client",
									"data_hot",
									"transform",
									"data_content",
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					}),
				},
			},
		},
		// This case we're using an empty deployment_template to ensure that
		// resources not present in the template cannot be expanded, receiving
		// an error instead.
		{
			name: "parses the resources with empty explicit declarations (Empty deployment template)",
			args: args{
				d:      deploymentEmptyTemplate,
				client: api.NewMock(mock.New200Response(emptyTpl())),
			},
			err: multierror.NewPrefixed("invalid configuration",
				errors.New("kibana specified but deployment template is not configured for it. Use a different template if you wish to add kibana"),
				errors.New("apm specified but deployment template is not configured for it. Use a different template if you wish to add apm"),
				errors.New("enterprise_search specified but deployment template is not configured for it. Use a different template if you wish to add enterprise_search"),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createResourceToModel(tt.args.d, tt.args.client)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_updateResourceToModel(t *testing.T) {
	deploymentRD := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleLegacyDeployment(),
		Schema: newSchema(),
	})
	var ioOptimizedTpl = func() io.ReadCloser {
		return fileAsResponseBody(t, "testdata/template-aws-io-optimized-v2.json")
	}
	deploymentEmptyRD := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleDeploymentEmptyRD(),
		Schema: newSchema(),
	})
	deploymentOverrideRd := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleDeploymentOverrides(),
		Schema: newSchema(),
	})

	hotWarmTpl := func() io.ReadCloser {
		return fileAsResponseBody(t, "testdata/template-aws-hot-warm-v2.json")
	}
	deploymentHotWarm := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		Schema: newSchema(),
		State: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-hot-warm-v2",
			"region":                 "us-east-1",
			"version":                "7.9.2",
			"elasticsearch":          []interface{}{map[string]interface{}{}},
			"kibana":                 []interface{}{map[string]interface{}{}},
		},
	})

	ccsTpl := func() io.ReadCloser {
		return fileAsResponseBody(t, "testdata/template-aws-cross-cluster-search-v2.json")
	}
	deploymentEmptyRDWithTemplateChange := util.NewResourceData(t, util.ResDataParams{
		ID:    mock.ValidClusterID,
		State: newSampleLegacyDeployment(),
		Change: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-cross-cluster-search-v2",
			"region":                 "us-east-1",
			"version":                "7.9.2",
			"elasticsearch":          []interface{}{map[string]interface{}{}},
			"kibana":                 []interface{}{map[string]interface{}{}},
		},
		Schema: newSchema(),
	})

	deploymentEmptyRDWithTemplateChangeWithDiffSize := util.NewResourceData(t, util.ResDataParams{
		ID: mock.ValidClusterID,
		State: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-io-optimized-v2",
			"region":                 "us-east-1",
			"version":                "7.9.2",
			"elasticsearch": []interface{}{
				map[string]interface{}{
					"topology": []interface{}{map[string]interface{}{
						"id":   "hot_content",
						"size": "16g",
					}},
				},
				map[string]interface{}{
					"topology": []interface{}{map[string]interface{}{
						"id":   "coordinating",
						"size": "16g",
					}},
				},
			},
			"kibana": []interface{}{map[string]interface{}{
				"topology": []interface{}{map[string]interface{}{
					"size": "2g",
				}},
			}},
			"apm": []interface{}{map[string]interface{}{
				"topology": []interface{}{map[string]interface{}{
					"size": "1g",
				}},
			}},
			"enterprise_search": []interface{}{map[string]interface{}{
				"topology": []interface{}{map[string]interface{}{
					"size": "2g",
				}},
			}},
		},
		Change: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-cross-cluster-search-v2",
			"region":                 "us-east-1",
			"version":                "7.9.2",
			"elasticsearch":          []interface{}{map[string]interface{}{}},
			"kibana":                 []interface{}{map[string]interface{}{}},
		},
		Schema: newSchema(),
	})

	emptyTpl := func() io.ReadCloser {
		return fileAsResponseBody(t, "testdata/template-empty.json")
	}
	deploymentChangeFromExplicitSizingToEmpty := util.NewResourceData(t, util.ResDataParams{
		ID: mock.ValidClusterID,
		State: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-io-optimized-v2",
			"region":                 "us-east-1",
			"version":                "7.9.2",
			"elasticsearch": []interface{}{
				map[string]interface{}{
					"topology": []interface{}{map[string]interface{}{
						"id":   "hot_content",
						"size": "16g",
					}},
				},
				map[string]interface{}{
					"topology": []interface{}{map[string]interface{}{
						"id":   "coordinating",
						"size": "16g",
					}},
				},
			},
			"kibana": []interface{}{map[string]interface{}{
				"topology": []interface{}{map[string]interface{}{
					"size": "2g",
				}},
			}},
			"apm": []interface{}{map[string]interface{}{
				"topology": []interface{}{map[string]interface{}{
					"size": "1g",
				}},
			}},
			"enterprise_search": []interface{}{map[string]interface{}{
				"topology": []interface{}{map[string]interface{}{
					"size": "8g",
				}},
			}},
		},
		Change: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-io-optimized-v2",
			"region":                 "us-east-1",
			"version":                "7.9.2",
			"elasticsearch":          []interface{}{map[string]interface{}{}},
			"kibana":                 []interface{}{map[string]interface{}{}},
			"apm":                    []interface{}{map[string]interface{}{}},
			"enterprise_search":      []interface{}{map[string]interface{}{}},
		},
		Schema: newSchema(),
	})

	deploymentChangeToEmptyDT := util.NewResourceData(t, util.ResDataParams{
		ID: mock.ValidClusterID,
		State: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-io-optimized-v2",
			"region":                 "us-east-1",
			"version":                "7.9.2",
			"elasticsearch": []interface{}{
				map[string]interface{}{
					"topology": []interface{}{map[string]interface{}{
						"id":   "hot_content",
						"size": "16g",
					}},
				},
				map[string]interface{}{
					"topology": []interface{}{map[string]interface{}{
						"id":   "coordinating",
						"size": "16g",
					}},
				},
			},
			"kibana": []interface{}{map[string]interface{}{
				"topology": []interface{}{map[string]interface{}{
					"size": "2g",
				}},
			}},
			"apm": []interface{}{map[string]interface{}{
				"topology": []interface{}{map[string]interface{}{
					"size": "1g",
				}},
			}},
			"enterprise_search": []interface{}{map[string]interface{}{
				"topology": []interface{}{map[string]interface{}{
					"size": "8g",
				}},
			}},
		},
		Change: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "empty-deployment-template",
			"region":                 "us-east-1",
			"version":                "7.9.2",
			"elasticsearch":          []interface{}{map[string]interface{}{}},
			"kibana":                 []interface{}{map[string]interface{}{}},
			"apm":                    []interface{}{map[string]interface{}{}},
			"enterprise_search":      []interface{}{map[string]interface{}{}},
		},
		Schema: newSchema(),
	})

	deploymentWithTags := util.NewResourceData(t, util.ResDataParams{
		ID: mock.ValidClusterID,
		State: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-io-optimized-v2",
			"region":                 "us-east-1",
			"version":                "7.10.1",
			"elasticsearch": []interface{}{
				map[string]interface{}{
					"version": "7.10.1",
					"topology": []interface{}{map[string]interface{}{
						"id":   "hot_content",
						"size": "8g",
					}},
				},
			},
		},
		Change: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-io-optimized-v2",
			"region":                 "us-east-1",
			"version":                "7.10.1",
			"elasticsearch": []interface{}{
				map[string]interface{}{
					"version": "7.10.1",
					"topology": []interface{}{map[string]interface{}{
						"id":   "hot_content",
						"size": "8g",
					}},
				},
			},
			"tags": map[string]interface{}{
				"aaa":         "bbb",
				"owner":       "elastic",
				"cost-center": "rnd",
			},
		},
		Schema: newSchema(),
	})

	type args struct {
		d      *schema.ResourceData
		client *api.API
	}
	tests := []struct {
		name string
		args args
		want *models.DeploymentUpdateRequest
		err  error
	}{
		{
			name: "parses the resources",
			args: args{
				d: deploymentRD,
				client: api.NewMock(
					mock.New200Response(ioOptimizedTpl()),
					mock.New200Response(
						mock.NewStructBody(models.DeploymentGetResponse{
							Healthy: ec.Bool(true),
							ID:      ec.String(mock.ValidClusterID),
							Resources: &models.DeploymentResources{
								Elasticsearch: []*models.ElasticsearchResourceInfo{{
									ID:    ec.String(mock.ValidClusterID),
									RefID: ec.String("main-elasticsearch"),
								}},
							},
						}),
					),
				),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				PruneOrphans: ec.Bool(true),
				Settings: &models.DeploymentUpdateSettings{
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
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
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
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					}),
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-apm"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{
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
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-enterprise_search"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
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
		{
			name: "parses the resources with empty declarations",
			args: args{
				d:      deploymentEmptyRD,
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				PruneOrphans: ec.Bool(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.7.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(8192),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   ec.Bool(true),
									Ingest: ec.Bool(true),
									Master: ec.Bool(true),
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					}),
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-apm"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{},
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
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-enterprise_search"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
								ClusterTopology: []*models.EnterpriseSearchTopologyElement{
									{
										ZoneCount:               2,
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
		{
			name: "parses the resources with topology overrides",
			args: args{
				d:      deploymentOverrideRd,
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				PruneOrphans: ec.Bool(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.7.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(4096),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   ec.Bool(true),
									Ingest: ec.Bool(true),
									Master: ec.Bool(true),
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					}),
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
								ClusterTopology: []*models.KibanaClusterTopologyElement{
									{
										ZoneCount:               1,
										InstanceConfigurationID: "aws.kibana.r5d",
										Size: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(2048),
										},
									},
								},
							},
						},
					},
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-apm"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{},
								ClusterTopology: []*models.ApmTopologyElement{{
									ZoneCount:               1,
									InstanceConfigurationID: "aws.apm.r5d",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								}},
							},
						},
					},
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-enterprise_search"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
								ClusterTopology: []*models.EnterpriseSearchTopologyElement{
									{
										ZoneCount:               2,
										InstanceConfigurationID: "aws.enterprisesearch.m5d",
										Size: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(4096),
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
		{
			name: "parses the resources with empty declarations (Hot Warm)",
			args: args{
				d:      deploymentHotWarm,
				client: api.NewMock(mock.New200Response(hotWarmTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				PruneOrphans: ec.Bool(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, hotWarmTpl(), false), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
							Curation:                  nil,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version:  "7.9.2",
								Curation: nil,
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-hot-warm-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeType: &models.ElasticsearchNodeType{
										Data:   ec.Bool(true),
										Ingest: ec.Bool(true),
										Master: ec.Bool(true),
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d2",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeType: &models.ElasticsearchNodeType{
										Data:   ec.Bool(true),
										Ingest: ec.Bool(true),
										Master: ec.Bool(false),
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "warm"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					}),
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
		{
			name: "toplogy change from hot / warm to cross cluster search",
			args: args{
				d:      deploymentEmptyRDWithTemplateChange,
				client: api.NewMock(mock.New200Response(ccsTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				PruneOrphans: ec.Bool(true),
				Settings: &models.DeploymentUpdateSettings{
					Observability: &models.DeploymentObservabilitySettings{},
				},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ccsTpl(), false), &models.ElasticsearchPayload{
						Region:   ec.String("us-east-1"),
						RefID:    ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{},
						Plan: &models.ElasticsearchClusterPlan{
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.9.2",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-cross-cluster-search-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID:                      "hot_content",
								ZoneCount:               1,
								InstanceConfigurationID: "aws.ccs.r5d",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   ec.Bool(true),
									Ingest: ec.Bool(true),
									Master: ec.Bool(true),
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								},
							}},
						},
					}),
					Kibana: []*models.KibanaPayload{{
						ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
						Region:                    ec.String("us-east-1"),
						RefID:                     ec.String("main-kibana"),
						Plan: &models.KibanaClusterPlan{
							Kibana: &models.KibanaConfiguration{},
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
					}},
				},
			},
		},
		// The behavior of this change should be:
		// * Resets the Elasticsearch topology: from 16g (due to unsetTopology call on DT change).
		// * Keeps the kibana toplogy size to 2g even though the topology element has been removed (saved value persists).
		// * Removes all other non present resources
		{
			name: "topology change with sizes not default from io optimized to cross cluster search",
			args: args{
				d:      deploymentEmptyRDWithTemplateChangeWithDiffSize,
				client: api.NewMock(mock.New200Response(ccsTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				PruneOrphans: ec.Bool(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ccsTpl(), false), &models.ElasticsearchPayload{
						Region:   ec.String("us-east-1"),
						RefID:    ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{},
						Plan: &models.ElasticsearchClusterPlan{
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.9.2",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-cross-cluster-search-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID:                      "hot_content",
								ZoneCount:               1,
								InstanceConfigurationID: "aws.ccs.r5d",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									// This field's value is reset.
									Value: ec.Int32(1024),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   ec.Bool(true),
									Ingest: ec.Bool(true),
									Master: ec.Bool(true),
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								},
							}},
						},
					}),
					Kibana: []*models.KibanaPayload{{
						ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
						Region:                    ec.String("us-east-1"),
						RefID:                     ec.String("main-kibana"),
						Plan: &models.KibanaClusterPlan{
							Kibana: &models.KibanaConfiguration{},
							ClusterTopology: []*models.KibanaClusterTopologyElement{
								{
									ZoneCount:               1,
									InstanceConfigurationID: "aws.kibana.r5d",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
								},
							},
						},
					}},
				},
			},
		},
		// The behavior of this change should be:
		// * Keeps all topology sizes as they were defined (saved value persists).
		{
			name: "topology change with sizes not default from explicit value to empty",
			args: args{
				d:      deploymentChangeFromExplicitSizingToEmpty,
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				PruneOrphans: ec.Bool(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.9.2",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(16384),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   ec.Bool(true),
									Ingest: ec.Bool(true),
									Master: ec.Bool(true),
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							},
							},
						},
					}),
					Kibana: []*models.KibanaPayload{{
						ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
						Region:                    ec.String("us-east-1"),
						RefID:                     ec.String("main-kibana"),
						Plan: &models.KibanaClusterPlan{
							Kibana: &models.KibanaConfiguration{},
							ClusterTopology: []*models.KibanaClusterTopologyElement{
								{
									ZoneCount:               1,
									InstanceConfigurationID: "aws.kibana.r5d",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
								},
							},
						},
					}},
					Apm: []*models.ApmPayload{{
						ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
						Region:                    ec.String("us-east-1"),
						RefID:                     ec.String("main-apm"),
						Plan: &models.ApmPlan{
							Apm: &models.ApmConfiguration{},
							ClusterTopology: []*models.ApmTopologyElement{{
								ZoneCount:               1,
								InstanceConfigurationID: "aws.apm.r5d",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(1024),
								},
							}},
						},
					}},
					EnterpriseSearch: []*models.EnterpriseSearchPayload{{
						ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
						Region:                    ec.String("us-east-1"),
						RefID:                     ec.String("main-enterprise_search"),
						Plan: &models.EnterpriseSearchPlan{
							EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
							ClusterTopology: []*models.EnterpriseSearchTopologyElement{
								{
									ZoneCount:               2,
									InstanceConfigurationID: "aws.enterprisesearch.m5d",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(8192),
									},
									NodeType: &models.EnterpriseSearchNodeTypes{
										Appserver: ec.Bool(true),
										Connector: ec.Bool(true),
										Worker:    ec.Bool(true),
									},
								},
							},
						},
					}},
				},
			},
		},
		{
			name: "does not migrate node_type to node_role on version upgrade",
			args: args{
				d: util.NewResourceData(t, util.ResDataParams{
					ID: mock.ValidClusterID,
					State: map[string]interface{}{
						"name":                   "my_deployment_name",
						"deployment_template_id": "aws-io-optimized-v2",
						"region":                 "us-east-1",
						"version":                "7.10.1",
						"elasticsearch": []interface{}{map[string]interface{}{
							"topology": []interface{}{map[string]interface{}{
								"id":               "hot_content",
								"size":             "16g",
								"node_type_data":   "true",
								"node_type_ingest": "true",
								"node_type_master": "true",
								"node_type_ml":     "false",
							}},
						}},
					},
					Change: map[string]interface{}{
						"name":                   "my_deployment_name",
						"deployment_template_id": "aws-io-optimized-v2",
						"region":                 "us-east-1",
						"version":                "7.11.1",
						"elasticsearch": []interface{}{map[string]interface{}{
							"topology": []interface{}{map[string]interface{}{
								"id":               "hot_content",
								"size":             "16g",
								"node_type_data":   "true",
								"node_type_ingest": "true",
								"node_type_master": "true",
								"node_type_ml":     "false",
							}},
						}},
					},
					Schema: newSchema(),
				}),
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				PruneOrphans: ec.Bool(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.11.1",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(16384),
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
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					}),
				},
			},
		},
		{
			name: "migrates node_type to node_role when the existing topology element size is updated",
			args: args{
				d: util.NewResourceData(t, util.ResDataParams{
					ID: mock.ValidClusterID,
					State: map[string]interface{}{
						"name":                   "my_deployment_name",
						"deployment_template_id": "aws-io-optimized-v2",
						"region":                 "us-east-1",
						"version":                "7.10.1",
						"elasticsearch": []interface{}{map[string]interface{}{
							"topology": []interface{}{map[string]interface{}{
								"id":               "hot_content",
								"size":             "16g",
								"node_type_data":   "true",
								"node_type_ingest": "true",
								"node_type_master": "true",
								"node_type_ml":     "false",
							}},
						}},
					},
					Change: map[string]interface{}{
						"name":                   "my_deployment_name",
						"deployment_template_id": "aws-io-optimized-v2",
						"region":                 "us-east-1",
						"version":                "7.10.1",
						"elasticsearch": []interface{}{map[string]interface{}{
							"topology": []interface{}{map[string]interface{}{
								"id":               "hot_content",
								"size":             "32g",
								"node_type_data":   "true",
								"node_type_ingest": "true",
								"node_type_master": "true",
								"node_type_ml":     "false",
							}},
						}},
					},
					Schema: newSchema(),
				}),
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				PruneOrphans: ec.Bool(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.10.1",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(32768),
								},
								NodeRoles: []string{
									"master",
									"ingest",
									"remote_cluster_client",
									"data_hot",
									"transform",
									"data_content",
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					}),
				},
			},
		},
		{
			name: "migrates node_type to node_role when the existing topology element size is updated and adds warm tier",
			args: args{
				d: util.NewResourceData(t, util.ResDataParams{
					ID: mock.ValidClusterID,
					State: map[string]interface{}{
						"name":                   "my_deployment_name",
						"deployment_template_id": "aws-io-optimized-v2",
						"region":                 "us-east-1",
						"version":                "7.10.1",
						"elasticsearch": []interface{}{map[string]interface{}{
							"topology": []interface{}{map[string]interface{}{
								"id":               "hot_content",
								"size":             "16g",
								"node_type_data":   "true",
								"node_type_ingest": "true",
								"node_type_master": "true",
								"node_type_ml":     "false",
							}},
						}},
					},
					Change: map[string]interface{}{
						"name":                   "my_deployment_name",
						"deployment_template_id": "aws-io-optimized-v2",
						"region":                 "us-east-1",
						"version":                "7.10.1",
						"elasticsearch": []interface{}{map[string]interface{}{
							"topology": []interface{}{
								map[string]interface{}{
									"id":               "hot_content",
									"size":             "16g",
									"node_type_data":   "true",
									"node_type_ingest": "true",
									"node_type_master": "true",
									"node_type_ml":     "false",
								},
								map[string]interface{}{
									"id":   "warm",
									"size": "8g",
								},
							},
						}},
					},
					Schema: newSchema(),
				}),
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				PruneOrphans: ec.Bool(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.10.1",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID: "hot_content",
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(16384),
									},
									NodeRoles: []string{
										"master",
										"ingest",
										"remote_cluster_client",
										"data_hot",
										"transform",
										"data_content",
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
								{
									ID: "warm",
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "warm"},
									},
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(8192),
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					}),
				},
			},
		},
		{
			name: "enables autoscaling with the default policies",
			args: args{
				d: util.NewResourceData(t, util.ResDataParams{
					ID: mock.ValidClusterID,
					State: map[string]interface{}{
						"name":                   "my_deployment_name",
						"deployment_template_id": "aws-io-optimized-v2",
						"region":                 "us-east-1",
						"version":                "7.12.1",
						"elasticsearch": []interface{}{map[string]interface{}{
							"topology": []interface{}{map[string]interface{}{
								"id":   "hot_content",
								"size": "16g",
							}},
						}},
					},
					Change: map[string]interface{}{
						"name":                   "my_deployment_name",
						"deployment_template_id": "aws-io-optimized-v2",
						"region":                 "us-east-1",
						"version":                "7.12.1",
						"elasticsearch": []interface{}{map[string]interface{}{
							"autoscale": "true",
							"topology": []interface{}{
								map[string]interface{}{
									"id":   "hot_content",
									"size": "16g",
								},
								map[string]interface{}{
									"id":   "warm",
									"size": "8g",
								},
							},
						}},
					},
					Schema: newSchema(),
				}),
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				PruneOrphans: ec.Bool(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(true),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.12.1",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID: "hot_content",
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(16384),
									},
									NodeRoles: []string{
										"master",
										"ingest",
										"remote_cluster_client",
										"data_hot",
										"transform",
										"data_content",
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
								{
									ID: "warm",
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "warm"},
									},
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(8192),
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					}),
				},
			},
		},
		{
			name: "parses the resources with tags",
			args: args{
				d:      deploymentWithTags,
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				PruneOrphans: ec.Bool(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{Tags: []*models.MetadataItem{
					{Key: ec.String("aaa"), Value: ec.String("bbb")},
					{Key: ec.String("cost-center"), Value: ec.String("rnd")},
					{Key: ec.String("owner"), Value: ec.String("elastic")},
				}},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.10.1",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(8192),
								},
								NodeRoles: []string{
									"master",
									"ingest",
									"remote_cluster_client",
									"data_hot",
									"transform",
									"data_content",
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					}),
				},
			},
		},
		{
			name: "topology change with invalid resources returns an error",
			args: args{
				d:      deploymentChangeToEmptyDT,
				client: api.NewMock(mock.New200Response(emptyTpl())),
			},
			err: multierror.NewPrefixed("invalid configuration",
				errors.New("kibana specified but deployment template is not configured for it. Use a different template if you wish to add kibana"),
				errors.New("apm specified but deployment template is not configured for it. Use a different template if you wish to add apm"),
				errors.New("enterprise_search specified but deployment template is not configured for it. Use a different template if you wish to add enterprise_search"),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := updateResourceToModel(tt.args.d, tt.args.client)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
