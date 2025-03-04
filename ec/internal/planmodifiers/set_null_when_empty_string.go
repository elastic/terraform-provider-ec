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

func SetNullWhenEmptyString() planmodifier.String {
	return setNullWhenEmptyString{}
}

type setNullWhenEmptyString struct{}

func (m setNullWhenEmptyString) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	if req.PlanValue.ValueString() != "" {
		return
	}

	resp.PlanValue = types.StringNull()
}

// Description returns a human-readable description of the plan modifier.
func (r setNullWhenEmptyString) Description(ctx context.Context) string {
	return "Set the plan value to null if currently configured to null."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r setNullWhenEmptyString) MarkdownDescription(ctx context.Context) string {
	return "Set the plan value to null if currently configured to null."
}
