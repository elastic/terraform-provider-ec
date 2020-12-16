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

package trafficfilterresource

import "testing"

func Test_trafficFilterRuleHash(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "hash a rule without description",
			args: args{v: map[string]interface{}{
				"source": "8.8.8.8/24",
			}},
			want: 1202035824,
		},
		{
			name: "hash a rule with description",
			args: args{v: map[string]interface{}{
				"source":      "8.8.8.8/24",
				"description": "google dns",
			}},
			want: 1579348650,
		},
		{
			name: "hash a rule different without description",
			args: args{v: map[string]interface{}{
				"source": "8.8.4.4/24",
			}},
			want: 2058478515,
		},
		{
			name: "hash a rule different with description",
			args: args{v: map[string]interface{}{
				"source":      "8.8.4.4/24",
				"description": "alternate google dns",
			}},
			want: 766352945,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trafficFilterRuleHash(tt.args.v); got != tt.want {
				t.Errorf("trafficFilterRuleHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
