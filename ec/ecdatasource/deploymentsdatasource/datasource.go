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

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"

	"github.com/elastic/terraform-provider-ec/ec/internal"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

var _ datasource.DataSource = &DataSource{}
var _ datasource.DataSourceWithConfigure = &DataSource{}
var _ datasource.DataSourceWithConfigValidators = &DataSource{}

type DataSource struct {
	client *api.API
}

func (d *DataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	client, diags := internal.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	d.client = client
}

func (d *DataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		// Only one of name_prefix and name should be configured
		datasourcevalidator.Conflicting(
			path.MatchRoot("name_prefix"),
			path.MatchRoot("name"),
		),
	}
}

func (d *DataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_deployments"
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

	query, diags := expandFilters(ctx, newState)
	response.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	res, err := deploymentapi.Search(deploymentapi.SearchParams{
		API:     d.client,
		Request: query,
	})
	if err != nil {
		response.Diagnostics.AddError(
			"Failed searching deployments",
			fmt.Sprintf("Failed searching deployments version: %s", err),
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

func modelToState(ctx context.Context, res *models.DeploymentsSearchResponse, state *modelV0) diag.Diagnostics {
	var diags diag.Diagnostics

	if b, _ := res.MarshalBinary(); len(b) > 0 {
		state.ID = types.StringValue(util.HashString(string(b)))
	}
	state.ReturnCount = types.Int64Value(int64(*res.ReturnCount))

	var result = make([]deploymentModelV0, 0, len(res.Deployments))
	for _, deployment := range res.Deployments {
		var m deploymentModelV0

		m.DeploymentID = types.StringValue(*deployment.ID)
		m.Alias = types.StringValue(deployment.Alias)

		if deployment.Name != nil {
			m.Name = types.StringValue(*deployment.Name)
		}

		if len(deployment.Resources.Elasticsearch) > 0 {
			m.ElasticsearchResourceID = types.StringValue(*deployment.Resources.Elasticsearch[0].ID)
			m.ElasticsearchRefID = types.StringValue(*deployment.Resources.Elasticsearch[0].RefID)
		}

		if len(deployment.Resources.Kibana) > 0 {
			m.KibanaResourceID = types.StringValue(*deployment.Resources.Kibana[0].ID)
			m.KibanaRefID = types.StringValue(*deployment.Resources.Kibana[0].RefID)
		}

		if len(deployment.Resources.Apm) > 0 {
			m.ApmResourceID = types.StringValue(*deployment.Resources.Apm[0].ID)
			m.ApmRefID = types.StringValue(*deployment.Resources.Apm[0].RefID)
		}

		if len(deployment.Resources.IntegrationsServer) > 0 {
			m.IntegrationsServerResourceID = types.StringValue(*deployment.Resources.IntegrationsServer[0].ID)
			m.IntegrationsServerRefID = types.StringValue(*deployment.Resources.IntegrationsServer[0].RefID)
		}

		if len(deployment.Resources.EnterpriseSearch) > 0 {
			m.EnterpriseSearchResourceID = types.StringValue(*deployment.Resources.EnterpriseSearch[0].ID)
			m.EnterpriseSearchRefID = types.StringValue(*deployment.Resources.EnterpriseSearch[0].RefID)
		}

		result = append(result, m)

	}

	diags.Append(tfsdk.ValueFrom(ctx, result, types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: deploymentAttrTypes(),
		},
	}, &state.Deployments)...)

	return diags
}
