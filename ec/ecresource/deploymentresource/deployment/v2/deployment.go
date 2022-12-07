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
	"fmt"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deptemplateapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/esremoteclustersapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"

	apmv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/apm/v2"
	elasticsearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
	enterprisesearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/enterprisesearch/v2"
	integrationsserverv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/integrationsserver/v2"
	kibanav2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/kibana/v2"
	observabilityv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/observability/v2"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/utils"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DeploymentTF struct {
	Id                    types.String `tfsdk:"id"`
	Alias                 types.String `tfsdk:"alias"`
	Version               types.String `tfsdk:"version"`
	Region                types.String `tfsdk:"region"`
	DeploymentTemplateId  types.String `tfsdk:"deployment_template_id"`
	Name                  types.String `tfsdk:"name"`
	RequestId             types.String `tfsdk:"request_id"`
	ElasticsearchUsername types.String `tfsdk:"elasticsearch_username"`
	ElasticsearchPassword types.String `tfsdk:"elasticsearch_password"`
	ApmSecretToken        types.String `tfsdk:"apm_secret_token"`
	TrafficFilter         types.Set    `tfsdk:"traffic_filter"`
	Tags                  types.Map    `tfsdk:"tags"`
	Elasticsearch         types.Object `tfsdk:"elasticsearch"`
	Kibana                types.Object `tfsdk:"kibana"`
	Apm                   types.Object `tfsdk:"apm"`
	IntegrationsServer    types.Object `tfsdk:"integrations_server"`
	EnterpriseSearch      types.Object `tfsdk:"enterprise_search"`
	Observability         types.Object `tfsdk:"observability"`
}

type Deployment struct {
	Id                    string                                   `tfsdk:"id"`
	Alias                 string                                   `tfsdk:"alias"`
	Version               string                                   `tfsdk:"version"`
	Region                string                                   `tfsdk:"region"`
	DeploymentTemplateId  string                                   `tfsdk:"deployment_template_id"`
	Name                  string                                   `tfsdk:"name"`
	RequestId             string                                   `tfsdk:"request_id"`
	ElasticsearchUsername string                                   `tfsdk:"elasticsearch_username"`
	ElasticsearchPassword string                                   `tfsdk:"elasticsearch_password"`
	ApmSecretToken        *string                                  `tfsdk:"apm_secret_token"`
	TrafficFilter         []string                                 `tfsdk:"traffic_filter"`
	Tags                  map[string]string                        `tfsdk:"tags"`
	Elasticsearch         *elasticsearchv2.Elasticsearch           `tfsdk:"elasticsearch"`
	Kibana                *kibanav2.Kibana                         `tfsdk:"kibana"`
	Apm                   *apmv2.Apm                               `tfsdk:"apm"`
	IntegrationsServer    *integrationsserverv2.IntegrationsServer `tfsdk:"integrations_server"`
	EnterpriseSearch      *enterprisesearchv2.EnterpriseSearch     `tfsdk:"enterprise_search"`
	Observability         *observabilityv2.Observability           `tfsdk:"observability"`
}

// Nullify Elasticsearch topologies that have zero size and are not specified in plan
func (dep *Deployment) NullifyNotUsedEsTopologies(ctx context.Context, esPlan *elasticsearchv2.ElasticsearchTF) {
	if dep.Elasticsearch == nil {
		return
	}

	if esPlan == nil {
		return
	}

	dep.Elasticsearch.HotTier = nullifyUnspecifiedZeroSizedTier(esPlan.HotContentTier, dep.Elasticsearch.HotTier)

	dep.Elasticsearch.WarmTier = nullifyUnspecifiedZeroSizedTier(esPlan.WarmTier, dep.Elasticsearch.WarmTier)

	dep.Elasticsearch.ColdTier = nullifyUnspecifiedZeroSizedTier(esPlan.ColdTier, dep.Elasticsearch.ColdTier)

	dep.Elasticsearch.FrozenTier = nullifyUnspecifiedZeroSizedTier(esPlan.FrozenTier, dep.Elasticsearch.FrozenTier)

	dep.Elasticsearch.MlTier = nullifyUnspecifiedZeroSizedTier(esPlan.MlTier, dep.Elasticsearch.MlTier)

	dep.Elasticsearch.MasterTier = nullifyUnspecifiedZeroSizedTier(esPlan.MasterTier, dep.Elasticsearch.MasterTier)

	dep.Elasticsearch.CoordinatingTier = nullifyUnspecifiedZeroSizedTier(esPlan.CoordinatingTier, dep.Elasticsearch.CoordinatingTier)
}

func nullifyUnspecifiedZeroSizedTier(tierPlan types.Object, tier *elasticsearchv2.ElasticsearchTopology) *elasticsearchv2.ElasticsearchTopology {

	if tierPlan.IsNull() && tier != nil {

		size, err := converters.ParseTopologySize(tier.Size, tier.SizeResource)

		// we can ignore returning an error here - it's handled in readers
		if err == nil && size != nil && size.Value != nil && *size.Value == 0 {
			tier = nil
		}
	}

	return tier
}

func ReadDeployment(res *models.DeploymentGetResponse, remotes *models.RemoteResources, deploymentResources []*models.DeploymentResource) (*Deployment, error) {
	var dep Deployment

	if res.ID == nil {
		return nil, utils.MissingField("ID")
	}
	dep.Id = *res.ID

	dep.Alias = res.Alias

	if res.Name == nil {
		return nil, utils.MissingField("Name")
	}
	dep.Name = *res.Name

	if res.Metadata != nil {
		dep.Tags = converters.TagsToMap(res.Metadata.Tags)
	}

	if res.Resources == nil {
		return nil, nil
	}

	templateID, err := utils.GetDeploymentTemplateID(res.Resources)
	if err != nil {
		return nil, err
	}

	dep.DeploymentTemplateId = templateID

	dep.Region = utils.GetRegion(res.Resources)

	// We're reconciling the version and storing the lowest version of any
	// of the deployment resources. This ensures that if an upgrade fails,
	// the state version will be lower than the desired version, making
	// retries possible. Once more resource types are added, the function
	// needs to be modified to check those as well.
	version, err := utils.GetLowestVersion(res.Resources)
	if err != nil {
		// This code path is highly unlikely, but we're bubbling up the
		// error in case one of the versions isn't parseable by semver.
		return nil, fmt.Errorf("failed reading deployment: %w", err)
	}
	dep.Version = version

	dep.Elasticsearch, err = elasticsearchv2.ReadElasticsearches(res.Resources.Elasticsearch, remotes)
	if err != nil {
		return nil, err
	}

	if dep.Kibana, err = kibanav2.ReadKibanas(res.Resources.Kibana); err != nil {
		return nil, err
	}

	if dep.Apm, err = apmv2.ReadApms(res.Resources.Apm); err != nil {
		return nil, err
	}

	if dep.IntegrationsServer, err = integrationsserverv2.ReadIntegrationsServers(res.Resources.IntegrationsServer); err != nil {
		return nil, err
	}

	if dep.EnterpriseSearch, err = enterprisesearchv2.ReadEnterpriseSearches(res.Resources.EnterpriseSearch); err != nil {
		return nil, err
	}

	if dep.TrafficFilter, err = ReadTrafficFilters(res.Settings); err != nil {
		return nil, err
	}

	if dep.Observability, err = observabilityv2.ReadObservability(res.Settings); err != nil {
		return nil, err
	}

	if err := dep.parseCredentials(deploymentResources); err != nil {
		return nil, err
	}

	return &dep, nil
}

func (dep DeploymentTF) CreateRequest(ctx context.Context, client *api.API) (*models.DeploymentCreateRequest, diag.Diagnostics) {
	var result = models.DeploymentCreateRequest{
		Name:      dep.Name.Value,
		Alias:     dep.Alias.Value,
		Resources: &models.DeploymentCreateResources{},
		Settings:  &models.DeploymentCreateSettings{},
		Metadata:  &models.DeploymentCreateMetadata{},
	}

	dtID := dep.DeploymentTemplateId.Value
	version := dep.Version.Value

	var diagsnostics diag.Diagnostics

	template, err := deptemplateapi.Get(deptemplateapi.GetParams{
		API:                        client,
		TemplateID:                 dtID,
		Region:                     dep.Region.Value,
		HideInstanceConfigurations: true,
	})
	if err != nil {
		diagsnostics.AddError("Deployment template get error", err.Error())
		return nil, diagsnostics
	}

	useNodeRoles, err := utils.CompatibleWithNodeRoles(version)
	if err != nil {
		diagsnostics.AddError("Deployment parse error", err.Error())
		return nil, diagsnostics
	}

	elasticsearchPayload, diags := elasticsearchv2.ElasticsearchPayload(ctx, dep.Elasticsearch, template, dtID, version, useNodeRoles, false)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if elasticsearchPayload != nil {
		result.Resources.Elasticsearch = []*models.ElasticsearchPayload{elasticsearchPayload}
	}

	kibanaPayload, diags := kibanav2.KibanaPayload(ctx, dep.Kibana, template)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if kibanaPayload != nil {
		result.Resources.Kibana = []*models.KibanaPayload{kibanaPayload}
	}

	apmPayload, diags := apmv2.ApmPayload(ctx, dep.Apm, template)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if apmPayload != nil {
		result.Resources.Apm = []*models.ApmPayload{apmPayload}
	}

	integrationsServerPayload, diags := integrationsserverv2.IntegrationsServerPayload(ctx, dep.IntegrationsServer, template)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if integrationsServerPayload != nil {
		result.Resources.IntegrationsServer = []*models.IntegrationsServerPayload{integrationsServerPayload}
	}

	enterpriseSearchPayload, diags := enterprisesearchv2.EnterpriseSearchesPayload(ctx, dep.EnterpriseSearch, template)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if enterpriseSearchPayload != nil {
		result.Resources.EnterpriseSearch = []*models.EnterpriseSearchPayload{enterpriseSearchPayload}
	}

	if diags := TrafficFilterToModel(ctx, dep.TrafficFilter, &result); diags.HasError() {
		diagsnostics.Append(diags...)
	}

	observabilityPayload, diags := observabilityv2.ObservabilityPayload(ctx, dep.Observability, client)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	result.Settings.Observability = observabilityPayload

	result.Metadata.Tags, diags = converters.TFmapToTags(ctx, dep.Tags)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	return &result, diagsnostics
}

func ReadTrafficFilters(in *models.DeploymentSettings) ([]string, error) {
	if in == nil || in.TrafficFilterSettings == nil || len(in.TrafficFilterSettings.Rulesets) == 0 {
		return nil, nil
	}

	var rules []string

	return append(rules, in.TrafficFilterSettings.Rulesets...), nil
}

// TrafficFilterToModel expands the flattened "traffic_filter" settings to a DeploymentCreateRequest.
func TrafficFilterToModel(ctx context.Context, set types.Set, req *models.DeploymentCreateRequest) diag.Diagnostics {
	if len(set.Elems) == 0 || req == nil {
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

// parseCredentials parses the Create or Update response Resources populating
// credential settings in the Terraform state if the keys are found, currently
// populates the following credentials in plain text:
// * Elasticsearch username and Password
func (dep *Deployment) parseCredentials(resources []*models.DeploymentResource) error {
	for _, res := range resources {

		if creds := res.Credentials; creds != nil {
			if creds.Username != nil && *creds.Username != "" {
				dep.ElasticsearchUsername = *creds.Username
			}

			if creds.Password != nil && *creds.Password != "" {
				dep.ElasticsearchPassword = *creds.Password
			}
		}

		if res.SecretToken != "" {
			dep.ApmSecretToken = &res.SecretToken
		}
	}

	return nil
}

func (dep *Deployment) ProcessSelfInObservability() {

	if dep.Observability == nil {
		return
	}

	if dep.Observability.DeploymentId == nil {
		return
	}

	if *dep.Observability.DeploymentId == dep.Id {
		*dep.Observability.DeploymentId = "self"
	}
}

func (dep *Deployment) SetCredentialsIfEmpty(state *DeploymentTF) {
	if state == nil {
		return
	}

	if dep.ElasticsearchPassword == "" && state.ElasticsearchPassword.Value != "" {
		dep.ElasticsearchPassword = state.ElasticsearchPassword.Value
	}

	if dep.ElasticsearchUsername == "" && state.ElasticsearchUsername.Value != "" {
		dep.ElasticsearchUsername = state.ElasticsearchUsername.Value
	}

	if (dep.ApmSecretToken == nil || *dep.ApmSecretToken == "") && state.ApmSecretToken.Value != "" {
		dep.ApmSecretToken = &state.ApmSecretToken.Value
	}
}

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

	useNodeRoles, diags := utils.UseNodeRoles(state.Version, plan.Version)

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

	result.Metadata.Tags, diags = converters.TFmapToTags(ctx, plan.Tags)
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

// func HandleRemoteClusters(ctx context.Context, client *api.API, newState, oldState DeploymentTF) diag.Diagnostics {
func HandleRemoteClusters(ctx context.Context, client *api.API, deploymentId string, esObj types.Object) diag.Diagnostics {
	remoteClusters, refId, diags := ElasticsearchRemoteClustersPayload(ctx, client, deploymentId, esObj)

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

func ElasticsearchRemoteClustersPayload(ctx context.Context, client *api.API, deploymentId string, esObj types.Object) (*models.RemoteResources, string, diag.Diagnostics) {
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

	return remoteRes, es.RefId.Value, nil
}
