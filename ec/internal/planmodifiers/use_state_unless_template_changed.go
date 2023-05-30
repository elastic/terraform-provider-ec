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

// Use current state for a topology's attribute if the topology's state is not nil and the template attribute has not changed
func UseStateForUnknownUnlessTemplateChanged() planmodifier.String {
	return useStateForUnknownUnlessTemplateChanged{}
}

type useStateForUnknownUnlessTemplateChanged struct{}

func (m useStateForUnknownUnlessTemplateChanged) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !resp.PlanValue.IsUnknown() {
		return
	}

	// if the config is the unknown value, use the unknown value otherwise, interpolation gets messed up
	if req.ConfigValue.IsUnknown() {
		return
	}

	if req.StateValue.IsNull() {
		return
	}

	templateChanged, diags := AttributeChanged(ctx, path.Root("deployment_template_id"), req.Plan, req.State)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if templateChanged {
		return
	}

	resp.PlanValue = req.StateValue
}

func (r useStateForUnknownUnlessTemplateChanged) Description(ctx context.Context) string {
	return "Use tier's state if it's defined and template is the same."
}

func (r useStateForUnknownUnlessTemplateChanged) MarkdownDescription(ctx context.Context) string {
	return "Use tier's state if it's defined and template is the same."
}
