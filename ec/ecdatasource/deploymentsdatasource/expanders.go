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
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
)

// expandFilters expands all filters into a search request model
func expandFilters(ctx context.Context, state modelV0) (*models.SearchRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	var queries []*models.QueryContainer

	namePrefix := state.NamePrefix.ValueString()
	if namePrefix != "" {
		queries = append(queries, &models.QueryContainer{
			Prefix: map[string]models.PrefixQuery{
				// The "keyword" addition denotes that the query will be using a keyword
				// field rather than a text field in order to ensure the query is not analyzed
				"name.keyword": {Value: &namePrefix},
			},
		})
	}

	name := state.Name.ValueString()
	if name != "" {
		queries = append(queries, &models.QueryContainer{
			Term: map[string]models.TermQuery{
				"name.keyword": {Value: &name},
			},
		})
	}

	depTemplateID := state.DeploymentTemplateID.ValueString()
	if depTemplateID != "" {
		esPath := "resources.elasticsearch"
		tplTermPath := esPath + ".info.plan_info.current.plan.deployment_template.id"

		queries = append(queries, newNestedTermQuery(esPath, tplTermPath, depTemplateID))
	}

	healthy := state.Healthy.ValueString()
	if healthy != "" {
		if healthy != "true" && healthy != "false" {
			diags.AddError("invalid value for healthy",
				fmt.Sprintf("expected either [true] or [false] but got [%s]", healthy))
			return nil, diags
		}

		queries = append(queries, &models.QueryContainer{
			Term: map[string]models.TermQuery{
				"healthy": {Value: &healthy},
			},
		})
	}

	var tags = make(map[string]string)
	diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
	if diags.HasError() {
		return nil, diags
	}

	var tagQueries []*models.QueryContainer
	for key, value := range tags {
		tagQueries = append(tagQueries,
			newNestedTagQuery(key, value),
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
	type resourceFilter struct {
		resourceKind string
		settings     *types.List
	}

	resourceFilters := []resourceFilter{
		{resourceKind: util.Elasticsearch, settings: &state.Elasticsearch},
		{resourceKind: util.Kibana, settings: &state.Kibana},
		{resourceKind: util.Apm, settings: &state.Apm},
		{resourceKind: util.EnterpriseSearch, settings: &state.EnterpriseSearch},
		{resourceKind: util.IntegrationsServer, settings: &state.IntegrationsServer},
	}

	for _, filter := range resourceFilters {
		req, diags := expandResourceFilters(ctx, filter.settings, filter.resourceKind)
		if diags.HasError() {
			return nil, diags
		}
		queries = append(queries, req...)
	}

	searchReq := models.SearchRequest{
		Size: int32(state.Size.ValueInt64()),
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
func expandResourceFilters(ctx context.Context, resources *types.List, resourceKind string) ([]*models.QueryContainer, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(resources.Elements()) == 0 {
		return nil, nil
	}
	var filters []resourceFiltersModelV0
	var queries []*models.QueryContainer
	diags.Append(resources.ElementsAs(ctx, &filters, false)...)
	if diags.HasError() {
		return nil, diags
	}
	for _, filter := range filters {
		resourceKindPath := "resources." + resourceKind

		if filter.Status.ValueString() != "" {
			statusTermPath := resourceKindPath + ".info.status"

			queries = append(queries,
				newNestedTermQuery(resourceKindPath, statusTermPath, filter.Status.ValueString()))
		}

		if filter.Version.ValueString() != "" {
			versionTermPath := resourceKindPath + ".info.plan_info.current.plan." +
				resourceKind + ".version"

			queries = append(queries,
				newNestedTermQuery(resourceKindPath, versionTermPath, filter.Version.ValueString()))
		}

		if filter.Healthy.ValueString() != "" {
			healthyTermPath := resourceKindPath + ".info.healthy"

			if filter.Healthy.ValueString() != "true" && filter.Healthy.ValueString() != "false" {
				diags.AddError("invalid value for healthy", fmt.Sprintf("expected either [true] or [false] but got [%s]", filter.Healthy.ValueString()))
				return nil, diags
			}

			queries = append(queries,
				newNestedTermQuery(resourceKindPath, healthyTermPath, filter.Healthy.ValueString()))
		}
	}

	return queries, nil
}

func newNestedTermQuery(path, term string, value string) *models.QueryContainer {
	return &models.QueryContainer{
		Nested: &models.NestedQuery{
			Path: &path,
			Query: &models.QueryContainer{
				Term: map[string]models.TermQuery{
					term: {
						Value: &value,
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
