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
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/trafficfilterapi"
	"github.com/elastic/cloud-sdk-go/pkg/client/deployments"
	deploymentv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/deployment/v2"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.ready(&resp.Diagnostics) {
		return
	}

	var state deploymentv2.DeploymentTF

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	//TODO retries

	if _, err := deploymentapi.Shutdown(deploymentapi.ShutdownParams{
		API: r.client, DeploymentID: state.Id.Value,
	}); err != nil {
		if alreadyDestroyed(err) {
			return
		}
	}

	if err := WaitForPlanCompletion(r.client, state.Id.Value); err != nil {
		resp.Diagnostics.AddError("deployment deletion error", err.Error())
		return
	}

	// We don't particularly care if delete succeeds or not. It's better to
	// remove it, but it might fail on ESS. For example, when user's aren't
	// allowed to delete deployments, or on ECE when the cluster is "still
	// being shutdown". Sumarizing, even if the call fails the deployment
	// won't be there.
	_, _ = deploymentapi.Delete(deploymentapi.DeleteParams{
		API: r.client, DeploymentID: state.Id.Value,
	})
}

func alreadyDestroyed(err error) bool {
	var destroyed *deployments.ShutdownDeploymentNotFound
	return errors.As(err, &destroyed)
}

func removeRule(ruleID, deploymentID string, client *api.API) error {
	res, err := trafficfilterapi.Get(trafficfilterapi.GetParams{
		API: client, ID: ruleID, IncludeAssociations: true,
	})

	// If the rule is gone (403 or 404), return nil.
	if err != nil {
		if util.TrafficFilterNotFound(err) {
			return nil
		}
		return err
	}

	// If the rule is found, then delete the association.
	for _, assoc := range res.Associations {
		if deploymentID == *assoc.ID {
			return trafficfilterapi.DeleteAssociation(trafficfilterapi.DeleteAssociationParams{
				API:        client,
				ID:         ruleID,
				EntityID:   *assoc.ID,
				EntityType: *assoc.EntityType,
			})
		}
	}

	return nil
}
