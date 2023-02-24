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

package elasticsearchkeystoreresource

import (
	"context"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/eskeystoreapi"
)

// Create will create an item in the Elasticsearch keystore
func (r Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	if !resourceReady(r, &response.Diagnostics) {
		return
	}

	var newState modelV0

	diags := request.Plan.Get(ctx, &newState)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if _, err := eskeystoreapi.Update(eskeystoreapi.UpdateParams{
		API:          r.client,
		DeploymentID: newState.DeploymentID.Value,
		Contents:     expandModel(ctx, newState),
	}); err != nil {
		response.Diagnostics.AddError(err.Error(), err.Error())
		return
	}

	newState.ID = types.String{Value: hashID(newState.DeploymentID.Value, newState.SettingName.Value)}

	found, diags := r.read(ctx, newState.DeploymentID.Value, &newState)
	response.Diagnostics.Append(diags...)
	if !found {
		response.Diagnostics.AddError(
			"Failed to read Elasticsearch keystore after create.",
			"Failed to read Elasticsearch keystore after create.",
		)
		response.State.RemoveResource(ctx)
		return
	}
	if response.Diagnostics.HasError() {
		return
	}

	// Finally, set the state
	response.Diagnostics.Append(response.State.Set(ctx, newState)...)
}

func hashID(elem ...string) string {
	return strconv.Itoa(schema.HashString(strings.Join(elem, "-")))
}
