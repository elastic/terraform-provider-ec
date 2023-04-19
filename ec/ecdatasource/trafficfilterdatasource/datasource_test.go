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
	remoteState := models.TrafficFilterRulesets{
		Rulesets: []*models.TrafficFilterRulesetInfo{
			{
				ID:               ec.String("basic"),
				Name:             ec.String("my traffic filter"),
				IncludeByDefault: ec.Bool(false),
				Region:           ec.String("us-east-1"),
				Description:      *ec.String("hhh"),
				Rules: []*models.TrafficFilterRule{
					{ID: "rule-1", Source: "1.1.1.1", Description: "desc"},
				},
			},
		},
	}

	want := newSampleTrafficFilterRuleset("basic")

	type args struct {
		in *models.TrafficFilterRulesets
	}

	tests := []struct {
		name string
		args args
		err  error
		want modelV0
	}{
		{
			name: "flattens the resource",
			args: args{in: &remoteState},
			want: want,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := modelV0{
				Id: types.String{Value: "basic"},
			}
			diags := modelToState(context.Background(), tt.args.in, &state)

			if tt.err != nil {
				assert.Equal(t, diags, tt.err)
			} else {
				assert.Empty(t, diags)
			}

			assert.Equal(t, tt.want, state)
		})
	}
}

func newSampleTrafficFilterRuleset(id string) modelV0 {
	return modelV0{
		Id: types.String{Value: id},
		Rulesets: types.List{
			ElemType: rulesetsElemType(),
			Elems: []attr.Value{
				types.Object{
					AttrTypes: rulesetsAttrTypes(),
					Attrs: map[string]attr.Value{
						"id":                 types.String{Value: id},
						"name":               types.String{Value: "my traffic filter"},
						"region":             types.String{Value: "us-east-1"},
						"include_by_default": types.Bool{Value: false},
						"description":        types.String{Value: "hhh"},
						"rules": types.List{
							ElemType: ruleElemType(),
							Elems: []attr.Value{
								types.Object{
									AttrTypes: ruleAttrTypes(),
									Attrs: map[string]attr.Value{
										"id":          types.String{Value: "rule-1"},
										"source":      types.String{Value: "1.1.1.1"},
										"description": types.String{Value: "desc"},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
