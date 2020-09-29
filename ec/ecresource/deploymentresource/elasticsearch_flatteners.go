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
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

// flattenEsResources takes in Elasticsearch resource models and returns its
// flattened form.
func flattenEsResources(in []*models.ElasticsearchResourceInfo, name string) []interface{} {
	var result = make([]interface{}, 0, len(in))
	for _, res := range in {
		var m = make(map[string]interface{})
		if util.IsCurrentEsPlanEmpty(res) || isEsResourceStopped(res) {
			continue
		}

		if res.Info.ClusterID != nil && *res.Info.ClusterID != "" {
			m["resource_id"] = *res.Info.ClusterID
		}

		if res.RefID != nil && *res.RefID != "" {
			m["ref_id"] = *res.RefID
		}

		var plan = res.Info.PlanInfo.Current.Plan
		if plan.Elasticsearch != nil {
			m["version"] = plan.Elasticsearch.Version
		}

		if res.Region != nil {
			m["region"] = *res.Region
		}

		if topology := flattenEsTopology(plan); len(topology) > 0 {
			m["topology"] = topology
		}

		for k, v := range flattenEsSettings(res.Info) {
			m[k] = v
		}

		var metadata = res.Info.Metadata
		if metadata != nil && metadata.CloudID != "" {
			m["cloud_id"] = metadata.CloudID
		}

		for k, v := range flattenClusterEndpoint(res.Info.Metadata) {
			m[k] = v
		}

		if c := flattenEsConfig(plan.Elasticsearch); len(c) > 0 {
			m["config"] = c
		}

		result = append(result, m)
	}

	return result
}

func flattenEsTopology(plan *models.ElasticsearchClusterPlan) []interface{} {
	var result = make([]interface{}, 0, len(plan.ClusterTopology))
	for _, topology := range plan.ClusterTopology {
		var m = make(map[string]interface{})
		if topology.Size == nil || topology.Size.Value == nil || *topology.Size.Value == 0 {
			continue
		}

		if topology.InstanceConfigurationID != "" {
			m["instance_configuration_id"] = topology.InstanceConfigurationID
		}

		// TODO: Check legacy plans.
		// if topology.MemoryPerNode > 0 {
		// 	m["memory_per_node"] = strconv.Itoa(int(topology.MemoryPerNode))
		// }

		if *topology.Size.Resource == "memory" {
			m["memory_per_node"] = util.MemoryToState(*topology.Size.Value)
		}

		if nt := topology.NodeType; nt != nil {
			if nt.Data != nil {
				m["node_type_data"] = *nt.Data
			}

			if nt.Ingest != nil {
				m["node_type_ingest"] = *nt.Ingest
			}

			if nt.Master != nil {
				m["node_type_master"] = *nt.Master
			}

			if nt.Ml != nil {
				m["node_type_ml"] = *nt.Ml
			}
		}

		if topology.NodeCountPerZone > 0 {
			m["node_count_per_zone"] = topology.NodeCountPerZone
		}

		m["zone_count"] = topology.ZoneCount

		if c := flattenEsConfig(topology.Elasticsearch); len(c) > 0 {
			m["config"] = c
		}

		result = append(result, m)
	}

	return result
}

func flattenEsConfig(cfg *models.ElasticsearchConfiguration) []interface{} {
	var m = make(map[string]interface{})
	if cfg == nil {
		return nil
	}

	if len(cfg.EnabledBuiltInPlugins) > 0 {
		m["plugins"] = schema.NewSet(schema.HashString,
			util.StringToItems(cfg.EnabledBuiltInPlugins...),
		)
	}

	if cfg.UserSettingsYaml != "" {
		m["user_settings_yaml"] = cfg.UserSettingsYaml
	}

	if cfg.UserSettingsOverrideYaml != "" {
		m["user_settings_override_yaml"] = cfg.UserSettingsOverrideYaml
	}

	if cfg.UserSettingsJSON != nil {
		m["user_settings_json"] = cfg.UserSettingsJSON
	}

	if cfg.UserSettingsOverrideJSON != nil {
		m["user_settings_override_json"] = cfg.UserSettingsOverrideJSON
	}

	if len(m) == 0 {
		return nil
	}

	return []interface{}{m}
}

func flattenEsSettings(info *models.ElasticsearchClusterInfo) map[string]interface{} {
	// TODO Check if this is set in ECE; if not, remove entirely.
	// var validMonitoringSettings = info.Settings != nil && info.Settings.Monitoring != nil
	// validMonitoringSettings = validMonitoringSettings && info.Settings.Monitoring.TargetClusterID != nil
	// if validMonitoringSettings {
	// 	m["monitoring_settings"] = []interface{}{map[string]interface{}{
	// 		"target_cluster_id": *info.Settings.Monitoring.TargetClusterID,
	// 	}}
	// }

	var m = make(map[string]interface{})
	var monitoringInfo = info.ElasticsearchMonitoringInfo != nil
	monitoringInfo = monitoringInfo && info.ElasticsearchMonitoringInfo != nil
	if monitoringInfo && len(info.ElasticsearchMonitoringInfo.DestinationClusterIds) > 0 {
		m["monitoring_settings"] = []interface{}{map[string]interface{}{
			"target_cluster_id": info.ElasticsearchMonitoringInfo.DestinationClusterIds[0],
		}}
	}

	return m
}
