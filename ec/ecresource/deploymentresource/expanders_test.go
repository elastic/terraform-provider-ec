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
	buf.WriteString("[\n")
	if _, err := io.Copy(buf, f); err != nil {
		t.Fatal(err)
	}
	buf.WriteString("]\n")
	buf.WriteString("\n")

	return ioutil.NopCloser(buf)
}

func Test_createResourceToModel(t *testing.T) {
	deploymentRD := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleDeployment(),
		Schema: newSchema(),
	})
	ioOptimizedTpl := func() io.ReadCloser {
		return fileAsResponseBody(t, "testdata/aws-io-optimized-v2.json")
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
	deploymentOverrideICRd := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleDeploymentOverridesIC(),
		Schema: newSchema(),
	})
	hotWarmTpl := func() io.ReadCloser {
		return fileAsResponseBody(t, "testdata/aws-hot-warm-v2.json")
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
	deploymentHotWarmILM := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		Schema: newSchema(),
		State: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-hot-warm-v2",
			"region":                 "us-east-1",
			"version":                "7.9.2",
			"disable_curation":       true,
			"elasticsearch":          []interface{}{map[string]interface{}{}},
			"kibana":                 []interface{}{map[string]interface{}{}},
		},
	})
	ccsTpl := func() io.ReadCloser {
		return fileAsResponseBody(t, "testdata/aws-cross-cluster-search-v2.json")
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
		return fileAsResponseBody(t, "testdata/empty-deployment_template.json")
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
				Name: "my_deployment_name",
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
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{
						{
							Region: ec.String("us-east-1"),
							RefID:  ec.String("main-elasticsearch"),
							Settings: &models.ElasticsearchClusterSettings{
								DedicatedMastersThreshold: 6,
							},
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
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										UserSettingsYaml:         `some.setting: value`,
										UserSettingsOverrideYaml: `some.setting: value2`,
										UserSettingsJSON: map[string]interface{}{
											"some.setting": "value",
										},
										UserSettingsOverrideJSON: map[string]interface{}{
											"some.setting": "value2",
										},
									},
								}},
							},
						},
					},
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
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
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-apm"),
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
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-enterprise_search"),
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
		{
			name: "parses the resources with empty declarations (IO Optimized)",
			args: args{
				d:      deploymentEmptyRD,
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name: "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{
					TrafficFilterSettings: &models.TrafficFilterSettings{
						Rulesets: []string{"0.0.0.0/0", "192.168.10.0/24"},
					},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{
						{
							Region: ec.String("us-east-1"),
							RefID:  ec.String("main-elasticsearch"),
							Settings: &models.ElasticsearchClusterSettings{
								DedicatedMastersThreshold: 6,
							},
							Plan: &models.ElasticsearchClusterPlan{
								Elasticsearch: &models.ElasticsearchConfiguration{},
								DeploymentTemplate: &models.DeploymentTemplateReference{
									ID: ec.String("aws-io-optimized-v2"),
								},
								ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
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
								}},
							},
						},
					},
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
				Name: "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{
					TrafficFilterSettings: &models.TrafficFilterSettings{
						Rulesets: []string{"0.0.0.0/0", "192.168.10.0/24"},
					},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{
						{
							Region: ec.String("us-east-1"),
							RefID:  ec.String("main-elasticsearch"),
							Settings: &models.ElasticsearchClusterSettings{
								DedicatedMastersThreshold: 6,
							},
							Plan: &models.ElasticsearchClusterPlan{
								Elasticsearch: &models.ElasticsearchConfiguration{},
								DeploymentTemplate: &models.DeploymentTemplateReference{
									ID: ec.String("aws-io-optimized-v2"),
								},
								ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
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
								}},
							},
						},
					},
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
				Name: "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{
					TrafficFilterSettings: &models.TrafficFilterSettings{
						Rulesets: []string{"0.0.0.0/0", "192.168.10.0/24"},
					},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{
						{
							Region: ec.String("us-east-1"),
							RefID:  ec.String("main-elasticsearch"),
							Settings: &models.ElasticsearchClusterSettings{
								DedicatedMastersThreshold: 6,
							},
							Plan: &models.ElasticsearchClusterPlan{
								Elasticsearch: &models.ElasticsearchConfiguration{},
								DeploymentTemplate: &models.DeploymentTemplateReference{
									ID: ec.String("aws-io-optimized-v2"),
								},
								ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
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
								}},
							},
						},
					},
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
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{
						{
							Region: ec.String("us-east-1"),
							RefID:  ec.String("main-elasticsearch"),
							Settings: &models.ElasticsearchClusterSettings{
								DedicatedMastersThreshold: 6,
								Curation: &models.ClusterCurationSettings{
									Specs: []*models.ClusterCurationSpec{
										{
											IndexPattern:           ec.String("logstash-*"),
											TriggerIntervalSeconds: ec.Int32(86400),
										},
										{
											IndexPattern:           ec.String("filebeat-*"),
											TriggerIntervalSeconds: ec.Int32(86400),
										},
										{
											IndexPattern:           ec.String("metricbeat-*"),
											TriggerIntervalSeconds: ec.Int32(86400),
										},
									},
								},
							},
							Plan: &models.ElasticsearchClusterPlan{
								Elasticsearch: &models.ElasticsearchConfiguration{
									Curation: &models.ElasticsearchCuration{
										FromInstanceConfigurationID: ec.String("aws.data.highio.i3"),
										ToInstanceConfigurationID:   ec.String("aws.data.highstorage.d2"),
									},
								},
								DeploymentTemplate: &models.DeploymentTemplateReference{
									ID: ec.String("aws-hot-warm-v2"),
								},
								ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
									{
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
											NodeAttributes: map[string]string{
												"data": "hot",
											},
										},
									},
									{
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
									},
								},
							},
						},
					},
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
			name: "parses the resources with empty declarations (Hot Warm) and curation disabled",
			args: args{
				d:      deploymentHotWarmILM,
				client: api.NewMock(mock.New200Response(hotWarmTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{
						{
							Region: ec.String("us-east-1"),
							RefID:  ec.String("main-elasticsearch"),
							Settings: &models.ElasticsearchClusterSettings{
								DedicatedMastersThreshold: 6,
								Curation:                  nil,
							},
							Plan: &models.ElasticsearchClusterPlan{
								Elasticsearch: &models.ElasticsearchConfiguration{
									Curation: nil,
								},
								DeploymentTemplate: &models.DeploymentTemplateReference{
									ID: ec.String("aws-hot-warm-v2"),
								},
								ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
									{
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
											NodeAttributes: map[string]string{
												"data": "hot",
											},
										},
									},
									{
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
									},
								},
							},
						},
					},
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
			name: "parses the resources with empty declarations (Cross Cluster Search)",
			args: args{
				d:      deploymentCCS,
				client: api.NewMock(mock.New200Response(ccsTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{
						{
							Region:   ec.String("us-east-1"),
							RefID:    ec.String("main-elasticsearch"),
							Settings: &models.ElasticsearchClusterSettings{},
							Plan: &models.ElasticsearchClusterPlan{
								Elasticsearch: &models.ElasticsearchConfiguration{},
								DeploymentTemplate: &models.DeploymentTemplateReference{
									ID: ec.String("aws-cross-cluster-search-v2"),
								},
								ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
									{
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
									},
								},
							},
						},
					},
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
		State:  newSampleDeployment(),
		Schema: newSchema(),
	})
	var ioOptimizedTpl = func() io.ReadCloser {
		return fileAsResponseBody(t, "testdata/aws-io-optimized-v2.json")
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
		return fileAsResponseBody(t, "testdata/aws-hot-warm-v2.json")
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
	deploymentHotWarmILM := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		Schema: newSchema(),
		State: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-hot-warm-v2",
			"region":                 "us-east-1",
			"version":                "7.9.2",
			"disable_curation":       true,
			"elasticsearch":          []interface{}{map[string]interface{}{}},
			"kibana":                 []interface{}{map[string]interface{}{}},
		},
	})

	ccsTpl := func() io.ReadCloser {
		return fileAsResponseBody(t, "testdata/aws-cross-cluster-search-v2.json")
	}
	deploymentEmptyRDWithTemplateChange := util.NewResourceData(t, util.ResDataParams{
		ID:    mock.ValidClusterID,
		State: newSampleDeployment(),
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
						"instance_configuration_id": "aws.data.highio.i3",
						"size":                      "16g",
					}},
				},
				map[string]interface{}{
					"topology": []interface{}{map[string]interface{}{
						"instance_configuration_id": "aws.coordinating.m5d",
						"size":                      "16g",
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
		return fileAsResponseBody(t, "testdata/empty-deployment_template.json")
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
						"instance_configuration_id": "aws.data.highio.i3",
						"size":                      "16g",
					}},
				},
				map[string]interface{}{
					"topology": []interface{}{map[string]interface{}{
						"instance_configuration_id": "aws.coordinating.m5d",
						"size":                      "16g",
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
						"instance_configuration_id": "aws.data.highio.i3",
						"size":                      "16g",
					}},
				},
				map[string]interface{}{
					"topology": []interface{}{map[string]interface{}{
						"instance_configuration_id": "aws.coordinating.m5d",
						"size":                      "16g",
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
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{
						{
							Region: ec.String("us-east-1"),
							RefID:  ec.String("main-elasticsearch"),
							Settings: &models.ElasticsearchClusterSettings{
								DedicatedMastersThreshold: 6,
							},
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
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										UserSettingsYaml:         `some.setting: value`,
										UserSettingsOverrideYaml: `some.setting: value2`,
										UserSettingsJSON: map[string]interface{}{
											"some.setting": "value",
										},
										UserSettingsOverrideJSON: map[string]interface{}{
											"some.setting": "value2",
										},
									},
								}},
							},
						},
					},
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
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
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-apm"),
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
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-enterprise_search"),
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
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{
						{
							Region: ec.String("us-east-1"),
							RefID:  ec.String("main-elasticsearch"),
							Settings: &models.ElasticsearchClusterSettings{
								DedicatedMastersThreshold: 6,
							},
							Plan: &models.ElasticsearchClusterPlan{
								Elasticsearch: &models.ElasticsearchConfiguration{},
								DeploymentTemplate: &models.DeploymentTemplateReference{
									ID: ec.String("aws-io-optimized-v2"),
								},
								ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
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
								}},
							},
						},
					},
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
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{
						{
							Region: ec.String("us-east-1"),
							RefID:  ec.String("main-elasticsearch"),
							Settings: &models.ElasticsearchClusterSettings{
								DedicatedMastersThreshold: 6,
							},
							Plan: &models.ElasticsearchClusterPlan{
								Elasticsearch: &models.ElasticsearchConfiguration{},
								DeploymentTemplate: &models.DeploymentTemplateReference{
									ID: ec.String("aws-io-optimized-v2"),
								},
								ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
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
								}},
							},
						},
					},
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
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{
						{
							Region: ec.String("us-east-1"),
							RefID:  ec.String("main-elasticsearch"),
							Settings: &models.ElasticsearchClusterSettings{
								DedicatedMastersThreshold: 6,
								Curation: &models.ClusterCurationSettings{
									Specs: []*models.ClusterCurationSpec{
										{
											IndexPattern:           ec.String("logstash-*"),
											TriggerIntervalSeconds: ec.Int32(86400),
										},
										{
											IndexPattern:           ec.String("filebeat-*"),
											TriggerIntervalSeconds: ec.Int32(86400),
										},
										{
											IndexPattern:           ec.String("metricbeat-*"),
											TriggerIntervalSeconds: ec.Int32(86400),
										},
									},
								},
							},
							Plan: &models.ElasticsearchClusterPlan{
								Elasticsearch: &models.ElasticsearchConfiguration{
									Curation: &models.ElasticsearchCuration{
										FromInstanceConfigurationID: ec.String("aws.data.highio.i3"),
										ToInstanceConfigurationID:   ec.String("aws.data.highstorage.d2"),
									},
								},
								DeploymentTemplate: &models.DeploymentTemplateReference{
									ID: ec.String("aws-hot-warm-v2"),
								},
								ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
									{
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
											NodeAttributes: map[string]string{
												"data": "hot",
											},
										},
									},
									{
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
									},
								},
							},
						},
					},
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
			name: "parses the resources with empty declarations (Hot Warm) and curation disabled",
			args: args{
				d:      deploymentHotWarmILM,
				client: api.NewMock(mock.New200Response(hotWarmTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				PruneOrphans: ec.Bool(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{
						{
							Region: ec.String("us-east-1"),
							RefID:  ec.String("main-elasticsearch"),
							Settings: &models.ElasticsearchClusterSettings{
								DedicatedMastersThreshold: 6,
								Curation:                  nil,
							},
							Plan: &models.ElasticsearchClusterPlan{
								Elasticsearch: &models.ElasticsearchConfiguration{
									Curation: nil,
								},
								DeploymentTemplate: &models.DeploymentTemplateReference{
									ID: ec.String("aws-hot-warm-v2"),
								},
								ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
									{
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
											NodeAttributes: map[string]string{
												"data": "hot",
											},
										},
									},
									{
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
									},
								},
							},
						},
					},
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
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{{
						Region:   ec.String("us-east-1"),
						RefID:    ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{},
						Plan: &models.ElasticsearchClusterPlan{
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.7.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-cross-cluster-search-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
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
							}},
						},
					}},
					Kibana: []*models.KibanaPayload{{
						ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
						Region:                    ec.String("us-east-1"),
						RefID:                     ec.String("main-kibana"),
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
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{{
						Region:   ec.String("us-east-1"),
						RefID:    ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{},
						Plan: &models.ElasticsearchClusterPlan{
							Elasticsearch: &models.ElasticsearchConfiguration{},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-cross-cluster-search-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
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
							}},
						},
					}},
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
			name: "topology change with sizes not default from io optimized to cross cluster search",
			args: args{
				d:      deploymentChangeFromExplicitSizingToEmpty,
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				PruneOrphans: ec.Bool(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							Elasticsearch: &models.ElasticsearchConfiguration{},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
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
							}},
						},
					}},
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
