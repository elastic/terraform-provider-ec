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

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UseStateForUnknownOrUnknown copies the prior state value when the planned
// value is unknown, matching the behaviour of
// stringplanmodifier.UseStateForUnknown. If there is no prior state value
// (for example, a new entry in a map), it marks the plan value as unknown so
// that Terraform accepts a value computed by the provider during apply.
type useStateForUnknownOrUnknown struct{}

func UseStateForUnknownOrUnknown() planmodifier.String {
	return useStateForUnknownOrUnknown{}
}

func (m useStateForUnknownOrUnknown) Description(ctx context.Context) string {
	return "Use prior state when unknown, otherwise leave the value unknown."
}

func (m useStateForUnknownOrUnknown) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m useStateForUnknownOrUnknown) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If the configuration explicitly sets a value, keep it.
	if !req.ConfigValue.IsNull() && !req.ConfigValue.IsUnknown() {
		return
	}

	// If the prior state has a known value, use it. This keeps plans stable
	// for existing resources.
	if !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
		resp.PlanValue = req.StateValue
		return
	}

	// No configuration value and no prior state value. The value will be
	// supplied by the provider during apply, so leave it unknown.
	resp.PlanValue = types.StringUnknown()
}
