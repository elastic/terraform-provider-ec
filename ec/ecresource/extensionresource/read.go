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

package extensionresource

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/extensionapi"
	"github.com/elastic/cloud-sdk-go/pkg/client/extensions"
	"github.com/elastic/cloud-sdk-go/pkg/models"
)

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	if !resourceReady(r, &response.Diagnostics) {
		return
	}

	var newState modelV0

	diags := request.State.Get(ctx, &newState)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	found, diags := r.read(newState.ID.ValueString(), &newState)
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

func (r *Resource) read(id string, state *modelV0) (found bool, diags diag.Diagnostics) {
	res, err := extensionapi.Get(extensionapi.GetParams{
		API:         r.client,
		ExtensionID: id,
	})
	if err != nil {
		if extensionNotFound(err) {
			return false, diags
		}
		diags.AddError("failed reading extension", err.Error())
		return true, diags
	}

	modelToState(res, state)
	return true, diags
}

func extensionNotFound(err error) bool {
	// We're using the As() call since we do not care about the error value
	// but do care about the error's contents type since it's an implicit 404.
	var extensionNotFound *extensions.GetExtensionNotFound
	return errors.As(err, &extensionNotFound)
}

func modelToState(model *models.Extension, state *modelV0) {
	if model.Name != nil {
		state.Name = types.StringValue(*model.Name)
	} else {
		state.Name = types.StringNull()
	}

	if model.Version != nil {
		state.Version = types.StringValue(*model.Version)
	} else {
		state.Version = types.StringNull()
	}

	if model.ExtensionType != nil {
		state.ExtensionType = types.StringValue(*model.ExtensionType)
	} else {
		state.ExtensionType = types.StringNull()
	}

	state.Description = types.StringValue(model.Description)

	if model.URL != nil {
		state.URL = types.StringValue(*model.URL)
	} else {
		state.URL = types.StringNull()
	}

	state.DownloadURL = types.StringValue(model.DownloadURL)

	if metadata := model.FileMetadata; metadata != nil {
		state.LastModified = types.StringValue(metadata.LastModifiedDate.String())
		state.Size = types.Int64Value(metadata.Size)
	} else {
		state.LastModified = types.StringNull()
		state.Size = types.Int64Null()
	}
}
