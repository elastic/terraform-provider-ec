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

			"apm":               kindResourceSchema(),
			"enterprise_search": kindResourceSchema(),
			"elasticsearch":     kindResourceSchema(),
			"kibana":            kindResourceSchema(),
		},
	}, nil
}

func kindResourceSchema() tfsdk.Attribute {
	// TODO should we use tfsdk.ListNestedAttributes here? - see https://github.com/hashicorp/terraform-provider-hashicups-pf/blob/8f222d805d39445673e442a674168349a45bc054/hashicups/data_source_coffee.go#L22
	return tfsdk.Attribute{
		Computed: true,
		Type: types.ListType{ElemType: types.ObjectType{
			AttrTypes: resourceKindConfigAttrTypes(),
		}},
	}
}

func resourceKindConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"denylist":                 types.ListType{ElemType: types.StringType},
		"capacity_constraints_max": types.Int64Type,
		"capacity_constraints_min": types.Int64Type,
		"compatible_node_types":    types.ListType{ElemType: types.StringType},
		"docker_image":             types.StringType,
		"plugins":                  types.ListType{ElemType: types.StringType},
		"default_plugins":          types.ListType{ElemType: types.StringType},

		// node_types not added. It is highly unlikely they will be used
		// for anything, and if they're needed in the future, then we can
		// invest on adding them.
	}
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
	Elasticsearch     types.List   `tfsdk:"elasticsearch"`     //< resourceKindConfigModelV0
	Kibana            types.List   `tfsdk:"kibana"`            //< resourceKindConfigModelV0
}

type resourceKindConfigModelV0 struct {
	DenyList               types.List   `tfsdk:"denylist"`
	CapacityConstraintsMax types.Int64  `tfsdk:"capacity_constraints_max"`
	CapacityConstraintsMin types.Int64  `tfsdk:"capacity_constraints_min"`
	CompatibleNodeTypes    types.List   `tfsdk:"compatible_node_types"`
	DockerImage            types.String `tfsdk:"docker_image"`
	Plugins                types.List   `tfsdk:"plugins"`
	DefaultPlugins         types.List   `tfsdk:"default_plugins"`
}
