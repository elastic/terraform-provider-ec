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
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/stretchr/testify/assert"
)

func TestFlattenObservability(t *testing.T) {
	type args struct {
		settings *models.DeploymentSettings
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "flattens no observability settings when empty",
			args: args{},
		},
		{
			name: "flattens no observability settings when empty",
			args: args{settings: &models.DeploymentSettings{}},
		},
		{
			name: "flattens no observability settings when empty",
			args: args{settings: &models.DeploymentSettings{Observability: &models.DeploymentObservabilitySettings{}}},
		},
		{
			name: "flattens observability settings",
			args: args{settings: &models.DeploymentSettings{
				Observability: &models.DeploymentObservabilitySettings{
					Logging: &models.DeploymentLoggingSettings{
						Destination: &models.AbsoluteRefID{
							DeploymentID: &mock.ValidClusterID,
							RefID:        ec.String("main-elasticsearch"),
						},
					},
				},
			}},
			want: []interface{}{map[string]interface{}{
				"deployment_id": &mock.ValidClusterID,
				"ref_id":        ec.String("main-elasticsearch"),
				"logs":          true,
			}},
		},
		{
			name: "flattens observability settings",
			args: args{settings: &models.DeploymentSettings{
				Observability: &models.DeploymentObservabilitySettings{
					Metrics: &models.DeploymentMetricsSettings{
						Destination: &models.AbsoluteRefID{
							DeploymentID: &mock.ValidClusterID,
							RefID:        ec.String("main-elasticsearch"),
						},
					},
				},
			}},
			want: []interface{}{map[string]interface{}{
				"deployment_id": &mock.ValidClusterID,
				"ref_id":        ec.String("main-elasticsearch"),
				"metrics":       true,
			}},
		},
		{
			name: "flattens observability settings",
			args: args{settings: &models.DeploymentSettings{
				Observability: &models.DeploymentObservabilitySettings{
					Logging: &models.DeploymentLoggingSettings{
						Destination: &models.AbsoluteRefID{
							DeploymentID: &mock.ValidClusterID,
							RefID:        ec.String("main-elasticsearch"),
						},
					},
					Metrics: &models.DeploymentMetricsSettings{
						Destination: &models.AbsoluteRefID{
							DeploymentID: &mock.ValidClusterID,
							RefID:        ec.String("main-elasticsearch"),
						},
					},
				},
			}},
			want: []interface{}{map[string]interface{}{
				"deployment_id": &mock.ValidClusterID,
				"ref_id":        ec.String("main-elasticsearch"),
				"logs":          true,
				"metrics":       true,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := flattenObservability(tt.args.settings)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExpandObservability(t *testing.T) {
	type args struct {
		v []interface{}
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
				v: []interface{}{map[string]interface{}{
					"deployment_id": mock.ValidClusterID,
					"ref_id":        "main-elasticsearch",
					"metrics":       true,
					"logs":          true,
				}},
			},
			want: &models.DeploymentObservabilitySettings{
				Logging: &models.DeploymentLoggingSettings{
					Destination: &models.AbsoluteRefID{
						DeploymentID: &mock.ValidClusterID,
						RefID:        ec.String("main-elasticsearch"),
					},
				},
				Metrics: &models.DeploymentMetricsSettings{
					Destination: &models.AbsoluteRefID{
						DeploymentID: &mock.ValidClusterID,
						RefID:        ec.String("main-elasticsearch"),
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
				v: []interface{}{map[string]interface{}{
					"deployment_id": mock.ValidClusterID,
					"metrics":       true,
					"logs":          true,
				}},
			},
			want: &models.DeploymentObservabilitySettings{
				Logging: &models.DeploymentLoggingSettings{
					Destination: &models.AbsoluteRefID{
						DeploymentID: &mock.ValidClusterID,
						RefID:        ec.String("main-elasticsearch"),
					},
				},
				Metrics: &models.DeploymentMetricsSettings{
					Destination: &models.AbsoluteRefID{
						DeploymentID: &mock.ValidClusterID,
						RefID:        ec.String("main-elasticsearch"),
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
				v: []interface{}{map[string]interface{}{
					"deployment_id": mock.ValidClusterID,
					"metrics":       false,
					"logs":          true,
				}},
			},
			want: &models.DeploymentObservabilitySettings{
				Logging: &models.DeploymentLoggingSettings{
					Destination: &models.AbsoluteRefID{
						DeploymentID: &mock.ValidClusterID,
						RefID:        ec.String("main-elasticsearch"),
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
				v: []interface{}{map[string]interface{}{
					"deployment_id": mock.ValidClusterID,
					"metrics":       true,
					"logs":          false,
				}},
			},
			want: &models.DeploymentObservabilitySettings{
				Metrics: &models.DeploymentMetricsSettings{
					Destination: &models.AbsoluteRefID{
						DeploymentID: &mock.ValidClusterID,
						RefID:        ec.String("main-elasticsearch"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := expandObservability(tt.args.v, tt.args.API)
			assert.Equal(t, tt.want, got)
		})
	}
}
