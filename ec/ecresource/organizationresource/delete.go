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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// It is not possible to delete an organization
}

func (r *Resource) deleteMember(email string, member OrganizationMember, organizationID string, diags *diag.Diagnostics) {
	if member.InvitationPending.ValueBool() {
		r.deleteInvitation(email, organizationID, diags)
	} else {
		_, err := organizationapi.DeleteMember(organizationapi.DeleteMemberParams{
			API:            r.client,
			OrganizationID: organizationID,
			UserIDs:        []string{member.UserID.ValueString()},
		})
		if err != nil {
			diags.Append(diag.NewErrorDiagnostic("Removing organization member failed.", err.Error()))
			return
		}
	}
}

func (r *Resource) deleteInvitation(email string, organizationID string, diags *diag.Diagnostics) {
	invitations, err := organizationapi.ListInvitations(organizationapi.ListInvitationsParams{
		API:            r.client,
		OrganizationID: organizationID,
	})
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("Listing organization members failed", err.Error()))
		return
	}
	for _, invitation := range invitations.Invitations {
		if *invitation.Email == email {
			_, err := organizationapi.DeleteInvitation(organizationapi.DeleteInvitationParams{
				API:              r.client,
				OrganizationID:   organizationID,
				InvitationTokens: []string{*invitation.Token},
			})
			if err != nil {
				diags.Append(diag.NewErrorDiagnostic("Removing member invitation failed", err.Error()))
				return
			}
			return
		}
	}
}
