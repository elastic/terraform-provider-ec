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
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/models"
)

// flattenStackVersionApmConfig takes a StackVersionApmConfigs and flattens it.
func flattenStackVersionApmConfig(ctx context.Context, res *models.StackVersionApmConfig, target interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	model := newResourceKindConfigModelV0()
	empty := true

	if res == nil {
		return diags
	}

	if len(res.Blacklist) > 0 {
		diags.Append(tfsdk.ValueFrom(ctx, res.Blacklist, types.ListType{ElemType: types.StringType}, &model.DenyList)...)
		empty = false
	}

	if res.CapacityConstraints != nil {
		model.CapacityConstraintsMax = types.Int64{Value: int64(*res.CapacityConstraints.Max)}
		model.CapacityConstraintsMin = types.Int64{Value: int64(*res.CapacityConstraints.Min)}
		empty = false
	}

	if len(res.CompatibleNodeTypes) > 0 {
		diags.Append(tfsdk.ValueFrom(ctx, res.CompatibleNodeTypes, types.ListType{ElemType: types.StringType}, &model.CompatibleNodeTypes)...)
		empty = false
	}

	if res.DockerImage != nil && *res.DockerImage != "" {
		model.DockerImage = types.String{Value: *res.DockerImage}
		empty = false
	}

	if empty {
		return diags
	}

	diags.Append(tfsdk.ValueFrom(ctx, []resourceKindConfigModelV0{model}, types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: resourceKindConfigAttrTypes(Apm),
		},
	}, target)...)

	return diags
}
