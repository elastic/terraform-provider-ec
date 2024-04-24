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

package planmodifiers

import (
	"context"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deploymentsize"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	es "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var objectAsOptions = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false}

func UpdateDedicatedMasterTier(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
	template models.DeploymentTemplateInfoV2,
) {
	var planElasticsearch es.ElasticsearchTF
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("elasticsearch"), &planElasticsearch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dedicatedMastersThreshold := getDedicatedMastersThreshold(template)
	if dedicatedMastersThreshold == 0 {
		// No automatic dedicated masters management
		return
	}

	nodesInCluster := countNodesInCluster(ctx, planElasticsearch)
	if resp.Diagnostics.HasError() {
		return
	}

	var config es.ElasticsearchTF
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("elasticsearch"), &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.MasterTier.IsNull() && !config.MasterTier.IsUnknown() {
		// Master tier is explicitly configured -> No changes will be made
		tflog.Debug(ctx, "UpdateDedicatedMasterTier: Could not enable master tier, as it has no instance-config.")
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	if nodesInCluster < dedicatedMastersThreshold {
		// Disable master tier
		resp.Plan.SetAttribute(ctx,
			path.Root("elasticsearch").AtName("master"),
			types.ObjectNull(es.ElasticsearchTopologyAttrs()),
		)
	} else {
		// Enable master tier
		instanceConfiguration := getInstanceConfiguration(template)
		if instanceConfiguration == nil {
			tflog.Debug(ctx, "UpdateDedicatedMasterTier: Could not enable master tier, as it has no instance-config.")
			return
		}

		if instanceConfiguration.DiscreteSizes == nil {
			return
		}

		// Set zones
		zones := instanceConfiguration.MaxZones
		if zones == 0 {
			return
		}
		resp.Plan.SetAttribute(ctx,
			path.Root("elasticsearch").AtName("master").AtName("zone_count"),
			zones,
		)

		// Set Size
		defaultSize := util.MemoryToState(instanceConfiguration.DiscreteSizes.DefaultSize)
		resp.Plan.SetAttribute(ctx,
			path.Root("elasticsearch").AtName("master").AtName("size"),
			defaultSize,
		)
	}
}

func getInstanceConfiguration(template models.DeploymentTemplateInfoV2) *models.InstanceConfigurationInfo {
	// Find master tier
	if template.DeploymentTemplate == nil ||
		template.DeploymentTemplate.Resources == nil ||
		len(template.DeploymentTemplate.Resources.Elasticsearch) == 0 ||
		template.DeploymentTemplate.Resources.Elasticsearch[0].Plan == nil {
		return nil
	}
	var masterTier *models.ElasticsearchClusterTopologyElement
	for _, topology := range template.DeploymentTemplate.Resources.Elasticsearch[0].Plan.ClusterTopology {
		if topology.ID == "master" {
			masterTier = topology
		}
	}
	if masterTier == nil {
		return nil
	}

	// Find IC for master tier
	for _, ic := range template.InstanceConfigurations {
		if ic.ID == masterTier.InstanceConfigurationID {
			return ic
		}
	}

	return nil
}

func getDedicatedMastersThreshold(template models.DeploymentTemplateInfoV2) int64 {
	dedicatedMastersThreshold := 0
	if template.DeploymentTemplate != nil && template.DeploymentTemplate.Resources != nil {
		for _, e := range template.DeploymentTemplate.Resources.Elasticsearch {
			if e.Settings == nil {
				continue
			}
			dedicatedMastersThreshold = int(e.Settings.DedicatedMastersThreshold)
			break
		}
	}
	return int64(dedicatedMastersThreshold)
}

func countNodesInCluster(
	ctx context.Context,
	planElasticsearch es.ElasticsearchTF,
) int64 {
	nodesInDeployment := int64(0)
	tiers := []types.Object{
		planElasticsearch.HotContentTier,
		planElasticsearch.CoordinatingTier,
		planElasticsearch.WarmTier,
		planElasticsearch.ColdTier,
		planElasticsearch.FrozenTier,
	}
	for _, tier := range tiers {
		if tier.IsNull() || tier.IsUnknown() {
			continue
		}

		var topology es.ElasticsearchTopologyTF
		diags := tier.As(ctx, &topology, objectAsOptions)
		if diags.HasError() {
			tflog.Debug(ctx, "countNodesInCluster: Failed to read topology.", map[string]interface{}{"error": diags.Errors()})
			continue
		}

		size, err := deploymentsize.ParseGb(topology.Size.ValueString())
		if err != nil {
			tflog.Debug(ctx, "countNodesInCluster: Failed to parse topology size.", map[string]interface{}{"error": err})
			continue
		}

		zoneCount := topology.ZoneCount.ValueInt64()
		if size > 0 && zoneCount > 0 {
			nodesInDeployment += zoneCount
		}
	}

	return nodesInDeployment
}
