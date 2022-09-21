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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifier"
)

type ResourceKind int

const (
	Apm ResourceKind = iota
	Elasticsearch
	EnterpriseSearch
	IntegrationsServer
	Kibana
)

func (d *DataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"name_prefix": {
				Type:        types.StringType,
				Description: "Prefix that one or several deployment names have in common.",
				Optional:    true,
			},
			"healthy": {
				Type:        types.StringType,
				Description: "Overall health status of the deployment.",
				Optional:    true,
			},
			"deployment_template_id": {
				Type:        types.StringType,
				Description: "ID of the deployment template used to create the deployment.",
				Optional:    true,
			},
			"tags": {
				Type:        types.MapType{ElemType: types.StringType},
				Description: "Key value map of arbitrary string tags for the deployment.\n",
				Optional:    true,
			},
			"size": {
				Type:                types.Int64Type,
				Description:         "The maximum number of deployments to return. Defaults to 100.",
				MarkdownDescription: "The maximum number of deployments to return. Defaults to `100`.",
				Optional:            true,
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
				Type:        types.Int64Type,
				Description: "The number of deployments actually returned.",
				Computed:    true,
			},
			"deployments": deploymentsListSchema(),
		},
		Blocks: map[string]tfsdk.Block{
			// Deployment resources
			"elasticsearch":       resourceFiltersSchema(Elasticsearch),
			"kibana":              resourceFiltersSchema(Kibana),
			"apm":                 resourceFiltersSchema(Apm),
			"integrations_server": resourceFiltersSchema(IntegrationsServer),
			"enterprise_search":   resourceFiltersSchema(EnterpriseSearch),
		},
	}, nil
}

func deploymentsListSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "List of deployments which match the specified query.",
		Computed:    true,
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"deployment_id": {
				Type:        types.StringType,
				Description: "The deployment unique ID.",
				Computed:    true,
			},
			"name": {
				Type:        types.StringType,
				Description: "The name of the deployment.",
				Computed:    true,
			},
			"alias": {
				Type:        types.StringType,
				Description: "Deployment alias.",
				Computed:    true,
			},
			"elasticsearch_resource_id": {
				Type:        types.StringType,
				Description: "The Elasticsearch resource unique ID.",
				Computed:    true,
			},
			"elasticsearch_ref_id": {
				Type:        types.StringType,
				Description: "The Elasticsearch resource reference.",
				Computed:    true,
			},
			"kibana_resource_id": {
				Type:        types.StringType,
				Description: "The Kibana resource unique ID.",
				Computed:    true,
			},
			"kibana_ref_id": {
				Type:        types.StringType,
				Description: "The Kibana resource reference.",
				Computed:    true,
			},
			"apm_resource_id": {
				Type:        types.StringType,
				Description: "The APM resource unique ID.",
				Computed:    true,
			},
			"apm_ref_id": {
				Type:        types.StringType,
				Description: "The APM resource reference.",
				Computed:    true,
			},
			"integrations_server_resource_id": {
				Type:        types.StringType,
				Description: "The Integrations Server resource unique ID.",
				Computed:    true,
			},
			"integrations_server_ref_id": {
				Type:        types.StringType,
				Description: "The Integrations Server resource reference.",
				Computed:    true,
			},
			"enterprise_search_resource_id": {
				Type:        types.StringType,
				Description: "The Enterprise Search resource unique ID.",
				Computed:    true,
			},
			"enterprise_search_ref_id": {
				Type:        types.StringType,
				Description: "The Enterprise Search resource reference.",
				Computed:    true,
			},
		}),
	}
}

func deploymentAttrTypes() map[string]attr.Type {
	return deploymentsListSchema().Attributes.Type().(types.ListType).ElemType.(types.ObjectType).AttrTypes

}

func (rk ResourceKind) Name() string {
	switch rk {
	case Apm:
		return "APM"
	case Elasticsearch:
		return "Elasticsearch"
	case EnterpriseSearch:
		return "Enterprise Search"
	case IntegrationsServer:
		return "Integrations Server"
	case Kibana:
		return "Kibana"
	default:
		return "unknown"
	}
}

func resourceFiltersSchema(resourceKind ResourceKind) tfsdk.Block {
	return tfsdk.Block{
		Description: fmt.Sprintf("Filter by %s resource kind status or configuration.", resourceKind.Name()),
		NestingMode: tfsdk.BlockNestingModeList,
		Attributes: map[string]tfsdk.Attribute{
			"healthy": {
				Type:     types.StringType,
				Optional: true,
			},
			"status": {
				Type:     types.StringType,
				Optional: true,
			},
			"version": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}
}

func resourceFiltersAttrTypes(resourceKind ResourceKind) map[string]attr.Type {
	return resourceFiltersSchema(resourceKind).Type().(types.ListType).ElemType.(types.ObjectType).AttrTypes

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
