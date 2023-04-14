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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/elastic/cloud-sdk-go/pkg/models"
)

type ElasticsearchSnapshot struct {
	Enabled    bool                                 `tfsdk:"enabled"`
	Repository *ElasticsearchSnapshotRepositoryInfo `tfsdk:"repository"`
}

type ElasticsearchSnapshotRepositoryInfo struct {
	Reference *ElasticsearchSnapshotRepositoryReference `tfsdk:"reference"`
}

type ElasticsearchSnapshotRepositoryReference struct {
	RepositoryName string `tfsdk:"repository_name"`
}

func readElasticsearchSnapshot(in *models.ElasticsearchClusterSettings) (*ElasticsearchSnapshot, error) {
	if in == nil || in.Snapshot == nil {
		return nil, nil
	}

	var snapshot ElasticsearchSnapshot

	if in.Snapshot.Enabled != nil {
		snapshot.Enabled = *in.Snapshot.Enabled
	}
	if in.Snapshot.Repository != nil {
		snapshot.Repository = &ElasticsearchSnapshotRepositoryInfo{}
		if in.Snapshot.Repository.Reference != nil {
			snapshot.Repository.Reference = &ElasticsearchSnapshotRepositoryReference{
				RepositoryName: in.Snapshot.Repository.Reference.RepositoryName,
			}
		}
	}

	return &snapshot, nil
}

func elasticsearchSnapshotPayload(ctx context.Context, srcObj attr.Value, model *models.ElasticsearchClusterSettings) (*models.ElasticsearchClusterSettings, diag.Diagnostics) {
	var snapshot *ElasticsearchSnapshot
	if srcObj.IsNull() || srcObj.IsUnknown() {
		return model, nil
	}

	if diags := tfsdk.ValueAs(ctx, srcObj, &snapshot); diags.HasError() {
		return model, diags
	}

	if model == nil {
		model = &models.ElasticsearchClusterSettings{}
	}
	model.Snapshot = &models.ClusterSnapshotSettings{
		Enabled:    &snapshot.Enabled,
		Repository: &models.ClusterSnapshotRepositoryInfo{},
	}

	if snapshot.Repository != nil && snapshot.Repository.Reference != nil {
		model.Snapshot.Repository.Reference = &models.ClusterSnapshotRepositoryReference{
			RepositoryName: snapshot.Repository.Reference.RepositoryName,
		}
	}

	return model, nil
}
