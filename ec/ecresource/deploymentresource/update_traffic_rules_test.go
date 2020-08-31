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

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func Test_getChange(t *testing.T) {
	type args struct {
		oldInterface interface{}
		newInterface interface{}
	}
	tests := []struct {
		name          string
		args          args
		wantAdditions []interface{}
		wantDeletions []interface{}
	}{
		{
			name: "diffs totally different slices",
			args: args{
				oldInterface: schema.NewSet(schema.HashString, []interface{}{
					"rule 1", "rule 2",
				}),
				newInterface: schema.NewSet(schema.HashString, []interface{}{
					"rule 3", "rule 4",
				}),
			},
			wantAdditions: []interface{}{"rule 4", "rule 3"},
			wantDeletions: []interface{}{"rule 1", "rule 2"},
		},
		{
			name: "diffs equal slices",
			args: args{
				oldInterface: schema.NewSet(schema.HashString, []interface{}{
					"rule 1", "rule 2",
				}),
				newInterface: schema.NewSet(schema.HashString, []interface{}{
					"rule 1", "rule 2",
				}),
			},
			wantAdditions: make([]interface{}, 0),
			wantDeletions: make([]interface{}, 0),
		},
		{
			name: "diffs equal slightly slices",
			args: args{
				oldInterface: schema.NewSet(schema.HashString, []interface{}{
					"rule 1", "rule 2",
				}),
				newInterface: schema.NewSet(schema.HashString, []interface{}{
					"rule 1", "rule 2", "rule 3",
				}),
			},
			wantAdditions: []interface{}{"rule 3"},
			wantDeletions: make([]interface{}, 0),
		},
		{
			name: "diffs a removal",
			args: args{
				newInterface: schema.NewSet(schema.HashString, nil),
				oldInterface: schema.NewSet(schema.HashString, []interface{}{
					"rule 1", "rule 2",
				}),
			},
			wantDeletions: []interface{}{"rule 1", "rule 2"},
			wantAdditions: make([]interface{}, 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAdditions, gotDeletions := getChange(tt.args.oldInterface, tt.args.newInterface)
			assert.Equal(t, tt.wantAdditions, gotAdditions.List(), "Additions")
			assert.Equal(t, tt.wantDeletions, gotDeletions.List(), "Deletions")
		})
	}
}
