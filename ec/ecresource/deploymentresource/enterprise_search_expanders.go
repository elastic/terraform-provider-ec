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

// expandEssResources expands Enterprise Search resources into their models.
func expandEssResources(ess []interface{}, tpl *models.EnterpriseSearchPayload) ([]*models.EnterpriseSearchPayload, error) {
	if len(ess) == 0 {
		return nil, nil
	}

	result := make([]*models.EnterpriseSearchPayload, 0, len(ess))
	for _, raw := range ess {
		resResource, err := expandEssResource(raw, tpl)
		if err != nil {
			return nil, err
		}
		result = append(result, resResource)
	}

	return result, nil
}

func expandEssResource(raw interface{}, res *models.EnterpriseSearchPayload) (*models.EnterpriseSearchPayload, error) {
	var es = raw.(map[string]interface{})

	if esRefID, ok := es["elasticsearch_cluster_ref_id"]; ok {
		res.ElasticsearchClusterRefID = ec.String(esRefID.(string))
	}

	if refID, ok := es["ref_id"]; ok {
		res.RefID = ec.String(refID.(string))
	}

	if version, ok := es["version"]; ok {
		res.Plan.EnterpriseSearch.Version = version.(string)
	}

	if region, ok := es["region"]; ok {
		if r := region.(string); r != "" {
			res.Region = ec.String(r)
		}
	}

	if cfg, ok := es["config"]; ok {
		if c := expandEssConfig(cfg); c != nil {
			version := res.Plan.EnterpriseSearch.Version
			res.Plan.EnterpriseSearch = c
			res.Plan.EnterpriseSearch.Version = version
		}
	}

	if rt, ok := es["topology"]; ok && len(rt.([]interface{})) > 0 {
		topology, err := expandEssTopology(rt, res.Plan.ClusterTopology)
		if err != nil {
			return nil, err
		}
		res.Plan.ClusterTopology = topology
	} else {
		res.Plan.ClusterTopology = defaultEssTopology(res.Plan.ClusterTopology)
	}

	return res, nil
}

func expandEssTopology(raw interface{}, topologies []*models.EnterpriseSearchTopologyElement) ([]*models.EnterpriseSearchTopologyElement, error) {
	var rawTopologies = raw.([]interface{})
	var res = make([]*models.EnterpriseSearchTopologyElement, 0, len(rawTopologies))
	for _, rawTop := range rawTopologies {
		var topology = rawTop.(map[string]interface{})
		var icID string
		if id, ok := topology["instance_configuration_id"]; ok {
			icID = id.(string)
		}
		size, err := util.ParseTopologySize(topology)
		if err != nil {
			return nil, err
		}

		elem, err := matchEssTopology(icID, topologies)
		if err != nil {
			return nil, err
		}
		elem.Size = &size

		if id, ok := topology["instance_configuration_id"]; ok {
			elem.InstanceConfigurationID = id.(string)
		}

		if zones, ok := topology["zone_count"]; ok {
			elem.ZoneCount = int32(zones.(int))
		}

		if c, ok := topology["config"]; ok {
			elem.EnterpriseSearch = expandEssConfig(c)
		}

		res = append(res, elem)
	}

	return res, nil
}

func expandEssConfig(raw interface{}) *models.EnterpriseSearchConfiguration {
	var res = &models.EnterpriseSearchConfiguration{}
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

	if !reflect.DeepEqual(res, &models.EnterpriseSearchConfiguration{}) {
		return res
	}

	return nil
}

func defaultEssTopology(topology []*models.EnterpriseSearchTopologyElement) []*models.EnterpriseSearchTopologyElement {
	for _, t := range topology {
		if *t.Size.Value > defaultEnterpriseSearchSize || *t.Size.Value == 0 {
			t.Size.Value = ec.Int32(defaultEnterpriseSearchSize)
		}
		if t.ZoneCount > 1 {
			t.ZoneCount = 1
		}
	}

	return topology
}

func matchEssTopology(id string, topologies []*models.EnterpriseSearchTopologyElement) (*models.EnterpriseSearchTopologyElement, error) {
	for _, t := range topologies {
		if t.InstanceConfigurationID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf(
		`enterprise_search topology: invalid instance_configuration_id: "%s" doesn't match any of the deployment template instance configurations`,
		id,
	)
}
