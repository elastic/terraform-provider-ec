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

func TestStringItems(t *testing.T) {
	type args struct {
		elems []string
	}
	tests := []struct {
		name       string
		args       args
		wantResult []interface{}
	}{
		{
			name: "empty list returns nil",
		},
		{
			name:       "populated list returns the results as []interface{}",
			args:       args{elems: []string{"some", "some-other", ""}},
			wantResult: []interface{}{"some", "some-other", ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := StringToItems(tt.args.elems...)
			assert.Equal(t, tt.wantResult, gotResult)
		})
	}
}

func TestItemsToString(t *testing.T) {
	type args struct {
		elems []interface{}
	}
	tests := []struct {
		name       string
		args       args
		wantResult []string
	}{
		{
			name: "empty list returns nil",
		},
		{
			name:       "populated list returns the results as []string{}",
			args:       args{elems: []interface{}{"some", "some-other", ""}},
			wantResult: []string{"", "some", "some-other"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := ItemsToString(tt.args.elems)
			assert.Equal(t, tt.wantResult, gotResult)
		})
	}
}
