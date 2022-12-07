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

	"github.com/elastic/cloud-sdk-go/pkg/models"
)

func TestParseTrafficFiltering(t *testing.T) {
	type args struct {
		settings *models.DeploymentSettings
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "parses no rules when they're empty",
			args: args{},
		},
		{
			name: "parses no rules when they're empty",
			args: args{settings: &models.DeploymentSettings{}},
		},
		{
			name: "parses no rules when they're empty",
			args: args{settings: &models.DeploymentSettings{
				TrafficFilterSettings: &models.TrafficFilterSettings{},
			}},
		},
		{
			name: "parses no rules when they're empty",
			args: args{settings: &models.DeploymentSettings{
				TrafficFilterSettings: &models.TrafficFilterSettings{
					Rulesets: []string{},
				},
			}},
		},
		{
			name: "parses no rules when they're empty",
			args: args{settings: &models.DeploymentSettings{
				TrafficFilterSettings: &models.TrafficFilterSettings{
					Rulesets: []string{
						"one-id-of-a-rule",
						"another-id-of-another-rule",
					},
				},
			}},
			want: []string{
				"one-id-of-a-rule",
				"another-id-of-another-rule",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadTrafficFilters(tt.args.settings)
			assert.Nil(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_trafficFilterToModel(t *testing.T) {
	type args struct {
		filters []string
		req     *models.DeploymentCreateRequest
	}
	tests := []struct {
		name string
		args args
		want *models.DeploymentCreateRequest
	}{
		{
			name: "empty returns an empty request",
			args: args{},
		},
		{
			name: "parses all the traffic filtering rules",
			args: args{
				filters: []string{"0.0.0.0/0", "192.168.1.0/24"},
				req:     &models.DeploymentCreateRequest{},
			},
			want: &models.DeploymentCreateRequest{Settings: &models.DeploymentCreateSettings{
				TrafficFilterSettings: &models.TrafficFilterSettings{Rulesets: []string{
					"0.0.0.0/0", "192.168.1.0/24",
				}},
			}},
		},
		{
			name: "parses all the traffic filtering rules",
			args: args{
				filters: []string{"0.0.0.0/0", "192.168.1.0/24"},
				req:     &models.DeploymentCreateRequest{Settings: &models.DeploymentCreateSettings{}},
			},
			want: &models.DeploymentCreateRequest{Settings: &models.DeploymentCreateSettings{
				TrafficFilterSettings: &models.TrafficFilterSettings{Rulesets: []string{
					"0.0.0.0/0", "192.168.1.0/24",
				}},
			}},
		},
		{
			name: "parses all the traffic filtering rules",
			args: args{
				filters: []string{"0.0.0.0/0", "192.168.1.0/24"},
				req: &models.DeploymentCreateRequest{Settings: &models.DeploymentCreateSettings{
					TrafficFilterSettings: &models.TrafficFilterSettings{
						Rulesets: []string{"192.168.0.0/24"},
					},
				}},
			},
			want: &models.DeploymentCreateRequest{Settings: &models.DeploymentCreateSettings{
				TrafficFilterSettings: &models.TrafficFilterSettings{Rulesets: []string{
					"192.168.0.0/24", "0.0.0.0/0", "192.168.1.0/24",
				}},
			}},
		},
		{
			name: "parses no traffic filtering rules",
			args: args{
				filters: nil,
				req:     &models.DeploymentCreateRequest{},
			},
			want: &models.DeploymentCreateRequest{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filters types.Set
			diags := tfsdk.ValueFrom(context.Background(), tt.args.filters, types.SetType{ElemType: types.StringType}, &filters)
			assert.Nil(t, diags)

			diags = TrafficFilterToModel(context.Background(), filters, tt.args.req)
			assert.Nil(t, diags)
			assert.Equal(t, tt.want, tt.args.req)
		})
	}
}
