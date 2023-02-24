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

// Use current state for a topology's attribute if the topology's state is not nil and the template attribute has not changed
func UseTopologyStateForUnknown(topologyAttributeName string) tfsdk.AttributePlanModifier {
	return useTopologyState{topologyAttributeName: topologyAttributeName}
}

type useTopologyState struct {
	topologyAttributeName string
}

func (m useTopologyState) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
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

	// we check state of entire topology state instead of topology attributes states because nil can be a valid state for some topology attributes
	// e.g. `aws-io-optimized-v2` template doesn't specify `autoscaling_min` for `hot_content` so `min_size`'s state is nil
	topologyStateDefined, diags := attributeStateDefined(ctx, path.Root("elasticsearch").AtName(m.topologyAttributeName), req)

	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	if !topologyStateDefined {
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

func (r useTopologyState) Description(ctx context.Context) string {
	return "Use tier's state if it's defined and template is the same."
}

func (r useTopologyState) MarkdownDescription(ctx context.Context) string {
	return "Use tier's state if it's defined and template is the same."
}

func attributeStateDefined(ctx context.Context, p path.Path, req tfsdk.ModifyAttributePlanRequest) (bool, diag.Diagnostics) {
	var val attr.Value

	if diags := req.State.GetAttribute(ctx, p, &val); diags.HasError() {
		return false, diags
	}

	return !val.IsNull() && !val.IsUnknown(), nil
}
