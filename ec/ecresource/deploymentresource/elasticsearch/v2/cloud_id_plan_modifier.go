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

// Use current state for instead of unknown, unless the deployment name change or Kibana is being enabled/disabled
func UseStateForUnknownUnlessNameOrKibanaStateChanges() planmodifier.String {
	return useStateForUnknownUnlessNameOrKibanaStateChanges{}
}

type useStateForUnknownUnlessNameOrKibanaStateChanges struct{}

func (m useStateForUnknownUnlessNameOrKibanaStateChanges) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	// Do nothing if the deployment name has changed
	nameChanged, diags := planmodifiers.AttributeChanged(ctx, path.Root("name"), req.Plan, req.State)
	if resp.Diagnostics.Append(diags...); diags.HasError() {
		return
	}
	if nameChanged {
		return
	}

	kibanaChanged, diags := kibanaStateChanging(ctx, req.Plan, req.State)
	if resp.Diagnostics.Append(diags...); diags.HasError() {
		return
	}
	if kibanaChanged {
		return
	}

	resp.PlanValue = req.StateValue
}

func (r useStateForUnknownUnlessNameOrKibanaStateChanges) Description(ctx context.Context) string {
	return "Use current state for instead of unknown, unless the deployment name change or Kibana is being enabled/disabled."
}

func (r useStateForUnknownUnlessNameOrKibanaStateChanges) MarkdownDescription(ctx context.Context) string {
	return "Use current state for instead of unknown, unless the deployment name change or Kibana is being enabled/disabled."
}

func kibanaStateChanging(ctx context.Context, plan tfsdk.Plan, state tfsdk.State) (bool, diag.Diagnostics) {
	var planValue attr.Value
	p := path.Root("kibana")

	if diags := plan.GetAttribute(ctx, p, &planValue); diags.HasError() {
		return false, diags
	}

	var stateValue attr.Value

	if diags := state.GetAttribute(ctx, p, &stateValue); diags.HasError() {
		return false, diags
	}

	return planValue.IsNull() != stateValue.IsNull(), nil
}
