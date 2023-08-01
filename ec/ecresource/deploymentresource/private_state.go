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

package deploymentresource

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type PrivateState interface {
	GetKey(context.Context, string) ([]byte, diag.Diagnostics)
	SetKey(context.Context, string, []byte) diag.Diagnostics
}

const trafficFilterStateKey = "traffic_filters"

func readPrivateStateTrafficFilters(ctx context.Context, state PrivateState) ([]string, diag.Diagnostics) {
	privateFilterBytes, diags := state.GetKey(ctx, trafficFilterStateKey)
	if privateFilterBytes == nil || diags.HasError() {
		return []string{}, diags
	}

	var privateFilters []string
	err := json.Unmarshal(privateFilterBytes, &privateFilters)
	if err != nil {
		diags.AddError("failed to parse private state", err.Error())
		return []string{}, diags
	}

	return privateFilters, diags
}

func updatePrivateStateTrafficFilters(ctx context.Context, state PrivateState, filters []string) diag.Diagnostics {
	var diags diag.Diagnostics
	filterBytes, err := json.Marshal(filters)
	if err != nil {
		diags.AddError("failed to update private state", err.Error())
		return diags
	}

	return state.SetKey(ctx, trafficFilterStateKey, filterBytes)
}
