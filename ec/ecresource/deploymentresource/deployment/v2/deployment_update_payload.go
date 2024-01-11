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

func (plan DeploymentTF) getBaseUpdatePayloads(ctx context.Context, client *api.API, state DeploymentTF) (*models.DeploymentUpdateResources, error) {
	newDtId := plan.DeploymentTemplateId.ValueString()
	prevDtId := state.DeploymentTemplateId.ValueString()

	template, err := deptemplateapi.Get(deptemplateapi.GetParams{
		API:                        client,
		TemplateID:                 newDtId,
		Region:                     plan.Region.ValueString(),
		HideInstanceConfigurations: true,
	})

	if err != nil {
		return nil, err
	}

	// Similarly to deployment creation, we don't want this setting to be inferred from deployment template
	removeAutoscalingTierOverridesFromTemplate(template)

	baseUpdatePayloads := &models.DeploymentUpdateResources{
		Apm:                template.DeploymentTemplate.Resources.Apm,
		Appsearch:          template.DeploymentTemplate.Resources.Appsearch,
		Elasticsearch:      template.DeploymentTemplate.Resources.Elasticsearch,
		EnterpriseSearch:   template.DeploymentTemplate.Resources.EnterpriseSearch,
		IntegrationsServer: template.DeploymentTemplate.Resources.IntegrationsServer,
		Kibana:             template.DeploymentTemplate.Resources.Kibana,
	}

	// If the deployment template has changed then we should use the template migration API
	// to build the base update payloads
	if newDtId != prevDtId && prevDtId != "" {
		// Get an update request from the template migration API
		migrateUpdateRequest, err := client.V1API.Deployments.MigrateDeploymentTemplate(
			deployments.NewMigrateDeploymentTemplateParams().WithDeploymentID(plan.Id.ValueString()).WithTemplateID(newDtId),
			client.AuthWriter,
		)

		if err != nil {
			return nil, err
		}

		if len(migrateUpdateRequest.Payload.Resources.Apm) > 0 {
			baseUpdatePayloads.Apm = migrateUpdateRequest.Payload.Resources.Apm
		}

		if len(migrateUpdateRequest.Payload.Resources.Appsearch) > 0 {
			baseUpdatePayloads.Appsearch = migrateUpdateRequest.Payload.Resources.Appsearch
		}

		if len(migrateUpdateRequest.Payload.Resources.Elasticsearch) > 0 {
			baseUpdatePayloads.Elasticsearch = migrateUpdateRequest.Payload.Resources.Elasticsearch
		}

		if len(migrateUpdateRequest.Payload.Resources.EnterpriseSearch) > 0 {
			baseUpdatePayloads.EnterpriseSearch = migrateUpdateRequest.Payload.Resources.EnterpriseSearch
		}

		if len(migrateUpdateRequest.Payload.Resources.IntegrationsServer) > 0 {
			baseUpdatePayloads.IntegrationsServer = migrateUpdateRequest.Payload.Resources.IntegrationsServer
		}

		if len(migrateUpdateRequest.Payload.Resources.Kibana) > 0 {
			baseUpdatePayloads.Kibana = migrateUpdateRequest.Payload.Resources.Kibana
		}
	}

	return baseUpdatePayloads, nil
}

func (plan DeploymentTF) UpdateRequest(ctx context.Context, client *api.API, state DeploymentTF) (*models.DeploymentUpdateRequest, diag.Diagnostics) {
	var result = models.DeploymentUpdateRequest{
		Name:         plan.Name.ValueString(),
		Alias:        plan.Alias.ValueString(),
		PruneOrphans: ec.Bool(true),
		Resources:    &models.DeploymentUpdateResources{},
		Settings:     &models.DeploymentUpdateSettings{},
		Metadata:     &models.DeploymentUpdateMetadata{},
	}

	dtID := plan.DeploymentTemplateId.ValueString()

	var diagsnostics diag.Diagnostics

	basePayloads, err := plan.getBaseUpdatePayloads(ctx, client, state)
	if err != nil {
		diagsnostics.AddError("Failed to get base update payloads for deployment", err.Error())
		return nil, diagsnostics
	}

	useNodeRoles, diags := elasticsearchv2.UseNodeRoles(ctx, state.Version, plan.Version, plan.Elasticsearch)

	if diags.HasError() {
		return nil, diags
	}

	elasticsearchPayload, diags := elasticsearchv2.ElasticsearchPayload(ctx, plan.Elasticsearch, &state.Elasticsearch, basePayloads, dtID, plan.Version.ValueString(), useNodeRoles)

	if diags.HasError() {
		diagsnostics.Append(diags...)
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
		diagsnostics.Append(diags...)
	}

	if kibanaPayload != nil {
		result.Resources.Kibana = append(result.Resources.Kibana, kibanaPayload)
	}

	apmPayload, diags := apmv2.ApmPayload(ctx, plan.Apm, basePayloads)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if apmPayload != nil {
		result.Resources.Apm = append(result.Resources.Apm, apmPayload)
	}

	integrationsServerPayload, diags := integrationsserverv2.IntegrationsServerPayload(ctx, plan.IntegrationsServer, basePayloads)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if integrationsServerPayload != nil {
		result.Resources.IntegrationsServer = append(result.Resources.IntegrationsServer, integrationsServerPayload)
	}

	enterpriseSearchPayload, diags := enterprisesearchv2.EnterpriseSearchesPayload(ctx, plan.EnterpriseSearch, basePayloads)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if enterpriseSearchPayload != nil {
		result.Resources.EnterpriseSearch = append(result.Resources.EnterpriseSearch, enterpriseSearchPayload)
	}

	observabilityPayload, diags := observabilityv2.ObservabilityPayload(ctx, plan.Observability, client)
	if diags.HasError() {
		diagsnostics.Append(diags...)
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
		diagsnostics.Append(diags...)
	}

	return &result, diagsnostics
}

func ensurePartialSnapshotStrategy(es *models.ElasticsearchPayload) {
	transient := es.Plan.Transient
	if transient == nil || transient.RestoreSnapshot == nil {
		return
	}
	transient.RestoreSnapshot.Strategy = "partial"
}
