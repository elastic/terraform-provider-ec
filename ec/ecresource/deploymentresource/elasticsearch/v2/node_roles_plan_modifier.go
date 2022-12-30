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

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func UseNodeRolesDefault() tfsdk.AttributePlanModifier {
	return nodeRolesDefault{}
}

type nodeRolesDefault struct{}

func (r nodeRolesDefault) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	useState, useNodeRoles := useStateAndNodeRolesInPlanModifiers(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		return
	}

	if !useState {
		return
	}

	// If useNodeRoles is false, we can use the current state and
	// 	it should be null in this case - we don't migrate back from node_roles to node_types
	if !useNodeRoles && !req.AttributeState.IsNull() {
		// it should not happen
		return
	}

	// If useNodeRoles is true, then there is either
	// 	* state already uses node_roles or
	// 	* state uses node_types but we need to migrate to node_roles.
	// We cannot use state in the second case (migration to node_roles)
	// It happens when the attriubute state is null.
	if useNodeRoles && req.AttributeState.IsNull() {
		return
	}

	resp.AttributePlan = req.AttributeState
}

// Description returns a human-readable description of the plan modifier.
func (r nodeRolesDefault) Description(ctx context.Context) string {
	return "Use current state if it's still valid."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r nodeRolesDefault) MarkdownDescription(ctx context.Context) string {
	return "Use current state if it's still valid."
}
