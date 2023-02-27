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
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifier"

	"github.com/elastic/terraform-provider-ec/ec/internal"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}
var _ resource.ResourceWithConfigValidators = &Resource{}

func (r *Resource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"name": {
				Type:        types.StringType,
				Description: "Required name of the ruleset",
				Required:    true,
			},
			"description": {
				Type:        types.StringType,
				Description: "Description for extension",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.String{Value: ""}),
				}},
			"extension_type": {
				Type:        types.StringType,
				Description: "Extension type. bundle or plugin",
				Required:    true,
			},
			"version": {
				Type:        types.StringType,
				Description: "Elasticsearch version",
				Required:    true,
			},
			"download_url": {
				Type:        types.StringType,
				Description: "download url",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.String{Value: ""}),
				},
			},

			// Uploading file via API
			"file_path": {
				Type:        types.StringType,
				Description: "file path",
				Optional:    true,
			},
			"file_hash": {
				Type:        types.StringType,
				Description: "file hash",
				Optional:    true,
			},
			"url": {
				Type:        types.StringType,
				Description: "",
				Computed:    true,
			},
			"last_modified": {
				Type:        types.StringType,
				Description: "",
				Computed:    true,
			},
			"size": {
				Type:        types.Int64Type,
				Description: "",
				Computed:    true,
			},
			// Computed attributes
			"id": {
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "Unique identifier of this resource.",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
		},
	}, nil
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
