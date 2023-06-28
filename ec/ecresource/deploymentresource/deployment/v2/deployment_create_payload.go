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
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/esremoteclustersapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"

	apmv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/apm/v2"
	elasticsearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
	enterprisesearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/enterprisesearch/v2"
	integrationsserverv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/integrationsserver/v2"
	kibanav2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/kibana/v2"
	observabilityv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/observability/v2"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DeploymentTF struct {
	Id                         types.String `tfsdk:"id"`
	Alias                      types.String `tfsdk:"alias"`
	Version                    types.String `tfsdk:"version"`
	Region                     types.String `tfsdk:"region"`
	DeploymentTemplateId       types.String `tfsdk:"deployment_template_id"`
	Name                       types.String `tfsdk:"name"`
	RequestId                  types.String `tfsdk:"request_id"`
	ElasticsearchUsername      types.String `tfsdk:"elasticsearch_username"`
	ElasticsearchPassword      types.String `tfsdk:"elasticsearch_password"`
	ApmSecretToken             types.String `tfsdk:"apm_secret_token"`
	TrafficFilter              types.Set    `tfsdk:"traffic_filter"`
	Tags                       types.Map    `tfsdk:"tags"`
	Elasticsearch              types.Object `tfsdk:"elasticsearch"`
	Kibana                     types.Object `tfsdk:"kibana"`
	Apm                        types.Object `tfsdk:"apm"`
	IntegrationsServer         types.Object `tfsdk:"integrations_server"`
	EnterpriseSearch           types.Object `tfsdk:"enterprise_search"`
	Observability              types.Object `tfsdk:"observability"`
	ResetElasticsearchPassword types.Bool   `tfsdk:"reset_elasticsearch_password"`
}

func (dep DeploymentTF) CreateRequest(ctx context.Context, client *api.API) (*models.DeploymentCreateRequest, diag.Diagnostics) {
	var result = models.DeploymentCreateRequest{
		Name:      dep.Name.ValueString(),
		Alias:     dep.Alias.ValueString(),
		Resources: &models.DeploymentCreateResources{},
		Settings:  &models.DeploymentCreateSettings{},
		Metadata:  &models.DeploymentCreateMetadata{},
	}

	dtID := dep.DeploymentTemplateId.ValueString()
	version := dep.Version.ValueString()

	var diagsnostics diag.Diagnostics

	template, err := deptemplateapi.Get(deptemplateapi.GetParams{
		API:                        client,
		TemplateID:                 dtID,
		Region:                     dep.Region.ValueString(),
		HideInstanceConfigurations: true,
	})
	if err != nil {
		diagsnostics.AddError("Deployment template get error", err.Error())
		return nil, diagsnostics
	}

	baseUpdatePayloads := &models.DeploymentUpdateResources{
		Apm:                template.DeploymentTemplate.Resources.Apm,
		Appsearch:          template.DeploymentTemplate.Resources.Appsearch,
		Elasticsearch:      template.DeploymentTemplate.Resources.Elasticsearch,
		EnterpriseSearch:   template.DeploymentTemplate.Resources.EnterpriseSearch,
		IntegrationsServer: template.DeploymentTemplate.Resources.IntegrationsServer,
		Kibana:             template.DeploymentTemplate.Resources.Kibana,
	}

	useNodeRoles, err := elasticsearchv2.CompatibleWithNodeRoles(version)
	if err != nil {
		diagsnostics.AddError("Deployment parse error", err.Error())
		return nil, diagsnostics
	}

	elasticsearchPayload, diags := elasticsearchv2.ElasticsearchPayload(ctx, dep.Elasticsearch, baseUpdatePayloads, dtID, version, useNodeRoles)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if elasticsearchPayload != nil {
		result.Resources.Elasticsearch = []*models.ElasticsearchPayload{elasticsearchPayload}
	}

	kibanaPayload, diags := kibanav2.KibanaPayload(ctx, dep.Kibana, baseUpdatePayloads)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if kibanaPayload != nil {
		result.Resources.Kibana = []*models.KibanaPayload{kibanaPayload}
	}

	apmPayload, diags := apmv2.ApmPayload(ctx, dep.Apm, baseUpdatePayloads)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if apmPayload != nil {
		result.Resources.Apm = []*models.ApmPayload{apmPayload}
	}

	integrationsServerPayload, diags := integrationsserverv2.IntegrationsServerPayload(ctx, dep.IntegrationsServer, baseUpdatePayloads)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if integrationsServerPayload != nil {
		result.Resources.IntegrationsServer = []*models.IntegrationsServerPayload{integrationsServerPayload}
	}

	enterpriseSearchPayload, diags := enterprisesearchv2.EnterpriseSearchesPayload(ctx, dep.EnterpriseSearch, baseUpdatePayloads)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if enterpriseSearchPayload != nil {
		result.Resources.EnterpriseSearch = []*models.EnterpriseSearchPayload{enterpriseSearchPayload}
	}

	if diags := trafficFilterToModel(ctx, dep.TrafficFilter, &result); diags.HasError() {
		diagsnostics.Append(diags...)
	}

	observabilityPayload, diags := observabilityv2.ObservabilityPayload(ctx, dep.Observability, client)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	result.Settings.Observability = observabilityPayload

	result.Metadata.Tags, diags = converters.TypesMapToModelsTags(ctx, dep.Tags)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	return &result, diagsnostics
}

// trafficFilterToModel expands the flattened "traffic_filter" settings to a DeploymentCreateRequest.
func trafficFilterToModel(ctx context.Context, set types.Set, req *models.DeploymentCreateRequest) diag.Diagnostics {
	if len(set.Elements()) == 0 || req == nil {
		return nil
	}

	if req.Settings == nil {
		req.Settings = &models.DeploymentCreateSettings{}
	}

	if req.Settings.TrafficFilterSettings == nil {
		req.Settings.TrafficFilterSettings = &models.TrafficFilterSettings{}
	}

	var rulesets []string
	if diags := tfsdk.ValueAs(ctx, set, &rulesets); diags.HasError() {
		return diags
	}

	req.Settings.TrafficFilterSettings.Rulesets = append(
		req.Settings.TrafficFilterSettings.Rulesets,
		rulesets...,
	)

	return nil
}

func HandleRemoteClusters(ctx context.Context, client *api.API, deploymentId string, esObj types.Object) diag.Diagnostics {
	remoteClusters, refId, diags := elasticsearchRemoteClustersPayload(ctx, client, deploymentId, esObj)

	if diags.HasError() {
		return diags
	}

	if err := esremoteclustersapi.Update(esremoteclustersapi.UpdateParams{
		API:             client,
		DeploymentID:    deploymentId,
		RefID:           refId,
		RemoteResources: remoteClusters,
	}); err != nil {
		diags.AddError("cannot update remote clusters", err.Error())
		return diags
	}

	return nil
}

func elasticsearchRemoteClustersPayload(ctx context.Context, client *api.API, deploymentId string, esObj types.Object) (*models.RemoteResources, string, diag.Diagnostics) {
	var es *elasticsearchv2.ElasticsearchTF

	diags := tfsdk.ValueAs(ctx, esObj, &es)

	if diags.HasError() {
		return nil, "", diags
	}

	if es == nil {
		var diags diag.Diagnostics
		diags.AddError("failed create remote clusters payload", "there is no elasticsearch")
		return nil, "", diags
	}

	remoteRes, diags := elasticsearchv2.ElasticsearchRemoteClustersPayload(ctx, es.RemoteCluster)
	if diags.HasError() {
		return nil, "", diags
	}

	return remoteRes, es.RefId.ValueString(), nil
}
