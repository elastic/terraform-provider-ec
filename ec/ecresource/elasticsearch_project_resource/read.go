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

package elasticsearchprojectresource

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_elasticsearch_project"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	if !resourceReady(r, &response.Diagnostics) {
		return
	}

	var model resource_elasticsearch_project.ElasticsearchProjectModel
	response.Diagnostics.Append(request.State.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	found, diags := r.read(ctx, model.Id.ValueString(), &model)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if !found {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, model)...)
}

func (r *Resource) read(ctx context.Context, id string, state *resource_elasticsearch_project.ElasticsearchProjectModel) (found bool, diags diag.Diagnostics) {
	resp, err := r.client.GetElasticsearchProjectWithResponse(ctx, id)
	if err != nil {
		return false, diag.Diagnostics{
			diag.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}

	if resp.JSON404 != nil {
		return false, nil
	}

	if resp.JSON200 == nil {
		return false, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Failed to create elasticsearch_project",
				fmt.Sprintf("The API request failed with: %d %s\n%s",
					resp.StatusCode(),
					resp.Status(),
					resp.Body),
			),
		}
	}

	state.Id = basetypes.NewStringValue(id)
	state.Alias = basetypes.NewStringValue(reformatAlias(resp.JSON200.Alias, id))
	state.CloudId = basetypes.NewStringValue(resp.JSON200.CloudId)

	endpoints, diags := resource_elasticsearch_project.NewEndpointsValue(
		state.Endpoints.AttributeTypes(ctx),
		map[string]attr.Value{
			"elasticsearch": basetypes.NewStringValue(resp.JSON200.Endpoints.Elasticsearch),
			"kibana":        basetypes.NewStringValue(resp.JSON200.Endpoints.Kibana),
		},
	)
	if diags.HasError() {
		return false, diags
	}
	state.Endpoints = endpoints

	metadataValues := map[string]attr.Value{
		"created_at":       basetypes.NewStringValue(resp.JSON200.Metadata.CreatedAt.String()),
		"created_by":       basetypes.NewStringValue(resp.JSON200.Metadata.CreatedBy),
		"organization_id":  basetypes.NewStringValue(resp.JSON200.Metadata.OrganizationId),
		"suspended_at":     basetypes.NewStringNull(),
		"suspended_reason": basetypes.NewStringPointerValue(resp.JSON200.Metadata.SuspendedReason),
	}

	if resp.JSON200.Metadata.SuspendedAt != nil {
		metadataValues["suspended_at"] = basetypes.NewStringValue(resp.JSON200.Metadata.SuspendedAt.String())
	}

	metadata, diags := resource_elasticsearch_project.NewMetadataValue(
		state.Metadata.AttributeTypes(ctx),
		metadataValues,
	)
	if diags.HasError() {
		return false, diags
	}
	state.Metadata = metadata

	state.Name = basetypes.NewStringValue(resp.JSON200.Name)
	state.OptimizedFor = basetypes.NewStringValue(string(resp.JSON200.OptimizedFor))
	state.RegionId = basetypes.NewStringValue(resp.JSON200.RegionId)
	state.Type = basetypes.NewStringValue(string(resp.JSON200.Type))

	searchLakeValues := map[string]attr.Value{
		"boost_window": basetypes.NewInt64Null(),
		"search_power": basetypes.NewInt64Null(),
	}

	if resp.JSON200.SearchLake != nil {
		if resp.JSON200.SearchLake.BoostWindow != nil {
			searchLakeValues["boost_window"] = basetypes.NewInt64Value(int64(*resp.JSON200.SearchLake.BoostWindow))
		}

		if resp.JSON200.SearchLake.SearchPower != nil {
			searchLakeValues["search_power"] = basetypes.NewInt64Value(int64(*resp.JSON200.SearchLake.SearchPower))
		}
	}
	searchLake, diags := resource_elasticsearch_project.NewSearchLakeValue(
		state.SearchLake.AttributeTypes(ctx),
		searchLakeValues,
	)
	if diags.HasError() {
		return false, diags
	}
	state.SearchLake = searchLake

	return true, nil
}

func reformatAlias(apiAlias string, id string) string {
	shortId := id[0:6]
	reformattedAlias, _ := strings.CutSuffix(apiAlias, fmt.Sprintf("-%s", shortId))
	return reformattedAlias
}
