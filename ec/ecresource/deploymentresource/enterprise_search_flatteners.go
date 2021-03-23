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
	"bytes"
	"encoding/json"

	"github.com/elastic/cloud-sdk-go/pkg/models"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

// flattenEssResources flattens Enterprise Search resources into its flattened structure.
func flattenEssResources(in []*models.EnterpriseSearchResourceInfo, name string) []interface{} {
	result := make([]interface{}, 0, len(in))
	for _, res := range in {
		m := make(map[string]interface{})
		if util.IsCurrentEssPlanEmpty(res) || isEssResourceStopped(res) {
			continue
		}

		if res.RefID != nil && *res.RefID != "" {
			m["ref_id"] = *res.RefID
		}

		if res.Info.ID != nil && *res.Info.ID != "" {
			m["resource_id"] = *res.Info.ID
		}

		if res.Region != nil {
			m["region"] = *res.Region
		}

		plan := res.Info.PlanInfo.Current.Plan
		if topology := flattenEssTopology(plan); len(topology) > 0 {
			m["topology"] = topology
		}

		if res.ElasticsearchClusterRefID != nil {
			m["elasticsearch_cluster_ref_id"] = *res.ElasticsearchClusterRefID
		}

		if urls := util.FlattenClusterEndpoint(res.Info.Metadata); len(urls) > 0 {
			for k, v := range urls {
				m[k] = v
			}
		}

		if c := flattenEssConfig(plan.EnterpriseSearch); len(c) > 0 {
			m["config"] = c
		}

		result = append(result, m)
	}

	return result
}

func flattenEssTopology(plan *models.EnterpriseSearchPlan) []interface{} {
	var result = make([]interface{}, 0, len(plan.ClusterTopology))
	for _, topology := range plan.ClusterTopology {
		var m = make(map[string]interface{})
		if topology.Size == nil || topology.Size.Value == nil || *topology.Size.Value == 0 {
			continue
		}

		if topology.InstanceConfigurationID != "" {
			m["instance_configuration_id"] = topology.InstanceConfigurationID
		}

		if topology.Size != nil {
			m["size"] = util.MemoryToState(*topology.Size.Value)
			m["size_resource"] = *topology.Size.Resource
		}

		if nt := topology.NodeType; nt != nil {
			if nt.Appserver != nil {
				m["node_type_appserver"] = *nt.Appserver
			}

			if nt.Connector != nil {
				m["node_type_connector"] = *nt.Connector
			}

			if nt.Worker != nil {
				m["node_type_worker"] = *nt.Worker
			}
		}

		m["zone_count"] = topology.ZoneCount

		result = append(result, m)
	}

	return result
}

func flattenEssConfig(cfg *models.EnterpriseSearchConfiguration) []interface{} {
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

	if o := cfg.UserSettingsJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			m["user_settings_json"] = string(b)
		}
	}

	if o := cfg.UserSettingsOverrideJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			m["user_settings_override_json"] = string(b)
		}
	}

	if len(m) == 0 {
		return nil
	}

	return []interface{}{m}
}
