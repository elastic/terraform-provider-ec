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

package projectresource

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-ec/ec/internal"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_elasticsearch_project"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &Resource[resource_elasticsearch_project.ElasticsearchProjectModel]{}
var _ resource.ResourceWithConfigure = &Resource[resource_elasticsearch_project.ElasticsearchProjectModel]{}
var _ resource.ResourceWithModifyPlan = &Resource[resource_elasticsearch_project.ElasticsearchProjectModel]{}

type Resource[T any] struct {
	modelReader modelHandler[T]
	api         api[T]
	name        string
}

type modelGetter interface {
	Get(ctx context.Context, target interface{}) diag.Diagnostics
}

type modelHandler[T any] interface {
	readFrom(context.Context, modelGetter) (*T, diag.Diagnostics)
	getID(T) string
	modify(T, T, T) T
}

type api[TModel any] interface {
	create(context.Context, TModel) (TModel, diag.Diagnostics)
	patch(context.Context, TModel) diag.Diagnostics
	ensureInitialised(context.Context, TModel) diag.Diagnostics
	read(context.Context, string, TModel) (bool, TModel, diag.Diagnostics)
	delete(context.Context, TModel) diag.Diagnostics
	withClient(serverless.ClientWithResponsesInterface) api[TModel]
	ready() bool
}

func resourceReady[T any](r *Resource[T], dg *diag.Diagnostics) bool {
	if !r.api.ready() {
		dg.AddError(
			"Unconfigured API Client",
			"Expected configured API client. Please report this issue to the provider developers.",
		)

		return false
	}
	return true
}

func (r *Resource[T]) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	clients, diags := internal.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	r.api = r.api.withClient(clients.Serverless)
}

func (r *Resource[T]) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = fmt.Sprintf("%s_%s_project", request.ProviderTypeName, r.name)
}

func (r *Resource[T]) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_elasticsearch_project.ElasticsearchProjectResourceSchema(ctx)
}

func (r Resource[T]) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	cfgModel, diags := r.modelReader.readFrom(ctx, req.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	planModel, diags := r.modelReader.readFrom(ctx, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateModel, diags := r.modelReader.readFrom(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If state is nil then we're creating, if planModel is nil then we're deleting.
	// There's no need for further modification in either case
	if stateModel == nil || planModel == nil {
		return
	}

	modifiedModel := r.modelReader.modify(*planModel, *stateModel, *cfgModel)
	resp.Diagnostics.Append(resp.Plan.Set(ctx, modifiedModel)...)
}

func useStateForUnknown[T basetypes.ObjectValuable](planValue T, stateValue T) T {
	if stateValue.IsNull() || stateValue.IsUnknown() {
		return planValue
	}

	if planValue.IsUnknown() {
		return stateValue
	}

	return planValue
}
