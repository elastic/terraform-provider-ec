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

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/stretchr/testify/assert"
)

func TestIsApmResourceStopped(t *testing.T) {
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
			got := IsApmResourceStopped(tt.args.res)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsEsResourceStopped(t *testing.T) {
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
			got := IsEsResourceStopped(tt.args.res)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsEssResourceStopped(t *testing.T) {
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
			got := IsEssResourceStopped(tt.args.res)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsKibanaResourceStopped(t *testing.T) {
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
			got := IsKibanaResourceStopped(tt.args.res)
			assert.Equal(t, tt.want, got)
		})
	}
}
