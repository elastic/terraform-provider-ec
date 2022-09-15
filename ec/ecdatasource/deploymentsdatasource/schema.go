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

package deploymentsdatasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifier"
)

func (d *DataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"name_prefix": {
				Type:     types.StringType,
				Optional: true,
			},
			"healthy": {
				Type:     types.StringType,
				Optional: true,
			},
			"deployment_template_id": {
				Type:     types.StringType,
				Optional: true,
			},
			"tags": {
				Type:     types.MapType{ElemType: types.StringType},
				Optional: true,
			},
			"size": {
				Type:     types.Int64Type,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.Int64{Value: 100}),
				},
			},

			// Computed
			"id": {
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "Unique identifier of this data source.",
			},
			"return_count": {
				Type:     types.Int64Type,
				Computed: true,
			},
			"deployments": deploymentsListSchema(),

			// Deployment resources
			"elasticsearch":       resourceFiltersSchema(),
			"kibana":              resourceFiltersSchema(),
			"apm":                 resourceFiltersSchema(),
			"integrations_server": resourceFiltersSchema(),
			"enterprise_search":   resourceFiltersSchema(),
		},
	}, nil
}

func deploymentsListSchema() tfsdk.Attribute {
	// TODO should we use tfsdk.ListNestedAttributes here? - see https://github.com/hashicorp/terraform-provider-hashicups-pf/blob/8f222d805d39445673e442a674168349a45bc054/hashicups/data_source_coffee.go#L22
	return tfsdk.Attribute{
		Computed: true,
		Type: types.ListType{ElemType: types.ObjectType{
			AttrTypes: deploymentAttrTypes(),
		}},
	}
}

func deploymentAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"deployment_id":                   types.StringType,
		"name":                            types.StringType,
		"alias":                           types.StringType,
		"elasticsearch_resource_id":       types.StringType,
		"elasticsearch_ref_id":            types.StringType,
		"kibana_resource_id":              types.StringType,
		"kibana_ref_id":                   types.StringType,
		"apm_resource_id":                 types.StringType,
		"apm_ref_id":                      types.StringType,
		"integrations_server_resource_id": types.StringType,
		"integrations_server_ref_id":      types.StringType,
		"enterprise_search_resource_id":   types.StringType,
		"enterprise_search_ref_id":        types.StringType,
	}
}

func resourceFiltersSchema() tfsdk.Attribute {
	// TODO should we use tfsdk.ListNestedAttributes here? - see https://github.com/hashicorp/terraform-provider-hashicups-pf/blob/8f222d805d39445673e442a674168349a45bc054/hashicups/data_source_coffee.go#L22
	return tfsdk.Attribute{
		Optional: true,
		Type: types.ListType{ElemType: types.ObjectType{
			AttrTypes: resourceFiltersAttrTypes(),
		}},
	}
}

func resourceFiltersAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"healthy": types.StringType,
		"status":  types.StringType,
		"version": types.StringType,
	}
}

type modelV0 struct {
	ID                   types.String `tfsdk:"id"`
	NamePrefix           types.String `tfsdk:"name_prefix"`
	Healthy              types.String `tfsdk:"healthy"`
	DeploymentTemplateID types.String `tfsdk:"deployment_template_id"`
	Tags                 types.Map    `tfsdk:"tags"`
	Size                 types.Int64  `tfsdk:"size"`
	ReturnCount          types.Int64  `tfsdk:"return_count"`
	Deployments          types.List   `tfsdk:"deployments"`         //< deploymentModelV0
	Elasticsearch        types.List   `tfsdk:"elasticsearch"`       //< resourceFiltersModelV0
	Kibana               types.List   `tfsdk:"kibana"`              //< resourceFiltersModelV0
	Apm                  types.List   `tfsdk:"apm"`                 //< resourceFiltersModelV0
	IntegrationsServer   types.List   `tfsdk:"integrations_server"` //< resourceFiltersModelV0
	EnterpriseSearch     types.List   `tfsdk:"enterprise_search"`   //< resourceFiltersModelV0
}

type deploymentModelV0 struct {
	DeploymentID                 types.String `tfsdk:"deployment_id"`
	Name                         types.String `tfsdk:"name"`
	Alias                        types.String `tfsdk:"alias"`
	ElasticSearchResourceID      types.String `tfsdk:"elasticsearch_resource_id"`
	ElasticSearchRefID           types.String `tfsdk:"elasticsearch_ref_id"`
	KibanaResourceID             types.String `tfsdk:"kibana_resource_id"`
	KibanaRefID                  types.String `tfsdk:"kibana_ref_id"`
	ApmResourceID                types.String `tfsdk:"apm_resource_id"`
	ApmRefID                     types.String `tfsdk:"apm_ref_id"`
	IntegrationsServerResourceID types.String `tfsdk:"integrations_server_resource_id"`
	IntegrationsServerRefID      types.String `tfsdk:"integrations_server_ref_id"`
	EnterpriseSearchResourceID   types.String `tfsdk:"enterprise_search_resource_id"`
	EnterpriseSearchRefID        types.String `tfsdk:"enterprise_search_ref_id"`
}

type resourceFiltersModelV0 struct {
	Healthy types.String `tfsdk:"healthy"`
	Status  types.String `tfsdk:"status"`
	Version types.String `tfsdk:"version"`
}
