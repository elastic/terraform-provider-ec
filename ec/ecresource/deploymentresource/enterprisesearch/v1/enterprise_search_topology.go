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

type EnterpriseSearchTopologyTF struct {
	InstanceConfigurationId types.String `tfsdk:"instance_configuration_id"`
	Size                    types.String `tfsdk:"size"`
	SizeResource            types.String `tfsdk:"size_resource"`
	ZoneCount               types.Int64  `tfsdk:"zone_count"`
	NodeTypeAppserver       types.Bool   `tfsdk:"node_type_appserver"`
	NodeTypeConnector       types.Bool   `tfsdk:"node_type_connector"`
	NodeTypeWorker          types.Bool   `tfsdk:"node_type_worker"`
}

type EnterpriseSearchTopology struct {
	InstanceConfigurationId *string `tfsdk:"instance_configuration_id"`
	Size                    *string `tfsdk:"size"`
	SizeResource            *string `tfsdk:"size_resource"`
	ZoneCount               int     `tfsdk:"zone_count"`
	NodeTypeAppserver       *bool   `tfsdk:"node_type_appserver"`
	NodeTypeConnector       *bool   `tfsdk:"node_type_connector"`
	NodeTypeWorker          *bool   `tfsdk:"node_type_worker"`
}

type EnterpriseSearchTopologies []EnterpriseSearchTopology
