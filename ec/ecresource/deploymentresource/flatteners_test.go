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
	_ = deploymentSchemaArg.Set("region", "us-east-1")

	wantDeployment := util.NewResourceData(t, util.ResDataParams{
		ID:        mock.ValidClusterID,
		Resources: newSampleDeployment(),
		Schema:    newSchema(),
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
