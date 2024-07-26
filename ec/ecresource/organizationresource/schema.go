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
	"fmt"
	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

type Organization struct {
	ID      types.String `tfsdk:"id"`
	Members types.Map    `tfsdk:"members"` //< OrganizationMember
}

type OrganizationMember struct {
	Email                     types.String `tfsdk:"email"`
	InvitationPending         types.Bool   `tfsdk:"invitation_pending"`
	UserID                    types.String `tfsdk:"user_id"`
	OrganizationRole          types.String `tfsdk:"organization_role"`
	DeploymentRoles           types.Set    `tfsdk:"deployment_roles"`            //< DeploymentRoleAssignment
	ProjectElasticsearchRoles types.Set    `tfsdk:"project_elasticsearch_roles"` //< ProjectRoleAssignment
	ProjectObservabilityRoles types.Set    `tfsdk:"project_observability_roles"` //< ProjectRoleAssignment
	ProjectSecurityRoles      types.Set    `tfsdk:"project_security_roles"`      //< ProjectRoleAssignment
}

type DeploymentRoleAssignment struct {
	Role              types.String `tfsdk:"role"`
	ForAllDeployments types.Bool   `tfsdk:"for_all_deployments"`
	DeploymentIDs     types.Set    `tfsdk:"deployment_ids"`
	ApplicationRoles  types.Set    `tfsdk:"application_roles"`
}

type ProjectRoleAssignment struct {
	Role             types.String `tfsdk:"role"`
	ForAllProjects   types.Bool   `tfsdk:"for_all_projects"`
	ProjectIDs       types.Set    `tfsdk:"project_ids"`
	ApplicationRoles types.Set    `tfsdk:"application_roles"`
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages an Elastic Cloud organization membership.

  ~> **This resource can only be used with Elastic Cloud SaaS**`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Organization ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"members": organizationMembersSchema(),
		},
	}
}

func organizationMembersSchema() schema.MapNestedAttribute {
	return schema.MapNestedAttribute{
		MarkdownDescription: "Manages the members of an Elastic Cloud organization. The key of each entry should be the email of the member.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"email": schema.StringAttribute{
					MarkdownDescription: "Email address of the user.",
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"invitation_pending": schema.BoolAttribute{
					MarkdownDescription: "Set to true while the user has not yet accepted their invitation to the organization.",
					Computed:            true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
					},
				},
				"user_id": schema.StringAttribute{
					MarkdownDescription: "User ID.",
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"organization_role": schema.StringAttribute{
					MarkdownDescription: "The optional organization role for the member. Can be one of `organization-admin`, `billing-admin`. For more info see: [Organization roles](https://www.elastic.co/guide/en/cloud/current/ec-user-privileges.html#ec_organization_level_roles)",
					Optional:            true,
				},
				"deployment_roles":            deploymentRoleAssignmentsSchema(),
				"project_elasticsearch_roles": projectElasticsearchRolesSchema(),
				"project_observability_roles": projectObservabilityRolesSchema(),
				"project_security_roles":      projectSecurityRolesSchema(),
			},
		},
	}
}

func deploymentRoleAssignmentsSchema() schema.SetNestedAttribute {
	return schema.SetNestedAttribute{
		MarkdownDescription: "Grant access to one or more deployments. For more info see: [Deployment instance roles](https://www.elastic.co/guide/en/cloud/current/ec-user-privileges.html#ec_instance_access_roles).",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"role": schema.StringAttribute{
					MarkdownDescription: "Assigned role. Must be on of `viewer`, `editor` or `admin`.",
					Required:            true,
				},
				"for_all_deployments": schema.BoolAttribute{
					MarkdownDescription: "Role applies to all deployments in the organization.",
					Optional:            true,
					PlanModifiers: []planmodifier.Bool{
						planmodifiers.BoolDefaultValue(false), // consider unknown as false
					},
				},
				"deployment_ids": schema.SetAttribute{
					MarkdownDescription: "Role applies to deployments listed here.",
					Optional:            true,
					ElementType:         types.StringType,
				},
				"application_roles": schema.SetAttribute{
					MarkdownDescription: "If provided, the user assigned this role assignment will be granted this application role when signing in to the deployment(s) specified in the role assignment.",
					Optional:            true,
					ElementType:         types.StringType,
				},
			},
		},
	}
}

func projectElasticsearchRolesSchema() schema.SetNestedAttribute {
	elementSchema := projectRoleAssignmentSchema([]string{
		"admin",
		"developer",
		"viewer",
	})
	return schema.SetNestedAttribute{
		MarkdownDescription: "Roles assigned for elasticsearch projects. For more info see: [Serverless elasticsearch roles](https://www.elastic.co/docs/current/serverless/general/assign-user-roles#es) ",
		Optional:            true,
		Computed:            true,
		NestedObject:        elementSchema,
		PlanModifiers: []planmodifier.Set{
			planmodifiers.SetDefaultValue(elementSchema.Type(), nil),
		},
	}
}

func projectObservabilityRolesSchema() schema.SetNestedAttribute {
	elementSchema := projectRoleAssignmentSchema([]string{
		"admin",
		"editor",
		"viewer",
	})
	return schema.SetNestedAttribute{
		MarkdownDescription: "Roles assigned for observability projects. For more info see: [Serverless observability roles](https://www.elastic.co/docs/current/serverless/general/assign-user-roles#observability)",
		Optional:            true,
		Computed:            true,
		NestedObject:        elementSchema,
		PlanModifiers: []planmodifier.Set{
			planmodifiers.SetDefaultValue(elementSchema.Type(), nil),
		},
	}
}

func projectSecurityRolesSchema() schema.SetNestedAttribute {
	elementSchema := projectRoleAssignmentSchema([]string{
		"admin",
		"editor",
		"viewer",
		"t1-analyst",
		"t2-analyst",
		"t3-analyst",
		"threat-intel-analyst",
		"rule-author",
		"soc-manager",
		"endpoint-operations-analyst",
		"platform-engineer",
		"detections-admin",
		"endpoint-policy-manager",
	})
	return schema.SetNestedAttribute{
		MarkdownDescription: "Roles assigned for security projects. For more info see: [Serverless security roles](https://www.elastic.co/docs/current/serverless/general/assign-user-roles#security)",
		Optional:            true,
		Computed:            true,
		NestedObject:        elementSchema,
		PlanModifiers: []planmodifier.Set{
			planmodifiers.SetDefaultValue(elementSchema.Type(), []attr.Value{}),
		},
	}
}

func projectRoleAssignmentSchema(roles []string) schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"role": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Assigned role. (Allowed values: %s)", "`" + strings.Join(roles, "`, `") + "`"),
				Required:            true,
			},
			"for_all_projects": schema.BoolAttribute{
				MarkdownDescription: "Role applies to all deployments in the organization.",
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					planmodifiers.BoolDefaultValue(false), // consider unknown as false
				},
			},
			"project_ids": schema.SetAttribute{
				MarkdownDescription: "Role applies to deployments listed here.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"application_roles": schema.SetAttribute{
				MarkdownDescription: "If provided, the user assigned this role assignment will be granted this application role when signing in to the project(s) specified in the role assignment.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func setElementObjectType(schema schema.SetNestedAttribute) types.ObjectType {
	return types.ObjectType{AttrTypes: schema.GetType().(types.SetType).ElemType.(types.ObjectType).AttrTypes}
}

func mapElementObjectType(schema schema.MapNestedAttribute) types.ObjectType {
	return types.ObjectType{AttrTypes: schema.GetType().(types.MapType).ElemType.(types.ObjectType).AttrTypes}
}
