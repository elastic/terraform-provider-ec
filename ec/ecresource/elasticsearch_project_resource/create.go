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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	if !resourceReady(r, &response.Diagnostics) {
		return
	}

	var model resource_elasticsearch_project.ElasticsearchProjectModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	createBody := serverless.CreateElasticsearchProjectRequest{
		Name:     model.Name.ValueString(),
		RegionId: model.RegionId.ValueString(),
	}

	if model.Alias.ValueString() != "" {
		createBody.Alias = model.Alias.ValueStringPointer()
	}

	if model.OptimizedFor.ValueString() != "" {
		createBody.OptimizedFor = (*serverless.ElasticsearchOptimizedFor)(model.OptimizedFor.ValueStringPointer())
	}

	if util.IsKnown(model.SearchLake) {
		createBody.SearchLake = &serverless.ElasticsearchSearchLake{}

		if util.IsKnown(model.SearchLake.BoostWindow) {
			boostWindow := int(model.SearchLake.BoostWindow.ValueInt64())
			createBody.SearchLake.BoostWindow = &boostWindow
		}

		if util.IsKnown(model.SearchLake.SearchPower) {
			searchPower := int(model.SearchLake.SearchPower.ValueInt64())
			createBody.SearchLake.SearchPower = &searchPower
		}
	}

	resp, err := r.client.CreateElasticsearchProjectWithResponse(ctx, createBody)
	if err != nil {
		response.Diagnostics.AddError(err.Error(), err.Error())
		return
	}

	if resp.JSON201 == nil {
		response.Diagnostics.AddError(
			"Failed to create elasticsearch_project",
			fmt.Sprintf("The API request failed with: %d %s\n%s",
				resp.StatusCode(),
				resp.Status(),
				resp.Body),
		)
		return
	}

	model.Id = types.StringValue(resp.JSON201.Id)

	creds, diags := resource_elasticsearch_project.NewCredentialsValue(
		model.Credentials.AttributeTypes(ctx),
		map[string]attr.Value{
			"username": types.StringValue(resp.JSON201.Credentials.Username),
			"password": types.StringValue(resp.JSON201.Credentials.Password),
		},
	)
	response.Diagnostics.Append(diags...)
	model.Credentials = creds

	response.Diagnostics.Append(response.State.Set(ctx, model)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(r.waitUntilInitialised(ctx, resp.JSON201.Id)...)
	if response.Diagnostics.HasError() {
		return
	}

	found, diags := r.read(ctx, resp.JSON201.Id, &model)
	response.Diagnostics.Append(diags...)

	if !found {
		response.Diagnostics.AddError(
			"Failed to read created Elasticsearch project",
			"The Elasticsearch project was successfully created and initialised, but could then not be read back from the API",
		)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, model)...)
}

func (r *Resource) waitUntilInitialised(ctx context.Context, id string) diag.Diagnostics {
	for {
		resp, err := r.client.GetElasticsearchProjectStatusWithResponse(ctx, id)
		if err != nil {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(err.Error(), err.Error()),
			}
		}

		if resp.JSON200 == nil {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Failed to create elasticsearch_project",
					fmt.Sprintf("The API request failed with: %d %s\n%s",
						resp.StatusCode(),
						resp.Status(),
						resp.Body),
				),
			}
		}

		if resp.JSON200.Phase == serverless.Initialized {
			return nil
		}
	}
}
