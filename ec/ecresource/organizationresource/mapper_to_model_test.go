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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"testing"
)

func TestEmptyDataCanBeMappedWithoutPanic(t *testing.T) {
	tests := []struct {
		name string
		data models.OrganizationMembership
	}{
		{
			name: "Empty membership",
			data: models.OrganizationMembership{},
		},
		{
			name: "Empty role assignments",
			data: models.OrganizationMembership{
				RoleAssignments: &models.RoleAssignments{},
			},
		},
		{
			name: "Empty roles - Deployment",
			data: models.OrganizationMembership{
				RoleAssignments: &models.RoleAssignments{
					Deployment:   []*models.DeploymentRoleAssignment{{}},
					Organization: []*models.OrganizationRoleAssignment{{}},
				},
			},
		},
		{
			name: "Empty roles - Elasticsearch",
			data: models.OrganizationMembership{
				RoleAssignments: &models.RoleAssignments{
					Project: &models.ProjectRoleAssignments{
						Elasticsearch: []*models.ProjectRoleAssignment{{}},
					},
				},
			},
		},
		{
			name: "Empty roles - Observability",
			data: models.OrganizationMembership{
				RoleAssignments: &models.RoleAssignments{
					Project: &models.ProjectRoleAssignments{
						Observability: []*models.ProjectRoleAssignment{{}},
					},
				},
			},
		},
		{
			name: "Empty roles - Security",
			data: models.OrganizationMembership{
				RoleAssignments: &models.RoleAssignments{
					Project: &models.ProjectRoleAssignments{
						Security: []*models.ProjectRoleAssignment{{}},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			diags := diag.Diagnostics{}
			apiToModel(context.Background(), test.data, false, &diags)
		})
	}
}
