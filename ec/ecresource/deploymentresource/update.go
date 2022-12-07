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
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/trafficfilterapi"
	v2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/deployment/v2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan v2.DeploymentTF

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var state v2.DeploymentTF

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateReq, diags := plan.UpdateRequest(ctx, r.client, state)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	res, err := deploymentapi.Update(deploymentapi.UpdateParams{
		API:          r.client,
		DeploymentID: plan.Id.Value,
		Request:      updateReq,
		Overrides: deploymentapi.PayloadOverrides{
			Version: plan.Version.Value,
			Region:  plan.Region.Value,
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("failed updating deployment", err.Error())
		return
	}

	if err := WaitForPlanCompletion(r.client, plan.Id.Value); err != nil {
		resp.Diagnostics.AddError("failed tracking update progress", err.Error())
		return
	}

	resp.Diagnostics.Append(handleTrafficFilterChange(ctx, r.client, plan, state)...)

	resp.Diagnostics.Append(v2.HandleRemoteClusters(ctx, r.client, plan.Id.Value, plan.Elasticsearch)...)

	deployment, diags := r.read(ctx, plan.Id.Value, &state, plan, res.Resources)

	resp.Diagnostics.Append(diags...)

	if deployment == nil {
		resp.Diagnostics.AddError("cannot read just updated resource", "")
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, deployment)...)
}

func handleTrafficFilterChange(ctx context.Context, client *api.API, plan, state v2.DeploymentTF) diag.Diagnostics {
	if plan.TrafficFilter.IsNull() || plan.TrafficFilter.Equal(state.TrafficFilter) {
		return nil
	}

	var planRules, stateRules ruleSet
	if diags := plan.TrafficFilter.ElementsAs(ctx, &planRules, true); diags.HasError() {
		return diags
	}

	if diags := state.TrafficFilter.ElementsAs(ctx, &stateRules, true); diags.HasError() {
		return diags
	}

	var rulesToAdd, rulesToDelete []string

	for _, rule := range planRules {
		if !stateRules.exist(rule) {
			rulesToAdd = append(rulesToAdd, rule)
		}
	}

	for _, rule := range stateRules {
		if !planRules.exist(rule) {
			rulesToDelete = append(rulesToDelete, rule)
		}
	}

	var diags diag.Diagnostics
	for _, rule := range rulesToAdd {
		if err := associateRule(rule, plan.Id.Value, client); err != nil {
			diags.AddError("cannot associate traffic filter rule", err.Error())
		}
	}

	for _, rule := range rulesToDelete {
		if err := removeRule(rule, plan.Id.Value, client); err != nil {
			diags.AddError("cannot remove traffic filter rule", err.Error())
		}
	}

	return diags
}

type ruleSet []string

func (rs ruleSet) exist(rule string) bool {
	for _, r := range rs {
		if r == rule {
			return true
		}
	}
	return false
}

func associateRule(ruleID, deploymentID string, client *api.API) error {
	res, err := trafficfilterapi.Get(trafficfilterapi.GetParams{
		API: client, ID: ruleID, IncludeAssociations: true,
	})
	if err != nil {
		return err
	}

	// When the rule has already been associated, return.
	for _, assoc := range res.Associations {
		if deploymentID == *assoc.ID {
			return nil
		}
	}

	// Create assignment.
	if err := trafficfilterapi.CreateAssociation(trafficfilterapi.CreateAssociationParams{
		API: client, ID: ruleID, EntityType: "deployment", EntityID: deploymentID,
	}); err != nil {
		return err
	}
	return nil
}
