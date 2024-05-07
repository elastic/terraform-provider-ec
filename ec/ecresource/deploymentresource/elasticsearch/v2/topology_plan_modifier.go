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
)

// UseTopologyStateForUnknown Use current state for a topology's attribute, unless one of the following scenarios occurs:
//  1. The deployment template attribute has changed
//  2. `migrate_to_latest_hardware` is set to `true` and there is a migration available to be performed
func UseTopologyStateForUnknown(topologyAttributeName string) useTopologyState {
	return useTopologyState{topologyAttributeName: topologyAttributeName}
}

type useTopologyState struct {
	topologyAttributeName string
}

func (m useTopologyState) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	useState, diags := m.UseState(ctx, req.ConfigValue, req.Plan, req.State, resp.PlanValue)
	resp.Diagnostics.Append(diags...)
	if useState {
		resp.PlanValue = req.StateValue
	}
}

func (m useTopologyState) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	useState, diags := m.UseState(ctx, req.ConfigValue, req.Plan, req.State, resp.PlanValue)
	resp.Diagnostics.Append(diags...)
	if useState {
		resp.PlanValue = req.StateValue
	}
}

func (m useTopologyState) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	useState, diags := m.UseState(ctx, req.ConfigValue, req.Plan, req.State, resp.PlanValue)
	resp.Diagnostics.Append(diags...)
	if useState {
		resp.PlanValue = req.StateValue
	}
}

func (m useTopologyState) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	useState, diags := m.UseState(ctx, req.ConfigValue, req.Plan, req.State, resp.PlanValue)
	resp.Diagnostics.Append(diags...)
	if useState {
		resp.PlanValue = req.StateValue
	}
}

type PlanModifierResponse interface {
	planmodifier.StringResponse | planmodifier.Int64Response
}

func (m useTopologyState) UseState(ctx context.Context, configValue attr.Value, plan tfsdk.Plan, state tfsdk.State, planValue attr.Value) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !planValue.IsUnknown() {
		return false, nil
	}

	if configValue.IsUnknown() {
		return false, nil
	}

	// we check state of entire topology state instead of topology attributes states because nil can be a valid state for some topology attributes
	// e.g. `aws-io-optimized-v2` template doesn't specify `autoscaling_min` for `hot_content` so `min_size`'s state is nil
	topologyStateDefined, d := planmodifiers.AttributeStateDefined(ctx, path.Root("elasticsearch").AtName(m.topologyAttributeName), state)
	diags.Append(d...)

	if diags.HasError() {
		return false, diags
	}

	if !topologyStateDefined {
		return false, diags
	}

	templateChanged, d := planmodifiers.AttributeChanged(ctx, path.Root("deployment_template_id"), plan, state)
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

	isMigrationAvailable, d := planmodifiers.CheckAvailableMigration(ctx, plan, state, path.Root("elasticsearch").AtName(m.topologyAttributeName))
	diags.Append(d...)

	if diags.HasError() {
		return false, diags
	}

	if isMigrationAvailable {
		return false, diags
	}

	return true, diags
}

func (r useTopologyState) Description(ctx context.Context) string {
	return "Use tier's state if it's defined and template is the same."
}

func (r useTopologyState) MarkdownDescription(ctx context.Context) string {
	return "Use tier's state if it's defined and template is the same."
}
