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
	"reflect"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-ec/ec/util"
)

// ExpandResources expands Elasticsearch resources
func ExpandResources(ess []interface{}, dt string) ([]*models.ElasticsearchPayload, error) {
	if len(ess) == 0 {
		return nil, nil
	}

	result := make([]*models.ElasticsearchPayload, 0, len(ess))
	for _, raw := range ess {
		resResource, err := expandResource(raw, dt)
		if err != nil {
			return nil, err
		}
		result = append(result, resResource)
	}

	return result, nil
}

// expandResource expands a single Elasticsearch resource
func expandResource(raw interface{}, dt string) (*models.ElasticsearchPayload, error) {
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
		topology, err := ExpandTopology(rawTopology)
		if err != nil {
			return nil, err
		}
		res.Plan.ClusterTopology = topology
	}

	if cfg, ok := es["config"]; ok {
		if c := expandConfig(cfg); c != nil {
			version := res.Plan.Elasticsearch.Version
			res.Plan.Elasticsearch = c
			res.Plan.Elasticsearch.Version = version
		}
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

	return &res, nil
}

// ExpandTopology expands a flattened topology
func ExpandTopology(raw interface{}) ([]*models.ElasticsearchClusterTopologyElement, error) {
	var rawTopologies = raw.([]interface{})
	var res = make([]*models.ElasticsearchClusterTopologyElement, 0, len(rawTopologies))
	for _, rawTop := range rawTopologies {
		var topology = rawTop.(map[string]interface{})
		var nodeType = parseNodeType(topology)

		size, err := util.ParseTopologySize(topology)
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

		if c, ok := topology["config"]; ok {
			elem.Elasticsearch = expandConfig(c)
		}

		res = append(res, &elem)
	}

	return res, nil
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

func expandConfig(raw interface{}) *models.ElasticsearchConfiguration {
	var res = &models.ElasticsearchConfiguration{}
	for _, rawCfg := range raw.([]interface{}) {
		var cfg = rawCfg.(map[string]interface{})
		if settings, ok := cfg["user_settings_json"]; ok && settings != nil {
			if s, ok := settings.(string); ok && s != "" {
				res.UserSettingsJSON = settings
			}
		}
		if settings, ok := cfg["user_settings_override_json"]; ok && settings != nil {
			if s, ok := settings.(string); ok && s != "" {
				res.UserSettingsOverrideJSON = settings
			}
		}
		if settings, ok := cfg["user_settings_yaml"]; ok {
			res.UserSettingsYaml = settings.(string)
		}
		if settings, ok := cfg["user_settings_override_yaml"]; ok {
			res.UserSettingsOverrideYaml = settings.(string)
		}

		if v, ok := cfg["plugins"]; ok {
			res.EnabledBuiltInPlugins = util.ItemsToString(v.(*schema.Set).List())
		}
	}

	if !reflect.DeepEqual(res, &models.ElasticsearchConfiguration{}) {
		return res
	}

	return nil
}
