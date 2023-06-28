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

package stackdatasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

// flattenApmConfig takes a StackVersionApmConfigs and flattens it.
func flattenApmConfig(ctx context.Context, res *models.StackVersionApmConfig) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	model := newResourceKindConfigModelV0()

	target := types.ListNull(resourceKindConfigSchema(util.ApmResourceKind).GetType().(types.ListType).ElemType)
	empty := true

	if res == nil {
		return target, nil
	}

	if len(res.Blacklist) > 0 {
		var d diag.Diagnostics
		model.DenyList, d = types.ListValueFrom(ctx, types.StringType, res.Blacklist)
		diags.Append(d...)
		empty = false
	}

	if res.CapacityConstraints != nil {
		model.CapacityConstraintsMax = types.Int64Value(int64(*res.CapacityConstraints.Max))
		model.CapacityConstraintsMin = types.Int64Value(int64(*res.CapacityConstraints.Min))
		empty = false
	}

	if len(res.CompatibleNodeTypes) > 0 {
		var d diag.Diagnostics
		model.CompatibleNodeTypes, d = types.ListValueFrom(ctx, types.StringType, res.CompatibleNodeTypes)
		diags.Append(d...)
		empty = false
	}

	if res.DockerImage != nil && *res.DockerImage != "" {
		model.DockerImage = types.StringValue(*res.DockerImage)
		empty = false
	}

	if empty {
		return target, diags
	}

	var d diag.Diagnostics
	target, d = types.ListValueFrom(ctx, target.ElementType(ctx), []resourceKindConfigModelV0{model})
	diags.Append(d...)

	return target, diags
}
