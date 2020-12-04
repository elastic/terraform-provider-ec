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

package util

import (
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/apierror"
	"github.com/elastic/cloud-sdk-go/pkg/client/deployments_traffic_filter"
	"github.com/go-openapi/runtime"
)

func TestTrafficFilterNotFound(t *testing.T) {
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
				err: &apierror.Error{Err: &deployments_traffic_filter.GetTrafficFilterRulesetInternalServerError{}},
			},
		},
		{
			name: "When the deployment traffic filter rule is not found, it returns true",
			args: args{
				err: &apierror.Error{Err: &deployments_traffic_filter.GetTrafficFilterRulesetNotFound{}},
			},
			want: true,
		},
		{
			name: "When the deployment traffic filter rule is not found in ESS it returns true",
			args: args{
				err: &apierror.Error{Err: &runtime.APIError{Code: 403}},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrafficFilterNotFound(tt.args.err); got != tt.want {
				t.Errorf("TrafficFilterNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}
