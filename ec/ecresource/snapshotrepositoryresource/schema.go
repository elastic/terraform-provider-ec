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

package snapshotrepositoryresource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
var _ resource.ResourceWithImportState = &Resource{}
var _ resource.ResourceWithConfigValidators = &Resource{}

func (r *Resource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `Manages Elastic Cloud Enterprise snapshot repositories.

  ~> **This resource can only be used with Elastic Cloud Enterprise** For Elastic Cloud SaaS please use the [elasticstack_elasticsearch_snapshot_repository](https://registry.terraform.io/providers/elastic/elasticstack/latest/docs/resources/elasticsearch_snapshot_repository) resource from the [Elastic Stack terraform provider](https://registry.terraform.io/providers/elastic/elasticstack/latest).`,
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:                types.StringType,
				MarkdownDescription: "Unique identifier of this resource.",
				Computed:            true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"name": {
				Type:        types.StringType,
				Description: "The name of the snapshot repository configuration.",
				Required:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
			"generic": genericSchema(),
			"s3":      s3Schema(),
		},
	}, nil
}

func s3Schema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "S3 repository settings.",
		Optional:    true,
		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			"region": {
				Type:        types.StringType,
				Description: "Allows specifying the signing region to use. Specifying this setting manually should not be necessary for most use cases. Generally, the SDK will correctly guess the signing region to use. It should be considered an expert level setting to support S3-compatible APIs that require v4 signatures and use a region other than the default us-east-1. Defaults to empty string which means that the SDK will try to automatically determine the correct signing region.",
				Optional:    true,
			},
			"bucket": {
				Type:        types.StringType,
				Description: "Name of the S3 bucket to use for snapshots.",
				Required:    true,
			},
			"access_key": {
				Type:        types.StringType,
				Description: "An S3 access key. If set, the secret_key setting must also be specified. If unset, the client will use the instance or container role instead.",
				Optional:    true,
			},
			"secret_key": {
				Type:        types.StringType,
				Description: "An S3 secret key. If set, the access_key setting must also be specified.",
				Optional:    true,
				Sensitive:   true,
			},
			"server_side_encryption": {
				Type:        types.BoolType,
				Description: "When set to true files are encrypted on server side using AES256 algorithm. Defaults to false.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.Bool{Value: false}),
				},
			},
			"endpoint": {
				Type:        types.StringType,
				Description: "The S3 service endpoint to connect to. This defaults to s3.amazonaws.com but the AWS documentation lists alternative S3 endpoints. If you are using an S3-compatible service then you should set this to the service’s endpoint.",
				Optional:    true,
			},
			"path_style_access": {
				Type:        types.BoolType,
				Description: "Whether to force the use of the path style access pattern. If true, the path style access pattern will be used. If false, the access pattern will be automatically determined by the AWS Java SDK (See AWS documentation for details). Defaults to false.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifier.DefaultValue(types.Bool{Value: false}),
				},
			},
		}),
	}
}

func genericSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Generic repository settings.",
		Optional:    true,
		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			"type": {
				Type:        types.StringType,
				Description: "Repository type",
				Required:    true,
			},
			"settings": {
				Type:        types.StringType,
				Description: "An arbitrary JSON object containing the repository settings.",
				Required:    true,
			},
		}),
	}
}

type Resource struct {
	client *api.API
}

func (r *Resource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("generic"),
			path.MatchRoot("s3"),
		),
	}
}

func (r *Resource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client, diags := internal.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	r.client = client
}

func (r *Resource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_snapshot_repository"
}

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
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

type modelV0 struct {
	ID      types.String         `tfsdk:"id"`
	Name    types.String         `tfsdk:"name"`
	S3      *s3RepositoryV0      `tfsdk:"s3"`
	Generic *genericRepositoryV0 `tfsdk:"generic"`
}

type s3RepositoryV0 struct {
	Region               types.String `tfsdk:"region"`
	Bucket               types.String `tfsdk:"bucket"`
	AccessKey            types.String `tfsdk:"access_key"`
	SecretKey            types.String `tfsdk:"secret_key"`
	ServerSideEncryption types.Bool   `tfsdk:"server_side_encryption"`
	Endpoint             types.String `tfsdk:"endpoint"`
	PathStyleAccess      types.Bool   `tfsdk:"path_style_access"`
}
type genericRepositoryV0 struct {
	Type     types.String `tfsdk:"type"`
	Settings types.String `tfsdk:"settings"`
}
