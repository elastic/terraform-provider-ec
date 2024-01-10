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
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	v1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/kibana/v1"
	topologyv1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/topology/v1"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type KibanaTF struct {
	ElasticsearchClusterRefId          types.String `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                              types.String `tfsdk:"ref_id"`
	ResourceId                         types.String `tfsdk:"resource_id"`
	Region                             types.String `tfsdk:"region"`
	HttpEndpoint                       types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint                      types.String `tfsdk:"https_endpoint"`
	InstanceConfigurationId            types.String `tfsdk:"instance_configuration_id"`
	LatestInstanceConfigurationId      types.String `tfsdk:"latest_instance_configuration_id"`
	InstanceConfigurationVersion       types.Int64  `tfsdk:"instance_configuration_version"`
	LatestInstanceConfigurationVersion types.Int64  `tfsdk:"latest_instance_configuration_version"`
	Size                               types.String `tfsdk:"size"`
	SizeResource                       types.String `tfsdk:"size_resource"`
	ZoneCount                          types.Int64  `tfsdk:"zone_count"`
	Config                             types.Object `tfsdk:"config"`
}

func (kibana KibanaTF) payload(ctx context.Context, payload models.KibanaPayload) (*models.KibanaPayload, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !kibana.ElasticsearchClusterRefId.IsNull() {
		payload.ElasticsearchClusterRefID = ec.String(kibana.ElasticsearchClusterRefId.ValueString())
	}

	if !kibana.RefId.IsNull() {
		payload.RefID = ec.String(kibana.RefId.ValueString())
	}

	if kibana.Region.ValueString() != "" {
		payload.Region = ec.String(kibana.Region.ValueString())
	}

	if !kibana.Config.IsNull() && !kibana.Config.IsUnknown() {
		var config *v1.KibanaConfigTF

		ds := tfsdk.ValueAs(ctx, kibana.Config, &config)

		diags.Append(ds...)

		if !ds.HasError() {
			diags.Append(kibanaConfigPayload(config, payload.Plan.Kibana)...)
		}
	}

	topologyTF := topologyv1.TopologyTF{
		InstanceConfigurationId:      kibana.InstanceConfigurationId,
		InstanceConfigurationVersion: kibana.InstanceConfigurationVersion,
		Size:                         kibana.Size,
		SizeResource:                 kibana.SizeResource,
		ZoneCount:                    kibana.ZoneCount,
	}

	topologyPayload, ds := kibanaTopologyPayload(ctx, topologyTF, defaultKibanaTopology(payload.Plan.ClusterTopology)[0])

	diags.Append(ds...)

	if !ds.HasError() && topologyPayload != nil {
		payload.Plan.ClusterTopology = []*models.KibanaClusterTopologyElement{topologyPayload}
	}

	return &payload, diags
}

func KibanaPayload(ctx context.Context, kibanaObj types.Object, updateResources *models.DeploymentUpdateResources) (*models.KibanaPayload, diag.Diagnostics) {
	var kibanaTF *KibanaTF

	var diags diag.Diagnostics

	if diags = tfsdk.ValueAs(ctx, kibanaObj, &kibanaTF); diags.HasError() {
		return nil, diags
	}

	if kibanaTF == nil {
		return nil, nil
	}

	templatePlayload := payloadFromUpdate(updateResources)

	if templatePlayload == nil {
		diags.AddError("kibana payload error", "kibana specified but deployment template is not configured for it. Use a different template if you wish to add kibana")
		return nil, diags
	}

	payload, diags := kibanaTF.payload(ctx, *templatePlayload)

	if diags.HasError() {
		return nil, diags
	}

	return payload, nil
}

// payloadFromUpdate returns the KibanaPayload from a deployment
// template or an empty version of the payload.
func payloadFromUpdate(updateResources *models.DeploymentUpdateResources) *models.KibanaPayload {
	if updateResources == nil || len(updateResources.Kibana) == 0 {
		return nil
	}
	return updateResources.Kibana[0]
}
