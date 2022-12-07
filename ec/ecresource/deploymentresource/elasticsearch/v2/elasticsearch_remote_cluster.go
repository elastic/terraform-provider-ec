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

package v2

import (
	"context"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchRemoteClusterTF struct {
	DeploymentId    types.String `tfsdk:"deployment_id"`
	Alias           types.String `tfsdk:"alias"`
	RefId           types.String `tfsdk:"ref_id"`
	SkipUnavailable types.Bool   `tfsdk:"skip_unavailable"`
}

type ElasticsearchRemoteCluster struct {
	DeploymentId    *string `tfsdk:"deployment_id"`
	Alias           *string `tfsdk:"alias"`
	RefId           *string `tfsdk:"ref_id"`
	SkipUnavailable *bool   `tfsdk:"skip_unavailable"`
}

type ElasticsearchRemoteClusters []ElasticsearchRemoteCluster

func ReadElasticsearchRemoteClusters(in []*models.RemoteResourceRef) (ElasticsearchRemoteClusters, error) {
	if len(in) == 0 {
		return nil, nil
	}

	clusters := make(ElasticsearchRemoteClusters, 0, len(in))

	for _, model := range in {
		cluster, err := ReadElasticsearchRemoteCluster(model)
		if err != nil {
			return nil, err
		}
		// clusters[*cluster.DeploymentId] = *cluster
		clusters = append(clusters, *cluster)
	}

	return clusters, nil
}

func ElasticsearchRemoteClustersPayload(ctx context.Context, clustersTF types.Set) (*models.RemoteResources, diag.Diagnostics) {
	payloads := models.RemoteResources{Resources: []*models.RemoteResourceRef{}}

	for _, elem := range clustersTF.Elems {
		var cluster ElasticsearchRemoteClusterTF
		diags := tfsdk.ValueAs(ctx, elem, &cluster)

		if diags.HasError() {
			return nil, diags
		}
		var payload models.RemoteResourceRef

		if !cluster.DeploymentId.IsNull() {
			payload.DeploymentID = &cluster.DeploymentId.Value
		}

		if !cluster.RefId.IsNull() {
			payload.ElasticsearchRefID = &cluster.RefId.Value
		}

		if !cluster.Alias.IsNull() {
			payload.Alias = &cluster.Alias.Value
		}

		if !cluster.SkipUnavailable.IsNull() {
			payload.SkipUnavailable = &cluster.SkipUnavailable.Value
		}

		payloads.Resources = append(payloads.Resources, &payload)
	}

	return &payloads, nil
}

func ReadElasticsearchRemoteCluster(in *models.RemoteResourceRef) (*ElasticsearchRemoteCluster, error) {
	var cluster ElasticsearchRemoteCluster

	if in.DeploymentID != nil && *in.DeploymentID != "" {
		cluster.DeploymentId = in.DeploymentID
	}

	if in.ElasticsearchRefID != nil && *in.ElasticsearchRefID != "" {
		cluster.RefId = in.ElasticsearchRefID
	}

	if in.Alias != nil && *in.Alias != "" {
		cluster.Alias = in.Alias
	}

	if in.SkipUnavailable != nil {
		cluster.SkipUnavailable = in.SkipUnavailable
	}

	return &cluster, nil
}
