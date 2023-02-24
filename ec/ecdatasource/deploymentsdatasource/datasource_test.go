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
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
)

func Test_modelToState(t *testing.T) {
	state := modelV0{
		ID:                   types.String{Value: "test"},
		NamePrefix:           types.String{Value: "test"},
		Healthy:              types.String{Value: "true"},
		DeploymentTemplateID: types.String{Value: "azure-compute-optimized"},
	}

	wantDeployments := modelV0{
		ID:                   types.String{Value: "2705093922"},
		NamePrefix:           types.String{Value: "test"},
		ReturnCount:          types.Int64{Value: 1},
		DeploymentTemplateID: types.String{Value: "azure-compute-optimized"},
		Healthy:              types.String{Value: "true"},
		Deployments: types.List{
			ElemType: types.ObjectType{AttrTypes: deploymentAttrTypes()},
			Elems: []attr.Value{types.Object{
				AttrTypes: deploymentAttrTypes(),
				Attrs: map[string]attr.Value{
					"name":                            types.String{Value: "test-hello"},
					"alias":                           types.String{Value: "dev"},
					"apm_resource_id":                 types.String{Value: "9884c76ae1cd4521a0d9918a454a700d"},
					"apm_ref_id":                      types.String{Value: "apm"},
					"deployment_id":                   types.String{Value: "a8f22a9b9e684a7f94a89df74aa14331"},
					"elasticsearch_resource_id":       types.String{Value: "a98dd0dac15a48d5b3953384c7e571b9"},
					"elasticsearch_ref_id":            types.String{Value: "elasticsearch"},
					"enterprise_search_resource_id":   types.String{Value: "f17e4d8a61b14c12b020d85b723357ba"},
					"enterprise_search_ref_id":        types.String{Value: "enterprise_search"},
					"kibana_resource_id":              types.String{Value: "c75297d672b54da68faecededf372f87"},
					"kibana_ref_id":                   types.String{Value: "kibana"},
					"integrations_server_resource_id": types.String{Value: "3b3025a012fd3dd5c9dcae2a1ac89c6f"},
					"integrations_server_ref_id":      types.String{Value: "integrations_server"},
				},
			}},
		},
	}

	searchResponse := &models.DeploymentsSearchResponse{
		ReturnCount: ec.Int32(1),
		Deployments: []*models.DeploymentSearchResponse{
			{
				Healthy: ec.Bool(true),
				ID:      ec.String("a8f22a9b9e684a7f94a89df74aa14331"),
				Name:    ec.String("test-hello"),
				Alias:   "dev",
				Resources: &models.DeploymentResources{
					Elasticsearch: []*models.ElasticsearchResourceInfo{
						{
							RefID: ec.String("elasticsearch"),
							ID:    ec.String("a98dd0dac15a48d5b3953384c7e571b9"),
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
							ID:    ec.String("c75297d672b54da68faecededf372f87"),
							RefID: ec.String("kibana"),
						},
					},
					Apm: []*models.ApmResourceInfo{
						{
							ID:    ec.String("9884c76ae1cd4521a0d9918a454a700d"),
							RefID: ec.String("apm"),
						},
					},
					EnterpriseSearch: []*models.EnterpriseSearchResourceInfo{
						{
							ID:    ec.String("f17e4d8a61b14c12b020d85b723357ba"),
							RefID: ec.String("enterprise_search"),
						},
					},
					IntegrationsServer: []*models.IntegrationsServerResourceInfo{
						{
							ID:    ec.String("3b3025a012fd3dd5c9dcae2a1ac89c6f"),
							RefID: ec.String("integrations_server"),
						},
					},
				},
			},
		},
	}

	type args struct {
		state modelV0
		res   *models.DeploymentsSearchResponse
	}
	tests := []struct {
		name  string
		args  args
		want  modelV0
		diags error
	}{
		{
			name: "flattens deployment resources",
			want: wantDeployments,
			args: args{
				state: state,
				res:   searchResponse,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state = tt.args.state
			diags := modelToState(context.Background(), tt.args.res, &state)
			if tt.diags != nil {
				assert.Equal(t, tt.diags, diags)
			} else {
				assert.Empty(t, diags)
			}

			assert.Equal(t, tt.want, state)
		})
	}
}
