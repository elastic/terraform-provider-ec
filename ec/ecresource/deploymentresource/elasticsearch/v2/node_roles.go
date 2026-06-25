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

	"github.com/blang/semver/v4"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/utils"
	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifiers"
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
	compatibleWithNodeRoles, err := CompatibleWithNodeRoles(planVersion.ValueString())

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
	if stateVersion.ValueString() == "" || stateVersion.ValueString() == planVersion.ValueString() {
		return true, nil
	}

	var diags diag.Diagnostics
	oldVersion, err := semver.Parse(stateVersion.ValueString())
	if err != nil {
		diags.AddError("Failed to parse previous Elasticsearch version", err.Error())
		return false, diags
	}
	newVersion, err := semver.Parse(planVersion.ValueString())
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
	hasNodeTypes, d := PlanHasNodeTypes(ctx, planElasticsearch)

	diags.Append(d...)

	return !hasNodeTypes, d
}

func PlanHasNodeTypes(ctx context.Context, planElasticsearch types.Object) (bool, diag.Diagnostics) {
	var es *ElasticsearchTF

	diags := tfsdk.ValueAs(ctx, planElasticsearch, &es)

	if diags.HasError() {
		return false, diags
	}

	if es == nil {
		diags.AddError("Cannot determine if node types are defined", "cannot find elasticsearch object")
		return false, diags
	}

	tiers, diags := es.topologies(ctx)

	if diags.HasError() {
		return false, diags
	}

	for _, tier := range tiers {
		if tier != nil && tier.HasNodeTypes() {
			return true, nil
		}
	}

	return false, nil
}

// if useState is false, useNodeRoles is always false
func useStateAndNodeRolesInPlanModifiers(ctx context.Context, configValue attr.Value, plan tfsdk.Plan, state tfsdk.State, planValue attr.Value) (useState bool, useNodeRoles bool, diags diag.Diagnostics) {
	if !planValue.IsUnknown() {
		return false, false, nil
	}

	if configValue.IsUnknown() {
		return false, false, nil
	}

	var stateVersion types.String

	if diags := state.GetAttribute(ctx, path.Root("version"), &stateVersion); diags.HasError() {
		return false, false, diags
	}

	// If resource has state, then it should contain version.
	// So if there is no version in state, plan modifier is called for Create.
	// In that case there is no state to use.
	// We cannot use StateValue from request parameter for this purpose,
	// because null can be a valid state for node_roles and node_types in Update.
	// E.g. node_roles' state can be null if node_types are used.
	if stateVersion.IsNull() {
		return false, false, nil
	}

	// if template changed return
	templateChanged, diags := planmodifiers.AttributeChanged(ctx, path.Root("deployment_template_id"), plan, state)
	if diags.HasError() {
		return false, false, diags
	}

	if templateChanged {
		return false, false, nil
	}

	var planVersion types.String

	if diags := plan.GetAttribute(ctx, path.Root("version"), &planVersion); diags.HasError() {
		return false, false, diags
	}

	var elasticsearch types.Object

	if diags := plan.GetAttribute(ctx, path.Root("elasticsearch"), &elasticsearch); diags.HasError() {
		return false, false, diags
	}

	if useNodeRoles, diags = UseNodeRoles(ctx, stateVersion, planVersion, elasticsearch); diags.HasError() {
		return false, false, diags
	}

	return true, useNodeRoles, nil
}
