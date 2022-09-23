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

package elasticsearchkeystoreresource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/terraform-provider-ec/ec/internal"
	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifier"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithGetSchema = &Resource{}

var _ resource.ResourceWithMetadata = &Resource{}

func (r *Resource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:                types.StringType,
				MarkdownDescription: "Unique identifier of this resource.",
				Computed:            true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"deployment_id": {
				Type:        types.StringType,
				Description: `Required deployment ID of the Deployment that holds the Elasticsearch cluster where the keystore setting will be written to.`,
				Required:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
			"setting_name": {
				Type:        types.StringType,
				Description: "Required name for the keystore setting, if the setting already exists in the Elasticsearch cluster, it will be overridden.",
				Required:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
			"value": {
				Type:        types.StringType,
				Description: "Required value of this setting. This can either be a string or a JSON object that is stored as a JSON string in the keystore.",
				Sensitive:   true,
				Required:    true,
			},
			"as_file": {
				Type:        types.BoolType,
				Description: "Optionally stores the remote keystore setting as a file. The default is false, which stores the keystore setting as string when value is a plain string.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.Bool{Value: false}),
				},
			},
		},
	}, nil
}

type Resource struct {
	client *api.API
}

func (r *Resource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client, diags := internal.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	r.client = client
}

func (r *Resource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_deployment_elasticsearch_keystore"
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

type modelV0 struct {
	ID           types.String `tfsdk:"id"`
	DeploymentID types.String `tfsdk:"deployment_id"`
	SettingName  types.String `tfsdk:"setting_name"`
	Value        types.String `tfsdk:"value"`
	AsFile       types.Bool   `tfsdk:"as_file"`
}
