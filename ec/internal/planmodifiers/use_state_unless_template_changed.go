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

package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// UseStateForUnknownUnlessMigrationIsRequired Use current state for a topology's attribute, unless one of the following scenarios occurs:
//  1. The attribute is not nullable (`isNullable = false`) and the topology's state is nil
//  2. The deployment template attribute has changed
//  3. `migrate_to_latest_hardware` is set to `true` and there is a migration available to be performed
//  4. The state of the parent attribute is nil
func UseStateForUnknownUnlessMigrationIsRequired(resourceKind string, isNullable bool) useStateForUnknownUnlessMigrationIsRequired {
	return useStateForUnknownUnlessMigrationIsRequired{resourceKind: resourceKind, isNullable: isNullable}
}

type useStateForUnknownUnlessMigrationIsRequired struct {
	resourceKind string
	isNullable   bool
}

type PlanModifierResponse interface {
	planmodifier.StringResponse | planmodifier.Int64Response
}

func (m useStateForUnknownUnlessMigrationIsRequired) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	useState, diags := m.UseState(ctx, req.ConfigValue, req.Plan, req.State, resp.PlanValue, req.StateValue)
	resp.Diagnostics.Append(diags...)
	if useState {
		resp.PlanValue = req.StateValue
	}
}

func (m useStateForUnknownUnlessMigrationIsRequired) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	useState, diags := m.UseState(ctx, req.ConfigValue, req.Plan, req.State, resp.PlanValue, req.StateValue)
	resp.Diagnostics.Append(diags...)
	if useState {
		resp.PlanValue = req.StateValue
	}
}

func (m useStateForUnknownUnlessMigrationIsRequired) UseState(ctx context.Context, configValue attr.Value, plan tfsdk.Plan, state tfsdk.State, planValue attr.Value, stateValue attr.Value) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	var parentResState attr.Value
	if d := state.GetAttribute(ctx, path.Root(m.resourceKind), &parentResState); d.HasError() {
		return false, d
	}

	resourceIsBeingCreated := parentResState.IsNull()

	if resourceIsBeingCreated {
		return false, nil
	}

	if stateValue.IsNull() && !m.isNullable {
		return false, nil
	}

	if !planValue.IsUnknown() {
		return false, nil
	}

	// if the config is the unknown value, use the unknown value otherwise, interpolation gets messed up
	if configValue.IsUnknown() {
		return false, nil
	}

	templateChanged, d := AttributeChanged(ctx, path.Root("deployment_template_id"), plan, state)
	diags.Append(d...)

	// If template changed, we won't use state
	if templateChanged {
		return false, diags
	}

	var migrateToLatestHw bool
	plan.GetAttribute(ctx, path.Root("migrate_to_latest_hardware"), &migrateToLatestHw)

	// If migrate_to_latest_hardware isn't set, we want to use state
	if !migrateToLatestHw {
		return true, diags
	}

	isMigrationAvailable, d := CheckAvailableMigration(ctx, plan, state, path.Root(m.resourceKind))
	diags.Append(d...)

	if diags.HasError() {
		return false, diags
	}

	if isMigrationAvailable {
		return false, diags
	}

	return true, diags
}

func (r useStateForUnknownUnlessMigrationIsRequired) Description(ctx context.Context) string {
	return "Use tier's state if it's defined and template is the same."
}

func (r useStateForUnknownUnlessMigrationIsRequired) MarkdownDescription(ctx context.Context) string {
	return "Use tier's state if it's defined and template is the same."
}

func diffStateAttributes(ctx context.Context, p1 path.Path, p2 path.Path, state tfsdk.State) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	var p1Value attr.Value

	d1 := state.GetAttribute(ctx, p1, &p1Value)

	diags.Append(d1...)

	var p2Value attr.Value

	d2 := state.GetAttribute(ctx, p2, &p2Value)

	diags.Append(d2...)

	return !p1Value.Equal(p2Value), diags
}

func attributePlanDefined(ctx context.Context, p path.Path, plan tfsdk.Plan) (bool, diag.Diagnostics) {
	var value attr.Value

	diags := plan.GetAttribute(ctx, p, &value)

	return !value.IsNull() && !value.IsUnknown(), diags
}

func CheckAvailableMigration(ctx context.Context, plan tfsdk.Plan, state tfsdk.State, topologyPath path.Path) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	planHasInstanceConfigId, d := attributePlanDefined(ctx, topologyPath.AtName("instance_configuration_id"), plan)

	diags.Append(d...)

	planHasInstanceConfigVersion, d := attributePlanDefined(ctx, topologyPath.AtName("instance_configuration_version"), plan)

	diags.Append(d...)

	// We won't migrate this topology element if 'instance_configuration_id' or 'instance_configuration_version' are
	// defined on the TF configuration. Otherwise, we may be setting an incorrect value for 'size', in case the
	// template IC has different size increments
	if planHasInstanceConfigId || planHasInstanceConfigVersion {
		return false, diags
	}

	instanceConfigIdsDiff, d := diffStateAttributes(ctx, topologyPath.AtName("instance_configuration_id"), topologyPath.AtName("latest_instance_configuration_id"), state)

	diags.Append(d...)

	instanceConfigVersionsDiff, d := diffStateAttributes(ctx, topologyPath.AtName("instance_configuration_version"), topologyPath.AtName("latest_instance_configuration_version"), state)

	diags.Append(d...)

	// We consider that a migration is available when:
	//    * the current instance config ID doesn't match the one in the template
	//    * the instance config IDs match but the instance config versions differ
	return instanceConfigIdsDiff || instanceConfigVersionsDiff, diags
}
