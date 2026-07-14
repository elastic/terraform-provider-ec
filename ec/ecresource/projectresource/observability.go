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
		api:          observabilityApi{sleeper: realSleeper{}},
		name:         "observability",
	}
}

type observabilityModelReader struct{}

func (obs observabilityModelReader) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_observability_project.ObservabilityProjectResourceSchema(ctx)
	patchMetadataSchema(resp)
}

func (obs observabilityModelReader) ReadFrom(ctx context.Context, getter modelGetter) (*resource_observability_project.ObservabilityProjectModel, diag.Diagnostics) {
	return readFrom[resource_observability_project.ObservabilityProjectModel](ctx, getter)
}

func (obs observabilityModelReader) GetID(model resource_observability_project.ObservabilityProjectModel) string {
	return model.Id.ValueString()
}

func (obs observabilityModelReader) Modify(plan resource_observability_project.ObservabilityProjectModel, state resource_observability_project.ObservabilityProjectModel, cfg resource_observability_project.ObservabilityProjectModel) resource_observability_project.ObservabilityProjectModel {
	plan.Credentials = useStateForUnknown(plan.Credentials, state.Credentials)
	plan.Endpoints = useStateForUnknown(plan.Endpoints, state.Endpoints)
	plan.PrivateEndpoints = useStateForUnknown(plan.PrivateEndpoints, state.PrivateEndpoints)
	plan.Metadata = useStateForUnknown(plan.Metadata, state.Metadata)
	plan.Linked = useStateForUnknownOrNull(plan.Linked, state.Linked, resource_observability_project.NewLinkedValueNull())
	if plan.ProductTier.IsUnknown() && !state.ProductTier.IsNull() {
		plan.ProductTier = state.ProductTier
	}

	nameHasChanged := !plan.Name.Equal(state.Name)
	aliasIsConfigured := util.IsKnown(cfg.Alias)
	aliasHasChanged := !plan.Alias.Equal(state.Alias)

	cloudIDIsUnknown := nameHasChanged || aliasHasChanged
	aliasIsUnknown := nameHasChanged && !aliasIsConfigured
	endpointsAreUnknown := aliasHasChanged || (!aliasIsConfigured && nameHasChanged)

	if aliasIsUnknown {
		plan.Alias = basetypes.NewStringUnknown()
	}

	if cloudIDIsUnknown {
		plan.CloudId = basetypes.NewStringUnknown()
	}

	if endpointsAreUnknown {
		plan.Endpoints = resource_observability_project.NewEndpointsValueUnknown()
		plan.PrivateEndpoints = resource_observability_project.NewPrivateEndpointsValueUnknown()
	}

	// system_tags includes _alias, which is derived from the project alias/name.
	// When either changes, system_tags must be recomputed by Read rather than
	// preserved from state, otherwise the stale _alias causes an inconsistent
	// result after apply.
	if cloudIDIsUnknown && !plan.Metadata.IsUnknown() && !plan.Metadata.IsNull() {
		plan.Metadata.SystemTags = types.MapUnknown(types.StringType)
	}

	return plan
}

type observabilityApi struct {
	client  serverless.ClientWithResponsesInterface
	sleeper sleeper
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

	if !model.ProductTier.IsNull() && !model.ProductTier.IsUnknown() {
		productTier := serverless.ObservabilityProjectProductTier(model.ProductTier.ValueString())
		createBody.ProductTier = &productTier
	}

	createBody.TrafficFilters = expandTrafficFilterIdsForCreate(ctx, model.TrafficFilterIds)

	if util.IsKnown(model.Metadata) && !model.Metadata.IsNull() {
		metaReq, metaDiags := projectMetadataRequestFromTFMetadata(ctx, model.Metadata.Tags)
		if metaDiags.HasError() {
			return model, metaDiags
		}
		if metaReq != nil {
			createBody.Metadata = metaReq
		}
	}

	createBody.Linked = expandLinkedForCreateObservability(model)

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

	if resp.JSON201.ProductTier != nil {
		model.ProductTier = basetypes.NewStringValue(string(*resp.JSON201.ProductTier))
	} else if model.ProductTier.IsUnknown() {
		model.ProductTier = basetypes.NewStringValue(string(serverless.ObservabilityProjectProductTierComplete))
	}

	linked, linkedDiags := flattenObservabilityLinked(ctx, resp.JSON201.Linked)
	if linkedDiags.HasError() {
		return model, linkedDiags
	}
	model.Linked = linked

	return model, diags
}

func (obs observabilityApi) Patch(ctx context.Context, plan, state resource_observability_project.ObservabilityProjectModel) diag.Diagnostics {
	updateBody := serverless.PatchObservabilityProjectRequest{
		Name: plan.Name.ValueStringPointer(),
	}

	if plan.Alias.ValueString() != "" {
		updateBody.Alias = plan.Alias.ValueStringPointer()
	}

	if !plan.ProductTier.IsNull() && !plan.ProductTier.IsUnknown() {
		productTier := serverless.ObservabilityProjectProductTier(plan.ProductTier.ValueString())
		updateBody.ProductTier = &productTier
	}

	updateBody.TrafficFilters = expandTrafficFilterIdsForPatch(ctx, plan.TrafficFilterIds)

	stateTags := types.MapNull(types.StringType)
	if util.IsKnown(state.Metadata) && !state.Metadata.IsNull() {
		stateTags = state.Metadata.Tags
	}
	if util.IsKnown(plan.Metadata) && !plan.Metadata.IsNull() {
		om, metaDiags := optionalMetadataForTagPatch(ctx, plan.Metadata.Tags, stateTags)
		if metaDiags.HasError() {
			return metaDiags
		}
		if om != nil {
			updateBody.Metadata = om
		}
	}

	updateBody.Linked = expandLinkedForPatchObservability(plan, state)

	resp, err := obs.client.PatchObservabilityProjectWithResponse(ctx, plan.Id.ValueString(), nil, updateBody)
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
	return waitForProjectInitialised(ctx, contextualSleep, func(ctx context.Context, id string) (serverless.ProjectStatusPhase, error) {
		resp, err := obs.client.GetObservabilityProjectStatusWithResponse(ctx, id)
		if err != nil {
			return "", err
		}
		if resp.JSON200 == nil {
			return "", fmt.Errorf("failed to get observability_project status: %d %s\n%s",
				resp.StatusCode(), resp.Status(), resp.Body)
		}
		return resp.JSON200.Phase, nil
	}, model.Id.ValueString())
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

	tagsVal, tagsDiags := metadataTagsFromAPI(ctx, resp.JSON200.Metadata.Tags)
	if tagsDiags.HasError() {
		return false, model, tagsDiags
	}
	metadataValues["tags"] = tagsVal

	systemTagsVal, systemTagsDiags := metadataSystemTagsFromAPI(ctx, resp.JSON200.Metadata.SystemTags)
	if systemTagsDiags.HasError() {
		return false, model, systemTagsDiags
	}
	metadataValues["system_tags"] = systemTagsVal

	metadata, diags := resource_observability_project.NewMetadataValue(
		model.Metadata.AttributeTypes(ctx),
		metadataValues,
	)
	if diags.HasError() {
		return false, model, diags
	}
	model.Metadata = metadata

	linked, linkedDiags := flattenObservabilityLinked(ctx, resp.JSON200.Linked)
	if linkedDiags.HasError() {
		return false, model, linkedDiags
	}
	model.Linked = linked

	if resp.JSON200.PrivateEndpoints != nil {
		privateEP, peDiags := resource_observability_project.NewPrivateEndpointsValue(
			model.PrivateEndpoints.AttributeTypes(ctx),
			map[string]attr.Value{
				"apm":           basetypes.NewStringValue(resp.JSON200.PrivateEndpoints.Apm),
				"elasticsearch": basetypes.NewStringValue(resp.JSON200.PrivateEndpoints.Elasticsearch),
				"ingest":        basetypes.NewStringValue(resp.JSON200.PrivateEndpoints.Ingest),
				"kibana":        basetypes.NewStringValue(resp.JSON200.PrivateEndpoints.Kibana),
			},
		)
		if peDiags.HasError() {
			return false, model, peDiags
		}
		model.PrivateEndpoints = privateEP
	} else {
		model.PrivateEndpoints = resource_observability_project.NewPrivateEndpointsValueNull()
	}

	model.Name = basetypes.NewStringValue(resp.JSON200.Name)
	model.RegionId = basetypes.NewStringValue(resp.JSON200.RegionId)
	model.Type = basetypes.NewStringValue(string(resp.JSON200.Type))

	// Set product_tier from API response, defaulting to "complete" if not present
	if resp.JSON200.ProductTier != nil {
		model.ProductTier = basetypes.NewStringValue(string(*resp.JSON200.ProductTier))
	} else {
		// Default value as per schema
		model.ProductTier = basetypes.NewStringValue(string(serverless.ObservabilityProjectProductTierComplete))
	}

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

func flattenObservabilityLinked(ctx context.Context, linked *serverless.LinkConfiguration) (resource_observability_project.LinkedValue, diag.Diagnostics) {
	if linked == nil || len(linked.Projects) == 0 {
		return resource_observability_project.NewLinkedValueNull(), nil
	}

	projectsMap := make(map[string]attr.Value, len(linked.Projects))
	statusesMap := make(map[string]attr.Value, len(linked.Projects))
	for projectID, project := range linked.Projects {
		pv, projectDiags := resource_observability_project.NewProjectsValue(
			resource_observability_project.ProjectsValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"type": basetypes.NewStringValue(string(project.Type)),
			},
		)
		if projectDiags.HasError() {
			return resource_observability_project.NewLinkedValueUnknown(), projectDiags
		}
		projectsMap[projectID] = pv
		statusesMap[projectID] = basetypes.NewStringValue(string(project.Status))
	}

	projects, projectsDiags := types.MapValue(resource_observability_project.ProjectsValue{}.Type(ctx), projectsMap)
	if projectsDiags.HasError() {
		return resource_observability_project.NewLinkedValueUnknown(), projectsDiags
	}

	statuses, statusesDiags := types.MapValue(types.StringType, statusesMap)
	if statusesDiags.HasError() {
		return resource_observability_project.NewLinkedValueUnknown(), statusesDiags
	}

	typedLinked, linkedDiags := resource_observability_project.NewLinkedValue(
		resource_observability_project.LinkedValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"projects": projects,
			"statuses": statuses,
		},
	)
	if linkedDiags.HasError() {
		return resource_observability_project.NewLinkedValueUnknown(), linkedDiags
	}

	return typedLinked, nil
}

func expandLinkedForCreateObservability(model resource_observability_project.ObservabilityProjectModel) *serverless.CreateLinkedRequest {
	if !util.IsKnown(model.Linked) || model.Linked.IsNull() || model.Linked.Projects.IsNull() {
		return nil
	}

	projects := make(map[string]serverless.CreateLinkedProjectRequest, len(model.Linked.Projects.Elements()))
	for projectID, v := range model.Linked.Projects.Elements() {
		pv, ok := v.(resource_observability_project.ProjectsValue)
		if !ok {
			continue
		}
		projects[projectID] = serverless.CreateLinkedProjectRequest{
			Type: serverless.ProjectType(pv.ProjectsType.ValueString()),
		}
	}

	if len(projects) == 0 {
		return nil
	}
	return &serverless.CreateLinkedRequest{Projects: projects}
}

func expandLinkedForPatchObservability(plan, state resource_observability_project.ObservabilityProjectModel) *serverless.OptionalLinkConfiguration {
	var planProjects, stateProjects basetypes.MapValue
	if util.IsKnown(plan.Linked) && !plan.Linked.IsNull() {
		planProjects = plan.Linked.Projects
	}
	if util.IsKnown(state.Linked) && !state.Linked.IsNull() {
		stateProjects = state.Linked.Projects
	}

	return expandLinkedProjectsForPatch(planProjects, stateProjects, func(v attr.Value) *serverless.OptionalLinkedProject {
		pv, ok := v.(resource_observability_project.ProjectsValue)
		if !ok {
			return nil
		}
		return &serverless.OptionalLinkedProject{
			Type: serverless.ProjectType(pv.ProjectsType.ValueString()),
		}
	})
}
