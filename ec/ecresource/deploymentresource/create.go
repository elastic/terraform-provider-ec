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

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	deploymentv "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/deployment/v2"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.ready(&resp.Diagnostics) {
		return
	}

	var config deploymentv.DeploymentTF
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan deploymentv.DeploymentTF
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

	requestId := deploymentapi.RequestID(plan.RequestId.Value)

	res, err := deploymentapi.Create(deploymentapi.CreateParams{
		API:       r.client,
		RequestID: requestId,
		Request:   request,
		Overrides: &deploymentapi.PayloadOverrides{
			Name:    plan.Name.Value,
			Version: plan.Version.Value,
			Region:  plan.Region.Value,
		},
	})

	if err != nil {
		resp.Diagnostics.AddError("failed creating deployment", err.Error())
		resp.Diagnostics.AddError("failed creating deployment", newCreationError(requestId).Error())
		return
	}

	if err := WaitForPlanCompletion(r.client, *res.ID); err != nil {
		resp.Diagnostics.AddError("failed tracking create progress", newCreationError(requestId).Error())
		return
	}

	tflog.Trace(ctx, "created a resource")

	resp.Diagnostics.Append(deploymentv.HandleRemoteClusters(ctx, r.client, *res.ID, plan.Elasticsearch)...)

	deployment, diags := r.read(ctx, *res.ID, nil, plan, res.Resources)

	resp.Diagnostics.Append(diags...)

	if deployment == nil {
		resp.Diagnostics.AddError("cannot read just created resource", "")
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, deployment)...)
}

func newCreationError(reqID string) error {
	return fmt.Errorf(
		`set "request_id" to "%s" to recreate the deployment resources`, reqID,
	)
}
