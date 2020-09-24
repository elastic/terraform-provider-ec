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

	"github.com/stretchr/testify/assert"
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
