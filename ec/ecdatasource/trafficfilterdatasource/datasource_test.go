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
				Id: types.String{Value: "matching-id"},
			}},
			want: want,
		},
		{
			name: "has no matching id or anything else",
			args: args{in: &remoteStateForMatchingId, state: modelV0{
				Id: types.String{Value: "no-matches"},
			}},
			want: wantNoMatches,
		},
		{
			name: "has matching name",
			args: args{in: &remoteStateForMatchingName, state: modelV0{
				Name: types.String{Value: "my traffic filter"},
			}},
			want: wantmatchingName,
		},
		{
			name: "has matching region",
			args: args{in: &remoteStateForMatchingRegion, state: modelV0{
				Region: types.String{Value: "us-east-1"},
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
		Region: types.String{Value: region},
		Rulesets: types.List{
			ElemType: rulesetElemType(),
			Elems: []attr.Value{
				types.Object{
					AttrTypes: rulesetAttrTypes(),
					Attrs: map[string]attr.Value{
						"id":                 types.String{Value: "matching-region"},
						"name":               types.String{Value: "my traffic filter"},
						"region":             types.String{Value: "us-east-1"},
						"include_by_default": types.Bool{Value: false},
						"description":        types.String{Value: "description"},
						"rules":              newSampleTrafficFilterRule("matching-region"),
					},
				},
				types.Object{
					AttrTypes: rulesetAttrTypes(),
					Attrs: map[string]attr.Value{
						"id":                 types.String{Value: "matching-region"},
						"name":               types.String{Value: "my traffic filter"},
						"region":             types.String{Value: "us-east-1"},
						"include_by_default": types.Bool{Value: false},
						"description":        types.String{Value: "description"},
						"rules":              newSampleTrafficFilterRule("matching-region"),
					},
				},
			},
		},
	}
}

func hasMatchingId(id string) modelV0 {
	return modelV0{
		Id: types.String{Value: id},
		Rulesets: types.List{
			ElemType: rulesetElemType(),
			Elems: []attr.Value{
				types.Object{
					AttrTypes: rulesetAttrTypes(),
					Attrs: map[string]attr.Value{
						"id":                 types.String{Value: id},
						"name":               types.String{Value: "my traffic filter"},
						"region":             types.String{Value: "us-east-1"},
						"include_by_default": types.Bool{Value: false},
						"description":        types.String{Value: "description"},
						"rules":              newSampleTrafficFilterRule(id),
					},
				},
			},
		},
	}
}

func hasMatchingName(name string) modelV0 {
	return modelV0{
		Name: types.String{Value: name},
		Rulesets: types.List{
			ElemType: rulesetElemType(),
			Elems: []attr.Value{
				types.Object{
					AttrTypes: rulesetAttrTypes(),
					Attrs: map[string]attr.Value{
						"id":                 types.String{Value: "matching-name"},
						"name":               types.String{Value: name},
						"region":             types.String{Value: "us-east-1"},
						"include_by_default": types.Bool{Value: false},
						"description":        types.String{Value: "description"},
						"rules":              newSampleTrafficFilterRule("matching-name"),
					},
				},
			},
		},
	}
}

func emptyResultSet(id string) modelV0 {
	return modelV0{
		Id: types.String{Value: id},
		Rulesets: types.List{
			ElemType: rulesetElemType(),
			Elems:    []attr.Value{}},
	}
}

func newSampleTrafficFilterRule(id string) types.List {
	return types.List{
		ElemType: ruleElemType(),
		Elems: []attr.Value{
			types.Object{
				AttrTypes: ruleAttrTypes(),
				Attrs: map[string]attr.Value{
					"id":          types.String{Value: id},
					"source":      types.String{Value: "1.1.1.1"},
					"description": types.String{Value: "desc"},
				},
			},
		},
	}
}
