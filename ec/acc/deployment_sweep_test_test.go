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

package acc

import (
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/go-openapi/runtime"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/apierror"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/client/deployments"
)

func Test_staleDeployment(t *testing.T) {
	type args struct {
		lastModified time.Time
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "deployment younger than stale window is not stale",
			args: args{
				lastModified: time.Now().Add(-deploymentStaleAfter + time.Minute),
			},
		},
		{
			name: "10m old deployment is not stale",
			args: args{
				lastModified: time.Now().Add(-time.Minute * 10),
			},
		},
		{
			name: "deployment older than stale window is stale",
			args: args{
				lastModified: time.Now().Add(-deploymentStaleAfter - time.Minute*2),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := staleDeployment(tt.args.lastModified); got != tt.want {
				t.Errorf("staleDeployment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_alreadyDestroyed(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil is not destroyed",
		},
		{
			name: "shutdown 404 is destroyed",
			err:  &deployments.ShutdownDeploymentNotFound{},
			want: true,
		},
		{
			name: "wrapped shutdown 404 is destroyed",
			err:  apierror.Wrap(&deployments.ShutdownDeploymentNotFound{}),
			want: true,
		},
		{
			name: "get 404 is destroyed",
			err:  &deployments.GetDeploymentNotFound{},
			want: true,
		},
		{
			name: "generic 404 is destroyed",
			err:  &runtime.APIError{Code: http.StatusNotFound},
			want: true,
		},
		{
			name: "403 is not destroyed",
			err:  &runtime.APIError{Code: http.StatusForbidden},
		},
		{
			name: "500 is not destroyed",
			err:  &runtime.APIError{Code: http.StatusInternalServerError},
		},
		{
			name: "unrelated error is not destroyed",
			err:  errors.New("boom"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := alreadyDestroyed(tt.err); got != tt.want {
				t.Errorf("alreadyDestroyed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_shutdownDeployment_notFound(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	err := shutdownDeployment(api.NewMock(mock.SampleNotFoundError()), mock.ValidClusterID, &wg)
	wg.Wait()
	if err != nil {
		t.Errorf("shutdownDeployment() not-found error = %v, want nil", err)
	}
}

func Test_shutdownDeployment_serverError(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	err := shutdownDeployment(api.NewMock(mock.SampleInternalError()), mock.ValidClusterID, &wg)
	wg.Wait()
	if err == nil {
		t.Fatal("shutdownDeployment() 500 error = nil, want error")
	}
}
