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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func apiToModel(ctx context.Context, member models.OrganizationMembership, invitationPending bool, diagnostics *diag.Diagnostics) *OrganizationMember {
	organizationRole := organizationRoleApiToModel(member)

	deploymentRoles := deploymentRolesApiToModel(ctx, member, diagnostics)
	if diagnostics.HasError() {
		return nil
	}

	projectElasticsearchRoles := elasticsearchRolesApiToModel(ctx, member, diagnostics)
	if diagnostics.HasError() {
		return nil
	}

	projectObservabilityRoles := observabilityRolesApiToModel(ctx, member, diagnostics)
	if diagnostics.HasError() {
		return nil
	}

	projectSecurityRoles := securityRolesApiToModel(ctx, member, diagnostics)
	if diagnostics.HasError() {
		return nil
	}

	return &OrganizationMember{
		Email:                     types.StringValue(member.Email),
		InvitationPending:         types.BoolValue(invitationPending),
		UserID:                    types.StringValue(nilToEmpty(member.UserID)),
		OrganizationRole:          organizationRole,
		DeploymentRoles:           *deploymentRoles,
		ProjectElasticsearchRoles: *projectElasticsearchRoles,
		ProjectObservabilityRoles: *projectObservabilityRoles,
		ProjectSecurityRoles:      *projectSecurityRoles,
	}
}

func nilToEmpty(id *string) string {
	if id == nil {
		return ""
	}
	return *id
}

func organizationRoleApiToModel(member models.OrganizationMembership) types.String {
	if member.RoleAssignments != nil &&
		member.RoleAssignments.Organization != nil &&
		len(member.RoleAssignments.Organization) > 0 &&
		member.RoleAssignments.Organization[0] != nil &&
		member.RoleAssignments.Organization[0].RoleID != nil {
		id := member.RoleAssignments.Organization[0].RoleID
		return types.StringValue(*id)
	} else {
		return types.StringNull()
	}
}

func deploymentRolesApiToModel(ctx context.Context, member models.OrganizationMembership, diagnostics *diag.Diagnostics) *types.Set {
	var result []DeploymentRoleAssignment
	if member.RoleAssignments != nil {
		for _, roleAssignment := range member.RoleAssignments.Deployment {
			deploymentIds, diags := types.SetValueFrom(ctx, types.StringType, roleAssignment.DeploymentIds)
			if diags.HasError() {
				diagnostics.Append(diags...)
				return nil
			}
			applicationRoles, diags := types.SetValueFrom(ctx, types.StringType, roleAssignment.ApplicationRoles)
			if diags.HasError() {
				diagnostics.Append(diags...)
				return nil
			}

			result = append(result, DeploymentRoleAssignment{
				Role:              types.StringValue(roleApiToModel(*roleAssignment.RoleID, deployment)),
				ForAllDeployments: forAllApiToModel(roleAssignment.All),
				DeploymentIDs:     deploymentIds,
				ApplicationRoles:  applicationRoles,
			})
		}
	}
	roleAssignments, diags := types.SetValueFrom(ctx, deploymentRoleAssignmentsSchema().NestedObject.GetAttributes().Type(), result)
	if diags.HasError() {
		diagnostics.Append(diags...)
		return nil
	}

	return &roleAssignments
}

func elasticsearchRolesApiToModel(ctx context.Context, member models.OrganizationMembership, diagnostics *diag.Diagnostics) *types.Set {
	rolesSchema := projectElasticsearchRolesSchema()
	if member.RoleAssignments != nil && member.RoleAssignments.Project != nil && member.RoleAssignments.Project.Elasticsearch != nil {
		return rolesApiToModel(ctx, member.RoleAssignments.Project.Elasticsearch, rolesSchema, "elasticsearch", diagnostics)
	} else {
		return emptySet()
	}
}

func observabilityRolesApiToModel(ctx context.Context, member models.OrganizationMembership, diagnostics *diag.Diagnostics) *types.Set {
	rolesSchema := projectObservabilityRolesSchema()
	if member.RoleAssignments != nil && member.RoleAssignments.Project != nil && member.RoleAssignments.Project.Observability != nil {
		return rolesApiToModel(ctx, member.RoleAssignments.Project.Observability, rolesSchema, "observability", diagnostics)
	} else {
		return emptySet()
	}
}

func securityRolesApiToModel(ctx context.Context, member models.OrganizationMembership, diagnostics *diag.Diagnostics) *types.Set {
	rolesSchema := projectSecurityRolesSchema()
	if member.RoleAssignments != nil && member.RoleAssignments.Project != nil && member.RoleAssignments.Project.Security != nil {
		return rolesApiToModel(ctx, member.RoleAssignments.Project.Security, rolesSchema, "security", diagnostics)
	} else {
		return emptySet()
	}
}

func emptySet() *types.Set {
	value := types.SetValueMust(projectRoleAssignmentSchema([]string{}).Type(), []attr.Value{})
	return &value
}

func rolesApiToModel(
	ctx context.Context,
	apiRoleAssignments []*models.ProjectRoleAssignment,
	schema schema.SetNestedAttribute,
	roleType RoleType,
	diagnostics *diag.Diagnostics,
) *types.Set {
	var result []ProjectRoleAssignment

	for _, roleAssignment := range apiRoleAssignments {
		projectIds, diags := types.SetValueFrom(ctx, types.StringType, roleAssignment.ProjectIds)
		if diags.HasError() {
			diagnostics.Append(diags...)
			return nil
		}
		applicationRoles, diags := types.SetValueFrom(ctx, types.StringType, roleAssignment.ApplicationRoles)
		if diags.HasError() {
			diagnostics.Append(diags...)
			return nil
		}

		result = append(result, ProjectRoleAssignment{
			Role:             types.StringValue(roleApiToModel(*roleAssignment.RoleID, roleType)),
			ForAllProjects:   forAllApiToModel(roleAssignment.All),
			ProjectIDs:       projectIds,
			ApplicationRoles: applicationRoles,
		})
	}

	roleAssignments, diags := types.SetValueFrom(ctx, schema.NestedObject.GetAttributes().Type(), result)
	if diags.HasError() {
		diagnostics.Append(diags...)
		return nil
	}

	return &roleAssignments
}

func forAllApiToModel(apiAll *bool) types.Bool {
	if apiAll == nil {
		return types.BoolValue(false)
	}
	return types.BoolValue(*apiAll)
}
