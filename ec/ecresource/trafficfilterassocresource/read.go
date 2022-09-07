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
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/trafficfilterapi"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (t trafficFilterAssocResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state modelV0

	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	res, err := trafficfilterapi.Get(trafficfilterapi.GetParams{
		API:                 t.provider.GetClient(),
		ID:                  state.TrafficFilterID.Value,
		IncludeAssociations: true,
	})
	if err != nil {
		if util.TrafficFilterNotFound(err) {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError(err.Error(), err.Error())
		return
	}

	if res == nil {
		response.State.RemoveResource(ctx)
		return
	}

	var found bool
	for _, assoc := range res.Associations {
		if *assoc.EntityType == entityTypeDeployment && *assoc.ID == state.DeploymentID.Value {
			found = true
		}
	}

	if !found {
		response.State.RemoveResource(ctx)
		return
	}
}
