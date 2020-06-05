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
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/stretchr/testify/assert"
)

func Test_createResourceToModel(t *testing.T) {
	deploymentRD := newResourceData(t, resDataParams{
		ID:        mock.ValidClusterID,
		Resources: newSampleDeployment(),
	})
	type args struct {
		d *schema.ResourceData
	}
	tests := []struct {
		name string
		args args
		want *models.DeploymentCreateRequest
		err  error
	}{
		{
			name: "parses the resources",
			args: args{d: deploymentRD},
			want: &models.DeploymentCreateRequest{
				Name: "my_deployment_name",
				Resources: &models.DeploymentCreateResources{
					EnterpriseSearch: make([]*models.EnterpriseSearchPayload, 0),
					Elasticsearch: []*models.ElasticsearchPayload{
						{
							DisplayName: "some-name",
							Region:      ec.String("some-region"),
							RefID:       ec.String("main-elasticsearch"),
							Settings: &models.ElasticsearchClusterSettings{
								Monitoring: &models.ManagedMonitoringSettings{
									TargetClusterID: ec.String("some"),
								},
							},
							Plan: &models.ElasticsearchClusterPlan{
								Elasticsearch: &models.ElasticsearchConfiguration{
									Version: "7.7.0",
								},
								DeploymentTemplate: &models.DeploymentTemplateReference{
									ID: ec.String("aws-io-optimized"),
								},
								ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
									{
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
									},
								},
							},
						},
					},
					Kibana: []*models.KibanaPayload{
						{
							DisplayName:               "some-kibana-name",
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("some-region"),
							RefID:                     ec.String("main-kibana"),
							Settings:                  &models.KibanaClusterSettings{},
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{
									Version: "7.7.0",
								},
								ClusterTopology: []*models.KibanaClusterTopologyElement{
									{
										ZoneCount:               1,
										InstanceConfigurationID: "aws.kibana.r4",
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
							DisplayName:               "some-apm-name",
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("some-region"),
							RefID:                     ec.String("main-apm"),
							Settings:                  &models.ApmSettings{},
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{
									Version: "7.7.0",
								},
								ClusterTopology: []*models.ApmTopologyElement{
									{
										ZoneCount:               1,
										InstanceConfigurationID: "aws.apm.r4",
										Size: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(512),
										},
									},
								},
							},
						},
					},
					Appsearch: []*models.AppSearchPayload{
						{
							DisplayName:               "some-appsearch-name",
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("some-region"),
							RefID:                     ec.String("main-appsearch"),
							Settings:                  &models.AppSearchSettings{},
							Plan: &models.AppSearchPlan{
								Appsearch: &models.AppSearchConfiguration{
									Version: "7.7.0",
								},
								ClusterTopology: []*models.AppSearchTopologyElement{
									{
										ZoneCount:               1,
										InstanceConfigurationID: "aws.appsearch.m5",
										Size: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(2048),
										},
										NodeType: &models.AppSearchNodeTypes{
											Appserver: ec.Bool(true),
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createResourceToModel(tt.args.d)
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
	deploymentRD := newResourceData(t, resDataParams{
		ID:        mock.ValidClusterID,
		Resources: newSampleDeployment(),
	})
	type args struct {
		d *schema.ResourceData
	}
	tests := []struct {
		name string
		args args
		want *models.DeploymentUpdateRequest
		err  error
	}{
		{
			name: "parses the resources",
			args: args{d: deploymentRD},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				PruneOrphans: ec.Bool(false),
				Resources: &models.DeploymentUpdateResources{
					EnterpriseSearch: make([]*models.EnterpriseSearchPayload, 0),
					Elasticsearch: []*models.ElasticsearchPayload{
						{
							DisplayName: "some-name",
							Region:      ec.String("some-region"),
							RefID:       ec.String("main-elasticsearch"),
							Settings: &models.ElasticsearchClusterSettings{
								Monitoring: &models.ManagedMonitoringSettings{
									TargetClusterID: ec.String("some"),
								},
							},
							Plan: &models.ElasticsearchClusterPlan{
								Elasticsearch: &models.ElasticsearchConfiguration{
									Version: "7.7.0",
								},
								DeploymentTemplate: &models.DeploymentTemplateReference{
									ID: ec.String("aws-io-optimized"),
								},
								ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
									{
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
									},
								},
							},
						},
					},
					Kibana: []*models.KibanaPayload{
						{
							DisplayName:               "some-kibana-name",
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("some-region"),
							RefID:                     ec.String("main-kibana"),
							Settings:                  &models.KibanaClusterSettings{},
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{
									Version: "7.7.0",
								},
								ClusterTopology: []*models.KibanaClusterTopologyElement{
									{
										ZoneCount:               1,
										InstanceConfigurationID: "aws.kibana.r4",
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
							DisplayName:               "some-apm-name",
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("some-region"),
							RefID:                     ec.String("main-apm"),
							Settings:                  &models.ApmSettings{},
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{
									Version: "7.7.0",
								},
								ClusterTopology: []*models.ApmTopologyElement{
									{
										ZoneCount:               1,
										InstanceConfigurationID: "aws.apm.r4",
										Size: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(512),
										},
									},
								},
							},
						},
					},
					Appsearch: []*models.AppSearchPayload{
						{
							DisplayName:               "some-appsearch-name",
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("some-region"),
							RefID:                     ec.String("main-appsearch"),
							Settings:                  &models.AppSearchSettings{},
							Plan: &models.AppSearchPlan{
								Appsearch: &models.AppSearchConfiguration{
									Version: "7.7.0",
								},
								ClusterTopology: []*models.AppSearchTopologyElement{
									{
										ZoneCount:               1,
										InstanceConfigurationID: "aws.appsearch.m5",
										Size: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(2048),
										},
										NodeType: &models.AppSearchNodeTypes{
											Appserver: ec.Bool(true),
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := updateResourceToModel(tt.args.d)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
