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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_modelToState(t *testing.T) {
	wantDeployment := newSampleDeployment(t)
	type args struct {
		res *models.DeploymentGetResponse
	}
	tests := []struct {
		name string
		args args
		want modelV0
		err  error
	}{
		{
			name: "flattens deployment resources",
			want: wantDeployment,
			args: args{
				res: &models.DeploymentGetResponse{
					Alias:   "some-alias",
					ID:      &mock.ValidClusterID,
					Healthy: ec.Bool(true),
					Name:    ec.String("my_deployment_name"),
					Settings: &models.DeploymentSettings{
						TrafficFilterSettings: &models.TrafficFilterSettings{
							Rulesets: []string{"0.0.0.0/0", "192.168.10.0/24"},
						},
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
						IntegrationsServer: []*models.IntegrationsServerResourceInfo{{
							Info: &models.IntegrationsServerInfo{
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
					Metadata: &models.DeploymentMetadata{
						Tags: []*models.MetadataItem{
							{
								Key:   ec.String("foo"),
								Value: ec.String("bar"),
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := modelV0{
				ID: types.StringValue(mock.ValidClusterID),
			}
			diags := modelToState(context.Background(), tt.args.res, &model)
			if tt.err != nil {
				assert.Equal(t, diags, tt.err)
			} else {
				assert.Empty(t, diags)
			}

			assert.Equal(t, tt.want, model)
		})
	}
}

func newSampleDeployment(t *testing.T) modelV0 {
	return modelV0{
		ID:                   types.StringValue(mock.ValidClusterID),
		Name:                 types.StringValue("my_deployment_name"),
		Alias:                types.StringValue("some-alias"),
		DeploymentTemplateID: types.StringValue("aws-io-optimized"),
		Healthy:              types.BoolValue(true),
		Region:               types.StringValue("us-east-1"),
		TrafficFilter:        util.StringListAsType(t, []string{"0.0.0.0/0", "192.168.10.0/24"}),
		Observability: func() types.List {
			res, diags := types.ListValueFrom(
				context.Background(),
				types.ObjectType{AttrTypes: observabilitySettingsAttrTypes()},
				[]observabilitySettingsModel{
					{
						DeploymentID: types.StringValue(mock.ValidClusterID),
						RefID:        types.StringValue("main-elasticsearch"),
						Logs:         types.BoolValue(true),
						Metrics:      types.BoolValue(true),
					},
				},
			)
			assert.Nil(t, diags)

			return res
		}(),
		Elasticsearch: func() types.List {
			topology, diags := types.ListValue(
				types.ObjectType{AttrTypes: elasticsearchTopologyAttrTypes()},
				[]attr.Value{},
			)
			assert.Nil(t, diags)

			res, diags := types.ListValueFrom(
				context.Background(),
				types.ObjectType{AttrTypes: elasticsearchResourceInfoAttrTypes()},
				[]elasticsearchResourceInfoModelV0{
					{
						Healthy:  types.BoolValue(true),
						Topology: topology,
					},
				},
			)
			assert.Nil(t, diags)

			return res
		}(),
		Kibana: func() types.List {
			res, diags := types.ListValueFrom(
				context.Background(),
				types.ObjectType{AttrTypes: kibanaResourceInfoAttrTypes()},
				[]kibanaResourceInfoModelV0{
					{
						Healthy: types.BoolValue(true),
						Topology: types.ListNull(
							types.ObjectType{AttrTypes: kibanaTopologyAttrTypes()},
						),
					},
				},
			)
			assert.Nil(t, diags)

			return res
		}(),
		Apm: func() types.List {
			res, diags := types.ListValueFrom(
				context.Background(),
				types.ObjectType{AttrTypes: apmResourceInfoAttrTypes()},
				[]apmResourceInfoModelV0{
					{
						Healthy: types.BoolValue(true),
						Topology: types.ListNull(
							types.ObjectType{AttrTypes: apmTopologyAttrTypes()},
						),
					},
				},
			)
			assert.Nil(t, diags)

			return res
		}(),
		IntegrationsServer: func() types.List {
			res, diags := types.ListValueFrom(
				context.Background(),
				types.ObjectType{AttrTypes: integrationsServerResourceInfoAttrTypes()},
				[]integrationsServerResourceInfoModelV0{
					{
						Healthy: types.BoolValue(true),
						Topology: types.ListNull(
							types.ObjectType{AttrTypes: integrationsServerTopologyAttrTypes()},
						),
					},
				},
			)
			assert.Nil(t, diags)

			return res
		}(),
		EnterpriseSearch: func() types.List {
			res, diags := types.ListValueFrom(
				context.Background(),
				types.ObjectType{AttrTypes: enterpriseSearchResourceInfoAttrTypes()},
				[]enterpriseSearchResourceInfoModelV0{
					{
						Healthy: types.BoolValue(true),
						Topology: types.ListNull(
							types.ObjectType{AttrTypes: enterpriseSearchTopologyAttrTypes()},
						),
					},
				},
			)
			assert.Nil(t, diags)

			return res
		}(),
		Tags: util.StringMapAsType(t, map[string]string{"foo": "bar"}),
	}
}
