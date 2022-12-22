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
	"strconv"
	"strings"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/utils"
)

type ElasticsearchTF struct {
	Autoscale        types.String `tfsdk:"autoscale"`
	RefId            types.String `tfsdk:"ref_id"`
	ResourceId       types.String `tfsdk:"resource_id"`
	Region           types.String `tfsdk:"region"`
	CloudID          types.String `tfsdk:"cloud_id"`
	HttpEndpoint     types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint    types.String `tfsdk:"https_endpoint"`
	HotContentTier   types.Object `tfsdk:"hot"`
	CoordinatingTier types.Object `tfsdk:"coordinating"`
	MasterTier       types.Object `tfsdk:"master"`
	WarmTier         types.Object `tfsdk:"warm"`
	ColdTier         types.Object `tfsdk:"cold"`
	FrozenTier       types.Object `tfsdk:"frozen"`
	MlTier           types.Object `tfsdk:"ml"`
	Config           types.Object `tfsdk:"config"`
	RemoteCluster    types.Set    `tfsdk:"remote_cluster"`
	SnapshotSource   types.Object `tfsdk:"snapshot_source"`
	Extension        types.Set    `tfsdk:"extension"`
	TrustAccount     types.Set    `tfsdk:"trust_account"`
	TrustExternal    types.Set    `tfsdk:"trust_external"`
	Strategy         types.String `tfsdk:"strategy"`
}

func ElasticsearchPayload(ctx context.Context, esObj types.Object, template *models.DeploymentTemplateInfoV2, dtID, version string, useNodeRoles bool, skipTopologies bool) (*models.ElasticsearchPayload, diag.Diagnostics) {
	var es *ElasticsearchTF

	if esObj.IsNull() || esObj.IsUnknown() {
		return nil, nil
	}

	if diags := tfsdk.ValueAs(ctx, esObj, &es); diags.HasError() {
		return nil, diags
	}

	if es == nil {
		return nil, nil
	}

	if es == nil {
		var diags diag.Diagnostics
		diags.AddError("Elasticsearch payload error", "cannot find elasticsearch data")
		return nil, diags
	}

	templatePayload := utils.EnrichElasticsearchTemplate(utils.EsResource(template), dtID, version, useNodeRoles)

	payload, diags := es.Payload(ctx, templatePayload, skipTopologies)
	if diags.HasError() {
		return nil, diags
	}

	return payload, nil
}

func (es *ElasticsearchTF) Payload(ctx context.Context, res *models.ElasticsearchPayload, skipTopologies bool) (*models.ElasticsearchPayload, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !es.RefId.IsNull() {
		res.RefID = &es.RefId.Value
	}

	if es.Region.Value != "" {
		res.Region = &es.Region.Value
	}

	// Unsetting the curation properties is since they're deprecated since
	// >= 6.6.0 which is when ILM is introduced in Elasticsearch.
	unsetElasticsearchCuration(res)

	var ds diag.Diagnostics

	if !skipTopologies {
		diags.Append(es.topologiesPayload(ctx, res.Plan.ClusterTopology)...)
	}

	// Fixes the node_roles field to remove the dedicated tier roles from the
	// list when these are set as a dedicated tier as a topology element.
	updateNodeRolesOnDedicatedTiers(res.Plan.ClusterTopology)

	res.Plan.Elasticsearch, ds = ElasticsearchConfigPayload(ctx, es.Config, res.Plan.Elasticsearch)
	diags.Append(ds...)

	diags.Append(elasticsearchSnapshotSourcePayload(ctx, es.SnapshotSource, res.Plan)...)

	diags.Append(elasticsearchExtensionPayload(ctx, es.Extension, res.Plan.Elasticsearch)...)

	if es.Autoscale.Value != "" {
		autoscaleBool, err := strconv.ParseBool(es.Autoscale.Value)
		if err != nil {
			diags.AddError("failed parsing autoscale value", err.Error())
		} else {
			res.Plan.AutoscalingEnabled = &autoscaleBool
		}
	}

	res.Settings, ds = ElasticsearchTrustAccountPayload(ctx, es.TrustAccount, res.Settings)
	diags.Append(ds...)

	res.Settings, ds = ElasticsearchTrustExternalPayload(ctx, es.TrustExternal, res.Settings)
	diags.Append(ds...)

	elasticsearchStrategyPayload(es.Strategy, res.Plan)

	return res, diags
}

func (es *ElasticsearchTF) topologiesPayload(ctx context.Context, topologies []*models.ElasticsearchClusterTopologyElement) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(topologyPayload(ctx, es.HotContentTier, "hot_content", topologies)...)
	diags.Append(topologyPayload(ctx, es.CoordinatingTier, "coordinating", topologies)...)
	diags.Append(topologyPayload(ctx, es.MasterTier, "master", topologies)...)
	diags.Append(topologyPayload(ctx, es.WarmTier, "warm", topologies)...)
	diags.Append(topologyPayload(ctx, es.ColdTier, "cold", topologies)...)
	diags.Append(topologyPayload(ctx, es.FrozenTier, "frozen", topologies)...)
	diags.Append(topologyPayload(ctx, es.MlTier, "ml", topologies)...)

	return diags
}

func topologyPayload(ctx context.Context, topologyObj types.Object, id string, topologies []*models.ElasticsearchClusterTopologyElement) diag.Diagnostics {
	var diags diag.Diagnostics

	if !topologyObj.IsNull() && !topologyObj.IsUnknown() {
		var topology ElasticsearchTopologyTF

		ds := tfsdk.ValueAs(ctx, topologyObj, &topology)
		diags.Append(ds...)

		if !ds.HasError() {
			diags.Append(topology.Payload(ctx, id, topologies)...)
		}
	}

	return diags
}

func unsetElasticsearchCuration(payload *models.ElasticsearchPayload) {
	if payload.Plan.Elasticsearch != nil {
		payload.Plan.Elasticsearch.Curation = nil
	}

	if payload.Settings != nil {
		payload.Settings.Curation = nil
	}
}

func updateNodeRolesOnDedicatedTiers(topologies []*models.ElasticsearchClusterTopologyElement) {
	dataTier, hasMasterTier, hasIngestTier := dedicatedTopoogies(topologies)
	// This case is not very likely since all deployments will have a data tier.
	// It's here because the code path is technically possible and it's better
	// than a straight panic.
	if dataTier == nil {
		return
	}

	if hasIngestTier {
		dataTier.NodeRoles = removeItemFromSlice(
			dataTier.NodeRoles, ingestDataTierRole,
		)
	}
	if hasMasterTier {
		dataTier.NodeRoles = removeItemFromSlice(
			dataTier.NodeRoles, masterDataTierRole,
		)
	}
}

func removeItemFromSlice(slice []string, item string) []string {
	var hasItem bool
	var itemIndex int
	for i, str := range slice {
		if str == item {
			hasItem = true
			itemIndex = i
		}
	}
	if hasItem {
		copy(slice[itemIndex:], slice[itemIndex+1:])
		return slice[:len(slice)-1]
	}
	return slice
}

func dedicatedTopoogies(topologies []*models.ElasticsearchClusterTopologyElement) (dataTier *models.ElasticsearchClusterTopologyElement, hasMasterTier, hasIngestTier bool) {
	for _, topology := range topologies {
		var hasSomeDataRole bool
		var hasMasterRole bool
		var hasIngestRole bool
		for _, role := range topology.NodeRoles {
			sizeNonZero := *topology.Size.Value > 0
			if strings.HasPrefix(role, dataTierRolePrefix) && sizeNonZero {
				hasSomeDataRole = true
			}
			if role == ingestDataTierRole && sizeNonZero {
				hasIngestRole = true
			}
			if role == masterDataTierRole && sizeNonZero {
				hasMasterRole = true
			}
		}

		if !hasSomeDataRole && hasMasterRole {
			hasMasterTier = true
		}

		if !hasSomeDataRole && hasIngestRole {
			hasIngestTier = true
		}

		if hasSomeDataRole && hasMasterRole {
			dataTier = topology
		}
	}

	return dataTier, hasMasterTier, hasIngestTier
}

func elasticsearchStrategyPayload(strategy types.String, payload *models.ElasticsearchClusterPlan) {
	createModelIfNeeded := func() {
		if payload.Transient == nil {
			payload.Transient = &models.TransientElasticsearchPlanConfiguration{
				Strategy: &models.PlanStrategy{},
			}
		}
	}

	switch strategy.Value {
	case autodetect:
		createModelIfNeeded()
		payload.Transient.Strategy.Autodetect = new(models.AutodetectStrategyConfig)
	case growAndShrink:
		createModelIfNeeded()
		payload.Transient.Strategy.GrowAndShrink = new(models.GrowShrinkStrategyConfig)
	case rollingGrowAndShrink:
		createModelIfNeeded()
		payload.Transient.Strategy.RollingGrowAndShrink = new(models.RollingGrowShrinkStrategyConfig)
	case rollingAll:
		createModelIfNeeded()
		payload.Transient.Strategy.Rolling = &models.RollingStrategyConfig{
			GroupBy: "__all__",
		}
	}
}
