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
	"testing"
	"time"
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
			name: "hour old deployment is not stale",
			args: args{
				lastModified: time.Now().Add(-time.Minute * 59),
			},
		},
		{
			name: "10m old deployment is not stale",
			args: args{
				lastModified: time.Now().Add(-time.Minute * 10),
			},
		},
		{
			name: "hour old+ deployment is stale",
			args: args{
				lastModified: time.Now().Add(-time.Minute * 62),
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
