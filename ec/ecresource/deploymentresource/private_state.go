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
	"encoding/json"
	"github.com/elastic/cloud-sdk-go/pkg/client/deployments"
	"github.com/elastic/cloud-sdk-go/pkg/models"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type PrivateState interface {
	GetKey(context.Context, string) ([]byte, diag.Diagnostics)
	SetKey(context.Context, string, []byte) diag.Diagnostics
}

const trafficFilterStateKey = "traffic_filters"
const migrationUpdateRequestKey = "migration_update_request"
const instanceConfigurationsKey = "instance_configurations"

func readPrivateStateTrafficFilters(ctx context.Context, state PrivateState) ([]string, diag.Diagnostics) {
	privateFilterBytes, diags := state.GetKey(ctx, trafficFilterStateKey)
	if privateFilterBytes == nil || diags.HasError() {
		return []string{}, diags
	}

	var privateFilters []string
	err := json.Unmarshal(privateFilterBytes, &privateFilters)
	if err != nil {
		diags.AddError("failed to parse private state", err.Error())
		return []string{}, diags
	}

	return privateFilters, diags
}

func updatePrivateStateTrafficFilters(ctx context.Context, state PrivateState, filters []string) diag.Diagnostics {
	var diags diag.Diagnostics
	filterBytes, err := json.Marshal(filters)
	if err != nil {
		diags.AddError("failed to update private state", err.Error())
		return diags
	}

	return state.SetKey(ctx, trafficFilterStateKey, filterBytes)
}

func ReadPrivateStateMigrateTemplateRequest(ctx context.Context, state PrivateState) (*deployments.MigrateDeploymentTemplateOK, diag.Diagnostics) {
	migrationUpdateRequestBytes, diags := state.GetKey(ctx, migrationUpdateRequestKey)
	if migrationUpdateRequestBytes == nil || diags.HasError() {
		return nil, diags
	}

	var migrationUpdateRequest models.DeploymentUpdateRequest
	err := migrationUpdateRequest.UnmarshalBinary(migrationUpdateRequestBytes)
	if err != nil {
		diags.AddError("failed to parse private state", err.Error())
		return nil, diags
	}

	migrateTemplateRequest := deployments.NewMigrateDeploymentTemplateOK()
	migrateTemplateRequest.Payload = &migrationUpdateRequest

	return migrateTemplateRequest, diags
}

func UpdatePrivateStateMigrateTemplateRequest(ctx context.Context, state PrivateState, migrateTemplateRequest *deployments.MigrateDeploymentTemplateOK) diag.Diagnostics {
	var diags diag.Diagnostics

	if migrateTemplateRequest == nil {
		return diags
	}

	migrationUpdateRequestBytes, err := migrateTemplateRequest.Payload.MarshalBinary()
	if err != nil {
		diags.AddError("failed to update private state", err.Error())
		return diags
	}

	return state.SetKey(ctx, migrationUpdateRequestKey, migrationUpdateRequestBytes)
}

func ReadPrivateStateInstanceConfigurations(
	ctx context.Context,
	state PrivateState,
) ([]models.InstanceConfigurationInfo, diag.Diagnostics) {
	data, diags := state.GetKey(ctx, instanceConfigurationsKey)
	if data == nil || diags.HasError() {
		return nil, diags
	}

	var instanceConfigurations []models.InstanceConfigurationInfo
	err := json.Unmarshal(data, &instanceConfigurations)
	if err != nil {
		diags.AddError("instance-configurations: failed to parse private state", err.Error())
		return nil, diags
	}

	return instanceConfigurations, diags
}

func UpdatePrivateStateInstanceConfigurations(
	ctx context.Context,
	state PrivateState,
	instanceConfigurations []*models.InstanceConfigurationInfo,
) diag.Diagnostics {
	var ics []models.InstanceConfigurationInfo
	for _, ic := range instanceConfigurations {
		if ic != nil {
			ics = append(ics, *ic)
		}
	}

	data, err := json.Marshal(ics)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("failed to update private state", err.Error())
		return diags
	}

	return state.SetKey(ctx, instanceConfigurationsKey, data)
}
