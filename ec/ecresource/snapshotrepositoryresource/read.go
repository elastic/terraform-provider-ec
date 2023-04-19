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
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api/apierror"
	"github.com/elastic/cloud-sdk-go/pkg/api/platformapi/snaprepoapi"
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

	found, diags := r.read(newState.ID.Value, &newState)
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
	res, err := snaprepoapi.Get(snaprepoapi.GetParams{
		API:    r.client,
		Region: "ece-region", // This resource is only usable for ECE installations. Thus, we can default to ece-region.
		Name:   id,
	})
	if err != nil {
		if apierror.IsRuntimeStatusCode(err, 404) {
			return false, diags
		}
		diags.AddError("failed reading snapshot repository", err.Error())
		return true, diags
	}

	diags.Append(modelToState(res, state)...)
	return true, diags
}

func modelToState(model *models.RepositoryConfig, state *modelV0) diag.Diagnostics {
	var diags diag.Diagnostics

	if model.RepositoryName != nil {
		state.Name = types.String{Value: *model.RepositoryName}
	}

	config, _ := model.Config.(map[string]interface{})
	if repositoryType, ok := config["type"]; ok && repositoryType != nil {
		if settingsInterface, ok := config["settings"]; ok && settingsInterface != nil {
			settings := settingsInterface.(map[string]interface{})
			// Parse into S3 schema if possible, but fall back to Generic when custom settings have been used.
			if repositoryType.(string) == "s3" && containsOnlyKnownS3Settings(settings) {
				if state.S3 == nil {
					state.S3 = &s3RepositoryV0{}
				}
				if region, ok := settings["region"]; ok && region != nil {
					state.S3.Region = types.String{Value: region.(string)}
				}
				if bucket, ok := settings["bucket"]; ok && bucket != nil {
					state.S3.Bucket = types.String{Value: bucket.(string)}
				}
				if accessKey, ok := settings["access_key"]; ok && accessKey != nil {
					state.S3.AccessKey = types.String{Value: accessKey.(string)}
				}
				if secretKey, ok := settings["secret_key"]; ok && secretKey != nil {
					state.S3.SecretKey = types.String{Value: secretKey.(string)}
				}
				if serverSideEncryption, ok := settings["server_side_encryption"]; ok && serverSideEncryption != nil {
					state.S3.ServerSideEncryption = types.Bool{Value: serverSideEncryption.(bool)}
				}
				if endpoint, ok := settings["endpoint"]; ok && endpoint != nil {
					state.S3.Endpoint = types.String{Value: endpoint.(string)}
				}
				if pathStyleAccess, ok := settings["path_style_access"]; ok && pathStyleAccess != nil {
					state.S3.PathStyleAccess = types.Bool{Value: pathStyleAccess.(bool)}
				}
			} else {
				if state.Generic == nil {
					state.Generic = &genericRepositoryV0{}
				}
				state.Generic.Type = types.String{Value: repositoryType.(string)}
				jsonSettings, err := json.Marshal(settings)
				if err != nil {
					diags.AddError(
						fmt.Sprintf("failed reading snapshot repository: unable to marshal settings - %s", err),
						fmt.Sprintf("failed reading snapshot repository: unable to marshal settings - %s", err),
					)
				} else {
					state.Generic.Settings = types.String{Value: string(jsonSettings)}
				}
			}
		}
	}
	return diags
}

func containsOnlyKnownS3Settings(settings map[string]interface{}) bool {
	attributes := s3Schema().Attributes.GetAttributes()
	for key := range settings {
		if _, ok := attributes[key]; !ok {
			return false
		}
	}
	return true
}
