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

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		want []interface{}
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
			want: []interface{}{
				"one-id-of-a-rule",
				"another-id-of-another-rule",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotSlice []interface{}
			if got := flattenTrafficFiltering(tt.args.settings); got != nil {
				gotSlice = got.List()
			}
			assert.Equal(t, tt.want, gotSlice)
		})
	}
}

func Test_expandTrafficFilterCreate(t *testing.T) {
	type args struct {
		v   *schema.Set
		req *models.DeploymentCreateRequest
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
				v:   schema.NewSet(schema.HashString, []interface{}{"0.0.0.0/0", "192.168.1.0/24"}),
				req: &models.DeploymentCreateRequest{},
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
				v:   schema.NewSet(schema.HashString, []interface{}{"0.0.0.0/0", "192.168.1.0/24"}),
				req: &models.DeploymentCreateRequest{Settings: &models.DeploymentCreateSettings{}},
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
				v: schema.NewSet(schema.HashString, []interface{}{"0.0.0.0/0", "192.168.1.0/24"}),
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
				v:   schema.NewSet(schema.HashString, nil),
				req: &models.DeploymentCreateRequest{},
			},
			want: &models.DeploymentCreateRequest{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expandTrafficFilterCreate(tt.args.v, tt.args.req)
			assert.Equal(t, tt.want, tt.args.req)
		})
	}
}
