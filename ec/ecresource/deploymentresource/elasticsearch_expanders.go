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
	"fmt"
	"reflect"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

// expandEsResources expands Elasticsearch resources
func expandEsResources(ess []interface{}, tpl *models.ElasticsearchPayload) ([]*models.ElasticsearchPayload, error) {
	if len(ess) == 0 {
		return nil, nil
	}

	result := make([]*models.ElasticsearchPayload, 0, len(ess))
	for _, raw := range ess {
		resResource, err := expandEsResource(raw, tpl)
		if err != nil {
			return nil, err
		}
		result = append(result, resResource)
	}

	return result, nil
}

// expandEsResource expands a single Elasticsearch resource
func expandEsResource(raw interface{}, res *models.ElasticsearchPayload) (*models.ElasticsearchPayload, error) {
	var es = raw.(map[string]interface{})

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

	if rt, ok := es["topology"]; ok && len(rt.([]interface{})) > 0 {
		topology, err := expandEsTopology(rt, res.Plan.ClusterTopology)
		if err != nil {
			return nil, err
		}
		res.Plan.ClusterTopology = topology
	} else {
		res.Plan.ClusterTopology = defaultEsTopology(res.Plan.ClusterTopology)
	}

	if cfg, ok := es["config"]; ok {
		if c := expandEsConfig(cfg); c != nil {
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

	return res, nil
}

// expandEsTopology expands a flattened topology
func expandEsTopology(raw interface{}, topologies []*models.ElasticsearchClusterTopologyElement) ([]*models.ElasticsearchClusterTopologyElement, error) {
	var rawTopologies = raw.([]interface{})
	var res = make([]*models.ElasticsearchClusterTopologyElement, 0)
	for i, rawTop := range rawTopologies {
		var topology = rawTop.(map[string]interface{})
		var icID string
		if id, ok := topology["instance_configuration_id"]; ok {
			icID = id.(string)
		}
		// When a topology element is set but no instance_configuration_id
		// is set, then obtain the instance_configuration_id from the topology
		// element.
		if t := defaultEsTopology(topologies); icID == "" && len(t) >= i {
			icID = t[i].InstanceConfigurationID
		}

		size, err := util.ParseTopologySize(topology)
		if err != nil {
			return nil, err
		}

		elem, err := matchEsTopology(icID, topologies)
		if err != nil {
			return nil, err
		}
		if size != nil {
			elem.Size = size
		}

		if zones, ok := topology["zone_count"]; ok {
			if z := zones.(int); z > 0 {
				elem.ZoneCount = int32(z)
			}
		}

		if nodecount, ok := topology["node_count_per_zone"]; ok {
			elem.NodeCountPerZone = int32(nodecount.(int))
		}

		if c, ok := topology["config"]; ok {
			elem.Elasticsearch = expandEsConfig(c)
		}

		res = append(res, elem)
	}

	return res, nil
}

func expandEsConfig(raw interface{}) *models.ElasticsearchConfiguration {
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

func discardEsZeroSize(topologies []*models.ElasticsearchClusterTopologyElement) (result []*models.ElasticsearchClusterTopologyElement) {
	for _, topology := range topologies {
		if topology.Size == nil || topology.Size.Value == nil || *topology.Size.Value == 0 {
			continue
		}
		result = append(result, topology)
	}
	return result
}

// defaultEsTopology iterates over all the templated topology elements and
// sets the size to the default when the template size is smaller than the
// deployment template default, the same is done on the ZoneCount. It discards
// any elements where the size is == 0, since it means that different Instance
// configurations are available to configure but are not included in the
// default deployment template.
func defaultEsTopology(topology []*models.ElasticsearchClusterTopologyElement) []*models.ElasticsearchClusterTopologyElement {
	topology = discardEsZeroSize(topology)
	for _, t := range topology {
		if *t.Size.Value < minimumElasticsearchSize {
			t.Size.Value = ec.Int32(minimumElasticsearchSize)
		}
		if t.ZoneCount < minimumZoneCount {
			t.ZoneCount = minimumZoneCount
		}
	}

	return topology
}

func matchEsTopology(id string, topologies []*models.ElasticsearchClusterTopologyElement) (*models.ElasticsearchClusterTopologyElement, error) {
	for _, t := range topologies {
		if t.InstanceConfigurationID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf(
		`elasticsearch topology: invalid instance_configuration_id: "%s" doesn't match any of the deployment template instance configurations`,
		id,
	)
}
