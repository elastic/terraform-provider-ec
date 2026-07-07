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
	"github.com/elastic/cloud-sdk-go/pkg/client/deployments"
	"io"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	apmv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/apm/v2"
	enterprisesearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/enterprisesearch/v2"
	kibanav2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/kibana/v2"
	observabilityv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/observability/v2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	elasticsearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
)

func Test_updateResourceToModel(t *testing.T) {
	defaultHotTier := elasticsearchv2.CreateTierForTest(
		"hot_content",
		elasticsearchv2.ElasticsearchTopology{
			Autoscaling: &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
		},
	)

	defaultElasticsearch := &elasticsearchv2.Elasticsearch{
		HotTier: defaultHotTier,
	}

	var ioOptimizedTpl = func() io.ReadCloser {
		return fileAsResponseBody(t, "../../testdata/template-aws-io-optimized-v2.json")
	}

	hotWarmTpl := func() io.ReadCloser {
		return fileAsResponseBody(t, "../../testdata/template-aws-hot-warm-v2.json")
	}

	ccsTpl := func() io.ReadCloser {
		return fileAsResponseBody(t, "../../testdata/template-aws-cross-cluster-search-v2.json")
	}

	ccsDeploymentUpdate := func() io.ReadCloser {
		return fileAsResponseBody(t, "../../testdata/deployment-update-aws-cross-cluster-search-v2.json")
	}

	type args struct {
		plan   Deployment
		state  *Deployment
		client *api.API
	}
	tests := []struct {
		name  string
		args  args
		want  *models.DeploymentUpdateRequest
		diags diag.Diagnostics
	}{
		{
			name: "parses the resources",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Alias:                "my-deployment",
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.7.0",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId:      new("main-elasticsearch"),
						ResourceId: new(mock.ValidClusterID),
						Region:     new("us-east-1"),
						Config: &elasticsearchv2.ElasticsearchConfig{
							UserSettingsYaml:         new("some.setting: value"),
							UserSettingsOverrideYaml: new("some.setting: value2"),
							UserSettingsJson:         new("{\"some.setting\":\"value\"}"),
							UserSettingsOverrideJson: new("{\"some.setting\":\"value2\"}"),
						},
						HotTier: elasticsearchv2.CreateTierForTest(
							"hot_content",
							elasticsearchv2.ElasticsearchTopology{
								InstanceConfigurationId: new("aws.data.highio.i3"),
								Size:                    new("2g"),
								NodeTypeData:            new("true"),
								NodeTypeIngest:          new("true"),
								NodeTypeMaster:          new("true"),
								NodeTypeMl:              new("false"),
								ZoneCount:               1,
								Autoscaling:             &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
					},
					Kibana: &kibanav2.Kibana{
						ElasticsearchClusterRefId: new("main-elasticsearch"),
						RefId:                     new("main-kibana"),
						ResourceId:                new(mock.ValidClusterID),
						Region:                    new("us-east-1"),
						InstanceConfigurationId:   new("aws.kibana.r5d"),
						Size:                      new("1g"),
						ZoneCount:                 1,
					},
					Apm: &apmv2.Apm{
						ElasticsearchClusterRefId: new("main-elasticsearch"),
						RefId:                     new("main-apm"),
						ResourceId:                new(mock.ValidClusterID),
						Region:                    new("us-east-1"),
						Config: &apmv2.ApmConfig{
							DebugEnabled: new(false),
						},
						InstanceConfigurationId: new("aws.apm.r5d"),
						Size:                    new("0.5g"),
						ZoneCount:               1,
					},
					EnterpriseSearch: &enterprisesearchv2.EnterpriseSearch{
						ElasticsearchClusterRefId: new("main-elasticsearch"),
						RefId:                     new("main-enterprise_search"),
						ResourceId:                new(mock.ValidClusterID),
						Region:                    new("us-east-1"),
						InstanceConfigurationId:   new("aws.enterprisesearch.m5d"),
						Size:                      new("2g"),
						ZoneCount:                 1,
						NodeTypeAppserver:         new(true),
						NodeTypeConnector:         new(true),
						NodeTypeWorker:            new(true),
					},
					Observability: &observabilityv2.Observability{
						DeploymentId: new(mock.ValidClusterID),
						RefId:        new("main-elasticsearch"),
						Logs:         true,
						Metrics:      true,
					},
					TrafficFilter: []string{"0.0.0.0/0", "192.168.10.0/24"},
				},
				client: api.NewMock(
					mock.New200Response(ioOptimizedTpl()),
					mock.New200Response(
						mock.NewStructBody(models.DeploymentGetResponse{
							Healthy: new(true),
							ID:      new(mock.ValidClusterID),
							Resources: &models.DeploymentResources{
								Elasticsearch: []*models.ElasticsearchResourceInfo{{
									ID:    new(mock.ValidClusterID),
									RefID: new("main-elasticsearch"),
								}},
							},
						}),
					),
				),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				Alias:        new("my-deployment"),
				PruneOrphans: new(true),
				Settings: &models.DeploymentUpdateSettings{
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
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{elasticsearchv2.EnrichWithEmptyTopologies(elasticsearchPayloadFromReader(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: new("us-east-1"),
						RefID:  new("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: new(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version:                  "7.7.0",
								UserSettingsYaml:         `some.setting: value`,
								UserSettingsOverrideYaml: `some.setting: value2`,
								UserSettingsJSON: map[string]any{
									"some.setting": "value",
								},
								UserSettingsOverrideJSON: map[string]any{
									"some.setting": "value2",
								},
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: new("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               1,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: new("memory"),
									Value:    ec.Int32(2048),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   new(true),
									Ingest: new(true),
									Master: new(true),
									Ml:     new(false),
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: new("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: new("memory"),
								},
							}},
						},
					})},
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: new("main-elasticsearch"),
							Region:                    new("us-east-1"),
							RefID:                     new("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
								ClusterTopology: []*models.KibanaClusterTopologyElement{
									{
										ZoneCount:               1,
										InstanceConfigurationID: "aws.kibana.r5d",
										Size: &models.TopologySize{
											Resource: new("memory"),
											Value:    ec.Int32(1024),
										},
									},
								},
							},
						},
					},
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: new("main-elasticsearch"),
							Region:                    new("us-east-1"),
							RefID:                     new("main-apm"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{
									SystemSettings: &models.ApmSystemSettings{
										DebugEnabled: new(false),
									},
								},
								ClusterTopology: []*models.ApmTopologyElement{{
									ZoneCount:               1,
									InstanceConfigurationID: "aws.apm.r5d",
									Size: &models.TopologySize{
										Resource: new("memory"),
										Value:    ec.Int32(512),
									},
								}},
							},
						},
					},
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: new("main-elasticsearch"),
							Region:                    new("us-east-1"),
							RefID:                     new("main-enterprise_search"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
								ClusterTopology: []*models.EnterpriseSearchTopologyElement{
									{
										ZoneCount:               1,
										InstanceConfigurationID: "aws.enterprisesearch.m5d",
										Size: &models.TopologySize{
											Resource: new("memory"),
											Value:    ec.Int32(2048),
										},
										NodeType: &models.EnterpriseSearchNodeTypes{
											Appserver: new(true),
											Connector: new(true),
											Worker:    new(true),
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
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Alias:                "",
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.7.0",
					Elasticsearch:        defaultElasticsearch,
					Kibana:               &kibanav2.Kibana{},
					Apm:                  &apmv2.Apm{},
					EnterpriseSearch:     &enterprisesearchv2.EnterpriseSearch{},
					TrafficFilter:        []string{"0.0.0.0/0", "192.168.10.0/24"},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				Alias:        new(""),
				PruneOrphans: new(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{elasticsearchv2.EnrichWithEmptyTopologies(elasticsearchPayloadFromReader(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: new("us-east-1"),
						RefID:  new("es-ref-id"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: new(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.7.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: new("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: new("memory"),
									Value:    ec.Int32(8192),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   new(true),
									Ingest: new(true),
									Master: new(true),
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: new("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: new("memory"),
								},
							}},
						},
					})},
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: new("es-ref-id"),
							Region:                    new("us-east-1"),
							RefID:                     new("kibana-ref-id"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
								ClusterTopology: []*models.KibanaClusterTopologyElement{
									{
										ZoneCount:               1,
										InstanceConfigurationID: "aws.kibana.r5d",
										Size: &models.TopologySize{
											Resource: new("memory"),
											Value:    ec.Int32(1024),
										},
									},
								},
							},
						},
					},
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: new("es-ref-id"),
							Region:                    new("us-east-1"),
							RefID:                     new("apm-ref-id"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{},
								ClusterTopology: []*models.ApmTopologyElement{{
									ZoneCount:               1,
									InstanceConfigurationID: "aws.apm.r5d",
									Size: &models.TopologySize{
										Resource: new("memory"),
										Value:    ec.Int32(512),
									},
								}},
							},
						},
					},
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: new("es-ref-id"),
							Region:                    new("us-east-1"),
							RefID:                     new("enterprise_search-ref-id"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
								ClusterTopology: []*models.EnterpriseSearchTopologyElement{
									{
										ZoneCount:               2,
										InstanceConfigurationID: "aws.enterprisesearch.m5d",
										Size: &models.TopologySize{
											Resource: new("memory"),
											Value:    ec.Int32(2048),
										},
										NodeType: &models.EnterpriseSearchNodeTypes{
											Appserver: new(true),
											Connector: new(true),
											Worker:    new(true),
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
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Alias:                "my-deployment",
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.7.0",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId: new("main-elasticsearch"),
						HotTier: elasticsearchv2.CreateTierForTest(
							"hot_content",
							elasticsearchv2.ElasticsearchTopology{
								Size:        new("4g"),
								Autoscaling: &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
					},
					Kibana: &kibanav2.Kibana{
						RefId:                     new("main-kibana"),
						ElasticsearchClusterRefId: new("main-elasticsearch"),
						Size:                      new("2g"),
					},
					Apm: &apmv2.Apm{
						RefId:                     new("main-apm"),
						ElasticsearchClusterRefId: new("main-elasticsearch"),
						Size:                      new("1g"),
					},
					EnterpriseSearch: &enterprisesearchv2.EnterpriseSearch{
						RefId:                     new("main-enterprise_search"),
						ElasticsearchClusterRefId: new("main-elasticsearch"),
						Size:                      new("4g"),
					},
					TrafficFilter: []string{"0.0.0.0/0", "192.168.10.0/24"},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				Alias:        new("my-deployment"),
				PruneOrphans: new(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{elasticsearchv2.EnrichWithEmptyTopologies(elasticsearchPayloadFromReader(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: new("us-east-1"),
						RefID:  new("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: new(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.7.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: new("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: new("memory"),
									Value:    ec.Int32(4096),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   new(true),
									Ingest: new(true),
									Master: new(true),
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: new("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: new("memory"),
								},
							}},
						},
					})},
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: new("main-elasticsearch"),
							Region:                    new("us-east-1"),
							RefID:                     new("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
								ClusterTopology: []*models.KibanaClusterTopologyElement{
									{
										ZoneCount:               1,
										InstanceConfigurationID: "aws.kibana.r5d",
										Size: &models.TopologySize{
											Resource: new("memory"),
											Value:    ec.Int32(2048),
										},
									},
								},
							},
						},
					},
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: new("main-elasticsearch"),
							Region:                    new("us-east-1"),
							RefID:                     new("main-apm"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{},
								ClusterTopology: []*models.ApmTopologyElement{{
									ZoneCount:               1,
									InstanceConfigurationID: "aws.apm.r5d",
									Size: &models.TopologySize{
										Resource: new("memory"),
										Value:    ec.Int32(1024),
									},
								}},
							},
						},
					},
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: new("main-elasticsearch"),
							Region:                    new("us-east-1"),
							RefID:                     new("main-enterprise_search"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
								ClusterTopology: []*models.EnterpriseSearchTopologyElement{
									{
										ZoneCount:               2,
										InstanceConfigurationID: "aws.enterprisesearch.m5d",
										Size: &models.TopologySize{
											Resource: new("memory"),
											Value:    ec.Int32(4096),
										},
										NodeType: &models.EnterpriseSearchNodeTypes{
											Appserver: new(true),
											Connector: new(true),
											Worker:    new(true),
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
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-hot-warm-v2",
					Region:               "us-east-1",
					Version:              "7.9.2",
					Elasticsearch:        defaultElasticsearch,
					Kibana:               &kibanav2.Kibana{},
				},
				client: api.NewMock(mock.New200Response(hotWarmTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				Alias:        new(""),
				PruneOrphans: new(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{elasticsearchv2.EnrichWithEmptyTopologies(elasticsearchPayloadFromReader(t, hotWarmTpl(), false), &models.ElasticsearchPayload{
						Region: new("us-east-1"),
						RefID:  new("es-ref-id"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
							Curation:                  nil,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: new(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version:  "7.9.2",
								Curation: nil,
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: new("aws-hot-warm-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: new("memory"),
										Value:    ec.Int32(4096),
									},
									NodeType: &models.ElasticsearchNodeType{
										Data:   new(true),
										Ingest: new(true),
										Master: new(true),
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: new("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: new("memory"),
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d2",
									Size: &models.TopologySize{
										Resource: new("memory"),
										Value:    ec.Int32(4096),
									},
									NodeType: &models.ElasticsearchNodeType{
										Data:   new(true),
										Ingest: new(true),
										Master: new(false),
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "warm"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: new("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: new("memory"),
									},
								},
							},
						},
					})},
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: new("es-ref-id"),
							Region:                    new("us-east-1"),
							RefID:                     new("kibana-ref-id"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
								ClusterTopology: []*models.KibanaClusterTopologyElement{
									{
										ZoneCount:               1,
										InstanceConfigurationID: "aws.kibana.r5d",
										Size: &models.TopologySize{
											Resource: new("memory"),
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
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Alias:                "my-deployment",
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-cross-cluster-search-v2",
					Region:               "us-east-1",
					Version:              "7.9.2",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId:   new("main-elasticsearch"),
						HotTier: defaultHotTier,
					},
					Kibana: &kibanav2.Kibana{
						RefId:                     new("main-kibana"),
						ElasticsearchClusterRefId: new("main-elasticsearch"),
					},
				},
				state: &Deployment{
					Id:                   mock.ValidClusterID,
					Alias:                "my-deployment",
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.7.0",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId:      new("main-elasticsearch"),
						ResourceId: new(mock.ValidClusterID),
						Region:     new("us-east-1"),
						Config: &elasticsearchv2.ElasticsearchConfig{
							UserSettingsYaml:         new("some.setting: value"),
							UserSettingsOverrideYaml: new("some.setting: value2"),
							UserSettingsJson:         new("{\"some.setting\":\"value\"}"),
							UserSettingsOverrideJson: new("{\"some.setting\":\"value2\"}"),
						},
						HotTier: elasticsearchv2.CreateTierForTest(
							"hot_content",
							elasticsearchv2.ElasticsearchTopology{
								InstanceConfigurationId: new("aws.data.highio.i3"),
								Size:                    new("2g"),
								NodeTypeData:            new("true"),
								NodeTypeIngest:          new("true"),
								NodeTypeMaster:          new("true"),
								NodeTypeMl:              new("false"),
								ZoneCount:               1,
								Autoscaling:             &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
					},
					Kibana: &kibanav2.Kibana{
						ElasticsearchClusterRefId: new("main-elasticsearch"),
						RefId:                     new("main-kibana"),
						ResourceId:                new(mock.ValidClusterID),
						Region:                    new("us-east-1"),
						InstanceConfigurationId:   new("aws.kibana.r5d"),
						Size:                      new("1g"),
						ZoneCount:                 1,
					},
					Apm: &apmv2.Apm{
						ElasticsearchClusterRefId: new("main-elasticsearch"),
						RefId:                     new("main-apm"),
						ResourceId:                new(mock.ValidClusterID),
						Region:                    new("us-east-1"),
						Config: &apmv2.ApmConfig{
							DebugEnabled: new(false),
						},
						InstanceConfigurationId: new("aws.apm.r5d"),
						Size:                    new("0.5g"),
						ZoneCount:               1,
					},
					EnterpriseSearch: &enterprisesearchv2.EnterpriseSearch{
						ElasticsearchClusterRefId: new("main-elasticsearch"),
						RefId:                     new("main-enterprise_search"),
						ResourceId:                new(mock.ValidClusterID),
						Region:                    new("us-east-1"),
						InstanceConfigurationId:   new("aws.enterprisesearch.m5d"),
						Size:                      new("2g"),
						ZoneCount:                 1,
						NodeTypeAppserver:         new(true),
						NodeTypeConnector:         new(true),
						NodeTypeWorker:            new(true),
					},
					Observability: &observabilityv2.Observability{
						DeploymentId: new(mock.ValidClusterID),
						RefId:        new("main-elasticsearch"),
						Logs:         true,
						Metrics:      true,
					},
					TrafficFilter: []string{"0.0.0.0/0", "192.168.10.0/24"},
				},
				client: api.NewMock(
					mock.New200Response(ccsTpl()),
					mock.New200Response(ccsDeploymentUpdate()),
				),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				Alias:        new("my-deployment"),
				PruneOrphans: new(true),
				Settings: &models.DeploymentUpdateSettings{
					Observability: &models.DeploymentObservabilitySettings{},
				},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{elasticsearchv2.EnrichWithEmptyTopologies(elasticsearchPayloadFromReader(t, ccsTpl(), false), &models.ElasticsearchPayload{
						Region:   new("us-east-1"),
						RefID:    new("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{},
						Plan: &models.ElasticsearchClusterPlan{
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.9.2",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: new("aws-cross-cluster-search-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID:                      "hot_content",
								ZoneCount:               1,
								InstanceConfigurationID: "aws.ccs.r5d",
								Size: &models.TopologySize{
									Resource: new("memory"),
									Value:    ec.Int32(1024),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   new(true),
									Ingest: new(true),
									Master: new(true),
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: new("memory"),
										Value:    ec.Int32(1024),
									},
								},
							}},
						},
					})},
					Kibana: []*models.KibanaPayload{{
						ElasticsearchClusterRefID: new("main-elasticsearch"),
						Region:                    new("us-east-1"),
						RefID:                     new("main-kibana"),
						Plan: &models.KibanaClusterPlan{
							Kibana: &models.KibanaConfiguration{},
							ClusterTopology: []*models.KibanaClusterTopologyElement{
								{
									ZoneCount:               1,
									InstanceConfigurationID: "aws.kibana.r5d",
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

		// The behavior of this change should be:
		// * Resets the Elasticsearch topology: from 16g (due to unsetTopology call on DT change).
		// * Keeps the kibana toplogy size to 2g even though the topology element has been removed (saved value persists).
		// * Removes all other non present resources
		{
			name: "topology change with sizes not default from io optimized to cross cluster search",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-cross-cluster-search-v2",
					Region:               "us-east-1",
					Version:              "7.9.2",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId:   new("main-elasticsearch"),
						HotTier: defaultHotTier,
					},
					Kibana: &kibanav2.Kibana{
						ElasticsearchClusterRefId: new("main-elasticsearch"),
						RefId:                     new("main-kibana"),
					},
				},
				state: &Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.9.2",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId: new("main-elasticsearch"),
						HotTier: elasticsearchv2.CreateTierForTest(
							"hot_content",
							elasticsearchv2.ElasticsearchTopology{
								Size:        new("16g"),
								Autoscaling: &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
						CoordinatingTier: elasticsearchv2.CreateTierForTest(
							"coordinating",
							elasticsearchv2.ElasticsearchTopology{
								Size:        new("16g"),
								Autoscaling: &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
					},
					Kibana: &kibanav2.Kibana{
						ElasticsearchClusterRefId: new("main-elasticsearch"),
						RefId:                     new("main-kibana"),
						Size:                      new("2g"),
					},
					Apm: &apmv2.Apm{
						ElasticsearchClusterRefId: new("main-elasticsearch"),
						RefId:                     new("main-apm"),
						Size:                      new("1g"),
					},
					EnterpriseSearch: &enterprisesearchv2.EnterpriseSearch{
						ElasticsearchClusterRefId: new("main-elasticsearch"),
						RefId:                     new("main-enterprise_search"),
						Size:                      new("2g"),
					},
				},
				client: api.NewMock(
					mock.New200Response(ccsTpl()),
					mock.New200Response(ccsDeploymentUpdate()),
				),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				Alias:        new(""),
				PruneOrphans: new(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{elasticsearchv2.EnrichWithEmptyTopologies(elasticsearchPayloadFromReader(t, ccsTpl(), false), &models.ElasticsearchPayload{
						Region:   new("us-east-1"),
						RefID:    new("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{},
						Plan: &models.ElasticsearchClusterPlan{
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.9.2",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: new("aws-cross-cluster-search-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID:                      "hot_content",
								ZoneCount:               1,
								InstanceConfigurationID: "aws.ccs.r5d",
								Size: &models.TopologySize{
									Resource: new("memory"),
									// This field's value is reset.
									Value: ec.Int32(1024),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   new(true),
									Ingest: new(true),
									Master: new(true),
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: new("memory"),
										Value:    ec.Int32(1024),
									},
								},
							}},
						},
					})},
					Kibana: []*models.KibanaPayload{{
						ElasticsearchClusterRefID: new("main-elasticsearch"),
						Region:                    new("us-east-1"),
						RefID:                     new("main-kibana"),
						Plan: &models.KibanaClusterPlan{
							Kibana: &models.KibanaConfiguration{},
							ClusterTopology: []*models.KibanaClusterTopologyElement{
								{
									ZoneCount:               1,
									InstanceConfigurationID: "aws.kibana.r5d",
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

		// The behavior of this change should be:
		// * Keeps all topology sizes as they were defined (saved value persists).
		{
			name: "topology change with sizes not default from explicit value to empty",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.9.2",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId:   new("main-elasticsearch"),
						HotTier: defaultHotTier,
					},
					Kibana: &kibanav2.Kibana{
						ElasticsearchClusterRefId: new("main-elasticsearch"),
						RefId:                     new("main-kibana"),
					},
					Apm: &apmv2.Apm{
						ElasticsearchClusterRefId: new("main-elasticsearch"),
						RefId:                     new("main-apm"),
					},
					EnterpriseSearch: &enterprisesearchv2.EnterpriseSearch{
						ElasticsearchClusterRefId: new("main-elasticsearch"),
						RefId:                     new("main-enterprise_search"),
					},
				},
				state: &Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.9.2",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId: new("main-elasticsearch"),
						HotTier: elasticsearchv2.CreateTierForTest(
							"hot_content",
							elasticsearchv2.ElasticsearchTopology{
								Size:        new("16g"),
								Autoscaling: &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
						CoordinatingTier: elasticsearchv2.CreateTierForTest(
							"coordinating",
							elasticsearchv2.ElasticsearchTopology{
								Size:        new("16g"),
								Autoscaling: &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
					},
					Kibana: &kibanav2.Kibana{
						ElasticsearchClusterRefId: new("main-elasticsearch"),
						RefId:                     new("main-kibana"),
						Size:                      new("2g"),
					},
					Apm: &apmv2.Apm{
						ElasticsearchClusterRefId: new("main-elasticsearch"),
						RefId:                     new("main-apm"),
						Size:                      new("1g"),
					},
					EnterpriseSearch: &enterprisesearchv2.EnterpriseSearch{
						ElasticsearchClusterRefId: new("main-elasticsearch"),
						RefId:                     new("main-enterprise_search"),
						Size:                      new("8g"),
					},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				Alias:        new(""),
				PruneOrphans: new(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{elasticsearchv2.EnrichWithEmptyTopologies(elasticsearchPayloadFromReader(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: new("us-east-1"),
						RefID:  new("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: new(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.9.2",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: new("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: new("memory"),
									Value:    ec.Int32(8192),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   new(true),
									Ingest: new(true),
									Master: new(true),
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: new("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: new("memory"),
								},
							},
							},
						},
					})},
					Kibana: []*models.KibanaPayload{{
						ElasticsearchClusterRefID: new("main-elasticsearch"),
						Region:                    new("us-east-1"),
						RefID:                     new("main-kibana"),
						Plan: &models.KibanaClusterPlan{
							Kibana: &models.KibanaConfiguration{},
							ClusterTopology: []*models.KibanaClusterTopologyElement{
								{
									ZoneCount:               1,
									InstanceConfigurationID: "aws.kibana.r5d",
									Size: &models.TopologySize{
										Resource: new("memory"),
										Value:    ec.Int32(1024),
									},
								},
							},
						},
					}},
					Apm: []*models.ApmPayload{{
						ElasticsearchClusterRefID: new("main-elasticsearch"),
						Region:                    new("us-east-1"),
						RefID:                     new("main-apm"),
						Plan: &models.ApmPlan{
							Apm: &models.ApmConfiguration{},
							ClusterTopology: []*models.ApmTopologyElement{{
								ZoneCount:               1,
								InstanceConfigurationID: "aws.apm.r5d",
								Size: &models.TopologySize{
									Resource: new("memory"),
									Value:    ec.Int32(512),
								},
							}},
						},
					}},
					EnterpriseSearch: []*models.EnterpriseSearchPayload{{
						ElasticsearchClusterRefID: new("main-elasticsearch"),
						Region:                    new("us-east-1"),
						RefID:                     new("main-enterprise_search"),
						Plan: &models.EnterpriseSearchPlan{
							EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
							ClusterTopology: []*models.EnterpriseSearchTopologyElement{
								{
									ZoneCount:               2,
									InstanceConfigurationID: "aws.enterprisesearch.m5d",
									Size: &models.TopologySize{
										Resource: new("memory"),
										Value:    ec.Int32(2048),
									},
									NodeType: &models.EnterpriseSearchNodeTypes{
										Appserver: new(true),
										Connector: new(true),
										Worker:    new(true),
									},
								},
							},
						},
					}},
				},
			},
		},

		{
			name: "does not migrate node_type to node_role on version upgrade that's lower than 7.10.0",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.11.1",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId: new("main-elasticsearch"),
						HotTier: elasticsearchv2.CreateTierForTest(
							"hot_content",
							elasticsearchv2.ElasticsearchTopology{
								Size:           new("16g"),
								NodeTypeData:   new("true"),
								NodeTypeIngest: new("true"),
								NodeTypeMaster: new("true"),
								NodeTypeMl:     new("false"),
								Autoscaling:    &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
					},
				},
				state: &Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.9.1",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId: new("main-elasticsearch"),
						HotTier: elasticsearchv2.CreateTierForTest(
							"hot_content",
							elasticsearchv2.ElasticsearchTopology{
								Size:           new("16g"),
								NodeTypeData:   new("true"),
								NodeTypeIngest: new("true"),
								NodeTypeMaster: new("true"),
								NodeTypeMl:     new("false"),
								Autoscaling:    &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
					},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				Alias:        new(""),
				PruneOrphans: new(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{elasticsearchv2.EnrichWithEmptyTopologies(elasticsearchPayloadFromReader(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: new("us-east-1"),
						RefID:  new("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: new(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.11.1",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: new("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: new("memory"),
									Value:    ec.Int32(16384),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   new(true),
									Ingest: new(true),
									Master: new(true),
									Ml:     new(false),
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: new("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: new("memory"),
								},
							}},
						},
					})},
				},
			},
		},

		{
			name: "does not migrate node_type to node_role on version upgrade that's higher than 7.10.0",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.11.1",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId: new("main-elasticsearch"),
						HotTier: elasticsearchv2.CreateTierForTest(
							"hot_content",
							elasticsearchv2.ElasticsearchTopology{
								Size:           new("16g"),
								NodeTypeData:   new("true"),
								NodeTypeIngest: new("true"),
								NodeTypeMaster: new("true"),
								NodeTypeMl:     new("false"),
							},
						),
					},
				},
				state: &Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.10.1",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId: new("main-elasticsearch"),
						HotTier: elasticsearchv2.CreateTierForTest(
							"hot_content",
							elasticsearchv2.ElasticsearchTopology{
								Size:           new("16g"),
								NodeTypeData:   new("true"),
								NodeTypeIngest: new("true"),
								NodeTypeMaster: new("true"),
								NodeTypeMl:     new("false"),
							},
						),
					},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				Alias:        new(""),
				PruneOrphans: new(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{elasticsearchv2.EnrichWithEmptyTopologies(elasticsearchPayloadFromReader(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: new("us-east-1"),
						RefID:  new("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: new(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.11.1",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: new("aws-io-optimized-v2"),
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
										Resource: new("memory"),
										Value:    ec.Int32(16384),
									},
									NodeType: &models.ElasticsearchNodeType{
										Data:   new(true),
										Ingest: new(true),
										Master: new(true),
										Ml:     new(false),
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: new("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: new("memory"),
									},
								},
							},
						},
					})},
				},
			},
		},

		{
			name: "migrates node_type to node_role when the existing topology element size is updated",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.10.1",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId: new("main-elasticsearch"),
						HotTier: elasticsearchv2.CreateTierForTest(
							"hot_content",
							elasticsearchv2.ElasticsearchTopology{
								Size:           new("32g"),
								NodeTypeData:   new("true"),
								NodeTypeIngest: new("true"),
								NodeTypeMaster: new("true"),
								NodeTypeMl:     new("false"),
								Autoscaling:    &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
					},
				},
				state: &Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.10.1",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId: new("main-elasticsearch"),
						HotTier: elasticsearchv2.CreateTierForTest(
							"hot_content",
							elasticsearchv2.ElasticsearchTopology{
								Size:           new("16g"),
								NodeTypeData:   new("true"),
								NodeTypeIngest: new("true"),
								NodeTypeMaster: new("true"),
								NodeTypeMl:     new("false"),
								Autoscaling:    &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
					},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				Alias:        new(""),
				PruneOrphans: new(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{elasticsearchv2.EnrichWithEmptyTopologies(elasticsearchPayloadFromReader(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: new("us-east-1"),
						RefID:  new("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: new(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.10.1",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: new("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: new("memory"),
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
										Resource: new("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: new("memory"),
								},
							}},
						},
					})},
				},
			},
		},

		{
			name: "migrates node_type to node_role when the existing topology element size is updated and adds warm tier",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.10.1",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId: new("main-elasticsearch"),
						HotTier: elasticsearchv2.CreateTierForTest(
							"hot_content",
							elasticsearchv2.ElasticsearchTopology{
								Size:           new("16g"),
								NodeTypeData:   new("true"),
								NodeTypeIngest: new("true"),
								NodeTypeMaster: new("true"),
								NodeTypeMl:     new("false"),
								Autoscaling:    &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
						WarmTier: elasticsearchv2.CreateTierForTest(
							"warm",
							elasticsearchv2.ElasticsearchTopology{
								Size:        new("8g"),
								Autoscaling: &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
					},
				},
				state: &Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.10.1",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId: new("main-elasticsearch"),
						HotTier: elasticsearchv2.CreateTierForTest(
							"hot_content",
							elasticsearchv2.ElasticsearchTopology{
								Size:           new("16g"),
								NodeTypeData:   new("true"),
								NodeTypeIngest: new("true"),
								NodeTypeMaster: new("true"),
								NodeTypeMl:     new("false"),
								Autoscaling:    &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
					},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				Alias:        new(""),
				PruneOrphans: new(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{elasticsearchv2.EnrichWithEmptyTopologies(elasticsearchPayloadFromReader(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: new("us-east-1"),
						RefID:  new("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: new(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.10.1",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: new("aws-io-optimized-v2"),
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
										Resource: new("memory"),
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
											Resource: new("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: new("memory"),
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
										Resource: new("memory"),
										Value:    ec.Int32(8192),
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: new("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: new("memory"),
									},
								},
							},
						},
					})},
				},
			},
		},

		{
			name: "enables autoscaling with the default policies",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.12.1",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId:     new("main-elasticsearch"),
						Autoscale: new(true),
						HotTier: elasticsearchv2.CreateTierForTest(
							"hot_content",
							elasticsearchv2.ElasticsearchTopology{
								Size:        new("16g"),
								Autoscaling: &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
						WarmTier: elasticsearchv2.CreateTierForTest(
							"warm",
							elasticsearchv2.ElasticsearchTopology{
								Size:        new("8g"),
								Autoscaling: &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
					},
				},
				state: &Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.12.1",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId:     new("main-elasticsearch"),
						Autoscale: new(true),
						HotTier: elasticsearchv2.CreateTierForTest(
							"hot_content",
							elasticsearchv2.ElasticsearchTopology{
								Size:        new("16g"),
								Autoscaling: &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
					},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				Alias:        new(""),
				PruneOrphans: new(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{elasticsearchv2.EnrichWithEmptyTopologies(elasticsearchPayloadFromReader(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: new("us-east-1"),
						RefID:  new("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: new(true),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.12.1",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: new("aws-io-optimized-v2"),
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
										Resource: new("memory"),
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
											Resource: new("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: new("memory"),
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
										Resource: new("memory"),
										Value:    ec.Int32(8192),
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: new("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: new("memory"),
									},
								},
							},
						},
					})},
				},
			},
		},

		{
			name: "parses the resources with tags",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.10.1",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId: new("main-elasticsearch"),
						HotTier: elasticsearchv2.CreateTierForTest(
							"hot_content",
							elasticsearchv2.ElasticsearchTopology{
								Size:        new("8g"),
								Autoscaling: &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
					},
					Tags: map[string]string{
						"aaa":         "bbb",
						"owner":       "elastic",
						"cost-center": "rnd",
					},
				},
				state: &Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.10.1",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId: new("main-elasticsearch"),
						HotTier: elasticsearchv2.CreateTierForTest(
							"hot_content",
							elasticsearchv2.ElasticsearchTopology{
								Size:        new("8g"),
								Autoscaling: &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
					},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				Alias:        new(""),
				PruneOrphans: new(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{Tags: []*models.MetadataItem{
					{Key: new("aaa"), Value: new("bbb")},
					{Key: new("cost-center"), Value: new("rnd")},
					{Key: new("owner"), Value: new("elastic")},
				}},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{elasticsearchv2.EnrichWithEmptyTopologies(elasticsearchPayloadFromReader(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: new("us-east-1"),
						RefID:  new("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: new(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.10.1",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: new("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: new("memory"),
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
										Resource: new("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: new("memory"),
								},
							}},
						},
					})},
				},
			},
		},

		{
			name: "handles a snapshot_source block adding Strategy: partial",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.10.1",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId: new("main-elasticsearch"),
						HotTier: elasticsearchv2.CreateTierForTest(
							"hot_content",
							elasticsearchv2.ElasticsearchTopology{
								Size:        new("8g"),
								Autoscaling: &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
						SnapshotSource: &elasticsearchv2.ElasticsearchSnapshotSource{
							SourceElasticsearchClusterId: "8c63b87af9e24ea49b8a4bfe550e5fe9",
						},
					},
				},
				state: &Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.10.1",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId: new("main-elasticsearch"),
						HotTier: elasticsearchv2.CreateTierForTest(
							"hot_content",
							elasticsearchv2.ElasticsearchTopology{
								Size:        new("8g"),
								Autoscaling: &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
					},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				Alias:        new(""),
				PruneOrphans: new(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{elasticsearchv2.EnrichWithEmptyTopologies(elasticsearchPayloadFromReader(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: new("us-east-1"),
						RefID:  new("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							Transient: &models.TransientElasticsearchPlanConfiguration{
								RestoreSnapshot: &models.RestoreSnapshotConfiguration{
									SourceClusterID: "8c63b87af9e24ea49b8a4bfe550e5fe9",
									SnapshotName:    new(""),
									Strategy:        "partial",
								},
							},
							AutoscalingEnabled: new(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.10.1",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: new("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: new("memory"),
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
										Resource: new("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: new("memory"),
								},
							}},
						},
					})},
				},
			},
		},

		{
			name: "handles empty Elasticsearch empty config block",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.10.1",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId: new("main-elasticsearch"),
						HotTier: elasticsearchv2.CreateTierForTest(
							"hot_content",
							elasticsearchv2.ElasticsearchTopology{
								Size:        new("8g"),
								Autoscaling: &elasticsearchv2.ElasticsearchTopologyAutoscaling{},
							},
						),
					},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				Alias:        new(""),
				PruneOrphans: new(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{elasticsearchv2.EnrichWithEmptyTopologies(elasticsearchPayloadFromReader(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: new("us-east-1"),
						RefID:  new("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: new(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.10.1",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: new("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: new("memory"),
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
										Resource: new("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: new("memory"),
								},
							}},
						},
					})},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := DeploymentSchema()

			var plan DeploymentTF
			diags := tfsdk.ValueFrom(context.Background(), &tt.args.plan, schema.Type(), &plan)
			assert.Nil(t, diags)

			state := tt.args.state
			if state == nil {
				state = &tt.args.plan
			}

			var stateTF DeploymentTF

			diags = tfsdk.ValueFrom(context.Background(), state, schema.Type(), &stateTF)
			assert.Nil(t, diags)

			var migrateTemplateRequest *deployments.MigrateDeploymentTemplateOK
			got, diags := plan.UpdateRequest(context.Background(), tt.args.client, stateTF, migrateTemplateRequest)
			if tt.diags != nil {
				assert.Equal(t, tt.diags, diags)
			} else {
				require.Nil(t, diags)
				require.NotNil(t, got)
				require.Equal(t, *tt.want, *got)
			}
		})
	}
}
