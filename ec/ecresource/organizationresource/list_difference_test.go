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

package organizationresource

import (
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"reflect"
	"testing"
)

func Test_difference(t *testing.T) {
	type args struct {
		a []*string
		b []*string
	}
	type testCase struct {
		name string
		args args
		want []*string
	}
	tests := []testCase{
		{
			name: "returns elements that are only in a",
			args: args{
				a: []*string{ec.String("a"), ec.String("b")},
				b: []*string{ec.String("b"), ec.String("c")},
			},
			want: []*string{ec.String("a")},
		},
		{
			name: "both lists empty, returns empty list",
			args: args{
				a: nil,
				b: nil,
			},
			want: nil,
		},
		{
			name: "if b empty, returns a",
			args: args{
				a: []*string{ec.String("a")},
				b: nil,
			},
			want: []*string{ec.String("a")},
		},
		{
			name: "if b has no elements in a, return a",
			args: args{
				a: []*string{ec.String("a")},
				b: []*string{ec.String("b")},
			},
			want: []*string{ec.String("a")},
		},
		{
			name: "if b has all elements of a, return empty list",
			args: args{
				a: []*string{ec.String("a")},
				b: []*string{ec.String("a"), ec.String("b")},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		getKey := func(a string) string {
			return a
		}
		t.Run(tt.name, func(t *testing.T) {
			if got := difference(tt.args.a, tt.args.b, getKey); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("difference() = %v, want %v", got, tt.want)
			}
		})
	}
}
