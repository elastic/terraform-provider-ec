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

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Provides an Elastic Cloud Deployment Elasticsearch keystore resource, which allows you to create and update Elasticsearch keystore settings.

  Elasticsearch keystore settings can be created and updated through this resource, **each resource represents a single Elasticsearch Keystore setting**. After adding a key and its secret value to the keystore, you can use the key in place of the secret value when you configure sensitive settings.

  ~> **Note on Elastic keystore settings** This resource offers weaker consistency guarantees and will not detect and update keystore setting values that have been modified outside of the scope of Terraform, usually referred to as _drift_. For example, consider the following scenario:
    1. A keystore setting is created using this resource.
    2. The keystore setting's value is modified to a different value using the Elasticsearch Service API.
    3. Running <code>terraform apply</code> fails to detect the changes and does not update the keystore setting to the value defined in the terraform configuration.
    To force the keystore setting to the value it is configured to hold, you may want to taint the resource and force its recreation.

  Before you create Elasticsearch keystore settings, check the [official Elasticsearch keystore documentation](https://www.elastic.co/guide/en/elasticsearch/reference/master/elasticsearch-keystore.html) and the [Elastic Cloud specific documentation](https://www.elastic.co/guide/en/cloud/current/ec-configuring-keystore.html).`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deployment_id": schema.StringAttribute{
				Description: `Deployment ID of the Deployment that holds the Elasticsearch cluster where the keystore setting will be written to.`,
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"setting_name": schema.StringAttribute{
				Description: "Name for the keystore setting, if the setting already exists in the Elasticsearch cluster, it will be overridden.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Description: "Value of this setting. This can either be a string or a JSON object that is stored as a JSON string in the keystore.",
				Sensitive:   true,
				Required:    true,
			},
			"as_file": schema.BoolAttribute{
				Description: "Indicates the the remote keystore setting should be stored as a file. The default is false, which stores the keystore setting as string when value is a plain string.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					planmodifiers.BoolDefaultValue(false),
				},
			},
		},
	}
}

type Resource struct {
	client *api.API
}

func (r *Resource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	clients, diags := internal.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	r.client = clients.Stateful
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
