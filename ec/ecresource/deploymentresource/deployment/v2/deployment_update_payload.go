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

package v2

import (
	"context"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deptemplateapi"
	"github.com/elastic/cloud-sdk-go/pkg/client/deployments"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"

	apmv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/apm/v2"
	elasticsearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
	enterprisesearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/enterprisesearch/v2"
	integrationsserverv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/integrationsserver/v2"
	kibanav2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/kibana/v2"
	observabilityv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/observability/v2"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func (plan DeploymentTF) getBaseUpdatePayloads(ctx context.Context, client *api.API, state DeploymentTF, migrateTemplateRequest *deployments.MigrateDeploymentTemplateOK) (*models.DeploymentUpdateResources, diag.Diagnostics) {
	var diags diag.Diagnostics

	newDtId := plan.DeploymentTemplateId.ValueString()
	prevDtId := state.DeploymentTemplateId.ValueString()

	template, err := deptemplateapi.Get(deptemplateapi.GetParams{
		API:                        client,
		TemplateID:                 newDtId,
		Region:                     plan.Region.ValueString(),
		HideInstanceConfigurations: true,
	})

	if err != nil {
		diags.AddError("Failed to get template", err.Error())
		return nil, diags
	}

	baseUpdatePayloads := &models.DeploymentUpdateResources{
		Apm:                template.DeploymentTemplate.Resources.Apm,
		Appsearch:          template.DeploymentTemplate.Resources.Appsearch,
		Elasticsearch:      template.DeploymentTemplate.Resources.Elasticsearch,
		EnterpriseSearch:   template.DeploymentTemplate.Resources.EnterpriseSearch,
		IntegrationsServer: template.DeploymentTemplate.Resources.IntegrationsServer,
		Kibana:             template.DeploymentTemplate.Resources.Kibana,
	}

	planHasNodeTypes, diags := elasticsearchv2.PlanHasNodeTypes(ctx, plan.Elasticsearch)

	if diags.HasError() {
		return nil, diags
	}

	templateChanged := newDtId != prevDtId && prevDtId != ""

	// If the deployment template has changed or MigrateToLatestHardware is true, we should use the template migration
	// API to build the base update payloads
	migrateToLatest := plan.MigrateToLatestHardware.ValueBool() || templateChanged

	// Template migration isn't available for deployments using node types
	if migrateToLatest && !planHasNodeTypes {
		// If the template has changed, we can't use the migrate request from private state.
		// In this case, we fetch a new update request again from the template migration API
		if migrateTemplateRequest == nil {
			migrateTemplateRequest, err = client.V1API.Deployments.MigrateDeploymentTemplate(
				deployments.NewMigrateDeploymentTemplateParams().WithDeploymentID(plan.Id.ValueString()).WithTemplateID(newDtId),
				client.AuthWriter,
			)
		}

		if err != nil {
			diags.AddError("Failed to get template migration request", err.Error())
			return nil, diags
		}

		if len(migrateTemplateRequest.Payload.Resources.Apm) > 0 {
			baseUpdatePayloads.Apm = migrateTemplateRequest.Payload.Resources.Apm
		}

		if len(migrateTemplateRequest.Payload.Resources.Appsearch) > 0 {
			baseUpdatePayloads.Appsearch = migrateTemplateRequest.Payload.Resources.Appsearch
		}

		if len(migrateTemplateRequest.Payload.Resources.Elasticsearch) > 0 {
			baseUpdatePayloads.Elasticsearch = migrateTemplateRequest.Payload.Resources.Elasticsearch
		}

		if len(migrateTemplateRequest.Payload.Resources.EnterpriseSearch) > 0 {
			baseUpdatePayloads.EnterpriseSearch = migrateTemplateRequest.Payload.Resources.EnterpriseSearch
		}

		if len(migrateTemplateRequest.Payload.Resources.IntegrationsServer) > 0 {
			baseUpdatePayloads.IntegrationsServer = migrateTemplateRequest.Payload.Resources.IntegrationsServer
		}

		if len(migrateTemplateRequest.Payload.Resources.Kibana) > 0 {
			baseUpdatePayloads.Kibana = migrateTemplateRequest.Payload.Resources.Kibana
		}
	}

	return baseUpdatePayloads, diags
}

func (plan DeploymentTF) UpdateRequest(ctx context.Context, client *api.API, state DeploymentTF, migrateTemplateRequest *deployments.MigrateDeploymentTemplateOK) (*models.DeploymentUpdateRequest, diag.Diagnostics) {
	var result = models.DeploymentUpdateRequest{
		Name:         plan.Name.ValueString(),
		Alias:        plan.Alias.ValueString(),
		PruneOrphans: ec.Bool(true),
		Resources:    &models.DeploymentUpdateResources{},
		Settings:     &models.DeploymentUpdateSettings{},
		Metadata:     &models.DeploymentUpdateMetadata{},
	}

	dtID := plan.DeploymentTemplateId.ValueString()

	var diagnostics diag.Diagnostics

	basePayloads, diags := plan.getBaseUpdatePayloads(ctx, client, state, migrateTemplateRequest)

	if diags.HasError() {
		return nil, diags
	}

	useNodeRoles, diags := elasticsearchv2.UseNodeRoles(ctx, state.Version, plan.Version, plan.Elasticsearch)

	if diags.HasError() {
		return nil, diags
	}

	elasticsearchPayload, diags := elasticsearchv2.ElasticsearchPayload(ctx, plan.Elasticsearch, &state.Elasticsearch, basePayloads, dtID, plan.Version.ValueString(), useNodeRoles)

	if diags.HasError() {
		diagnostics.Append(diags...)
	}

	if elasticsearchPayload != nil {
		// if the restore snapshot operation has been specified, the snapshot restore
		// can't be full once the cluster has been created, so the Strategy must be set
		// to "partial".
		ensurePartialSnapshotStrategy(elasticsearchPayload)

		result.Resources.Elasticsearch = append(result.Resources.Elasticsearch, elasticsearchPayload)
	}

	kibanaPayload, diags := kibanav2.KibanaPayload(ctx, plan.Kibana, basePayloads)
	if diags.HasError() {
		diagnostics.Append(diags...)
	}

	if kibanaPayload != nil {
		result.Resources.Kibana = append(result.Resources.Kibana, kibanaPayload)
	}

	apmPayload, diags := apmv2.ApmPayload(ctx, plan.Apm, basePayloads)
	if diags.HasError() {
		diagnostics.Append(diags...)
	}

	if apmPayload != nil {
		result.Resources.Apm = append(result.Resources.Apm, apmPayload)
	}

	integrationsServerPayload, diags := integrationsserverv2.IntegrationsServerPayload(ctx, plan.IntegrationsServer, basePayloads)
	if diags.HasError() {
		diagnostics.Append(diags...)
	}

	if integrationsServerPayload != nil {
		result.Resources.IntegrationsServer = append(result.Resources.IntegrationsServer, integrationsServerPayload)
	}

	enterpriseSearchPayload, diags := enterprisesearchv2.EnterpriseSearchesPayload(ctx, plan.EnterpriseSearch, basePayloads)
	if diags.HasError() {
		diagnostics.Append(diags...)
	}

	if enterpriseSearchPayload != nil {
		result.Resources.EnterpriseSearch = append(result.Resources.EnterpriseSearch, enterpriseSearchPayload)
	}

	observabilityPayload, diags := observabilityv2.ObservabilityPayload(ctx, plan.Observability, client)
	if diags.HasError() {
		diagnostics.Append(diags...)
	}
	result.Settings.Observability = observabilityPayload

	// In order to stop shipping logs and metrics, an empty Observability
	// object must be passed, as opposed to a nil object when creating a
	// deployment without observability settings.
	if plan.Observability.IsNull() && !state.Observability.IsNull() {
		result.Settings.Observability = &models.DeploymentObservabilitySettings{}
	}

	result.Metadata.Tags, diags = converters.TypesMapToModelsTags(ctx, plan.Tags)
	if diags.HasError() {
		diagnostics.Append(diags...)
	}

	return &result, diagnostics
}

func ensurePartialSnapshotStrategy(es *models.ElasticsearchPayload) {
	transient := es.Plan.Transient
	if transient == nil || transient.RestoreSnapshot == nil {
		return
	}
	transient.RestoreSnapshot.Strategy = "partial"
}
