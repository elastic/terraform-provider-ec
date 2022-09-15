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

	"github.com/elastic/cloud-sdk-go/pkg/models"
)

func Test_flattenTrafficFiltering(t *testing.T) {
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
			want: []string{},
		},
		{
			name: "parses rules",
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
			var newState modelV0
			diags := flattenTrafficFiltering(context.Background(), tt.args.settings, &newState.TrafficFilter)
			assert.Empty(t, diags)
			var got []string
			newState.TrafficFilter.ElementsAs(context.Background(), &got, false)
			assert.Equal(t, tt.want, got)
		})
	}
}
