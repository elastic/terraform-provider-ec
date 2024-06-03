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

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_elasticsearch_project"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	if !resourceReady(r, &response.Diagnostics) {
		return
	}

	var model resource_elasticsearch_project.ElasticsearchProjectModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	updateBody := serverless.PatchElasticsearchProjectRequest{
		Name: model.Name.ValueStringPointer(),
	}

	if model.Alias.ValueString() != "" {
		updateBody.Alias = model.Alias.ValueStringPointer()
	}

	if util.IsKnown(model.SearchLake) {
		updateBody.SearchLake = &serverless.OptionalElasticsearchSearchLake{}

		if util.IsKnown(model.SearchLake.BoostWindow) {
			boostWindow := int(model.SearchLake.BoostWindow.ValueInt64())
			updateBody.SearchLake.BoostWindow = &boostWindow
		}

		if util.IsKnown(model.SearchLake.SearchPower) {
			searchPower := int(model.SearchLake.SearchPower.ValueInt64())
			updateBody.SearchLake.SearchPower = &searchPower
		}
	}

	resp, err := r.client.PatchElasticsearchProjectWithResponse(ctx, model.Id.ValueString(), nil, updateBody)
	if err != nil {
		response.Diagnostics.AddError(err.Error(), err.Error())
		return
	}

	if resp.JSON200 == nil {
		response.Diagnostics.AddError(
			"Failed to update elasticsearch_project",
			fmt.Sprintf("The API request failed with: %d %s\n%s",
				resp.StatusCode(),
				resp.Status(),
				resp.Body),
		)
		return
	}

	found, diags := r.read(ctx, resp.JSON200.Id, &model)
	response.Diagnostics.Append(diags...)

	if !found {
		response.Diagnostics.AddError(
			"Failed to read updated Elasticsearch project",
			"The Elasticsearch project was successfully update, but could then not be read back from the API",
		)
	}

	response.Diagnostics.Append(response.State.Set(ctx, model)...)
}
