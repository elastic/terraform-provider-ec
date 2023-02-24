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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *DataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"version_regex": {
				Type:     types.StringType,
				Required: true,
			},
			"region": {
				Type:     types.StringType,
				Required: true,
			},
			"lock": {
				Type:     types.BoolType,
				Optional: true,
			},

			// Computed attributes
			"id": {
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "Unique identifier of this data source.",
			},
			"version": {
				Type:     types.StringType,
				Computed: true,
			},
			"accessible": {
				Type:     types.BoolType,
				Computed: true,
			},
			"min_upgradable_from": {
				Type:     types.StringType,
				Computed: true,
			},
			"upgradable_to": {
				Type:     types.ListType{ElemType: types.StringType},
				Computed: true,
			},
			"allowlisted": {
				Type:     types.BoolType,
				Computed: true,
			},
			"apm":               resourceKindConfigSchema(util.ApmResourceKind),
			"enterprise_search": resourceKindConfigSchema(util.EnterpriseSearchResourceKind),
			"elasticsearch":     elasticsearchConfigSchema(),
			"kibana":            resourceKindConfigSchema(util.KibanaResourceKind),
		},
	}, nil
}

func elasticsearchConfigSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Information for Elasticsearch workloads on this stack version.",
		Computed:    true,
		Validators:  []tfsdk.AttributeValidator{listvalidator.SizeAtMost(1)},
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"denylist": {
				Type:        types.ListType{ElemType: types.StringType},
				Description: "List of configuration options that cannot be overridden by user settings.",
				Computed:    true,
			},
			"capacity_constraints_max": {
				Type:        types.Int64Type,
				Description: "Maximum size of the instances.",
				Computed:    true,
			},
			"capacity_constraints_min": {
				Type:        types.Int64Type,
				Description: "Minimum size of the instances.",
				Computed:    true,
			},
			"compatible_node_types": {
				Type:        types.ListType{ElemType: types.StringType},
				Description: "List of node types compatible with this one.",
				Computed:    true,
			},
			"docker_image": {
				Type:        types.StringType,
				Description: "Docker image to use for the Elasticsearch cluster instances.",
				Computed:    true,
			},
			"plugins": {
				Type:        types.ListType{ElemType: types.StringType},
				Description: "List of available plugins to be specified by users in Elasticsearch cluster instances.",
				Computed:    true,
			},
			"default_plugins": {
				Type:        types.ListType{ElemType: types.StringType},
				Description: "List of default plugins.",
				Computed:    true,
			},
			// node_types not added. It is highly unlikely they will be used
			// for anything, and if they're needed in the future, then we can
			// invest on adding them.
		}),
	}
}

func elasticsearchConfigAttrTypes() map[string]attr.Type {
	return elasticsearchConfigSchema().Attributes.Type().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}

func resourceKindConfigSchema(resourceKind util.ResourceKind) tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: fmt.Sprintf("Information for %s workloads on this stack version.", resourceKind.Name()),
		Computed:    true,
		Validators:  []tfsdk.AttributeValidator{listvalidator.SizeAtMost(1)},
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"denylist": {
				Type:        types.ListType{ElemType: types.StringType},
				Description: "List of configuration options that cannot be overridden by user settings.",
				Computed:    true,
			},
			"capacity_constraints_max": {
				Type:        types.Int64Type,
				Description: "Maximum size of the instances.",
				Computed:    true,
			},
			"capacity_constraints_min": {
				Type:        types.Int64Type,
				Description: "Minimum size of the instances.",
				Computed:    true,
			},
			"compatible_node_types": {
				Type:        types.ListType{ElemType: types.StringType},
				Description: "List of node types compatible with this one.",
				Computed:    true,
			},
			"docker_image": {
				Type:        types.StringType,
				Description: fmt.Sprintf("Docker image to use for the %s instance.", resourceKind.Name()),
				Computed:    true,
			},
			// node_types not added. It is highly unlikely they will be used
			// for anything, and if they're needed in the future, then we can
			// invest on adding them.
		}),
	}
}

func resourceKindConfigAttrTypes(resourceKind util.ResourceKind) map[string]attr.Type {
	return resourceKindConfigSchema(resourceKind).Attributes.Type().(types.ListType).ElemType.(types.ObjectType).AttrTypes
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
