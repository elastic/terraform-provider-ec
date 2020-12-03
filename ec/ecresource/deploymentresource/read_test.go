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
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/apierror"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/client/deployments"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/go-openapi/runtime"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_readResource(t *testing.T) {
	tc500Err := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleDeployment(),
		Schema: newSchema(),
	})
	wantTC500 := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleDeployment(),
		Schema: newSchema(),
	})

	tc404Err := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleDeployment(),
		Schema: newSchema(),
	})

	wantTC404 := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleDeployment(),
		Schema: newSchema(),
	})
	wantTC404.SetId("")

	tc200Stopped := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleDeployment(),
		Schema: newSchema(),
	})

	wantTC200Stopped := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleDeployment(),
		Schema: newSchema(),
	})
	wantTC200Stopped.SetId("")

	type args struct {
		ctx  context.Context
		d    *schema.ResourceData
		meta interface{}
	}
	tests := []struct {
		name   string
		args   args
		want   diag.Diagnostics
		wantRD *schema.ResourceData
	}{
		{
			name: "returns an error when it receives a 500",
			args: args{
				d: tc500Err,
				meta: api.NewMock(mock.NewErrorResponse(500, mock.APIError{
					Code: "some", Message: "message",
				})),
			},
			want: diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "failed reading deployment: 1 error occurred:\n\t* api error: some: message\n\n",
				},
			},
			wantRD: wantTC500,
		},
		{
			name: "returns nil and unsets the state when the error is known",
			args: args{
				d: tc404Err,
				meta: api.NewMock(mock.NewErrorResponse(404, mock.APIError{
					Code: "some", Message: "message",
				})),
			},
			want:   nil,
			wantRD: wantTC404,
		},
		{
			name: "returns nil and unsets the state when none of the deployment resources are running",
			args: args{
				d: tc200Stopped,
				meta: api.NewMock(mock.New200StructResponse(models.DeploymentGetResponse{
					Resources: &models.DeploymentResources{
						Elasticsearch: []*models.ElasticsearchResourceInfo{{
							Info: &models.ElasticsearchClusterInfo{Status: ec.String("stopped")},
						}},
					},
				})),
			},
			want:   nil,
			wantRD: wantTC200Stopped,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := readResource(tt.args.ctx, tt.args.d, tt.args.meta)
			assert.Equal(t, tt.want, got)
			var want interface{}
			if tt.wantRD != nil {
				if s := tt.wantRD.State(); s != nil {
					want = s.Attributes
				}
			}

			var gotState interface{}
			if s := tt.args.d.State(); s != nil {
				gotState = s.Attributes
			}

			assert.Equal(t, want, gotState)
		})
	}
}

func Test_deploymentNotFound(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "When the error is empty, it returns false",
		},
		{
			name: "When the error is something else (500), it returns false",
			args: args{
				err: &apierror.Error{Err: &runtime.APIError{Code: 500}},
			},
		},
		{
			name: "When the error is something else (401), it returns false",
			args: args{
				err: &apierror.Error{Err: &deployments.GetDeploymentUnauthorized{}},
			},
		},
		{
			name: "When the deployment is not found, it returns true",
			args: args{
				err: &apierror.Error{Err: &deployments.GetDeploymentNotFound{}},
			},
			want: true,
		},
		{
			name: "When the deployment is not authorized it returns false",
			args: args{
				err: &apierror.Error{Err: &runtime.APIError{Code: 403}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deploymentNotFound(tt.args.err); got != tt.want {
				t.Errorf("deploymentNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}
