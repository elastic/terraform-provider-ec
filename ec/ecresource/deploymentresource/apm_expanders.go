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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

// expandApmResources expands apm resources into their models.
func expandApmResources(apms []interface{}, tpl *models.ApmPayload) ([]*models.ApmPayload, error) {
	if len(apms) == 0 {
		return nil, nil
	}

	if tpl == nil {
		return nil, errors.New("apm specified but deployment template is not configured for it. Use a different template if you wish to add apm")
	}

	result := make([]*models.ApmPayload, 0, len(apms))
	for _, raw := range apms {
		resResource, err := expandApmResource(raw, tpl)
		if err != nil {
			return nil, err
		}
		result = append(result, resResource)
	}

	return result, nil
}

func expandApmResource(raw interface{}, res *models.ApmPayload) (*models.ApmPayload, error) {
	var apm = raw.(map[string]interface{})

	if esRefID, ok := apm["elasticsearch_cluster_ref_id"]; ok {
		res.ElasticsearchClusterRefID = ec.String(esRefID.(string))
	}

	if refID, ok := apm["ref_id"]; ok {
		res.RefID = ec.String(refID.(string))
	}

	if region, ok := apm["region"]; ok {
		if r := region.(string); r != "" {
			res.Region = ec.String(r)
		}
	}

	if cfg, ok := apm["config"]; ok {
		if err := expandApmConfig(cfg, res.Plan.Apm); err != nil {
			return nil, err
		}
	}

	if rt, ok := apm["topology"]; ok && len(rt.([]interface{})) > 0 {
		topology, err := expandApmTopology(rt, res.Plan.ClusterTopology)
		if err != nil {
			return nil, err
		}
		res.Plan.ClusterTopology = topology
	} else {
		res.Plan.ClusterTopology = defaultApmTopology(res.Plan.ClusterTopology)
	}

	return res, nil
}

func expandApmTopology(raw interface{}, topologies []*models.ApmTopologyElement) ([]*models.ApmTopologyElement, error) {
	rawTopologies := raw.([]interface{})
	res := make([]*models.ApmTopologyElement, 0, len(rawTopologies))

	for i, rawTop := range rawTopologies {
		topology := rawTop.(map[string]interface{})
		var icID string
		if id, ok := topology["instance_configuration_id"]; ok {
			icID = id.(string)
		}
		// When a topology element is set but no instance_configuration_id
		// is set, then obtain the instance_configuration_id from the topology
		// element.
		if t := defaultApmTopology(topologies); icID == "" && len(t) >= i {
			icID = t[i].InstanceConfigurationID
		}

		size, err := util.ParseTopologySize(topology)
		if err != nil {
			return nil, err
		}

		elem, err := matchApmTopology(icID, topologies)
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

		res = append(res, elem)
	}

	return res, nil
}

func expandApmConfig(raw interface{}, res *models.ApmConfiguration) error {
	for _, rawCfg := range raw.([]interface{}) {
		var cfg = rawCfg.(map[string]interface{})

		if debugEnabled, ok := cfg["debug_enabled"]; ok {
			if res.SystemSettings == nil {
				res.SystemSettings = &models.ApmSystemSettings{}
			}
			res.SystemSettings.DebugEnabled = ec.Bool(debugEnabled.(bool))
		}

		if settings, ok := cfg["user_settings_json"]; ok && settings != nil {
			if s, ok := settings.(string); ok && s != "" {
				if err := json.Unmarshal([]byte(s), &res.UserSettingsJSON); err != nil {
					return fmt.Errorf("failed expanding apm user_settings_json: %w", err)
				}
			}
		}
		if settings, ok := cfg["user_settings_override_json"]; ok && settings != nil {
			if s, ok := settings.(string); ok && s != "" {
				if err := json.Unmarshal([]byte(s), &res.UserSettingsOverrideJSON); err != nil {
					return fmt.Errorf("failed expanding apm user_settings_override_json: %w", err)
				}
			}
		}
		if settings, ok := cfg["user_settings_yaml"]; ok {
			res.UserSettingsYaml = settings.(string)
		}
		if settings, ok := cfg["user_settings_override_yaml"]; ok {
			res.UserSettingsOverrideYaml = settings.(string)
		}
	}

	return nil
}

// defaultApmTopology iterates over all the templated topology elements and
// sets the size to the default when the template size is smaller than the
// deployment template default, the same is done on the ZoneCount.
func defaultApmTopology(topology []*models.ApmTopologyElement) []*models.ApmTopologyElement {
	for _, t := range topology {
		if *t.Size.Value < minimumApmSize {
			t.Size.Value = ec.Int32(minimumApmSize)
		}
		if t.ZoneCount < minimumZoneCount {
			t.ZoneCount = minimumZoneCount
		}
	}

	return topology
}

func matchApmTopology(id string, topologies []*models.ApmTopologyElement) (*models.ApmTopologyElement, error) {
	for _, t := range topologies {
		if t.InstanceConfigurationID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf(
		`apm topology: invalid instance_configuration_id: "%s" doesn't match any of the deployment template instance configurations`,
		id,
	)
}

// apmResource returns the ApmPayload from a deployment
// template or an empty version of the payload.
func apmResource(res *models.DeploymentTemplateInfoV2) *models.ApmPayload {
	if len(res.DeploymentTemplate.Resources.Apm) == 0 {
		return nil
	}
	return res.DeploymentTemplate.Resources.Apm[0]
}
