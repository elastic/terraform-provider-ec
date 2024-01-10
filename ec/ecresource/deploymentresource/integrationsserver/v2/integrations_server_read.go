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
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

type IntegrationsServer struct {
	ElasticsearchClusterRefId          *string                   `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                              *string                   `tfsdk:"ref_id"`
	ResourceId                         *string                   `tfsdk:"resource_id"`
	Region                             *string                   `tfsdk:"region"`
	HttpEndpoint                       *string                   `tfsdk:"http_endpoint"`
	HttpsEndpoint                      *string                   `tfsdk:"https_endpoint"`
	Endpoints                          *Endpoints                `tfsdk:"endpoints"`
	InstanceConfigurationId            *string                   `tfsdk:"instance_configuration_id"`
	LatestInstanceConfigurationId      *string                   `tfsdk:"latest_instance_configuration_id"`
	InstanceConfigurationVersion       *int                      `tfsdk:"instance_configuration_version"`
	LatestInstanceConfigurationVersion *int                      `tfsdk:"latest_instance_configuration_version"`
	Size                               *string                   `tfsdk:"size"`
	SizeResource                       *string                   `tfsdk:"size_resource"`
	ZoneCount                          int                       `tfsdk:"zone_count"`
	Config                             *IntegrationsServerConfig `tfsdk:"config"`
}

type Endpoints struct {
	Fleet *string `tfsdk:"fleet"`
	APM   *string `tfsdk:"apm"`
}

func ReadIntegrationsServers(in []*models.IntegrationsServerResourceInfo) (*IntegrationsServer, error) {
	for _, model := range in {
		if util.IsCurrentIntegrationsServerPlanEmpty(model) || IsIntegrationsServerStopped(model) {
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
		srv.InstanceConfigurationVersion = topologies[0].InstanceConfigurationVersion
		srv.Size = topologies[0].Size
		srv.SizeResource = topologies[0].SizeResource
		srv.ZoneCount = topologies[0].ZoneCount
	}

	srv.ElasticsearchClusterRefId = in.ElasticsearchClusterRefID

	srv.HttpEndpoint, srv.HttpsEndpoint = converters.ExtractEndpoints(in.Info.Metadata)
	srv.Endpoints = readEndpoints(in)

	cfg, err := readIntegrationsServerConfigs(plan.IntegrationsServer)

	if err != nil {
		return nil, err
	}

	srv.Config = cfg

	return &srv, nil
}

func readEndpoints(in *models.IntegrationsServerResourceInfo) *Endpoints {
	endpoints := &Endpoints{}
	hasValidEndpoints := false
	for _, url := range in.Info.Metadata.ServicesUrls {
		if url.Service == nil || url.URL == nil {
			continue
		}

		switch *url.Service {
		case "apm":
			endpoints.APM = url.URL
			hasValidEndpoints = true
		case "fleet":
			endpoints.Fleet = url.URL
			hasValidEndpoints = true
		}
	}

	if !hasValidEndpoints {
		return nil
	}

	return endpoints
}

// IsIntegrationsServerStopped returns true if the resource is stopped.
func IsIntegrationsServerStopped(res *models.IntegrationsServerResourceInfo) bool {
	return res == nil || res.Info == nil || res.Info.Status == nil ||
		*res.Info.Status == "stopped"
}

func SetLatestInstanceConfigInfo(currentTopology *IntegrationsServer, latestTopology *models.IntegrationsServerTopologyElement) {
	if currentTopology != nil && latestTopology != nil {
		currentTopology.LatestInstanceConfigurationId = &latestTopology.InstanceConfigurationID
		if latestTopology.InstanceConfigurationVersion != nil {
			latestVersion := int(*latestTopology.InstanceConfigurationVersion)
			currentTopology.LatestInstanceConfigurationVersion = &latestVersion
		}
	}
}
