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

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
)

func Test_modelToState(t *testing.T) {
	state := modelV0{
		ID:                   types.StringValue("test"),
		NamePrefix:           types.StringValue("test"),
		Healthy:              types.StringValue("true"),
		DeploymentTemplateID: types.StringValue("azure-compute-optimized"),
	}

	wantDeployments := modelV0{
		ID:                   types.StringValue("2705093922"),
		NamePrefix:           types.StringValue("test"),
		ReturnCount:          types.Int64Value(1),
		DeploymentTemplateID: types.StringValue("azure-compute-optimized"),
		Healthy:              types.StringValue("true"),
		Deployments: func() types.List {
			res, diags := types.ListValueFrom(
				context.Background(),
				types.ObjectType{AttrTypes: deploymentAttrTypes()},
				[]deploymentModelV0{
					{
						Name:                         types.StringValue("test-hello"),
						Alias:                        types.StringValue("dev"),
						ApmResourceID:                types.StringValue("9884c76ae1cd4521a0d9918a454a700d"),
						ApmRefID:                     types.StringValue("apm"),
						DeploymentID:                 types.StringValue("a8f22a9b9e684a7f94a89df74aa14331"),
						ElasticsearchResourceID:      types.StringValue("a98dd0dac15a48d5b3953384c7e571b9"),
						ElasticsearchRefID:           types.StringValue("elasticsearch"),
						EnterpriseSearchResourceID:   types.StringValue("f17e4d8a61b14c12b020d85b723357ba"),
						EnterpriseSearchRefID:        types.StringValue("enterprise_search"),
						KibanaResourceID:             types.StringValue("c75297d672b54da68faecededf372f87"),
						KibanaRefID:                  types.StringValue("kibana"),
						IntegrationsServerResourceID: types.StringValue("3b3025a012fd3dd5c9dcae2a1ac89c6f"),
						IntegrationsServerRefID:      types.StringValue("integrations_server"),
					},
				},
			)
			assert.Nil(t, diags)

			return res
		}(),
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
