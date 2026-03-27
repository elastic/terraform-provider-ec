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
	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/go-openapi/strfmt"
)

var orgId = new("123")

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

func getMembersFails() mock.Response {
	return mock.New404ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "GET",
			Path:   "/api/v1/organizations/123/members",
		},
		mock.NewStructBody(models.BasicFailedReply{
			Errors: []*models.BasicFailedReplyElement{
				{
					Message: new("organization-does-not-exist"),
				},
			}}),
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

func getInvitationsFails() mock.Response {
	return mock.New404ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "GET",
			Path:   "/api/v1/organizations/123/invitations",
		},
		mock.NewStructBody(models.BasicFailedReply{
			Errors: []*models.BasicFailedReplyElement{
				{
					Message: new("organization-does-not-exist"),
				},
			}}),
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

func createInvitationFails(invitation *models.OrganizationInvitation) mock.Response {
	return mock.New400ResponseAssertion(
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
		mock.NewStructBody(models.BasicFailedReply{
			Errors: []*models.BasicFailedReplyElement{
				{
					Message: new("organization.invitation_invalid_email"),
				},
			}}),
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

func deleteInvitationFails(invitation *models.OrganizationInvitation) mock.Response {
	return mock.New400ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "DELETE",
			Path:   "/api/v1/organizations/123/invitations/" + *invitation.Token,
		},
		mock.NewStructBody(models.BasicFailedReply{
			Errors: []*models.BasicFailedReplyElement{
				{
					Message: new("organization.invitation_token_invalid"),
				},
			}}),
	)
}

func buildInvitationModel(email string) *models.OrganizationInvitation {
	timestamp, _ := strfmt.ParseDateTime("2021-01-07T22:13:42.999Z")
	expiration, _ := strfmt.ParseDateTime("2023-01-07T22:13:42.999Z")
	assignments := &models.RoleAssignments{
		Deployment: nil,
		Organization: []*models.OrganizationRoleAssignment{
			{
				OrganizationID: new("123"),
				RoleID:         new("billing-admin"),
			},
		},
		Platform: nil,
		Project:  &models.ProjectRoleAssignments{},
	}
	return &models.OrganizationInvitation{
		Token:      new("invitation-token"),
		AcceptedAt: strfmt.DateTime{},
		CreatedAt:  &timestamp,
		Email:      new(email),
		Expired:    new(false),
		ExpiresAt:  &expiration,
		Organization: &models.Organization{
			ID: new("123"),
		},
		RoleAssignments: assignments,
	}
}

func addRoleAssignments(deploymentAssignments []*models.DeploymentRoleAssignment) mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "POST",
			Path:   "/api/v1/users/userid2/role_assignments",
			Body: mock.NewStructBody(models.RoleAssignments{
				Deployment: deploymentAssignments,
				Project:    &models.ProjectRoleAssignments{},
			}),
		},
		mock.NewStringBody("{}"),
	)
}

func addRoleAssignmentsFails() mock.Response {
	return mock.New400ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "POST",
			Path:   "/api/v1/users/userid2/role_assignments",
			Body: mock.NewStructBody(models.RoleAssignments{
				Deployment: []*models.DeploymentRoleAssignment{
					{
						All:            new(false),
						OrganizationID: orgId,
						RoleID:         new("deployment-editor"),
						DeploymentIds:  []string{"abc"},
					},
					{
						OrganizationID: orgId,
						RoleID:         new("deployment-viewer"),
						All:            new(true),
					},
				},
				Project: &models.ProjectRoleAssignments{},
			}),
		},
		mock.NewStructBody(models.BasicFailedReply{
			Errors: []*models.BasicFailedReplyElement{
				{
					Message: new("role_assignments.invalid_config"),
				},
			}}),
	)
}

func removeRoleAssignments(deploymentAssignments []*models.DeploymentRoleAssignment, orgAssignments []*models.OrganizationRoleAssignment) mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "DELETE",
			Path:   "/api/v1/users/userid2/role_assignments",
			Body: mock.NewStructBody(models.RoleAssignments{
				Organization: orgAssignments,
				Deployment:   deploymentAssignments,
				Project:      &models.ProjectRoleAssignments{},
			}),
		},
		mock.NewStringBody("{}"),
	)
}

func removeRoleAssignmentsFails() mock.Response {
	return mock.New400ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "DELETE",
			Path:   "/api/v1/users/userid2/role_assignments",
			Body: mock.NewStructBody(models.RoleAssignments{
				Organization: []*models.OrganizationRoleAssignment{
					{
						OrganizationID: orgId,
						RoleID:         new("organization-admin"),
					},
				},
				Deployment: []*models.DeploymentRoleAssignment{
					{
						All:            new(false),
						OrganizationID: orgId,
						RoleID:         new("deployment-editor"),
						DeploymentIds:  []string{"abc"},
					},
				},
				Project: &models.ProjectRoleAssignments{},
			}),
		},
		mock.NewStructBody(models.BasicFailedReply{
			Errors: []*models.BasicFailedReplyElement{
				{
					Message: new("role_assignments.invalid_config"),
				},
			}}),
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

func removeMemberFails() mock.Response {
	return mock.New404ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "DELETE",
			Path:   "/api/v1/organizations/123/members/userid2",
		},
		mock.NewStructBody(models.BasicFailedReply{
			Errors: []*models.BasicFailedReplyElement{
				{
					Message: new("organization.membership_not_found"),
				},
			}}),
	)
}

func buildExistingMember() *models.OrganizationMembership {
	return &models.OrganizationMembership{
		UserID:         new("userid"),
		Email:          "user@example.com",
		OrganizationID: orgId,
		RoleAssignments: &models.RoleAssignments{
			Organization: []*models.OrganizationRoleAssignment{
				{
					OrganizationID: orgId,
					RoleID:         new("billing-admin"),
				},
			},
			Deployment: []*models.DeploymentRoleAssignment{
				{
					OrganizationID: orgId,
					RoleID:         new("deployment-viewer"),
					All:            new(true),
				},
				{
					OrganizationID: orgId,
					RoleID:         new("deployment-editor"),
					DeploymentIds:  []string{"abc"},
				},
			},
			Project: &models.ProjectRoleAssignments{
				Elasticsearch: []*models.ProjectRoleAssignment{
					{
						OrganizationID: orgId,
						RoleID:         new("elasticsearch-viewer"),
						All:            new(true),
					},
					{
						OrganizationID: orgId,
						RoleID:         new("elasticsearch-developer"),
						ProjectIds:     []string{"qwe"},
					},
				},
				Observability: []*models.ProjectRoleAssignment{
					{
						OrganizationID: orgId,
						RoleID:         new("observability-viewer"),
						All:            new(true),
					},
					{
						OrganizationID: orgId,
						RoleID:         new("observability-editor"),
						ProjectIds:     []string{"rty"},
					},
				},
				Security: []*models.ProjectRoleAssignment{
					{
						OrganizationID: orgId,
						RoleID:         new("security-viewer"),
						All:            new(true),
					},
					{
						OrganizationID: orgId,
						RoleID:         new("security-editor"),
						ProjectIds:     []string{"uio"},
					},
				},
			},
		},
	}
}

func buildNewMember() *models.OrganizationMembership {
	return &models.OrganizationMembership{
		UserID:         new("userid2"),
		Email:          "newuser@example.com",
		OrganizationID: orgId,
		RoleAssignments: &models.RoleAssignments{
			Organization: []*models.OrganizationRoleAssignment{
				{
					OrganizationID: orgId,
					RoleID:         new("organization-admin"),
				},
			},
		},
	}
}
