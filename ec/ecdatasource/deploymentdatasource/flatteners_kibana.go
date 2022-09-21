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
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/models"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

// flattenKibanaResources takes in Kibana resource models and returns its
// flattened form.
func flattenKibanaResources(ctx context.Context, in []*models.KibanaResourceInfo, target interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var result = make([]kibanaResourceInfoModelV0, 0, len(in))

	for _, res := range in {
		model := kibanaResourceInfoModelV0{
			Topology: types.List{ElemType: types.ObjectType{AttrTypes: kibanaTopologyAttrTypes()}},
		}

		if res.ElasticsearchClusterRefID != nil {
			model.ElasticsearchClusterRefID = types.String{Value: *res.ElasticsearchClusterRefID}
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

			if !util.IsCurrentKibanaPlanEmpty(res) {
				var plan = res.Info.PlanInfo.Current.Plan

				if plan.Kibana != nil {
					model.Version = types.String{Value: plan.Kibana.Version}
				}

				diags.Append(flattenKibanaTopology(ctx, plan, &model.Topology)...)
			}

			if res.Info.Metadata != nil {
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
			AttrTypes: kibanaResourceInfoAttrTypes(),
		},
	}, target)...)

	return diags
}

func flattenKibanaTopology(ctx context.Context, plan *models.KibanaClusterPlan, target interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var result = make([]kibanaTopologyModelV0, 0, len(plan.ClusterTopology))
	for _, topology := range plan.ClusterTopology {
		var model kibanaTopologyModelV0

		if isKibanaSizePopulated(topology) && *topology.Size.Value == 0 {
			continue
		}

		model.InstanceConfigurationID = types.String{Value: topology.InstanceConfigurationID}

		if isKibanaSizePopulated(topology) {
			model.Size = types.String{Value: util.MemoryToState(*topology.Size.Value)}
			model.SizeResource = types.String{Value: *topology.Size.Resource}
		}

		model.ZoneCount = types.Int64{Value: int64(topology.ZoneCount)}

		result = append(result, model)
	}

	diags.Append(tfsdk.ValueFrom(ctx, result, types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: kibanaTopologyAttrTypes(),
		},
	}, target)...)

	return diags
}

func isKibanaSizePopulated(topology *models.KibanaClusterTopologyElement) bool {
	if topology.Size != nil && topology.Size.Value != nil {
		return true
	}

	return false
}
