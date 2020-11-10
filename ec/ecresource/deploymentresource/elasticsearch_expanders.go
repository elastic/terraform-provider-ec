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
		if err := expandEsConfig(cfg, res.Plan.Elasticsearch); err != nil {
			return nil, err
		}
	}

	if snap, ok := es["snapshot_source"]; ok && len(snap.([]interface{})) > 0 {
		res.Plan.Transient = &models.TransientElasticsearchPlanConfiguration{
			RestoreSnapshot: &models.RestoreSnapshotConfiguration{},
		}
		expandSnapshotSource(snap, res.Plan.Transient.RestoreSnapshot)
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

		if c, ok := topology["config"]; ok {
			if elem.Elasticsearch == nil && c != nil {
				elem.Elasticsearch = &models.ElasticsearchConfiguration{}
			}
			if err = expandEsConfig(c, elem.Elasticsearch); err != nil {
				return nil, err
			}
			if reflect.DeepEqual(elem.Elasticsearch, &models.ElasticsearchConfiguration{}) {
				elem.Elasticsearch = nil
			}
		}

		res = append(res, elem)
	}

	return res, nil
}

func expandEsConfig(raw interface{}, esCfg *models.ElasticsearchConfiguration) error {
	for _, rawCfg := range raw.([]interface{}) {
		var cfg = rawCfg.(map[string]interface{})
		if settings, ok := cfg["user_settings_json"]; ok && settings != nil {
			if s, ok := settings.(string); ok && s != "" {
				if err := json.Unmarshal([]byte(s), &esCfg.UserSettingsJSON); err != nil {
					return fmt.Errorf(
						"failed expanding elasticsearch user_settings_json: %w", err,
					)
				}
			}
		}
		if settings, ok := cfg["user_settings_override_json"]; ok && settings != nil {
			if s, ok := settings.(string); ok && s != "" {
				if err := json.Unmarshal([]byte(s), &esCfg.UserSettingsOverrideJSON); err != nil {
					return fmt.Errorf(
						"failed expanding elasticsearch user_settings_override_json: %w", err,
					)
				}
			}
		}
		if settings, ok := cfg["user_settings_yaml"]; ok {
			esCfg.UserSettingsYaml = settings.(string)
		}
		if settings, ok := cfg["user_settings_override_yaml"]; ok {
			esCfg.UserSettingsOverrideYaml = settings.(string)
		}

		if v, ok := cfg["plugins"]; ok {
			esCfg.EnabledBuiltInPlugins = util.ItemsToString(v.(*schema.Set).List())
		}
	}

	return nil
}

func expandSnapshotSource(raw interface{}, restore *models.RestoreSnapshotConfiguration) {
	for _, rawRestore := range raw.([]interface{}) {
		var rs = rawRestore.(map[string]interface{})
		if clusterID, ok := rs["source_elasticsearch_cluster_id"]; ok {
			restore.SourceClusterID = clusterID.(string)
		}

		if snapshotName, ok := rs["snapshot_name"]; ok {
			restore.SnapshotName = ec.String(snapshotName.(string))
		}

	}
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

// esResource returns the ElaticsearchPayload from a deployment
// template or an empty version of the payload.
func esResource(res *models.DeploymentTemplateInfoV2) *models.ElasticsearchPayload {
	if len(res.DeploymentTemplate.Resources.Elasticsearch) == 0 {
		return &models.ElasticsearchPayload{
			Plan: &models.ElasticsearchClusterPlan{
				Elasticsearch: &models.ElasticsearchConfiguration{},
			},
			Settings: &models.ElasticsearchClusterSettings{},
		}
	}
	return res.DeploymentTemplate.Resources.Elasticsearch[0]
}
