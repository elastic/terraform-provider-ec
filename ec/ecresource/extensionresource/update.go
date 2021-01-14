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

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/extensionapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func updateResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.API)

	_, err := updateRequest(client, d)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, ok := d.GetOk("file_path"); ok && d.HasChanges("file_hash", "last_modified", "size") {
		if err := uploadExtension(client, d); err != nil {
			return diag.FromErr(multierror.NewPrefixed("failed to upload file", err))
		}
	}

	return readResource(ctx, d, meta)
}

func updateRequest(client *api.API, d *schema.ResourceData) (*models.Extension, error) {
	name := d.Get("name").(string)
	version := d.Get("version").(string)
	extensionType := d.Get("extension_type").(string)
	description := d.Get("description").(string)
	downloadURL := d.Get("download_url").(string)

	body := extensionapi.UpdateParams{
		API:         client,
		ExtensionID: d.Id(),
		Name:        name,
		Version:     version,
		Type:        extensionType,
		Description: description,
		DownloadURL: downloadURL,
	}

	res, err := extensionapi.Update(body)
	if err != nil {
		return nil, err
	}

	return res, nil
}
