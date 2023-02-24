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

package v2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
)

func Test_observabilityPayload(t *testing.T) {
	type args struct {
		observability *Observability
		*api.API
	}
	tests := []struct {
		name string
		args args
		want *models.DeploymentObservabilitySettings
	}{
		{
			name: "empty returns an empty request",
			args: args{},
		},
		{
			name: "expands all observability settings with given refID",
			args: args{
				observability: &Observability{
					DeploymentId: &mock.ValidClusterID,
					RefId:        ec.String("main-elasticsearch"),
					Metrics:      true,
					Logs:         true,
				},
			},
			want: &models.DeploymentObservabilitySettings{
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
		{
			name: "expands all observability settings",
			args: args{
				API: api.NewMock(
					mock.New200Response(
						mock.NewStructBody(models.DeploymentGetResponse{
							Healthy: ec.Bool(true),
							ID:      ec.String(mock.ValidClusterID),
							Resources: &models.DeploymentResources{
								Elasticsearch: []*models.ElasticsearchResourceInfo{{
									ID:    ec.String(mock.ValidClusterID),
									RefID: ec.String("main-elasticsearch"),
								}},
							},
						}),
					),
				),
				observability: &Observability{
					DeploymentId: &mock.ValidClusterID,
					Metrics:      true,
					Logs:         true,
				},
			},
			want: &models.DeploymentObservabilitySettings{
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
		{
			name: "expands logging observability settings",
			args: args{
				API: api.NewMock(
					mock.New200Response(
						mock.NewStructBody(models.DeploymentGetResponse{
							Healthy: ec.Bool(true),
							ID:      ec.String(mock.ValidClusterID),
							Resources: &models.DeploymentResources{
								Elasticsearch: []*models.ElasticsearchResourceInfo{{
									ID:    ec.String(mock.ValidClusterID),
									RefID: ec.String("main-elasticsearch"),
								}},
							},
						}),
					),
				),
				observability: &Observability{
					DeploymentId: &mock.ValidClusterID,
					Metrics:      false,
					Logs:         true,
				},
			},
			want: &models.DeploymentObservabilitySettings{
				Logging: &models.DeploymentLoggingSettings{
					Destination: &models.ObservabilityAbsoluteDeployment{
						DeploymentID: &mock.ValidClusterID,
						RefID:        "main-elasticsearch",
					},
				},
			},
		},
		{
			name: "expands metrics observability settings",
			args: args{
				API: api.NewMock(
					mock.New200Response(
						mock.NewStructBody(models.DeploymentGetResponse{
							Healthy: ec.Bool(true),
							ID:      ec.String(mock.ValidClusterID),
							Resources: &models.DeploymentResources{
								Elasticsearch: []*models.ElasticsearchResourceInfo{{
									ID:    ec.String(mock.ValidClusterID),
									RefID: ec.String("main-elasticsearch"),
								}},
							},
						}),
					),
				),
				observability: &Observability{
					DeploymentId: &mock.ValidClusterID,
					Metrics:      true,
					Logs:         false,
				},
			},
			want: &models.DeploymentObservabilitySettings{
				Metrics: &models.DeploymentMetricsSettings{
					Destination: &models.ObservabilityAbsoluteDeployment{
						DeploymentID: &mock.ValidClusterID,
						RefID:        "main-elasticsearch",
					},
				},
			},
		},
		{
			name: "observability targeting self without ref-id",
			args: args{
				API: api.NewMock(
					mock.New200Response(
						mock.NewStructBody(models.DeploymentGetResponse{
							Healthy: ec.Bool(true),
							ID:      ec.String(mock.ValidClusterID),
							Resources: &models.DeploymentResources{
								Elasticsearch: []*models.ElasticsearchResourceInfo{{
									ID:    ec.String(mock.ValidClusterID),
									RefID: ec.String("main-elasticsearch"),
								}},
							},
						}),
					),
				),
				observability: &Observability{
					DeploymentId: ec.String("self"),
					Metrics:      true,
					Logs:         false,
				},
			},
			want: &models.DeploymentObservabilitySettings{
				Metrics: &models.DeploymentMetricsSettings{
					Destination: &models.ObservabilityAbsoluteDeployment{
						DeploymentID: ec.String("self"),
						RefID:        "",
					},
				},
			},
		},
		{
			name: "observability targeting self with ref-id",
			args: args{
				API: api.NewMock(
					mock.New200Response(
						mock.NewStructBody(models.DeploymentGetResponse{
							Healthy: ec.Bool(true),
							ID:      ec.String(mock.ValidClusterID),
							Resources: &models.DeploymentResources{
								Elasticsearch: []*models.ElasticsearchResourceInfo{{
									ID:    ec.String(mock.ValidClusterID),
									RefID: ec.String("main-elasticsearch"),
								}},
							},
						}),
					),
				),
				observability: &Observability{
					DeploymentId: ec.String("self"),
					RefId:        ec.String("main-elasticsearch"),
					Metrics:      true,
					Logs:         false,
				},
			},
			want: &models.DeploymentObservabilitySettings{
				Metrics: &models.DeploymentMetricsSettings{
					Destination: &models.ObservabilityAbsoluteDeployment{
						DeploymentID: ec.String("self"),
						RefID:        "main-elasticsearch",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var observability types.Object
			diags := tfsdk.ValueFrom(context.Background(), tt.args.observability, ObservabilitySchema().FrameworkType(), &observability)
			assert.Nil(t, diags)

			got, diags := ObservabilityPayload(context.Background(), observability, tt.args.API)
			assert.Nil(t, diags)
			assert.Equal(t, tt.want, got)
		})
	}
}
