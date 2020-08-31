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
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Delete shuts down and deletes the remote deployment.
func Delete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.API)

	if _, err := deploymentapi.Shutdown(deploymentapi.ShutdownParams{
		API: client, DeploymentID: d.Id(),
	}); err != nil {
		return diag.FromErr(err)
	}

	if err := WaitForPlanCompletion(client, d.Id()); err != nil {
		return diag.FromErr(err)
	}

	if err := handleTrafficFilterChange(d, client); err != nil {
		return diag.FromErr(err)
	}

	// We don't particularly care if delete succeeds or not. It's better to
	// remove it, but it might fail on ESS. For example, when user's aren't
	// allowed to delete deployments, or on ECE when the cluster is "still
	// being shutdown". Sumarizing, even if the call fails the deployment
	// won't be there.
	_, _ = deploymentapi.Delete(deploymentapi.DeleteParams{
		API: client, DeploymentID: d.Id(),
	})

	d.SetId("")
	return nil
}
