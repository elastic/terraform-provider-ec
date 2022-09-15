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
	wantDeployment := newSampleDeployment()
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
				ID: types.String{Value: mock.ValidClusterID},
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

func newSampleDeployment() modelV0 {
	return modelV0{
		ID:                   types.String{Value: mock.ValidClusterID},
		Name:                 types.String{Value: "my_deployment_name"},
		Alias:                types.String{Value: "some-alias"},
		DeploymentTemplateID: types.String{Value: "aws-io-optimized"},
		Healthy:              types.Bool{Value: true},
		Region:               types.String{Value: "us-east-1"},
		TrafficFilter:        util.StringListAsType([]string{"0.0.0.0/0", "192.168.10.0/24"}),
		Observability: types.List{
			ElemType: types.ObjectType{AttrTypes: observabilitySettingsAttrTypes()},
			Elems: []attr.Value{
				types.Object{
					AttrTypes: observabilitySettingsAttrTypes(),
					Attrs: map[string]attr.Value{
						"deployment_id": types.String{Value: mock.ValidClusterID},
						"ref_id":        types.String{Value: "main-elasticsearch"},
						"logs":          types.Bool{Value: true},
						"metrics":       types.Bool{Value: true},
					},
				},
			},
		},
		Elasticsearch: types.List{
			ElemType: types.ObjectType{AttrTypes: elasticsearchResourceInfoAttrTypes()},
			Elems: []attr.Value{
				types.Object{
					AttrTypes: elasticsearchResourceInfoAttrTypes(),
					Attrs: map[string]attr.Value{
						"cloud_id":       types.String{Value: ""},
						"healthy":        types.Bool{Value: true},
						"autoscale":      types.String{Value: ""},
						"http_endpoint":  types.String{Value: ""},
						"https_endpoint": types.String{Value: ""},
						"ref_id":         types.String{Value: ""},
						"resource_id":    types.String{Value: ""},
						"status":         types.String{Value: ""},
						"version":        types.String{Value: ""},
						"topology": types.List{
							ElemType: types.ObjectType{AttrTypes: elasticsearchTopologyAttrTypes()},
							Elems:    []attr.Value{},
						},
					},
				},
			},
		},
		Kibana: types.List{
			ElemType: types.ObjectType{AttrTypes: kibanaResourceInfoAttrTypes()},
			Elems: []attr.Value{
				types.Object{
					AttrTypes: kibanaResourceInfoAttrTypes(),
					Attrs: map[string]attr.Value{
						"elasticsearch_cluster_ref_id": types.String{Value: ""},
						"healthy":                      types.Bool{Value: true},
						"http_endpoint":                types.String{Value: ""},
						"https_endpoint":               types.String{Value: ""},
						"ref_id":                       types.String{Value: ""},
						"resource_id":                  types.String{Value: ""},
						"status":                       types.String{Value: ""},
						"version":                      types.String{Value: ""},
						"topology": types.List{
							ElemType: types.ObjectType{AttrTypes: kibanaTopologyAttrTypes()},
							Elems:    []attr.Value{},
						},
					},
				},
			},
		},
		Apm: types.List{
			ElemType: types.ObjectType{AttrTypes: apmResourceInfoAttrTypes()},
			Elems: []attr.Value{
				types.Object{
					AttrTypes: apmResourceInfoAttrTypes(),
					Attrs: map[string]attr.Value{
						"elasticsearch_cluster_ref_id": types.String{Value: ""},
						"healthy":                      types.Bool{Value: true},
						"http_endpoint":                types.String{Value: ""},
						"https_endpoint":               types.String{Value: ""},
						"ref_id":                       types.String{Value: ""},
						"resource_id":                  types.String{Value: ""},
						"status":                       types.String{Value: ""},
						"version":                      types.String{Value: ""},
						"topology": types.List{
							ElemType: types.ObjectType{AttrTypes: apmTopologyAttrTypes()},
							Elems:    []attr.Value{},
						},
					},
				},
			},
		},
		IntegrationsServer: types.List{
			ElemType: types.ObjectType{AttrTypes: integrationsServerResourceInfoAttrTypes()},
			Elems: []attr.Value{
				types.Object{
					AttrTypes: integrationsServerResourceInfoAttrTypes(),
					Attrs: map[string]attr.Value{
						"elasticsearch_cluster_ref_id": types.String{Value: ""},
						"healthy":                      types.Bool{Value: true},
						"http_endpoint":                types.String{Value: ""},
						"https_endpoint":               types.String{Value: ""},
						"ref_id":                       types.String{Value: ""},
						"resource_id":                  types.String{Value: ""},
						"status":                       types.String{Value: ""},
						"version":                      types.String{Value: ""},
						"topology": types.List{
							ElemType: types.ObjectType{AttrTypes: integrationsServerTopologyAttrTypes()},
							Elems:    []attr.Value{},
						},
					},
				},
			},
		},
		EnterpriseSearch: types.List{
			ElemType: types.ObjectType{AttrTypes: enterpriseSearchResourceInfoAttrTypes()},
			Elems: []attr.Value{
				types.Object{
					AttrTypes: enterpriseSearchResourceInfoAttrTypes(),
					Attrs: map[string]attr.Value{
						"elasticsearch_cluster_ref_id": types.String{Value: ""},
						"healthy":                      types.Bool{Value: true},
						"http_endpoint":                types.String{Value: ""},
						"https_endpoint":               types.String{Value: ""},
						"ref_id":                       types.String{Value: ""},
						"resource_id":                  types.String{Value: ""},
						"status":                       types.String{Value: ""},
						"version":                      types.String{Value: ""},
						"topology": types.List{
							ElemType: types.ObjectType{AttrTypes: enterpriseSearchTopologyAttrTypes()},
							Elems:    []attr.Value{},
						},
					},
				},
			},
		},
		Tags: util.StringMapAsType(map[string]string{"foo": "bar"}),
	}
}
