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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

func UseNodeRolesDefault() nodeRolesDefault {
	return nodeRolesDefault{}
}

type nodeRolesDefault struct{}

func (m nodeRolesDefault) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	useState, useNodeRoles, diags := useStateAndNodeRolesInPlanModifiers(ctx, req.ConfigValue, req.Plan, req.State, resp.PlanValue)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !useState {
		return
	}

	// If useNodeRoles is false, we can use the current state and
	// 	it should be null in this case - we don't migrate back from node_roles to node_types
	if !useNodeRoles && !req.StateValue.IsNull() {
		// it should not happen
		return
	}

	// If useNodeRoles is true, then either
	// 	* state already uses node_roles or
	// 	* state uses node_types but we need to migrate to node_roles.
	// We cannot use state in the second case (migration to node_roles)
	// It happens when node_roles state is null.
	if useNodeRoles && req.StateValue.IsNull() {
		return
	}

	resp.PlanValue = req.StateValue
}

// Description returns a human-readable description of the plan modifier.
func (r nodeRolesDefault) Description(ctx context.Context) string {
	return "Use current state if it's still valid."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r nodeRolesDefault) MarkdownDescription(ctx context.Context) string {
	return "Use current state if it's still valid."
}

func SetUnknownOnTopologySizeChange() planmodifier.Set {
	return setUnknownOnTopologyChanges{}
}

type setUnknownOnTopologyChanges struct{}

var (
	tierNames        = []string{"hot", "coordinating", "master", "warm", "cold", "frozen"}
	sizingAttributes = []string{"size", "zone_count"}
)

func (m setUnknownOnTopologyChanges) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if req.PlanValue.IsUnknown() || req.PlanValue.IsNull() {
		return
	}

	for _, tierName := range tierNames {
		var tierValue attr.Value
		resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("elasticsearch").AtName(tierName), &tierValue)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, attrName := range sizingAttributes {
			attrPath := path.Root("elasticsearch").AtName(tierName).AtName(attrName)
			var planValue attr.Value
			resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, attrPath, &planValue)...)
			if resp.Diagnostics.HasError() {
				return
			}

			var stateValue attr.Value

			resp.Diagnostics.Append(req.State.GetAttribute(ctx, attrPath, &stateValue)...)
			if resp.Diagnostics.HasError() {
				return
			}

			// If the plan value is unknown then planmodifiers haven't run for this topology element
			// Eventually the plan value will be set to the state value and it will be unchanged.
			// The tier should be directly checked for unknown, since the planValue will be null in that case (instead of unknown).
			// See: https://github.com/hashicorp/terraform-plugin-framework/issues/186
			if (planValue.IsUnknown() || tierValue.IsUnknown()) && !stateValue.IsUnknown() && !stateValue.IsNull() {
				continue
			}

			hasChanged, diags := planmodifiers.AttributeChanged(ctx, attrPath, req.Plan, req.State)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			if hasChanged {
				resp.PlanValue = types.SetUnknown(types.StringType)
				return
			}
		}
	}
}

// Description returns a human-readable description of the plan modifier.
func (r setUnknownOnTopologyChanges) Description(ctx context.Context) string {
	return "Sets the plan value to unknown if the size of any topology element has changed."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r setUnknownOnTopologyChanges) MarkdownDescription(ctx context.Context) string {
	return "Sets the plan value to unknown if the size of any topology element has changed."
}
