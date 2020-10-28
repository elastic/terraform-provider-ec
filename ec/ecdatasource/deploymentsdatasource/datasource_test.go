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

package deploymentsdatasource

import (
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_modelToState(t *testing.T) {
	deploymentsSchemaArg := schema.TestResourceDataRaw(t, newSchema(), nil)
	deploymentsSchemaArg.SetId("myID")
	_ = deploymentsSchemaArg.Set("name_prefix", "test")
	_ = deploymentsSchemaArg.Set("healthy", "true")
	_ = deploymentsSchemaArg.Set("deployment_template_id", "azure-compute-optimized")

	wantDeployments := util.NewResourceData(t, util.ResDataParams{
		ID:     "myID",
		State:  newSampleDeployments(),
		Schema: newSchema(),
	})

	type args struct {
		d   *schema.ResourceData
		res *models.DeploymentsSearchResponse
	}
	tests := []struct {
		name string
		args args
		want *schema.ResourceData
		err  error
	}{
		{
			name: "flattens deployment resources",
			want: wantDeployments,
			args: args{
				d: deploymentsSchemaArg,
				res: &models.DeploymentsSearchResponse{
					ReturnCount: ec.Int32(1),
					Deployments: []*models.DeploymentSearchResponse{
						{
							Healthy: ec.Bool(true),
							ID:      ec.String("a8f22a9b9e684a7f94a89df74aa14331"),
							Name:    ec.String("test-hello"),
							Resources: &models.DeploymentResources{
								Elasticsearch: []*models.ElasticsearchResourceInfo{
									{
										ID: ec.String("a98dd0dac15a48d5b3953384c7e571b9"),
										Info: &models.ElasticsearchClusterInfo{
											Healthy: ec.Bool(true),
											PlanInfo: &models.ElasticsearchClusterPlansInfo{
												Current: &models.ElasticsearchClusterPlanInfo{
													Plan: &models.ElasticsearchClusterPlan{
														DeploymentTemplate: &models.DeploymentTemplateReference{
															ID: ec.String("azure-compute-optimized"),
														},
													},
												},
											},
										},
									},
								},
								Kibana: []*models.KibanaResourceInfo{
									{
										ID: ec.String("c75297d672b54da68faecededf372f87"),
									},
								},
								Apm: []*models.ApmResourceInfo{
									{
										ID: ec.String("9884c76ae1cd4521a0d9918a454a700d"),
									},
								},
								EnterpriseSearch: []*models.EnterpriseSearchResourceInfo{
									{
										ID: ec.String("f17e4d8a61b14c12b020d85b723357ba"),
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
			if tt.err != nil || err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want.State().Attributes, tt.args.d.State().Attributes)
		})
	}
}

func newSampleDeployments() map[string]interface{} {
	return map[string]interface{}{
		"id":                     "myID",
		"name_prefix":            "test",
		"return_count":           1,
		"deployment_template_id": "azure-compute-optimized",
		"healthy":                "true",
		"deployments": []interface{}{map[string]interface{}{
			"apm_resource_id":               "9884c76ae1cd4521a0d9918a454a700d",
			"deployment_id":                 "a8f22a9b9e684a7f94a89df74aa14331",
			"elasticsearch_resource_id":     "a98dd0dac15a48d5b3953384c7e571b9",
			"enterprise_search_resource_id": "f17e4d8a61b14c12b020d85b723357ba",
			"kibana_resource_id":            "c75297d672b54da68faecededf372f87",
		}},
	}
}
