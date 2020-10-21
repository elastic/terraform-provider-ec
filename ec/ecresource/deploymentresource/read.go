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
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deputil"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/esremoteclustersapi"
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
		return diag.FromErr(multierror.NewPrefixed("failed reading deployment", err))
	}

	var diags diag.Diagnostics
	remotes, err := esremoteclustersapi.Get(esremoteclustersapi.GetParams{
		API: client, DeploymentID: d.Id(),
		RefID: d.Get("elasticsearch.0.ref_id").(string),
	})
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if remotes == nil {
		remotes = &models.RemoteResources{}
	}

	if err := modelToState(d, res, *remotes); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}
