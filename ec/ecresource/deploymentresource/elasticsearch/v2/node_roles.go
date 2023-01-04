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
	"fmt"

	"github.com/blang/semver"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func CompatibleWithNodeRoles(version string) (bool, error) {
	deploymentVersion, err := semver.Parse(version)
	if err != nil {
		return false, fmt.Errorf("failed to parse Elasticsearch version: %w", err)
	}

	return deploymentVersion.GE(utils.DataTiersVersion), nil
}

func UseNodeRoles(ctx context.Context, stateVersion, planVersion types.String, planElasticsearch types.Object) (bool, diag.Diagnostics) {
	compatibleWithNodeRoles, err := CompatibleWithNodeRoles(planVersion.Value)

	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Failed to determine whether to use node_roles", err.Error())
		return false, diags
	}

	if !compatibleWithNodeRoles {
		return false, nil
	}

	convertLegacy, diags := legacyToNodeRoles(ctx, stateVersion, planVersion, planElasticsearch)

	if diags.HasError() {
		return false, diags
	}

	return convertLegacy, nil
}

// legacyToNodeRoles returns true when the legacy  "node_type_*" should be
// migrated over to node_roles. Which will be true when:
// * The version field doesn't change.
// * The version field changes but:
//   - The Elasticsearch.0.toplogy doesn't have any node_type_* set.
func legacyToNodeRoles(ctx context.Context, stateVersion, planVersion types.String, planElasticsearch types.Object) (bool, diag.Diagnostics) {
	if stateVersion.Value == "" || stateVersion.Value == planVersion.Value {
		return true, nil
	}

	var diags diag.Diagnostics
	oldVersion, err := semver.Parse(stateVersion.Value)
	if err != nil {
		diags.AddError("Failed to parse previous Elasticsearch version", err.Error())
		return false, diags
	}
	newVersion, err := semver.Parse(planVersion.Value)
	if err != nil {
		diags.AddError("Failed to parse new Elasticsearch version", err.Error())
		return false, diags
	}

	// if the version change moves from non-node_roles to one
	// that supports node roles, do not migrate on that step.
	if oldVersion.LT(utils.DataTiersVersion) && newVersion.GE(utils.DataTiersVersion) {
		return false, nil
	}

	// When any topology elements in the state have the node_type_*
	// properties set, the node_role field cannot be used, since
	// we'd be changing the version AND migrating over `node_role`s
	// which is not permitted by the API.

	var es *ElasticsearchTF

	if diags := tfsdk.ValueAs(ctx, planElasticsearch, &es); diags.HasError() {
		return false, diags
	}

	if es == nil {
		diags.AddError("Cannot migrate node types to node roles", "cannot find elasticsearch object")
		return false, diags
	}

	tiers, diags := es.topologies(ctx)

	if diags.HasError() {
		return false, diags
	}

	for _, tier := range tiers {
		if tier != nil && tier.HasNodeType() {
			return false, nil
		}
	}

	return true, nil
}

func useStateAndNodeRolesInPlanModifiers(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) (useState, useNodeRoles bool) {
	if req.AttributeState == nil || resp.AttributePlan == nil || req.AttributeConfig == nil {
		return false, false
	}

	if !resp.AttributePlan.IsUnknown() {
		return false, false
	}

	// if the config is the unknown value, use the unknown value otherwise, interpolation gets messed up
	// it's the precaution taken from the Framework's `UseStateForUnknown` plan modifier
	if req.AttributeConfig.IsUnknown() {
		return false, false
	}

	// if there is no state for "version" return
	var stateVersion types.String

	if diags := req.State.GetAttribute(ctx, path.Root("version"), &stateVersion); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return false, false
	}

	if stateVersion.IsNull() {
		return false, false
	}

	// if template changed return
	templateChanged, diags := isAttributeChanged(ctx, path.Root("deployment_template_id"), req)

	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return false, false
	}

	if templateChanged {
		return false, false
	}

	// get version for plan and state and calculate useNodeRoles

	var planVersion types.String

	if diags := req.Plan.GetAttribute(ctx, path.Root("version"), &planVersion); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return false, false
	}

	var elasticsearch types.Object

	if diags := req.Plan.GetAttribute(ctx, path.Root("elasticsearch"), &elasticsearch); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return false, false
	}

	useNodeRoles, diags = UseNodeRoles(ctx, stateVersion, planVersion, elasticsearch)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return false, false
	}

	return true, useNodeRoles
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
