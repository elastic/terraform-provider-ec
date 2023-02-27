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

type ElasticsearchTF struct {
	Autoscale      types.String `tfsdk:"autoscale"`
	RefId          types.String `tfsdk:"ref_id"`
	ResourceId     types.String `tfsdk:"resource_id"`
	Region         types.String `tfsdk:"region"`
	CloudID        types.String `tfsdk:"cloud_id"`
	HttpEndpoint   types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint  types.String `tfsdk:"https_endpoint"`
	Topology       types.List   `tfsdk:"topology"`
	Config         types.List   `tfsdk:"config"`
	RemoteCluster  types.Set    `tfsdk:"remote_cluster"`
	SnapshotSource types.List   `tfsdk:"snapshot_source"`
	Extension      types.Set    `tfsdk:"extension"`
	TrustAccount   types.Set    `tfsdk:"trust_account"`
	TrustExternal  types.Set    `tfsdk:"trust_external"`
	Strategy       types.List   `tfsdk:"strategy"`
}

type Elasticsearch struct {
	Autoscale      *string                      `tfsdk:"autoscale"`
	RefId          *string                      `tfsdk:"ref_id"`
	ResourceId     *string                      `tfsdk:"resource_id"`
	Region         *string                      `tfsdk:"region"`
	CloudID        *string                      `tfsdk:"cloud_id"`
	HttpEndpoint   *string                      `tfsdk:"http_endpoint"`
	HttpsEndpoint  *string                      `tfsdk:"https_endpoint"`
	Topology       ElasticsearchTopologies      `tfsdk:"topology"`
	Config         ElasticsearchConfigs         `tfsdk:"config"`
	RemoteCluster  ElasticsearchRemoteClusters  `tfsdk:"remote_cluster"`
	SnapshotSource ElasticsearchSnapshotSources `tfsdk:"snapshot_source"`
	Extension      ElasticsearchExtensions      `tfsdk:"extension"`
	TrustAccount   ElasticsearchTrustAccounts   `tfsdk:"trust_account"`
	TrustExternal  ElasticsearchTrustExternals  `tfsdk:"trust_external"`
	Strategy       ElasticsearchStrategies      `tfsdk:"strategy"`
}

type Elasticsearches []Elasticsearch
