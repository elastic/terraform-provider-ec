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
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/models"
)

type ElasticsearchSnapshot struct {
	Enabled    bool                                 `tfsdk:"enabled"`
	Repository *ElasticsearchSnapshotRepositoryInfo `tfsdk:"repository"`
}

type ElasticsearchSnapshotTF struct {
	Enabled    bool         `tfsdk:"enabled"`
	Repository types.Object `tfsdk:"repository"` //< ElasticsearchSnapshotRepositoryInfo
}

type ElasticsearchSnapshotRepositoryInfo struct {
	Reference *ElasticsearchSnapshotRepositoryReference `tfsdk:"reference"`
}

type ElasticsearchSnapshotRepositoryInfoTF struct {
	Reference types.Object `tfsdk:"reference"` //< ElasticsearchSnapshotRepositoryReference
}

type ElasticsearchSnapshotRepositoryReference struct {
	RepositoryName string `tfsdk:"repository_name"`
}

type ElasticsearchSnapshotRepositoryReferenceTF struct {
	RepositoryName types.String `tfsdk:"repository_name"`
}

func readElasticsearchSnapshot(in *models.ElasticsearchClusterSettings) (*ElasticsearchSnapshot, error) {
	if in == nil || in.Snapshot == nil {
		return nil, nil
	}

	var snapshot ElasticsearchSnapshot

	if in.Snapshot.Enabled != nil {
		snapshot.Enabled = *in.Snapshot.Enabled
	}

	if in.Snapshot.Repository == nil || in.Snapshot.Repository.Reference == nil {
		return &snapshot, nil
	}

	snapshot.Repository = &ElasticsearchSnapshotRepositoryInfo{
		Reference: &ElasticsearchSnapshotRepositoryReference{
			RepositoryName: in.Snapshot.Repository.Reference.RepositoryName,
		},
	}

	return &snapshot, nil
}

func elasticsearchSnapshotPayload(ctx context.Context, srcObj attr.Value, model *models.ElasticsearchClusterSettings, state *ElasticsearchTF) (*models.ElasticsearchClusterSettings, diag.Diagnostics) {
	var snapshot ElasticsearchSnapshotTF
	if srcObj.IsNull() || srcObj.IsUnknown() {
		return model, nil
	}

	// Only put snapshot updates into the payload, if the plan is making changes to the snapshot settings
	// (To avoid overwriting changes made outside, i.e. with the elasticstack provider)
	if state != nil && state.Snapshot.Equal(srcObj) {
		return model, nil
	}

	if diags := tfsdk.ValueAs(ctx, srcObj, &snapshot); diags.HasError() {
		return model, diags
	}

	if model == nil {
		model = &models.ElasticsearchClusterSettings{}
	}
	model.Snapshot = &models.ClusterSnapshotSettings{
		Enabled: &snapshot.Enabled,
	}

	if snapshot.Repository.IsNull() || snapshot.Repository.IsUnknown() {
		return model, nil
	}

	var repo ElasticsearchSnapshotRepositoryInfoTF
	if diags := tfsdk.ValueAs(ctx, snapshot.Repository, &repo); diags.HasError() {
		return model, diags
	}

	if repo.Reference.IsNull() || repo.Reference.IsUnknown() {
		return model, nil
	}

	var reference ElasticsearchSnapshotRepositoryReferenceTF
	if diags := tfsdk.ValueAs(ctx, repo.Reference, &reference); diags.HasError() {
		return model, diags
	}

	model.Snapshot.Repository = &models.ClusterSnapshotRepositoryInfo{
		Reference: &models.ClusterSnapshotRepositoryReference{
			RepositoryName: reference.RepositoryName.ValueString(),
		},
	}

	return model, nil
}
