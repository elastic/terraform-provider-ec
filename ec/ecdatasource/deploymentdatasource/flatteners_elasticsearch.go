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
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

// flattenElasticsearchResources takes in Elasticsearch resource models and returns its
// flattened form.
func flattenElasticsearchResources(in []*models.ElasticsearchResourceInfo) ([]interface{}, error) {
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

				if plan.AutoscalingEnabled != nil {
					m["autoscale"] = strconv.FormatBool(*plan.AutoscalingEnabled)
				}

				top, err := flattenElasticsearchTopology(plan)
				if err != nil {
					return nil, err
				}
				m["topology"] = top
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

	return result, nil
}

func flattenElasticsearchTopology(plan *models.ElasticsearchClusterPlan) ([]interface{}, error) {
	var result = make([]interface{}, 0, len(plan.ClusterTopology))
	for _, topology := range plan.ClusterTopology {
		var m = make(map[string]interface{})

		if isSizePopulated(topology) && *topology.Size.Value == 0 {
			continue
		}

		m["instance_configuration_id"] = topology.InstanceConfigurationID

		if isSizePopulated(topology) {
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

		if len(topology.NodeRoles) > 0 {
			m["node_roles"] = schema.NewSet(schema.HashString, util.StringToItems(
				topology.NodeRoles...,
			))
		}

		autoscaling := make(map[string]interface{})
		if ascale := topology.AutoscalingMax; ascale != nil {
			autoscaling["max_size_resource"] = *ascale.Resource
			autoscaling["max_size"] = util.MemoryToState(*ascale.Value)
		}

		if ascale := topology.AutoscalingMin; ascale != nil {
			autoscaling["min_size_resource"] = *ascale.Resource
			autoscaling["min_size"] = util.MemoryToState(*ascale.Value)
		}

		if topology.AutoscalingPolicyOverrideJSON != nil {
			b, err := json.Marshal(topology.AutoscalingPolicyOverrideJSON)
			if err != nil {
				return nil, fmt.Errorf(
					"elasticsearch topology %s: unable to persist policy_override_json: %w",
					topology.ID, err,
				)
			}
			autoscaling["policy_override_json"] = string(b)
		}

		if len(autoscaling) > 0 {
			m["autoscaling"] = []interface{}{autoscaling}
		}

		result = append(result, m)
	}

	return result, nil
}

func isSizePopulated(topology *models.ElasticsearchClusterTopologyElement) bool {
	if topology.Size != nil && topology.Size.Value != nil {
		return true
	}

	return false
}
