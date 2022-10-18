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
	"os"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/extensionapi"
)

func (r *Resource) uploadExtension(state modelV0) diag.Diagnostics {
	var diags diag.Diagnostics

	reader, err := os.Open(state.FilePath.Value)
	if err != nil {
		diags.AddError("failed to open file", err.Error())
		return diags
	}

	_, err = extensionapi.Upload(extensionapi.UploadParams{
		API:         r.client,
		ExtensionID: state.ID.Value,
		File:        reader,
	})
	if err != nil {
		diags.AddError("failed to upload file", err.Error())
		return diags
	}

	return diags
}
