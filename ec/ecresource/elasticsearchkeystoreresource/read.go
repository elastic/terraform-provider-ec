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

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/eskeystoreapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
)

// Read queries the remote Elasticsearch keystore state and updates the local state.
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

	found, diags := r.read(ctx, newState.DeploymentID.Value, &newState)
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

func (r Resource) read(ctx context.Context, deploymentID string, state *modelV0) (found bool, diags diag.Diagnostics) {
	res, err := eskeystoreapi.Get(eskeystoreapi.GetParams{
		API:          r.client,
		DeploymentID: deploymentID,
	})
	if err != nil {
		diags.AddError(err.Error(), err.Error())
		return true, diags
	}

	return modelToState(ctx, res, state)
}

// This modelToState function is a little different from others in that it does
// not set any other fields than "as_file". This is because the "value" is not
// returned by the API for obvious reasons, and thus we cannot reconcile that the
// value of the secret is the same in the remote as it is in the configuration.
func modelToState(ctx context.Context, res *models.KeystoreContents, state *modelV0) (found bool, diags diag.Diagnostics) {
	if secret, ok := res.Secrets[state.SettingName.Value]; ok {
		if secret.AsFile != nil {
			state.AsFile = types.Bool{Value: *secret.AsFile}
		}
		return true, nil
	}

	// When the secret is not found in the returned map of secrets, the resource should be removed from state.
	// Would only happen if secrets are removed from the underlying Deployment.
	return false, nil
}
