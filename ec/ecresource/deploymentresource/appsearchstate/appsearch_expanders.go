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

package appsearchstate

import (
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/terraform-providers/terraform-provider-ec/ec/ecresource/deploymentresource/deploymentstate"
)

// ExpandResources expands appsearch resources into their models.
func ExpandResources(apss []interface{}) ([]*models.AppSearchPayload, error) {
	if len(apss) == 0 {
		return nil, nil
	}

	result := make([]*models.AppSearchPayload, 0, len(apss))
	for _, raw := range apss {
		resResource, err := expandResource(raw)
		if err != nil {
			return nil, err
		}
		result = append(result, resResource)
	}

	return result, nil
}

func expandResource(raw interface{}) (*models.AppSearchPayload, error) {
	var es = raw.(map[string]interface{})
	var res = models.AppSearchPayload{
		Plan: &models.AppSearchPlan{
			Appsearch: &models.AppSearchConfiguration{},
		},
		Settings: &models.AppSearchSettings{},
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
		res.Plan.Appsearch.Version = version.(string)
	}

	if region, ok := es["region"]; ok {
		if r := region.(string); r != "" {
			res.Region = ec.String(r)
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

func expandTopology(raw interface{}) ([]*models.AppSearchTopologyElement, error) {
	var rawTopologies = raw.([]interface{})
	var res = make([]*models.AppSearchTopologyElement, 0, len(rawTopologies))
	for _, rawTop := range rawTopologies {
		var topology = rawTop.(map[string]interface{})
		var nodeType = parseNodeType(topology)

		size, err := deploymentstate.ParseTopologySize(topology)
		if err != nil {
			return nil, err
		}

		var elem = models.AppSearchTopologyElement{
			Size:     &size,
			NodeType: &nodeType,
		}

		if id, ok := topology["instance_configuration_id"]; ok {
			elem.InstanceConfigurationID = id.(string)
		}

		if zones, ok := topology["zone_count"]; ok {
			elem.ZoneCount = int32(zones.(int))
		}

		res = append(res, &elem)
	}

	return res, nil
}

func parseNodeType(topology map[string]interface{}) models.AppSearchNodeTypes {
	var result models.AppSearchNodeTypes
	if val, ok := topology["node_type_appserver"]; ok {
		result.Appserver = ec.Bool(val.(bool))
	}

	if val, ok := topology["node_type_worker"]; ok {
		result.Worker = ec.Bool(val.(bool))
	}

	return result
}
