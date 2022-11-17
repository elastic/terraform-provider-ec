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
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deploymentsize"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

// These constants are only used to determine whether or not a dedicated
// tier of masters or ingest (coordinating) nodes are set.
const (
	dataTierRolePrefix   = "data_"
	ingestDataTierRole   = "ingest"
	masterDataTierRole   = "master"
	autodetect           = "autodetect"
	growAndShrink        = "grow_and_shrink"
	rollingGrowAndShrink = "rolling_grow_and_shrink"
	rollingAll           = "rolling_all"
)

// List of update strategies availables.
var strategiesList = []string{
	autodetect, growAndShrink, rollingGrowAndShrink, rollingAll,
}

// expandEsResources expands Elasticsearch resources
func expandEsResources(ess []interface{}, tpl *models.ElasticsearchPayload) ([]*models.ElasticsearchPayload, error) {
	if len(ess) == 0 {
		return nil, nil
	}

	result := make([]*models.ElasticsearchPayload, 0, len(ess))
	for _, raw := range ess {
		resResource, err := expandEsResource(raw, tpl)
		if err != nil {
			return nil, err
		}
		result = append(result, resResource)
	}

	return result, nil
}

// expandEsResource expands a single Elasticsearch resource
func expandEsResource(raw interface{}, res *models.ElasticsearchPayload) (*models.ElasticsearchPayload, error) {
	es := raw.(map[string]interface{})

	if refID, ok := es["ref_id"].(string); ok {
		res.RefID = ec.String(refID)
	}

	if region, ok := es["region"].(string); ok && region != "" {
		res.Region = ec.String(region)
	}

	// Unsetting the curation properties is since they're deprecated since
	// >= 6.6.0 which is when ILM is introduced in Elasticsearch.
	unsetElasticsearchCuration(res)

	if rt, ok := es["topology"].([]interface{}); ok && len(rt) > 0 {
		topology, err := expandEsTopology(rt, res.Plan.ClusterTopology)
		if err != nil {
			return nil, err
		}
		res.Plan.ClusterTopology = topology
	}

	// Fixes the node_roles field to remove the dedicated tier roles from the
	// list when these are set as a dedicated tier as a topology element.
	updateNodeRolesOnDedicatedTiers(res.Plan.ClusterTopology)

	if cfg, ok := es["config"].([]interface{}); ok {
		if err := expandEsConfig(cfg, res.Plan.Elasticsearch); err != nil {
			return nil, err
		}
	}

	if snap, ok := es["snapshot_source"].([]interface{}); ok && len(snap) > 0 {
		res.Plan.Transient = &models.TransientElasticsearchPlanConfiguration{
			RestoreSnapshot: &models.RestoreSnapshotConfiguration{},
		}
		expandSnapshotSource(snap, res.Plan.Transient.RestoreSnapshot)
	}

	if ext, ok := es["extension"].(*schema.Set); ok && ext.Len() > 0 {
		expandEsExtension(ext.List(), res.Plan.Elasticsearch)
	}

	if autoscale, ok := es["autoscale"].(string); ok && autoscale != "" {
		autoscaleBool, err := strconv.ParseBool(autoscale)
		if err != nil {
			return nil, fmt.Errorf("failed parsing autoscale value: %w", err)
		}
		res.Plan.AutoscalingEnabled = &autoscaleBool
	}

	if trust, ok := es["trust_account"].(*schema.Set); ok && trust.Len() > 0 {
		if res.Settings == nil {
			res.Settings = &models.ElasticsearchClusterSettings{}
		}
		expandAccountTrust(trust.List(), res.Settings)
	}

	if trust, ok := es["trust_external"].(*schema.Set); ok && trust.Len() > 0 {
		if res.Settings == nil {
			res.Settings = &models.ElasticsearchClusterSettings{}
		}
		expandExternalTrust(trust.List(), res.Settings)
	}

	if strategy, ok := es["strategy"].([]interface{}); ok && len(strategy) > 0 {
		if res.Plan.Transient == nil {
			res.Plan.Transient = &models.TransientElasticsearchPlanConfiguration{
				Strategy: &models.PlanStrategy{},
			}
		}
		expandStrategy(strategy, res.Plan.Transient.Strategy)
	}

	return res, nil
}

// expandStrategy expands the Configuration Strategy.
func expandStrategy(raw []interface{}, strategy *models.PlanStrategy) {
	for _, rawStrategy := range raw {
		strategyCfg, ok := rawStrategy.(map[string]interface{})
		if !ok {
			continue
		}

		rawValue, ok := strategyCfg["type"].(string)
		if !ok {
			continue
		}

		if rawValue == autodetect {
			strategy.Autodetect = new(models.AutodetectStrategyConfig)
		} else if rawValue == growAndShrink {
			strategy.GrowAndShrink = new(models.GrowShrinkStrategyConfig)
		} else if rawValue == rollingGrowAndShrink {
			strategy.RollingGrowAndShrink = new(models.RollingGrowShrinkStrategyConfig)
		} else if rawValue == rollingAll {
			strategy.Rolling = &models.RollingStrategyConfig{
				GroupBy: "__all__",
			}
		}
	}
}

// expandEsTopology expands a flattened topology
func expandEsTopology(rawTopologies []interface{}, topologies []*models.ElasticsearchClusterTopologyElement) ([]*models.ElasticsearchClusterTopologyElement, error) {
	res := topologies

	for _, rawTop := range rawTopologies {
		topology, ok := rawTop.(map[string]interface{})
		if !ok {
			continue
		}

		var topologyID string
		if id, ok := topology["id"].(string); ok {
			topologyID = id
		}

		size, err := util.ParseTopologySize(topology)
		if err != nil {
			return nil, err
		}

		elem, err := matchEsTopologyID(topologyID, topologies)
		if err != nil {
			return nil, fmt.Errorf("elasticsearch topology %s: %w", topologyID, err)
		}
		if size != nil {
			elem.Size = size
		}

		if zones, ok := topology["zone_count"].(int); ok && zones > 0 {
			elem.ZoneCount = int32(zones)
		}

		if err := parseLegacyNodeType(topology, elem.NodeType); err != nil {
			return nil, err
		}

		if nrSet, ok := topology["node_roles"].(*schema.Set); ok && nrSet.Len() > 0 {
			elem.NodeRoles = util.ItemsToString(nrSet.List())
			elem.NodeType = nil
		}

		if autoscalingRaw, ok := topology["autoscaling"].([]interface{}); ok && len(autoscalingRaw) > 0 {
			for _, autoscaleRaw := range autoscalingRaw {
				autoscale, ok := autoscaleRaw.(map[string]interface{})
				if !ok {
					continue
				}

				if elem.AutoscalingMax == nil {
					elem.AutoscalingMax = new(models.TopologySize)
				}

				if elem.AutoscalingMin == nil {
					elem.AutoscalingMin = new(models.TopologySize)
				}

				err := expandAutoscalingDimension(autoscale, elem.AutoscalingMax, "max")
				if err != nil {
					return nil, err
				}

				err = expandAutoscalingDimension(autoscale, elem.AutoscalingMin, "min")
				if err != nil {
					return nil, err
				}

				// Ensure that if the Min and Max are empty, they're nil.
				if reflect.DeepEqual(elem.AutoscalingMin, new(models.TopologySize)) {
					elem.AutoscalingMin = nil
				}
				if reflect.DeepEqual(elem.AutoscalingMax, new(models.TopologySize)) {
					elem.AutoscalingMax = nil
				}

				if policy, ok := autoscale["policy_override_json"].(string); ok && policy != "" {
					if err := json.Unmarshal([]byte(policy),
						&elem.AutoscalingPolicyOverrideJSON,
					); err != nil {
						return nil, fmt.Errorf(
							"elasticsearch topology %s: unable to load policy_override_json: %w",
							topologyID, err,
						)
					}
				}
			}
		}

		if cfg, ok := topology["config"].([]interface{}); ok {
			if elem.Elasticsearch == nil {
				elem.Elasticsearch = &models.ElasticsearchConfiguration{}
			}
			if err := expandEsConfig(cfg, elem.Elasticsearch); err != nil {
				return nil, err
			}
		}
	}

	return res, nil
}

// expandAutoscalingDimension centralises processing of %_size and %_size_resource attributes
// Due to limitations in the Terraform SDK, it's not possible to specify a Default on a Computed schema member
// to work around this limitation, this function will default the %_size_resource attribute to `memory`.
// Without this default, setting autoscaling limits on tiers which do not have those limits in the deployment
// template leads to an API error due to the empty resource field on the TopologySize model.
func expandAutoscalingDimension(autoscale map[string]interface{}, model *models.TopologySize, dimension string) error {
	sizeAttribute := fmt.Sprintf("%s_size", dimension)
	resourceAttribute := fmt.Sprintf("%s_size_resource", dimension)

	if size, ok := autoscale[sizeAttribute].(string); ok && size != "" {
		val, err := deploymentsize.ParseGb(size)
		if err != nil {
			return err
		}
		model.Value = &val

		if model.Resource == nil {
			model.Resource = ec.String("memory")
		}
	}

	if sizeResource, ok := autoscale[resourceAttribute].(string); ok && sizeResource != "" {
		model.Resource = ec.String(sizeResource)
	}

	return nil
}

func expandEsConfig(raw []interface{}, esCfg *models.ElasticsearchConfiguration) error {
	for _, rawCfg := range raw {
		cfg, ok := rawCfg.(map[string]interface{})
		if !ok {
			continue
		}
		if settings, ok := cfg["user_settings_json"].(string); ok && settings != "" {
			if err := json.Unmarshal([]byte(settings), &esCfg.UserSettingsJSON); err != nil {
				return fmt.Errorf(
					"failed expanding elasticsearch user_settings_json: %w", err,
				)
			}
		}
		if settings, ok := cfg["user_settings_override_json"].(string); ok && settings != "" {
			if err := json.Unmarshal([]byte(settings), &esCfg.UserSettingsOverrideJSON); err != nil {
				return fmt.Errorf(
					"failed expanding elasticsearch user_settings_override_json: %w", err,
				)
			}
		}
		if settings, ok := cfg["user_settings_yaml"].(string); ok && settings != "" {
			esCfg.UserSettingsYaml = settings
		}
		if settings, ok := cfg["user_settings_override_yaml"].(string); ok && settings != "" {
			esCfg.UserSettingsOverrideYaml = settings
		}

		if v, ok := cfg["plugins"].(*schema.Set); ok && v.Len() > 0 {
			esCfg.EnabledBuiltInPlugins = util.ItemsToString(v.List())
		}

		if v, ok := cfg["docker_image"].(string); ok {
			esCfg.DockerImage = v
		}
	}

	return nil
}

func expandSnapshotSource(raw []interface{}, restore *models.RestoreSnapshotConfiguration) {
	for _, rawRestore := range raw {
		var rs, ok = rawRestore.(map[string]interface{})
		if !ok {
			continue
		}

		if clusterID, ok := rs["source_elasticsearch_cluster_id"].(string); ok {
			restore.SourceClusterID = clusterID
		}

		if snapshotName, ok := rs["snapshot_name"].(string); ok {
			restore.SnapshotName = ec.String(snapshotName)
		}
	}
}

func matchEsTopologyID(id string, topologies []*models.ElasticsearchClusterTopologyElement) (*models.ElasticsearchClusterTopologyElement, error) {
	for _, t := range topologies {
		if t.ID == id {
			return t, nil
		}
	}

	topIDs := topologyIDs(topologies)
	for i, id := range topIDs {
		topIDs[i] = "\"" + id + "\""
	}

	return nil, fmt.Errorf(`invalid id: valid topology IDs are %s`,
		strings.Join(topIDs, ", "),
	)
}

func emptyEsResource() *models.ElasticsearchPayload {
	return &models.ElasticsearchPayload{
		Plan: &models.ElasticsearchClusterPlan{
			Elasticsearch: &models.ElasticsearchConfiguration{},
		},
		Settings: &models.ElasticsearchClusterSettings{},
	}
}

// esResource returns the ElaticsearchPayload from a deployment
// template or an empty version of the payload.
func esResource(res *models.DeploymentTemplateInfoV2) *models.ElasticsearchPayload {
	if len(res.DeploymentTemplate.Resources.Elasticsearch) == 0 {
		return emptyEsResource()
	}
	return res.DeploymentTemplate.Resources.Elasticsearch[0]
}

// esResourceFromUpdate returns the ElaticsearchPayload from a deployment
// update request or an empty version of the payload.
func esResourceFromUpdate(res *models.DeploymentUpdateResources) *models.ElasticsearchPayload {
	if len(res.Elasticsearch) == 0 {
		return emptyEsResource()
	}

	return res.Elasticsearch[0]
}

func unsetElasticsearchCuration(payload *models.ElasticsearchPayload) {
	if payload.Plan.Elasticsearch != nil {
		payload.Plan.Elasticsearch.Curation = nil
	}

	if payload.Settings != nil {
		payload.Settings.Curation = nil
	}
}

func topologyIDs(topologies []*models.ElasticsearchClusterTopologyElement) []string {
	var result []string

	for _, topology := range topologies {
		result = append(result, topology.ID)
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func parseLegacyNodeType(topology map[string]interface{}, nodeType *models.ElasticsearchNodeType) error {
	if nodeType == nil {
		return nil
	}

	if ntData, ok := topology["node_type_data"].(string); ok && ntData != "" {
		nt, err := strconv.ParseBool(ntData)
		if err != nil {
			return fmt.Errorf("failed parsing node_type_data value: %w", err)
		}
		nodeType.Data = ec.Bool(nt)
	}

	if ntMaster, ok := topology["node_type_master"].(string); ok && ntMaster != "" {
		nt, err := strconv.ParseBool(ntMaster)
		if err != nil {
			return fmt.Errorf("failed parsing node_type_master value: %w", err)
		}
		nodeType.Master = ec.Bool(nt)
	}

	if ntIngest, ok := topology["node_type_ingest"].(string); ok && ntIngest != "" {
		nt, err := strconv.ParseBool(ntIngest)
		if err != nil {
			return fmt.Errorf("failed parsing node_type_ingest value: %w", err)
		}
		nodeType.Ingest = ec.Bool(nt)
	}

	if ntMl, ok := topology["node_type_ml"].(string); ok && ntMl != "" {
		nt, err := strconv.ParseBool(ntMl)
		if err != nil {
			return fmt.Errorf("failed parsing node_type_ml value: %w", err)
		}
		nodeType.Ml = ec.Bool(nt)
	}

	return nil
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

func expandEsExtension(raw []interface{}, es *models.ElasticsearchConfiguration) {
	for _, rawExt := range raw {
		m := rawExt.(map[string]interface{})

		var version string
		if v, ok := m["version"].(string); ok {
			version = v
		}

		var url string
		if u, ok := m["url"].(string); ok {
			url = u
		}

		var name string
		if n, ok := m["name"].(string); ok {
			name = n
		}

		if t, ok := m["type"].(string); ok && t == "bundle" {
			es.UserBundles = append(es.UserBundles, &models.ElasticsearchUserBundle{
				Name:                 &name,
				ElasticsearchVersion: &version,
				URL:                  &url,
			})
		}

		if t, ok := m["type"].(string); ok && t == "plugin" {
			es.UserPlugins = append(es.UserPlugins, &models.ElasticsearchUserPlugin{
				Name:                 &name,
				ElasticsearchVersion: &version,
				URL:                  &url,
			})
		}
	}
}

func expandAccountTrust(raw []interface{}, es *models.ElasticsearchClusterSettings) {
	var accounts []*models.AccountTrustRelationship
	for _, rawTrust := range raw {
		m := rawTrust.(map[string]interface{})

		var id string
		if v, ok := m["account_id"].(string); ok {
			id = v
		}

		var all bool
		if a, ok := m["trust_all"].(bool); ok {
			all = a
		}

		var allowlist []string
		if al, ok := m["trust_allowlist"].(*schema.Set); ok && al.Len() > 0 {
			allowlist = util.ItemsToString(al.List())
		}

		accounts = append(accounts, &models.AccountTrustRelationship{
			AccountID:      &id,
			TrustAll:       &all,
			TrustAllowlist: allowlist,
		})
	}

	if len(accounts) == 0 {
		return
	}

	if es.Trust == nil {
		es.Trust = &models.ElasticsearchClusterTrustSettings{}
	}

	es.Trust.Accounts = append(es.Trust.Accounts, accounts...)
}

func expandExternalTrust(raw []interface{}, es *models.ElasticsearchClusterSettings) {
	var external []*models.ExternalTrustRelationship
	for _, rawTrust := range raw {
		m := rawTrust.(map[string]interface{})

		var id string
		if v, ok := m["relationship_id"].(string); ok {
			id = v
		}

		var all bool
		if a, ok := m["trust_all"].(bool); ok {
			all = a
		}

		var allowlist []string
		if al, ok := m["trust_allowlist"].(*schema.Set); ok && al.Len() > 0 {
			allowlist = util.ItemsToString(al.List())
		}

		external = append(external, &models.ExternalTrustRelationship{
			TrustRelationshipID: &id,
			TrustAll:            &all,
			TrustAllowlist:      allowlist,
		})
	}

	if len(external) == 0 {
		return
	}

	if es.Trust == nil {
		es.Trust = &models.ElasticsearchClusterTrustSettings{}
	}

	es.Trust.External = append(es.Trust.External, external...)
}
