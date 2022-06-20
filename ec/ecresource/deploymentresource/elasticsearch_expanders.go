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
	rolling              = "rolling"
)

// List of update strategies availables.
var strategiesList = []string{
	autodetect, growAndShrink, rollingGrowAndShrink, rolling,
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

	if refID, ok := es["ref_id"]; ok {
		res.RefID = ec.String(refID.(string))
	}

	if region, ok := es["region"]; ok {
		if r := region.(string); r != "" {
			res.Region = ec.String(r)
		}
	}

	// Unsetting the curation properties is since they're deprecated since
	// >= 6.6.0 which is when ILM is introduced in Elasticsearch.
	unsetElasticsearchCuration(res)

	if rt, ok := es["topology"]; ok && len(rt.([]interface{})) > 0 {
		topology, err := expandEsTopology(rt, res.Plan.ClusterTopology)
		if err != nil {
			return nil, err
		}
		res.Plan.ClusterTopology = topology
	}

	// Fixes the node_roles field to remove the dedicated tier roles from the
	// list when these are set as a dedicated tier as a topology element.
	updateNodeRolesOnDedicatedTiers(res.Plan.ClusterTopology)

	if cfg, ok := es["config"]; ok {
		if err := expandEsConfig(cfg, res.Plan.Elasticsearch); err != nil {
			return nil, err
		}
	}

	if snap, ok := es["snapshot_source"]; ok && len(snap.([]interface{})) > 0 {
		res.Plan.Transient = &models.TransientElasticsearchPlanConfiguration{
			RestoreSnapshot: &models.RestoreSnapshotConfiguration{},
		}
		expandSnapshotSource(snap, res.Plan.Transient.RestoreSnapshot)
	}

	if ext, ok := es["extension"]; ok {
		if e := ext.(*schema.Set); e.Len() > 0 {
			expandEsExtension(e.List(), res.Plan.Elasticsearch)
		}
	}

	if auto := es["autoscale"]; auto != nil {
		if autoscale := auto.(string); autoscale != "" {
			autoscaleBool, err := strconv.ParseBool(autoscale)
			if err != nil {
				return nil, fmt.Errorf("failed parsing autoscale value: %w", err)
			}
			res.Plan.AutoscalingEnabled = &autoscaleBool
		}
	}

	if trust, ok := es["trust_account"]; ok {
		if t := trust.(*schema.Set); t.Len() > 0 {
			if res.Settings == nil {
				res.Settings = &models.ElasticsearchClusterSettings{}
			}
			expandAccountTrust(t.List(), res.Settings)
		}
	}

	if trust, ok := es["trust_external"]; ok {
		if t := trust.(*schema.Set); t.Len() > 0 {
			if res.Settings == nil {
				res.Settings = &models.ElasticsearchClusterSettings{}
			}
			expandExternalTrust(t.List(), res.Settings)
		}
	}

	if strategy, ok := es["strategy"]; ok {
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
func expandStrategy(raw interface{}, strategy *models.PlanStrategy) (*models.PlanStrategy, error) {
	rawStrategy := raw.(map[string]interface{})
	res := strategy
	var err error = nil

	if _, ok := rawStrategy[autodetect]; ok {
		res.Autodetect = new(models.AutodetectStrategyConfig)
	} else if _, ok := rawStrategy[growAndShrink]; ok {
		res.GrowAndShrink = new(models.GrowShrinkStrategyConfig)
	} else if _, ok := rawStrategy[rollingGrowAndShrink]; ok {
		res.RollingGrowAndShrink = new(models.RollingGrowShrinkStrategyConfig)
	} else if rawValue, ok := rawStrategy[rolling]; ok {
		value := rawValue.(map[string]interface{})
		allowInlineResize := false
		skipSyncedFlush := false
		var shardInitWaitTime int64 = 600
		groupBy := "__all__"
		if v, ok := value["allowInlineResize"]; ok {
			allowInlineResize = v.(bool)
		}
		if v, ok := value["skipSyncedFlush"]; ok {
			skipSyncedFlush = v.(bool)
		}
		if v, ok := value["shardInitWaitTime"]; ok {
			shardInitWaitTime = v.(int64)
		}
		if v, ok := value["groupBy"]; ok {
			groupBy = v.(string)
		}
		res.Rolling = &models.RollingStrategyConfig{
			AllowInlineResize: &allowInlineResize,
			GroupBy:           groupBy,
			ShardInitWaitTime: shardInitWaitTime,
			SkipSyncedFlush:   &skipSyncedFlush,
		}
	} else {
		err = fmt.Errorf(`invalid strategy: valid strategies are %s`,
			strings.Join(strategiesList, ", "),
		)
	}
	return res, err
}

// expandEsTopology expands a flattened topology
func expandEsTopology(raw interface{}, topologies []*models.ElasticsearchClusterTopologyElement) ([]*models.ElasticsearchClusterTopologyElement, error) {
	rawTopologies := raw.([]interface{})
	res := topologies

	for _, rawTop := range rawTopologies {
		topology := rawTop.(map[string]interface{})

		var topologyID string
		if id, ok := topology["id"]; ok {
			topologyID = id.(string)
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

		if zones, ok := topology["zone_count"]; ok {
			if z := zones.(int); z > 0 {
				elem.ZoneCount = int32(z)
			}
		}

		if err := parseLegacyNodeType(topology, elem.NodeType); err != nil {
			return nil, err
		}

		if nr, ok := topology["node_roles"]; ok {
			if nrSet, ok := nr.(*schema.Set); ok && nrSet.Len() > 0 {
				elem.NodeRoles = util.ItemsToString(nrSet.List())
				elem.NodeType = nil
			}
		}

		if autoscalingRaw := topology["autoscaling"]; autoscalingRaw != nil {
			for _, autoscaleRaw := range autoscalingRaw.([]interface{}) {
				autoscale := autoscaleRaw.(map[string]interface{})

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

				if policy := autoscale["policy_override_json"]; policy != nil {
					if policyString := policy.(string); policyString != "" {
						if err := json.Unmarshal([]byte(policyString),
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
		}

		if cfg, ok := topology["config"]; ok {
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

	if size := autoscale[sizeAttribute]; size != nil {
		if size := size.(string); size != "" {
			val, err := deploymentsize.ParseGb(size)
			if err != nil {
				return err
			}
			model.Value = &val

			if model.Resource == nil {
				model.Resource = ec.String("memory")
			}
		}
	}

	if sizeResource := autoscale[resourceAttribute]; sizeResource != nil {
		if sizeResource := sizeResource.(string); sizeResource != "" {
			model.Resource = ec.String(sizeResource)
		}
	}

	return nil
}

func expandEsConfig(raw interface{}, esCfg *models.ElasticsearchConfiguration) error {
	for _, rawCfg := range raw.([]interface{}) {
		cfg, ok := rawCfg.(map[string]interface{})
		if !ok {
			continue
		}
		if settings, ok := cfg["user_settings_json"]; ok && settings != nil {
			if s, ok := settings.(string); ok && s != "" {
				if err := json.Unmarshal([]byte(s), &esCfg.UserSettingsJSON); err != nil {
					return fmt.Errorf(
						"failed expanding elasticsearch user_settings_json: %w", err,
					)
				}
			}
		}
		if settings, ok := cfg["user_settings_override_json"]; ok && settings != nil {
			if s, ok := settings.(string); ok && s != "" {
				if err := json.Unmarshal([]byte(s), &esCfg.UserSettingsOverrideJSON); err != nil {
					return fmt.Errorf(
						"failed expanding elasticsearch user_settings_override_json: %w", err,
					)
				}
			}
		}
		if settings, ok := cfg["user_settings_yaml"]; ok {
			esCfg.UserSettingsYaml = settings.(string)
		}
		if settings, ok := cfg["user_settings_override_yaml"]; ok {
			esCfg.UserSettingsOverrideYaml = settings.(string)
		}

		if v, ok := cfg["plugins"]; ok {
			esCfg.EnabledBuiltInPlugins = util.ItemsToString(v.(*schema.Set).List())
		}

		if v, ok := cfg["docker_image"]; ok {
			esCfg.DockerImage = v.(string)
		}
	}

	return nil
}

func expandSnapshotSource(raw interface{}, restore *models.RestoreSnapshotConfiguration) {
	for _, rawRestore := range raw.([]interface{}) {
		var rs = rawRestore.(map[string]interface{})
		if clusterID, ok := rs["source_elasticsearch_cluster_id"]; ok {
			restore.SourceClusterID = clusterID.(string)
		}

		if snapshotName, ok := rs["snapshot_name"]; ok {
			restore.SnapshotName = ec.String(snapshotName.(string))
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

// esResource returns the ElaticsearchPayload from a deployment
// template or an empty version of the payload.
func esResource(res *models.DeploymentTemplateInfoV2) *models.ElasticsearchPayload {
	if len(res.DeploymentTemplate.Resources.Elasticsearch) == 0 {
		return &models.ElasticsearchPayload{
			Plan: &models.ElasticsearchClusterPlan{
				Elasticsearch: &models.ElasticsearchConfiguration{},
			},
			Settings: &models.ElasticsearchClusterSettings{},
		}
	}
	return res.DeploymentTemplate.Resources.Elasticsearch[0]
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

	if ntData, ok := topology["node_type_data"]; ok && ntData.(string) != "" {
		nt, err := strconv.ParseBool(ntData.(string))
		if err != nil {
			return fmt.Errorf("failed parsing node_type_data value: %w", err)
		}
		nodeType.Data = ec.Bool(nt)
	}

	if ntMaster, ok := topology["node_type_master"]; ok && ntMaster.(string) != "" {
		nt, err := strconv.ParseBool(ntMaster.(string))
		if err != nil {
			return fmt.Errorf("failed parsing node_type_master value: %w", err)
		}
		nodeType.Master = ec.Bool(nt)
	}

	if ntIngest, ok := topology["node_type_ingest"]; ok && ntIngest.(string) != "" {
		nt, err := strconv.ParseBool(ntIngest.(string))
		if err != nil {
			return fmt.Errorf("failed parsing node_type_ingest value: %w", err)
		}
		nodeType.Ingest = ec.Bool(nt)
	}

	if ntMl, ok := topology["node_type_ml"]; ok && ntMl.(string) != "" {
		nt, err := strconv.ParseBool(ntMl.(string))
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
		if v, ok := m["version"]; ok {
			version = v.(string)
		}

		var url string
		if u, ok := m["url"]; ok {
			url = u.(string)
		}

		var name string
		if n, ok := m["name"]; ok {
			name = n.(string)
		}

		if t, ok := m["type"]; ok && t.(string) == "bundle" {
			es.UserBundles = append(es.UserBundles, &models.ElasticsearchUserBundle{
				Name:                 &name,
				ElasticsearchVersion: &version,
				URL:                  &url,
			})
		}

		if t, ok := m["type"]; ok && t.(string) == "plugin" {
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
		if v, ok := m["account_id"]; ok {
			id = v.(string)
		}

		var all bool
		if a, ok := m["trust_all"]; ok {
			all = a.(bool)
		}

		var allowlist []string
		if al, ok := m["trust_allowlist"]; ok {
			set := al.(*schema.Set)
			if set.Len() > 0 {
				allowlist = util.ItemsToString(set.List())
			}
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
		if v, ok := m["relationship_id"]; ok {
			id = v.(string)
		}

		var all bool
		if a, ok := m["trust_all"]; ok {
			all = a.(bool)
		}

		var allowlist []string
		if al, ok := m["trust_allowlist"]; ok {
			set := al.(*schema.Set)
			if set.Len() > 0 {
				allowlist = util.ItemsToString(set.List())
			}
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
