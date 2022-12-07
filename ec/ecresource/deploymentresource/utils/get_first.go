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

package utils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func GetFirst(ctx context.Context, list types.List, target any) diag.Diagnostics {
	if list.IsNull() || list.IsUnknown() || len(list.Elems) == 0 {
		return nil
	}

	if list.Elems[0].IsUnknown() || list.Elems[0].IsNull() {
		return nil
	}

	diags := tfsdk.ValueAs(ctx, list.Elems[0], target)

	if diags.HasError() {
		return diags
	}

	return nil
}
