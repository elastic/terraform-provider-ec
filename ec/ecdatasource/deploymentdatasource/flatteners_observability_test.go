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

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func TestFlattenObservability(t *testing.T) {
	type args struct {
		settings *models.DeploymentSettings
	}
	tests := []struct {
		name string
		args args
		want []observabilitySettingsModel
	}{
		{
			name: "flattens no observability settings when empty #1",
			args: args{},
		},
		{
			name: "flattens no observability settings when empty #2",
			args: args{settings: &models.DeploymentSettings{}},
		},
		{
			name: "flattens no observability settings when empty #3",
			args: args{settings: &models.DeploymentSettings{Observability: &models.DeploymentObservabilitySettings{}}},
		},
		{
			name: "flattens observability settings",
			args: args{settings: &models.DeploymentSettings{
				Observability: &models.DeploymentObservabilitySettings{
					Logging: &models.DeploymentLoggingSettings{
						Destination: &models.ObservabilityAbsoluteDeployment{
							DeploymentID: &mock.ValidClusterID,
							RefID:        "main-elasticsearch",
						},
					},
				},
			}},
			want: []observabilitySettingsModel{{
				DeploymentID: types.StringValue(mock.ValidClusterID),
				RefID:        types.StringValue("main-elasticsearch"),
				Logs:         types.BoolValue(true),
				Metrics:      types.BoolValue(false),
			}},
		},
		{
			name: "flattens observability settings",
			args: args{settings: &models.DeploymentSettings{
				Observability: &models.DeploymentObservabilitySettings{
					Metrics: &models.DeploymentMetricsSettings{
						Destination: &models.ObservabilityAbsoluteDeployment{
							DeploymentID: &mock.ValidClusterID,
							RefID:        "main-elasticsearch",
						},
					},
				},
			}},
			want: []observabilitySettingsModel{{
				DeploymentID: types.StringValue(mock.ValidClusterID),
				RefID:        types.StringValue("main-elasticsearch"),
				Logs:         types.BoolValue(false),
				Metrics:      types.BoolValue(true),
			}},
		},
		{
			name: "flattens observability settings",
			args: args{settings: &models.DeploymentSettings{
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
			}},
			want: []observabilitySettingsModel{{
				DeploymentID: types.StringValue(mock.ValidClusterID),
				RefID:        types.StringValue("main-elasticsearch"),
				Logs:         types.BoolValue(true),
				Metrics:      types.BoolValue(true),
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observability, diags := flattenObservability(context.Background(), tt.args.settings)
			assert.Empty(t, diags)
			var got []observabilitySettingsModel
			observability.ElementsAs(context.Background(), &got, false)
			assert.Equal(t, tt.want, got)
			util.CheckConverionToAttrValue(t, &DataSource{}, "observability", observability)
		})
	}
}
