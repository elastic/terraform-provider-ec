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

package state

import (
	"github.com/elastic/cloud-sdk-go/pkg/models"

	"github.com/terraform-providers/terraform-provider-ec/ec/util"
)

// FlattenEnterpriseSearchResources takes in EnterpriseSearch resource models and returns its
// flattened form.
func FlattenEnterpriseSearchResources(in []*models.EnterpriseSearchResourceInfo) []interface{} {
	var result = make([]interface{}, 0, len(in))
	for _, res := range in {
		var m = make(map[string]interface{})

		if res.RefID != nil {
			m["ref_id"] = *res.RefID
		}

		if res.ElasticsearchClusterRefID != nil {
			m["elasticsearch_cluster_ref_id"] = *res.ElasticsearchClusterRefID
		}

		if res.Info != nil {
			if res.Info.Healthy != nil {
				m["healthy"] = *res.Info.Healthy
			}

			if res.Info.ID != nil {
				m["resource_id"] = *res.Info.ID
			}

			if res.Info.Status != nil {
				m["status"] = *res.Info.Status
			}

			if res.Info.PlanInfo != nil && res.Info.PlanInfo.Current != nil &&
				res.Info.PlanInfo.Current.Plan != nil {
				var plan = res.Info.PlanInfo.Current.Plan

				if plan.EnterpriseSearch != nil {
					m["version"] = plan.EnterpriseSearch.Version
				}

				if plan.EnterpriseSearch != nil &&
					plan.EnterpriseSearch.SystemSettings != nil {
					m["secret_session_key"] = plan.EnterpriseSearch.SystemSettings.SecretSessionKey
				}

				m["topology"] = flattenEnterpriseSearchTopology(plan)
			}

			if res.Info.Metadata != nil {
				for k, v := range util.FlattenClusterEndpoint(res.Info.Metadata) {
					m[k] = v
				}
			}
		}
		result = append(result, m)
	}

	return result
}

func flattenEnterpriseSearchTopology(plan *models.EnterpriseSearchPlan) []interface{} {
	var result = make([]interface{}, 0, len(plan.ClusterTopology))
	for _, topology := range plan.ClusterTopology {
		var m = make(map[string]interface{})

		m["instance_configuration_id"] = topology.InstanceConfigurationID

		m["zone_count"] = topology.ZoneCount

		if topology.Size != nil && topology.Size.Value != nil {
			m["memory_per_node"] = util.MemoryToState(*topology.Size.Value)
		}

		if topology.NodeType != nil {
			if topology.NodeType.Appserver != nil {
				m["node_type_appserver"] = *topology.NodeType.Appserver
			}

			if topology.NodeType.Connector != nil {
				m["node_type_connector"] = *topology.NodeType.Connector
			}

			if topology.NodeType.Worker != nil {
				m["node_type_worker"] = *topology.NodeType.Worker
			}
		}

		result = append(result, m)
	}

	return result
}
