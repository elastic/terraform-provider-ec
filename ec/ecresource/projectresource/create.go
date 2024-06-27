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

package projectresource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource[T]) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	if !resourceReady(r, &response.Diagnostics) {
		return
	}

	model, diags := r.modelHandler.ReadFrom(ctx, request.Plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if model == nil {
		// Diags
		return
	}

	createdModel, diags := r.api.Create(ctx, *model)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, createdModel)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(r.api.EnsureInitialised(ctx, createdModel)...)
	if response.Diagnostics.HasError() {
		return
	}

	found, createdModel, diags := r.api.Read(ctx, r.modelHandler.GetID(createdModel), createdModel)
	response.Diagnostics.Append(diags...)

	if !found {
		response.Diagnostics.AddError(
			fmt.Sprintf("Failed to read created %s project", r.name),
			fmt.Sprintf("The %s project was successfully created and initialised, but could then not be read back from the API", r.name),
		)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, createdModel)...)
}
