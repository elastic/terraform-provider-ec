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

package deploymentdatasource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deputil"
	"github.com/elastic/cloud-sdk-go/pkg/models"

	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource"
	"github.com/elastic/terraform-provider-ec/ec/internal"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

var _ datasource.DataSource = &DataSource{}
var _ datasource.DataSourceWithConfigure = &DataSource{}

type DataSource struct {
	client *api.API
}

func (d *DataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	clients, diags := internal.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	d.client = clients.Stateful
}

func (d *DataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_deployment"
}

func (d DataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	// Prevent panic if the provider has not been configured.
	if d.client == nil {
		response.Diagnostics.AddError(
			"Unconfigured API Client",
			"Expected configured API client. Please report this issue to the provider developers.",
		)

		return
	}

	var newState modelV0
	response.Diagnostics.Append(request.Config.Get(ctx, &newState)...)
	if response.Diagnostics.HasError() {
		return
	}

	res, err := deploymentapi.Get(deploymentapi.GetParams{
		API:          d.client,
		DeploymentID: newState.ID.ValueString(),
		QueryParams: deputil.QueryParams{
			ShowPlans:        true,
			ShowSettings:     true,
			ShowMetadata:     true,
			ShowPlanDefaults: true,
		},
	})
	if err != nil {
		response.Diagnostics.AddError(
			"Failed retrieving deployment information",
			fmt.Sprintf("Failed retrieving deployment information: %s", err),
		)
		return
	}

	if !deploymentresource.HasRunningResources(res) {
		return
	}

	response.Diagnostics.Append(modelToState(ctx, res, &newState)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Finally, set the state
	response.Diagnostics.Append(response.State.Set(ctx, newState)...)
}

func modelToState(ctx context.Context, res *models.DeploymentGetResponse, state *modelV0) diag.Diagnostics {
	var diagsnostics diag.Diagnostics

	state.Name = types.StringValue(*res.Name)
	state.Healthy = types.BoolValue(*res.Healthy)
	state.Alias = types.StringValue(res.Alias)

	es := res.Resources.Elasticsearch[0]
	if es.Region != nil {
		state.Region = types.StringValue(*es.Region)
	}

	if !util.IsCurrentEsPlanEmpty(es) {
		state.DeploymentTemplateID = types.StringValue(*es.Info.PlanInfo.Current.Plan.DeploymentTemplate.ID)
	}

	var diags diag.Diagnostics

	state.TrafficFilter, diags = flattenTrafficFiltering(ctx, res.Settings)
	diagsnostics.Append(diags...)

	state.Observability, diags = flattenObservability(ctx, res.Settings)
	diagsnostics.Append(diags...)

	state.Elasticsearch, diags = flattenElasticsearchResources(ctx, res.Resources.Elasticsearch)
	diagsnostics.Append(diags...)

	state.Kibana, diags = flattenKibanaResources(ctx, res.Resources.Kibana)
	diagsnostics.Append(diags...)

	state.Apm, diags = flattenApmResources(ctx, res.Resources.Apm)
	diagsnostics.Append(diags...)

	state.IntegrationsServer, diags = flattenIntegrationsServerResources(ctx, res.Resources.IntegrationsServer)
	diagsnostics.Append(diags...)

	state.EnterpriseSearch, diags = flattenEnterpriseSearchResources(ctx, res.Resources.EnterpriseSearch)
	diagsnostics.Append(diags...)

	if res.Metadata != nil {
		state.Tags, diags = converters.ModelsTagsToTypesMap(res.Metadata.Tags)
		diagsnostics.Append(diags...)
	}

	return diagsnostics
}
