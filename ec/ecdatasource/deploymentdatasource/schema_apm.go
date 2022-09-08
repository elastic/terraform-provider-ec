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

func apmResourceInfoSchema() tfsdk.Attribute {
	// TODO should we use tfsdk.ListNestedAttributes here? - see https://github.com/hashicorp/terraform-provider-hashicups-pf/blob/8f222d805d39445673e442a674168349a45bc054/hashicups/data_source_coffee.go#L22
	return tfsdk.Attribute{
		Computed: true,
		Type: types.ListType{ElemType: types.ObjectType{
			AttrTypes: apmResourceInfoAttrTypes(),
		}},
	}
}

func apmResourceInfoAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"elasticsearch_cluster_ref_id": types.StringType,
		"healthy":                      types.BoolType,
		"http_endpoint":                types.StringType,
		"https_endpoint":               types.StringType,
		"ref_id":                       types.StringType,
		"resource_id":                  types.StringType,
		"status":                       types.StringType,
		"version":                      types.StringType,
		"topology":                     apmTopologySchema(),
	}
}
func apmTopologySchema() attr.Type {
	return types.ListType{ElemType: types.ObjectType{
		AttrTypes: apmTopologyAttrTypes(),
	}}
}

func apmTopologyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"instance_configuration_id": types.StringType,
		"size":                      types.StringType,
		"size_resource":             types.StringType,
		"zone_count":                types.Int64Type,
	}
}

type apmResourceModelV0 struct {
	ElasticsearchClusterRefID types.String `tfsdk:"elasticsearch_cluster_ref_id"`
	Healthy                   types.Bool   `tfsdk:"healthy"`
	HttpEndpoint              types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint             types.String `tfsdk:"https_endpoint"`
	RefID                     types.String `tfsdk:"ref_id"`
	ResourceID                types.String `tfsdk:"resource_id"`
	Status                    types.String `tfsdk:"status"`
	Version                   types.String `tfsdk:"version"`
	Topology                  types.List   `tfsdk:"topology"` //< apmTopologyModelV0
}

type apmTopologyModelV0 struct {
	InstanceConfigurationID types.String `tfsdk:"instance_configuration_id"`
	Size                    types.String `tfsdk:"size"`
	SizeResource            types.String `tfsdk:"size_resource"`
	ZoneCount               types.Int64  `tfsdk:"zone_count"`
}
