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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
			name: "parses no rules when they're empty 1",
			args: args{},
		},
		{
			name: "parses no rules when they're empty 2",
			args: args{settings: &models.DeploymentSettings{}},
		},
		{
			name: "parses no rules when they're empty 3",
			args: args{settings: &models.DeploymentSettings{
				TrafficFilterSettings: &models.TrafficFilterSettings{},
			}},
		},
		{
			name: "parses no rules when they're empty 4",
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
			trafficFilter, diags := flattenTrafficFiltering(context.Background(), tt.args.settings)
			assert.Empty(t, diags)
			var got []string
			trafficFilter.ElementsAs(context.Background(), &got, false)
			assert.Equal(t, tt.want, got)

			checkConverionToAttrValue(t, "traffic_filter", trafficFilter)
		})
	}
}

// checking conversion to attr.Value
// it should catch cases when e.g. the func under test returns types.List{}
func checkConverionToAttrValue(t *testing.T, attributeName string, attributeValue types.List) {
	var target types.List
	diags := tfsdk.ValueFrom(context.Background(), attributeValue, attributeType(t, attributeName), &target)
	assert.Nil(t, diags)
}

func attributeType(t *testing.T, attributeName string) attr.Type {
	var d DataSource
	schema, diags := d.GetSchema(context.Background())
	assert.Nil(t, diags)
	return schema.Attributes[attributeName].FrameworkType()
}
