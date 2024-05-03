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
	loadTemplate func() (*models.DeploymentTemplateInfoV2, error),
) {
	var config es.ElasticsearchTF
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("elasticsearch"), &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.MasterTier.IsNull() && !config.MasterTier.IsUnknown() {
		// Master tier is explicitly configured -> No changes will be made
		tflog.Debug(ctx, "Skip UpdateDedicatedMasterTier: Master tier has been explicitly configured")
		return
	}

	var planElasticsearch es.ElasticsearchTF
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("elasticsearch"), &planElasticsearch)...)
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
		resp.Plan.SetAttribute(ctx,
			path.Root("elasticsearch").AtName("master"),
			types.ObjectNull(es.ElasticsearchTopologyAttrs()),
		)
	} else {
		// Enable master tier
		instanceConfiguration := getInstanceConfiguration(*template, "master")
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
	tiers := []string{"hot_content", "coordinating", "warm", "cold", "frozen"}
	for _, tier := range tiers {
		var rawTopology types.Object
		switch tier {
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

		var topology es.ElasticsearchTopologyTF
		diags := rawTopology.As(ctx, &topology, objectAsOptions)
		if diags.HasError() {
			tflog.Debug(ctx, "countNodesInCluster: Failed to read topology.", map[string]interface{}{"error": diags.Errors()})
			continue
		}

		var size int32
		if !topology.Size.IsUnknown() && !topology.Size.IsNull() {
			var err error
			size, err = deploymentsize.ParseGb(topology.Size.ValueString())
			if err != nil {
				tflog.Debug(ctx, "countNodesInCluster: Failed to parse topology size.", map[string]interface{}{"error": err})
				continue
			}
		} else {
			// Fall back to template value
			size = getTopologySize(template, tier)
		}

		// Calculate if there are >1 nodes in each zone:
		// If the size is > the maximum size allowed in the zone, more nodes are added to reach the desired size.
		instanceConfig := getInstanceConfiguration(template, tier)
		maxSize := getMaxSize(instanceConfig)
		var nodesPerZone int32
		if size < maxSize || maxSize == 0 {
			nodesPerZone = 1
		} else {
			nodesPerZone = size / maxSize
		}

		var zoneCount int32
		if !topology.ZoneCount.IsUnknown() && !topology.ZoneCount.IsNull() {
			zoneCount = int32(topology.ZoneCount.ValueInt64())
		} else {
			// Fall back to template value
			zoneCount = getTopologyZoneCount(template, tier)
		}

		if size > 0 && zoneCount > 0 {
			nodesInDeployment += zoneCount * nodesPerZone
		}
	}

	return nodesInDeployment
}

func getInstanceConfiguration(template models.DeploymentTemplateInfoV2, topologyId string) *models.InstanceConfigurationInfo {
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
