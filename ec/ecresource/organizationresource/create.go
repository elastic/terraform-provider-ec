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
	"github.com/elastic/cloud-sdk-go/pkg/api/organizationapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	// It is not possible to create an organization, it already exists
	// Instead, just import the already existing organization
	response.Diagnostics.AddError("organization already exists", "please import the organization using terraform import")
}

func (r *Resource) createInvitation(ctx context.Context, email string, plan OrganizationMember, organizationID string, diagnostics *diag.Diagnostics) *OrganizationMember {
	apiModel := modelToApi(ctx, plan, organizationID, diagnostics)
	if diagnostics.HasError() {
		return nil
	}

	invitations, err := organizationapi.CreateInvitation(organizationapi.CreateInvitationParams{
		API:             r.client,
		OrganizationID:  organizationID,
		Emails:          []string{email},
		ExpiresIn:       "7d",
		RoleAssignments: apiModel.RoleAssignments,
	})
	if err != nil {
		diagnostics.Append(diag.NewErrorDiagnostic("Failed to create invitation", err.Error()))
		return nil
	}

	invitation := invitations.Invitations[0]
	organizationMember := apiToModel(ctx, models.OrganizationMembership{
		Email:           *invitation.Email,
		OrganizationID:  invitation.Organization.ID,
		RoleAssignments: invitation.RoleAssignments,
	}, true, diagnostics)

	return organizationMember
}
