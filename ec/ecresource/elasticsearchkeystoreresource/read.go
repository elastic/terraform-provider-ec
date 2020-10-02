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

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/eskeystoreapi"
)

// read queries the remote Elasticsearch keystore state and updates the local state.
func read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var client = meta.(*api.API)
	deploymentID := d.Get("deployment_id").(string)

	res, err := eskeystoreapi.Get(eskeystoreapi.GetParams{
		API:          client,
		DeploymentID: deploymentID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if err := modelToState(d, res); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func modelToState(d *schema.ResourceData, res *models.KeystoreContents) error {
	var result = make([]interface{}, 0, len(res.Secrets))
	for setting, secret := range res.Secrets {
		var m = make(map[string]interface{})

		m["setting_name"] = setting

		if secret.AsFile != nil {
			m["as_file"] = *secret.AsFile
		}

		m["value"] = secret.Value

		result = append(result, m)

		if len(result) > 0 {
			if err := d.Set("secrets", result); err != nil {
				return err
			}
		}
	}

	return nil
}
