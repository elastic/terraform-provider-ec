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

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	v1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/apm/v1"
	topologyv1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/topology/v1"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ApmTF struct {
	ElasticsearchClusterRefId          types.String `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                              types.String `tfsdk:"ref_id"`
	ResourceId                         types.String `tfsdk:"resource_id"`
	Region                             types.String `tfsdk:"region"`
	HttpEndpoint                       types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint                      types.String `tfsdk:"https_endpoint"`
	InstanceConfigurationId            types.String `tfsdk:"instance_configuration_id"`
	LatestInstanceConfigurationId      types.String `tfsdk:"latest_instance_configuration_id"`
	InstanceConfigurationVersion       types.Int64  `tfsdk:"instance_configuration_version"`
	LatestInstanceConfigurationVersion types.Int64  `tfsdk:"latest_instance_configuration_version"`
	Size                               types.String `tfsdk:"size"`
	SizeResource                       types.String `tfsdk:"size_resource"`
	ZoneCount                          types.Int64  `tfsdk:"zone_count"`
	Config                             types.Object `tfsdk:"config"`
}

func (apm ApmTF) payload(ctx context.Context, payload models.ApmPayload) (*models.ApmPayload, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !apm.ElasticsearchClusterRefId.IsNull() {
		payload.ElasticsearchClusterRefID = ec.String(apm.ElasticsearchClusterRefId.ValueString())
	}

	if !apm.RefId.IsNull() {
		payload.RefID = ec.String(apm.RefId.ValueString())
	}

	if apm.Region.ValueString() != "" {
		payload.Region = ec.String(apm.Region.ValueString())
	}

	if !apm.Config.IsNull() && !apm.Config.IsUnknown() {
		var cfg v1.ApmConfigTF

		ds := tfsdk.ValueAs(ctx, apm.Config, &cfg)

		diags.Append(ds...)

		if !ds.HasError() {
			diags.Append(apmConfigPayload(ctx, cfg, payload.Plan.Apm)...)
		}
	}

	topology := topologyv1.TopologyTF{
		InstanceConfigurationId:      apm.InstanceConfigurationId,
		InstanceConfigurationVersion: apm.InstanceConfigurationVersion,
		Size:                         apm.Size,
		SizeResource:                 apm.SizeResource,
		ZoneCount:                    apm.ZoneCount,
	}

	// Always use the first topology element - discard any other topology elements
	topologyPayload, ds := apmTopologyPayload(ctx, topology, defaultApmTopology(payload.Plan.ClusterTopology)[0])

	diags.Append(ds...)

	if !ds.HasError() && topologyPayload != nil {
		payload.Plan.ClusterTopology = []*models.ApmTopologyElement{topologyPayload}
	}

	return &payload, diags
}

func ApmPayload(ctx context.Context, apmObj types.Object, updateResources *models.DeploymentUpdateResources) (*models.ApmPayload, diag.Diagnostics) {
	apm, diags := objectToApm(ctx, apmObj)
	if diags.HasError() {
		return nil, diags
	}

	if apm == nil {
		return nil, nil
	}

	templatePayload := payloadFromUpdate(updateResources)

	if templatePayload == nil {
		diags.AddError("apm payload error", "apm specified but deployment template is not configured for it. Use a different template if you wish to add apm")
		return nil, diags
	}

	payload, diags := apm.payload(ctx, *templatePayload)

	if diags.HasError() {
		return nil, diags
	}

	return payload, nil
}

func objectToApm(ctx context.Context, plan types.Object) (*ApmTF, diag.Diagnostics) {
	var apm *ApmTF

	if plan.IsNull() || plan.IsUnknown() {
		return nil, nil
	}

	if diags := tfsdk.ValueAs(ctx, plan, &apm); diags.HasError() {
		return nil, diags
	}

	return apm, nil
}

func CheckAvailableMigration(ctx context.Context, plan types.Object, state types.Object) (bool, diag.Diagnostics) {
	apmPlan, diags := objectToApm(ctx, plan)
	if diags.HasError() {
		return false, diags
	}

	apmState, diags := objectToApm(ctx, state)
	if diags.HasError() {
		return false, diags
	}

	if apmPlan == nil || apmState == nil {
		return false, nil
	}

	// We won't migrate this topology element if 'instance_configuration_id' or 'instance_configuration_version' are
	// defined on the TF configuration. Otherwise, we may be setting an incorrect value for 'size', in case the
	// template IC has different size increments
	if !apmPlan.InstanceConfigurationId.IsUnknown() || !apmPlan.InstanceConfigurationVersion.IsUnknown() {
		return false, nil
	}

	instanceConfigIdsDiff := apmState.InstanceConfigurationId != apmState.LatestInstanceConfigurationId
	instanceConfigVersionsDiff := apmState.InstanceConfigurationVersion != apmState.LatestInstanceConfigurationVersion

	// We consider that a migration is available when:
	//    * the current instance config ID doesn't match the one in the template
	//    * the instance config IDs match but the instance config versions differ
	return instanceConfigIdsDiff || instanceConfigVersionsDiff, nil
}

// payloadFromUpdate returns the ApmPayload from a deployment
// template or an empty version of the payload.
func payloadFromUpdate(updateResources *models.DeploymentUpdateResources) *models.ApmPayload {
	if updateResources == nil || len(updateResources.Apm) == 0 {
		return nil
	}
	return updateResources.Apm[0]
}
