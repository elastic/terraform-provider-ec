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
	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// UseStateForUnknownInt64OrNull returns a plan modifier that copies a known prior state
// value into the planned value, even if the state is null.
func UseStateForUnknownInt64OrNull() planmodifier.Int64 {
	return useStateForUnknownInt64OrNullModifier{}
}

// useStateForUnknownInt64OrNullModifier implements the plan modifier.
type useStateForUnknownInt64OrNullModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m useStateForUnknownInt64OrNullModifier) Description(_ context.Context) string {
	return "Use state for unknown values, even if state is null."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m useStateForUnknownInt64OrNullModifier) MarkdownDescription(_ context.Context) string {
	return "Use state for unknown values, even if state is null."
}

// PlanModifyInt64 implements the plan modification logic.
func (m useStateForUnknownInt64OrNullModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	// Do nothing if there is no state value and deployment is being created
	deploymentIdDefined, d := AttributeStateDefined(ctx, path.Root("id"), req.State)
	if !deploymentIdDefined && req.StateValue.IsNull() {
		return
	}

	if d.HasError() {
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

	resp.PlanValue = req.StateValue
}
