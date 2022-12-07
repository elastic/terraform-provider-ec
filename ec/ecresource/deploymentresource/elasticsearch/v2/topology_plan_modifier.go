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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func UseTierStateForUnknown(tier string) tfsdk.AttributePlanModifier {
	return useTierState{tier: tier}
}

type useTierState struct {
	tier string
}

func (m useTierState) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
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

	// we check tier's state instead of tier attribute's state because nil can be a valid state
	// e.g. `aws-io-optimized-v2` template doesn't specify `autoscaling_min` for `hot_content` so `min_size` state is nil
	tierStateDefined, diags := attributeStateDefined(ctx, path.Root("elasticsearch").AtName(m.tier), req)

	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	if !tierStateDefined {
		return
	}

	templateChanged, diags := attributeChanged(ctx, path.Root("deployment_template_id"), req)

	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	if templateChanged {
		return
	}

	resp.AttributePlan = req.AttributeState
}

func (r useTierState) Description(ctx context.Context) string {
	return "Use tier's state if it's defined and template is the same."
}

func (r useTierState) MarkdownDescription(ctx context.Context) string {
	return "Use tier's state if it's defined and template is the same."
}

func attributeChanged(ctx context.Context, p path.Path, req tfsdk.ModifyAttributePlanRequest) (bool, diag.Diagnostics) {
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

func attributeStateDefined(ctx context.Context, p path.Path, req tfsdk.ModifyAttributePlanRequest) (bool, diag.Diagnostics) {
	var val attr.Value

	if diags := req.State.GetAttribute(ctx, p, &val); diags.HasError() {
		return false, diags
	}

	return !val.IsNull() && !val.IsUnknown(), nil
}
