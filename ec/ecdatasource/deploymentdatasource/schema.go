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

package deploymentdatasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"alias": schema.StringAttribute{
				Description: "Deployment alias.",
				Computed:    true,
			},
			"healthy": schema.BoolAttribute{
				Description: "Overall health status of the deployment.",
				Computed:    true,
			},
			"id": schema.StringAttribute{
				Description: "The unique ID of the deployment.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the deployment.",
				Computed:    true,
			},
			"region": schema.StringAttribute{
				Description: "Region where the deployment is hosted.",
				Computed:    true,
			},
			"deployment_template_id": schema.StringAttribute{
				Description: "ID of the deployment template this deployment is based off.",
				Computed:    true,
			},
			"traffic_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "Traffic filter block, which contains a list of traffic filter rule identifiers.",
				Computed:    true,
			},
			"tags": schema.MapAttribute{
				ElementType: types.StringType,
				Description: "Key value map of arbitrary string tags.",
				Computed:    true,
			},
			"observability":       observabilitySettingsSchema(),
			"elasticsearch":       elasticsearchResourceInfoSchema(),
			"kibana":              kibanaResourceInfoSchema(),
			"apm":                 apmResourceInfoSchema(),
			"integrations_server": integrationsServerResourceInfoSchema(),
			"enterprise_search":   enterpriseSearchResourceInfoSchema(),
		},
	}
}

type modelV0 struct {
	Alias                types.String `tfsdk:"alias"`
	Healthy              types.Bool   `tfsdk:"healthy"`
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Region               types.String `tfsdk:"region"`
	DeploymentTemplateID types.String `tfsdk:"deployment_template_id"`
	TrafficFilter        types.List   `tfsdk:"traffic_filter"`      //< string
	Observability        types.List   `tfsdk:"observability"`       //< observabilitySettingsModel
	Tags                 types.Map    `tfsdk:"tags"`                //< string
	Elasticsearch        types.List   `tfsdk:"elasticsearch"`       //< elasticsearchResourceInfoModelV0
	Kibana               types.List   `tfsdk:"kibana"`              //< kibanaResourceInfoModelV0
	Apm                  types.List   `tfsdk:"apm"`                 //< apmResourceInfoModelV0
	IntegrationsServer   types.List   `tfsdk:"integrations_server"` //< integrationsServerResourceInfoModelV0
	EnterpriseSearch     types.List   `tfsdk:"enterprise_search"`   //< enterpriseSearchResourceInfoModelV0
}
