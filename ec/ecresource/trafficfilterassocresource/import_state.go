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

package trafficfilterassocresource

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"strings"
)

func (t trafficFilterAssocResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	idParts := strings.Split(request.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: deployment_id,traffic_filter_id. Got: %q", request.ID),
		)
		return
	}
	deploymentId := idParts[0]
	trafficFilterId := idParts[1]

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), fmt.Sprintf("%v-%v", deploymentId, trafficFilterId))...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("deployment_id"), deploymentId)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("traffic_filter_id"), trafficFilterId)...)
}
