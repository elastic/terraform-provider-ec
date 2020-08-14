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

package elasticsearchstate

import (
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/terraform-providers/terraform-provider-ec/ec/ecresource/deploymentresource/deploymentstate"
)

// FlattenResources takes in Elasticsearch resource models and returns its
// flattened form.
func FlattenResources(in []*models.ElasticsearchResourceInfo, name string) []interface{} {
	var result = make([]interface{}, 0, len(in))
	for _, res := range in {
		var m = make(map[string]interface{})
		if IsCurrentPlanEmpty(res) {
			continue
		}

		if res.Info.ClusterName != nil && *res.Info.ClusterName != name && *res.Info.ClusterName != "" {
			m["display_name"] = *res.Info.ClusterName
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

		if topology := flattenTopology(plan); len(topology) > 0 {
			m["topology"] = topology
		}

		for k, v := range flattenElasticsearchSettings(res.Info) {
			m[k] = v
		}

		var metadata = res.Info.Metadata
		if metadata != nil && metadata.CloudID != "" {
			m["cloud_id"] = metadata.CloudID
		}

		for k, v := range deploymentstate.FlattenClusterEndpoint(res.Info.Metadata) {
			m[k] = v
		}

		// TODO: Flatten repository state.
		// Determine what to do with the default snapshot repository.

		result = append(result, m)
	}

	return result
}

func flattenTopology(plan *models.ElasticsearchClusterPlan) []interface{} {
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
			m["memory_per_node"] = deploymentstate.MemoryToState(*topology.Size.Value)
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

		result = append(result, m)
	}

	return result
}

func flattenElasticsearchSettings(info *models.ElasticsearchClusterInfo) map[string]interface{} {
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

// IsCurrentPlanEmpty checks the elasticsearch resource current plan is empty.
func IsCurrentPlanEmpty(res *models.ElasticsearchResourceInfo) bool {
	return res.Info == nil || res.Info.PlanInfo == nil ||
		res.Info.PlanInfo.Current == nil ||
		res.Info.PlanInfo.Current.Plan == nil
}
