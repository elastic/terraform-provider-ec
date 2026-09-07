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
	"regexp"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/stackapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"

	"github.com/elastic/terraform-provider-ec/ec/internal"
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
	response.TypeName = request.ProviderTypeName + "_stack"
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

	res, err := stackapi.List(stackapi.ListParams{
		API:    d.client,
		Region: newState.Region.ValueString(),
	})
	if err != nil {
		response.Diagnostics.AddError(
			"Failed retrieving the specified stack version",
			fmt.Sprintf("Failed retrieving the specified stack version: %s", err),
		)
		return
	}

	// Data sources do not receive prior state. An explicit version in config
	// (or an exact version_regex) is the only pin that can survive a yank.
	stack, warning, err := resolveStack(resolveParams{
		Expr:    newState.VersionRegex.ValueString(),
		Version: newState.Version.ValueString(),
		Stacks:  res.Stacks,
	})
	if err != nil {
		response.Diagnostics.AddError(err.Error(), err.Error())
		return
	}
	if warning != "" {
		response.Diagnostics.AddWarning("Stack version is not advertised", warning)
	}

	response.Diagnostics.Append(modelToState(ctx, stack, &newState)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, newState)...)
}

func modelToState(ctx context.Context, stack *models.StackVersionConfig, state *modelV0) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	state.ID = types.StringValue(stack.Version)
	state.Version = types.StringValue(stack.Version)
	if stack.Accessible != nil {
		state.Accessible = types.BoolValue(*stack.Accessible)
	}

	state.MinUpgradableFrom = types.StringValue(stack.MinUpgradableFrom)

	if len(stack.UpgradableTo) > 0 {
		diagnostics.Append(tfsdk.ValueFrom(ctx, stack.UpgradableTo, types.ListType{ElemType: types.StringType}, &state.UpgradableTo)...)
	}

	if stack.Whitelisted != nil {
		state.AllowListed = types.BoolValue(*stack.Whitelisted)
	}

	var diags diag.Diagnostics
	state.Apm, diags = flattenApmConfig(ctx, stack.Apm)
	diagnostics.Append(diags...)

	state.Elasticsearch, diags = flattenElasticsearchConfig(ctx, stack.Elasticsearch)
	diagnostics.Append(diags...)

	state.EnterpriseSearch, diags = flattenEnterpriseSearchConfig(ctx, stack.EnterpriseSearch)
	diagnostics.Append(diags...)

	state.Kibana, diags = flattenKibanaConfig(ctx, stack.Kibana)
	diagnostics.Append(diags...)

	return diagnostics
}

type resolveParams struct {
	Expr    string
	Version string
	Stacks  []*models.StackVersionConfig
}

func resolveStack(p resolveParams) (*models.StackVersionConfig, string, error) {
	pin := p.Version
	if pin == "" {
		if v, ok := exactVersion(p.Expr); ok {
			pin = v
		}
	}

	if pin != "" {
		if advertised, err := stackFromFilters(regexp.QuoteMeta(pin), p.Stacks); err == nil {
			return advertised, "", nil
		}
		return &models.StackVersionConfig{Version: pin}, fmt.Sprintf(
			"%s is not in the advertised stack list; using it anyway. Creating a deployment on a withdrawn version will still fail; keeping an existing deployment on this version is allowed.",
			pin,
		), nil
	}

	stack, err := stackFromFilters(p.Expr, p.Stacks)
	return stack, "", err
}

func stackFromFilters(expr string, stacks []*models.StackVersionConfig) (*models.StackVersionConfig, error) {
	if expr == "latest" {
		if len(stacks) == 0 {
			return nil, fmt.Errorf(`failed to obtain a stack version matching "latest": ` +
				`please specify a valid version_regex`)
		}
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

// exactVersion returns expr if it is a concrete stack version (optionally
// wrapped in ^...$), not a broader regex such as "9.5.?" or "latest".
func exactVersion(expr string) (string, bool) {
	if expr == "" || expr == "latest" {
		return "", false
	}
	trimmed := strings.TrimSuffix(strings.TrimPrefix(expr, "^"), "$")
	if _, err := semver.Parse(trimmed); err != nil {
		return "", false
	}
	re, err := regexp.Compile(expr)
	if err != nil || !re.MatchString(trimmed) {
		return "", false
	}
	return trimmed, true
}

func newElasticsearchConfigModelV0() elasticsearchConfigModelV0 {
	return elasticsearchConfigModelV0{
		DenyList:            types.ListNull(types.StringType),
		CompatibleNodeTypes: types.ListNull(types.StringType),
		Plugins:             types.ListNull(types.StringType),
		DefaultPlugins:      types.ListNull(types.StringType),
	}
}
func newResourceKindConfigModelV0() resourceKindConfigModelV0 {
	return resourceKindConfigModelV0{
		DenyList:            types.ListNull(types.StringType),
		CompatibleNodeTypes: types.ListNull(types.StringType),
	}
}
