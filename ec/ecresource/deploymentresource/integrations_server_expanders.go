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

// expandIntegrationsServerResources expands IntegrationsServer resources into their models.
func expandIntegrationsServerResources(integrationsServers []interface{}, tpl *models.IntegrationsServerPayload) ([]*models.IntegrationsServerPayload, error) {
	if len(integrationsServers) == 0 {
		return nil, nil
	}

	if tpl == nil {
		return nil, errors.New("IntegrationsServer specified but deployment template is not configured for it. Use a different template if you wish to add IntegrationsServer")
	}

	result := make([]*models.IntegrationsServerPayload, 0, len(integrationsServers))
	for _, raw := range integrationsServers {
		resResource, err := expandIntegrationsServerResource(raw, tpl)
		if err != nil {
			return nil, err
		}
		result = append(result, resResource)
	}

	return result, nil
}

func expandIntegrationsServerResource(raw interface{}, res *models.IntegrationsServerPayload) (*models.IntegrationsServerPayload, error) {
	var integrationsServer = raw.(map[string]interface{})

	if esRefID, ok := integrationsServer["elasticsearch_cluster_ref_id"].(string); ok {
		res.ElasticsearchClusterRefID = ec.String(esRefID)
	}

	if refID, ok := integrationsServer["ref_id"].(string); ok {
		res.RefID = ec.String(refID)
	}

	if region, ok := integrationsServer["region"].(string); ok && region != "" {
		res.Region = ec.String(region)
	}

	if cfg, ok := integrationsServer["config"].([]interface{}); ok {
		if err := expandIntegrationsServerConfig(cfg, res.Plan.IntegrationsServer); err != nil {
			return nil, err
		}
	}

	if rt, ok := integrationsServer["topology"].([]interface{}); ok && len(rt) > 0 {
		topology, err := expandIntegrationsServerTopology(rt, res.Plan.ClusterTopology)
		if err != nil {
			return nil, err
		}
		res.Plan.ClusterTopology = topology
	} else {
		res.Plan.ClusterTopology = defaultIntegrationsServerTopology(res.Plan.ClusterTopology)
	}

	return res, nil
}

func expandIntegrationsServerTopology(rawTopologies []interface{}, topologies []*models.IntegrationsServerTopologyElement) ([]*models.IntegrationsServerTopologyElement, error) {
	res := make([]*models.IntegrationsServerTopologyElement, 0, len(rawTopologies))

	for i, rawTop := range rawTopologies {
		topology, ok := rawTop.(map[string]interface{})
		if !ok {
			continue
		}

		var icID string
		if id, ok := topology["instance_configuration_id"].(string); ok {
			icID = id
		}
		// When a topology element is set but no instance_configuration_id
		// is set, then obtain the instance_configuration_id from the topology
		// element.
		if t := defaultIntegrationsServerTopology(topologies); icID == "" && len(t) > i {
			icID = t[i].InstanceConfigurationID
		}

		size, err := util.ParseTopologySize(topology)
		if err != nil {
			return nil, err
		}

		elem, err := matchIntegrationsServerTopology(icID, topologies)
		if err != nil {
			return nil, err
		}
		if size != nil {
			elem.Size = size
		}

		if zones, ok := topology["zone_count"].(int); ok && zones > 0 {
			elem.ZoneCount = int32(zones)
		}

		res = append(res, elem)
	}

	return res, nil
}

func expandIntegrationsServerConfig(raw []interface{}, res *models.IntegrationsServerConfiguration) error {
	for _, rawCfg := range raw {
		cfg, ok := rawCfg.(map[string]interface{})
		if !ok {
			continue
		}

		if debugEnabled, ok := cfg["debug_enabled"].(bool); ok {
			if res.SystemSettings == nil {
				res.SystemSettings = &models.IntegrationsServerSystemSettings{}
			}
			res.SystemSettings.DebugEnabled = ec.Bool(debugEnabled)
		}

		if settings, ok := cfg["user_settings_json"].(string); ok && settings != "" {
			if err := json.Unmarshal([]byte(settings), &res.UserSettingsJSON); err != nil {
				return fmt.Errorf("failed expanding IntegrationsServer user_settings_json: %w", err)
			}
		}
		if settings, ok := cfg["user_settings_override_json"].(string); ok && settings != "" {
			if err := json.Unmarshal([]byte(settings), &res.UserSettingsOverrideJSON); err != nil {
				return fmt.Errorf("failed expanding IntegrationsServer user_settings_override_json: %w", err)
			}
		}
		if settings, ok := cfg["user_settings_yaml"].(string); ok && settings != "" {
			res.UserSettingsYaml = settings
		}
		if settings, ok := cfg["user_settings_override_yaml"].(string); ok && settings != "" {
			res.UserSettingsOverrideYaml = settings
		}

		if v, ok := cfg["docker_image"].(string); ok {
			res.DockerImage = v
		}
	}

	return nil
}

// defaultIntegrationsServerTopology iterates over all the templated topology elements and
// sets the size to the default when the template size is smaller than the
// deployment template default, the same is done on the ZoneCount.
func defaultIntegrationsServerTopology(topology []*models.IntegrationsServerTopologyElement) []*models.IntegrationsServerTopologyElement {
	for _, t := range topology {
		if *t.Size.Value < minimumIntegrationsServerSize {
			t.Size.Value = ec.Int32(minimumIntegrationsServerSize)
		}
		if t.ZoneCount < minimumZoneCount {
			t.ZoneCount = minimumZoneCount
		}
	}

	return topology
}

func matchIntegrationsServerTopology(id string, topologies []*models.IntegrationsServerTopologyElement) (*models.IntegrationsServerTopologyElement, error) {
	for _, t := range topologies {
		if t.InstanceConfigurationID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf(
		`IntegrationsServer topology: invalid instance_configuration_id: "%s" doesn't match any of the deployment template instance configurations`,
		id,
	)
}

// IntegrationsServerResource returns the IntegrationsServerPayload from a deployment
// template or an empty version of the payload.
func integrationsServerResource(res *models.DeploymentTemplateInfoV2) *models.IntegrationsServerPayload {
	if len(res.DeploymentTemplate.Resources.IntegrationsServer) == 0 {
		return nil
	}
	return res.DeploymentTemplate.Resources.IntegrationsServer[0]
}

// integrationsServerResourceFromUpdate returns the IntegrationsServerPayload from a deployment
// update request or an empty version of the payload.
func integrationsServerResourceFromUpdate(res *models.DeploymentUpdateResources) *models.IntegrationsServerPayload {
	if len(res.IntegrationsServer) == 0 {
		return nil
	}
	return res.IntegrationsServer[0]
}
