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

// expandKibanaResources expands the flattened kibana resources into its models.
func expandKibanaResources(kibanas []interface{}, tpl *models.KibanaPayload) ([]*models.KibanaPayload, error) {
	if len(kibanas) == 0 {
		return nil, nil
	}

	if tpl == nil {
		return nil, errors.New("kibana specified but deployment template is not configured for it. Use a different template if you wish to add kibana")
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
	kibana := raw.(map[string]interface{})

	if esRefID, ok := kibana["elasticsearch_cluster_ref_id"].(string); ok {
		res.ElasticsearchClusterRefID = ec.String(esRefID)
	}

	if refID, ok := kibana["ref_id"].(string); ok {
		res.RefID = ec.String(refID)
	}

	if region, ok := kibana["region"].(string); ok && region != "" {
		res.Region = ec.String(region)
	}

	if cfg, ok := kibana["config"]; ok {
		if err := expandKibanaConfig(cfg, res.Plan.Kibana); err != nil {
			return nil, err
		}
	}

	if rt, ok := kibana["topology"].([]interface{}); ok && len(rt) > 0 {
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

func expandKibanaTopology(rawTopologies []interface{}, topologies []*models.KibanaClusterTopologyElement) ([]*models.KibanaClusterTopologyElement, error) {
	var res = make([]*models.KibanaClusterTopologyElement, 0, len(rawTopologies))
	for i, rawTop := range rawTopologies {
		var topology, ok = rawTop.(map[string]interface{})
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
		if t := defaultKibanaTopology(topologies); icID == "" && len(t) > i {
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

		if zones, ok := topology["zone_count"].(int); ok && zones > 0 {
			elem.ZoneCount = int32(zones)
		}

		res = append(res, elem)
	}

	return res, nil
}

func expandKibanaConfig(raw interface{}, res *models.KibanaConfiguration) error {
	for _, rawCfg := range raw.([]interface{}) {
		cfg, ok := rawCfg.(map[string]interface{})
		if !ok {
			continue
		}
		if settings, ok := cfg["user_settings_json"].(string); ok && settings != "" {
			if err := json.Unmarshal([]byte(settings), &res.UserSettingsJSON); err != nil {
				return fmt.Errorf("failed expanding kibana user_settings_json: %w", err)
			}
		}
		if settings, ok := cfg["user_settings_override_json"].(string); ok && settings != "" {
			if err := json.Unmarshal([]byte(settings), &res.UserSettingsOverrideJSON); err != nil {
				return fmt.Errorf("failed expanding kibana user_settings_override_json: %w", err)
			}
		}
		if settings, ok := cfg["user_settings_yaml"].(string); ok && settings != "" {
			res.UserSettingsYaml = settings
		}
		if settings, ok := cfg["user_settings_override_yaml"].(string); ok && settings != "" {
			res.UserSettingsOverrideYaml = settings
		}

		if v, ok := cfg["docker_image"].(string); ok && v != "" {
			res.DockerImage = v
		}
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

// kibanaResource returns the KibanaPayload from a deployment
// template or an empty version of the payload.
func kibanaResource(res *models.DeploymentTemplateInfoV2) *models.KibanaPayload {
	if len(res.DeploymentTemplate.Resources.Kibana) == 0 {
		return nil
	}
	return res.DeploymentTemplate.Resources.Kibana[0]
}

// kibanaResourceFromUpdate returns the KibanaPayload from a deployment
// update request or an empty version of the payload.
func kibanaResourceFromUpdate(res *models.DeploymentUpdateResources) *models.KibanaPayload {
	if len(res.Kibana) == 0 {
		return nil
	}
	return res.Kibana[0]
}
