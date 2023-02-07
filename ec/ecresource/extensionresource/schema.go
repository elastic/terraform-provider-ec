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

package extensionresource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api"

	"github.com/elastic/terraform-provider-ec/ec/internal"
	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifiers"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}
var _ resource.ResourceWithConfigValidators = &Resource{}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Required name of the ruleset",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description for extension",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					planmodifiers.StringDefaultValue(""),
				},
			},
			"extension_type": schema.StringAttribute{
				Description: "Extension type. bundle or plugin",
				Required:    true,
			},
			"version": schema.StringAttribute{
				Description: "Elasticsearch version",
				Required:    true,
			},
			"download_url": schema.StringAttribute{
				Description: "download url",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					planmodifiers.StringDefaultValue(""),
				},
			},
			// Uploading file via API
			"file_path": schema.StringAttribute{
				Description: "file path",
				Optional:    true,
			},
			"file_hash": schema.StringAttribute{
				Description: "file hash",
				Optional:    true,
			},
			"url": schema.StringAttribute{
				Description: "",
				Computed:    true,
			},
			"last_modified": schema.StringAttribute{
				Description: "",
				Computed:    true,
			},
			"size": schema.Int64Attribute{
				Description: "",
				Computed:    true,
			},
			// Computed attributes
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of this resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *Resource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.RequiredTogether(
			path.MatchRoot("file_path"),
			path.MatchRoot("file_hash"),
		),
	}
}

type Resource struct {
	client *api.API
}

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), request.ID)...)
}

func resourceReady(r *Resource, dg *diag.Diagnostics) bool {
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
	client, diags := internal.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	r.client = client
}

func (r *Resource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_deployment_extension"
}

type modelV0 struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	ExtensionType types.String `tfsdk:"extension_type"`
	Version       types.String `tfsdk:"version"`
	DownloadURL   types.String `tfsdk:"download_url"`
	FilePath      types.String `tfsdk:"file_path"`
	FileHash      types.String `tfsdk:"file_hash"`
	URL           types.String `tfsdk:"url"`
	LastModified  types.String `tfsdk:"last_modified"`
	Size          types.Int64  `tfsdk:"size"`
}
