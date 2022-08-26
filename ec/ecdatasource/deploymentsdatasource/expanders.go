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
	"fmt"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// expandFilters expands all filters into a search request model
func expandFilters(d *schema.ResourceData) (*models.SearchRequest, error) {
	var queries []*models.QueryContainer

	namePrefix := d.Get("name_prefix").(string)
	if namePrefix != "" {
		queries = append(queries, &models.QueryContainer{
			Prefix: map[string]models.PrefixQuery{
				// The "keyword" addition denotes that the query will be using a keyword
				// field rather than a text field in order to ensure the query is not analyzed
				"name.keyword": {Value: ec.String(namePrefix)},
			},
		})
	}

	depTemplateID := d.Get("deployment_template_id").(string)
	if depTemplateID != "" {
		esPath := "resources.elasticsearch"
		tplTermPath := esPath + ".info.plan_info.current.plan.deployment_template.id"

		queries = append(queries, newNestedTermQuery(esPath, tplTermPath, depTemplateID))
	}

	healthy := d.Get("healthy").(string)
	if healthy != "" {
		if healthy != "true" && healthy != "false" {
			return nil, fmt.Errorf("invalid value for healthy (true|false): '%s'", healthy)
		}

		queries = append(queries, &models.QueryContainer{
			Term: map[string]models.TermQuery{
				"healthy": {Value: &healthy},
			},
		})
	}

	tags := d.Get("tags").(map[string]interface{})
	var tagQueries []*models.QueryContainer
	for key, value := range tags {
		tagQueries = append(tagQueries,
			newNestedTagQuery(key, value.(string)),
		)
	}
	if len(tagQueries) > 0 {
		queries = append(queries, &models.QueryContainer{
			Bool: &models.BoolQuery{
				MinimumShouldMatch: int32(len(tags)),
				Should:             tagQueries,
			},
		})
	}

	validResourceKinds := []string{util.Elasticsearch, util.Kibana,
		util.Apm, util.EnterpriseSearch, util.IntegrationsServer}

	for _, resourceKind := range validResourceKinds {
		req, err := expandResourceFilters(d.Get(resourceKind).([]interface{}), resourceKind)
		if err != nil {
			return nil, err
		}
		queries = append(queries, req...)
	}

	searchReq := models.SearchRequest{
		Size: int32(d.Get("size").(int)),
		Sort: []interface{}{"id"},
	}

	if len(queries) > 0 {
		searchReq.Query = &models.QueryContainer{
			Bool: &models.BoolQuery{
				Filter: []*models.QueryContainer{
					{
						Bool: &models.BoolQuery{
							Must: queries,
						},
					},
				},
			},
		}
	}

	return &searchReq, nil
}

// expandResourceFilters expands filters from a specific resource kind into query models
func expandResourceFilters(resources []interface{}, resourceKind string) ([]*models.QueryContainer, error) {
	if len(resources) == 0 {
		return nil, nil
	}

	var queries []*models.QueryContainer

	for _, raw := range resources {
		var q = raw.(map[string]interface{})

		resourceKindPath := "resources." + resourceKind

		if status, ok := q["status"].(string); ok && status != "" {
			statusTermPath := resourceKindPath + ".info.status"

			queries = append(queries,
				newNestedTermQuery(resourceKindPath, statusTermPath, status))
		}

		if version, ok := q["version"].(string); ok && version != "" {
			versionTermPath := resourceKindPath + ".info.plan_info.current.plan." +
				resourceKind + ".version"

			queries = append(queries,
				newNestedTermQuery(resourceKindPath, versionTermPath, version))
		}

		if healthy, ok := q["healthy"].(string); ok && healthy != "" {
			healthyTermPath := resourceKindPath + ".info.healthy"

			if healthy != "true" && healthy != "false" {
				return nil, fmt.Errorf("invalid value for healthy (true|false): '%s'", healthy)
			}

			queries = append(queries,
				newNestedTermQuery(resourceKindPath, healthyTermPath, healthy))
		}
	}

	return queries, nil
}

func newNestedTermQuery(path, term string, value string) *models.QueryContainer {
	return &models.QueryContainer{
		Nested: &models.NestedQuery{
			Path: ec.String(path),
			Query: &models.QueryContainer{
				Term: map[string]models.TermQuery{
					term: {
						Value: ec.String(value),
					},
				},
			},
		},
	}
}

// newNestedTagQuery returns a nested query for a metadata tag
func newNestedTagQuery(key string, value string) *models.QueryContainer {
	return &models.QueryContainer{
		Nested: &models.NestedQuery{
			Path: ec.String("metadata.tags"),
			Query: &models.QueryContainer{
				Bool: &models.BoolQuery{
					Filter: []*models.QueryContainer{
						{
							Term: map[string]models.TermQuery{
								"metadata.tags.key": {
									Value: &key,
								},
							},
						},
						{
							Term: map[string]models.TermQuery{
								"metadata.tags.value": {
									Value: &value,
								},
							},
						},
					},
				},
			},
		},
	}
}
