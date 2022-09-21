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
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/models"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

// flattenElasticsearchResources takes in Elasticsearch resource models and returns its
// flattened form.
func flattenElasticsearchResources(ctx context.Context, in []*models.ElasticsearchResourceInfo, target interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var result = make([]elasticsearchResourceInfoModelV0, 0, len(in))

	for _, res := range in {
		model := elasticsearchResourceInfoModelV0{
			Topology: types.List{ElemType: types.ObjectType{AttrTypes: elasticsearchTopologyAttrTypes()}},
		}

		if res.RefID != nil {
			model.RefID = types.String{Value: *res.RefID}
		}

		if res.Info != nil {
			if res.Info.Healthy != nil {
				model.Healthy = types.Bool{Value: *res.Info.Healthy}
			}

			if res.Info.ClusterID != nil {
				model.ResourceID = types.String{Value: *res.Info.ClusterID}
			}

			if res.Info.Status != nil {
				model.Status = types.String{Value: *res.Info.Status}
			}

			if !util.IsCurrentEsPlanEmpty(res) {
				var plan = res.Info.PlanInfo.Current.Plan

				if plan.Elasticsearch != nil {
					model.Version = types.String{Value: plan.Elasticsearch.Version}
				}

				if plan.AutoscalingEnabled != nil {
					model.Autoscale = types.String{Value: strconv.FormatBool(*plan.AutoscalingEnabled)}
				}

				diags.Append(flattenElasticsearchTopology(ctx, plan, &model.Topology)...)
			}

			if res.Info.Metadata != nil {
				model.CloudID = types.String{Value: res.Info.Metadata.CloudID}

				endpoints := util.FlattenClusterEndpoint(res.Info.Metadata)
				if endpoints != nil {
					model.HttpEndpoint = types.String{Value: endpoints["http_endpoint"].(string)}
					model.HttpsEndpoint = types.String{Value: endpoints["https_endpoint"].(string)}
				}
			}
		}

		result = append(result, model)
	}

	diags.Append(tfsdk.ValueFrom(ctx, result, types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: elasticsearchResourceInfoAttrTypes(),
		},
	}, target)...)

	return diags
}

func flattenElasticsearchTopology(ctx context.Context, plan *models.ElasticsearchClusterPlan, target interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var result = make([]elasticsearchTopologyModelV0, 0, len(plan.ClusterTopology))
	for _, topology := range plan.ClusterTopology {
		model := elasticsearchTopologyModelV0{
			NodeRoles: types.Set{ElemType: types.StringType},
		}

		if isElasticsearchSizePopulated(topology) && *topology.Size.Value == 0 {
			continue
		}

		model.InstanceConfigurationID = types.String{Value: topology.InstanceConfigurationID}

		if isElasticsearchSizePopulated(topology) {
			model.Size = types.String{Value: util.MemoryToState(*topology.Size.Value)}
			model.SizeResource = types.String{Value: *topology.Size.Resource}
		}

		model.ZoneCount = types.Int64{Value: int64(topology.ZoneCount)}

		if topology.NodeType != nil {
			if topology.NodeType.Data != nil {
				model.NodeTypeData = types.Bool{Value: *topology.NodeType.Data}
			}

			if topology.NodeType.Ingest != nil {
				model.NodeTypeIngest = types.Bool{Value: *topology.NodeType.Ingest}
			}

			if topology.NodeType.Master != nil {
				model.NodeTypeMaster = types.Bool{Value: *topology.NodeType.Master}
			}

			if topology.NodeType.Ml != nil {
				model.NodeTypeMl = types.Bool{Value: *topology.NodeType.Ml}
			}
		}

		if len(topology.NodeRoles) > 0 {
			diags.Append(tfsdk.ValueFrom(ctx, util.StringToItems(topology.NodeRoles...), types.SetType{ElemType: types.StringType}, &model.NodeRoles)...)
		}

		var autoscaling elasticsearchAutoscalingModel
		var empty = true
		if limit := topology.AutoscalingMax; limit != nil {
			autoscaling.MaxSizeResource = types.String{Value: *limit.Resource}
			autoscaling.MaxSize = types.String{Value: util.MemoryToState(*limit.Value)}
			empty = false
		}

		if limit := topology.AutoscalingMin; limit != nil {
			autoscaling.MinSizeResource = types.String{Value: *limit.Resource}
			autoscaling.MinSize = types.String{Value: util.MemoryToState(*limit.Value)}
			empty = false
		}

		if topology.AutoscalingPolicyOverrideJSON != nil {
			b, err := json.Marshal(topology.AutoscalingPolicyOverrideJSON)
			if err != nil {
				diags.AddError(
					"Invalid elasticsearch topology policy_override_json",
					fmt.Sprintf("elasticsearch topology %s: unable to persist policy_override_json: %v", topology.ID, err),
				)
			} else {
				autoscaling.PolicyOverrideJson = types.String{Value: string(b)}
				empty = false
			}
		}

		if !empty {
			diags.Append(tfsdk.ValueFrom(ctx, []elasticsearchAutoscalingModel{autoscaling}, elasticsearchAutoscalingListType(), &model.Autoscaling)...)
		}

		result = append(result, model)
	}

	diags.Append(tfsdk.ValueFrom(ctx, result, types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: elasticsearchTopologyAttrTypes(),
		},
	}, target)...)

	return diags
}

func isElasticsearchSizePopulated(topology *models.ElasticsearchClusterTopologyElement) bool {
	if topology.Size != nil && topology.Size.Value != nil {
		return true
	}

	return false
}
