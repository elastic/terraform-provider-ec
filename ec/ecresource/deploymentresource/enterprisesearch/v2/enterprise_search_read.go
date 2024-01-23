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

package v2

import (
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

type EnterpriseSearch struct {
	ElasticsearchClusterRefId          *string                 `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                              *string                 `tfsdk:"ref_id"`
	ResourceId                         *string                 `tfsdk:"resource_id"`
	Region                             *string                 `tfsdk:"region"`
	HttpEndpoint                       *string                 `tfsdk:"http_endpoint"`
	HttpsEndpoint                      *string                 `tfsdk:"https_endpoint"`
	InstanceConfigurationId            *string                 `tfsdk:"instance_configuration_id"`
	LatestInstanceConfigurationId      *string                 `tfsdk:"latest_instance_configuration_id"`
	InstanceConfigurationVersion       *int                    `tfsdk:"instance_configuration_version"`
	LatestInstanceConfigurationVersion *int                    `tfsdk:"latest_instance_configuration_version"`
	Size                               *string                 `tfsdk:"size"`
	SizeResource                       *string                 `tfsdk:"size_resource"`
	ZoneCount                          int                     `tfsdk:"zone_count"`
	NodeTypeAppserver                  *bool                   `tfsdk:"node_type_appserver"`
	NodeTypeConnector                  *bool                   `tfsdk:"node_type_connector"`
	NodeTypeWorker                     *bool                   `tfsdk:"node_type_worker"`
	Config                             *EnterpriseSearchConfig `tfsdk:"config"`
}

type EnterpriseSearches []EnterpriseSearch

func ReadEnterpriseSearch(in *models.EnterpriseSearchResourceInfo) (*EnterpriseSearch, error) {
	if util.IsCurrentEssPlanEmpty(in) || IsEnterpriseSearchStopped(in) {
		return nil, nil
	}

	var ess EnterpriseSearch

	ess.RefId = in.RefID

	ess.ResourceId = in.Info.ID

	ess.Region = in.Region

	plan := in.Info.PlanInfo.Current.Plan

	topologies, err := readEnterpriseSearchTopologies(plan.ClusterTopology)

	if err != nil {
		return nil, err
	}

	if len(topologies) > 0 {
		ess.InstanceConfigurationId = topologies[0].InstanceConfigurationId
		ess.InstanceConfigurationVersion = topologies[0].InstanceConfigurationVersion
		ess.Size = topologies[0].Size
		ess.SizeResource = topologies[0].SizeResource
		ess.ZoneCount = topologies[0].ZoneCount
		ess.NodeTypeAppserver = topologies[0].NodeTypeAppserver
		ess.NodeTypeConnector = topologies[0].NodeTypeConnector
		ess.NodeTypeWorker = topologies[0].NodeTypeWorker
	}

	ess.ElasticsearchClusterRefId = in.ElasticsearchClusterRefID

	ess.HttpEndpoint, ess.HttpsEndpoint = converters.ExtractEndpoints(in.Info.Metadata)

	cfg, err := readEnterpriseSearchConfig(plan.EnterpriseSearch)
	if err != nil {
		return nil, err
	}
	ess.Config = cfg

	return &ess, nil
}

func ReadEnterpriseSearches(in []*models.EnterpriseSearchResourceInfo) (*EnterpriseSearch, error) {
	for _, model := range in {
		if util.IsCurrentEssPlanEmpty(model) || IsEnterpriseSearchStopped(model) {
			continue
		}

		es, err := ReadEnterpriseSearch(model)
		if err != nil {
			return nil, err
		}

		return es, nil
	}

	return nil, nil
}

// IsEnterpriseSearchStopped returns true if the resource is stopped.
func IsEnterpriseSearchStopped(res *models.EnterpriseSearchResourceInfo) bool {
	return res == nil || res.Info == nil || res.Info.Status == nil ||
		*res.Info.Status == "stopped"
}

func SetLatestInstanceConfigInfo(currentTopology *EnterpriseSearch, latestTopology *models.EnterpriseSearchTopologyElement) {
	if currentTopology != nil && latestTopology != nil {
		currentTopology.LatestInstanceConfigurationId = &latestTopology.InstanceConfigurationID
		if latestTopology.InstanceConfigurationVersion != nil {
			currentTopology.LatestInstanceConfigurationVersion = ec.Int(int(*latestTopology.InstanceConfigurationVersion))
		}
	}
}

func SetLatestInstanceConfigInfoToCurrent(topology *EnterpriseSearch) {
	if topology != nil {
		topology.LatestInstanceConfigurationId = topology.InstanceConfigurationId
		topology.LatestInstanceConfigurationVersion = topology.InstanceConfigurationVersion
	}
}
