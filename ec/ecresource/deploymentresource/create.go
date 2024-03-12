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
	"fmt"
	"time"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	deploymentv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/deployment/v2"
	v2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/deployment/v2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.ready(&resp.Diagnostics) {
		return
	}

	var config v2.DeploymentTF
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan v2.DeploymentTF
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	request, diags := plan.CreateRequest(ctx, r.client)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	requestId := deploymentapi.RequestID(plan.RequestId.ValueString())

	res, err := deploymentapi.Create(deploymentapi.CreateParams{
		API:       r.client,
		RequestID: requestId,
		Request:   request,
		Overrides: &deploymentapi.PayloadOverrides{
			Name:    plan.Name.ValueString(),
			Version: plan.Version.ValueString(),
			Region:  plan.Region.ValueString(),
		},
	})

	if err != nil {
		resp.Diagnostics.AddError("failed creating deployment", err.Error())
		resp.Diagnostics.AddError("failed creating deployment", newCreationError(requestId).Error())
		return
	}

	// Set the ID immediately so the deployment is managed by Terraform.
	// If the rest of this fails the deployment will be tainted and recreated,
	// but that's preferable to leaving an unmanaged resource.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), *res.ID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := WaitForPlanCompletion(r.client, *res.ID); err != nil {
		resp.Diagnostics.AddError("failed tracking create progress", err.Error())
		resp.Diagnostics.AddError("failed tracking create progress", newCreationError(requestId).Error())
		return
	}

	tflog.Trace(ctx, "created deployment resource")

	resp.Diagnostics.Append(v2.HandleRemoteClusters(ctx, r.client, *res.ID, plan.Elasticsearch)...)

	filters := []string{}
	if request.Settings != nil && request.Settings.TrafficFilterSettings != nil && request.Settings.TrafficFilterSettings.Rulesets != nil {
		filters = request.Settings.TrafficFilterSettings.Rulesets
	}

	deployment, diags := r.readUntilEndpointsAreAvailable(ctx, *res.ID, nil, &plan, res.Resources, filters, nil)
	updatePrivateStateTrafficFilters(ctx, resp.Private, filters)

	resp.Diagnostics.Append(diags...)
	if deployment == nil {
		resp.Diagnostics.AddError("cannot read just created resource", "")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, deployment)...)
}

func (r *Resource) readUntilEndpointsAreAvailable(ctx context.Context, id string, state *deploymentv2.DeploymentTF, plan *deploymentv2.DeploymentTF, deploymentResources []*models.DeploymentResource, privateFilters []string, readResponse *resource.ReadResponse) (deployment *deploymentv2.Deployment, diags diag.Diagnostics) {
	for i := 12; i > 0; i-- {
		deployment, diags = r.read(ctx, id, state, plan, deploymentResources, privateFilters, readResponse)
		if diags.HasError() {
			return deployment, diags
		}

		if deployment.IntegrationsServer == nil {
			return deployment, diags
		}

		if deployment.IntegrationsServer.Endpoints != nil && deployment.IntegrationsServer.Endpoints.APM != nil && deployment.IntegrationsServer.Endpoints.Fleet != nil {
			return deployment, diags
		}

		time.Sleep(15 * time.Second)
	}

	return deployment, diags
}

func newCreationError(reqID string) error {
	return fmt.Errorf(
		`set "request_id" to "%s" to recreate the deployment resources`, reqID,
	)
}
