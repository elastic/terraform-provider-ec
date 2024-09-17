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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	provider "github.com/elastic/terraform-provider-ec/ec"
)


func TestOrganizationResourceAgainstMockedAPI(t *testing.T) {
	resourceName := "ec_organization.myorg"

	baseConfig := buildConfig("")
	configWithNewMember := buildConfig(addedMember)
	configWithUpdatedNewMember := buildConfig(addedMemberWithUpdate)
	configWithAddedRoles := buildConfig(memberWithNewRoles)
	configWithRemovedRoles := buildConfig(memberWithRemovedRoles)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			// The mocked calls are very important for the validity of the tests
			// For each testcase below, the correct API responses have to be mocked in here
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
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.deployment_roles.1.all_deployments", "true"),
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
					resource.TestCheckResourceAttr(resourceName, "members.newuser@example.com.deployment_roles.0.all_deployments", "true"),
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

func protoV6ProviderFactoriesWithMockClient(client *api.API) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"ec": func() (tfprotov6.ProviderServer, error) {
			return providerserver.NewProtocol6(provider.ProviderWithClient(client, "unit-tests"))(), nil
		},
	}
}
