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

package deploymentresource_test

import (
	"context"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/client/deployments"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type mockPrivateState struct {
	data map[string][]byte
}

func (m *mockPrivateState) GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics) {
	value, ok := m.data[key]
	if !ok {
		return nil, nil
	}
	return value, nil
}
func (m *mockPrivateState) SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics {
	m.data[key] = value
	return nil
}

func TestReadPrivateStateMigrateTemplateRequest(t *testing.T) {
	type args struct {
		ctx   context.Context
		state *mockPrivateState
	}

	tests := []struct {
		name         string
		args         args
		want         *deployments.MigrateDeploymentTemplateOK
		diagHasError bool
	}{
		{
			name: "reads valid migration request",
			args: args{
				ctx:   context.Background(),
				state: &mockPrivateState{data: map[string][]byte{"migration_update_request": getSampleMigrationRequestJson()}}},
			want:         getSampleMigrationRequest(),
			diagHasError: false,
		},
		{
			name: "reads invalid migration request",
			args: args{
				ctx:   context.Background(),
				state: &mockPrivateState{data: map[string][]byte{"migration_update_request": []byte("{invalid json}")}}},
			want:         nil,
			diagHasError: true,
		},
		{
			name: "reads non-existent migration request",
			args: args{
				ctx:   context.Background(),
				state: &mockPrivateState{data: map[string][]byte{}}},
			want:         nil,
			diagHasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, d := deploymentresource.ReadPrivateStateMigrateTemplateRequest(tt.args.ctx, tt.args.state)

			invalidDiags := d.HasError() != tt.diagHasError
			invalidReq := got != nil && tt.want != nil && (got.Payload.Name != tt.want.Payload.Name || got.Payload.Alias != tt.want.Payload.Alias)

			if invalidDiags || invalidReq {
				t.Errorf("ReadPrivateStateMigrateTemplateRequest() = (req = %v, d.HasError = %v), want (req = %v, d.HasError = %v)",
					got,
					d.HasError(),
					tt.want,
					tt.diagHasError)
			}
		})
	}
}

func TestUpdatePrivateStateMigrateTemplateRequest(t *testing.T) {
	migrationUpdateRequestKey := "migration_update_request"
	sampleMigrationReq := getSampleMigrationRequest()

	state := &mockPrivateState{
		data: make(map[string][]byte),
	}

	type args struct {
		ctx   context.Context
		state *mockPrivateState
		req   *deployments.MigrateDeploymentTemplateOK
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "store valid migration request",
			args: args{
				ctx:   context.Background(),
				state: state,
				req:   sampleMigrationReq,
			},
		},
		{
			name: "store empty migration request",
			args: args{
				ctx:   context.Background(),
				state: state,
				req: &deployments.MigrateDeploymentTemplateOK{
					Payload: nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploymentresource.UpdatePrivateStateMigrateTemplateRequest(tt.args.ctx, tt.args.state, tt.args.req)

			reqFromStateBytes, _ := tt.args.state.GetKey(tt.args.ctx, migrationUpdateRequestKey)

			expectedReq := string(getSampleMigrationRequestJson())
			reqFromState := string(reqFromStateBytes)

			if reqFromState == expectedReq {
				t.Errorf("ReadPrivateStateMigrateTemplateRequest() = (%v), want (%v)", reqFromState, expectedReq)
			}
		})
	}
}

func getSampleMigrationRequestJson() []byte {
	return []byte(`
	{
		"alias": "my-deployment-name",
		"name": "my_deployment_name"
	}
	`)
}

func getSampleMigrationRequest() *deployments.MigrateDeploymentTemplateOK {
	return &deployments.MigrateDeploymentTemplateOK{
		Payload: &models.DeploymentUpdateRequest{
			Name:  "my_deployment_name",
			Alias: "my-deployment-name",
		},
	}
}
