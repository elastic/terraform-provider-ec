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

package v2

import (
	"context"

	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func UseNullUnlessAddingAPMOrIntegrationsServer() planmodifier.String {
	return useNullUnlessAddingAPMOrIntegrationsServer{}
}

type useNullUnlessAddingAPMOrIntegrationsServer struct{}

var _ planmodifier.String = useNullUnlessAddingAPMOrIntegrationsServer{}

func (m useNullUnlessAddingAPMOrIntegrationsServer) Description(ctx context.Context) string {
	return m.MarkdownDescription(ctx)
}

func (m useNullUnlessAddingAPMOrIntegrationsServer) MarkdownDescription(ctx context.Context) string {
	return "Sets the plan value to null if there is no apm or integrations_server resource"
}

func (m useNullUnlessAddingAPMOrIntegrationsServer) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// if the config is the unknown value, use the unknown value otherwise, interpolation gets messed up
	if req.ConfigValue.IsUnknown() {
		return
	}

	// Critically, we'll return here if this value has been set from state.
	// The rest of this function only applies if there is no value already in state.
	if !req.PlanValue.IsUnknown() {
		return
	}

	addedAPM, diags := wasAttributeAdded(ctx, path.Root("apm"), req.Plan, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	addedIntegrationsServer, diags := wasAttributeAdded(ctx, path.Root("integrations_server"), req.Plan, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if addedAPM || addedIntegrationsServer {
		return
	}

	resp.PlanValue = types.StringNull()
}

func wasAttributeAdded(ctx context.Context, p path.Path, plan tfsdk.Plan, state tfsdk.State) (bool, diag.Diagnostics) {
	hasIntegrationsServer, diags := planmodifiers.HasAttribute(ctx, p, plan)
	if diags.HasError() {
		return false, diags
	}

	if hasIntegrationsServer {
		var value attr.Value
		diags.Append(state.GetAttribute(ctx, p, &value)...)
		if diags.HasError() {
			return false, diags
		}

		// Check if Integrations Server has been enabled, i.e exists in plan, but not in state
		if value.IsNull() {
			return true, diags
		}
	}

	return false, diags
}
