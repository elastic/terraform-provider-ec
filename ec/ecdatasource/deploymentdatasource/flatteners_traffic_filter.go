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
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/models"
)

// flattenTrafficFiltering parses a deployment's traffic filtering settings.
func flattenTrafficFiltering(ctx context.Context, settings *models.DeploymentSettings) (types.List, diag.Diagnostics) {
	target := types.List{ElemType: types.StringType}

	if settings == nil || settings.TrafficFilterSettings == nil {
		target.Null = true
		return target, nil
	}

	diags := tfsdk.ValueFrom(ctx, settings.TrafficFilterSettings.Rulesets, target.Type(ctx), &target)

	return target, diags
}
