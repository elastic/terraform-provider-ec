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

package utils

import (
	"errors"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/stretchr/testify/assert"
)

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
			got, err := GetDeploymentTemplateID(tt.args.res)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
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
			if got := HasRunningResources(tt.args.res); got != tt.want {
				t.Errorf("hasRunningResources() = %v, want %v", got, tt.want)
			}
		})
	}
}
