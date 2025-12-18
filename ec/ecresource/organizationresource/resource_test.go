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

package organizationresource_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	provider "github.com/elastic/terraform-provider-ec/ec"
)

func Test(t *testing.T) {
	resourceName := "ec_organization.myorg"

	baseConfig := buildConfig("")
	configWithNewMember := buildConfig(addedMember)
	configWithUpdatedNewMember := buildConfig(addedMemberWithUpdate)
	configWithAddedRoles := buildConfig(memberWithNewRoles)
	configWithRemovedRoles := buildConfig(memberWithRemovedRoles)
	configWithOverlappingRoles := buildConfig(memberWithOverlappingRoles)

	newUserInvitation := buildInvitationModel("newuser@example.com")
	updatedUserInvitation := buildInvitationModel("newuser@example.com")
	updatedUserInvitation.RoleAssignments.Organization[0].RoleID = ec.String("organization-admin")

	existingMember := buildExistingMember()
	newMember := buildNewMember()
	oneMember := []*models.OrganizationMembership{existingMember}

	newMemberWithAddedRoles := buildNewMember()
	newMemberWithAddedRoles.RoleAssignments.Deployment = []*models.DeploymentRoleAssignment{
		{
			All:            ec.Bool(false),
			OrganizationID: orgId,
			RoleID:         ec.String("deployment-editor"),
			DeploymentIds:  []string{"abc"},
		},
		{
			OrganizationID: orgId,
			RoleID:         ec.String("deployment-viewer"),
			All:            ec.Bool(true),
		},
	}
	newMemberWithRemovedRoles := buildNewMember()
	newMemberWithRemovedRoles.RoleAssignments.Organization = []*models.OrganizationRoleAssignment{}
	newMemberWithRemovedRoles.RoleAssignments.Deployment = []*models.DeploymentRoleAssignment{
		{
			OrganizationID: orgId,
			RoleID:         ec.String("deployment-viewer"),
			All:            ec.Bool(true),
		},
	}
	newMemberWithOverlappingRoles := buildNewMember()
	newMemberWithOverlappingRoles.RoleAssignments.Deployment = []*models.DeploymentRoleAssignment{
		{
			All:            ec.Bool(false),
			OrganizationID: orgId,
			RoleID:         ec.String("deployment-editor"),
			DeploymentIds:  []string{"abc", "def"},
		},
		{
			OrganizationID: orgId,
			RoleID:         ec.String("deployment-viewer"),
			All:            ec.Bool(true),
		},
	}

	tests := []struct {
		name    string
		steps   []resource.TestStep
		apiMock []mock.Response
	}{
		{
			name: "import should correctly set the state",
			steps: []resource.TestStep{
				{
					ImportState:        true,
					ResourceName:       "ec_organization.myorg",
					ImportStateId:      "123",
					Config:             baseConfig,
					ImportStatePersist: true,
				},
				{
					Config: baseConfig,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "id", "123"),
						resource.TestCheckResourceAttr(resourceName, "members.%", "1"),
						resource.TestCheckResourceAttr(resourceName, "members.user@example.com.email", "user@example.com"),
						resource.TestCheckResourceAttr(resourceName, "members.user@example.com.invitation_pending", "false"),
						resource.TestCheckResourceAttr(resourceName, "members.user@example.com.user_id", "userid"),

						// Organization role
						resource.TestCheckResourceAttr(resourceName, "members.user@example.com.organization_role", "billing-admin"),

						// Deployment roles
						resource.TestCheckResourceAttr(resourceName, "members.user@example.com.deployment_roles.0.role", "editor"),
						resource.TestCheckResourceAttr(resourceName, "members.user@example.com.deployment_roles.0.deployment_ids.0", "abc"),
						resource.TestCheckResourceAttr(resourceName, "members.user@example.com.deployment_roles.1.role", "viewer"),
						resource.TestCheckResourceAttr(resourceName, "members.user@example.com.deployment_roles.1.all_deployments", "true"),

						// Elasticsearch roles
						resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_elasticsearch_roles.0.role", "developer"),
						resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_elasticsearch_roles.0.project_ids.0", "qwe"),
						resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_elasticsearch_roles.1.role", "viewer"),
						resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_elasticsearch_roles.1.all_projects", "true"),

						// Observability roles
						resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_observability_roles.0.role", "editor"),
						resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_observability_roles.0.project_ids.0", "rty"),
						resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_observability_roles.1.role", "viewer"),
						resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_observability_roles.1.all_projects", "true"),

						// Project roles
						resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_security_roles.0.role", "editor"),
						resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_security_roles.0.project_ids.0", "uio"),
						resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_security_roles.1.role", "viewer"),
						resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_security_roles.1.all_projects", "true"),
					),
				},
			},
			apiMock: []mock.Response{
				// Import
				getMembers(oneMember),
				getInvitations(nil),
				getMembers(oneMember),
				getInvitations(nil),
				// Apply
				getMembers(oneMember),
				getInvitations(nil),
				getMembers(oneMember),
				getInvitations(nil),
			},
		},
		{
			name: "a newly added member should be invited to the organization",
			steps: []resource.TestStep{
				{
					ImportState:        true,
					ResourceName:       "ec_organization.myorg",
					ImportStateId:      "123",
					Config:             baseConfig,
					ImportStatePersist: true,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "members.%", "1"),
					),
				},
				{
					Config: configWithNewMember,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "members.%", "2"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.email", "newuser@example.com"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.invitation_pending", "true"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.organization_role", "billing-admin"),
					),
				},
			},
			apiMock: []mock.Response{
				// Import
				getMembers(oneMember),
				getInvitations(nil),
				getMembers(oneMember),
				getInvitations(nil),
				// Apply
				getMembers(oneMember),
				getInvitations(nil),
				createInvitation(newUserInvitation),
				getMembers(oneMember),
				getInvitations([]*models.OrganizationInvitation{newUserInvitation}),
				getMembers(oneMember),
				getInvitations([]*models.OrganizationInvitation{newUserInvitation}),
			},
		},
		{
			name: "if the invited members roles are changed, the invitation should be cancelled and re-sent (invitations can't be updated)",
			steps: []resource.TestStep{
				{
					ImportState:        true,
					ResourceName:       "ec_organization.myorg",
					ImportStateId:      "123",
					Config:             configWithNewMember,
					ImportStatePersist: true,
				},
				{
					Config: configWithUpdatedNewMember,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "members.%", "2"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.email", "newuser@example.com"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.invitation_pending", "true"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.organization_role", "organization-admin"),
					),
				},
			},
			apiMock: []mock.Response{
				// Import
				getMembers(oneMember),
				getInvitations([]*models.OrganizationInvitation{newUserInvitation}),
				getMembers(oneMember),
				getInvitations([]*models.OrganizationInvitation{newUserInvitation}),
				// Update
				getMembers(oneMember),
				getInvitations([]*models.OrganizationInvitation{newUserInvitation}),
				getInvitations([]*models.OrganizationInvitation{newUserInvitation}),
				deleteInvitation(newUserInvitation),
				createInvitation(updatedUserInvitation),
				getMembers(oneMember),
				getInvitations([]*models.OrganizationInvitation{updatedUserInvitation}),
				getMembers(oneMember),
				getInvitations([]*models.OrganizationInvitation{updatedUserInvitation}),
			},
		},
		{
			name: "if the invited member accepts, the next apply should just update the state with the user-id and set invitation_pending to false",
			steps: []resource.TestStep{
				{
					ImportState:        true,
					ResourceName:       "ec_organization.myorg",
					ImportStateId:      "123",
					Config:             configWithUpdatedNewMember,
					ImportStatePersist: true,
				},
				{
					Config: configWithUpdatedNewMember,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "members.%", "2"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.email", "newuser@example.com"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.invitation_pending", "false"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.organization_role", "organization-admin"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.user_id", "userid2"),
					),
				},
			},
			apiMock: []mock.Response{
				// Import
				getMembers(oneMember),
				getInvitations([]*models.OrganizationInvitation{updatedUserInvitation}),
				getMembers(oneMember),
				getInvitations([]*models.OrganizationInvitation{updatedUserInvitation}),
				// Plan
				getMembers([]*models.OrganizationMembership{existingMember, newMember}),
				getInvitations(nil),
				getMembers([]*models.OrganizationMembership{existingMember, newMember}),
				getInvitations(nil),
			},
		},
		{
			name: "adding roles to member",
			steps: []resource.TestStep{
				{
					ImportState:        true,
					ResourceName:       "ec_organization.myorg",
					ImportStateId:      "123",
					Config:             configWithUpdatedNewMember,
					ImportStatePersist: true,
				},
				{
					Config: configWithAddedRoles,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "members.%", "2"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.email", "newuser@example.com"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.organization_role", "organization-admin"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.deployment_roles.0.role", "editor"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.deployment_roles.0.deployment_ids.0", "abc"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.deployment_roles.1.role", "viewer"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.deployment_roles.1.all_deployments", "true"),
					),
				},
			},
			apiMock: []mock.Response{
				// Import
				getMembers([]*models.OrganizationMembership{existingMember, newMember}),
				getInvitations(nil),
				getMembers([]*models.OrganizationMembership{existingMember, newMember}),
				getInvitations(nil),
				// Apply
				getMembers([]*models.OrganizationMembership{existingMember, newMember}),
				getInvitations(nil),
				addRoleAssignments(newMemberWithAddedRoles.RoleAssignments.Deployment),
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithAddedRoles}),
				getInvitations(nil),
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithAddedRoles}),
				getInvitations(nil),
			},
		},
		{
			name: "removing roles from member should work",
			steps: []resource.TestStep{
				{
					ImportState:        true,
					ResourceName:       "ec_organization.myorg",
					ImportStateId:      "123",
					Config:             configWithAddedRoles,
					ImportStatePersist: true,
				},
				{
					Config: configWithRemovedRoles,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "members.%", "2"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.email", "newuser@example.com"),
						resource.TestCheckNoResourceAttr(resourceName, "members.newuser@example.com.organization_role"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.deployment_roles.0.role", "viewer"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.deployment_roles.0.all_deployments", "true"),
					),
				},
			},
			apiMock: []mock.Response{
				// Import
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithAddedRoles}),
				getInvitations(nil),
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithAddedRoles}),
				getInvitations(nil),
				// Apply
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithAddedRoles}),
				getInvitations(nil),
				removeRoleAssignments([]*models.DeploymentRoleAssignment{
					newMemberWithAddedRoles.RoleAssignments.Deployment[0],
				}, []*models.OrganizationRoleAssignment{
					{
						OrganizationID: orgId,
						RoleID:         ec.String("organization-admin"),
					},
				}),
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithRemovedRoles}),
				getInvitations(nil),
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithRemovedRoles}),
				getInvitations(nil),
			},
		},
		{
			name: "overlapping roles should first remove the existing assignments before adding the new ones",
			steps: []resource.TestStep{
				{
					ImportState:        true,
					ResourceName:       "ec_organization.myorg",
					ImportStateId:      "123",
					Config:             configWithAddedRoles,
					ImportStatePersist: true,
				},
				{
					Config: configWithOverlappingRoles,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "members.%", "2"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.email", "newuser@example.com"),
						resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.organization_role", "organization-admin"),
						resource.TestCheckTypeSetElemNestedAttrs(resourceName, "members.newuser@example.com.deployment_roles.*", map[string]string{
							"role":            "viewer",
							"all_deployments": "true",
						}),
						resource.TestCheckTypeSetElemNestedAttrs(resourceName, "members.newuser@example.com.deployment_roles.*", map[string]string{
							"role":             "editor",
							"all_deployments":  "false",
							"deployment_ids.0": "abc",
							"deployment_ids.1": "def",
						}),
					),
				},
			},
			apiMock: []mock.Response{
				// Import
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithAddedRoles}),
				getInvitations(nil),
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithAddedRoles}),
				getInvitations(nil),
				// Apply
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithAddedRoles}),
				getInvitations(nil),
				removeRoleAssignments([]*models.DeploymentRoleAssignment{
					newMemberWithAddedRoles.RoleAssignments.Deployment[0],
				}, nil),
				addRoleAssignments([]*models.DeploymentRoleAssignment{
					newMemberWithOverlappingRoles.RoleAssignments.Deployment[0],
				}),
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithOverlappingRoles}),
				getInvitations(nil),
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithOverlappingRoles}),
				getInvitations(nil),
			},
		},
		{
			name: "remove member from organization should work",
			steps: []resource.TestStep{
				{
					ImportState:        true,
					ResourceName:       "ec_organization.myorg",
					ImportStateId:      "123",
					Config:             configWithRemovedRoles,
					ImportStatePersist: true,
				},
				{
					Config: baseConfig,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "members.%", "1"),
						resource.TestCheckNoResourceAttr(resourceName, "members.newuser@example.com"),
					),
				},
			},
			apiMock: []mock.Response{
				// Import
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithRemovedRoles}),
				getInvitations(nil),
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithRemovedRoles}),
				getInvitations(nil),
				// Apply
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithAddedRoles}),
				getInvitations(nil),
				removeMember(),
				getMembers([]*models.OrganizationMembership{existingMember}),
				getInvitations(nil),
				getMembers([]*models.OrganizationMembership{existingMember}),
				getInvitations(nil),
			},
		},
		{
			name: "un-invite member before the member accepted the invitation should work",
			steps: []resource.TestStep{
				{
					ImportState:        true,
					ResourceName:       "ec_organization.myorg",
					ImportStateId:      "123",
					Config:             configWithNewMember,
					ImportStatePersist: true,
				},
				// Un-invite member (where the member is removed before they have accepted the invitation)
				{
					Config: baseConfig,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "members.%", "1"),
						resource.TestCheckNoResourceAttr(resourceName, "members.newuser@example.com"),
					),
				},
			},
			apiMock: []mock.Response{
				// Import
				getMembers(oneMember),
				getInvitations([]*models.OrganizationInvitation{newUserInvitation}),
				getMembers(oneMember),
				getInvitations([]*models.OrganizationInvitation{newUserInvitation}),
				// Remove member before invitation was accepted (cancelling invitation)
				getMembers(oneMember),
				getInvitations([]*models.OrganizationInvitation{newUserInvitation}),
				getInvitations([]*models.OrganizationInvitation{newUserInvitation}),
				deleteInvitation(newUserInvitation),
				getMembers(oneMember),
				getInvitations(nil),
				getMembers(oneMember),
				getInvitations(nil),
			},
		},
		{
			name: "show API error if import fails because organization does not exist",
			steps: []resource.TestStep{
				{
					ImportState:   true,
					ResourceName:  "ec_organization.myorg",
					ImportStateId: "123",
					Config:        baseConfig,
					ExpectError:   regexp.MustCompile("organization-does-not-exist"),
				},
			},
			apiMock: []mock.Response{
				// Import
				getMembersFails(),
			},
		},
		{
			name: "show API error if import fails because invitations could not be listed",
			steps: []resource.TestStep{
				{
					ImportState:   true,
					ResourceName:  "ec_organization.myorg",
					ImportStateId: "123",
					Config:        baseConfig,
					ExpectError:   regexp.MustCompile("organization-does-not-exist"),
				},
			},
			apiMock: []mock.Response{
				// Import
				getMembers(oneMember),
				getInvitationsFails(),
			},
		},
		{
			name: "show API error if inviting a member fails due to invalid config",
			steps: []resource.TestStep{
				{
					ImportState:        true,
					ResourceName:       "ec_organization.myorg",
					ImportStateId:      "123",
					Config:             baseConfig,
					ImportStatePersist: true,
				},
				{
					Config:      configWithNewMember,
					ExpectError: regexp.MustCompile("organization.invitation_invalid_email"),
				},
			},
			apiMock: []mock.Response{
				// Import
				getMembers(oneMember),
				getInvitations(nil),
				getMembers(oneMember),
				getInvitations(nil),
				// Apply
				getMembers(oneMember),
				getInvitations(nil),
				createInvitationFails(newUserInvitation),
			},
		},
		{
			name: "show API error if adding roles fails",
			steps: []resource.TestStep{
				{
					ImportState:        true,
					ResourceName:       "ec_organization.myorg",
					ImportStateId:      "123",
					Config:             configWithUpdatedNewMember,
					ImportStatePersist: true,
				},
				{
					Config:      configWithAddedRoles,
					ExpectError: regexp.MustCompile("role_assignments.invalid_config"),
				},
			},
			apiMock: []mock.Response{
				// Import
				getMembers([]*models.OrganizationMembership{existingMember, newMember}),
				getInvitations(nil),
				getMembers([]*models.OrganizationMembership{existingMember, newMember}),
				getInvitations(nil),
				// Apply
				getMembers([]*models.OrganizationMembership{existingMember, newMember}),
				getInvitations(nil),
				addRoleAssignmentsFails(),
			},
		},
		{
			name: "show API error if removing roles fails due to API error",
			steps: []resource.TestStep{
				{
					ImportState:        true,
					ResourceName:       "ec_organization.myorg",
					ImportStateId:      "123",
					Config:             configWithAddedRoles,
					ImportStatePersist: true,
				},
				{
					Config:      configWithRemovedRoles,
					ExpectError: regexp.MustCompile("role_assignments.invalid_config"),
				},
			},
			apiMock: []mock.Response{
				// Import
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithAddedRoles}),
				getInvitations(nil),
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithAddedRoles}),
				getInvitations(nil),
				// Apply
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithAddedRoles}),
				getInvitations(nil),
				removeRoleAssignmentsFails(),
			},
		},
		{
			name: "show API error if remove member fails",
			steps: []resource.TestStep{
				{
					ImportState:        true,
					ResourceName:       "ec_organization.myorg",
					ImportStateId:      "123",
					Config:             configWithRemovedRoles,
					ImportStatePersist: true,
				},
				{
					Config:      baseConfig,
					ExpectError: regexp.MustCompile("organization.membership_not_found"),
				},
			},
			apiMock: []mock.Response{
				// Import
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithRemovedRoles}),
				getInvitations(nil),
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithRemovedRoles}),
				getInvitations(nil),
				// Apply
				getMembers([]*models.OrganizationMembership{existingMember, newMemberWithAddedRoles}),
				getInvitations(nil),
				removeMemberFails(),
			},
		},
		{
			name: "show API error if invitation delete fails",
			steps: []resource.TestStep{
				{
					ImportState:        true,
					ResourceName:       "ec_organization.myorg",
					ImportStateId:      "123",
					Config:             configWithNewMember,
					ImportStatePersist: true,
				},
				{
					Config:      baseConfig,
					ExpectError: regexp.MustCompile("organization.invitation_token_invalid"),
				},
			},
			apiMock: []mock.Response{
				// Import
				getMembers(oneMember),
				getInvitations([]*models.OrganizationInvitation{newUserInvitation}),
				getMembers(oneMember),
				getInvitations([]*models.OrganizationInvitation{newUserInvitation}),
				// Remove member before invitation was accepted (cancelling invitation)
				getMembers(oneMember),
				getInvitations([]*models.OrganizationInvitation{newUserInvitation}),
				getInvitations([]*models.OrganizationInvitation{newUserInvitation}),
				deleteInvitationFails(newUserInvitation),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
					api.NewMock(test.apiMock...),
				),
				Steps: test.steps,
			})
		})
	}
}

func buildConfig(newUser string) string {
	return fmt.Sprintf(`
resource "ec_organization" "myorg" {
  members = {
    "user@example.com" = {
      organization_role = "billing-admin"
      
      deployment_roles = [
        {
          role = "viewer"
          all_deployments = true
        },
        {
          role = "editor"
          deployment_ids = ["abc"]
        }
      ]

      project_elasticsearch_roles = [
        {
          role = "viewer"
          all_projects = true
        },
        {
          role = "developer"
          project_ids = ["qwe"]
        }
      ]

      project_observability_roles = [
        {
          role = "viewer"
          all_projects = true
        },
        {
          role = "editor"
          project_ids = ["rty"]
        }
      ]

      project_security_roles = [
        {
          role = "viewer"
          all_projects = true
        },
        {
          role = "editor"
          project_ids = ["uio"]
        }
      ]
    }
    %s
  }
}
`, newUser)
}

const addedMember = `
    "newuser@example.com" = {
      organization_role = "billing-admin"
    }
`

const addedMemberWithUpdate = `
    "newuser@example.com" = {
      organization_role = "organization-admin"
    }
`

const memberWithNewRoles = `
    "newuser@example.com" = {
      organization_role = "organization-admin"

      deployment_roles = [
        {
          role = "viewer"
          all_deployments = true
        },
        {
          role = "editor"
          deployment_ids = ["abc"]
        }
      ]
    }
`

const memberWithRemovedRoles = `
    "newuser@example.com" = {
      deployment_roles = [
        {
          role = "viewer"
          all_deployments = true
        }
      ]
    }
`

const memberWithOverlappingRoles = `
    "newuser@example.com" = {
      organization_role = "organization-admin"

      deployment_roles = [
        {
          role = "viewer"
          all_deployments = true
        },
        {
          role = "editor"
          deployment_ids = ["abc", "def"]
        }
      ]
    }
`

func protoV6ProviderFactoriesWithMockClient(client *api.API) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"ec": func() (tfprotov6.ProviderServer, error) {
			return providerserver.NewProtocol6(provider.ProviderWithClient(client, "unit-tests"))(), nil
		},
	}
}
