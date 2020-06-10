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

package apmstate

import (
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/terraform-providers/terraform-provider-ec/ec/ecresource/deploymentresource/deploymentstate"
)

// FlattenResources flattens apm resources into its flattened structure.
func FlattenResources(in []*models.ApmResourceInfo, name string) []interface{} {
	var result = make([]interface{}, 0, len(in))
	for _, res := range in {
		var m = make(map[string]interface{})
		if isCurrentPlanEmpty(res) {
			continue
		}

		if res.Info.Name != nil && *res.Info.Name != name && *res.Info.Name != "" {
			m["display_name"] = *res.Info.Name
		}

		if res.RefID != nil && *res.RefID != "" {
			m["ref_id"] = *res.RefID
		}

		if res.Info.ID != nil && *res.Info.ID != "" {
			m["resource_id"] = *res.Info.ID
		}

		var plan = res.Info.PlanInfo.Current.Plan
		if plan.Apm != nil {
			m["version"] = plan.Apm.Version
		}

		if res.Region != nil {
			m["region"] = *res.Region
		}

		if topology := flattenTopology(plan); len(topology) > 0 {
			m["topology"] = topology
		}

		if res.ElasticsearchClusterRefID != nil {
			m["elasticsearch_cluster_ref_id"] = *res.ElasticsearchClusterRefID
		}

		if endpoint := deploymentstate.FlattenClusterEndpoint(res.Info.Metadata); len(endpoint) > 0 {
			m["endpoint"] = endpoint
		}

		result = append(result, m)
	}

	return result
}

func flattenTopology(plan *models.ApmPlan) []interface{} {
	var result = make([]interface{}, 0, len(plan.ClusterTopology))
	for _, topology := range plan.ClusterTopology {
		var m = make(map[string]interface{})
		if topology.Size == nil || topology.Size.Value == nil || *topology.Size.Value == 0 {
			continue
		}

		if topology.InstanceConfigurationID != "" {
			m["instance_configuration_id"] = topology.InstanceConfigurationID
		}

		if *topology.Size.Resource == "memory" {
			m["memory_per_node"] = deploymentstate.MemoryToState(*topology.Size.Value)
		}

		m["zone_count"] = topology.ZoneCount

		result = append(result, m)
	}

	return result
}

func isCurrentPlanEmpty(res *models.ApmResourceInfo) bool {
	var emptyPlanInfo = res.Info == nil || res.Info.PlanInfo == nil || res.Info.PlanInfo.Current == nil
	return emptyPlanInfo || res.Info.PlanInfo.Current.Plan == nil
}
