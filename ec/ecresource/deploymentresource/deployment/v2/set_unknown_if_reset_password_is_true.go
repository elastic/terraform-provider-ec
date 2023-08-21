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

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type setUnknownIfResetPasswordIsTrue struct{}

var _ planmodifier.String = setUnknownIfResetPasswordIsTrue{}

func (m setUnknownIfResetPasswordIsTrue) Description(ctx context.Context) string {
	return m.MarkdownDescription(ctx)
}

func (m setUnknownIfResetPasswordIsTrue) MarkdownDescription(ctx context.Context) string {
	return "Sets the planned value to unknown if the reset_elasticsearch_password config value is true"
}

func (m setUnknownIfResetPasswordIsTrue) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// if the config is the unknown value, use the unknown value otherwise, interpolation gets messed up
	if req.ConfigValue.IsUnknown() {
		return
	}

	var isResetting *bool
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("reset_elasticsearch_password"), &isResetting)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if isResetting != nil && *isResetting {
		resp.PlanValue = types.StringUnknown()
	}
}
