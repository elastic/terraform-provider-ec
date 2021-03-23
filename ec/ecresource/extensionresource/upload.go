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

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/extensionapi"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func uploadExtension(client *api.API, d *schema.ResourceData) error {
	filePath := d.Get("file_path").(string)
	reader, err := os.Open(filePath)
	if err != nil {
		return multierror.NewPrefixed("failed to open file", err)
	}

	_, err = extensionapi.Upload(extensionapi.UploadParams{
		API:         client,
		ExtensionID: d.Id(),
		File:        reader,
	})
	if err != nil {
		return err
	}

	return nil
}
