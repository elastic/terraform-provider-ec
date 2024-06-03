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

package elasticsearchprojectresource

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_elasticsearch_project"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	if !resourceReady(r, &response.Diagnostics) {
		return
	}

	var model resource_elasticsearch_project.ElasticsearchProjectModel
	response.Diagnostics.Append(request.State.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, err := r.client.DeleteElasticsearchProjectWithResponse(ctx, model.Id.ValueString(), nil)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete elasticsearch_project", err.Error())
	}

	statusCode := resp.StatusCode()
	if statusCode != 200 && statusCode != 404 {
		response.Diagnostics.AddError(
			"Request to delete elasticsearch_project failed",
			fmt.Sprintf("The API request failed with: %d %s\n%s",
				resp.StatusCode(),
				resp.Status(),
				resp.Body),
		)
		return
	}

	response.State.RemoveResource(ctx)
}
