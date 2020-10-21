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
	"strings"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Update syncs the remote state with the local.
func updateResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.API)

	if hasDeploymentChange(d) {
		if err := updateDeployment(ctx, d, client); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := handleTrafficFilterChange(d, client); err != nil {
		return diag.FromErr(err)
	}

	if err := handleRemoteClusters(d, client); err != nil {
		return diag.FromErr(err)
	}

	return readResource(ctx, d, meta)
}

func updateDeployment(_ context.Context, d *schema.ResourceData, client *api.API) error {
	req, err := updateResourceToModel(d, client)
	if err != nil {
		return err
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
		return multierror.NewPrefixed("failed updating deployment", err)
	}

	if err := WaitForPlanCompletion(client, d.Id()); err != nil {
		return multierror.NewPrefixed("failed tracking update progress", err)
	}

	return parseCredentials(d, res.Resources)
}

// hasDeploymentChange checks if there's any change in the resource attributes
// except in the "traffic_filter" prefixed keys. If so, it returns true.
func hasDeploymentChange(d *schema.ResourceData) bool {
	for attr := range d.State().Attributes {
		if strings.HasPrefix(attr, "traffic_filter") {
			continue
		}
		// Check if any of the resource attributes has a change.
		if d.HasChange(attr) {
			return true
		}
	}
	return false
}
