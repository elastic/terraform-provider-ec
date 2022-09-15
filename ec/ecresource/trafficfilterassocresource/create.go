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
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/trafficfilterapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (t trafficFilterAssocResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var newState modelV0

	diags := request.Plan.Get(ctx, &newState)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := trafficfilterapi.CreateAssociation(trafficfilterapi.CreateAssociationParams{
		API:        t.provider.GetClient(),
		ID:         newState.TrafficFilterID.Value,
		EntityID:   newState.DeploymentID.Value,
		EntityType: entityTypeDeployment,
	}); err != nil {
		response.Diagnostics.AddError(err.Error(), err.Error())
		return
	}

	newState.ID = types.String{Value: fmt.Sprintf("%v-%v", newState.DeploymentID.Value, newState.TrafficFilterID.Value)}
	diags = response.State.Set(ctx, newState)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}
