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

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/extensionapi"
	"github.com/elastic/cloud-sdk-go/pkg/client/extensions"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func readResource(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.API)

	res, err := extensionapi.Get(extensionapi.GetParams{
		API:         client,
		ExtensionID: d.Id(),
	})
	if err != nil {
		if extensionNotFound(err) {
			d.SetId("")
			return nil
		}

		return diag.FromErr(multierror.NewPrefixed("failed reading extension", err))
	}

	if err := modelToState(d, res); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func extensionNotFound(err error) bool {
	// We're using the As() call since we do not care about the error value
	// but do care about the error's contents type since it's an implicit 404.
	var extensionNotFound *extensions.GetExtensionNotFound
	return errors.As(err, &extensionNotFound)
}

func modelToState(d *schema.ResourceData, model *models.Extension) error {
	if err := d.Set("name", model.Name); err != nil {
		return err
	}

	if err := d.Set("version", model.Version); err != nil {
		return err
	}

	if err := d.Set("extension_type", model.ExtensionType); err != nil {
		return err
	}

	if err := d.Set("description", model.Description); err != nil {
		return err
	}

	if err := d.Set("url", model.URL); err != nil {
		return err
	}

	if err := d.Set("download_url", model.DownloadURL); err != nil {
		return err
	}

	if filemeta := model.FileMetadata; filemeta != nil {
		if err := d.Set("last_modified", filemeta.LastModifiedDate.String()); err != nil {
			return err
		}

		if err := d.Set("size", filemeta.Size); err != nil {
			return err
		}
	}

	return nil
}
