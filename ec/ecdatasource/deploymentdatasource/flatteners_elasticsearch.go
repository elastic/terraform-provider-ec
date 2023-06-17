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
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/models"

	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

// flattenElasticsearchResources takes in Elasticsearch resource models and returns its
// flattened form.
func flattenElasticsearchResources(ctx context.Context, in []*models.ElasticsearchResourceInfo) (types.List, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	var result = make([]elasticsearchResourceInfoModelV0, 0, len(in))

	for _, res := range in {
		model := elasticsearchResourceInfoModelV0{
			Topology: types.ListNull(types.ObjectType{AttrTypes: elasticsearchTopologyAttrTypes()}),
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

			if !util.IsCurrentEsPlanEmpty(res) {
				var plan = res.Info.PlanInfo.Current.Plan

				if plan.Elasticsearch != nil {
					model.Version = types.StringValue(plan.Elasticsearch.Version)
				}

				if plan.AutoscalingEnabled != nil {
					model.Autoscale = types.StringValue(strconv.FormatBool(*plan.AutoscalingEnabled))
				}

				var diags diag.Diagnostics
				model.Topology, diags = flattenElasticsearchTopology(ctx, plan)
				diagnostics.Append(diags...)
			}

			if res.Info.Metadata != nil {
				model.CloudID = types.StringValue(res.Info.Metadata.CloudID)
				model.HttpEndpoint, model.HttpsEndpoint = converters.ExtractEndpointsToTypes(res.Info.Metadata)
			}
		}

		result = append(result, model)
	}

	target, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: elasticsearchResourceInfoAttrTypes()}, result)
	diagnostics.Append(diags...)

	return target, diagnostics
}

func flattenElasticsearchTopology(ctx context.Context, plan *models.ElasticsearchClusterPlan) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result = make([]elasticsearchTopologyModelV0, 0, len(plan.ClusterTopology))
	for _, topology := range plan.ClusterTopology {
		model := elasticsearchTopologyModelV0{
			NodeRoles:   types.SetNull(types.StringType),
			Autoscaling: types.ListNull(elasticsearchAutoscalingElemType()),
		}

		if isElasticsearchSizePopulated(topology) && *topology.Size.Value == 0 {
			continue
		}

		model.InstanceConfigurationID = types.StringValue(topology.InstanceConfigurationID)

		if isElasticsearchSizePopulated(topology) {
			model.Size = types.StringValue(util.MemoryToState(*topology.Size.Value))
			model.SizeResource = types.StringValue(*topology.Size.Resource)
		}

		model.ZoneCount = types.Int64Value(int64(topology.ZoneCount))

		if topology.NodeType != nil {
			if topology.NodeType.Data != nil {
				model.NodeTypeData = types.BoolValue(*topology.NodeType.Data)
			}

			if topology.NodeType.Ingest != nil {
				model.NodeTypeIngest = types.BoolValue(*topology.NodeType.Ingest)
			}

			if topology.NodeType.Master != nil {
				model.NodeTypeMaster = types.BoolValue(*topology.NodeType.Master)
			}

			if topology.NodeType.Ml != nil {
				model.NodeTypeMl = types.BoolValue(*topology.NodeType.Ml)
			}
		}

		if len(topology.NodeRoles) > 0 {
			var d diag.Diagnostics
			model.NodeRoles, d = types.SetValueFrom(ctx, types.StringType, topology.NodeRoles)
			diags.Append(d...)
		}

		var autoscaling elasticsearchAutoscalingModel
		var hasAutoscalingModel = false
		if limit := topology.AutoscalingMax; limit != nil {
			autoscaling.MaxSizeResource = types.StringValue(*limit.Resource)
			autoscaling.MaxSize = types.StringValue(util.MemoryToState(*limit.Value))
			hasAutoscalingModel = true
		}

		if limit := topology.AutoscalingMin; limit != nil {
			autoscaling.MinSizeResource = types.StringValue(*limit.Resource)
			autoscaling.MinSize = types.StringValue(util.MemoryToState(*limit.Value))
			hasAutoscalingModel = true
		}

		if topology.AutoscalingPolicyOverrideJSON != nil {
			b, err := json.Marshal(topology.AutoscalingPolicyOverrideJSON)
			if err != nil {
				diags.AddError(
					"Invalid elasticsearch topology policy_override_json",
					fmt.Sprintf("elasticsearch topology %s: unable to persist policy_override_json: %v", topology.ID, err),
				)
			} else {
				autoscaling.PolicyOverrideJson = types.StringValue(string(b))
				hasAutoscalingModel = true
			}
		}

		if hasAutoscalingModel {
			var d diag.Diagnostics
			model.Autoscaling, d = types.ListValueFrom(ctx, elasticsearchAutoscalingElemType(), []elasticsearchAutoscalingModel{autoscaling})
			diags.Append(d...)
		}

		result = append(result, model)
	}

	var d diag.Diagnostics
	target, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: elasticsearchTopologyAttrTypes()}, result)
	diags.Append(d...)

	return target, diags
}

func isElasticsearchSizePopulated(topology *models.ElasticsearchClusterTopologyElement) bool {
	if topology.Size != nil && topology.Size.Value != nil {
		return true
	}

	return false
}
