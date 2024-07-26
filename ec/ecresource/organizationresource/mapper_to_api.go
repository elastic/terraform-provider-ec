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
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"sort"
)

func modelToApi(ctx context.Context, m OrganizationMember, organizationID string, diagnostics *diag.Diagnostics) *models.OrganizationMembership {
	// org
	var apiOrgRoleAssignments []*models.OrganizationRoleAssignment
	if !m.OrganizationRole.IsNull() && !m.OrganizationRole.IsUnknown() {
		apiOrgRoleAssignments = append(apiOrgRoleAssignments, &models.OrganizationRoleAssignment{
			OrganizationID: ec.String(organizationID),
			RoleID:         m.OrganizationRole.ValueStringPointer(),
		})
	}

	// deployment
	var modelDeploymentRoles []DeploymentRoleAssignment
	diags := m.DeploymentRoles.ElementsAs(ctx, &modelDeploymentRoles, false)
	if diags.HasError() {
		diagnostics.Append(diags...)
		return nil
	}
	var apiDeploymentRoleAssignments []*models.DeploymentRoleAssignment
	for _, roleAssignment := range modelDeploymentRoles {

		var deploymentIds []string
		diags = roleAssignment.DeploymentIDs.ElementsAs(ctx, &deploymentIds, false)
		if diags.HasError() {
			diagnostics.Append(diags...)
			return nil
		}
		sort.Strings(deploymentIds)

		var applicationRoles []string
		diags = roleAssignment.ApplicationRoles.ElementsAs(ctx, &applicationRoles, false)
		if diags.HasError() {
			diagnostics.Append(diags...)
			return nil
		}
		sort.Strings(applicationRoles)

		apiDeploymentRoleAssignments = append(apiDeploymentRoleAssignments, &models.DeploymentRoleAssignment{
			OrganizationID:   ec.String(organizationID),
			RoleID:           roleModelToApi(roleAssignment.Role.ValueString(), Deployment),
			All:              roleAssignment.ForAllDeployments.ValueBoolPointer(),
			DeploymentIds:    deploymentIds,
			ApplicationRoles: applicationRoles,
		})
	}

	// elasticsearch
	apiElasticsearchRoles := projectRolesModelToApi(ctx, m.ProjectElasticsearchRoles, ProjectElasticsearch, organizationID, diagnostics)
	if diagnostics.HasError() {
		return nil
	}

	// observability
	apiObservabilityRoles := projectRolesModelToApi(ctx, m.ProjectObservabilityRoles, ProjectObservability, organizationID, diagnostics)
	if diagnostics.HasError() {
		return nil
	}

	// security
	apiSecurityRoles := projectRolesModelToApi(ctx, m.ProjectSecurityRoles, ProjectSecurity, organizationID, diagnostics)
	if diagnostics.HasError() {
		return nil
	}

	apiRoleAssignments := models.RoleAssignments{
		Organization: apiOrgRoleAssignments,
		Deployment:   apiDeploymentRoleAssignments,
		Project: &models.ProjectRoleAssignments{
			Elasticsearch: apiElasticsearchRoles,
			Observability: apiObservabilityRoles,
			Security:      apiSecurityRoles,
		},
	}
	return &models.OrganizationMembership{
		Email:           m.Email.ValueString(),
		UserID:          m.UserID.ValueStringPointer(),
		RoleAssignments: &apiRoleAssignments,
	}
}

func projectRolesModelToApi(ctx context.Context, roles types.Set, roleType RoleType, organizationID string, diagnostics *diag.Diagnostics) []*models.ProjectRoleAssignment {
	var modelRoles []ProjectRoleAssignment
	diags := roles.ElementsAs(ctx, &modelRoles, false)
	if diags.HasError() {
		diagnostics.Append(diags...)
		return nil
	}
	var apiRoles []*models.ProjectRoleAssignment
	for _, roleAssignment := range modelRoles {

		var projectIds []string
		diags = roleAssignment.ProjectIDs.ElementsAs(ctx, &projectIds, false)
		if diags.HasError() {
			diagnostics.Append(diags...)
			return nil
		}
		sort.Strings(projectIds)

		var applicationRoles []string
		diags = roleAssignment.ApplicationRoles.ElementsAs(ctx, &applicationRoles, false)
		if diags.HasError() {
			diagnostics.Append(diags...)
			return nil
		}
		sort.Strings(applicationRoles)

		apiRoles = append(apiRoles, &models.ProjectRoleAssignment{
			OrganizationID:   ec.String(organizationID),
			RoleID:           roleModelToApi(roleAssignment.Role.ValueString(), roleType),
			All:              roleAssignment.ForAllProjects.ValueBoolPointer(),
			ProjectIds:       projectIds,
			ApplicationRoles: applicationRoles,
		})
	}
	return apiRoles
}
