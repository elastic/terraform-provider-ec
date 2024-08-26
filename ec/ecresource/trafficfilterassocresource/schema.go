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

package trafficfilterassocresource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api"

	"github.com/elastic/terraform-provider-ec/ec/internal"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

const entityTypeDeployment = "deployment"

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Provides an Elastic Cloud traffic filter association resource, which allows traffic filter rules to be associated to an Elastic Cloud deployment outside of the control of Terraform. Associations can be created and deleted.`,
		Attributes: map[string]schema.Attribute{
			"deployment_id": schema.StringAttribute{
				Description: `Required deployment ID where the traffic filter will be associated`,
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"traffic_filter_id": schema.StringAttribute{
				Description: "Required traffic filter ruleset ID to tie to a deployment",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// Computed attributes
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of this resource.",
			},
		},
	}
}

type Resource struct {
	client *api.API
}

func resourceReady(r Resource, dg *diag.Diagnostics) bool {
	if r.client == nil {
		dg.AddError(
			"Unconfigured API Client",
			"Expected configured API client. Please report this issue to the provider developers.",
		)

		return false
	}
	return true
}

func (r *Resource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	clients, diags := internal.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	r.client = clients.Stateful
}

func (r *Resource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_deployment_traffic_filter_association"
}

type modelV0 struct {
	ID              types.String `tfsdk:"id"`
	DeploymentID    types.String `tfsdk:"deployment_id"`
	TrafficFilterID types.String `tfsdk:"traffic_filter_id"`
}
