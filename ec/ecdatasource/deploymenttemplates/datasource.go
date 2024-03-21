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

package deploymenttemplates

import (
	"context"
	"fmt"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deptemplateapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/terraform-provider-ec/ec/internal"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

type DataSource struct {
	client *api.API
}

var _ datasource.DataSource = &DataSource{}
var _ datasource.DataSourceWithConfigure = &DataSource{}

func (d *DataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_deployment_templates"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of available deployment templates.",
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Description: "Region to select. For Elastic Cloud Enterprise (ECE) installations, use `ece-region`.",
				Required:    true,
			},
			"stack_version": schema.StringAttribute{
				Description: "Filters for deployment templates compatible with this stack version.",
				Optional:    true,
			},
			"show_hidden": schema.BoolAttribute{
				Description: "Enable to also show hidden deployment templates. (Set to false by default.)",
				Optional:    true,
			},
			"templates": deploymentTemplatesListSchema(),
		},
	}
}

func deploymentTemplatesListSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Description: "List of available deployment templates.",
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Description: "The id of the deployment template.",
					Computed:    true,
				},
				"name": schema.StringAttribute{
					Description: "The name of the deployment template.",
					Computed:    true,
				},
				"description": schema.StringAttribute{
					Description: "The description of the deployment template.",
					Computed:    true,
				},
				"min_stack_version": schema.StringAttribute{
					Description: "The minimum stack version that can used with this deployment template.",
					Computed:    true,
				},
				"hidden": schema.BoolAttribute{
					Description: "If the template is visible by default. (Outdated templates are hidden, but can still be used)",
					Computed:    true,
				},
			},
		},
	}
}

func (d *DataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	client, diags := internal.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	d.client = client
}

func (d DataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	// Prevent panic if the provider has not been configured.
	if d.client == nil {
		response.Diagnostics.AddError(
			"Unconfigured API Client",
			"Expected configured API client. Please report this issue to the provider developers.",
		)
		return
	}

	var data deploymentTemplatesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	res, err := deptemplateapi.List(deptemplateapi.ListParams{
		API:                        d.client,
		MetadataFilter:             "",
		Region:                     data.Region.ValueString(),
		StackVersion:               data.StackVersion.ValueString(),
		ShowHidden:                 false,
		HideInstanceConfigurations: true,
	})

	if err != nil {
		response.Diagnostics.AddError(
			"Failed retrieving deployment template list",
			fmt.Sprintf("Failed retrieving deployment template list: %s", err),
		)
		return
	}

	showHidden := data.ShowHidden.ValueBool()

	templates := mapResponseToModel(res, showHidden)
	from, diagnostics := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: deploymentTemplateAttrTypes()}, templates)
	response.Diagnostics.Append(diagnostics...)
	if response.Diagnostics.HasError() {
		return
	}

	data.Templates = from

	// Finally, set the state
	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func mapResponseToModel(response []*models.DeploymentTemplateInfoV2, showHidden bool) []deploymentTemplateModel {
	templates := make([]deploymentTemplateModel, 0, len(response))
	for _, template := range response {
		hidden := isHidden(template)
		if !showHidden && hidden {
			continue
		}
		templateModel := deploymentTemplateModel{
			ID:              types.StringValue(*template.ID),
			Name:            types.StringValue(*template.Name),
			Description:     types.StringValue(template.Description),
			MinStackVersion: types.StringValue(template.MinVersion),
			Hidden:          types.BoolValue(hidden),
		}
		templates = append(templates, templateModel)
	}
	return templates
}

func isHidden(template *models.DeploymentTemplateInfoV2) bool {
	for _, metadatum := range template.Metadata {
		if *metadatum.Key == "hidden" && *metadatum.Value == "true" {
			return true
		}
	}
	return false
}

func deploymentTemplateAttrTypes() map[string]attr.Type {
	return deploymentTemplatesListSchema().GetType().(types.ListType).ElemType.(types.ObjectType).AttrTypes

}

type deploymentTemplatesDataSourceModel struct {
	Region       types.String `tfsdk:"region"`
	StackVersion types.String `tfsdk:"stack_version"`
	ShowHidden   types.Bool   `tfsdk:"show_hidden"`
	Templates    types.List   `tfsdk:"templates"` //< deploymentTemplateModel
}

type deploymentTemplateModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	MinStackVersion types.String `tfsdk:"min_stack_version"`
	Hidden          types.Bool   `tfsdk:"hidden"`
}
