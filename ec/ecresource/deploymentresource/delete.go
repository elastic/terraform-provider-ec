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
	"strings"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/client/deployments"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Delete shuts down and deletes the remote deployment retrying up to 3 times
// the Shutdown API call in case the plan returns with a failure that contains
// the "Timeout Exceeded" string, which is a fairly common transient error state
// returned from the API.
func deleteResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	const maxRetries = 3
	var retries int
	timeout := d.Timeout(schema.TimeoutDelete)
	client := meta.(*api.API)

	return diag.FromErr(resource.RetryContext(ctx, timeout, func() *resource.RetryError {
		if _, err := deploymentapi.Shutdown(deploymentapi.ShutdownParams{
			API: client, DeploymentID: d.Id(),
		}); err != nil {
			if alreadyDestroyed(err) {
				d.SetId("")
				return nil
			}
			return resource.NonRetryableError(multierror.NewPrefixed(
				"failed shutting down the deployment", err,
			))
		}

		if err := WaitForPlanCompletion(client, d.Id()); err != nil {
			if shouldRetryShutdown(err, retries, maxRetries) {
				retries++
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		if err := handleTrafficFilterChange(d, client); err != nil {
			return resource.NonRetryableError(err)
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
	}))
}

func alreadyDestroyed(err error) bool {
	var destroyed *deployments.ShutdownDeploymentNotFound
	return errors.As(err, &destroyed)
}

func shouldRetryShutdown(err error, retries, maxRetries int) bool {
	const timeout = "Timeout exceeded"
	needsRetry := retries < maxRetries

	var isTimeout, isFailDeallocate bool
	if err != nil {
		isTimeout = strings.Contains(err.Error(), timeout)
		isFailDeallocate = strings.Contains(
			err.Error(), "Some instances were not stopped",
		)
	}
	return (needsRetry && isTimeout) ||
		(needsRetry && isFailDeallocate)
}
