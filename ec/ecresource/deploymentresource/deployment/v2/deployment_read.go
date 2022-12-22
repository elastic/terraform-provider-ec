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

	"github.com/elastic/cloud-sdk-go/pkg/models"

	apmv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/apm/v2"
	elasticsearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
	enterprisesearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/enterprisesearch/v2"
	integrationsserverv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/integrationsserver/v2"
	kibanav2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/kibana/v2"
	observabilityv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/observability/v2"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/utils"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
