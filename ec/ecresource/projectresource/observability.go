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

package projectresource

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_observability_project"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func NewObservabilityProjectResource() *Resource[resource_observability_project.ObservabilityProjectModel] {
	return &Resource[resource_observability_project.ObservabilityProjectModel]{
		modelHandler: observabilityModelReader{},
		api:          observabilityApi{},
		name:         "observability",
	}
}

type observabilityModelReader struct{}

func (obs observabilityModelReader) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_observability_project.ObservabilityProjectResourceSchema(ctx)
}

func (obs observabilityModelReader) ReadFrom(ctx context.Context, getter modelGetter) (*resource_observability_project.ObservabilityProjectModel, diag.Diagnostics) {
	var model *resource_observability_project.ObservabilityProjectModel
	diags := getter.Get(ctx, &model)

	return model, diags
}

func (obs observabilityModelReader) GetID(model resource_observability_project.ObservabilityProjectModel) string {
	return model.Id.ValueString()
}

func (obs observabilityModelReader) Modify(plan resource_observability_project.ObservabilityProjectModel, state resource_observability_project.ObservabilityProjectModel, cfg resource_observability_project.ObservabilityProjectModel) resource_observability_project.ObservabilityProjectModel {
	plan.Credentials = useStateForUnknown(plan.Credentials, state.Credentials)
	plan.Endpoints = useStateForUnknown(plan.Endpoints, state.Endpoints)
	plan.Metadata = useStateForUnknown(plan.Metadata, state.Metadata)

	nameHasChanged := !plan.Name.Equal(state.Name)
	aliasIsConfigured := util.IsKnown(cfg.Alias)
	aliasHasChanged := !plan.Alias.Equal(state.Alias)

	cloudIDIsUnknown := nameHasChanged || aliasHasChanged
	aliasIsUnknown := nameHasChanged && !aliasIsConfigured
	endpointsAreUnknown := aliasHasChanged || (!aliasIsConfigured && nameHasChanged)

	if cloudIDIsUnknown {
		plan.CloudId = basetypes.NewStringUnknown()
	}

	if aliasIsUnknown {
		plan.Alias = basetypes.NewStringUnknown()
	}

	if endpointsAreUnknown {
		plan.Endpoints = resource_observability_project.NewEndpointsValueUnknown()
	}

	return plan
}

type observabilityApi struct {
	client serverless.ClientWithResponsesInterface
}

func (obs observabilityApi) Ready() bool {
	return obs.client != nil
}

func (obs observabilityApi) WithClient(client serverless.ClientWithResponsesInterface) api[resource_observability_project.ObservabilityProjectModel] {
	obs.client = client
	return obs
}

func (obs observabilityApi) Create(ctx context.Context, model resource_observability_project.ObservabilityProjectModel) (resource_observability_project.ObservabilityProjectModel, diag.Diagnostics) {
	createBody := serverless.CreateObservabilityProjectRequest{
		Name:     model.Name.ValueString(),
		RegionId: model.RegionId.ValueString(),
	}

	if model.Alias.ValueString() != "" {
		createBody.Alias = model.Alias.ValueStringPointer()
	}

	resp, err := obs.client.CreateObservabilityProjectWithResponse(ctx, createBody)
	if err != nil {
		return model, diag.Diagnostics{
			diag.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}

	if resp.JSON201 == nil {
		return model, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Failed to create observability_project",
				fmt.Sprintf("The API request failed with: %d %s\n%s",
					resp.StatusCode(),
					resp.Status(),
					resp.Body),
			),
		}
	}

	model.Id = types.StringValue(resp.JSON201.Id)

	creds, diags := resource_observability_project.NewCredentialsValue(
		model.Credentials.AttributeTypes(ctx),
		map[string]attr.Value{
			"username": types.StringValue(resp.JSON201.Credentials.Username),
			"password": types.StringValue(resp.JSON201.Credentials.Password),
		},
	)
	model.Credentials = creds
	return model, diags
}

func (obs observabilityApi) Patch(ctx context.Context, model resource_observability_project.ObservabilityProjectModel) diag.Diagnostics {
	updateBody := serverless.PatchObservabilityProjectRequest{
		Name: model.Name.ValueStringPointer(),
	}

	if model.Alias.ValueString() != "" {
		updateBody.Alias = model.Alias.ValueStringPointer()
	}

	resp, err := obs.client.PatchObservabilityProjectWithResponse(ctx, model.Id.ValueString(), nil, updateBody)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}

	if resp.JSON200 == nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Failed to update observability_project",
				fmt.Sprintf("The API request failed with: %d %s\n%s",
					resp.StatusCode(),
					resp.Status(),
					resp.Body),
			),
		}
	}

	return nil
}

func (obs observabilityApi) EnsureInitialised(ctx context.Context, model resource_observability_project.ObservabilityProjectModel) diag.Diagnostics {
	id := model.Id.ValueString()
	for {
		resp, err := obs.client.GetObservabilityProjectStatusWithResponse(ctx, id)
		if err != nil {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(err.Error(), err.Error()),
			}
		}

		if resp.JSON200 == nil {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Failed to get observability_project status",
					fmt.Sprintf("The API request failed with: %d %s\n%s",
						resp.StatusCode(),
						resp.Status(),
						resp.Body),
				),
			}
		}

		if resp.JSON200.Phase == serverless.Initialized {
			return nil
		}
	}
}

func (obs observabilityApi) Read(ctx context.Context, id string, model resource_observability_project.ObservabilityProjectModel) (bool, resource_observability_project.ObservabilityProjectModel, diag.Diagnostics) {
	resp, err := obs.client.GetObservabilityProjectWithResponse(ctx, id)
	if err != nil {
		return false, model, diag.Diagnostics{
			diag.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}

	if resp.HTTPResponse != nil && resp.HTTPResponse.StatusCode == http.StatusNotFound {
		return false, model, nil
	}

	if resp.JSON200 == nil {
		return false, model, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Failed to read observability_project",
				fmt.Sprintf("The API request failed with: %d %s\n%s",
					resp.StatusCode(),
					resp.Status(),
					resp.Body),
			),
		}
	}

	model.Id = basetypes.NewStringValue(id)
	model.Alias = basetypes.NewStringValue(reformatAlias(resp.JSON200.Alias, id))
	model.CloudId = basetypes.NewStringValue(resp.JSON200.CloudId)

	endpoints, diags := resource_observability_project.NewEndpointsValue(
		model.Endpoints.AttributeTypes(ctx),
		map[string]attr.Value{
			"elasticsearch": basetypes.NewStringValue(resp.JSON200.Endpoints.Elasticsearch),
			"kibana":        basetypes.NewStringValue(resp.JSON200.Endpoints.Kibana),
			"apm":           basetypes.NewStringValue(resp.JSON200.Endpoints.Apm),
		},
	)
	if diags.HasError() {
		return false, model, diags
	}
	model.Endpoints = endpoints

	metadataValues := map[string]attr.Value{
		"created_at":       basetypes.NewStringValue(resp.JSON200.Metadata.CreatedAt.String()),
		"created_by":       basetypes.NewStringValue(resp.JSON200.Metadata.CreatedBy),
		"organization_id":  basetypes.NewStringValue(resp.JSON200.Metadata.OrganizationId),
		"suspended_at":     basetypes.NewStringNull(),
		"suspended_reason": basetypes.NewStringPointerValue(resp.JSON200.Metadata.SuspendedReason),
	}

	if resp.JSON200.Metadata.SuspendedAt != nil {
		metadataValues["suspended_at"] = basetypes.NewStringValue(resp.JSON200.Metadata.SuspendedAt.String())
	}

	metadata, diags := resource_observability_project.NewMetadataValue(
		model.Metadata.AttributeTypes(ctx),
		metadataValues,
	)
	if diags.HasError() {
		return false, model, diags
	}
	model.Metadata = metadata

	model.Name = basetypes.NewStringValue(resp.JSON200.Name)
	model.RegionId = basetypes.NewStringValue(resp.JSON200.RegionId)
	model.Type = basetypes.NewStringValue(string(resp.JSON200.Type))

	return true, model, nil
}

func (obs observabilityApi) Delete(ctx context.Context, model resource_observability_project.ObservabilityProjectModel) diag.Diagnostics {
	resp, err := obs.client.DeleteObservabilityProjectWithResponse(ctx, model.Id.ValueString(), nil)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic("Failed to delete observability_project", err.Error()),
		}
	}

	statusCode := resp.StatusCode()
	if statusCode != 200 && statusCode != 404 {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Request to delete observability_project failed",
				fmt.Sprintf("The API request failed with: %d %s\n%s",
					resp.StatusCode(),
					resp.Status(),
					resp.Body),
			),
		}
	}

	return nil
}
