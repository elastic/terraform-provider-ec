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

package deploymentdatasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/models"

	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

// flattenApmResources takes in Apm resource models and returns its
// flattened form.
func flattenApmResources(ctx context.Context, in []*models.ApmResourceInfo) (types.List, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	var result = make([]apmResourceInfoModelV0, 0, len(in))

	for _, res := range in {
		model := apmResourceInfoModelV0{
			Topology: types.ListNull(types.ObjectType{AttrTypes: apmTopologyAttrTypes()}),
		}

		if res.ElasticsearchClusterRefID != nil {
			model.ElasticsearchClusterRefID = types.StringValue(*res.ElasticsearchClusterRefID)
		}

		if res.RefID != nil {
			model.RefID = types.StringValue(*res.RefID)
		}

		if res.Info != nil {
			if res.Info.Healthy != nil {
				model.Healthy = types.BoolValue(*res.Info.Healthy)
			}

			if res.Info.ID != nil {
				model.ResourceID = types.StringValue(*res.Info.ID)
			}

			if res.Info.Status != nil {
				model.Status = types.StringValue(*res.Info.Status)
			}

			if !util.IsCurrentApmPlanEmpty(res) {
				var plan = res.Info.PlanInfo.Current.Plan

				if plan.Apm != nil {
					model.Version = types.StringValue(plan.Apm.Version)
				}

				var diags diag.Diagnostics
				model.Topology, diags = flattenApmTopology(ctx, plan)
				diagnostics.Append(diags...)
			}

			if res.Info.Metadata != nil {
				model.HttpEndpoint, model.HttpsEndpoint = converters.ExtractEndpointsToTypes(res.Info.Metadata)
			}
		}

		result = append(result, model)
	}

	target, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: apmResourceInfoAttrTypes()}, result)
	diagnostics.Append(diags...)

	return target, diagnostics
}

func flattenApmTopology(ctx context.Context, plan *models.ApmPlan) (types.List, diag.Diagnostics) {
	var result = make([]apmTopologyModelV0, 0, len(plan.ClusterTopology))
	for _, topology := range plan.ClusterTopology {
		var model apmTopologyModelV0

		if isApmSizePopulated(topology) && *topology.Size.Value == 0 {
			continue
		}

		model.InstanceConfigurationID = types.StringValue(topology.InstanceConfigurationID)

		if isApmSizePopulated(topology) {
			model.Size = types.StringValue(util.MemoryToState(*topology.Size.Value))
			model.SizeResource = types.StringValue(*topology.Size.Resource)
		}

		model.ZoneCount = types.Int64Value(int64(topology.ZoneCount))

		result = append(result, model)
	}

	target, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: apmTopologyAttrTypes()}, result)

	return target, diags
}

func isApmSizePopulated(topology *models.ApmTopologyElement) bool {
	if topology.Size != nil && topology.Size.Value != nil {
		return true
	}

	return false
}
