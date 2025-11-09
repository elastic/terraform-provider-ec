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

// productTypesOrderInsensitivePlanModifier ignores order differences in product_types list
type productTypesOrderInsensitivePlanModifier struct{}

func (m productTypesOrderInsensitivePlanModifier) Description(ctx context.Context) string {
	return "Ignores order differences in product_types list when semantically equivalent"
}

func (m productTypesOrderInsensitivePlanModifier) MarkdownDescription(ctx context.Context) string {
	return "Ignores order differences in product_types list when semantically equivalent"
}

func (m productTypesOrderInsensitivePlanModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// If either value is null or unknown, don't modify
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() || req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	// Get both lists
	var planItems, stateItems []resource_security_project.ProductTypesValue
	req.PlanValue.ElementsAs(ctx, &planItems, false)
	req.StateValue.ElementsAs(ctx, &stateItems, false)

	// If different lengths, they're actually different
	if len(planItems) != len(stateItems) {
		return
	}

	// Create maps of product_line -> product_tier for comparison
	planMap := make(map[string]string)
	for _, item := range planItems {
		planMap[item.ProductLine.ValueString()] = item.ProductTier.ValueString()
	}

	stateMap := make(map[string]string)
	for _, item := range stateItems {
		stateMap[item.ProductLine.ValueString()] = item.ProductTier.ValueString()
	}

	// If maps are equal, use state value (same content, different order)
	mapsEqual := len(planMap) == len(stateMap)
	if mapsEqual {
		for k, v := range planMap {
			if stateMap[k] != v {
				mapsEqual = false
				break
			}
		}
	}

	if mapsEqual {
		resp.PlanValue = req.StateValue
	}
}

func (sec securityModelReader) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_security_project.SecurityProjectResourceSchema(ctx)

	// Add plan modifiers to admin_features_package and product_types to preserve state values
	// when these fields are not configured. The API returns these values, and they may change
	// over time (e.g., tier upgrades), but if not explicitly configured we should keep the
	// current state value rather than forcing a recomputation.
	adminFeaturesAttr := resp.Schema.Attributes["admin_features_package"].(schema.StringAttribute)
	adminFeaturesAttr.PlanModifiers = append(adminFeaturesAttr.PlanModifiers, stringplanmodifier.UseStateForUnknown())
	resp.Schema.Attributes["admin_features_package"] = adminFeaturesAttr

	productTypesAttr := resp.Schema.Attributes["product_types"].(schema.ListNestedAttribute)
	productTypesAttr.PlanModifiers = append(productTypesAttr.PlanModifiers,
		listplanmodifier.UseStateForUnknown(),
		productTypesOrderInsensitivePlanModifier{},
	)
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

	// Populate admin_features_package from API response when available
	// If API doesn't return it, preserve the configured/state value
	if resp.JSON200.AdminFeaturesPackage != nil {
		pkgStr := string(*resp.JSON200.AdminFeaturesPackage)
		model.AdminFeaturesPackage = basetypes.NewStringValue(pkgStr)
	} else if model.AdminFeaturesPackage.IsNull() || model.AdminFeaturesPackage.IsUnknown() {
		// Only set to null if it wasn't already configured
		model.AdminFeaturesPackage = basetypes.NewStringNull()
	}
	// Otherwise, preserve the existing configured value

	// Populate product_types from API response when available
	if resp.JSON200.ProductTypes != nil {
		// If we have product_types in the state/config, we want to preserve that ordering
		// to avoid inconsistent results. Otherwise, use API ordering.
		var sourceProductTypes []resource_security_project.ProductTypesValue
		if !model.ProductTypes.IsNull() && !model.ProductTypes.IsUnknown() {
			model.ProductTypes.ElementsAs(ctx, &sourceProductTypes, false)
		}

		productTypeValues := []attr.Value{}

		if len(sourceProductTypes) > 0 {
			// Use the ordering from state/config, but with values from API
			apiProductTypesMap := make(map[string]serverless.SecurityProductType)
			for _, pt := range *resp.JSON200.ProductTypes {
				apiProductTypesMap[string(pt.ProductLine)] = pt
			}

			// Build result in the same order as source
			for _, sourcePt := range sourceProductTypes {
				productLine := sourcePt.ProductLine.ValueString()
				if apiPt, exists := apiProductTypesMap[productLine]; exists {
					productTypeValue, diags := resource_security_project.NewProductTypesValue(
						resource_security_project.ProductTypesValue{}.AttributeTypes(ctx),
						map[string]attr.Value{
							"product_line": basetypes.NewStringValue(string(apiPt.ProductLine)),
							"product_tier": basetypes.NewStringValue(string(apiPt.ProductTier)),
						},
					)
					if diags.HasError() {
						return false, model, diags
					}
					productTypeValues = append(productTypeValues, productTypeValue)
					delete(apiProductTypesMap, productLine)
				}
			}

			// Add any new product types from API that weren't in source
			for _, apiPt := range apiProductTypesMap {
				productTypeValue, diags := resource_security_project.NewProductTypesValue(
					resource_security_project.ProductTypesValue{}.AttributeTypes(ctx),
					map[string]attr.Value{
						"product_line": basetypes.NewStringValue(string(apiPt.ProductLine)),
						"product_tier": basetypes.NewStringValue(string(apiPt.ProductTier)),
					},
				)
				if diags.HasError() {
					return false, model, diags
				}
				productTypeValues = append(productTypeValues, productTypeValue)
			}
		} else {
			// No source ordering, use API ordering
			for _, pt := range *resp.JSON200.ProductTypes {
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
		}

		productTypesList, diags := types.ListValue(
			resource_security_project.ProductTypesValue{}.Type(ctx),
			productTypeValues,
		)
		if diags.HasError() {
			return false, model, diags
		}
		model.ProductTypes = productTypesList
	} else {
		// If API doesn't return product_types, preserve the configured/state value
		if model.ProductTypes.IsNull() || model.ProductTypes.IsUnknown() {
			// Only set to null if it wasn't already configured
			model.ProductTypes = types.ListNull(resource_security_project.ProductTypesValue{}.Type(ctx))
		}
		// Otherwise, preserve the existing configured value
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
