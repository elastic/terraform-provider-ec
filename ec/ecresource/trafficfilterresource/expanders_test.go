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

package trafficfilterresource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
)

func Test_expandModel(t *testing.T) {
	trafficFilterRD := newSampleTrafficFilter(t, "some-random-id")

	trafficFilterMultipleRD := modelV0{
		ID:               types.StringValue("some-random-id"),
		Name:             types.StringValue("my traffic filter"),
		Type:             types.StringValue("ip"),
		IncludeByDefault: types.BoolValue(false),
		Region:           types.StringValue("us-east-1"),
		Rule: func() types.Set {
			res, diags := types.SetValue(
				trafficFilterRuleElemType(),
				[]attr.Value{
					newSampleTrafficFilterRule(t, "1.1.1.1/24", "", "", "", "", "", ""),
					newSampleTrafficFilterRule(t, "1.1.1.0/16", "", "", "", "", "", ""),
					newSampleTrafficFilterRule(t, "0.0.0.0/0", "", "", "", "", "", ""),
					newSampleTrafficFilterRule(t, "1.1.1.1", "", "", "", "", "", ""),
				},
			)
			assert.Nil(t, diags)
			return res
		}(),
	}
	type args struct {
		state modelV0
	}
	tests := []struct {
		name string
		args args
		want *models.TrafficFilterRulesetRequest
	}{
		{
			name: "parses the resource",
			args: args{state: trafficFilterRD},
			want: &models.TrafficFilterRulesetRequest{
				Name:             ec.String("my traffic filter"),
				Type:             ec.String("ip"),
				IncludeByDefault: ec.Bool(false),
				Region:           ec.String("us-east-1"),
				Rules: []*models.TrafficFilterRule{
					{Source: "1.1.1.1"},
					{Source: "0.0.0.0/0"},
				},
			},
		},
		{
			name: "parses the resource with a lot of traffic rules",
			args: args{state: trafficFilterMultipleRD},
			want: &models.TrafficFilterRulesetRequest{
				Name:             ec.String("my traffic filter"),
				Type:             ec.String("ip"),
				IncludeByDefault: ec.Bool(false),
				Region:           ec.String("us-east-1"),
				Rules: []*models.TrafficFilterRule{
					{Source: "1.1.1.1/24"},
					{Source: "1.1.1.0/16"},
					{Source: "0.0.0.0/0"},
					{Source: "1.1.1.1"},
				},
			},
		},
		{
			name: "parses an Azure privatelink resource",
			args: args{
				state: modelV0{
					ID:               types.StringValue("some-random-id"),
					Name:             types.StringValue("my traffic filter"),
					Type:             types.StringValue("azure_private_endpoint"),
					IncludeByDefault: types.BoolValue(false),
					Region:           types.StringValue("azure-australiaeast"),
					Rule: func() types.Set {
						res, diags := types.SetValue(
							trafficFilterRuleElemType(),
							[]attr.Value{
								newSampleTrafficFilterRule(t, "", "", "my-azure-pl", "1231312-1231-1231-1231-1231312", "", "", ""),
							},
						)
						assert.Nil(t, diags)
						return res
					}(),
				},
			},
			want: &models.TrafficFilterRulesetRequest{
				Name:             ec.String("my traffic filter"),
				Type:             ec.String("azure_private_endpoint"),
				IncludeByDefault: ec.Bool(false),
				Region:           ec.String("azure-australiaeast"),
				Rules: []*models.TrafficFilterRule{
					{
						AzureEndpointGUID: "1231312-1231-1231-1231-1231312",
						AzureEndpointName: "my-azure-pl",
					},
				},
			},
		},
		{
			name: "parses a remote cluster resource",
			args: args{
				state: modelV0{
					ID:               types.StringValue("some-random-id"),
					Name:             types.StringValue("my traffic filter"),
					Type:             types.StringValue("remote_cluster"),
					IncludeByDefault: types.BoolValue(false),
					Region:           types.StringValue("us-east-1"),
					Rule: func() types.Set {
						res, diags := types.SetValue(
							trafficFilterRuleElemType(),
							[]attr.Value{
								newSampleTrafficFilterRule(t, "", "", "", "", "remote-cluster-id-123", "123123123", ""),
							},
						)
						assert.Nil(t, diags)
						return res
					}(),
				},
			},
			want: &models.TrafficFilterRulesetRequest{
				Name:             ec.String("my traffic filter"),
				Type:             ec.String("remote_cluster"),
				IncludeByDefault: ec.Bool(false),
				Region:           ec.String("us-east-1"),
				Rules: []*models.TrafficFilterRule{
					{
						RemoteClusterID:    "remote-cluster-id-123",
						RemoteClusterOrgID: "123123123",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, diags := expandModel(context.Background(), tt.args.state)
			assert.Empty(t, diags)
			assert.Equal(t, tt.want, got)
		})
	}
}
