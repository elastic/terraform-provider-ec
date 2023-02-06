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
	topologyv1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/topology/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IntegrationsServerTF struct {
	ElasticsearchClusterRefId types.String `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     types.String `tfsdk:"ref_id"`
	ResourceId                types.String `tfsdk:"resource_id"`
	Region                    types.String `tfsdk:"region"`
	HttpEndpoint              types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint             types.String `tfsdk:"https_endpoint"`
	Topology                  types.List   `tfsdk:"topology"`
	Config                    types.List   `tfsdk:"config"`
}

type IntegrationsServer struct {
	ElasticsearchClusterRefId *string                   `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     *string                   `tfsdk:"ref_id"`
	ResourceId                *string                   `tfsdk:"resource_id"`
	Region                    *string                   `tfsdk:"region"`
	HttpEndpoint              *string                   `tfsdk:"http_endpoint"`
	HttpsEndpoint             *string                   `tfsdk:"https_endpoint"`
	Topology                  topologyv1.Topologies     `tfsdk:"topology"`
	Config                    IntegrationsServerConfigs `tfsdk:"config"`
}

type IntegrationsServers []IntegrationsServer
