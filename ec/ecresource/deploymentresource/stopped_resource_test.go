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

	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
)

func Test_isApmResourceStopped(t *testing.T) {
	type args struct {
		res *models.ApmResourceInfo
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "started resource returns false",
			args: args{res: &models.ApmResourceInfo{Info: &models.ApmInfo{
				Status: ec.String("started"),
			}}},
			want: false,
		},
		{
			name: "stopped resource returns true",
			args: args{res: &models.ApmResourceInfo{Info: &models.ApmInfo{
				Status: ec.String("stopped"),
			}}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isApmResourceStopped(tt.args.res)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_isEsResourceStopped(t *testing.T) {
	type args struct {
		res *models.ElasticsearchResourceInfo
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "started resource returns false",
			args: args{res: &models.ElasticsearchResourceInfo{Info: &models.ElasticsearchClusterInfo{
				Status: ec.String("started"),
			}}},
			want: false,
		},
		{
			name: "stopped resource returns true",
			args: args{res: &models.ElasticsearchResourceInfo{Info: &models.ElasticsearchClusterInfo{
				Status: ec.String("stopped"),
			}}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isEsResourceStopped(tt.args.res)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_isEssResourceStopped(t *testing.T) {
	type args struct {
		res *models.EnterpriseSearchResourceInfo
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "started resource returns false",
			args: args{res: &models.EnterpriseSearchResourceInfo{Info: &models.EnterpriseSearchInfo{
				Status: ec.String("started"),
			}}},
			want: false,
		},
		{
			name: "stopped resource returns true",
			args: args{res: &models.EnterpriseSearchResourceInfo{Info: &models.EnterpriseSearchInfo{
				Status: ec.String("stopped"),
			}}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isEssResourceStopped(tt.args.res)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_isKibanaResourceStopped(t *testing.T) {
	type args struct {
		res *models.KibanaResourceInfo
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "started resource returns false",
			args: args{res: &models.KibanaResourceInfo{Info: &models.KibanaClusterInfo{
				Status: ec.String("started"),
			}}},
			want: false,
		},
		{
			name: "stopped resource returns true",
			args: args{res: &models.KibanaResourceInfo{Info: &models.KibanaClusterInfo{
				Status: ec.String("stopped"),
			}}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isKibanaResourceStopped(tt.args.res)
			assert.Equal(t, tt.want, got)
		})
	}
}
