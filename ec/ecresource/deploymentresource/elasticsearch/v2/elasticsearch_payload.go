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
	"slices"
	"strings"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchTF struct {
	Autoscale        types.Bool   `tfsdk:"autoscale"`
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
	Snapshot         types.Object `tfsdk:"snapshot"`
	SnapshotSource   types.Object `tfsdk:"snapshot_source"`
	Extension        types.Set    `tfsdk:"extension"`
	TrustAccount     types.Set    `tfsdk:"trust_account"`
	TrustExternal    types.Set    `tfsdk:"trust_external"`
	Strategy         types.String `tfsdk:"strategy"`
	KeystoreContents types.Map    `tfsdk:"keystore_contents"`
}

func ElasticsearchPayload(ctx context.Context, plan types.Object, state *types.Object, updateResources *models.DeploymentUpdateResources, dtID, version string, useNodeRoles bool) (*models.ElasticsearchPayload, diag.Diagnostics) {
	es, diags := objectToElasticsearch(ctx, plan)
	if diags.HasError() {
		return nil, diags
	}

	if es == nil {
		return nil, nil
	}

	var esState *ElasticsearchTF
	if state != nil {
		esState, diags = objectToElasticsearch(ctx, *state)
		if diags.HasError() {
			return nil, diags
		}
	}

	templatePayload := EnrichElasticsearchTemplate(payloadFromUpdate(updateResources), dtID, version, useNodeRoles)

	payload, diags := es.payload(ctx, templatePayload, esState)
	if diags.HasError() {
		return nil, diags
	}

	return payload, nil
}

func objectToElasticsearch(ctx context.Context, plan types.Object) (*ElasticsearchTF, diag.Diagnostics) {
	var es *ElasticsearchTF

	if plan.IsNull() || plan.IsUnknown() {
		return nil, nil
	}

	if diags := tfsdk.ValueAs(ctx, plan, &es); diags.HasError() {
		return nil, diags
	}

	return es, nil
}

func CheckAvailableMigration(ctx context.Context, plan types.Object, state types.Object) (bool, diag.Diagnostics) {
	esPlan, diags := objectToElasticsearch(ctx, plan)
	if diags.HasError() {
		return false, diags
	}

	esState, diags := objectToElasticsearch(ctx, state)
	if diags.HasError() {
		return false, diags
	}

	if esPlan == nil || esState == nil {
		return false, nil
	}

	planTiers, diags := esPlan.topologies(ctx)
	if diags.HasError() {
		return false, diags
	}

	stateTiers, diags := esState.topologies(ctx)
	if diags.HasError() {
		return false, diags
	}

	for topologyId, tier := range planTiers {
		if tier != nil && stateTiers[topologyId] != nil && tier.checkAvailableMigration(stateTiers[topologyId]) {
			return true, nil
		}
	}

	return false, nil
}

func (es *ElasticsearchTF) payload(ctx context.Context, res *models.ElasticsearchPayload, state *ElasticsearchTF) (*models.ElasticsearchPayload, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !es.RefId.IsNull() {
		res.RefID = ec.String(es.RefId.ValueString())
	}

	if es.Region.ValueString() != "" {
		res.Region = ec.String(es.Region.ValueString())
	}

	// Unsetting the curation properties is since they're deprecated since
	// >= 6.6.0 which is when ILM is introduced in Elasticsearch.
	unsetElasticsearchCuration(res)

	var ds diag.Diagnostics

	diags.Append(es.topologiesPayload(ctx, res.Plan.ClusterTopology)...)

	// Fixes the node_roles field to remove the dedicated tier roles from the
	// list when these are set as a dedicated tier as a topology element.
	updateNodeRolesOnDedicatedTiers(res.Plan.ClusterTopology)

	res.Plan.Elasticsearch, ds = elasticsearchConfigPayload(ctx, es.Config, res.Plan.Elasticsearch)
	diags.Append(ds...)

	res.Settings, ds = elasticsearchSnapshotPayload(ctx, es.Snapshot, res.Settings, state)
	diags.Append(ds...)

	diags.Append(elasticsearchSnapshotSourcePayload(ctx, es.SnapshotSource, res.Plan)...)

	diags.Append(elasticsearchExtensionPayload(ctx, es.Extension, res.Plan.Elasticsearch)...)

	if !es.Autoscale.IsNull() && !es.Autoscale.IsUnknown() {
		res.Plan.AutoscalingEnabled = ec.Bool(es.Autoscale.ValueBool())
	}

	// Only add trust settings to update payload if trust has changed
	if state == nil || !es.TrustAccount.Equal(state.TrustAccount) || !es.TrustExternal.Equal(state.TrustExternal) {
		res.Settings, ds = elasticsearchTrustAccountPayload(ctx, es.TrustAccount, res.Settings)
		diags.Append(ds...)

		res.Settings, ds = elasticsearchTrustExternalPayload(ctx, es.TrustExternal, res.Settings)
		diags.Append(ds...)
	}

	res.Settings, ds = elasticsearchKeystoreContentsPayload(ctx, es.KeystoreContents, res.Settings, state)
	diags.Append(ds...)

	elasticsearchStrategyPayload(es.Strategy, res.Plan)

	return res, diags
}

func (es *ElasticsearchTF) topologyObjects() map[string]types.Object {
	return map[string]types.Object{
		"hot_content":  es.HotContentTier,
		"warm":         es.WarmTier,
		"cold":         es.ColdTier,
		"frozen":       es.FrozenTier,
		"ml":           es.MlTier,
		"master":       es.MasterTier,
		"coordinating": es.CoordinatingTier,
	}
}

func (es *ElasticsearchTF) topologies(ctx context.Context) (map[string]*ElasticsearchTopologyTF, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	tierObjects := es.topologyObjects()
	res := make(map[string]*ElasticsearchTopologyTF, len(tierObjects))

	for topologyId, topologyObject := range tierObjects {
		tier, diags := objectToTopology(ctx, topologyObject)
		diagnostics.Append(diags...)
		res[topologyId] = tier
	}

	return res, diagnostics
}

func (es *ElasticsearchTF) topologiesPayload(ctx context.Context, topologyModels []*models.ElasticsearchClusterTopologyElement) diag.Diagnostics {
	tiers, diags := es.topologies(ctx)

	if diags.HasError() {
		return diags
	}

	for tierId, tier := range tiers {
		if tier != nil {
			diags.Append(tier.payload(ctx, tierId, topologyModels)...)
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
	i := slices.Index(slice, item)

	if i == -1 {
		return slice
	}

	return slices.Delete(slice, i, i+1)
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
			payload.Transient = &models.TransientElasticsearchPlanConfiguration{}
		}
		if payload.Transient.Strategy == nil {
			payload.Transient.Strategy = &models.PlanStrategy{}
		}
	}

	switch strategy.ValueString() {
	case strategyAutodetect:
		createModelIfNeeded()
		payload.Transient.Strategy.Autodetect = models.AutodetectStrategyConfig(map[string]interface{}{})
	case strategyGrowAndShrink:
		createModelIfNeeded()
		payload.Transient.Strategy.GrowAndShrink = models.GrowShrinkStrategyConfig(map[string]interface{}{})
	case strategyRollingGrowAndShrink:
		createModelIfNeeded()
		payload.Transient.Strategy.RollingGrowAndShrink = models.RollingGrowShrinkStrategyConfig(map[string]interface{}{})
	case strategyRollingAll:
		createModelIfNeeded()
		payload.Transient.Strategy.Rolling = &models.RollingStrategyConfig{
			GroupBy: "__all__",
		}
	}
}

func payloadFromUpdate(updateResources *models.DeploymentUpdateResources) *models.ElasticsearchPayload {
	if updateResources == nil || len(updateResources.Elasticsearch) == 0 {
		return &models.ElasticsearchPayload{
			Plan: &models.ElasticsearchClusterPlan{
				Elasticsearch: &models.ElasticsearchConfiguration{},
			},
			Settings: &models.ElasticsearchClusterSettings{},
		}
	}
	return updateResources.Elasticsearch[0]
}

func EnrichElasticsearchTemplate(tpl *models.ElasticsearchPayload, templateId, version string, useNodeRoles bool) *models.ElasticsearchPayload {
	if tpl.Plan.DeploymentTemplate == nil {
		tpl.Plan.DeploymentTemplate = &models.DeploymentTemplateReference{}
	}

	if tpl.Plan.DeploymentTemplate.ID == nil || *tpl.Plan.DeploymentTemplate.ID == "" {
		tpl.Plan.DeploymentTemplate.ID = ec.String(templateId)
	}

	if tpl.Plan.Elasticsearch.Version == "" {
		tpl.Plan.Elasticsearch.Version = version
	}

	for _, topology := range tpl.Plan.ClusterTopology {
		if useNodeRoles {
			topology.NodeType = nil
			continue
		}
		topology.NodeRoles = nil
	}

	return tpl
}
