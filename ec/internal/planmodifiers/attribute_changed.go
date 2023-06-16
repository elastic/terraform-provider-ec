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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func AttributeChanged(ctx context.Context, p path.Path, plan tfsdk.Plan, state tfsdk.State) (bool, diag.Diagnostics) {
	var planValue attr.Value

	if diags := plan.GetAttribute(ctx, p, &planValue); diags.HasError() {
		return false, diags
	}

	var stateValue attr.Value

	if diags := state.GetAttribute(ctx, p, &stateValue); diags.HasError() {
		return false, diags
	}

	return !planValue.Equal(stateValue), nil
}
