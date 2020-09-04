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

// FlattenElasticsearchResources takes in Elasticsearch resource models and returns its
// flattened form.
func FlattenElasticsearchResources(in []*models.ElasticsearchResourceInfo) []interface{} {
	var result = make([]interface{}, 0, len(in))
	for _, res := range in {
		var m = make(map[string]interface{})

		m["healthy"] = *res.Info.Healthy

		m["cloud_id"] = res.Info.Metadata.CloudID

		m["ref_id"] = *res.RefID

		m["resource_id"] = *res.Info.ClusterID

		var plan = res.Info.PlanInfo.Current.Plan
		m["version"] = plan.Elasticsearch.Version

		m["topology"] = flattenElasticsearchTopology(plan)

		for k, v := range util.FlattenClusterEndpoint(res.Info.Metadata) {
			m[k] = v
		}

		m["status"] = *res.Info.Status

		result = append(result, m)
	}

	return result
}

func flattenElasticsearchTopology(plan *models.ElasticsearchClusterPlan) []interface{} {
	var result = make([]interface{}, 0, len(plan.ClusterTopology))
	for _, topology := range plan.ClusterTopology {
		var m = make(map[string]interface{})

		m["instance_configuration_id"] = topology.InstanceConfigurationID

		m["memory_per_node"] = util.MemoryToState(*topology.Size.Value)

		m["zone_count"] = topology.ZoneCount

		m["node_count_per_zone"] = topology.NodeCountPerZone

		m["node_type_data"] = *topology.NodeType.Data

		m["node_type_ingest"] = *topology.NodeType.Ingest

		m["node_type_master"] = *topology.NodeType.Master

		m["node_type_ml"] = *topology.NodeType.Ml

		result = append(result, m)
	}

	return result
}
