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
	"time"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_security_project"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func NewSecurityProjectResource() *Resource[resource_security_project.SecurityProjectModel] {
	return &Resource[resource_security_project.SecurityProjectModel]{
		modelHandler: securityModelReader{},
		api:          securityApi{sleeper: realSleeper{}},
		name:         "security",
	}
}

type securityModelReader struct{}

func (sec securityModelReader) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_security_project.SecurityProjectResourceSchema(ctx)
}

func (sec securityModelReader) ReadFrom(ctx context.Context, getter modelGetter) (*resource_security_project.SecurityProjectModel, diag.Diagnostics) {
	return readFrom[resource_security_project.SecurityProjectModel](ctx, getter)
}

func (sec securityModelReader) GetID(model resource_security_project.SecurityProjectModel) string {
	return model.Id.ValueString()
}

func (sec securityModelReader) Modify(plan resource_security_project.SecurityProjectModel, state resource_security_project.SecurityProjectModel, cfg resource_security_project.SecurityProjectModel) resource_security_project.SecurityProjectModel {
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
		plan.Endpoints = resource_security_project.NewEndpointsValueUnknown()
	}

	return plan
}

type securityApi struct {
	client  serverless.ClientWithResponsesInterface
	sleeper sleeper
}

func (sec securityApi) Ready() bool {
	return sec.client != nil
}

func (sec securityApi) WithClient(client serverless.ClientWithResponsesInterface) api[resource_security_project.SecurityProjectModel] {
	sec.client = client
	return sec
}

func (sec securityApi) Create(ctx context.Context, model resource_security_project.SecurityProjectModel) (resource_security_project.SecurityProjectModel, diag.Diagnostics) {
	createBody := serverless.CreateSecurityProjectRequest{
		Name:     model.Name.ValueString(),
		RegionId: model.RegionId.ValueString(),
	}

	if model.Alias.ValueString() != "" {
		createBody.Alias = model.Alias.ValueStringPointer()
	}

	if model.AdminFeaturesPackage.ValueString() != "" {
		createBody.AdminFeaturesPackage = (*serverless.SecurityAdminFeaturesPackage)(model.AdminFeaturesPackage.ValueStringPointer())
	}

	if util.IsKnown(model.ProductTypes) {
		var productTypes []resource_security_project.ProductTypesValue
		diags := model.ProductTypes.ElementsAs(ctx, &productTypes, false)
		if diags.HasError() {
			return model, diags
		}

		createProductTypes := []serverless.SecurityProductType{}
		for _, productType := range productTypes {
			createProductTypes = append(createProductTypes, serverless.SecurityProductType{
				ProductLine: serverless.SecurityProductLine(productType.ProductLine.ValueString()),
				ProductTier: serverless.SecurityProductTier(productType.ProductTier.ValueString()),
			})
		}

		createBody.ProductTypes = &createProductTypes
	}

	resp, err := sec.client.CreateSecurityProjectWithResponse(ctx, createBody)
	if err != nil {
		return model, diag.Diagnostics{
			diag.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}

	if resp.JSON201 == nil {
		return model, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Failed to create security_project",
				fmt.Sprintf("The API request failed with: %d %s\n%s",
					resp.StatusCode(),
					resp.Status(),
					resp.Body),
			),
		}
	}

	model.Id = types.StringValue(resp.JSON201.Id)

	creds, diags := resource_security_project.NewCredentialsValue(
		model.Credentials.AttributeTypes(ctx),
		map[string]attr.Value{
			"username": types.StringValue(resp.JSON201.Credentials.Username),
			"password": types.StringValue(resp.JSON201.Credentials.Password),
		},
	)
	model.Credentials = creds
	return model, diags
}

func (sec securityApi) Patch(ctx context.Context, model resource_security_project.SecurityProjectModel) diag.Diagnostics {
	updateBody := serverless.PatchSecurityProjectRequest{
		Name: model.Name.ValueStringPointer(),
	}

	if model.Alias.ValueString() != "" {
		updateBody.Alias = model.Alias.ValueStringPointer()
	}

	resp, err := sec.client.PatchSecurityProjectWithResponse(ctx, model.Id.ValueString(), nil, updateBody)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}

	if resp.JSON200 == nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Failed to update security_project",
				fmt.Sprintf("The API request failed with: %d %s\n%s",
					resp.StatusCode(),
					resp.Status(),
					resp.Body),
			),
		}
	}

	return nil
}

func (sec securityApi) EnsureInitialised(ctx context.Context, model resource_security_project.SecurityProjectModel) diag.Diagnostics {
	id := model.Id.ValueString()
	for {
		resp, err := sec.client.GetSecurityProjectStatusWithResponse(ctx, id)
		if err != nil {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(err.Error(), err.Error()),
			}
		}

		if resp.JSON200 == nil {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Failed to get security_project status",
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

		sec.sleeper.Sleep(200 * time.Millisecond)
	}
}

func (sec securityApi) Read(ctx context.Context, id string, model resource_security_project.SecurityProjectModel) (bool, resource_security_project.SecurityProjectModel, diag.Diagnostics) {
	resp, err := sec.client.GetSecurityProjectWithResponse(ctx, id)
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
				"Failed to read security_project",
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

	endpoints, diags := resource_security_project.NewEndpointsValue(
		model.Endpoints.AttributeTypes(ctx),
		map[string]attr.Value{
			"elasticsearch": basetypes.NewStringValue(resp.JSON200.Endpoints.Elasticsearch),
			"kibana":        basetypes.NewStringValue(resp.JSON200.Endpoints.Kibana),
			"ingest":        basetypes.NewStringValue(resp.JSON200.Endpoints.Ingest),
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

	metadata, diags := resource_security_project.NewMetadataValue(
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

func (sec securityApi) Delete(ctx context.Context, model resource_security_project.SecurityProjectModel) diag.Diagnostics {
	resp, err := sec.client.DeleteSecurityProjectWithResponse(ctx, model.Id.ValueString(), nil)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic("Failed to delete security_project", err.Error()),
		}
	}

	statusCode := resp.StatusCode()
	if statusCode != 200 && statusCode != 404 {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Request to delete security_project failed",
				fmt.Sprintf("The API request failed with: %d %s\n%s",
					resp.StatusCode(),
					resp.Status(),
					resp.Body),
			),
		}
	}

	return nil
}
