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
	"github.com/elastic/cloud-sdk-go/pkg/api/apierror"
	"github.com/elastic/cloud-sdk-go/pkg/client/extensions"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/go-openapi/runtime"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func readRequest(d *schema.ResourceData, client *api.API) (*models.Extension, error) {
	res, err := client.V1API.Extensions.GetExtension(
		extensions.NewGetExtensionParams().WithExtensionID(d.Id()),
		client.AuthWriter)

	if err != nil {
		return nil, apierror.Wrap(err)
	}
	return res.Payload, nil
}

func createRequest(client *api.API, d *schema.ResourceData) (*models.Extension, error) {
	name := d.Get("name").(string)
	version := d.Get("version").(string)
	extensionsType := d.Get("extension_type").(string)
	description := d.Get("description").(string)

	body := &models.CreateExtensionRequest{
		Name:          &name,
		Version:       &version,
		ExtensionType: &extensionsType,
		Description:   description,
	}

	res, err := client.V1API.Extensions.CreateExtension(
		extensions.NewCreateExtensionParams().WithBody(body),
		client.AuthWriter)

	if err != nil {
		return nil, apierror.Wrap(err)
	}
	return res.Payload, nil
}

func updateRequest(client *api.API, d *schema.ResourceData) (*models.Extension, error) {
	name := d.Get("name").(string)
	version := d.Get("version").(string)
	extensionsType := d.Get("extension_type").(string)
	description := d.Get("description").(string)

	body := &models.UpdateExtensionRequest{
		Name:          &name,
		Version:       &version,
		ExtensionType: &extensionsType,
		Description:   description,
	}

	res, err := client.V1API.Extensions.UpdateExtension(
		extensions.NewUpdateExtensionParams().WithBody(body).WithExtensionID(d.Id()),
		client.AuthWriter)
	if err != nil {
		return nil, apierror.Wrap(err)
	}

	return res.Payload, nil
}

func deleteRequest(client *api.API, d *schema.ResourceData) error {
	if _, err := client.V1API.Extensions.DeleteExtension(
		extensions.NewDeleteExtensionParams().WithExtensionID(d.Id()),
		client.AuthWriter); err != nil {
		return apierror.Wrap(err)
	}

	return nil
}

func uploadRequest(client *api.API, d *schema.ResourceData) error {
	reader, err := os.Open(d.Get("file_path").(string))
	if err != nil {
		return multierror.NewPrefixed("failed open file", err)
	}

	if _, err := client.V1API.Extensions.UploadExtension(
		extensions.NewUploadExtensionParams().WithExtensionID(d.Id()).
			WithFile(runtime.NamedReader(d.Get("file_path").(string), reader)),
		client.AuthWriter); err != nil {
		return apierror.Wrap(err)
	}

	return nil
}
