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

package kibanastate

import (
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/terraform-providers/terraform-provider-ec/ec/util"
)

// FlattenResources takes the kibana resource models and returns them flattened.
func FlattenResources(in []*models.KibanaResourceInfo, name string) []interface{} {
	var result = make([]interface{}, 0, len(in))
	for _, res := range in {
		var m = make(map[string]interface{})
		if isCurrentPlanEmpty(res) {
			continue
		}

		if res.Info.ClusterName != nil && *res.Info.ClusterName != name && *res.Info.ClusterName != "" {
			m["display_name"] = *res.Info.ClusterName
		}

		if res.RefID != nil && *res.RefID != "" {
			m["ref_id"] = *res.RefID
		}

		if res.Info.ClusterID != nil && *res.Info.ClusterID != "" {
			m["resource_id"] = *res.Info.ClusterID
		}

		var plan = res.Info.PlanInfo.Current.Plan
		if plan.Kibana != nil {
			m["version"] = plan.Kibana.Version
		}

		if res.Region != nil {
			m["region"] = *res.Region
		}

		if topology := flattenKibanaTopology(plan); len(topology) > 0 {
			m["topology"] = topology
		}

		if res.ElasticsearchClusterRefID != nil {
			m["elasticsearch_cluster_ref_id"] = *res.ElasticsearchClusterRefID
		}

		for k, v := range util.FlattenClusterEndpoint(res.Info.Metadata) {
			m[k] = v
		}

		if c := flattenConfig(plan.Kibana); len(c) > 0 {
			m["config"] = c
		}

		result = append(result, m)
	}

	return result
}

func flattenKibanaTopology(plan *models.KibanaClusterPlan) []interface{} {
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

		if topology.NodeCountPerZone > 0 {
			m["node_count_per_zone"] = topology.NodeCountPerZone
		}

		m["zone_count"] = topology.ZoneCount

		if c := flattenConfig(topology.Kibana); len(c) > 0 {
			m["config"] = c
		}

		result = append(result, m)
	}

	return result
}

func flattenConfig(cfg *models.KibanaConfiguration) []interface{} {
	var m = make(map[string]interface{})
	if cfg == nil {
		return nil
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

func isCurrentPlanEmpty(res *models.KibanaResourceInfo) bool {
	var emptyPlanInfo = res.Info == nil || res.Info.PlanInfo == nil || res.Info.PlanInfo.Current == nil
	return emptyPlanInfo || res.Info.PlanInfo.Current.Plan == nil
}
