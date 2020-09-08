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

package deploymentdatasource

import (
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func Test_modelToState(t *testing.T) {
	deploymentSchemaArg := schema.TestResourceDataRaw(t, newSchema(), nil)
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
					ID:      &mock.ValidClusterID,
					Healthy: ec.Bool(true),
					Name:    ec.String("my_deployment_name"),
					Resources: &models.DeploymentResources{
						Elasticsearch: []*models.ElasticsearchResourceInfo{
							{
								Region: ec.String("us-east-1"),
								Info: &models.ElasticsearchClusterInfo{
									Healthy: ec.Bool(true),
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
						Kibana: []*models.KibanaResourceInfo{
							{
								Info: &models.KibanaClusterInfo{
									Healthy: ec.Bool(true),
								},
							},
						},
						Apm: []*models.ApmResourceInfo{{
							Info: &models.ApmInfo{
								Healthy: ec.Bool(true),
							},
						}},
						EnterpriseSearch: []*models.EnterpriseSearchResourceInfo{
							{
								Info: &models.EnterpriseSearchInfo{
									Healthy: ec.Bool(true),
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

type resDataParams struct {
	Resources map[string]interface{}
	ID        string
}

func newResourceData(t *testing.T, params resDataParams) *schema.ResourceData {
	raw := schema.TestResourceDataRaw(t, DataSource().Schema, params.Resources)
	raw.SetId(params.ID)

	return raw
}

func newSampleDeployment() map[string]interface{} {
	return map[string]interface{}{
		"id":                     mock.ValidClusterID,
		"name":                   "my_deployment_name",
		"deployment_template_id": "aws-io-optimized",
		"healthy":                true,
		"region":                 "us-east-1",
		"elasticsearch": []interface{}{map[string]interface{}{
			"healthy": true,
		}},
		"kibana": []interface{}{map[string]interface{}{
			"healthy": true,
		}},
		"apm": []interface{}{map[string]interface{}{
			"healthy": true,
		}},
		"enterprise_search": []interface{}{map[string]interface{}{
			"healthy": true,
		}},
	}
}
