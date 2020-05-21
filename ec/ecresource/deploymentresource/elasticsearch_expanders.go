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
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deploymentsize"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
)

func expandElasticsearchResources(ess []interface{}, dt string) ([]*models.ElasticsearchPayload, error) {
	if len(ess) == 0 {
		return nil, nil
	}

	result := make([]*models.ElasticsearchPayload, 0, len(ess))
	for _, raw := range ess {
		resResource, err := expandElasticsearchResource(raw, dt)
		if err != nil {
			return nil, err
		}
		result = append(result, resResource)
	}

	return result, nil
}

func expandElasticsearchResource(raw interface{}, dt string) (*models.ElasticsearchPayload, error) {
	var es = raw.(map[string]interface{})
	var res = models.ElasticsearchPayload{
		Plan: &models.ElasticsearchClusterPlan{
			Elasticsearch: &models.ElasticsearchConfiguration{},
			DeploymentTemplate: &models.DeploymentTemplateReference{
				ID: ec.String(dt),
			},
		},
		Settings: &models.ElasticsearchClusterSettings{},
	}

	if name, ok := es["display_name"]; ok {
		res.DisplayName = name.(string)
	}

	if refID, ok := es["ref_id"]; ok {
		res.RefID = ec.String(refID.(string))
	}

	if version, ok := es["version"]; ok {
		res.Plan.Elasticsearch.Version = version.(string)
	}

	if region, ok := es["region"]; ok {
		if r := region.(string); r != "" {
			res.Region = ec.String(r)
		}
	}

	if rawTopology, ok := es["topology"]; ok {
		topology, err := expandElasticsearchTopology(rawTopology)
		if err != nil {
			return nil, err
		}
		res.Plan.ClusterTopology = topology
	}

	if rawSettings, ok := es["monitoring_settings"]; ok {
		if settings := rawSettings.([]interface{}); len(settings) > 0 {
			ms := settings[0].((map[string]interface{}))
			res.Settings.Monitoring = &models.ManagedMonitoringSettings{
				TargetClusterID: ec.String(ms["target_cluster_id"].(string)),
			}
		}
		// } else {
		// if monitoring_settings isn't present, then setting the Monitoring
		// property to the object without an ID, which should stop the monitoring.
		// FIXME: This doesn't currently work, reported in a cloud issue:
		// https://github.com/elastic/cloud/issues/57821.
		// res.Settings.Monitoring = &models.ManagedMonitoringSettings{
		// 	TargetClusterID: nil,
		// }
		// }
	}

	// TODO: Verify that this works.
	if rawSettings, ok := es["snapshot_settings"]; ok {
		if settings := rawSettings.([]interface{}); len(settings) > 0 {
			res.Settings.Snapshot = &models.ClusterSnapshotSettings{}
			var ss = settings[0].((map[string]interface{}))
			res.Settings.Snapshot.Retention = &models.ClusterSnapshotRetention{}

			if enabled, ok := ss["enabled"].(bool); ok {
				res.Settings.Snapshot.Enabled = ec.Bool(enabled)
			}

			if interval, ok := ss["interval"].(string); ok {
				res.Settings.Snapshot.Interval = interval
			}

			if maxAge, ok := ss["retention_max_age"].(string); ok {
				res.Settings.Snapshot.Retention.MaxAge = maxAge
			}

			if snapshotRetention, ok := ss["retention_snapshots"].(int); ok {
				res.Settings.Snapshot.Retention.Snapshots = int32(snapshotRetention)
			}
		}
	}

	return &res, nil
}

func expandElasticsearchTopology(raw interface{}) ([]*models.ElasticsearchClusterTopologyElement, error) {
	var rawTopologies = raw.([]interface{})
	var res = make([]*models.ElasticsearchClusterTopologyElement, 0, len(rawTopologies))
	for _, rawTop := range rawTopologies {
		var topology = rawTop.(map[string]interface{})
		var nodeType = parseNodeType(topology)

		size, err := parseTopologySize(topology)
		if err != nil {
			return nil, err
		}

		var elem = models.ElasticsearchClusterTopologyElement{
			Size:     &size,
			NodeType: &nodeType,
		}

		if id, ok := topology["instance_configuration_id"]; ok {
			elem.InstanceConfigurationID = id.(string)
		}

		if zones, ok := topology["zone_count"]; ok {
			elem.ZoneCount = int32(zones.(int))
		}

		if nodecount, ok := topology["node_count_per_zone"]; ok {
			elem.NodeCountPerZone = int32(nodecount.(int))
		}

		res = append(res, &elem)
	}

	return res, nil
}

func parseTopologySize(topology map[string]interface{}) (models.TopologySize, error) {
	if mem, ok := topology["memory_per_node"]; ok {
		val, err := deploymentsize.Parse(mem.(string))
		if err != nil {
			return models.TopologySize{}, err
		}

		return models.TopologySize{
			// TODO: For now the resource is assumed to be "memory". This can
			// and will change in the future, we need to accommodate for this case.
			Value: ec.Int32(val), Resource: ec.String("memory"),
		}, nil
	}

	return models.TopologySize{}, nil
}

func parseNodeType(topology map[string]interface{}) models.ElasticsearchNodeType {
	var result models.ElasticsearchNodeType
	if val, ok := topology["node_type_data"]; ok {
		result.Data = ec.Bool(val.(bool))
	}

	if val, ok := topology["node_type_master"]; ok {
		result.Master = ec.Bool(val.(bool))
	}

	if val, ok := topology["node_type_ingest"]; ok {
		result.Ingest = ec.Bool(val.(bool))
	}

	if val, ok := topology["node_type_ml"]; ok {
		result.Ml = ec.Bool(val.(bool))
	}

	return result
}
