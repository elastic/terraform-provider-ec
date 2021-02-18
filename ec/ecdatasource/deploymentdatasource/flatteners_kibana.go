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

// flattenKibanaResources takes in Kibana resource models and returns its
// flattened form.
func flattenKibanaResources(in []*models.KibanaResourceInfo) []interface{} {
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

			if res.Info.ClusterID != nil {
				m["resource_id"] = *res.Info.ClusterID
			}

			if res.Info.Status != nil {
				m["status"] = *res.Info.Status
			}

			if !util.IsCurrentKibanaPlanEmpty(res) {
				var plan = res.Info.PlanInfo.Current.Plan

				if plan.Kibana != nil {
					m["version"] = plan.Kibana.Version
				}

				m["topology"] = flattenKibanaTopology(plan)
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

func flattenKibanaTopology(plan *models.KibanaClusterPlan) []interface{} {
	var result = make([]interface{}, 0, len(plan.ClusterTopology))
	for _, topology := range plan.ClusterTopology {
		var m = make(map[string]interface{})

		if isKibanaSizePopulated(topology) && *topology.Size.Value == 0 {
			continue
		}

		m["instance_configuration_id"] = topology.InstanceConfigurationID

		if isKibanaSizePopulated(topology) {
			m["size"] = util.MemoryToState(*topology.Size.Value)
			m["size_resource"] = *topology.Size.Resource
		}

		m["zone_count"] = topology.ZoneCount

		result = append(result, m)
	}

	return result
}

func isKibanaSizePopulated(topology *models.KibanaClusterTopologyElement) bool {
	if topology.Size != nil && topology.Size.Value != nil {
		return true
	}

	return false
}
