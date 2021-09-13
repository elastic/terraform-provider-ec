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
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/eskeystoreapi"
)

// delete will delete an existing element in the Elasticsearch keystore
func delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.API)
	contents := expandModel(d)

	// Since we're using the Update API (PATCH method), we need to se the Value
	// field to nil for the keystore setting to be unset.
	if secret, ok := contents.Secrets[d.Get("setting_name").(string)]; ok {
		secret.Value = nil
	}

	if _, err := eskeystoreapi.Update(eskeystoreapi.UpdateParams{
		API:          client,
		DeploymentID: d.Get("deployment_id").(string),
		Contents:     contents,
	}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return read(ctx, d, meta)
}
