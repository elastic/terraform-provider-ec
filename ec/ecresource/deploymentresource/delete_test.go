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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_deleteResource(t *testing.T) {
	tc500Err := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleLegacyDeployment(),
		Schema: newSchema(),
	})
	wantTC500 := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleLegacyDeployment(),
		Schema: newSchema(),
	})

	tc404Err := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleLegacyDeployment(),
		Schema: newSchema(),
	})
	wantTC404 := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleLegacyDeployment(),
		Schema: newSchema(),
	})
	wantTC404.SetId("")

	type args struct {
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
					Summary:  "failed shutting down the deployment: 1 error occurred:\n\t* api error: some: message\n\n",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deleteResource(context.Background(), tt.args.d, tt.args.meta)
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

func Test_shouldRetryShutdown(t *testing.T) {
	type args struct {
		err        error
		retries    int
		maxRetries int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "returns false when error doesn't contain timeout string",
			args: args{
				err:        errors.New("some error"),
				retries:    1,
				maxRetries: 10,
			},
			want: false,
		},
		{
			name: "returns false when the error is nil",
			args: args{
				retries:    1,
				maxRetries: 10,
			},
			want: false,
		},
		{
			name: "returns false when error doesn't contain timeout string",
			args: args{
				err:        errors.New("timeout exceeded"),
				retries:    1,
				maxRetries: 10,
			},
			want: false,
		},
		{
			name: "returns true when error contains timeout string",
			args: args{
				err:        errors.New("Timeout exceeded"),
				retries:    1,
				maxRetries: 10,
			},
			want: true,
		},
		{
			name: "returns true when error contains timeout string",
			args: args{
				err: multierror.NewPrefixed("aa",
					errors.New("Timeout exceeded"),
				),
				retries:    1,
				maxRetries: 10,
			},
			want: true,
		},
		{
			name: "returns true when error contains a deallocation failure string",
			args: args{
				err: multierror.NewPrefixed("aa",
					errors.New(`deployment [8f3c85f97536163ad117a6d37b377120] - [elasticsearch][39dd873845bc43f9b3b21b87fe1a3c99]: caught error: "Plan change failed: Some instances were not stopped`),
				),
				retries:    1,
				maxRetries: 10,
			},
			want: true,
		},
		{
			name: "returns false when error contains timeout string but exceeds max timeouts",
			args: args{
				err:        errors.New("Timeout exceeded"),
				retries:    10,
				maxRetries: 10,
			},
			want: false,
		},
		{
			name: "returns false when error contains a deallocation failure string",
			args: args{
				err: multierror.NewPrefixed("aa",
					errors.New(`deployment [8f3c85f97536163ad117a6d37b377120] - [elasticsearch][39dd873845bc43f9b3b21b87fe1a3c99]: caught error: "Plan change failed: Some instances were not stopped`),
				),
				retries:    10,
				maxRetries: 10,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldRetryShutdown(tt.args.err, tt.args.retries, tt.args.maxRetries)
			assert.Equal(t, tt.want, got)
		})
	}
}
