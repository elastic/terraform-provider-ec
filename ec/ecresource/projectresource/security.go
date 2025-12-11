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
	"sort"
	"time"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_security_project"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

// productTypesSemanticEqualityModifier uses the custom ProductTypesListValue semantic equality
type productTypesSemanticEqualityModifier struct{}

func (m productTypesSemanticEqualityModifier) Description(ctx context.Context) string {
	return "Ignores order differences in product_types list when semantically equivalent"
}

func (m productTypesSemanticEqualityModifier) MarkdownDescription(ctx context.Context) string {
	return "Ignores order differences in product_types list when semantically equivalent"
}

func (m productTypesSemanticEqualityModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// If either value is null or unknown, don't modify
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() || req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	// Use the ProductTypesListValue semantic equality check
	planValue := ProductTypesListValue{ListValue: req.PlanValue}
	stateValue := ProductTypesListValue{ListValue: req.StateValue}

	equal, diags := planValue.ListSemanticEquals(ctx, stateValue)
	resp.Diagnostics.Append(diags...)

	if equal {
		// Values are semantically equal, use state value to avoid false diffs
		resp.PlanValue = req.StateValue
	}
}

func (sec securityModelReader) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_security_project.SecurityProjectResourceSchema(ctx)

	// Add plan modifiers to admin_features_package and product_types
	// UseStateForUnknown prevents Terraform from showing a diff when the API doesn't return
	// these optional computed fields (they remain as the configured/state value)
	adminFeaturesAttr := resp.Schema.Attributes["admin_features_package"].(schema.StringAttribute)
	adminFeaturesAttr.PlanModifiers = []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	resp.Schema.Attributes["admin_features_package"] = adminFeaturesAttr

	// Add plan modifiers for product_types including order-insensitive semantic equality
	// The semantic equality modifier prevents spurious diffs when the API returns items in a different order
	productTypesAttr := resp.Schema.Attributes["product_types"].(schema.ListNestedAttribute)
	productTypesAttr.PlanModifiers = []planmodifier.List{
		listplanmodifier.UseStateForUnknown(),
		productTypesSemanticEqualityModifier{},
	}
	resp.Schema.Attributes["product_types"] = productTypesAttr
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

	// Populate admin_features_package from API response
	// The UseStateForUnknown plan modifier handles preserving the value when the API doesn't
	// return it, preventing Terraform from treating it as a configuration drift
	if resp.JSON200.AdminFeaturesPackage != nil {
		model.AdminFeaturesPackage = basetypes.NewStringValue(string(*resp.JSON200.AdminFeaturesPackage))
	} else {
		model.AdminFeaturesPackage = basetypes.NewStringNull()
	}

	// Populate product_types from API response
	// Sort the product_types to ensure consistent ordering, as the API may return them
	// in different orders. This prevents Terraform from detecting false differences.
	// The custom ProductTypesListType handles semantic equality (order-insensitive) during
	// planning, and the UseStateForUnknown plan modifier preserves the value when the API
	// doesn't return it, preventing configuration drift.
	if resp.JSON200.ProductTypes != nil {
		// Create a copy and sort by product_line, then product_tier for deterministic ordering
		productTypes := make([]serverless.SecurityProductType, len(*resp.JSON200.ProductTypes))
		copy(productTypes, *resp.JSON200.ProductTypes)

		sort.Slice(productTypes, func(i, j int) bool {
			if productTypes[i].ProductLine != productTypes[j].ProductLine {
				return productTypes[i].ProductLine < productTypes[j].ProductLine
			}
			return productTypes[i].ProductTier < productTypes[j].ProductTier
		})

		productTypeValues := []attr.Value{}
		for _, pt := range productTypes {
			// Validate that product line and tier are not empty
			if pt.ProductLine == "" || pt.ProductTier == "" {
				return false, model, diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid product type from API",
						fmt.Sprintf("API returned product type with empty product_line or product_tier"),
					),
				}
			}

			productTypeValue, diags := resource_security_project.NewProductTypesValue(
				resource_security_project.ProductTypesValue{}.AttributeTypes(ctx),
				map[string]attr.Value{
					"product_line": basetypes.NewStringValue(string(pt.ProductLine)),
					"product_tier": basetypes.NewStringValue(string(pt.ProductTier)),
				},
			)
			if diags.HasError() {
				return false, model, diags
			}
			productTypeValues = append(productTypeValues, productTypeValue)
		}

		productTypesList, diags := types.ListValueFrom(ctx,
			resource_security_project.ProductTypesValue{}.Type(ctx),
			productTypeValues,
		)
		if diags.HasError() {
			return false, model, diags
		}
		model.ProductTypes = productTypesList
	} else {
		model.ProductTypes = types.ListNull(resource_security_project.ProductTypesValue{}.Type(ctx))
	}

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
