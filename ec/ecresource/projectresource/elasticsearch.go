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
	"strings"
	"time"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_elasticsearch_project"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func NewElasticsearchProjectResource() *Resource[resource_elasticsearch_project.ElasticsearchProjectModel] {
	return &Resource[resource_elasticsearch_project.ElasticsearchProjectModel]{
		modelHandler: elasticsearchModelReader{},
		api: elasticsearchApi{
			sleeper: realSleeper{},
		},
		name: "elasticsearch",
	}
}

type elasticsearchModelReader struct{}

func (es elasticsearchModelReader) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_elasticsearch_project.ElasticsearchProjectResourceSchema(ctx)
	patchOptimizedForSchema(resp)
	patchMetadataSchema(resp)
}

func patchOptimizedForSchema(resp *resource.SchemaResponse) {
	optimizedForAttr := resp.Schema.Attributes["optimized_for"].(schema.StringAttribute)
	optimizedForAttr.Description = strings.ReplaceAll(optimizedForAttr.Description, "\n-", "\n\t-")
	optimizedForAttr.MarkdownDescription = strings.ReplaceAll(optimizedForAttr.MarkdownDescription, "\n-", "\n\t-")
	resp.Schema.Attributes["optimized_for"] = optimizedForAttr
}

func (es elasticsearchModelReader) ReadFrom(ctx context.Context, getter modelGetter) (*resource_elasticsearch_project.ElasticsearchProjectModel, diag.Diagnostics) {
	return readFrom[resource_elasticsearch_project.ElasticsearchProjectModel](ctx, getter)
}

func (es elasticsearchModelReader) GetID(model resource_elasticsearch_project.ElasticsearchProjectModel) string {
	return model.Id.ValueString()
}

func (es elasticsearchModelReader) Modify(plan resource_elasticsearch_project.ElasticsearchProjectModel, state resource_elasticsearch_project.ElasticsearchProjectModel, cfg resource_elasticsearch_project.ElasticsearchProjectModel) resource_elasticsearch_project.ElasticsearchProjectModel {
	plan.Credentials = useStateForUnknown(plan.Credentials, state.Credentials)
	plan.Endpoints = useStateForUnknown(plan.Endpoints, state.Endpoints)
	plan.PrivateEndpoints = useStateForUnknown(plan.PrivateEndpoints, state.PrivateEndpoints)
	plan.Linked = useStateForUnknownOrNull(plan.Linked, state.Linked, resource_elasticsearch_project.NewLinkedValueNull())

	nameHasChanged := !plan.Name.Equal(state.Name)
	aliasIsConfigured := util.IsKnown(cfg.Alias)
	aliasHasChanged := !plan.Alias.Equal(state.Alias)

	cloudIDIsUnknown := nameHasChanged || aliasHasChanged
	aliasIsUnknown := nameHasChanged && !aliasIsConfigured
	endpointsAreUnknown := aliasHasChanged || (!aliasIsConfigured && nameHasChanged)

	if aliasIsUnknown {
		plan.Alias = basetypes.NewStringUnknown()
	}

	plan.Metadata = preserveMetadataForPlan(plan.Metadata, state.Metadata)

	if cloudIDIsUnknown {
		plan.CloudId = basetypes.NewStringUnknown()
	}

	if endpointsAreUnknown {
		plan.Endpoints = resource_elasticsearch_project.NewEndpointsValueUnknown()
		plan.PrivateEndpoints = resource_elasticsearch_project.NewPrivateEndpointsValueUnknown()
	}

	return plan
}

type sleeper interface {
	Sleep(time.Duration)
}

type realSleeper struct{}

func (r realSleeper) Sleep(d time.Duration) {
	//lintignore:R018 // Intentionally wrapped for testability
	time.Sleep(d)
}

type elasticsearchApi struct {
	client  serverless.ClientWithResponsesInterface
	sleeper sleeper
}

func (es elasticsearchApi) Ready() bool {
	return es.client != nil
}

func (es elasticsearchApi) WithClient(client serverless.ClientWithResponsesInterface) api[resource_elasticsearch_project.ElasticsearchProjectModel] {
	es.client = client
	return es
}

func (es elasticsearchApi) Create(ctx context.Context, model resource_elasticsearch_project.ElasticsearchProjectModel) (resource_elasticsearch_project.ElasticsearchProjectModel, diag.Diagnostics) {
	createBody := serverless.CreateElasticsearchProjectRequest{
		Name:     model.Name.ValueString(),
		RegionId: model.RegionId.ValueString(),
	}

	if model.Alias.ValueString() != "" {
		createBody.Alias = model.Alias.ValueStringPointer()
	}

	if model.OptimizedFor.ValueString() != "" {
		createBody.OptimizedFor = (*serverless.ElasticsearchOptimizedFor)(model.OptimizedFor.ValueStringPointer())
	}

	if util.IsKnown(model.SearchLake) {
		createBody.SearchLake = &serverless.ElasticsearchSearchLake{}

		if util.IsKnown(model.SearchLake.BoostWindow) {
			boostWindow := int(model.SearchLake.BoostWindow.ValueInt64())
			createBody.SearchLake.BoostWindow = &boostWindow
		}

		if util.IsKnown(model.SearchLake.SearchPower) {
			searchPower := int(model.SearchLake.SearchPower.ValueInt64())
			createBody.SearchLake.SearchPower = &searchPower
		}
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

	createBody.Linked = expandLinkedForCreateElasticsearch(model)

	resp, err := es.client.CreateElasticsearchProjectWithResponse(ctx, createBody)
	if err != nil {
		return model, diag.Diagnostics{
			diag.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}

	if resp.JSON201 == nil {
		return model, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Failed to create elasticsearch_project",
				fmt.Sprintf("The API request failed with: %d %s\n%s",
					resp.StatusCode(),
					resp.Status(),
					resp.Body),
			),
		}
	}

	model.Id = types.StringValue(resp.JSON201.Id)

	creds, diags := resource_elasticsearch_project.NewCredentialsValue(
		model.Credentials.AttributeTypes(ctx),
		map[string]attr.Value{
			"username": types.StringValue(resp.JSON201.Credentials.Username),
			"password": types.StringValue(resp.JSON201.Credentials.Password),
		},
	)
	model.Credentials = creds

	linked, linkedDiags := flattenElasticsearchLinked(ctx, resp.JSON201.Linked)
	if linkedDiags.HasError() {
		return model, linkedDiags
	}
	model.Linked = linked

	return model, diags
}

func (es elasticsearchApi) Patch(ctx context.Context, plan, state resource_elasticsearch_project.ElasticsearchProjectModel) diag.Diagnostics {
	updateBody := serverless.PatchElasticsearchProjectRequest{
		Name: plan.Name.ValueStringPointer(),
	}

	if plan.Alias.ValueString() != "" {
		updateBody.Alias = plan.Alias.ValueStringPointer()
	}

	if util.IsKnown(plan.SearchLake) {
		updateBody.SearchLake = &serverless.OptionalElasticsearchSearchLake{}

		if util.IsKnown(plan.SearchLake.BoostWindow) {
			boostWindow := int(plan.SearchLake.BoostWindow.ValueInt64())
			updateBody.SearchLake.BoostWindow = &boostWindow
		}

		if util.IsKnown(plan.SearchLake.SearchPower) {
			searchPower := int(plan.SearchLake.SearchPower.ValueInt64())
			updateBody.SearchLake.SearchPower = &searchPower
		}
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

	updateBody.Linked = expandLinkedForPatchElasticsearch(plan)

	resp, err := es.client.PatchElasticsearchProjectWithResponse(ctx, plan.Id.ValueString(), nil, updateBody)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}

	if resp.JSON200 == nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Failed to update elasticsearch_project",
				fmt.Sprintf("The API request failed with: %d %s\n%s",
					resp.StatusCode(),
					resp.Status(),
					resp.Body),
			),
		}
	}

	return nil
}

func (es elasticsearchApi) EnsureInitialised(ctx context.Context, model resource_elasticsearch_project.ElasticsearchProjectModel) diag.Diagnostics {
	id := model.Id.ValueString()
	for {
		resp, err := es.client.GetElasticsearchProjectStatusWithResponse(ctx, id)
		if err != nil {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(err.Error(), err.Error()),
			}
		}

		if resp.JSON200 == nil {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Failed to get elasticsearch_project status",
					fmt.Sprintf("The API request failed with: %d %s\n%s",
						resp.StatusCode(),
						resp.Status(),
						resp.Body),
				),
			}
		}

		if resp.JSON200.Phase == serverless.ProjectStatusPhaseInitialized {
			return nil
		}

		es.sleeper.Sleep(200 * time.Millisecond)
	}
}

func (es elasticsearchApi) Read(ctx context.Context, id string, model resource_elasticsearch_project.ElasticsearchProjectModel) (bool, resource_elasticsearch_project.ElasticsearchProjectModel, diag.Diagnostics) {
	resp, err := es.client.GetElasticsearchProjectWithResponse(ctx, id)
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
				"Failed to read elasticsearch_project",
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

	endpoints, diags := resource_elasticsearch_project.NewEndpointsValue(
		model.Endpoints.AttributeTypes(ctx),
		map[string]attr.Value{
			"elasticsearch": basetypes.NewStringValue(resp.JSON200.Endpoints.Elasticsearch),
			"kibana":        basetypes.NewStringValue(resp.JSON200.Endpoints.Kibana),
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

	metadata, diags := resource_elasticsearch_project.NewMetadataValue(
		model.Metadata.AttributeTypes(ctx),
		metadataValues,
	)
	if diags.HasError() {
		return false, model, diags
	}
	model.Metadata = metadata

	linked, linkedDiags := flattenElasticsearchLinked(ctx, resp.JSON200.Linked)
	if linkedDiags.HasError() {
		return false, model, linkedDiags
	}
	model.Linked = linked

	if resp.JSON200.PrivateEndpoints != nil {
		privateEP, peDiags := resource_elasticsearch_project.NewPrivateEndpointsValue(
			model.PrivateEndpoints.AttributeTypes(ctx),
			map[string]attr.Value{
				"elasticsearch": basetypes.NewStringValue(resp.JSON200.PrivateEndpoints.Elasticsearch),
				"kibana":        basetypes.NewStringValue(resp.JSON200.PrivateEndpoints.Kibana),
			},
		)
		if peDiags.HasError() {
			return false, model, peDiags
		}
		model.PrivateEndpoints = privateEP
	} else {
		model.PrivateEndpoints = resource_elasticsearch_project.NewPrivateEndpointsValueNull()
	}

	model.Name = basetypes.NewStringValue(resp.JSON200.Name)
	model.OptimizedFor = basetypes.NewStringValue(string(resp.JSON200.OptimizedFor))
	model.RegionId = basetypes.NewStringValue(resp.JSON200.RegionId)
	model.Type = basetypes.NewStringValue(string(resp.JSON200.Type))

	searchLakeValues := map[string]attr.Value{
		"boost_window": basetypes.NewInt64Null(),
		"search_power": basetypes.NewInt64Null(),
	}

	if resp.JSON200.SearchLake != nil {
		if resp.JSON200.SearchLake.BoostWindow != nil {
			searchLakeValues["boost_window"] = basetypes.NewInt64Value(int64(*resp.JSON200.SearchLake.BoostWindow))
		}

		if resp.JSON200.SearchLake.SearchPower != nil {
			searchLakeValues["search_power"] = basetypes.NewInt64Value(int64(*resp.JSON200.SearchLake.SearchPower))
		}
	}
	searchLake, diags := resource_elasticsearch_project.NewSearchLakeValue(
		model.SearchLake.AttributeTypes(ctx),
		searchLakeValues,
	)
	if diags.HasError() {
		return false, model, nil
	}
	model.SearchLake = searchLake

	return true, model, nil
}

func (es elasticsearchApi) Delete(ctx context.Context, model resource_elasticsearch_project.ElasticsearchProjectModel) diag.Diagnostics {
	resp, err := es.client.DeleteElasticsearchProjectWithResponse(ctx, model.Id.ValueString(), nil)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic("Failed to delete elasticsearch_project", err.Error()),
		}
	}

	statusCode := resp.StatusCode()
	if statusCode != 200 && statusCode != 404 {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Request to delete elasticsearch_project failed",
				fmt.Sprintf("The API request failed with: %d %s\n%s",
					resp.StatusCode(),
					resp.Status(),
					resp.Body),
			),
		}
	}

	return nil
}

func flattenElasticsearchLinked(ctx context.Context, linked *serverless.LinkConfiguration) (resource_elasticsearch_project.LinkedValue, diag.Diagnostics) {
	if linked == nil || len(linked.Projects) == 0 {
		return resource_elasticsearch_project.NewLinkedValueNull(), nil
	}

	projectsMap := make(map[string]attr.Value, len(linked.Projects))
	for projectID, project := range linked.Projects {
		pv, projectDiags := resource_elasticsearch_project.NewProjectsValue(
			resource_elasticsearch_project.ProjectsValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"status": basetypes.NewStringValue(string(project.Status)),
				"type":   basetypes.NewStringValue(string(project.Type)),
			},
		)
		if projectDiags.HasError() {
			return resource_elasticsearch_project.NewLinkedValueUnknown(), projectDiags
		}
		projectsMap[projectID] = pv
	}

	projects, projectsDiags := types.MapValue(resource_elasticsearch_project.ProjectsValue{}.Type(ctx), projectsMap)
	if projectsDiags.HasError() {
		return resource_elasticsearch_project.NewLinkedValueUnknown(), projectsDiags
	}

	typedLinked, linkedDiags := resource_elasticsearch_project.NewLinkedValue(
		resource_elasticsearch_project.LinkedValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"projects": projects,
		},
	)
	if linkedDiags.HasError() {
		return resource_elasticsearch_project.NewLinkedValueUnknown(), linkedDiags
	}

	return typedLinked, nil
}

func expandLinkedForCreateElasticsearch(model resource_elasticsearch_project.ElasticsearchProjectModel) *serverless.CreateLinkedRequest {
	if !util.IsKnown(model.Linked) || model.Linked.IsNull() || model.Linked.Projects.IsNull() {
		return nil
	}

	projects := make(map[string]serverless.CreateLinkedProjectRequest, len(model.Linked.Projects.Elements()))
	for projectID, v := range model.Linked.Projects.Elements() {
		pv, ok := v.(resource_elasticsearch_project.ProjectsValue)
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

func expandLinkedForPatchElasticsearch(plan resource_elasticsearch_project.ElasticsearchProjectModel) *serverless.OptionalLinkConfiguration {
	if !util.IsKnown(plan.Linked) || plan.Linked.IsNull() || plan.Linked.Projects.IsNull() {
		return nil
	}

	projects := make(map[string]*serverless.OptionalLinkedProject, len(plan.Linked.Projects.Elements()))
	for projectID, v := range plan.Linked.Projects.Elements() {
		pv, ok := v.(resource_elasticsearch_project.ProjectsValue)
		if !ok {
			continue
		}
		projects[projectID] = &serverless.OptionalLinkedProject{
			Type: serverless.ProjectType(pv.ProjectsType.ValueString()),
		}
	}

	if len(projects) == 0 {
		return nil
	}
	return &serverless.OptionalLinkConfiguration{Projects: &projects}
}
