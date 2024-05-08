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

package deploymentresource

import (
	"context"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deptemplateapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/platformapi/instanceconfigapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	deploymentv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/deployment/v2"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.ResourceWithModifyPlan = &Resource{}

func (r Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		// Resource is being destroyed
		return
	}

	var plan deploymentv2.DeploymentTF
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	loadTemplate := func() (*models.DeploymentTemplateInfoV2, error) {
		return deptemplateapi.Get(deptemplateapi.GetParams{
			API:                        r.client,
			TemplateID:                 plan.DeploymentTemplateId.ValueString(),
			Region:                     plan.Region.ValueString(),
			HideInstanceConfigurations: false,
			ShowMaxZones:               true,
		})
	}

	loadInstanceConfig := func(id string, version *int64) (*models.InstanceConfiguration, error) {
		return instanceconfigapi.Get(instanceconfigapi.GetParams{
			API:           r.client,
			ID:            id,
			Region:        plan.Region.ValueString(),
			ShowDeleted:   true,
			ShowMaxZones:  true,
			ConfigVersion: version,
		})
	}

	planmodifiers.UpdateDedicatedMasterTier(ctx, req, resp, loadTemplate, loadInstanceConfig)
	if resp.Diagnostics.HasError() {
		return
	}
}
