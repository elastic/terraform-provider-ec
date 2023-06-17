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

// flattenKibanaResources takes in Kibana resource models and returns its
// flattened form.
func flattenKibanaResources(ctx context.Context, in []*models.KibanaResourceInfo) (types.List, diag.Diagnostics) {
	var diagsnostics diag.Diagnostics
	var result = make([]kibanaResourceInfoModelV0, 0, len(in))

	for _, res := range in {
		model := kibanaResourceInfoModelV0{
			Topology: types.ListNull(types.ObjectType{AttrTypes: kibanaTopologyAttrTypes()}),
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

			if res.Info.ClusterID != nil {
				model.ResourceID = types.StringValue(*res.Info.ClusterID)
			}

			if res.Info.Status != nil {
				model.Status = types.StringValue(*res.Info.Status)
			}

			if !util.IsCurrentKibanaPlanEmpty(res) {
				var plan = res.Info.PlanInfo.Current.Plan

				if plan.Kibana != nil {
					model.Version = types.StringValue(plan.Kibana.Version)
				}

				var diags diag.Diagnostics
				model.Topology, diags = flattenKibanaTopology(ctx, plan)
				diagsnostics.Append(diags...)
			}

			if res.Info.Metadata != nil {
				model.HttpEndpoint, model.HttpsEndpoint = converters.ExtractEndpointsToTypes(res.Info.Metadata)
			}
		}

		result = append(result, model)
	}

	target, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: kibanaResourceInfoAttrTypes()}, result)
	diagsnostics.Append(diags...)

	return target, diagsnostics
}

func flattenKibanaTopology(ctx context.Context, plan *models.KibanaClusterPlan) (types.List, diag.Diagnostics) {
	var result = make([]kibanaTopologyModelV0, 0, len(plan.ClusterTopology))
	for _, topology := range plan.ClusterTopology {
		var model kibanaTopologyModelV0

		if isKibanaSizePopulated(topology) && *topology.Size.Value == 0 {
			continue
		}

		model.InstanceConfigurationID = types.StringValue(topology.InstanceConfigurationID)

		if isKibanaSizePopulated(topology) {
			model.Size = types.StringValue(util.MemoryToState(*topology.Size.Value))
			model.SizeResource = types.StringValue(*topology.Size.Resource)
		}

		model.ZoneCount = types.Int64Value(int64(topology.ZoneCount))

		result = append(result, model)
	}

	return types.ListValueFrom(ctx, types.ObjectType{AttrTypes: kibanaTopologyAttrTypes()}, result)
}

func isKibanaSizePopulated(topology *models.KibanaClusterTopologyElement) bool {
	if topology.Size != nil && topology.Size.Value != nil {
		return true
	}

	return false
}
