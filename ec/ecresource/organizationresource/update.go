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

package organizationresource

import (
	"context"
	"fmt"
	"github.com/elastic/cloud-sdk-go/pkg/api/organizationapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"sort"
	"strings"
)

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	diagnostics := &response.Diagnostics

	var plan Organization
	var state Organization
	diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	diagnostics.Append(request.State.Get(ctx, &state)...)
	if diagnostics.HasError() {
		return
	}

	organizationID := plan.ID.ValueString()

	planMembers := make(map[string]types.Object)
	diags := plan.Members.ElementsAs(ctx, &planMembers, false)
	if diags.HasError() {
		diagnostics.Append(diags...)
		return
	}
	stateMembers := make(map[string]types.Object)
	diags = state.Members.ElementsAs(ctx, &stateMembers, false)
	if diags.HasError() {
		diagnostics.Append(diags...)
		return
	}

	// Create new members, update changed members
	for email, planMember := range planMembers {
		planMemberModel := toModel(ctx, planMember, diagnostics)
		if diagnostics.HasError() {
			continue
		}

		// create new invitation if member is in plan but not in state
		stateMember, ok := stateMembers[email]
		if !ok {
			r.createInvitation(ctx, email, planMemberModel, organizationID, diagnostics)
		} else {
			// member is in plan and state, update if there is a diff
			if !stateMember.Equal(planMember) {
				stateMemberModel := toModel(ctx, stateMember, diagnostics)
				if diagnostics.HasError() {
					continue
				}
				r.updateMember(ctx, stateMemberModel, planMemberModel, organizationID, diagnostics)
			}
		}
	}

	// Delete removed members
	for key, stateMember := range stateMembers {
		_, ok := planMembers[key]
		if !ok {
			// member is in state, but not in plan
			stateMemberModel := toModel(ctx, stateMember, diagnostics)
			if diagnostics.HasError() {
				continue
			}
			r.deleteMember(stateMemberModel, organizationID, diagnostics)
		}
	}

	// Re-read the whole org from the API to get the current state
	updatedOrganization := r.readFromApi(ctx, organizationID, diagnostics)
	if diagnostics.HasError() {
		return
	}
	diagnostics.Append(response.State.Set(ctx, *updatedOrganization)...)
}

func (r *Resource) updateMember(
	ctx context.Context,
	stateMember OrganizationMember,
	planMember OrganizationMember,
	organizationID string,
	diagnostics *diag.Diagnostics,
) {
	if planMember.InvitationPending.ValueBool() {
		// Invitations can't be updated, so while the invitation is pending the role assignments can't be changed
		// The only way to update them is by creating a new invitation with the right role-assignments.
		r.deleteInvitation(planMember, organizationID, diagnostics)
		r.createInvitation(ctx, stateMember.Email.ValueString(), planMember, organizationID, diagnostics)
	} else {
		// Add new role assignments
		planApiMember := modelToApi(ctx, planMember, organizationID, diagnostics)
		if diagnostics.HasError() {
			return
		}

		stateApiMember := modelToApi(ctx, stateMember, organizationID, diagnostics)
		if diagnostics.HasError() {
			return
		}

		add, remove := diffRoleAssignments(stateApiMember.RoleAssignments, planApiMember.RoleAssignments)

		if hasChanges(add) {
			_, err := organizationapi.AddRoleAssignments(organizationapi.AddRoleAssignmentsParams{
				API:             r.client,
				UserID:          planMember.UserID.ValueString(),
				RoleAssignments: add,
			})
			if err != nil {
				diagnostics.Append(diag.NewErrorDiagnostic("Updating member roles failed.", err.Error()))
				return
			}
		}

		if hasChanges(remove) {
			_, err := organizationapi.RemoveRoleAssignments(organizationapi.RemoveRoleAssignmentsParams{
				API:             r.client,
				UserID:          planMember.UserID.ValueString(),
				RoleAssignments: remove,
			})
			if err != nil {
				diagnostics.Append(diag.NewErrorDiagnostic("Updating member roles failed.", err.Error()))
				return
			}
		}
	}
}

func hasChanges(ra models.RoleAssignments) bool {
	if len(ra.Organization) > 0 {
		return true
	}
	if len(ra.Deployment) > 0 {
		return true
	}
	if ra.Project != nil {
		return len(ra.Project.Elasticsearch) > 0 ||
			len(ra.Project.Security) > 0 ||
			len(ra.Project.Observability) > 0
	}
	return false
}

func toModel(ctx context.Context, member types.Object, diags *diag.Diagnostics) OrganizationMember {
	var modelValue OrganizationMember
	var objectAsOptions = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false}
	diags.Append(member.As(ctx, &modelValue, objectAsOptions)...)
	return modelValue
}

func diffRoleAssignments(old, new *models.RoleAssignments) (models.RoleAssignments, models.RoleAssignments) {
	addOrganization, removeOrganization := diffOrganizationRoleAssignments(
		old.Organization,
		new.Organization,
	)

	addDeployment, removeDeployment := diffDeploymentRoleAssignments(
		old.Deployment,
		new.Deployment,
	)

	var addProject, removeProject *models.ProjectRoleAssignments
	if old.Project != nil && new.Project != nil {
		addProject, removeProject = diffProjectRoleAssignments(
			*old.Project,
			*new.Project,
		)
	} else if old.Project == nil && new.Project != nil {
		addProject, removeProject = new.Project, nil
	} else if old.Project != nil && new.Project == nil {
		addProject, removeProject = nil, old.Project
	}

	add := models.RoleAssignments{
		Organization: addOrganization,
		Deployment:   addDeployment,
		Project:      addProject,
	}
	remove := models.RoleAssignments{
		Organization: removeOrganization,
		Deployment:   removeDeployment,
		Project:      removeProject,
	}

	return add, remove
}

func diffOrganizationRoleAssignments(old, new []*models.OrganizationRoleAssignment) ([]*models.OrganizationRoleAssignment, []*models.OrganizationRoleAssignment) {
	getKey := func(ra models.OrganizationRoleAssignment) string {
		return *ra.RoleID
	}
	add := difference(new, old, getKey)
	remove := difference(old, new, getKey)
	return add, remove
}

func diffDeploymentRoleAssignments(old, new []*models.DeploymentRoleAssignment) ([]*models.DeploymentRoleAssignment, []*models.DeploymentRoleAssignment) {
	getKey := func(ra models.DeploymentRoleAssignment) string {
		var all bool
		if ra.All != nil {
			all = *ra.All
		}
		sort.Strings(ra.DeploymentIds)
		return fmt.Sprintf("%s-%t-%s", *ra.RoleID, all, strings.Join(ra.DeploymentIds, ","))
	}
	add := difference(new, old, getKey)
	remove := difference(old, new, getKey)
	return add, remove
}

func diffProjectRoleAssignments(old, new models.ProjectRoleAssignments) (*models.ProjectRoleAssignments, *models.ProjectRoleAssignments) {
	getKey := func(ra models.ProjectRoleAssignment) string {
		var all bool
		if ra.All != nil {
			all = *ra.All
		}
		sort.Strings(ra.ProjectIds)
		return fmt.Sprintf("%s-%t-%s", *ra.RoleID, all, strings.Join(ra.ProjectIds, ","))
	}

	addElasticsearch := difference(new.Elasticsearch, old.Elasticsearch, getKey)
	removeElasticsearch := difference(old.Elasticsearch, new.Elasticsearch, getKey)

	addObservability := difference(new.Observability, old.Observability, getKey)
	removeObservability := difference(old.Observability, new.Observability, getKey)

	addSecurity := difference(new.Security, old.Security, getKey)
	removeSecurity := difference(old.Security, new.Security, getKey)

	add := models.ProjectRoleAssignments{
		Elasticsearch: addElasticsearch,
		Observability: addObservability,
		Security:      addSecurity,
	}
	remove := models.ProjectRoleAssignments{
		Elasticsearch: removeElasticsearch,
		Observability: removeObservability,
		Security:      removeSecurity,
	}
	return &add, &remove
}

func difference[T interface{}](a, b []*T, getKey func(T) string) []*T {
	var diff []*T
	m := make(map[string]T)
	for _, item := range b {
		if item == nil {
			continue
		}
		key := getKey(*item)
		m[key] = *item
	}

	for _, item := range a {
		if item == nil {
			continue
		}
		key := getKey(*item)
		if _, ok := m[key]; !ok {
			diff = append(diff, item)
		}
	}

	return diff
}
