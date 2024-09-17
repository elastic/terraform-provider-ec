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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	diagnostics := &response.Diagnostics

	var organizationID string
	diagnostics.Append(request.State.GetAttribute(ctx, path.Root("id"), &organizationID)...)
	if diagnostics.HasError() {
		return
	}

	organization := r.readFromApi(ctx, organizationID, diagnostics)
	if diagnostics.HasError() {
		return
	}

	diagnostics.Append(response.State.Set(ctx, organization)...)
}

func (r *Resource) readFromApi(ctx context.Context, organizationID string, diagnostics *diag.Diagnostics) *Organization {
	members, err := organizationapi.ListMembers(organizationapi.ListMembersParams{
		API:            r.client,
		OrganizationID: organizationID,
	})
	if err != nil {
		diagnostics.Append(diag.NewErrorDiagnostic("Listing organization members failed", err.Error()))
		return nil
	}

	modelMembers := make(map[string]OrganizationMember)
	for _, member := range members.Members {
		model := apiToModel(ctx, *member, false, diagnostics)
		if diagnostics.HasError() {
			return nil
		}
		modelMembers[model.Email.ValueString()] = *model
	}

	// Members that were invited, but have not yet accepted, are listed as invitations
	invitations, err := organizationapi.ListInvitations(organizationapi.ListInvitationsParams{
		API:            r.client,
		OrganizationID: organizationID,
	})
	if err != nil {
		diagnostics.Append(diag.NewErrorDiagnostic("Listing organization members failed", err.Error()))
		return nil
	}

	for _, invitation := range invitations.Invitations {
		model := apiToModel(ctx, models.OrganizationMembership{
			Email:           *invitation.Email,
			OrganizationID:  invitation.Organization.ID,
			RoleAssignments: invitation.RoleAssignments,
		}, true, diagnostics)
		if diagnostics.HasError() {
			return nil
		}
		modelMembers[model.Email.ValueString()] = *model
	}

	membersMapValue, diags := types.MapValueFrom(ctx, organizationMembersSchema().NestedObject.GetAttributes().Type(), modelMembers)
	if diags.HasError() {
		diagnostics.Append(diags...)
		return nil
	}

	return &Organization{
		ID:      types.StringValue(organizationID),
		Members: membersMapValue,
	}
}
