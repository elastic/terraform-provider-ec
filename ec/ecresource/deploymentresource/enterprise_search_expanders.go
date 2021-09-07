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

// expandEssResources expands Enterprise Search resources into their models.
func expandEssResources(ess []interface{}, tpl *models.EnterpriseSearchPayload) ([]*models.EnterpriseSearchPayload, error) {
	if len(ess) == 0 {
		return nil, nil
	}

	if tpl == nil {
		return nil, errors.New("enterprise_search specified but deployment template is not configured for it. Use a different template if you wish to add enterprise_search")
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
	ess := raw.(map[string]interface{})

	if esRefID, ok := ess["elasticsearch_cluster_ref_id"]; ok {
		res.ElasticsearchClusterRefID = ec.String(esRefID.(string))
	}

	if refID, ok := ess["ref_id"]; ok {
		res.RefID = ec.String(refID.(string))
	}

	if version, ok := ess["version"]; ok {
		res.Plan.EnterpriseSearch.Version = version.(string)
	}

	if region, ok := ess["region"]; ok {
		if r := region.(string); r != "" {
			res.Region = ec.String(r)
		}
	}

	if cfg, ok := ess["config"]; ok {
		if err := expandEssConfig(cfg, res.Plan.EnterpriseSearch); err != nil {
			return nil, err
		}
	}

	if rt, ok := ess["topology"]; ok && len(rt.([]interface{})) > 0 {
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
	rawTopologies := raw.([]interface{})
	res := make([]*models.EnterpriseSearchTopologyElement, 0, len(rawTopologies))
	for i, rawTop := range rawTopologies {
		topology := rawTop.(map[string]interface{})
		var icID string
		if id, ok := topology["instance_configuration_id"]; ok {
			icID = id.(string)
		}

		// When a topology element is set but no instance_configuration_id
		// is set, then obtain the instance_configuration_id from the topology
		// element.
		if t := defaultEssTopology(topologies); icID == "" && len(t) >= i {
			icID = t[i].InstanceConfigurationID
		}
		size, err := util.ParseTopologySize(topology)
		if err != nil {
			return nil, err
		}

		// Since Enterprise Search is not enabled by default in the template,
		// if the size == nil, it means that the size hasn't been specified in
		// the definition.
		if size == nil {
			size = &models.TopologySize{
				Resource: ec.String("memory"),
				Value:    ec.Int32(minimumEnterpriseSearchSize),
			}
		}

		elem, err := matchEssTopology(icID, topologies)
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

func expandEssConfig(raw interface{}, res *models.EnterpriseSearchConfiguration) error {
	for _, rawCfg := range raw.([]interface{}) {
		cfg := rawCfg.(map[string]interface{})
		if settings, ok := cfg["user_settings_json"]; ok && settings != nil {
			if s, ok := settings.(string); ok && s != "" {
				if err := json.Unmarshal([]byte(s), &res.UserSettingsJSON); err != nil {
					return fmt.Errorf("failed expanding enterprise_search user_settings_json: %w", err)
				}
			}
		}
		if settings, ok := cfg["user_settings_override_json"]; ok && settings != nil {
			if s, ok := settings.(string); ok && s != "" {
				if err := json.Unmarshal([]byte(s), &res.UserSettingsOverrideJSON); err != nil {
					return fmt.Errorf("failed expanding enterprise_search user_settings_override_json: %w", err)
				}
			}
		}
		if settings, ok := cfg["user_settings_yaml"]; ok {
			res.UserSettingsYaml = settings.(string)
		}
		if settings, ok := cfg["user_settings_override_yaml"]; ok {
			res.UserSettingsOverrideYaml = settings.(string)
		}

		if v, ok := cfg["docker_image"]; ok {
			res.DockerImage = v.(string)
		}
	}

	return nil
}

// defaultApmTopology iterates over all the templated topology elements and
// sets the size to the default when the template size is smaller than the
// deployment template default, the same is done on the ZoneCount.
func defaultEssTopology(topology []*models.EnterpriseSearchTopologyElement) []*models.EnterpriseSearchTopologyElement {
	for _, t := range topology {
		if *t.Size.Value < minimumEnterpriseSearchSize || *t.Size.Value == 0 {
			t.Size.Value = ec.Int32(minimumEnterpriseSearchSize)
		}
		if t.ZoneCount < minimumZoneCount {
			t.ZoneCount = minimumZoneCount
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

// essResource returns the EnterpriseSearchPayload from a deployment
// template or an empty version of the payload.
func essResource(res *models.DeploymentTemplateInfoV2) *models.EnterpriseSearchPayload {
	if len(res.DeploymentTemplate.Resources.EnterpriseSearch) == 0 {
		return nil
	}
	return res.DeploymentTemplate.Resources.EnterpriseSearch[0]
}
