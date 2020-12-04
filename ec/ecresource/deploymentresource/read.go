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
	"errors"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/apierror"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deputil"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/esremoteclustersapi"
	"github.com/elastic/cloud-sdk-go/pkg/client/deployments"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Read queries the remote deployment state and updates the local state.
func readResource(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.API)

	res, err := deploymentapi.Get(deploymentapi.GetParams{
		API: client, DeploymentID: d.Id(),
		QueryParams: deputil.QueryParams{
			ShowSettings:     true,
			ShowPlans:        true,
			ShowMetadata:     true,
			ShowPlanDefaults: true,
		},
	})
	if err != nil {
		if deploymentNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(multierror.NewPrefixed("failed reading deployment", err))
	}

	if !hasRunningResources(res) {
		d.SetId("")
		return nil
	}

	var diags diag.Diagnostics
	remotes, err := esremoteclustersapi.Get(esremoteclustersapi.GetParams{
		API: client, DeploymentID: d.Id(),
		RefID: d.Get("elasticsearch.0.ref_id").(string),
	})
	if err != nil {
		diags = append(diags, diag.FromErr(
			multierror.NewPrefixed("failed reading remote clusters", err),
		)...)
	}

	if remotes == nil {
		remotes = &models.RemoteResources{}
	}

	if err := modelToState(d, res, *remotes); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func deploymentNotFound(err error) bool {
	// We're using the As() call since we do not care about the error value
	// but do care about the error's contents type since it's an implicit 404.
	var notDeploymentNotFound *deployments.GetDeploymentNotFound
	if errors.As(err, &notDeploymentNotFound) {
		return true
	}

	// We also check for the case where a 403 is thrown for ESS.
	return apierror.IsRuntimeStatusCode(err, 403)
}
