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

	"github.com/elastic/terraform-provider-ec/ec/internal"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_elasticsearch_project"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithModifyPlan = &Resource{}

type Resource struct {
	client serverless.ClientWithResponsesInterface
}

func resourceReady(r *Resource, dg *diag.Diagnostics) bool {
	if r.client == nil {
		dg.AddError(
			"Unconfigured API Client",
			"Expected configured API client. Please report this issue to the provider developers.",
		)

		return false
	}
	return true
}

func (r *Resource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	clients, diags := internal.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	r.client = clients.Serverless
}

func (r *Resource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_elasticsearch_project"
}

func (r *Resource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_elasticsearch_project.ElasticsearchProjectResourceSchema(ctx)
}

func (r Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var cfgModel *resource_elasticsearch_project.ElasticsearchProjectModel
	var planModel *resource_elasticsearch_project.ElasticsearchProjectModel
	var stateModel *resource_elasticsearch_project.ElasticsearchProjectModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfgModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &planModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If state is nil then we're creating, if planModel is nil then we're deleting.
	// There's no need for further modification in either case
	if stateModel == nil || planModel == nil {
		return
	}

	planModel.Credentials = useStateForUnknown(planModel.Credentials, stateModel.Credentials)
	planModel.Endpoints = useStateForUnknown(planModel.Endpoints, stateModel.Endpoints)
	planModel.Metadata = useStateForUnknown(planModel.Metadata, stateModel.Metadata)
	planModel.SearchLake = useStateForUnknown(planModel.SearchLake, stateModel.SearchLake)

	nameHasChanged := !planModel.Name.Equal(stateModel.Name)
	aliasIsConfigured := util.IsKnown(cfgModel.Alias)
	aliasHasChanged := !planModel.Alias.Equal(stateModel.Alias)

	cloudIDIsUnknown := nameHasChanged || aliasHasChanged
	aliasIsUnknown := nameHasChanged && !aliasIsConfigured
	endpointsAreUnknown := aliasHasChanged || (!aliasIsConfigured && nameHasChanged)

	if cloudIDIsUnknown {
		planModel.CloudId = basetypes.NewStringUnknown()
	}

	if aliasIsUnknown {
		planModel.Alias = basetypes.NewStringUnknown()
	}

	if endpointsAreUnknown {
		planModel.Endpoints = resource_elasticsearch_project.NewEndpointsValueUnknown()
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, planModel)...)
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
