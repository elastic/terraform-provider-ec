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

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestObjectRemoved(t *testing.T) {
	schemaMap := map[string]*schema.Schema{
		"object": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
	type args struct {
		d   *schema.ResourceData
		key string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "removes an object",
			args: args{
				key: "object",
				d: NewResourceData(t, ResDataParams{
					ID:     "id",
					Schema: schemaMap,
					State: map[string]interface{}{
						"object": []interface{}{"a", "b"},
					},
					Change: map[string]interface{}{
						"object": []interface{}{},
					},
				}),
			},
			want: true,
		},
		{
			name: "does not remove an object",
			args: args{
				key: "object",
				d: NewResourceData(t, ResDataParams{
					ID:     "id",
					Schema: schemaMap,
					State: map[string]interface{}{
						"object": []interface{}{"a", "b"},
					},
					Change: map[string]interface{}{
						"object": []interface{}{"b"},
					},
				}),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ObjectRemoved(tt.args.d, tt.args.key); got != tt.want {
				t.Errorf("ObjectRemoved() = %v, want %v", got, tt.want)
			}
		})
	}
}
