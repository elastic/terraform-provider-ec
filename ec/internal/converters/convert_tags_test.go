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

package converters

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
)

func TestFlattenTags(t *testing.T) {
	type args struct {
		metadata *models.DeploymentMetadata
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "flattens no metadata tags when empty",
			args: args{metadata: &models.DeploymentMetadata{}},
			want: map[string]string{},
		},
		{
			name: "flatten metadata tags",
			args: args{metadata: &models.DeploymentMetadata{
				Tags: []*models.MetadataItem{
					{
						Key:   ec.String("foo"),
						Value: ec.String("bar"),
					},
				},
			}},
			want: map[string]string{"foo": "bar"},
		},
		{
			name: "flatten metadata tags",
			args: args{metadata: &models.DeploymentMetadata{
				Tags: []*models.MetadataItem{
					{
						Key:   ec.String("foo"),
						Value: ec.String("bar"),
					},
					{
						Key:   ec.String("bar"),
						Value: ec.String("baz"),
					},
				},
			}},
			want: map[string]string{"foo": "bar", "bar": "baz"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TagsToTypeMap(tt.args.metadata.Tags)
			got := make(map[string]string, len(result.Elems))
			result.ElementsAs(context.Background(), &got, false)
			assert.Equal(t, tt.want, got)
		})
	}
}
