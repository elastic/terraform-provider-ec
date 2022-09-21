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
	"errors"
	"fmt"
	"regexp/syntax"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_modelToState(t *testing.T) {
	state := modelV0{
		Region:       types.String{Value: "us-east-1"},
		VersionRegex: types.String{Value: "latest"},
	}

	type args struct {
		state modelV0
		res   *models.StackVersionConfig
	}
	tests := []struct {
		name string
		args args
		want modelV0
		err  error
	}{
		{
			name: "flattens stack resources",
			want: newSampleStack(),
			args: args{
				state: state,
				res: &models.StackVersionConfig{
					Version:           "7.9.1",
					Accessible:        ec.Bool(true),
					Whitelisted:       ec.Bool(true),
					MinUpgradableFrom: "6.8.0",
					Apm: &models.StackVersionApmConfig{
						Blacklist: []string{"some"},
						CapacityConstraints: &models.StackVersionInstanceCapacityConstraint{
							Max: ec.Int32(8192),
							Min: ec.Int32(512),
						},
						DockerImage: ec.String("docker.elastic.co/cloud-assets/apm:7.9.1-0"),
					},
					Kibana: &models.StackVersionKibanaConfig{
						Blacklist: []string{"some"},
						CapacityConstraints: &models.StackVersionInstanceCapacityConstraint{
							Max: ec.Int32(8192),
							Min: ec.Int32(512),
						},
						DockerImage: ec.String("docker.elastic.co/cloud-assets/kibana:7.9.1-0"),
					},
					Elasticsearch: &models.StackVersionElasticsearchConfig{
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
					},
					EnterpriseSearch: &models.StackVersionEnterpriseSearchConfig{
						Blacklist: []string{"some"},
						CapacityConstraints: &models.StackVersionInstanceCapacityConstraint{
							Max: ec.Int32(8192),
							Min: ec.Int32(512),
						},
						DockerImage: ec.String("docker.elastic.co/cloud-assets/enterprise_search:7.9.1-0"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state = tt.args.state
			diags := modelToState(context.Background(), tt.args.res, &state)
			assert.Empty(t, diags)

			assert.Equal(t, tt.want, state)
		})
	}
}

func newSampleStack() modelV0 {
	return modelV0{
		ID:                types.String{Value: "7.9.1"},
		Region:            types.String{Value: "us-east-1"},
		Version:           types.String{Value: "7.9.1"},
		VersionRegex:      types.String{Value: "latest"},
		Accessible:        types.Bool{Value: true},
		AllowListed:       types.Bool{Value: true},
		MinUpgradableFrom: types.String{Value: "6.8.0"},
		Elasticsearch: types.List{
			ElemType: types.ObjectType{
				AttrTypes: elasticSearchConfigAttrTypes(),
			},
			Elems: []attr.Value{types.Object{
				AttrTypes: elasticSearchConfigAttrTypes(),
				Attrs: map[string]attr.Value{
					"denylist":                 util.StringListAsType([]string{"some"}),
					"capacity_constraints_max": types.Int64{Value: 8192},
					"capacity_constraints_min": types.Int64{Value: 512},
					"compatible_node_types":    util.StringListAsType(nil),
					"docker_image":             types.String{Value: "docker.elastic.co/cloud-assets/elasticsearch:7.9.1-0"},
					"plugins": util.StringListAsType([]string{
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
					"default_plugins": util.StringListAsType([]string{"repository-s3"}),
				},
			}},
		},
		Kibana: types.List{
			ElemType: types.ObjectType{
				AttrTypes: resourceKindConfigAttrTypes(Kibana),
			},
			Elems: []attr.Value{types.Object{
				AttrTypes: resourceKindConfigAttrTypes(Kibana),
				Attrs: map[string]attr.Value{
					"denylist":                 util.StringListAsType([]string{"some"}),
					"capacity_constraints_max": types.Int64{Value: 8192},
					"capacity_constraints_min": types.Int64{Value: 512},
					"compatible_node_types":    util.StringListAsType(nil),
					"docker_image":             types.String{Value: "docker.elastic.co/cloud-assets/kibana:7.9.1-0"},
				},
			}},
		},
		EnterpriseSearch: types.List{
			ElemType: types.ObjectType{
				AttrTypes: resourceKindConfigAttrTypes(EnterpriseSearch),
			},
			Elems: []attr.Value{types.Object{
				AttrTypes: resourceKindConfigAttrTypes(EnterpriseSearch),
				Attrs: map[string]attr.Value{
					"denylist":                 util.StringListAsType([]string{"some"}),
					"capacity_constraints_max": types.Int64{Value: 8192},
					"capacity_constraints_min": types.Int64{Value: 512},
					"compatible_node_types":    util.StringListAsType(nil),
					"docker_image":             types.String{Value: "docker.elastic.co/cloud-assets/enterprise_search:7.9.1-0"},
				},
			}},
		},
		Apm: types.List{
			ElemType: types.ObjectType{
				AttrTypes: resourceKindConfigAttrTypes(Apm),
			},
			Elems: []attr.Value{types.Object{
				AttrTypes: resourceKindConfigAttrTypes(Apm),
				Attrs: map[string]attr.Value{
					"denylist":                 util.StringListAsType([]string{"some"}),
					"capacity_constraints_max": types.Int64{Value: 8192},
					"capacity_constraints_min": types.Int64{Value: 512},
					"compatible_node_types":    util.StringListAsType(nil),
					"docker_image":             types.String{Value: "docker.elastic.co/cloud-assets/apm:7.9.1-0"},
				},
			}},
		},
	}
}

func Test_stackFromFilters(t *testing.T) {
	var stackPacks = []*models.StackVersionConfig{
		{Version: "7.9.1"},
		{Version: "7.9.0"},
		{Version: "7.8.1"},
		{Version: "7.8.0"},
	}
	type args struct {
		expr    string
		version string
		locked  bool
		stacks  []*models.StackVersionConfig
	}
	tests := []struct {
		name string
		args args
		want *models.StackVersionConfig
		err  error
	}{
		{
			name: "returns the stack pack with exact matching",
			args: args{expr: "7.9.0", stacks: stackPacks},
			want: &models.StackVersionConfig{Version: "7.9.0"},
		},
		{
			name: "returns the stack pack with patch regex",
			args: args{expr: "7.8.?", stacks: stackPacks},
			want: &models.StackVersionConfig{Version: "7.8.1"},
		},
		{
			name: "returns the latest stackpack",
			args: args{expr: "latest", stacks: stackPacks},
			want: &models.StackVersionConfig{Version: "7.9.1"},
		},
		{
			name: "returns the latest stackpack with a locked version",
			args: args{
				expr:    "latest",
				stacks:  stackPacks,
				locked:  true,
				version: "7.8.1",
			},
			want: &models.StackVersionConfig{Version: "7.8.1"},
		},
		{
			name: "returns an error when the expression doesn't match the stackpack",
			args: args{expr: "7.9.1", stacks: []*models.StackVersionConfig{
				{Version: "7.8.0"},
			}},
			err: errors.New(`failed to obtain a stack version matching "7.9.1": please specify a valid version_regex`),
		},
		{
			name: "returns an error when the regex can't be compiled",
			args: args{expr: `(?!`, stacks: []*models.StackVersionConfig{
				{Version: "7.8.0"},
			}},
			err: fmt.Errorf("failed to compile the version_regex: %w", &syntax.Error{
				Expr: `(?!`,
				Code: syntax.ErrInvalidPerlOp,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := stackFromFilters(tt.args.expr, tt.args.version, tt.args.locked, tt.args.stacks)
			if !assert.Equal(t, tt.err, err) {
				fmt.Println(err, "!= want ", tt.err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
