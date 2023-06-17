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

package trafficfilterdatasource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
)

func Test_modelToState(t *testing.T) {
	remoteStateForMatchingId := models.TrafficFilterRulesets{
		Rulesets: []*models.TrafficFilterRulesetInfo{
			{
				ID:               ec.String("matching-id"),
				Name:             ec.String("my traffic filter"),
				IncludeByDefault: ec.Bool(false),
				Region:           ec.String("us-east-1"),
				Description:      *ec.String("description"),
				Rules: []*models.TrafficFilterRule{
					{ID: "matching-id", Source: "1.1.1.1", Description: "desc"},
				},
			},
		},
	}

	remoteStateForMatchingName := models.TrafficFilterRulesets{
		Rulesets: []*models.TrafficFilterRulesetInfo{
			{
				ID:               ec.String("matching-name"),
				Name:             ec.String("my traffic filter"),
				IncludeByDefault: ec.Bool(false),
				Region:           ec.String("us-east-1"),
				Description:      *ec.String("description"),
				Rules: []*models.TrafficFilterRule{
					{ID: "matching-name", Source: "1.1.1.1", Description: "desc"},
				},
			},
		},
	}

	remoteStateForMatchingRegion := models.TrafficFilterRulesets{
		Rulesets: []*models.TrafficFilterRulesetInfo{
			{
				ID:               ec.String("matching-region"),
				Name:             ec.String("my traffic filter"),
				IncludeByDefault: ec.Bool(false),
				Region:           ec.String("us-east-1"),
				Description:      *ec.String("description"),
				Rules: []*models.TrafficFilterRule{
					{ID: "matching-region", Source: "1.1.1.1", Description: "desc"},
				},
			},
			{
				ID:               ec.String("matching-region"),
				Name:             ec.String("my traffic filter"),
				IncludeByDefault: ec.Bool(false),
				Region:           ec.String("us-east-1"),
				Description:      *ec.String("description"),
				Rules: []*models.TrafficFilterRule{
					{ID: "matching-region", Source: "1.1.1.1", Description: "desc"},
				},
			},
		},
	}

	want := hasMatchingId("matching-id")
	wantmatchingName := hasMatchingName("my traffic filter")
	wantNoMatches := emptyResultSet("no-matches")
	wantRegionMatches := hasMatchingRegion("us-east-1")

	type args struct {
		in    *models.TrafficFilterRulesets
		state modelV0
	}

	tests := []struct {
		name string
		args args
		want modelV0
	}{
		{
			name: "has a matching id",
			args: args{in: &remoteStateForMatchingId, state: modelV0{
				Id: types.StringValue("matching-id"),
			}},
			want: want,
		},
		{
			name: "has no matching id or anything else",
			args: args{in: &remoteStateForMatchingId, state: modelV0{
				Id: types.StringValue("no-matches"),
			}},
			want: wantNoMatches,
		},
		{
			name: "has matching name",
			args: args{in: &remoteStateForMatchingName, state: modelV0{
				Name: types.StringValue("my traffic filter"),
			}},
			want: wantmatchingName,
		},
		{
			name: "has matching region",
			args: args{in: &remoteStateForMatchingRegion, state: modelV0{
				Region: types.StringValue("us-east-1"),
			}},
			want: wantRegionMatches,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelToState(context.Background(), tt.args.in, &tt.args.state)
			assert.Equal(t, tt.want, tt.args.state)
		})
	}
}

func hasMatchingRegion(region string) modelV0 {
	return modelV0{
		Region: types.StringValue(region),
		Rulesets: types.ListValueMust(rulesetElemType(), []attr.Value{
			types.ObjectValueMust(rulesetAttrTypes(), map[string]attr.Value{
				"id":                 types.StringValue("matching-region"),
				"name":               types.StringValue("my traffic filter"),
				"region":             types.StringValue("us-east-1"),
				"include_by_default": types.BoolValue(false),
				"description":        types.StringValue("description"),
				"rules":              newSampleTrafficFilterRule("matching-region"),
			}),
			types.ObjectValueMust(rulesetAttrTypes(), map[string]attr.Value{
				"id":                 types.StringValue("matching-region"),
				"name":               types.StringValue("my traffic filter"),
				"region":             types.StringValue("us-east-1"),
				"include_by_default": types.BoolValue(false),
				"description":        types.StringValue("description"),
				"rules":              newSampleTrafficFilterRule("matching-region"),
			}),
		}),
	}
}

func hasMatchingId(id string) modelV0 {
	return modelV0{
		Id: types.StringValue(id),
		Rulesets: types.ListValueMust(rulesetElemType(), []attr.Value{
			types.ObjectValueMust(rulesetAttrTypes(), map[string]attr.Value{
				"id":                 types.StringValue(id),
				"name":               types.StringValue("my traffic filter"),
				"region":             types.StringValue("us-east-1"),
				"include_by_default": types.BoolValue(false),
				"description":        types.StringValue("description"),
				"rules":              newSampleTrafficFilterRule(id),
			}),
		}),
	}
}

func hasMatchingName(name string) modelV0 {
	return modelV0{
		Name: types.StringValue(name),
		Rulesets: types.ListValueMust(rulesetElemType(), []attr.Value{
			types.ObjectValueMust(rulesetAttrTypes(), map[string]attr.Value{
				"id":                 types.StringValue("matching-name"),
				"name":               types.StringValue(name),
				"region":             types.StringValue("us-east-1"),
				"include_by_default": types.BoolValue(false),
				"description":        types.StringValue("description"),
				"rules":              newSampleTrafficFilterRule("matching-name"),
			}),
		}),
	}
}

func emptyResultSet(id string) modelV0 {
	return modelV0{
		Id:       types.StringValue(id),
		Rulesets: types.ListValueMust(rulesetElemType(), []attr.Value{}),
	}
}

func newSampleTrafficFilterRule(id string) types.List {
	return types.ListValueMust(ruleElemType(), []attr.Value{
		types.ObjectValueMust(ruleAttrTypes(), map[string]attr.Value{
			"id":          types.StringValue(id),
			"source":      types.StringValue("1.1.1.1"),
			"description": types.StringValue("desc"),
		}),
	},
	)
}
