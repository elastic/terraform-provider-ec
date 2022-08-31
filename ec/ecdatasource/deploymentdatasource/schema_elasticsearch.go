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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func elasticsearchResourceInfoSchema() tfsdk.Attribute {
	// TODO should we use tfsdk.ListNestedAttributes here? - see https://github.com/hashicorp/terraform-provider-hashicups-pf/blob/8f222d805d39445673e442a674168349a45bc054/hashicups/data_source_coffee.go#L22
	return tfsdk.Attribute{
		Computed: true,
		Type: types.ListType{ElemType: types.ObjectType{
			AttrTypes: elasticsearchResourceInfoAttrTypes(),
		}},
	}
}

func elasticsearchResourceInfoAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"autoscale":      types.StringType,
		"healthy":        types.BoolType,
		"cloud_id":       types.StringType,
		"http_endpoint":  types.StringType,
		"https_endpoint": types.StringType,
		"ref_id":         types.StringType,
		"resource_id":    types.StringType,
		"status":         types.StringType,
		"version":        types.StringType,
		"topology":       elasticsearchTopologySchema(),
	}
}

func elasticsearchTopologySchema() attr.Type {
	return types.ListType{ElemType: types.ObjectType{
		AttrTypes: elasticsearchTopologyAttrTypes(),
	}}
}

func elasticsearchTopologyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"instance_configuration_id": types.StringType,
		"size":                      types.StringType,
		"size_resource":             types.StringType,
		"zone_count":                types.Int64Type,
		"node_type_data":            types.BoolType,
		"node_type_master":          types.BoolType,
		"node_type_ingest":          types.BoolType,
		"node_type_ml":              types.BoolType,
		"node_roles":                types.SetType{ElemType: types.StringType},
		"autoscaling":               elasticsearchAutoscalingSchema(), // Optional Elasticsearch autoscaling settings, such a maximum and minimum size and resources.
	}
}

func elasticsearchAutoscalingSchema() attr.Type {
	return types.ListType{ElemType: types.ObjectType{
		AttrTypes: elasticsearchAutoscalingAttrTypes(),
	}}
}

func elasticsearchAutoscalingAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"max_size_resource":    types.StringType, // Maximum resource type for the maximum autoscaling setting.
		"max_size":             types.StringType, // Maximum size value for the maximum autoscaling setting.
		"min_size_resource":    types.StringType, // Minimum resource type for the minimum autoscaling setting.
		"min_size":             types.StringType, // Minimum size value for the minimum autoscaling setting.
		"policy_override_json": types.StringType, // Computed policy overrides set directly via the API or other clients.
	}
}

type elasticsearchResourceModelV0 struct {
	Autoscale     types.String `tfsdk:"autoscale"`
	Healthy       types.Bool   `tfsdk:"healthy"`
	CloudID       types.String `tfsdk:"cloud_id"`
	HttpEndpoint  types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint types.String `tfsdk:"https_endpoint"`
	RefID         types.String `tfsdk:"ref_id"`
	ResourceID    types.String `tfsdk:"resource_id"`
	Status        types.String `tfsdk:"status"`
	Version       types.String `tfsdk:"version"`
	Topology      types.List   `tfsdk:"topology"` //< elasticsearchTopologyModelV0
}

type elasticsearchTopologyModelV0 struct {
	InstanceConfigurationID types.String `tfsdk:"instance_configuration_id"`
	Size                    types.String `tfsdk:"size"`
	SizeResource            types.String `tfsdk:"size_resource"`
	ZoneCount               types.Int64  `tfsdk:"zone_count"`
	NodeTypeData            types.Bool   `tfsdk:"node_type_data"`
	NodeTypeMaster          types.Bool   `tfsdk:"node_type_master"`
	NodeTypeIngest          types.Bool   `tfsdk:"node_type_ingest"`
	NodeTypeMl              types.Bool   `tfsdk:"node_type_ml"`
	NodeRoles               types.Set    `tfsdk:"node_roles"`
	Autoscaling             types.List   `tfsdk:"autoscaling"` //< elasticsearchAutoscalingModel
}

type elasticsearchAutoscalingModel struct {
	MaxSizeResource    types.String `tfsdk:"max_size_resource"`
	MaxSize            types.String `tfsdk:"max_size"`
	MinSizeResource    types.String `tfsdk:"min_size_resource"`
	MinSize            types.String `tfsdk:"min_size"`
	PolicyOverrideJson types.String `tfsdk:"policy_override_json"`
}
