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

package deploymentresource

import (
	"context"
	"errors"
	"fmt"

	"github.com/blang/semver"
	"github.com/elastic/cloud-sdk-go/pkg/api/apierror"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deputil"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/eskeystoreapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/esremoteclustersapi"
	"github.com/elastic/cloud-sdk-go/pkg/client/deployments"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	apmv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/apm/v2"
	deploymentv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/deployment/v2"
	elasticsearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
	enterprisesearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/enterprisesearch/v2"
	integrationsserverv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/integrationsserver/v2"
	kibanav2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/kibana/v2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	if !r.ready(&response.Diagnostics) {
		return
	}

	var curState deploymentv2.DeploymentTF

	diags := request.State.Get(ctx, &curState)

	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	var newState *deploymentv2.Deployment

	privateFilters, d := readPrivateStateTrafficFilters(ctx, request.Private)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	// use state for the plan (there is no plan and config during Read) - otherwise we can get unempty plan output
	newState, diags = r.read(ctx, curState.Id.ValueString(), &curState, nil, nil, privateFilters, response)

	response.Diagnostics.Append(diags...)

	if newState == nil {
		response.State.RemoveResource(ctx)
	}

	if newState != nil {
		diags = response.State.Set(ctx, newState)
	}

	response.Diagnostics.Append(diags...)
}

// at least one of state and plan should not be nil
func (r *Resource) read(ctx context.Context, id string, state *deploymentv2.DeploymentTF, plan *deploymentv2.DeploymentTF, deploymentResources []*models.DeploymentResource, privateFilters []string, readResponse *resource.ReadResponse) (*deploymentv2.Deployment, diag.Diagnostics) {
	var diags diag.Diagnostics

	var base deploymentv2.DeploymentTF

	switch {
	case plan != nil:
		base = *plan
	case state != nil:
		base = *state
	default:
		diags.AddError("both state and plan are empty", "please specify at least one of them")
		return nil, diags
	}

	response, err := deploymentapi.Get(deploymentapi.GetParams{
		API:          r.client,
		DeploymentID: id,
		QueryParams: deputil.QueryParams{
			ShowSettings:     true,
			ShowPlans:        true,
			ShowMetadata:     true,
			ShowPlanDefaults: true,
		},
	})
	if err != nil {
		if deploymentNotFound(err) {
			diags.AddError("Deployment not found", err.Error())
			return nil, diags
		}
		diags.AddError("Deployment get error", err.Error())
		return nil, diags
	}

	if !HasRunningResources(response) {
		return nil, nil
	}

	if response.Resources == nil || len(response.Resources.Elasticsearch) == 0 {
		diags.AddError("Get resource error", "cannot find Elasticsearch in response resources")
		return nil, diags
	}

	if response.Resources.Elasticsearch[0].Info.PlanInfo.Current != nil && response.Resources.Elasticsearch[0].Info.PlanInfo.Current.Plan != nil {
		if err := checkVersion(response.Resources.Elasticsearch[0].Info.PlanInfo.Current.Plan.Elasticsearch.Version); err != nil {
			diags.AddError("Get resource error", err.Error())
			return nil, diags
		}
	}

	refId := ""

	var baseElasticsearch *elasticsearchv2.ElasticsearchTF

	if diags = tfsdk.ValueAs(ctx, base.Elasticsearch, &baseElasticsearch); diags.HasError() {
		return nil, diags
	}

	if baseElasticsearch != nil {
		refId = baseElasticsearch.RefId.ValueString()
	}

	remotes, err := esremoteclustersapi.Get(esremoteclustersapi.GetParams{
		API: r.client, DeploymentID: id,
		RefID: refId,
	})
	if err != nil {
		diags.AddError("Remote clusters read error", err.Error())
		return nil, diags
	}
	if remotes == nil {
		remotes = &models.RemoteResources{}
	}

	deployment, err := deploymentv2.ReadDeployment(response, remotes, deploymentResources)
	if err != nil {
		diags.AddError("Deployment read error", err.Error())
		return nil, diags
	}

	deployment.RequestId = base.RequestId.ValueString()
	if !base.ResetElasticsearchPassword.IsNull() && !base.ResetElasticsearchPassword.IsUnknown() {
		deployment.ResetElasticsearchPassword = base.ResetElasticsearchPassword.ValueBoolPointer()
	}

	if !base.MigrateToLatestHardware.IsNull() && !base.MigrateToLatestHardware.IsUnknown() {
		deployment.MigrateToLatestHardware = base.MigrateToLatestHardware.ValueBoolPointer()
	}

	diags.Append(deployment.IncludePrivateStateTrafficFilters(ctx, base, privateFilters)...)

	deployment.SetCredentialsIfEmpty(state)

	deployment.ProcessSelfInObservability()

	deployment.NullifyUnusedEsTopologies(ctx, baseElasticsearch)
	diags.Append(deployment.PersistSnapshotSource(ctx, baseElasticsearch)...)

	if !deployment.HasNodeTypes() {
		// The MigrateDeploymentTemplate request can only be performed for deployments that use node roles.
		// We'll skip this logic for deployments with node types.
		migrateTemplateRequest, err := r.client.V1API.Deployments.MigrateDeploymentTemplate(
			deployments.NewMigrateDeploymentTemplateParams().WithDeploymentID(deployment.Id).WithTemplateID(deployment.DeploymentTemplateId),
			r.client.AuthWriter,
		)

		if err != nil {
			diags.AddError("Template migrate request error", err.Error())
			return nil, diags
		}

		// Store migrate request in private state
		if readResponse != nil {
			UpdatePrivateStateMigrateTemplateRequest(ctx, readResponse.Private, migrateTemplateRequest)
		}

		deployment.SetLatestInstanceConfigInfo(migrateTemplateRequest)
	} else {
		// Set latest_instance_configuration_* fields to current values
		// If this isn't done, when migrating a deployment to node roles, these fields will contain inconsistent values
		deployment.SetLatestInstanceConfigInfoToCurrent()
	}

	// Set Elasticsearch `strategy` to the one from plan.
	// We don't care about backend current `strategy`'s value and should not trigger a change,
	// if the backend's value differs from the local state.
	if baseElasticsearch != nil && !baseElasticsearch.Strategy.IsNull() {
		deployment.Elasticsearch.Strategy = baseElasticsearch.Strategy.ValueStringPointer()
	}

	// sync Elasticsearch keystore contents if plan or state defines it:
	// - all keystore entries that are not managed by the resource are left alone
	// - if backend doesn't contain some keystore entry, the entry should be removed from the future state as well
	if baseElasticsearch != nil && deployment.Elasticsearch != nil && !baseElasticsearch.KeystoreContents.IsNull() {
		ds := baseElasticsearch.KeystoreContents.ElementsAs(ctx, &deployment.Elasticsearch.KeystoreContents, true)
		diags.Append(ds...)

		keystoreContents, err := eskeystoreapi.Get(eskeystoreapi.GetParams{
			API:          r.client,
			DeploymentID: id,
		})
		if err != nil {
			diags.AddError("Deployment keystore read error", err.Error())
			return nil, diags
		}

		for entryName, entryVal := range deployment.Elasticsearch.KeystoreContents {
			secret, ok := keystoreContents.Secrets[entryName]
			if !ok {
				delete(deployment.Elasticsearch.KeystoreContents, entryName)
				continue
			}
			if secret.AsFile != nil {
				entryVal.AsFile = secret.AsFile
				deployment.Elasticsearch.KeystoreContents[entryName] = entryVal
			}
		}
	}

	// ReadDeployment returns empty config struct if there is no config, so we have to nullify it if plan doesn't contain it
	// we use state for plan in Read and there is no state during import so we need to check elasticsearchPlan against nil
	if baseElasticsearch != nil &&
		baseElasticsearch.Config.IsNull() &&
		deployment.Elasticsearch != nil &&
		deployment.Elasticsearch.Config != nil &&
		deployment.Elasticsearch.Config.IsEmpty() {
		deployment.Elasticsearch.Config = nil
	}

	return deployment, diags
}

func deploymentNotFound(err error) bool {
	// We're using the As() call since we do not care about the error value
	// but do care about the error's contents type since it's an implicit 404.
	var notDeploymentNotFound *deployments.GetDeploymentNotFound
	if errors.As(err, &notDeploymentNotFound) {
		return true
	}

	// We also check for the case where a 403 is thrown for ESS.
	return apierror.IsRuntimeStatusCode(err, 403)
}

var minimumSupportedVersion = semver.MustParse("6.6.0")

func checkVersion(version string) error {
	v, err := semver.New(version)

	if err != nil {
		return fmt.Errorf("unable to parse deployment version: %w", err)
	}

	if v.LT(minimumSupportedVersion) {
		return fmt.Errorf(
			`invalid deployment version "%s": minimum supported version is "%s"`,
			v.String(), minimumSupportedVersion.String(),
		)
	}

	return nil
}

func HasRunningResources(res *models.DeploymentGetResponse) bool {
	if res.Resources != nil {
		for _, r := range res.Resources.Elasticsearch {
			if !elasticsearchv2.IsElasticsearchStopped(r) {
				return true
			}
		}
		for _, r := range res.Resources.Kibana {
			if !kibanav2.IsKibanaStopped(r) {
				return true
			}
		}
		for _, r := range res.Resources.Apm {
			if !apmv2.IsApmStopped(r) {
				return true
			}
		}
		for _, r := range res.Resources.EnterpriseSearch {
			if !enterprisesearchv2.IsEnterpriseSearchStopped(r) {
				return true
			}
		}
		for _, r := range res.Resources.IntegrationsServer {
			if !integrationsserverv2.IsIntegrationsServerStopped(r) {
				return true
			}
		}
	}
	return false
}
