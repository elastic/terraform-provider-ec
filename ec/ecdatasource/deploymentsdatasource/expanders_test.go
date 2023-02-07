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

package deploymentsdatasource

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_expandFilters(t *testing.T) {
	type args struct {
		state modelV0
	}
	tests := []struct {
		name  string
		args  args
		want  *models.SearchRequest
		diags diag.Diagnostics
	}{
		{
			name: "parses the data source",
			args: args{state: newSampleFilters(t)},
			want: &models.SearchRequest{
				Size: 100,
				Sort: []interface{}{"id"},
				Query: &models.QueryContainer{
					Bool: &models.BoolQuery{
						Filter: []*models.QueryContainer{
							{
								Bool: &models.BoolQuery{
									Must: newTestQuery(),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "parses the data source with a different size",
			args: args{
				state: modelV0{
					NamePrefix: types.StringValue("test"),
					Healthy:    types.StringValue("true"),
					Size:       types.Int64Value(200),
					Tags:       util.StringMapAsType(t, map[string]string{"foo": "bar"}),
					Elasticsearch: func() types.List {
						res, diags := types.ListValueFrom(
							context.Background(),
							types.ObjectType{AttrTypes: resourceFiltersAttrTypes(util.ElasticsearchResourceKind)},
							[]resourceFiltersModelV0{
								{
									Healthy: types.StringNull(),
									Status:  types.StringNull(),
									Version: types.StringValue("7.9.1"),
								},
							},
						)

						assert.Nil(t, diags)
						return res
					}(),
					Kibana: func() types.List {
						res, diags := types.ListValueFrom(
							context.Background(),
							types.ObjectType{AttrTypes: resourceFiltersAttrTypes(util.KibanaResourceKind)},
							[]resourceFiltersModelV0{
								{
									Healthy: types.StringNull(),
									Status:  types.StringValue("started"),
									Version: types.StringNull(),
								},
							},
						)
						assert.Nil(t, diags)

						return res
					}(),
					Apm: func() types.List {
						res, diags := types.ListValueFrom(
							context.Background(),
							types.ObjectType{AttrTypes: resourceFiltersAttrTypes(util.ApmResourceKind)},
							[]resourceFiltersModelV0{
								{
									Healthy: types.StringValue("true"),
									Status:  types.StringNull(),
									Version: types.StringNull(),
								},
							},
						)
						assert.Nil(t, diags)

						return res
					}(),
					EnterpriseSearch: func() types.List {
						res, diags := types.ListValueFrom(
							context.Background(),
							types.ObjectType{AttrTypes: resourceFiltersAttrTypes(util.EnterpriseSearchResourceKind)},
							[]resourceFiltersModelV0{
								{
									Status:  types.StringNull(),
									Healthy: types.StringValue("false"),
									Version: types.StringNull(),
								},
							},
						)
						assert.Nil(t, diags)

						return res
					}(),
				},
			},
			want: &models.SearchRequest{
				Size: 200,
				Sort: []interface{}{"id"},
				Query: &models.QueryContainer{
					Bool: &models.BoolQuery{
						Filter: []*models.QueryContainer{
							{
								Bool: &models.BoolQuery{
									Must: newTestQuery(),
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "fails to parse the data source",
			args:  args{state: newInvalidFilters(t)},
			diags: diag.Diagnostics{diag.NewErrorDiagnostic("invalid value for healthy", "expected either [true] or [false] but got [invalid value]")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, diags := expandFilters(context.Background(), tt.args.state)
			if tt.diags != nil {
				assert.Equal(t, tt.diags, diags)
			} else {
				assert.Empty(t, diags)
			}

			jsonWant, err := json.MarshalIndent(tt.want, "", "  ")
			if err != nil {
				t.Error("Unable to marshal wanted struct to JSON")
			}

			jsonGot, err := json.MarshalIndent(got, "", "  ")
			if err != nil {
				t.Error("Unable to marshal received struct to JSON")
			}

			assert.Equal(t, string(jsonWant), string(jsonGot))
		})
	}
}

func newInvalidFilters(t *testing.T) modelV0 {
	return modelV0{
		Healthy: types.StringValue("invalid value"),
		Apm: func() types.List {
			res, diags := types.ListValueFrom(
				context.Background(),
				types.ObjectType{AttrTypes: resourceFiltersAttrTypes(util.ApmResourceKind)},
				[]resourceFiltersModelV0{
					{
						Healthy: types.StringValue("invalid healthy value"),
						Status:  types.StringValue("invalid status value"),
						Version: types.StringValue("invalid version value"),
					},
				},
			)
			assert.Nil(t, diags)

			return res
		}(),
	}
}

func newSampleFilters(t *testing.T) modelV0 {
	return modelV0{
		NamePrefix: types.StringValue("test"),
		Healthy:    types.StringValue("true"),
		Size:       types.Int64Value(100),
		Tags: func() types.Map {
			res, diags := types.MapValue(
				types.StringType,
				map[string]attr.Value{"foo": types.StringValue("bar")},
			)
			assert.Nil(t, diags)
			return res
		}(),
		Elasticsearch: func() types.List {
			res, diags := types.ListValueFrom(
				context.Background(),
				types.ObjectType{AttrTypes: resourceFiltersAttrTypes(util.ElasticsearchResourceKind)},
				[]resourceFiltersModelV0{
					{
						Healthy: types.StringNull(),
						Status:  types.StringNull(),
						Version: types.StringValue("7.9.1"),
					},
				},
			)
			assert.Nil(t, diags)

			return res
		}(),
		Kibana: func() types.List {
			res, diags := types.ListValueFrom(
				context.Background(),
				types.ObjectType{AttrTypes: resourceFiltersAttrTypes(util.KibanaResourceKind)},
				[]resourceFiltersModelV0{
					{
						Healthy: types.StringNull(),
						Status:  types.StringValue("started"),
						Version: types.StringNull(),
					},
				},
			)
			assert.Nil(t, diags)

			return res
		}(),
		Apm: func() types.List {
			res, diags := types.ListValueFrom(
				context.Background(),
				types.ObjectType{AttrTypes: resourceFiltersAttrTypes(util.ApmResourceKind)},
				[]resourceFiltersModelV0{
					{
						Healthy: types.StringValue("true"),
						Status:  types.StringNull(),
						Version: types.StringNull(),
					},
				},
			)
			assert.Nil(t, diags)

			return res
		}(),
		EnterpriseSearch: func() types.List {
			res, diags := types.ListValueFrom(
				context.Background(),
				types.ObjectType{AttrTypes: resourceFiltersAttrTypes(util.EnterpriseSearchResourceKind)},
				[]resourceFiltersModelV0{
					{
						Status:  types.StringNull(),
						Healthy: types.StringValue("false"),
						Version: types.StringNull(),
					},
				},
			)
			assert.Nil(t, diags)

			return res
		}(),
	}
}

func newTestQuery() []*models.QueryContainer {
	return []*models.QueryContainer{
		{
			Prefix: map[string]models.PrefixQuery{
				"name.keyword": {Value: ec.String("test")},
			},
		},
		{
			Term: map[string]models.TermQuery{
				"healthy": {Value: ec.String("true")},
			},
		},
		{
			Bool: &models.BoolQuery{
				MinimumShouldMatch: int32(1),
				Should: []*models.QueryContainer{
					{
						Nested: &models.NestedQuery{
							Path: ec.String("metadata.tags"),
							Query: &models.QueryContainer{
								Bool: &models.BoolQuery{
									Filter: []*models.QueryContainer{
										{
											Term: map[string]models.TermQuery{
												"metadata.tags.key": {
													Value: ec.String("foo"),
												},
											},
										},
										{
											Term: map[string]models.TermQuery{
												"metadata.tags.value": {
													Value: ec.String("bar"),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Nested: &models.NestedQuery{
				Path: ec.String("resources.elasticsearch"),
				Query: &models.QueryContainer{
					Term: map[string]models.TermQuery{
						"resources.elasticsearch.info.plan_info.current.plan.elasticsearch.version": {
							Value: ec.String("7.9.1"),
						},
					},
				},
			},
		},
		{
			Nested: &models.NestedQuery{
				Path: ec.String("resources.kibana"),
				Query: &models.QueryContainer{
					Term: map[string]models.TermQuery{
						"resources.kibana.info.status": {
							Value: ec.String("started"),
						},
					},
				},
			},
		},
		{
			Nested: &models.NestedQuery{
				Path: ec.String("resources.apm"),
				Query: &models.QueryContainer{
					Term: map[string]models.TermQuery{
						"resources.apm.info.healthy": {
							Value: ec.String("true"),
						},
					},
				},
			},
		},
		{
			Nested: &models.NestedQuery{
				Path: ec.String("resources.enterprise_search"),
				Query: &models.QueryContainer{
					Term: map[string]models.TermQuery{
						"resources.enterprise_search.info.healthy": {
							Value: ec.String("false"),
						},
					},
				},
			},
		},
	}
}
