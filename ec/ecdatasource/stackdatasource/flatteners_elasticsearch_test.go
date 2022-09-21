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
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_flattenElasticsearchResource(t *testing.T) {
	type args struct {
		res *models.StackVersionElasticsearchConfig
	}
	tests := []struct {
		name string
		args args
		want []elasticSearchConfigModelV0
	}{
		{
			name: "empty resource list returns empty list",
			args: args{},
			want: nil,
		},
		{
			name: "empty resource list returns empty list",
			args: args{res: &models.StackVersionElasticsearchConfig{}},
			want: nil,
		},
		{
			name: "parses the apm resource",
			args: args{res: &models.StackVersionElasticsearchConfig{
				Blacklist: []string{"some"},
				CapacityConstraints: &models.StackVersionInstanceCapacityConstraint{
					Max: ec.Int32(8192),
					Min: ec.Int32(512),
				},
				DockerImage:    ec.String("docker.elastic.co/cloud-assets/elasticsearch:7.9.1-0"),
				DefaultPlugins: []string{"repository-s3"},
				Plugins: []string{
					"analysis-icu",
					"analysis-kuromoji",
					"analysis-nori",
					"analysis-phonetic",
					"analysis-smartcn",
					"analysis-stempel",
					"analysis-ukrainian",
					"ingest-attachment",
					"mapper-annotated-text",
					"mapper-murmur3",
					"mapper-size",
					"repository-azure",
					"repository-gcs",
				},
			}},
			want: []elasticSearchConfigModelV0{{
				DenyList:               util.StringListAsType([]string{"some"}),
				CapacityConstraintsMax: types.Int64{Value: 8192},
				CapacityConstraintsMin: types.Int64{Value: 512},
				CompatibleNodeTypes:    util.StringListAsType(nil),
				DockerImage:            types.String{Value: "docker.elastic.co/cloud-assets/elasticsearch:7.9.1-0"},
				Plugins: util.StringListAsType([]string{
					"analysis-icu",
					"analysis-kuromoji",
					"analysis-nori",
					"analysis-phonetic",
					"analysis-smartcn",
					"analysis-stempel",
					"analysis-ukrainian",
					"ingest-attachment",
					"mapper-annotated-text",
					"mapper-murmur3",
					"mapper-size",
					"repository-azure",
					"repository-gcs",
				}),
				DefaultPlugins: util.StringListAsType([]string{"repository-s3"}),
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var newState modelV0
			diags := flattenStackVersionElasticsearchConfig(context.Background(), tt.args.res, &newState.Elasticsearch)
			assert.Empty(t, diags)

			var got []elasticSearchConfigModelV0
			newState.Elasticsearch.ElementsAs(context.Background(), &got, false)
			assert.Equal(t, tt.want, got)
		})
	}
}
