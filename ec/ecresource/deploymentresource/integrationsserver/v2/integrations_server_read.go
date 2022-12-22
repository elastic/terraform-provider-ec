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
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/utils"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

type IntegrationsServer struct {
	ElasticsearchClusterRefId *string                   `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     *string                   `tfsdk:"ref_id"`
	ResourceId                *string                   `tfsdk:"resource_id"`
	Region                    *string                   `tfsdk:"region"`
	HttpEndpoint              *string                   `tfsdk:"http_endpoint"`
	HttpsEndpoint             *string                   `tfsdk:"https_endpoint"`
	InstanceConfigurationId   *string                   `tfsdk:"instance_configuration_id"`
	Size                      *string                   `tfsdk:"size"`
	SizeResource              *string                   `tfsdk:"size_resource"`
	ZoneCount                 int                       `tfsdk:"zone_count"`
	Config                    *IntegrationsServerConfig `tfsdk:"config"`
}

func ReadIntegrationsServers(in []*models.IntegrationsServerResourceInfo) (*IntegrationsServer, error) {
	for _, model := range in {
		if util.IsCurrentIntegrationsServerPlanEmpty(model) || utils.IsIntegrationsServerResourceStopped(model) {
			continue
		}

		srv, err := readIntegrationsServer(model)
		if err != nil {
			return nil, err
		}

		return srv, nil
	}

	return nil, nil
}

func readIntegrationsServer(in *models.IntegrationsServerResourceInfo) (*IntegrationsServer, error) {

	var srv IntegrationsServer

	srv.RefId = in.RefID

	srv.ResourceId = in.Info.ID

	srv.Region = in.Region

	plan := in.Info.PlanInfo.Current.Plan

	topologies, err := readIntegrationsServerTopologies(plan.ClusterTopology)

	if err != nil {
		return nil, err
	}

	if len(topologies) > 0 {
		srv.InstanceConfigurationId = topologies[0].InstanceConfigurationId
		srv.Size = topologies[0].Size
		srv.SizeResource = topologies[0].SizeResource
		srv.ZoneCount = topologies[0].ZoneCount
	}

	srv.ElasticsearchClusterRefId = in.ElasticsearchClusterRefID

	srv.HttpEndpoint, srv.HttpsEndpoint = converters.ExtractEndpoints(in.Info.Metadata)

	cfg, err := readIntegrationsServerConfigs(plan.IntegrationsServer)

	if err != nil {
		return nil, err
	}

	srv.Config = cfg

	return &srv, nil
}
