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
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/models"
)

// flattenObservability parses a deployment's observability settings.
func flattenObservability(ctx context.Context, settings *models.DeploymentSettings) (types.List, diag.Diagnostics) {
	model := observabilitySettingsModel{
		Metrics: types.BoolValue(false),
		Logs:    types.BoolValue(false),
	}
	empty := true

	target := types.ListNull(
		types.ObjectType{
			AttrTypes: observabilitySettingsAttrTypes(),
		},
	)

	if settings == nil || settings.Observability == nil {
		return target, nil
	}

	// We are only accepting a single deployment ID and refID for both logs and metrics.
	// If either of them is not nil the deployment ID and refID will be filled.
	if settings.Observability.Metrics != nil {
		model.DeploymentID = types.StringValue(*settings.Observability.Metrics.Destination.DeploymentID)
		model.RefID = types.StringValue(settings.Observability.Metrics.Destination.RefID)
		model.Metrics = types.BoolValue(true)
		empty = false
	}

	if settings.Observability.Logging != nil {
		model.DeploymentID = types.StringValue(*settings.Observability.Logging.Destination.DeploymentID)
		model.RefID = types.StringValue(settings.Observability.Logging.Destination.RefID)
		model.Logs = types.BoolValue(true)
		empty = false
	}

	if empty {
		return target, nil
	}

	return types.ListValueFrom(ctx, target.ElementType(ctx), []observabilitySettingsModel{model})
}
