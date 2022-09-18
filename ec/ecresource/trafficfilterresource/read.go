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

package trafficfilterresource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/trafficfilterapi"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

// Read queries the remote deployment traffic filter ruleset state and updates
// the local state.
func (r Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	if !resourceReady(r, &response.Diagnostics) {
		return
	}

	var newState modelV0

	diags := request.State.Get(ctx, &newState)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	found, diags := r.read(ctx, newState.ID.Value, &newState)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	if !found {
		response.State.RemoveResource(ctx)
		return
	}

	// Finally, set the state
	response.Diagnostics.Append(response.State.Set(ctx, newState)...)
}

func (r Resource) read(ctx context.Context, id string, state *modelV0) (found bool, diags diag.Diagnostics) {
	res, err := trafficfilterapi.Get(trafficfilterapi.GetParams{
		API: r.client, ID: id, IncludeAssociations: false,
	})
	if err != nil {
		if util.TrafficFilterNotFound(err) {
			return false, diags
		}
		diags.AddError(err.Error(), err.Error())
		return true, diags
	}

	diags.Append(modelToState(ctx, res, state)...)
	return true, diags
}
