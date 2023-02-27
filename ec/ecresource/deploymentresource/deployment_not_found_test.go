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
	"testing"

	"github.com/go-openapi/runtime"

	"github.com/elastic/cloud-sdk-go/pkg/api/apierror"
	"github.com/elastic/cloud-sdk-go/pkg/client/deployments"
)

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
			name: "When the deployment is not authorized it returns true, to account for the DR case (ESS)",
			args: args{
				err: &apierror.Error{Err: &runtime.APIError{Code: 403}},
			},
			want: true,
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
