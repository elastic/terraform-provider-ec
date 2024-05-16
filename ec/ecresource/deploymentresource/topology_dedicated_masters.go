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
	"fmt"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deploymentsize"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	es "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var objectAsOptions = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false}

func UpdateDedicatedMasterTier(
	ctx context.Context,
	config tfsdk.Config,
	plan tfsdk.Plan,
	privateState PrivateState,
	resp *resource.ModifyPlanResponse,
	loadTemplate func() (*models.DeploymentTemplateInfoV2, error),
) {
	var esConfig es.ElasticsearchTF
	resp.Diagnostics.Append(config.GetAttribute(ctx, path.Root("elasticsearch"), &esConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !esConfig.MasterTier.IsNull() && !esConfig.MasterTier.IsUnknown() {
		// Master tier is explicitly configured -> No changes will be made
		tflog.Debug(ctx, "Skip UpdateDedicatedMasterTier: Master tier has been explicitly configured")
		return
	}

	var planElasticsearch es.ElasticsearchTF
	resp.Diagnostics.Append(plan.GetAttribute(ctx, path.Root("elasticsearch"), &planElasticsearch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	template, err := loadTemplate()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get deployment-template", "Error: "+err.Error())
		return
	}

	dedicatedMastersThreshold := getDedicatedMastersThreshold(*template)
	if dedicatedMastersThreshold == 0 {
		// No automatic dedicated masters management
		return
	}

	nodesInCluster := countNodesInCluster(ctx, planElasticsearch, *template)

	if nodesInCluster < dedicatedMastersThreshold {
		// Disable master tier
		if planElasticsearch.MasterTier.IsUnknown() || planElasticsearch.MasterTier.IsNull() {
			resp.Plan.SetAttribute(ctx,
				path.Root("elasticsearch").AtName("master"),
				types.ObjectNull(es.ElasticsearchTopologyAttrs()),
			)
		} else {
			resp.Plan.SetAttribute(ctx,
				path.Root("elasticsearch").AtName("master").AtName("size"),
				"0g",
			)
		}
	} else {
		var migrateToLatestHw bool
		plan.GetAttribute(ctx, path.Root("migrate_to_latest_hardware"), &migrateToLatestHw)

		// Skip update if the master tier is already enabled
		// If migrateToLatestHw is true, update the tier to values from latest IC
		if masterTierIsEnabled(ctx, planElasticsearch, *template) && !migrateToLatestHw {
			return
		}

		// Enable master tier

		instanceConfigurations, diags := ReadPrivateStateInstanceConfigurations(ctx, privateState)
		if diags.HasError() {
			tflog.Debug(ctx, "Failed to read instance-configs from private state", withDiags(diags))
			return
		}

		templateInstanceConfig := getTemplateInstanceConfiguration(*template, "master")
		instanceConfiguration := getInstanceConfiguration(ctx, planElasticsearch.MasterTier, instanceConfigurations)
		if instanceConfiguration == nil || instanceConfiguration.DiscreteSizes == nil {
			// Fall back to template IC
			instanceConfiguration = templateInstanceConfig
		}
		if instanceConfiguration == nil || instanceConfiguration.DiscreteSizes == nil {
			tflog.Debug(ctx, "UpdateDedicatedMasterTier: Could not enable master tier, as it has no instance-config.")
			return
		}

		// Zones are
		zones := instanceConfiguration.MaxZones
		if zones == 0 {
			// Fall back to template if no max-zones is set
			zones = templateInstanceConfig.MaxZones
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

func masterTierIsEnabled(ctx context.Context, planElasticsearch es.ElasticsearchTF, template models.DeploymentTemplateInfoV2) bool {
	if planElasticsearch.MasterTier.IsUnknown() || planElasticsearch.MasterTier.IsNull() {
		return false
	}

	size, zoneCount := getSizeAndZoneCount(ctx, "master", planElasticsearch.MasterTier, template)

	return size > 0 && zoneCount > 0
}

func getDedicatedMastersThreshold(template models.DeploymentTemplateInfoV2) int32 {
	dedicatedMastersThreshold := int32(0)
	if template.DeploymentTemplate != nil && template.DeploymentTemplate.Resources != nil {
		for _, e := range template.DeploymentTemplate.Resources.Elasticsearch {
			if e.Settings == nil {
				continue
			}
			dedicatedMastersThreshold = e.Settings.DedicatedMastersThreshold
			break
		}
	}
	return dedicatedMastersThreshold
}

func countNodesInCluster(ctx context.Context, esPlan es.ElasticsearchTF, template models.DeploymentTemplateInfoV2) int32 {
	nodesInDeployment := int32(0)
	tierTopologyIds := []string{"hot_content", "coordinating", "warm", "cold", "frozen"}
	for _, topologyId := range tierTopologyIds {
		var rawTopology types.Object
		switch topologyId {
		case "hot_content":
			rawTopology = esPlan.HotContentTier
		case "coordinating":
			rawTopology = esPlan.CoordinatingTier
		case "warm":
			rawTopology = esPlan.WarmTier
		case "cold":
			rawTopology = esPlan.ColdTier
		case "frozen":
			rawTopology = esPlan.FrozenTier
		}

		if rawTopology.IsNull() || rawTopology.IsUnknown() {
			continue
		}

		size, zoneCount := getSizeAndZoneCount(ctx, topologyId, rawTopology, template)

		// Calculate if there are >1 nodes in each zone:
		// If the size is > the maximum size allowed in the zone, more nodes are added to reach the desired size.
		instanceConfig := getTemplateInstanceConfiguration(template, topologyId)
		maxSize := getMaxSize(instanceConfig)
		var nodesPerZone int32
		if size < maxSize || maxSize == 0 {
			nodesPerZone = 1
		} else {
			nodesPerZone = size / maxSize
		}

		if size > 0 && zoneCount > 0 {
			nodesInDeployment += zoneCount * nodesPerZone
		}
	}

	return nodesInDeployment
}

func getSizeAndZoneCount(
	ctx context.Context,
	topologyId string,
	rawTopology types.Object,
	template models.DeploymentTemplateInfoV2,
) (int32, int32) {
	var topology es.ElasticsearchTopologyTF
	diags := rawTopology.As(ctx, &topology, objectAsOptions)
	if diags.HasError() {
		tflog.Debug(ctx, "getSizeAndZoneCount: Failed to read topology.", withDiags(diags))
		return 0, 0
	}

	var size int32
	if !topology.Size.IsUnknown() && !topology.Size.IsNull() {
		var err error
		size, err = deploymentsize.ParseGb(topology.Size.ValueString())
		if err != nil {
			tflog.Debug(ctx, "getSizeAndZoneCount: Failed to parse topology size.", withError(err))
			return 0, 0
		}
	} else {
		// Fall back to template value
		size = getTopologySize(template, topologyId)
	}

	var zoneCount int32
	if !topology.ZoneCount.IsUnknown() && !topology.ZoneCount.IsNull() {
		zoneCount = int32(topology.ZoneCount.ValueInt64())
	} else {
		// Fall back to template value
		zoneCount = getTopologyZoneCount(template, topologyId)
	}

	return size, zoneCount
}

// If the instance-config-id is set in the topology, loads that specific IC via API
// If no instance-config-id is set, uses the IC set in the deployment template
func getInstanceConfiguration(
	ctx context.Context,
	rawTopology types.Object,
	deploymentInstanceConfigs []models.InstanceConfigurationInfo,
) *models.InstanceConfigurationInfo {

	if rawTopology.IsUnknown() || rawTopology.IsNull() {
		return nil
	}

	var topology es.ElasticsearchTopologyTF
	diags := rawTopology.As(ctx, &topology, objectAsOptions)
	if diags.HasError() {
		tflog.Debug(ctx, "getInstanceConfigurationId: Failed to read topology.", withDiags(diags))
		return nil
	}

	if topology.InstanceConfigurationId.IsUnknown() || topology.InstanceConfigurationId.IsNull() {
		return nil
	}

	icId := topology.InstanceConfigurationId.ValueStringPointer()
	if icId == nil || *icId == "" {
		return nil
	}

	var deploymentIc *models.InstanceConfigurationInfo
	for _, ic := range deploymentInstanceConfigs {
		if ic.ID == *icId {
			deploymentIc = &ic
			break
		}
	}
	if deploymentIc == nil {
		tflog.Debug(ctx, fmt.Sprintf("UpdateDedicatedMasterTier: Instance-config not found: %s", *icId))
		return nil
	}

	return deploymentIc
}

func getTemplateInstanceConfiguration(template models.DeploymentTemplateInfoV2, topologyId string) *models.InstanceConfigurationInfo {
	if template.DeploymentTemplate == nil ||
		template.DeploymentTemplate.Resources == nil ||
		len(template.DeploymentTemplate.Resources.Elasticsearch) == 0 ||
		template.DeploymentTemplate.Resources.Elasticsearch[0].Plan == nil {
		return nil
	}
	var topologyElement *models.ElasticsearchClusterTopologyElement
	for _, topology := range template.DeploymentTemplate.Resources.Elasticsearch[0].Plan.ClusterTopology {
		if topology.ID == topologyId {
			topologyElement = topology
			break
		}
	}
	if topologyElement == nil {
		return nil
	}

	// Find IC for tier
	for _, ic := range template.InstanceConfigurations {
		if ic.ID == topologyElement.InstanceConfigurationID {
			return ic
		}
	}

	return nil
}

func getMaxSize(ic *models.InstanceConfigurationInfo) int32 {
	if ic == nil || ic.DiscreteSizes == nil {
		return 0
	}
	maxSize := int32(0)
	for _, size := range ic.DiscreteSizes.Sizes {
		if size > maxSize {
			maxSize = size
		}
	}
	return maxSize
}

func getTopologySize(template models.DeploymentTemplateInfoV2, tier string) int32 {
	if len(template.DeploymentTemplate.Resources.Elasticsearch) == 0 {
		return 0
	}

	elasticsearch := template.DeploymentTemplate.Resources.Elasticsearch[0]
	for _, topology := range elasticsearch.Plan.ClusterTopology {
		if topology.ID == tier {
			if topology.Size == nil || topology.Size.Value == nil {
				return 0
			}
			return *topology.Size.Value
		}
	}
	return 0
}

func getTopologyZoneCount(template models.DeploymentTemplateInfoV2, tier string) int32 {
	if len(template.DeploymentTemplate.Resources.Elasticsearch) == 0 {
		return 0
	}

	elasticsearch := template.DeploymentTemplate.Resources.Elasticsearch[0]
	for _, topology := range elasticsearch.Plan.ClusterTopology {
		if topology.ID == tier {
			return topology.ZoneCount
		}
	}
	return 0
}

func withError(err error) map[string]interface{} {
	return map[string]interface{}{"error": err}
}

func withDiags(diags diag.Diagnostics) map[string]interface{} {
	return map[string]interface{}{"error": diags.Errors()}
}
