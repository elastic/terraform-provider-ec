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

func Test_flattenEnterpriseSearchResources(t *testing.T) {
	type args struct {
		res *models.StackVersionEnterpriseSearchConfig
	}
	tests := []struct {
		name string
		args args
		want []resourceKindConfigModelV0
	}{
		{
			name: "empty resource list returns empty list #1",
			args: args{},
			want: nil,
		},
		{
			name: "empty resource list returns empty list #2",
			args: args{res: &models.StackVersionEnterpriseSearchConfig{}},
			want: nil,
		},
		{
			name: "parses the apm resource",
			args: args{res: &models.StackVersionEnterpriseSearchConfig{
				Blacklist: []string{"some"},
				CapacityConstraints: &models.StackVersionInstanceCapacityConstraint{
					Max: ec.Int32(8192),
					Min: ec.Int32(512),
				},
				DockerImage: ec.String("docker.elastic.co/cloud-assets/enterprise_search:7.9.1-0"),
			}},
			want: []resourceKindConfigModelV0{{
				DenyList:               util.StringListAsType(t, []string{"some"}),
				CapacityConstraintsMax: types.Int64Value(8192),
				CapacityConstraintsMin: types.Int64Value(512),
				CompatibleNodeTypes:    util.StringListAsType(t, nil),
				DockerImage:            types.StringValue("docker.elastic.co/cloud-assets/enterprise_search:7.9.1-0"),
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enterpriseSearch, diags := flattenEnterpriseSearchConfig(context.Background(), tt.args.res)
			assert.Empty(t, diags)

			var got []resourceKindConfigModelV0
			enterpriseSearch.ElementsAs(context.Background(), &got, false)
			assert.Equal(t, tt.want, got)

			util.CheckConverionToAttrValue(t, &DataSource{}, "enterprise_search", enterpriseSearch)
		})
	}
}
