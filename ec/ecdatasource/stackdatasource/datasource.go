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

package stackdatasource

import (
	"context"
	"fmt"
	"github.com/elastic/cloud-sdk-go/pkg/api/stackapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/terraform-provider-ec/ec/internal"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"regexp"
)

var _ provider.DataSourceType = (*DataSourceType)(nil)

type DataSourceType struct{}

func (s DataSourceType) NewDataSource(ctx context.Context, p provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	return &stackDataSource{
		p: p.(internal.Provider),
	}, nil
}

var _ datasource.DataSource = (*stackDataSource)(nil)

type stackDataSource struct {
	p internal.Provider
}

func (d stackDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var newState modelV0
	response.Diagnostics.Append(request.Config.Get(ctx, &newState)...)
	if response.Diagnostics.HasError() {
		return
	}

	res, err := stackapi.List(stackapi.ListParams{
		API:    d.p.GetClient(),
		Region: newState.Region.Value,
	})
	if err != nil {
		response.Diagnostics.AddError(
			"Failed retrieving the specified stack version",
			fmt.Sprintf("Failed retrieving the specified stack version: %s", err),
		)
		return
	}

	stack, err := stackFromFilters(newState.VersionRegex.Value, newState.Version.Value, newState.Lock.Value, res.Stacks)
	if err != nil {
		response.Diagnostics.AddError(err.Error(), err.Error())
		return
	}

	response.Diagnostics.Append(modelToState(ctx, stack, &newState)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Finally, set the state
	response.Diagnostics.Append(response.State.Set(ctx, newState)...)
}

func modelToState(ctx context.Context, stack *models.StackVersionConfig, state *modelV0) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.String{Value: stack.Version}
	state.Version = types.String{Value: stack.Version}
	if stack.Accessible != nil {
		state.Accessible = types.Bool{Value: *stack.Accessible}
	}

	state.MinUpgradableFrom = types.String{Value: stack.MinUpgradableFrom}

	if len(stack.UpgradableTo) > 0 {
		diags.Append(tfsdk.ValueFrom(ctx, stack.UpgradableTo, types.ListType{ElemType: types.StringType}, &state.UpgradableTo)...)
	}

	if stack.Whitelisted != nil {
		state.AllowListed = types.Bool{Value: *stack.Whitelisted}
	}

	diags.Append(flattenStackVersionApmConfig(ctx, stack.Apm, &state.Apm)...)
	diags.Append(flattenStackVersionElasticsearchConfig(ctx, stack.Elasticsearch, &state.Elasticsearch)...)
	diags.Append(flattenStackVersionEnterpriseSearchConfig(ctx, stack.EnterpriseSearch, &state.EnterpriseSearch)...)
	diags.Append(flattenStackVersionKibanaConfig(ctx, stack.Kibana, &state.Kibana)...)

	return diags
}

/* TODO - see https://github.com/multani/terraform-provider-camunda/pull/16/files
Timeouts: &schema.ResourceTimeout{
	Default: schema.DefaultTimeout(5 * time.Minute),
},
*/

func stackFromFilters(expr, version string, locked bool, stacks []*models.StackVersionConfig) (*models.StackVersionConfig, error) {
	if expr == "latest" && locked && version != "" {
		expr = version
	}

	if expr == "latest" {
		return stacks[0], nil
	}

	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("failed to compile the version_regex: %w", err)
	}

	for _, stack := range stacks {
		if re.MatchString(stack.Version) {
			return stack, nil
		}
	}

	return nil, fmt.Errorf(`failed to obtain a stack version matching "%s": `+
		`please specify a valid version_regex`, expr,
	)
}

func newResourceKindConfigModelV0() resourceKindConfigModelV0 {
	return resourceKindConfigModelV0{
		DenyList:            types.List{ElemType: types.StringType},
		CompatibleNodeTypes: types.List{ElemType: types.StringType},
		Plugins:             types.List{ElemType: types.StringType},
		DefaultPlugins:      types.List{ElemType: types.StringType},
	}
}
