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

	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Use `self` as value of `observability`'s `deployment_id` attribute
func UseNodeTypesDefault() tfsdk.AttributePlanModifier {
	return nodeTypesDefault{}
}

type nodeTypesDefault struct{}

func (r nodeTypesDefault) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if req.AttributeState == nil || resp.AttributePlan == nil || req.AttributeConfig == nil {
		return
	}

	if !resp.AttributePlan.IsUnknown() {
		return
	}

	// if the config is the unknown value, use the unknown value otherwise, interpolation gets messed up
	if req.AttributeConfig.IsUnknown() {
		return
	}

	// if there is no state for "version" return
	var stateVersion types.String

	if diags := req.State.GetAttribute(ctx, path.Root("version"), &stateVersion); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if stateVersion.IsNull() {
		return
	}

	// if template changed return
	templateChanged, diags := isAttributeChanged(ctx, path.Root("deployment_template_id"), req)

	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	if templateChanged {
		return
	}

	// get version for plan and state and calculate useNodeRoles

	var planVersion types.String

	if diags := req.Plan.GetAttribute(ctx, path.Root("version"), &planVersion); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	useNodeRoles, diags := utils.UseNodeRoles(stateVersion, planVersion)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if useNodeRoles && !req.AttributeState.IsNull() {
		return
	}

	resp.AttributePlan = req.AttributeState
}

// Description returns a human-readable description of the plan modifier.
func (r nodeTypesDefault) Description(ctx context.Context) string {
	return "Use current state if it's still valid."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r nodeTypesDefault) MarkdownDescription(ctx context.Context) string {
	return "Use current state if it's still valid."
}

func isAttributeChanged(ctx context.Context, p path.Path, req tfsdk.ModifyAttributePlanRequest) (bool, diag.Diagnostics) {
	var planValue attr.Value

	if diags := req.Plan.GetAttribute(ctx, p, &planValue); diags.HasError() {
		return false, diags
	}

	var stateValue attr.Value

	if diags := req.State.GetAttribute(ctx, p, &stateValue); diags.HasError() {
		return false, diags
	}

	return !planValue.Equal(stateValue), nil
}
