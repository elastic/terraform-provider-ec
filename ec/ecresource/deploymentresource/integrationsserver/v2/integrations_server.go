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
	"context"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	topologyv1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/topology/v1"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/utils"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IntegrationsServerTF struct {
	ElasticsearchClusterRefId types.String `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     types.String `tfsdk:"ref_id"`
	ResourceId                types.String `tfsdk:"resource_id"`
	Region                    types.String `tfsdk:"region"`
	HttpEndpoint              types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint             types.String `tfsdk:"https_endpoint"`
	InstanceConfigurationId   types.String `tfsdk:"instance_configuration_id"`
	Size                      types.String `tfsdk:"size"`
	SizeResource              types.String `tfsdk:"size_resource"`
	ZoneCount                 types.Int64  `tfsdk:"zone_count"`
	Config                    types.Object `tfsdk:"config"`
}

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

func (srv IntegrationsServerTF) Payload(ctx context.Context, payload models.IntegrationsServerPayload) (*models.IntegrationsServerPayload, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !srv.ElasticsearchClusterRefId.IsNull() {
		payload.ElasticsearchClusterRefID = &srv.ElasticsearchClusterRefId.Value
	}

	if !srv.RefId.IsNull() {
		payload.RefID = &srv.RefId.Value
	}

	if srv.Region.Value != "" {
		payload.Region = &srv.Region.Value
	}

	ds := integrationsServerConfigPayload(ctx, srv.Config, payload.Plan.IntegrationsServer)
	diags.Append(ds...)

	topologyTF := topologyv1.TopologyTF{
		InstanceConfigurationId: srv.InstanceConfigurationId,
		Size:                    srv.Size,
		SizeResource:            srv.SizeResource,
		ZoneCount:               srv.ZoneCount,
	}

	toplogyPayload, ds := integrationsServerTopologyPayload(ctx, topologyTF, defaultIntegrationsServerTopology(payload.Plan.ClusterTopology), 0)

	diags.Append(ds...)

	if !ds.HasError() && toplogyPayload != nil {
		payload.Plan.ClusterTopology = []*models.IntegrationsServerTopologyElement{toplogyPayload}
	}

	return &payload, diags
}

func IntegrationsServerPayload(ctx context.Context, srvObj types.Object, template *models.DeploymentTemplateInfoV2) (*models.IntegrationsServerPayload, diag.Diagnostics) {
	var diags diag.Diagnostics

	var srv *IntegrationsServerTF

	if diags = tfsdk.ValueAs(ctx, srvObj, &srv); diags.HasError() {
		return nil, diags
	}

	if srv == nil {
		return nil, nil
	}

	templatePayload := integrationsServerResource(template)

	if templatePayload == nil {
		diags.AddError("integrations_server payload error", "integrations_server specified but deployment template is not configured for it. Use a different template if you wish to add integrations_server")
		return nil, diags
	}

	payload, diags := srv.Payload(ctx, *templatePayload)

	if diags.HasError() {
		return nil, diags
	}

	return payload, nil
}

// integrationsServerResource returns the IntegrationsServerPayload from a deployment
// template or an empty version of the payload.
func integrationsServerResource(template *models.DeploymentTemplateInfoV2) *models.IntegrationsServerPayload {
	if template == nil || len(template.DeploymentTemplate.Resources.IntegrationsServer) == 0 {
		return nil
	}
	return template.DeploymentTemplate.Resources.IntegrationsServer[0]
}
