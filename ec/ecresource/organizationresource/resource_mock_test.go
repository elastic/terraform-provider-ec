package organizationresource_test

import (
	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/go-openapi/strfmt"
)

var orgId = ec.String("123")

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
