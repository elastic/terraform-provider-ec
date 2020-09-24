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

package stackdatasource

import (
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/stretchr/testify/assert"
)

func Test_flattenKibanaResources(t *testing.T) {
	type args struct {
		res *models.StackVersionKibanaConfig
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "empty resource list returns empty list",
			args: args{},
			want: nil,
		},
		{
			name: "empty resource list returns empty list",
			args: args{res: &models.StackVersionKibanaConfig{}},
			want: nil,
		},
		{
			name: "parses the apm resource",
			args: args{res: &models.StackVersionKibanaConfig{
				Blacklist: []string{"some"},
				CapacityConstraints: &models.StackVersionInstanceCapacityConstraint{
					Max: ec.Int32(8192),
					Min: ec.Int32(512),
				},
				DockerImage: ec.String("docker.elastic.co/cloud-assets/kibana:7.9.1-0"),
			}},
			want: []interface{}{map[string]interface{}{
				"denylist":                 []interface{}{"some"},
				"capacity_constraints_max": 8192,
				"capacity_constraints_min": 512,
				"docker_image":             "docker.elastic.co/cloud-assets/kibana:7.9.1-0",
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := flattenKibanaResources(tt.args.res)
			assert.Equal(t, tt.want, got)
		})
	}
}
