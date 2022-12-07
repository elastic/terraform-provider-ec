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

package v1

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchTopologyTF struct {
	Id                      types.String `tfsdk:"id"`
	InstanceConfigurationId types.String `tfsdk:"instance_configuration_id"`
	Size                    types.String `tfsdk:"size"`
	SizeResource            types.String `tfsdk:"size_resource"`
	ZoneCount               types.Int64  `tfsdk:"zone_count"`
	NodeTypeData            types.String `tfsdk:"node_type_data"`
	NodeTypeMaster          types.String `tfsdk:"node_type_master"`
	NodeTypeIngest          types.String `tfsdk:"node_type_ingest"`
	NodeTypeMl              types.String `tfsdk:"node_type_ml"`
	NodeRoles               types.Set    `tfsdk:"node_roles"`
	Autoscaling             types.List   `tfsdk:"autoscaling"`
	Config                  types.List   `tfsdk:"config"`
}

type ElasticsearchTopology struct {
	Id                      string                            `tfsdk:"id"`
	InstanceConfigurationId *string                           `tfsdk:"instance_configuration_id"`
	Size                    *string                           `tfsdk:"size"`
	SizeResource            *string                           `tfsdk:"size_resource"`
	ZoneCount               int                               `tfsdk:"zone_count"`
	NodeTypeData            *string                           `tfsdk:"node_type_data"`
	NodeTypeMaster          *string                           `tfsdk:"node_type_master"`
	NodeTypeIngest          *string                           `tfsdk:"node_type_ingest"`
	NodeTypeMl              *string                           `tfsdk:"node_type_ml"`
	NodeRoles               []string                          `tfsdk:"node_roles"`
	Autoscaling             ElasticsearchTopologyAutoscalings `tfsdk:"autoscaling"`
	Config                  ElasticsearchTopologyConfigs      `tfsdk:"config"`
}

type ElasticsearchTopologies []ElasticsearchTopology
