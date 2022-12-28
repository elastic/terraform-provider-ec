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

func (plan DeploymentTF) UpdateRequest(ctx context.Context, client *api.API, state DeploymentTF) (*models.DeploymentUpdateRequest, diag.Diagnostics) {
	var result = models.DeploymentUpdateRequest{
		Name:         plan.Name.Value,
		Alias:        plan.Alias.Value,
		PruneOrphans: ec.Bool(true),
		Resources:    &models.DeploymentUpdateResources{},
		Settings:     &models.DeploymentUpdateSettings{},
		Metadata:     &models.DeploymentUpdateMetadata{},
	}

	dtID := plan.DeploymentTemplateId.Value

	var diagsnostics diag.Diagnostics

	template, err := deptemplateapi.Get(deptemplateapi.GetParams{
		API:                        client,
		TemplateID:                 dtID,
		Region:                     plan.Region.Value,
		HideInstanceConfigurations: true,
	})
	if err != nil {
		diagsnostics.AddError("Deployment template get error", err.Error())
		return nil, diagsnostics
	}

	// When the deployment template is changed, we need to skip the missing
	// resource topologies to account for a new instance_configuration_id and
	// a different default value.
	skipEStopologies := plan.DeploymentTemplateId.Value != "" && plan.DeploymentTemplateId.Value != state.DeploymentTemplateId.Value && state.DeploymentTemplateId.Value != ""
	// If the deployment_template_id is changed, then we skip updating the
	// Elasticsearch topology to account for the case where the
	// instance_configuration_id changes, i.e. Hot / Warm, etc.
	// This might not be necessary going forward as we move to
	// tiered Elasticsearch nodes.

	useNodeRoles, diags := elasticsearchv2.UseNodeRoles(state.Version, plan.Version)

	if diags.HasError() {
		return nil, diags
	}

	elasticsearchPayload, diags := elasticsearchv2.ElasticsearchPayload(ctx, plan.Elasticsearch, template, dtID, plan.Version.Value, useNodeRoles, skipEStopologies)

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

	kibanaPayload, diags := kibanav2.KibanaPayload(ctx, plan.Kibana, template)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if kibanaPayload != nil {
		result.Resources.Kibana = append(result.Resources.Kibana, kibanaPayload)
	}

	apmPayload, diags := apmv2.ApmPayload(ctx, plan.Apm, template)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if apmPayload != nil {
		result.Resources.Apm = append(result.Resources.Apm, apmPayload)
	}

	integrationsServerPayload, diags := integrationsserverv2.IntegrationsServerPayload(ctx, plan.IntegrationsServer, template)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if integrationsServerPayload != nil {
		result.Resources.IntegrationsServer = append(result.Resources.IntegrationsServer, integrationsServerPayload)
	}

	enterpriseSearchPayload, diags := enterprisesearchv2.EnterpriseSearchesPayload(ctx, plan.EnterpriseSearch, template)
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
