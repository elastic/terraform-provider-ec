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

package deploymentdatasource

import (
	"github.com/elastic/cloud-sdk-go/pkg/models"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

// flattenElasticsearchResources takes in Elasticsearch resource models and returns its
// flattened form.
func flattenElasticsearchResources(in []*models.ElasticsearchResourceInfo) []interface{} {
	var result = make([]interface{}, 0, len(in))
	for _, res := range in {
		var m = make(map[string]interface{})

		if res.RefID != nil {
			m["ref_id"] = *res.RefID
		}

		if res.Info != nil {
			if res.Info.Healthy != nil {
				m["healthy"] = *res.Info.Healthy
			}

			if res.Info.ClusterID != nil {
				m["resource_id"] = *res.Info.ClusterID
			}

			if res.Info.Status != nil {
				m["status"] = *res.Info.Status
			}

			if !util.IsCurrentEsPlanEmpty(res) {
				var plan = res.Info.PlanInfo.Current.Plan

				if plan.Elasticsearch != nil {
					m["version"] = plan.Elasticsearch.Version
				}

				m["topology"] = flattenElasticsearchTopology(plan)
			}

			if res.Info.Metadata != nil {
				m["cloud_id"] = res.Info.Metadata.CloudID

				for k, v := range util.FlattenClusterEndpoint(res.Info.Metadata) {
					m[k] = v
				}
			}
		}
		result = append(result, m)
	}

	return result
}

func flattenElasticsearchTopology(plan *models.ElasticsearchClusterPlan) []interface{} {
	var result = make([]interface{}, 0, len(plan.ClusterTopology))
	for _, topology := range plan.ClusterTopology {
		var m = make(map[string]interface{})

		m["instance_configuration_id"] = topology.InstanceConfigurationID

		if topology.Size != nil && topology.Size.Value != nil {
			m["size"] = util.MemoryToState(*topology.Size.Value)
			m["size_resource"] = *topology.Size.Resource
		}

		m["zone_count"] = topology.ZoneCount

		if topology.NodeType != nil {
			if topology.NodeType.Data != nil {
				m["node_type_data"] = *topology.NodeType.Data
			}

			if topology.NodeType.Ingest != nil {
				m["node_type_ingest"] = *topology.NodeType.Ingest
			}

			if topology.NodeType.Master != nil {
				m["node_type_master"] = *topology.NodeType.Master
			}

			if topology.NodeType.Ml != nil {
				m["node_type_ml"] = *topology.NodeType.Ml
			}
		}

		result = append(result, m)
	}

	return result
}
