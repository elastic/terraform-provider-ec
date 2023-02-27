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
	v1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v1"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type ElasticsearchSnapshotSource v1.ElasticsearchSnapshotSource

func elasticsearchSnapshotSourcePayload(ctx context.Context, srcObj attr.Value, payload *models.ElasticsearchClusterPlan) diag.Diagnostics {
	var snapshot *v1.ElasticsearchSnapshotSourceTF

	if srcObj.IsNull() || srcObj.IsUnknown() {
		return nil
	}

	if diags := tfsdk.ValueAs(ctx, srcObj, &snapshot); diags.HasError() {
		return diags
	}

	if snapshot == nil {
		return nil
	}

	if payload.Transient == nil {
		payload.Transient = &models.TransientElasticsearchPlanConfiguration{
			RestoreSnapshot: &models.RestoreSnapshotConfiguration{},
		}
	}

	if !snapshot.SourceElasticsearchClusterId.IsNull() {
		payload.Transient.RestoreSnapshot.SourceClusterID = snapshot.SourceElasticsearchClusterId.Value
	}

	if !snapshot.SnapshotName.IsNull() {
		payload.Transient.RestoreSnapshot.SnapshotName = &snapshot.SnapshotName.Value
	}

	return nil
}
