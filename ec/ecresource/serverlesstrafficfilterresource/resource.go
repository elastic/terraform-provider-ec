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

package serverlesstrafficfilterresource

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-ec/ec/internal"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_serverless_traffic_filter"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

type Resource struct {
	client serverless.ClientWithResponsesInterface
}

func New() *Resource {
	return &Resource{}
}

func (r *Resource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = fmt.Sprintf("%s_serverless_traffic_filter", request.ProviderTypeName)
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	generatedSchema := resource_serverless_traffic_filter.ServerlessTrafficFilterResourceSchema(ctx)

	// Use generated schema as-is since rules is already a list attribute
	resp.Schema = generatedSchema

}

func (r *Resource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	clients, diags := internal.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	r.client = clients.Serverless
}

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	if !r.ready(&response.Diagnostics) {
		return
	}

	var model resource_serverless_traffic_filter.ServerlessTrafficFilterModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	createReq := toCreateRequest(ctx, model, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	resp, err := r.client.CreateTrafficFilterWithResponse(ctx, createReq)
	if err != nil {
		response.Diagnostics.AddError("Failed to create traffic filter", err.Error())
		return
	}

	if resp.StatusCode() != 201 {
		response.Diagnostics.AddError(
			"Failed to create traffic filter",
			fmt.Sprintf("Unexpected status code: %d, body: %s", resp.StatusCode(), string(resp.Body)),
		)
		return
	}

	model = fromTrafficFilterInfo(ctx, *resp.JSON201, &response.Diagnostics)
	response.Diagnostics.Append(response.State.Set(ctx, model)...)
}

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	if !r.ready(&response.Diagnostics) {
		return
	}

	var model resource_serverless_traffic_filter.ServerlessTrafficFilterModel
	response.Diagnostics.Append(request.State.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := model.Id.ValueString()
	resp, err := r.client.GetTrafficFilterWithResponse(ctx, id)
	if err != nil {
		response.Diagnostics.AddError("Failed to read traffic filter", err.Error())
		return
	}

	if resp.StatusCode() == 404 {
		response.State.RemoveResource(ctx)
		return
	}

	if resp.StatusCode() != 200 {
		response.Diagnostics.AddError(
			"Failed to read traffic filter",
			fmt.Sprintf("Unexpected status code: %d, body: %s", resp.StatusCode(), string(resp.Body)),
		)
		return
	}

	model = fromTrafficFilterInfo(ctx, *resp.JSON200, &response.Diagnostics)
	response.Diagnostics.Append(response.State.Set(ctx, model)...)
}

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	if !r.ready(&response.Diagnostics) {
		return
	}

	var model resource_serverless_traffic_filter.ServerlessTrafficFilterModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := model.Id.ValueString()
	patchReq := toPatchRequest(ctx, model, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	resp, err := r.client.PatchTrafficFilterWithResponse(ctx, id, patchReq)
	if err != nil {
		response.Diagnostics.AddError("Failed to update traffic filter", err.Error())
		return
	}

	if resp.StatusCode() != 200 {
		response.Diagnostics.AddError(
			"Failed to update traffic filter",
			fmt.Sprintf("Unexpected status code: %d, body: %s", resp.StatusCode(), string(resp.Body)),
		)
		return
	}

	model = fromTrafficFilterInfo(ctx, *resp.JSON200, &response.Diagnostics)
	response.Diagnostics.Append(response.State.Set(ctx, model)...)
}

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	if !r.ready(&response.Diagnostics) {
		return
	}

	var model resource_serverless_traffic_filter.ServerlessTrafficFilterModel
	response.Diagnostics.Append(request.State.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := model.Id.ValueString()
	resp, err := r.client.DeleteTrafficFilterWithResponse(ctx, id)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete traffic filter", err.Error())
		return
	}

	if resp.StatusCode() != 200 && resp.StatusCode() != 404 {
		response.Diagnostics.AddError(
			"Failed to delete traffic filter",
			fmt.Sprintf("Unexpected status code: %d, body: %s", resp.StatusCode(), string(resp.Body)),
		)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *Resource) ready(dg *diag.Diagnostics) bool {
	if r.client == nil {
		dg.AddError(
			"Unconfigured API Client",
			"Expected configured API client. Please report this issue to the provider developers.",
		)
		return false
	}
	return true
}
