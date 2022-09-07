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
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deputil"
	"github.com/elastic/cloud-sdk-go/pkg/models"

	"github.com/elastic/terraform-provider-ec/ec/internal"
	"github.com/elastic/terraform-provider-ec/ec/internal/flatteners"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

var _ provider.DataSourceType = (*DataSourceType)(nil)

type DataSourceType struct{}

func (s DataSourceType) NewDataSource(ctx context.Context, in provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	p, diags := internal.ConvertProviderType(in)

	return &deploymentDataSource{
		p: p,
	}, diags
}

var _ datasource.DataSource = (*deploymentDataSource)(nil)

type deploymentDataSource struct {
	p internal.Provider
}

func (d deploymentDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var newState modelV0
	response.Diagnostics.Append(request.Config.Get(ctx, &newState)...)
	if response.Diagnostics.HasError() {
		return
	}

	res, err := deploymentapi.Get(deploymentapi.GetParams{
		API:          d.p.GetClient(),
		DeploymentID: newState.ID.Value,
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

	response.Diagnostics.Append(modelToState(ctx, res, &newState)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Finally, set the state
	response.Diagnostics.Append(response.State.Set(ctx, newState)...)
}

/*
	TODO - see https://github.com/multani/terraform-provider-camunda/pull/16/files

	Timeouts: &schema.ResourceTimeout{
		Default: schema.DefaultTimeout(5 * time.Minute),
	},
*/

func modelToState(ctx context.Context, res *models.DeploymentGetResponse, state *modelV0) diag.Diagnostics {
	var diags diag.Diagnostics

	state.Name = types.String{Value: *res.Name}
	state.Healthy = types.Bool{Value: *res.Healthy}
	state.Alias = types.String{Value: res.Alias}

	es := res.Resources.Elasticsearch[0]
	if es.Region != nil {
		state.Region = types.String{Value: *es.Region}
	}

	if !util.IsCurrentEsPlanEmpty(es) {
		state.DeploymentTemplateID = types.String{Value: *es.Info.PlanInfo.Current.Plan.DeploymentTemplate.ID}
	}

	diags.Append(flattenTrafficFiltering(ctx, res.Settings, &state.TrafficFilter)...)
	diags.Append(flattenObservability(ctx, res.Settings, &state.Observability)...)
	diags.Append(flattenElasticsearchResources(ctx, res.Resources.Elasticsearch, &state.Elasticsearch)...)
	diags.Append(flattenKibanaResources(ctx, res.Resources.Kibana, &state.Kibana)...)
	diags.Append(flattenApmResources(ctx, res.Resources.Apm, &state.Apm)...)
	diags.Append(flattenIntegrationsServerResources(ctx, res.Resources.IntegrationsServer, &state.IntegrationsServer)...)
	diags.Append(flattenEnterpriseSearchResources(ctx, res.Resources.EnterpriseSearch, &state.EnterpriseSearch)...)

	if res.Metadata != nil {
		state.Tags = flatteners.FlattenTags(res.Metadata.Tags)
	}

	return diags
}
