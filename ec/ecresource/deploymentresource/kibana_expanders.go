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

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

// expandKibanaResources expands the flattened kibana resources into its models.
func expandKibanaResources(kibanas []interface{}, tpl *models.KibanaPayload) ([]*models.KibanaPayload, error) {
	if len(kibanas) == 0 {
		return nil, nil
	}

	result := make([]*models.KibanaPayload, 0, len(kibanas))
	for _, raw := range kibanas {
		resResource, err := expandKibanaResource(raw, tpl)
		if err != nil {
			return nil, err
		}
		result = append(result, resResource)
	}

	return result, nil
}

func expandKibanaResource(raw interface{}, res *models.KibanaPayload) (*models.KibanaPayload, error) {
	var es = raw.(map[string]interface{})

	if esRefID, ok := es["elasticsearch_cluster_ref_id"]; ok {
		res.ElasticsearchClusterRefID = ec.String(esRefID.(string))
	}

	if refID, ok := es["ref_id"]; ok {
		res.RefID = ec.String(refID.(string))
	}

	if version, ok := es["version"]; ok {
		res.Plan.Kibana.Version = version.(string)
	}

	if region, ok := es["region"]; ok {
		if r := region.(string); r != "" {
			res.Region = ec.String(r)
		}
	}

	if cfg, ok := es["config"]; ok {
		if c := expandKibanaConfig(cfg); c != nil {
			version := res.Plan.Kibana.Version
			res.Plan.Kibana = c
			res.Plan.Kibana.Version = version
		}
	}

	if rt, ok := es["topology"]; ok && len(rt.([]interface{})) > 0 {
		topology, err := expandKibanaTopology(rt, res.Plan.ClusterTopology)
		if err != nil {
			return nil, err
		}
		res.Plan.ClusterTopology = topology
	} else {
		res.Plan.ClusterTopology = defaultKibanaTopology(res.Plan.ClusterTopology)
	}

	return res, nil
}

func expandKibanaTopology(raw interface{}, topologies []*models.KibanaClusterTopologyElement) ([]*models.KibanaClusterTopologyElement, error) {
	var rawTopologies = raw.([]interface{})
	var res = make([]*models.KibanaClusterTopologyElement, 0, len(rawTopologies))
	for i, rawTop := range rawTopologies {
		var topology = rawTop.(map[string]interface{})
		var icID string
		if id, ok := topology["instance_configuration_id"]; ok {
			icID = id.(string)
		}
		// When a topology element is set but no instance_configuration_id
		// is set, then obtain the instance_configuration_id from the topology
		// element.
		if t := defaultKibanaTopology(topologies); icID == "" && len(t) >= i {
			icID = t[i].InstanceConfigurationID
		}
		size, err := util.ParseTopologySize(topology)
		if err != nil {
			return nil, err
		}

		elem, err := matchKibanaTopology(icID, topologies)
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

		if c, ok := topology["config"]; ok {
			elem.Kibana = expandKibanaConfig(c)
		}

		res = append(res, elem)
	}

	return res, nil
}

func expandKibanaConfig(raw interface{}) *models.KibanaConfiguration {
	var res = &models.KibanaConfiguration{}
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
	}

	if !reflect.DeepEqual(res, &models.KibanaConfiguration{}) {
		return res
	}

	return nil
}

// defaultApmTopology iterates over all the templated topology elements and
// sets the size to the default when the template size is greater than the
// local terraform default, the same is done on the ZoneCount.
func defaultKibanaTopology(topology []*models.KibanaClusterTopologyElement) []*models.KibanaClusterTopologyElement {
	for _, t := range topology {
		if *t.Size.Value > minimumKibanaSize {
			t.Size.Value = ec.Int32(minimumKibanaSize)
		}
		if t.ZoneCount > minimumZoneCount {
			t.ZoneCount = minimumZoneCount
		}
	}

	return topology
}

func matchKibanaTopology(id string, topologies []*models.KibanaClusterTopologyElement) (*models.KibanaClusterTopologyElement, error) {
	for _, t := range topologies {
		if t.InstanceConfigurationID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf(
		`kibana topology: invalid instance_configuration_id: "%s" doesn't match any of the deployment template instance configurations`,
		id,
	)
}
