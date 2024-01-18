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
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/depresourceapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/trafficfilterapi"
	v2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/deployment/v2"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
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

	// Read migrate request from private state
	migrateTemplateRequest, diags := readPrivateStateMigrateTemplateRequest(ctx, req.Private)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	updateReq, diags := plan.UpdateRequest(ctx, r.client, state, migrateTemplateRequest)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	res, err := deploymentapi.Update(deploymentapi.UpdateParams{
		API:          r.client,
		DeploymentID: plan.Id.ValueString(),
		Request:      updateReq,
		Overrides: deploymentapi.PayloadOverrides{
			Version: plan.Version.ValueString(),
			Region:  plan.Region.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("failed updating deployment", err.Error())
		return
	}

	if err := WaitForPlanCompletion(r.client, plan.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("failed tracking update progress", err.Error())
		return
	}

	privateFilters, d := readPrivateStateTrafficFilters(ctx, req.Private)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	planRules, diags := HandleTrafficFilterChange(ctx, r.client, plan, privateFilters)
	resp.Diagnostics.Append(diags...)
	updatePrivateStateTrafficFilters(ctx, resp.Private, planRules)
	resp.Diagnostics.Append(v2.HandleRemoteClusters(ctx, r.client, plan.Id.ValueString(), plan.Elasticsearch)...)

	deployment, diags := r.read(ctx, plan.Id.ValueString(), &state, &plan, res.Resources, planRules, nil)

	resp.Diagnostics.Append(diags...)

	if deployment == nil {
		resp.Diagnostics.AddError("cannot read just updated resource", "")
		resp.State.RemoveResource(ctx)
		return
	}

	if plan.ResetElasticsearchPassword.ValueBool() {
		newUsername, newPassword, diags := r.ResetElasticsearchPassword(plan.Id.ValueString(), *deployment.Elasticsearch.RefId)
		if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
			return
		}

		deployment.ElasticsearchUsername = newUsername
		deployment.ElasticsearchPassword = newPassword
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, deployment)...)
}

func (r *Resource) ResetElasticsearchPassword(deploymentID string, refID string) (string, string, diag.Diagnostics) {
	var diags diag.Diagnostics

	resetResp, err := depresourceapi.ResetElasticsearchPassword(depresourceapi.ResetElasticsearchPasswordParams{
		API:   r.client,
		ID:    deploymentID,
		RefID: refID,
	})

	if err != nil {
		diags.AddError("failed to reset elasticsearch password", err.Error())
		return "", "", diags
	}

	return *resetResp.Username, *resetResp.Password, diags
}

func HandleTrafficFilterChange(ctx context.Context, client *api.API, plan v2.DeploymentTF, stateRules ruleSet) ([]string, diag.Diagnostics) {
	var planRules ruleSet
	if diags := plan.TrafficFilter.ElementsAs(ctx, &planRules, true); diags.HasError() {
		return []string{}, diags
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
		if err := associateRule(rule, plan.Id.ValueString(), client); err != nil {
			diags.AddError("cannot associate traffic filter rule", err.Error())
		}
	}

	for _, rule := range rulesToDelete {
		if err := removeRule(rule, plan.Id.ValueString(), client); err != nil {
			diags.AddError("cannot remove traffic filter rule", err.Error())
		}
	}

	return planRules, diags
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

var (
	GetAssociation    = trafficfilterapi.Get
	CreateAssociation = trafficfilterapi.CreateAssociation
	DeleteAssociation = trafficfilterapi.DeleteAssociation
)

func associateRule(ruleID, deploymentID string, client *api.API) error {
	res, err := GetAssociation(trafficfilterapi.GetParams{
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
	if err := CreateAssociation(trafficfilterapi.CreateAssociationParams{
		API: client, ID: ruleID, EntityType: "deployment", EntityID: deploymentID,
	}); err != nil {
		return err
	}
	return nil
}

func removeRule(ruleID, deploymentID string, client *api.API) error {
	res, err := GetAssociation(trafficfilterapi.GetParams{
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
			return DeleteAssociation(trafficfilterapi.DeleteAssociationParams{
				API:        client,
				ID:         ruleID,
				EntityID:   *assoc.ID,
				EntityType: *assoc.EntityType,
			})
		}
	}

	return nil
}
