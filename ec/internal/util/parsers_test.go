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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
)

func TestMemoryToState(t *testing.T) {
	type args struct {
		mem int32
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "gigabytes",
			args: args{mem: 4096},
			want: "4g",
		},
		{
			name: "512 megabytes turns into 0.5g",
			args: args{mem: 512},
			want: "0.5g",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MemoryToState(tt.args.mem)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseTopologySize(t *testing.T) {
	type args struct {
		topology map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want *models.TopologySize
		err  error
	}{
		{
			name: "has no size returns nil",
		},
		{
			name: "has empty size returns nil",
			args: args{topology: map[string]interface{}{
				"size": "",
			}},
		},
		{
			name: "has badly formatted size returns error",
			args: args{topology: map[string]interface{}{
				"size": "asdasd",
			}},
			err: errors.New(`failed to convert "asdasd" to <size><g>`),
		},
		{
			name: "has size but no size_resource",
			args: args{topology: map[string]interface{}{
				"size": "15g",
			}},
			want: &models.TopologySize{
				Value:    ec.Int32(15360),
				Resource: ec.String("memory"),
			},
		},
		{
			name: "has size and explicit size_resource (memory)",
			args: args{topology: map[string]interface{}{
				"size":          "8g",
				"size_resource": "memory",
			}},
			want: &models.TopologySize{
				Value:    ec.Int32(8192),
				Resource: ec.String("memory"),
			},
		},
		{
			name: "has size and explicit size_resource (storage)",
			args: args{topology: map[string]interface{}{
				"size":          "4g",
				"size_resource": "storage",
			}},
			want: &models.TopologySize{
				Value:    ec.Int32(4096),
				Resource: ec.String("storage"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTopologySize(tt.args.topology)
			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
