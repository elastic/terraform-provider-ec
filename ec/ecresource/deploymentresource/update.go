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

package deploymentresource

import (
	"context"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Update syncs the remote state with the local.
func Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.API)

	req, err := updateResourceToModel(d)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := deploymentapi.Update(deploymentapi.UpdateParams{
		API:          client,
		DeploymentID: d.Id(),
		Request:      req,
		Overrides: deploymentapi.PayloadOverrides{
			Version: d.Get("version").(string),
			Region:  d.Get("region").(string),
		},
	})

	if err != nil {
		return diag.FromErr(multierror.NewPrefixed("failed updating deployment", err))
	}

	if err := WaitForPlanCompletion(client, d.Id()); err != nil {
		return diag.FromErr(multierror.NewPrefixed("failed tracking update progress", err))
	}

	if diag := Read(ctx, d, meta); diag != nil {
		return diag
	}

	if err := parseCredentials(d, res.Resources); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
