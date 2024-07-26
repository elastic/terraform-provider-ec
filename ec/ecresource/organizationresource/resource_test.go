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
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/go-openapi/strfmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	provider "github.com/elastic/terraform-provider-ec/ec"
)

var orgId = ec.String("123")

func TestOrganizationResourceAgainstMockedAPI(t *testing.T) {
	resourceName := "ec_organization.myorg"

	baseConfig := buildConfig("")
	configWithNewMember := buildConfig(addedMember)
	configWithUpdatedNewMember := buildConfig(addedMemberWithUpdate)
	configWithAddedRoles := buildConfig(memberWithNewRoles)
	configWithRemovedRoles := buildConfig(memberWithRemovedRoles)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			mockApi(),
		),
		Steps: []resource.TestStep{
			{
				ImportState:        true,
				ResourceName:       "ec_organization.myorg",
				ImportStateId:      "123",
				Config:             baseConfig,
				ImportStatePersist: true,
			},
			// Ensure the pre-existing member is correctly imported into the state
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
					resource.TestCheckResourceAttr(resourceName, "members.user@example.com.deployment_roles.1.for_all_deployments", "true"),

					// Elasticsearch roles
					resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_elasticsearch_roles.0.role", "developer"),
					resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_elasticsearch_roles.0.project_ids.0", "qwe"),
					resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_elasticsearch_roles.1.role", "viewer"),
					resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_elasticsearch_roles.1.for_all_projects", "true"),

					// Observability roles
					resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_observability_roles.0.role", "editor"),
					resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_observability_roles.0.project_ids.0", "rty"),
					resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_observability_roles.1.role", "viewer"),
					resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_observability_roles.1.for_all_projects", "true"),

					// Project roles
					resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_security_roles.0.role", "editor"),
					resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_security_roles.0.project_ids.0", "uio"),
					resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_security_roles.1.role", "viewer"),
					resource.TestCheckResourceAttr(resourceName, "members.user@example.com.project_security_roles.1.for_all_projects", "true"),
				),
			},
			// A newly added member should be invited to the organization
			{
				Config: configWithNewMember,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "members.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.email", "newuser@example.com"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.invitation_pending", "true"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.organization_role", "billing-admin"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.user_id", ""),
				),
			},
			// If the invited members roles are changed, the invitation is cancelled and re-sent (invitations can't be updated)
			{
				Config: configWithUpdatedNewMember,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "members.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.email", "newuser@example.com"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.invitation_pending", "true"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.organization_role", "organization-admin"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.user_id", ""),
				),
			},
			// If the invited member accepts, the next apply will just update the state with the user-id and set invitation_pending to false
			{
				Config:   configWithUpdatedNewMember,
				PlanOnly: true, // Has to be no-op plan
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "members.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.email", "newuser@example.com"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.invitation_pending", "false"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.organization_role", "organization-admin"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.user_id", "userid2"),
				),
			},
			// Adding roles to member
			{
				Config: configWithAddedRoles,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "members.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.email", "newuser@example.com"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.organization_role", "organization-admin"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.deployment_roles.0.role", "editor"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.deployment_roles.0.deployment_ids.0", "abc"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.deployment_roles.1.role", "viewer"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.deployment_roles.1.for_all_deployments", "true"),
				),
			},
			// Removing roles from member
			{
				Config: configWithRemovedRoles,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "members.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.email", "newuser@example.com"),
					resource.TestCheckNoResourceAttr(resourceName, "members.newuser@example.com.organization_role"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.deployment_roles.0.role", "viewer"),
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.deployment_roles.0.for_all_deployments", "true"),
				),
			},
			// Removing member from organization
			{
				Config: baseConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "members.%", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "members.newuser@example.com"),
				),
			},
			// Invite member
			{
				Config: configWithNewMember,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "members.%", "2"),
				),
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
	})
}

func mockApi() *api.API {
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

	return api.NewMock(
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

		// Add member
		getMembers(oneMember),
		getInvitations(nil),
		createInvitation(newUserInvitation),
		getMembers(oneMember),
		getInvitations([]*models.OrganizationInvitation{newUserInvitation}),
		getMembers(oneMember),
		getInvitations([]*models.OrganizationInvitation{newUserInvitation}),

		// Update invited member (before invitation is accepted)
		getMembers(oneMember),
		getInvitations([]*models.OrganizationInvitation{newUserInvitation}),
		getInvitations([]*models.OrganizationInvitation{newUserInvitation}),
		deleteInvitation(newUserInvitation),
		createInvitation(updatedUserInvitation),
		getMembers(oneMember),
		getInvitations([]*models.OrganizationInvitation{updatedUserInvitation}),
		getMembers(oneMember),
		getInvitations([]*models.OrganizationInvitation{updatedUserInvitation}),

		// Apply after invitation has been accepted
		getMembers([]*models.OrganizationMembership{existingMember, newMember}),
		getInvitations(nil),
		getMembers([]*models.OrganizationMembership{existingMember, newMember}),
		getInvitations(nil),

		// Add roles
		getMembers([]*models.OrganizationMembership{existingMember, newMember}),
		getInvitations(nil),
		addRoleAssignments(),
		getMembers([]*models.OrganizationMembership{existingMember, newMemberWithAddedRoles}),
		getInvitations(nil),
		getMembers([]*models.OrganizationMembership{existingMember, newMemberWithAddedRoles}),
		getInvitations(nil),

		// Removed roles
		getMembers([]*models.OrganizationMembership{existingMember, newMemberWithAddedRoles}),
		getInvitations(nil),
		removeRoleAssignments(),
		getMembers([]*models.OrganizationMembership{existingMember, newMemberWithRemovedRoles}),
		getInvitations(nil),
		getMembers([]*models.OrganizationMembership{existingMember, newMemberWithRemovedRoles}),
		getInvitations(nil),

		// Remove member
		getMembers([]*models.OrganizationMembership{existingMember, newMemberWithAddedRoles}),
		getInvitations(nil),
		removeMember(),
		getMembers([]*models.OrganizationMembership{existingMember}),
		getInvitations(nil),
		getMembers([]*models.OrganizationMembership{existingMember}),
		getInvitations(nil),

		// Add member
		getMembers(oneMember),
		getInvitations(nil),
		createInvitation(newUserInvitation),
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
	)
}

func getMembers(memberships []*models.OrganizationMembership) mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "GET",
			Path:   "/api/v1/organizations/123/members",
		},
		mock.NewStructBody(models.OrganizationMemberships{
			Members: memberships,
		}),
	)
}

func getInvitations(invitations []*models.OrganizationInvitation) mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "GET",
			Path:   "/api/v1/organizations/123/invitations",
		},
		mock.NewStructBody(models.OrganizationInvitations{
			Invitations: invitations,
		}),
	)
}

func createInvitation(invitation *models.OrganizationInvitation) mock.Response {

	return mock.New201ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "POST",
			Path:   "/api/v1/organizations/123/invitations",
			Body: mock.NewStructBody(models.OrganizationInvitationRequest{
				Emails:          []string{*invitation.Email},
				ExpiresIn:       "7d",
				RoleAssignments: invitation.RoleAssignments,
			}),
		},
		mock.NewStructBody(models.OrganizationInvitations{
			Invitations: []*models.OrganizationInvitation{
				invitation,
			},
		}),
	)
}

func deleteInvitation(invitation *models.OrganizationInvitation) mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "DELETE",
			Path:   "/api/v1/organizations/123/invitations/" + *invitation.Token,
		},
		mock.NewStringBody("{}"),
	)
}

func buildInvitationModel(email string) *models.OrganizationInvitation {
	timestamp, _ := strfmt.ParseDateTime("2021-01-07T22:13:42.999Z")
	expiration, _ := strfmt.ParseDateTime("2023-01-07T22:13:42.999Z")
	assignments := &models.RoleAssignments{
		Deployment: nil,
		Organization: []*models.OrganizationRoleAssignment{
			{
				OrganizationID: ec.String("123"),
				RoleID:         ec.String("billing-admin"),
			},
		},
		Platform: nil,
		Project:  &models.ProjectRoleAssignments{},
	}
	return &models.OrganizationInvitation{
		Token:      ec.String("invitation-token"),
		AcceptedAt: strfmt.DateTime{},
		CreatedAt:  &timestamp,
		Email:      ec.String(email),
		Expired:    ec.Bool(false),
		ExpiresAt:  &expiration,
		Organization: &models.Organization{
			ID: ec.String("123"),
		},
		RoleAssignments: assignments,
	}
}

func addRoleAssignments() mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "POST",
			Path:   "/api/v1/users/userid2/role_assignments",
			Body: mock.NewStructBody(models.RoleAssignments{
				Deployment: []*models.DeploymentRoleAssignment{
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
				},
				Project: &models.ProjectRoleAssignments{},
			}),
		},
		mock.NewStringBody("{}"),
	)
}

func removeRoleAssignments() mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "DELETE",
			Path:   "/api/v1/users/userid2/role_assignments",
			Body: mock.NewStructBody(models.RoleAssignments{
				Organization: []*models.OrganizationRoleAssignment{
					{
						OrganizationID: orgId,
						RoleID:         ec.String("organization-admin"),
					},
				},
				Deployment: []*models.DeploymentRoleAssignment{
					{
						All:            ec.Bool(false),
						OrganizationID: orgId,
						RoleID:         ec.String("deployment-editor"),
						DeploymentIds:  []string{"abc"},
					},
				},
				Project: &models.ProjectRoleAssignments{},
			}),
		},
		mock.NewStringBody("{}"),
	)
}

func removeMember() mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "DELETE",
			Path:   "/api/v1/organizations/123/members/userid2",
		},
		mock.NewStringBody("{}"),
	)
}

func buildExistingMember() *models.OrganizationMembership {
	return &models.OrganizationMembership{
		UserID:         ec.String("userid"),
		Email:          "user@example.com",
		OrganizationID: orgId,
		RoleAssignments: &models.RoleAssignments{
			Organization: []*models.OrganizationRoleAssignment{
				{
					OrganizationID: orgId,
					RoleID:         ec.String("billing-admin"),
				},
			},
			Deployment: []*models.DeploymentRoleAssignment{
				{
					OrganizationID: orgId,
					RoleID:         ec.String("deployment-viewer"),
					All:            ec.Bool(true),
				},
				{
					OrganizationID: orgId,
					RoleID:         ec.String("deployment-editor"),
					DeploymentIds:  []string{"abc"},
				},
			},
			Project: &models.ProjectRoleAssignments{
				Elasticsearch: []*models.ProjectRoleAssignment{
					{
						OrganizationID: orgId,
						RoleID:         ec.String("elasticsearch-viewer"),
						All:            ec.Bool(true),
					},
					{
						OrganizationID: orgId,
						RoleID:         ec.String("elasticsearch-developer"),
						ProjectIds:     []string{"qwe"},
					},
				},
				Observability: []*models.ProjectRoleAssignment{
					{
						OrganizationID: orgId,
						RoleID:         ec.String("observability-viewer"),
						All:            ec.Bool(true),
					},
					{
						OrganizationID: orgId,
						RoleID:         ec.String("observability-editor"),
						ProjectIds:     []string{"rty"},
					},
				},
				Security: []*models.ProjectRoleAssignment{
					{
						OrganizationID: orgId,
						RoleID:         ec.String("security-viewer"),
						All:            ec.Bool(true),
					},
					{
						OrganizationID: orgId,
						RoleID:         ec.String("security-editor"),
						ProjectIds:     []string{"uio"},
					},
				},
			},
		},
	}
}

func buildNewMember() *models.OrganizationMembership {
	return &models.OrganizationMembership{
		UserID:         ec.String("userid2"),
		Email:          "newuser@example.com",
		OrganizationID: orgId,
		RoleAssignments: &models.RoleAssignments{
			Organization: []*models.OrganizationRoleAssignment{
				{
					OrganizationID: orgId,
					RoleID:         ec.String("organization-admin"),
				},
			},
		},
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
          for_all_deployments = true
        },
        {
          role = "editor"
          deployment_ids = ["abc"]
        }
      ]

      project_elasticsearch_roles = [
        {
          role = "viewer"
          for_all_projects = true
        },
        {
          role = "developer"
          project_ids = ["qwe"]
        }
      ]

      project_observability_roles = [
        {
          role = "viewer"
          for_all_projects = true
        },
        {
          role = "editor"
          project_ids = ["rty"]
        }
      ]

      project_security_roles = [
        {
          role = "viewer"
          for_all_projects = true
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
          for_all_deployments = true
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
          for_all_deployments = true
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
