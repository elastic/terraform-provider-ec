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
	"strconv"

	"github.com/elastic/terraform-provider-ec/ec/internal"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var _ provider.DataSourceType = (*DataSourceType)(nil)

type DataSourceType struct{}

func (s DataSourceType) NewDataSource(ctx context.Context, in provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	p, diags := internal.ConvertProviderType(in)

	return &deploymentsDataSource{
		p: p,
	}, diags
}

var _ datasource.DataSource = (*deploymentsDataSource)(nil)

type deploymentsDataSource struct {
	p internal.Provider
}

func (d deploymentsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
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
		API:     d.p.GetClient(),
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

/* TODO - see https://github.com/multani/terraform-provider-camunda/pull/16/files
Timeouts: &schema.ResourceTimeout{
	Default: schema.DefaultTimeout(5 * time.Minute),
},
*/

func modelToState(ctx context.Context, res *models.DeploymentsSearchResponse, state *modelV0) diag.Diagnostics {
	var diags diag.Diagnostics

	if b, _ := res.MarshalBinary(); len(b) > 0 {
		state.ID = types.String{Value: strconv.Itoa(schema.HashString(string(b)))}
	}
	state.ReturnCount = types.Int64{Value: int64(*res.ReturnCount)}

	var result = make([]deploymentModelV0, 0, len(res.Deployments))
	for _, deployment := range res.Deployments {
		var m deploymentModelV0

		m.DeploymentID = types.String{Value: *deployment.ID}
		m.Alias = types.String{Value: deployment.Alias}

		if deployment.Name != nil {
			m.Name = types.String{Value: *deployment.Name}
		}

		if len(deployment.Resources.Elasticsearch) > 0 {
			m.ElasticSearchResourceID = types.String{Value: *deployment.Resources.Elasticsearch[0].ID}
			m.ElasticSearchRefID = types.String{Value: *deployment.Resources.Elasticsearch[0].RefID}
		}

		if len(deployment.Resources.Kibana) > 0 {
			m.KibanaResourceID = types.String{Value: *deployment.Resources.Kibana[0].ID}
			m.KibanaRefID = types.String{Value: *deployment.Resources.Kibana[0].RefID}
		}

		if len(deployment.Resources.Apm) > 0 {
			m.ApmResourceID = types.String{Value: *deployment.Resources.Apm[0].ID}
			m.ApmRefID = types.String{Value: *deployment.Resources.Apm[0].RefID}
		}

		if len(deployment.Resources.IntegrationsServer) > 0 {
			m.IntegrationsServerResourceID = types.String{Value: *deployment.Resources.IntegrationsServer[0].ID}
			m.IntegrationsServerRefID = types.String{Value: *deployment.Resources.IntegrationsServer[0].RefID}
		}

		if len(deployment.Resources.EnterpriseSearch) > 0 {
			m.EnterpriseSearchResourceID = types.String{Value: *deployment.Resources.EnterpriseSearch[0].ID}
			m.EnterpriseSearchRefID = types.String{Value: *deployment.Resources.EnterpriseSearch[0].RefID}
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
