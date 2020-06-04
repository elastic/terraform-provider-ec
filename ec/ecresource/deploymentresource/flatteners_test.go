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
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/stretchr/testify/assert"
)

func Test_modelToState(t *testing.T) {
	deploymentSchemaArg := schema.TestResourceDataRaw(t, NewSchema(), nil)
	deploymentSchemaArg.SetId(mock.ValidClusterID)

	wantDeployment := newResourceData(t, resDataParams{
		ID:        mock.ValidClusterID,
		Resources: newSampleDeployment(),
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
					Resources: &models.DeploymentResources{
						Elasticsearch: []*models.ElasticsearchResourceInfo{
							{
								Region: ec.String("some-region"),
								RefID:  ec.String("main-elasticsearch"),
								Info: &models.ElasticsearchClusterInfo{
									ClusterID:   &mock.ValidClusterID,
									ClusterName: ec.String("some-name"),
									Region:      "some-region",
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
								},
							},
						},
						Kibana: []*models.KibanaResourceInfo{
							{
								Region:                    ec.String("some-region"),
								RefID:                     ec.String("main-kibana"),
								ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
								Info: &models.KibanaClusterInfo{
									ClusterID:   &mock.ValidClusterID,
									ClusterName: ec.String("some-kibana-name"),
									Region:      "some-region",
									PlanInfo: &models.KibanaClusterPlansInfo{
										Current: &models.KibanaClusterPlanInfo{
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
								},
							},
						},
						Apm: []*models.ApmResourceInfo{
							{
								Region:                    ec.String("some-region"),
								RefID:                     ec.String("main-apm"),
								ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
								Info: &models.ApmInfo{
									ID:     &mock.ValidClusterID,
									Name:   ec.String("some-apm-name"),
									Region: "some-region",
									PlanInfo: &models.ApmPlansInfo{
										Current: &models.ApmPlanInfo{
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
								},
							},
						},
						Appsearch: []*models.AppSearchResourceInfo{
							{
								Region:                    ec.String("some-region"),
								RefID:                     ec.String("main-appsearch"),
								ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
								Info: &models.AppSearchInfo{
									ID:     &mock.ValidClusterID,
									Name:   ec.String("some-appsearch-name"),
									Region: "some-region",
									PlanInfo: &models.AppSearchPlansInfo{
										Current: &models.AppSearchPlanInfo{
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
					},
				},
			},
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

			assert.Equal(t, tt.want.State().Attributes, tt.args.d.State().Attributes)
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
	deploymentRD := newResourceData(t, resDataParams{
		ID:        mock.ValidClusterID,
		Resources: newSampleDeployment(),
	})

	rawData := newSampleDeployment()
	esData := rawData["elasticsearch"].([]interface{})[0].(map[string]interface{})
	esData["username"] = "my-username"
	esData["password"] = "my-password"
	apmData := rawData["apm"].([]interface{})[0].(map[string]interface{})
	apmData["secret_token"] = "my-secret-token"

	wantDeploymentRD := newResourceData(t, resDataParams{
		ID:        mock.ValidClusterID,
		Resources: rawData,
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
				resources: []*models.DeploymentResource{
					{
						Credentials: &models.ClusterCredentials{
							Username: ec.String("my-username"),
							Password: ec.String("my-password"),
						},
					},
					{
						SecretToken: "my-secret-token",
					},
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
