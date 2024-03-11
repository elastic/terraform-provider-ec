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
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/elastic/cloud-sdk-go/pkg/client/deployments"

	"github.com/blang/semver"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"

	apmv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/apm/v2"
	elasticsearchv1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v1"
	elasticsearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
	enterprisesearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/enterprisesearch/v2"
	integrationsserverv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/integrationsserver/v2"
	kibanav2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/kibana/v2"
	v1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/observability/v1"
	observabilityv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/observability/v2"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/utils"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type Deployment struct {
	Id                         string                                   `tfsdk:"id"`
	Alias                      string                                   `tfsdk:"alias"`
	Version                    string                                   `tfsdk:"version"`
	Region                     string                                   `tfsdk:"region"`
	DeploymentTemplateId       string                                   `tfsdk:"deployment_template_id"`
	Name                       string                                   `tfsdk:"name"`
	RequestId                  string                                   `tfsdk:"request_id"`
	ElasticsearchUsername      string                                   `tfsdk:"elasticsearch_username"`
	ElasticsearchPassword      string                                   `tfsdk:"elasticsearch_password"`
	ApmSecretToken             *string                                  `tfsdk:"apm_secret_token"`
	TrafficFilter              []string                                 `tfsdk:"traffic_filter"`
	Tags                       map[string]string                        `tfsdk:"tags"`
	Elasticsearch              *elasticsearchv2.Elasticsearch           `tfsdk:"elasticsearch"`
	Kibana                     *kibanav2.Kibana                         `tfsdk:"kibana"`
	Apm                        *apmv2.Apm                               `tfsdk:"apm"`
	IntegrationsServer         *integrationsserverv2.IntegrationsServer `tfsdk:"integrations_server"`
	EnterpriseSearch           *enterprisesearchv2.EnterpriseSearch     `tfsdk:"enterprise_search"`
	Observability              *observabilityv2.Observability           `tfsdk:"observability"`
	ResetElasticsearchPassword *bool                                    `tfsdk:"reset_elasticsearch_password"`
	MigrateToLatestHardware    *bool                                    `tfsdk:"migrate_to_latest_hardware"`
}

func (dep *Deployment) PersistSnapshotSource(ctx context.Context, esPlan *elasticsearchv2.ElasticsearchTF) diag.Diagnostics {
	if dep == nil || dep.Elasticsearch == nil {
		return nil
	}

	if esPlan == nil || esPlan.SnapshotSource.IsNull() || esPlan.SnapshotSource.IsUnknown() {
		return nil
	}

	var snapshotSource *elasticsearchv1.ElasticsearchSnapshotSourceTF
	if diags := tfsdk.ValueAs(ctx, esPlan.SnapshotSource, &snapshotSource); diags.HasError() {
		return diags
	}

	dep.Elasticsearch.SnapshotSource = &elasticsearchv2.ElasticsearchSnapshotSource{
		SourceElasticsearchClusterId: snapshotSource.SourceElasticsearchClusterId.ValueString(),
		SnapshotName:                 snapshotSource.SnapshotName.ValueString(),
	}

	return nil
}

// Nullify Elasticsearch topologies that have zero size and are not specified in plan
func (dep *Deployment) NullifyUnusedEsTopologies(ctx context.Context, esPlan *elasticsearchv2.ElasticsearchTF) {
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

// SetLatestInstanceConfigInfo Sets latest instance_configuration_id and instance_configuration_version for each
// topology element, based on the migrate template request
func (dep *Deployment) SetLatestInstanceConfigInfo(migrateUpdateRequest *deployments.MigrateDeploymentTemplateOK) {
	if migrateUpdateRequest == nil {
		return
	}

	if dep.Elasticsearch != nil {
		elasticsearchv2.SetLatestInstanceConfigInfo(dep.Elasticsearch.HotTier, elasticsearchv2.GetTopologyFromMigrateRequest(migrateUpdateRequest, "hot"))
		elasticsearchv2.SetLatestInstanceConfigInfo(dep.Elasticsearch.WarmTier, elasticsearchv2.GetTopologyFromMigrateRequest(migrateUpdateRequest, "warm"))
		elasticsearchv2.SetLatestInstanceConfigInfo(dep.Elasticsearch.ColdTier, elasticsearchv2.GetTopologyFromMigrateRequest(migrateUpdateRequest, "cold"))
		elasticsearchv2.SetLatestInstanceConfigInfo(dep.Elasticsearch.FrozenTier, elasticsearchv2.GetTopologyFromMigrateRequest(migrateUpdateRequest, "frozen"))
		elasticsearchv2.SetLatestInstanceConfigInfo(dep.Elasticsearch.MlTier, elasticsearchv2.GetTopologyFromMigrateRequest(migrateUpdateRequest, "ml"))
		elasticsearchv2.SetLatestInstanceConfigInfo(dep.Elasticsearch.MasterTier, elasticsearchv2.GetTopologyFromMigrateRequest(migrateUpdateRequest, "master"))
		elasticsearchv2.SetLatestInstanceConfigInfo(dep.Elasticsearch.CoordinatingTier, elasticsearchv2.GetTopologyFromMigrateRequest(migrateUpdateRequest, "coordinating"))
	}

	if migrateUpdateRequest.Payload.Resources.Apm != nil && len(migrateUpdateRequest.Payload.Resources.Apm) > 0 && len(migrateUpdateRequest.Payload.Resources.Apm[0].Plan.ClusterTopology) > 0 {
		apmv2.SetLatestInstanceConfigInfo(dep.Apm, migrateUpdateRequest.Payload.Resources.Apm[0].Plan.ClusterTopology[0])
	}

	if migrateUpdateRequest.Payload.Resources.EnterpriseSearch != nil && len(migrateUpdateRequest.Payload.Resources.EnterpriseSearch) > 0 && len(migrateUpdateRequest.Payload.Resources.EnterpriseSearch[0].Plan.ClusterTopology) > 0 {
		enterprisesearchv2.SetLatestInstanceConfigInfo(dep.EnterpriseSearch, migrateUpdateRequest.Payload.Resources.EnterpriseSearch[0].Plan.ClusterTopology[0])
	}

	if migrateUpdateRequest.Payload.Resources.IntegrationsServer != nil && len(migrateUpdateRequest.Payload.Resources.IntegrationsServer) > 0 && len(migrateUpdateRequest.Payload.Resources.IntegrationsServer[0].Plan.ClusterTopology) > 0 {
		integrationsserverv2.SetLatestInstanceConfigInfo(dep.IntegrationsServer, migrateUpdateRequest.Payload.Resources.IntegrationsServer[0].Plan.ClusterTopology[0])
	}

	if migrateUpdateRequest.Payload.Resources.Kibana != nil && len(migrateUpdateRequest.Payload.Resources.Kibana) > 0 && len(migrateUpdateRequest.Payload.Resources.Kibana[0].Plan.ClusterTopology) > 0 {
		kibanav2.SetLatestInstanceConfigInfo(dep.Kibana, migrateUpdateRequest.Payload.Resources.Kibana[0].Plan.ClusterTopology[0])
	}
}

// SetLatestInstanceConfigInfoToCurrent Sets latest instance_configuration_id and instance_configuration_version for each
// topology element, based on the current values
func (dep *Deployment) SetLatestInstanceConfigInfoToCurrent() {
	if dep.Elasticsearch != nil {
		elasticsearchv2.SetLatestInstanceConfigInfoToCurrent(dep.Elasticsearch.HotTier)
		elasticsearchv2.SetLatestInstanceConfigInfoToCurrent(dep.Elasticsearch.WarmTier)
		elasticsearchv2.SetLatestInstanceConfigInfoToCurrent(dep.Elasticsearch.ColdTier)
		elasticsearchv2.SetLatestInstanceConfigInfoToCurrent(dep.Elasticsearch.FrozenTier)
		elasticsearchv2.SetLatestInstanceConfigInfoToCurrent(dep.Elasticsearch.MlTier)
		elasticsearchv2.SetLatestInstanceConfigInfoToCurrent(dep.Elasticsearch.MasterTier)
		elasticsearchv2.SetLatestInstanceConfigInfoToCurrent(dep.Elasticsearch.CoordinatingTier)
	}

	apmv2.SetLatestInstanceConfigInfoToCurrent(dep.Apm)
	enterprisesearchv2.SetLatestInstanceConfigInfoToCurrent(dep.EnterpriseSearch)
	integrationsserverv2.SetLatestInstanceConfigInfoToCurrent(dep.IntegrationsServer)
	kibanav2.SetLatestInstanceConfigInfoToCurrent(dep.Kibana)
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
		dep.Tags = converters.ModelsTagsToMap(res.Metadata.Tags)
	}

	if res.Resources == nil {
		return nil, nil
	}

	templateID, err := getDeploymentTemplateID(res.Resources)
	if err != nil {
		return nil, err
	}

	dep.DeploymentTemplateId = templateID

	dep.Region = getRegion(res.Resources)

	// We're reconciling the version and storing the lowest version of any
	// of the deployment resources. This ensures that if an upgrade fails,
	// the state version will be lower than the desired version, making
	// retries possible. Once more resource types are added, the function
	// needs to be modified to check those as well.
	version, err := getLowestVersion(res.Resources)
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

	if dep.TrafficFilter, err = readTrafficFilters(res.Settings); err != nil {
		return nil, err
	}

	if dep.Observability, err = observabilityv2.ReadObservability(res.Settings); err != nil {
		return nil, err
	}

	dep.parseCredentials(deploymentResources)

	return &dep, nil
}

func readTrafficFilters(in *models.DeploymentSettings) ([]string, error) {
	if in == nil || in.TrafficFilterSettings == nil || len(in.TrafficFilterSettings.Rulesets) == 0 {
		return nil, nil
	}

	var rules []string

	return append(rules, in.TrafficFilterSettings.Rulesets...), nil
}

// parseCredentials parses the Create or Update response Resources populating
// credential settings in the Terraform state if the keys are found, currently
// populates the following credentials in plain text:
// * Elasticsearch username and Password
func (dep *Deployment) parseCredentials(resources []*models.DeploymentResource) {
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
}

func (dep *Deployment) ProcessSelfInObservability(ctx context.Context, base DeploymentTF) diag.Diagnostics {
	if dep == nil || dep.Observability == nil {
		return nil
	}

	if dep.Observability.DeploymentId == nil {
		return nil
	}

	var baseObservability v1.ObservabilityTF
	diags := base.Observability.As(ctx, &baseObservability, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return diags
	}

	deploymentIDIsKnown := !(baseObservability.DeploymentId.IsNull() || baseObservability.DeploymentId.IsUnknown())
	if deploymentIDIsKnown && baseObservability.DeploymentId.ValueString() != "self" {
		return nil
	}

	if *dep.Observability.DeploymentId == dep.Id {
		*dep.Observability.DeploymentId = "self"
	}

	return nil
}

func (dep *Deployment) IncludePrivateStateTrafficFilters(ctx context.Context, base DeploymentTF, privateFilters []string) diag.Diagnostics {
	var baseFilters []string
	diags := base.TrafficFilter.ElementsAs(ctx, &baseFilters, true)
	if diags.HasError() {
		return diags
	}

	for _, filter := range privateFilters {
		if !slices.Contains(baseFilters, filter) {
			baseFilters = append(baseFilters, filter)
		}
	}

	if len(baseFilters) == 0 {
		dep.TrafficFilter = baseFilters
	}

	intersectionFilters := []string{}
	for _, filter := range dep.TrafficFilter {
		if slices.Contains(baseFilters, filter) {
			intersectionFilters = append(intersectionFilters, filter)
		}
	}

	dep.TrafficFilter = intersectionFilters

	return diags
}

func (dep *Deployment) SetCredentialsIfEmpty(state *DeploymentTF) {
	if state == nil {
		return
	}

	if dep.ElasticsearchPassword == "" && state.ElasticsearchPassword.ValueString() != "" {
		dep.ElasticsearchPassword = state.ElasticsearchPassword.ValueString()
	}

	if dep.ElasticsearchUsername == "" && state.ElasticsearchUsername.ValueString() != "" {
		dep.ElasticsearchUsername = state.ElasticsearchUsername.ValueString()
	}

	if (dep.ApmSecretToken == nil || *dep.ApmSecretToken == "") && state.ApmSecretToken.ValueString() != "" {
		dep.ApmSecretToken = ec.String(state.ApmSecretToken.ValueString())
	}
}

func (dep *Deployment) HasNodeTypes() bool {
	if dep.Elasticsearch != nil {
		for _, t := range dep.Elasticsearch.GetTopologies() {
			if t.HasNodeTypes() {
				return true
			}
		}
	}
	return false
}

func getLowestVersion(res *models.DeploymentResources) (string, error) {
	// We're starting off with a very high version so it can be replaced.
	replaceVersion := `99.99.99`
	version := semver.MustParse(replaceVersion)
	for _, r := range res.Elasticsearch {
		if !util.IsCurrentEsPlanEmpty(r) {
			v := r.Info.PlanInfo.Current.Plan.Elasticsearch.Version
			if err := swapLowerVersion(&version, v); err != nil && !elasticsearchv2.IsElasticsearchStopped(r) {
				return "", fmt.Errorf("elasticsearch version '%s' is not semver compliant: %w", v, err)
			}
		}
	}

	for _, r := range res.Kibana {
		if !util.IsCurrentKibanaPlanEmpty(r) && !kibanav2.IsKibanaStopped(r) {
			v := r.Info.PlanInfo.Current.Plan.Kibana.Version
			if err := swapLowerVersion(&version, v); err != nil {
				return version.String(), fmt.Errorf("kibana version '%s' is not semver compliant: %w", v, err)
			}
		}
	}

	for _, r := range res.Apm {
		if !util.IsCurrentApmPlanEmpty(r) && !apmv2.IsApmStopped(r) {
			v := r.Info.PlanInfo.Current.Plan.Apm.Version
			if err := swapLowerVersion(&version, v); err != nil {
				return version.String(), fmt.Errorf("apm version '%s' is not semver compliant: %w", v, err)
			}
		}
	}

	for _, r := range res.IntegrationsServer {
		if !util.IsCurrentIntegrationsServerPlanEmpty(r) && !integrationsserverv2.IsIntegrationsServerStopped(r) {
			v := r.Info.PlanInfo.Current.Plan.IntegrationsServer.Version
			if err := swapLowerVersion(&version, v); err != nil {
				return version.String(), fmt.Errorf("integrations_server version '%s' is not semver compliant: %w", v, err)
			}
		}
	}

	for _, r := range res.EnterpriseSearch {
		if !util.IsCurrentEssPlanEmpty(r) && !enterprisesearchv2.IsEnterpriseSearchStopped(r) {
			v := r.Info.PlanInfo.Current.Plan.EnterpriseSearch.Version
			if err := swapLowerVersion(&version, v); err != nil {
				return version.String(), fmt.Errorf("enterprise search version '%s' is not semver compliant: %w", v, err)
			}
		}
	}

	if version.String() != replaceVersion {
		return version.String(), nil
	}
	return "", errors.New("unable to determine the lowest version for any the deployment components")
}

func swapLowerVersion(version *semver.Version, comp string) error {
	if comp == "" {
		return nil
	}

	v, err := semver.Parse(comp)
	if err != nil {
		return err
	}
	if v.LT(*version) {
		*version = v
	}
	return nil
}

func getRegion(res *models.DeploymentResources) string {
	for _, r := range res.Elasticsearch {
		if r.Region != nil && *r.Region != "" {
			return *r.Region
		}
	}

	return ""
}

func getDeploymentTemplateID(res *models.DeploymentResources) (string, error) {
	var deploymentTemplateID string
	var foundTemplates []string
	for _, esRes := range res.Elasticsearch {
		if util.IsCurrentEsPlanEmpty(esRes) {
			continue
		}

		var emptyDT = esRes.Info.PlanInfo.Current.Plan.DeploymentTemplate == nil
		if emptyDT {
			continue
		}

		if deploymentTemplateID == "" {
			deploymentTemplateID = *esRes.Info.PlanInfo.Current.Plan.DeploymentTemplate.ID
		}

		foundTemplates = append(foundTemplates,
			*esRes.Info.PlanInfo.Current.Plan.DeploymentTemplate.ID,
		)
	}

	if deploymentTemplateID == "" {
		return "", errors.New("failed to obtain the deployment template id")
	}

	if len(foundTemplates) > 1 {
		return "", fmt.Errorf(
			"there are more than 1 deployment templates specified on the deployment: \"%s\"", strings.Join(foundTemplates, ", "),
		)
	}

	return deploymentTemplateID, nil
}
