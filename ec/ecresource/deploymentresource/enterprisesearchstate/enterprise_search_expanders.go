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

package enterprisesearchstate

import (
	"reflect"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"

	"github.com/terraform-providers/terraform-provider-ec/ec/util"
)

// ExpandResources expands appsearch resources into their models.
func ExpandResources(ess []interface{}) ([]*models.EnterpriseSearchPayload, error) {
	if len(ess) == 0 {
		return nil, nil
	}

	result := make([]*models.EnterpriseSearchPayload, 0, len(ess))
	for _, raw := range ess {
		resResource, err := expandResource(raw)
		if err != nil {
			return nil, err
		}
		result = append(result, resResource)
	}

	return result, nil
}

func expandResource(raw interface{}) (*models.EnterpriseSearchPayload, error) {
	var es = raw.(map[string]interface{})
	var res = models.EnterpriseSearchPayload{
		Plan: &models.EnterpriseSearchPlan{
			EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
		},
		Settings: &models.EnterpriseSearchSettings{},
	}

	if esRefID, ok := es["elasticsearch_cluster_ref_id"]; ok {
		res.ElasticsearchClusterRefID = ec.String(esRefID.(string))
	}

	if name, ok := es["display_name"]; ok {
		res.DisplayName = name.(string)
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
		if c := expandConfig(cfg); c != nil {
			version := res.Plan.EnterpriseSearch.Version
			res.Plan.EnterpriseSearch = c
			res.Plan.EnterpriseSearch.Version = version
		}
	}

	if rawTopology, ok := es["topology"]; ok {
		topology, err := expandTopology(rawTopology)
		if err != nil {
			return nil, err
		}
		res.Plan.ClusterTopology = topology
	}

	return &res, nil
}

func expandTopology(raw interface{}) ([]*models.EnterpriseSearchTopologyElement, error) {
	var rawTopologies = raw.([]interface{})
	var res = make([]*models.EnterpriseSearchTopologyElement, 0, len(rawTopologies))
	for _, rawTop := range rawTopologies {
		var topology = rawTop.(map[string]interface{})
		var nodeType = parseNodeType(topology)

		size, err := util.ParseTopologySize(topology)
		if err != nil {
			return nil, err
		}

		var elem = models.EnterpriseSearchTopologyElement{
			Size:     &size,
			NodeType: &nodeType,
		}

		if id, ok := topology["instance_configuration_id"]; ok {
			elem.InstanceConfigurationID = id.(string)
		}

		if zones, ok := topology["zone_count"]; ok {
			elem.ZoneCount = int32(zones.(int))
		}

		if c, ok := topology["config"]; ok {
			elem.EnterpriseSearch = expandConfig(c)
		}

		res = append(res, &elem)
	}

	return res, nil
}

func parseNodeType(topology map[string]interface{}) models.EnterpriseSearchNodeTypes {
	var result models.EnterpriseSearchNodeTypes
	if val, ok := topology["node_type_appserver"]; ok {
		result.Appserver = ec.Bool(val.(bool))
	}

	if val, ok := topology["node_type_connector"]; ok {
		result.Connector = ec.Bool(val.(bool))
	}

	if val, ok := topology["node_type_worker"]; ok {
		result.Worker = ec.Bool(val.(bool))
	}

	return result
}

func expandConfig(raw interface{}) *models.EnterpriseSearchConfiguration {
	var res = &models.EnterpriseSearchConfiguration{}
	for _, rawCfg := range raw.([]interface{}) {
		var cfg = rawCfg.(map[string]interface{})
		if key, ok := cfg["secret_session_key"]; ok {
			if res.SystemSettings == nil {
				res.SystemSettings = &models.EnterpriseSearchSystemSettings{}
			}
			res.SystemSettings.SecretSessionKey = key.(string)
		}

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
