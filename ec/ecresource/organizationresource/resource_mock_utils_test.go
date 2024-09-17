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
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/go-openapi/strfmt"
)

var orgId = ec.String("123")

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
