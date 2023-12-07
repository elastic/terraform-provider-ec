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
	v1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/enterprisesearch/v1"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EnterpriseSearchTF struct {
	ElasticsearchClusterRefId    types.String `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                        types.String `tfsdk:"ref_id"`
	ResourceId                   types.String `tfsdk:"resource_id"`
	Region                       types.String `tfsdk:"region"`
	HttpEndpoint                 types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint                types.String `tfsdk:"https_endpoint"`
	InstanceConfigurationId      types.String `tfsdk:"instance_configuration_id"`
	InstanceConfigurationVersion types.Int64  `tfsdk:"instance_configuration_version"`
	Size                         types.String `tfsdk:"size"`
	SizeResource                 types.String `tfsdk:"size_resource"`
	ZoneCount                    types.Int64  `tfsdk:"zone_count"`
	NodeTypeAppserver            types.Bool   `tfsdk:"node_type_appserver"`
	NodeTypeConnector            types.Bool   `tfsdk:"node_type_connector"`
	NodeTypeWorker               types.Bool   `tfsdk:"node_type_worker"`
	Config                       types.Object `tfsdk:"config"`
}

func (es *EnterpriseSearchTF) payload(ctx context.Context, payload models.EnterpriseSearchPayload) (*models.EnterpriseSearchPayload, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !es.ElasticsearchClusterRefId.IsNull() {
		payload.ElasticsearchClusterRefID = ec.String(es.ElasticsearchClusterRefId.ValueString())
	}

	if !es.RefId.IsNull() {
		payload.RefID = ec.String(es.RefId.ValueString())
	}

	if es.Region.ValueString() != "" {
		payload.Region = ec.String(es.Region.ValueString())
	}

	if !es.Config.IsNull() && !es.Config.IsUnknown() {
		var config *v1.EnterpriseSearchConfigTF

		ds := tfsdk.ValueAs(ctx, es.Config, &config)

		diags.Append(ds...)

		if !ds.HasError() && config != nil {
			diags.Append(enterpriseSearchConfigPayload(ctx, *config, payload.Plan.EnterpriseSearch)...)
		}
	}

	topologyTF := v1.EnterpriseSearchTopologyTF{
		InstanceConfigurationId:      es.InstanceConfigurationId,
		InstanceConfigurationVersion: es.InstanceConfigurationVersion,
		Size:                         es.Size,
		SizeResource:                 es.SizeResource,
		ZoneCount:                    es.ZoneCount,
		NodeTypeAppserver:            es.NodeTypeAppserver,
		NodeTypeConnector:            es.NodeTypeConnector,
		NodeTypeWorker:               es.NodeTypeWorker,
	}

	// Always use the first topology element - discard any other topology elements
	topology, ds := enterpriseSearchTopologyPayload(ctx, topologyTF, defaultTopology(payload.Plan.ClusterTopology)[0])

	diags = append(diags, ds...)

	if topology != nil {
		payload.Plan.ClusterTopology = []*models.EnterpriseSearchTopologyElement{topology}
	}

	return &payload, diags
}

func EnterpriseSearchesPayload(ctx context.Context, esObj types.Object, updateResources *models.DeploymentUpdateResources) (*models.EnterpriseSearchPayload, diag.Diagnostics) {
	var diags diag.Diagnostics

	var es *EnterpriseSearchTF

	if diags = tfsdk.ValueAs(ctx, esObj, &es); diags.HasError() {
		return nil, diags
	}

	if es == nil {
		return nil, nil
	}

	templatePayload := payloadFromUpdate(updateResources)

	if templatePayload == nil {
		diags.AddError(
			"enterprise_search payload error",
			"enterprise_search specified but deployment template is not configured for it. Use a different template if you wish to add enterprise_search",
		)
		return nil, diags
	}

	payload, diags := es.payload(ctx, *templatePayload)

	if diags.HasError() {
		return nil, diags
	}

	return payload, nil
}

// payloadFromUpdate returns the EnterpriseSearchPayload from a deployment
// template or an empty version of the payload.
func payloadFromUpdate(updateResources *models.DeploymentUpdateResources) *models.EnterpriseSearchPayload {
	if updateResources == nil || len(updateResources.EnterpriseSearch) == 0 {
		return nil
	}
	return updateResources.EnterpriseSearch[0]
}
