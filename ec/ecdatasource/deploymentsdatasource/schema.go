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
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of IDs for the deployment and resource kinds, based on the specified query.",
		Attributes: map[string]schema.Attribute{
			"name_prefix": schema.StringAttribute{
				Description: "Prefix to filter the returned deployment list by.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Filter the result by the full deployment name.",
				Optional:    true,
			},
			"healthy": schema.StringAttribute{
				Description: "Filter the result set by their health status.",
				Optional:    true,
			},
			"deployment_template_id": schema.StringAttribute{
				Description: "Filter the result set by the ID of the deployment template the deployment is based off.",
				Optional:    true,
			},
			"tags": schema.MapAttribute{
				ElementType: types.StringType,
				Description: "Filter the result set by their assigned tags.",
				Optional:    true,
			},
			"size": schema.Int64Attribute{
				Description:         "The maximum number of deployments to return. Defaults to 100.",
				MarkdownDescription: "The maximum number of deployments to return. Defaults to `100`.",
				Optional:            true,
				// PlanModifiers: []tfsdk.AttributePlanModifier{
				// 	planmodifier.DefaultValue(types.Int64Value(100)),
				// },
			},

			// Computed
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of this data source.",
			},
			"return_count": schema.Int64Attribute{
				Description: "The number of deployments actually returned.",
				Computed:    true,
			},
			"deployments": deploymentsListSchema(),
		},
		Blocks: map[string]schema.Block{
			// Deployment resources
			"elasticsearch":       resourceFiltersSchema(util.ElasticsearchResourceKind),
			"kibana":              resourceFiltersSchema(util.KibanaResourceKind),
			"apm":                 resourceFiltersSchema(util.ApmResourceKind),
			"integrations_server": resourceFiltersSchema(util.IntegrationsServerResourceKind),
			"enterprise_search":   resourceFiltersSchema(util.EnterpriseSearchResourceKind),
		},
	}
}

func deploymentsListSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Description: "List of deployments which match the specified query.",
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"deployment_id": schema.StringAttribute{
					Description: "The deployment unique ID.",
					Computed:    true,
				},
				"name": schema.StringAttribute{
					Description: "The name of the deployment.",
					Computed:    true,
				},
				"alias": schema.StringAttribute{
					Description: "Deployment alias.",
					Computed:    true,
				},
				"elasticsearch_resource_id": schema.StringAttribute{
					Description: "The Elasticsearch resource unique ID.",
					Computed:    true,
				},
				"elasticsearch_ref_id": schema.StringAttribute{
					Description: "The Elasticsearch resource reference.",
					Computed:    true,
				},
				"kibana_resource_id": schema.StringAttribute{
					Description: "The Kibana resource unique ID.",
					Computed:    true,
				},
				"kibana_ref_id": schema.StringAttribute{
					Description: "The Kibana resource reference.",
					Computed:    true,
				},
				"apm_resource_id": schema.StringAttribute{
					Description: "The APM resource unique ID.",
					Computed:    true,
				},
				"apm_ref_id": schema.StringAttribute{
					Description: "The APM resource reference.",
					Computed:    true,
				},
				"integrations_server_resource_id": schema.StringAttribute{
					Description: "The Integrations Server resource unique ID.",
					Computed:    true,
				},
				"integrations_server_ref_id": schema.StringAttribute{
					Description: "The Integrations Server resource reference.",
					Computed:    true,
				},
				"enterprise_search_resource_id": schema.StringAttribute{
					Description: "The Enterprise Search resource unique ID.",
					Computed:    true,
				},
				"enterprise_search_ref_id": schema.StringAttribute{
					Description: "The Enterprise Search resource reference.",
					Computed:    true,
				},
			},
		},
	}
}

func deploymentAttrTypes() map[string]attr.Type {
	return deploymentsListSchema().GetType().(types.ListType).ElemType.(types.ObjectType).AttrTypes

}

func resourceFiltersSchema(resourceKind util.ResourceKind) schema.Block {
	return schema.ListNestedBlock{
		Description: fmt.Sprintf("Filter by %s resource kind status or configuration.", resourceKind.Name()),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"healthy": schema.StringAttribute{
					Optional:    true,
					Description: "Overall health status of the resource instances.",
				},
				"status": schema.StringAttribute{
					Optional:    true,
					Description: "Resource kind status. Can be one of `initializing`, `stopping`, `stopped`, `rebooting`, `restarting`.",
				},
				"version": schema.StringAttribute{
					Optional:    true,
					Description: "Elastic stack version.",
				},
			},
		},
	}
}

func resourceFiltersAttrTypes(resourceKind util.ResourceKind) map[string]attr.Type {
	return resourceFiltersSchema(resourceKind).Type().(types.ListType).ElemType.(types.ObjectType).AttrTypes

}

type modelV0 struct {
	ID                   types.String `tfsdk:"id"`
	NamePrefix           types.String `tfsdk:"name_prefix"`
	Name                 types.String `tfsdk:"name"`
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
	ElasticsearchResourceID      types.String `tfsdk:"elasticsearch_resource_id"`
	ElasticsearchRefID           types.String `tfsdk:"elasticsearch_ref_id"`
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
