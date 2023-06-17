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

package stackdatasource

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Use this data source to retrieve information about an existing Elastic Cloud stack.

  -> **Note on regions** Before you start, you might want to check the [full list](https://www.elastic.co/guide/en/cloud/current/ec-regions-templates-instances.html) of regions available in Elasticsearch Service (ESS).`,
		Attributes: map[string]schema.Attribute{
			"version_regex": schema.StringAttribute{
				Required:    true,
				Description: "Regex to filter the available stacks. Can be any valid regex expression, when multiple stacks are matched through a regex, the latest version is returned. `latest` is also accepted to obtain the latest available stack version.",
			},
			"region": schema.StringAttribute{
				Required:    true,
				Description: "Region where the stack pack is. For Elastic Cloud Enterprise (ECE) installations, use `ece-region`.",
			},
			"lock": schema.BoolAttribute{
				Optional:    true,
				Description: "Lock the `latest` `version_regex` obtained, so that the new stack release doesn't cascade the changes down to the deployments. It can be changed at any time.",
			},

			// Computed attributes
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of this data source.",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "The stack version",
			},
			"accessible": schema.BoolAttribute{
				Computed:    true,
				Description: "To have this version accessible/not accessible by the calling user. This is only relevant for Elasticsearch Service (ESS), not for ECE.",
			},
			"min_upgradable_from": schema.StringAttribute{
				Computed:    true,
				Description: "The minimum stack version which can be upgraded to this stack version.",
			},
			"upgradable_to": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "A list of stack versions which this stack version can be upgraded to.",
			},
			"allowlisted": schema.BoolAttribute{
				Computed:    true,
				Description: "To include/not include this version in the `allowlist`. This is only relevant for Elasticsearch Service (ESS), not for ECE.",
			},
			"apm":               resourceKindConfigSchema(util.ApmResourceKind),
			"enterprise_search": resourceKindConfigSchema(util.EnterpriseSearchResourceKind),
			"elasticsearch":     elasticsearchConfigSchema(),
			"kibana":            resourceKindConfigSchema(util.KibanaResourceKind),
		},
	}
}

func elasticsearchConfigSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Description: "Information for Elasticsearch workloads on this stack version.",
		Computed:    true,
		Validators:  []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"denylist": schema.ListAttribute{
					ElementType: types.StringType,
					Description: "List of configuration options that cannot be overridden by user settings.",
					Computed:    true,
				},
				"capacity_constraints_max": schema.Int64Attribute{
					Description: "Maximum size of the instances.",
					Computed:    true,
				},
				"capacity_constraints_min": schema.Int64Attribute{
					Description: "Minimum size of the instances.",
					Computed:    true,
				},
				"compatible_node_types": schema.ListAttribute{
					ElementType: types.StringType,
					Description: "List of node types compatible with this one.",
					Computed:    true,
				},
				"docker_image": schema.StringAttribute{
					Description: "Docker image to use for the Elasticsearch cluster instances.",
					Computed:    true,
				},
				"plugins": schema.ListAttribute{
					ElementType: types.StringType,
					Description: "List of available plugins to be specified by users in Elasticsearch cluster instances.",
					Computed:    true,
				},
				"default_plugins": schema.ListAttribute{
					ElementType: types.StringType,
					Description: "List of default plugins.",
					Computed:    true,
				},
				// node_types not added. It is highly unlikely they will be used
				// for anything, and if they're needed in the future, then we can
				// invest on adding them.
			},
		},
	}
}

func elasticsearchConfigAttrTypes() map[string]attr.Type {
	return elasticsearchConfigSchema().GetType().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}

func resourceKindConfigSchema(resourceKind util.ResourceKind) schema.Attribute {
	return schema.ListNestedAttribute{
		Description: fmt.Sprintf("Information for %s workloads on this stack version.", resourceKind.Name()),
		Computed:    true,
		Validators:  []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"denylist": schema.ListAttribute{
					ElementType: types.StringType,
					Description: "List of configuration options that cannot be overridden by user settings.",
					Computed:    true,
				},
				"capacity_constraints_max": schema.Int64Attribute{
					Description: "Maximum size of the instances.",
					Computed:    true,
				},
				"capacity_constraints_min": schema.Int64Attribute{
					Description: "Minimum size of the instances.",
					Computed:    true,
				},
				"compatible_node_types": schema.ListAttribute{
					ElementType: types.StringType,
					Description: "List of node types compatible with this one.",
					Computed:    true,
				},
				"docker_image": schema.StringAttribute{
					Description: fmt.Sprintf("Docker image to use for the %s instance.", resourceKind.Name()),
					Computed:    true,
				},
				// node_types not added. It is highly unlikely they will be used
				// for anything, and if they're needed in the future, then we can
				// invest on adding them.
			},
		},
	}
}

func resourceKindConfigAttrTypes(resourceKind util.ResourceKind) map[string]attr.Type {
	return resourceKindConfigSchema(resourceKind).GetType().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}

type modelV0 struct {
	ID                types.String `tfsdk:"id"`
	VersionRegex      types.String `tfsdk:"version_regex"`
	Region            types.String `tfsdk:"region"`
	Lock              types.Bool   `tfsdk:"lock"`
	Version           types.String `tfsdk:"version"`
	Accessible        types.Bool   `tfsdk:"accessible"`
	MinUpgradableFrom types.String `tfsdk:"min_upgradable_from"`
	UpgradableTo      types.List   `tfsdk:"upgradable_to"`
	AllowListed       types.Bool   `tfsdk:"allowlisted"`
	Apm               types.List   `tfsdk:"apm"`               //< resourceKindConfigModelV0
	EnterpriseSearch  types.List   `tfsdk:"enterprise_search"` //< resourceKindConfigModelV0
	Elasticsearch     types.List   `tfsdk:"elasticsearch"`     //< elasticsearchConfigModelV0
	Kibana            types.List   `tfsdk:"kibana"`            //< resourceKindConfigModelV0
}

type elasticsearchConfigModelV0 struct {
	DenyList               types.List   `tfsdk:"denylist"`
	CapacityConstraintsMax types.Int64  `tfsdk:"capacity_constraints_max"`
	CapacityConstraintsMin types.Int64  `tfsdk:"capacity_constraints_min"`
	CompatibleNodeTypes    types.List   `tfsdk:"compatible_node_types"`
	DockerImage            types.String `tfsdk:"docker_image"`
	Plugins                types.List   `tfsdk:"plugins"`
	DefaultPlugins         types.List   `tfsdk:"default_plugins"`
}

type resourceKindConfigModelV0 struct {
	DenyList               types.List   `tfsdk:"denylist"`
	CapacityConstraintsMax types.Int64  `tfsdk:"capacity_constraints_max"`
	CapacityConstraintsMin types.Int64  `tfsdk:"capacity_constraints_min"`
	CompatibleNodeTypes    types.List   `tfsdk:"compatible_node_types"`
	DockerImage            types.String `tfsdk:"docker_image"`
}
