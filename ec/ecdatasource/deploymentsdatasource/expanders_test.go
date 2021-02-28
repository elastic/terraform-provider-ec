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
	"encoding/json"
	"errors"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_expandFilters(t *testing.T) {
	deploymentsDS := util.NewResourceData(t, util.ResDataParams{
		ID:     "myID",
		State:  newSampleFilters(),
		Schema: newSchema(),
	})
	invalidDS := util.NewResourceData(t, util.ResDataParams{
		ID:     "myID",
		State:  newInvalidFilters(),
		Schema: newSchema(),
	})
	type args struct {
		d *schema.ResourceData
	}
	tests := []struct {
		name string
		args args
		want *models.SearchRequest
		err  error
	}{
		{
			name: "parses the data source",
			args: args{d: deploymentsDS},
			want: &models.SearchRequest{
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
			name: "fails to parse the data source",
			args: args{d: invalidDS},
			err:  errors.New("strconv.ParseBool: parsing \"invalid value\": invalid syntax"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandFilters(tt.args.d)
			if tt.err != nil || err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}

			jsonWant, err := json.Marshal(tt.want)
			if err != nil {
				panic(err)
			}
			jsonGot, err := json.Marshal(got)
			if err != nil {
				panic(err)
			}

			assert.Equal(t, jsonWant, jsonGot)

		})
	}
}

func newInvalidFilters() map[string]interface{} {
	return map[string]interface{}{
		"healthy": "invalid value",
		"apm": []interface{}{
			map[string]interface{}{
				"healthy": "invalid value",
			},
		},
	}
}

func newSampleFilters() map[string]interface{} {
	return map[string]interface{}{
		"name_prefix": "test",
		"healthy":     "true",
		"tags": map[string]interface{}{
			"foo": "bar",
		},
		"elasticsearch": []interface{}{
			map[string]interface{}{
				"version": "7.9.1",
			},
		},
		"kibana": []interface{}{
			map[string]interface{}{
				"status": "started",
			},
		},
		"apm": []interface{}{
			map[string]interface{}{
				"healthy": "true",
			},
		},
		"enterprise_search": []interface{}{
			map[string]interface{}{
				"healthy": "false",
			},
		},
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
				"healthy": {Value: true},
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
							Value: "started",
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
							Value: true,
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
							Value: false,
						},
					},
				},
			},
		},
	}
}
