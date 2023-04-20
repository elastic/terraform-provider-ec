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

package snapshotrepositoryresource

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/elastic/cloud-sdk-go/pkg/api/platformapi/snaprepoapi"
	"github.com/elastic/cloud-sdk-go/pkg/util"
)

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	if !resourceReady(r, &response.Diagnostics) {
		return
	}

	var newState modelV0

	diags := request.Plan.Get(ctx, &newState)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	var repositoryType string
	var repositoryConfig util.Validator
	if newState.S3 != nil {
		repositoryType = "s3"
		repositoryConfig = snaprepoapi.S3Config{
			Region:               newState.S3.Region.Value,
			Bucket:               newState.S3.Bucket.Value,
			AccessKey:            newState.S3.AccessKey.Value,
			SecretKey:            newState.S3.SecretKey.Value,
			ServerSideEncryption: newState.S3.ServerSideEncryption.Value,
			Endpoint:             newState.S3.Endpoint.Value,
			PathStyleAccess:      newState.S3.PathStyleAccess.Value,
		}
	} else {
		var err error
		repositoryType = newState.Generic.Type.Value
		repositoryConfig, err = snaprepoapi.ParseGenericConfig(strings.NewReader(newState.Generic.Settings.Value))
		if err != nil {
			response.Diagnostics.AddError(err.Error(), err.Error())
			return
		}
	}

	err := snaprepoapi.Set(
		snaprepoapi.SetParams{
			API:    r.client,
			Region: "ece-region", // This resource is only usable for ECE installations. Thus, we can default to ece-region.
			Name:   newState.Name.Value,
			Type:   repositoryType,
			Config: repositoryConfig,
		},
	)
	if err != nil {
		response.Diagnostics.AddError(err.Error(), err.Error())
		return
	}

	found, diags := r.read(newState.ID.Value, &newState)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	if !found {
		response.Diagnostics.AddError(
			"Failed to read snapshot repository after update.",
			"Failed to read snapshot repository after update.",
		)
		response.State.RemoveResource(ctx)
		return
	}

	// Finally, set the state
	response.Diagnostics.Append(response.State.Set(ctx, newState)...)
}
