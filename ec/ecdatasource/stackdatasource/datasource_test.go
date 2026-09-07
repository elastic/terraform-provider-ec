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

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_modelToState(t *testing.T) {
	state := modelV0{
		Region:       types.StringValue("us-east-1"),
		VersionRegex: types.StringValue("latest"),
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
			want: newSampleStack(t),
			args: args{
				state: state,
				res: &models.StackVersionConfig{
					Version:           "7.9.1",
					Accessible:        new(true),
					Whitelisted:       new(true),
					MinUpgradableFrom: "6.8.0",
					Apm: &models.StackVersionApmConfig{
						Blacklist: []string{"some"},
						CapacityConstraints: &models.StackVersionInstanceCapacityConstraint{
							Max: ec.Int32(8192),
							Min: ec.Int32(512),
						},
						DockerImage: new("docker.elastic.co/cloud-assets/apm:7.9.1-0"),
					},
					Kibana: &models.StackVersionKibanaConfig{
						Blacklist: []string{"some"},
						CapacityConstraints: &models.StackVersionInstanceCapacityConstraint{
							Max: ec.Int32(8192),
							Min: ec.Int32(512),
						},
						DockerImage: new("docker.elastic.co/cloud-assets/kibana:7.9.1-0"),
					},
					Elasticsearch: &models.StackVersionElasticsearchConfig{
						Blacklist: []string{"some"},
						CapacityConstraints: &models.StackVersionInstanceCapacityConstraint{
							Max: ec.Int32(8192),
							Min: ec.Int32(512),
						},
						DockerImage:    new("docker.elastic.co/cloud-assets/elasticsearch:7.9.1-0"),
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
						DockerImage: new("docker.elastic.co/cloud-assets/enterprise_search:7.9.1-0"),
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

func newSampleStack(t *testing.T) modelV0 {
	return modelV0{
		ID:                types.StringValue("7.9.1"),
		Region:            types.StringValue("us-east-1"),
		Version:           types.StringValue("7.9.1"),
		VersionRegex:      types.StringValue("latest"),
		Accessible:        types.BoolValue(true),
		AllowListed:       types.BoolValue(true),
		MinUpgradableFrom: types.StringValue("6.8.0"),
		Elasticsearch: func() types.List {
			res, diags := types.ListValueFrom(
				context.Background(),
				types.ObjectType{AttrTypes: elasticsearchConfigAttrTypes()},
				[]elasticsearchConfigModelV0{
					{
						DenyList:               util.StringListAsType(t, []string{"some"}),
						CapacityConstraintsMax: types.Int64Value(8192),
						CapacityConstraintsMin: types.Int64Value(512),
						CompatibleNodeTypes:    util.StringListAsType(t, nil),
						DockerImage:            types.StringValue("docker.elastic.co/cloud-assets/elasticsearch:7.9.1-0"),
						Plugins: util.StringListAsType(t,
							[]string{
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
						),
						DefaultPlugins: util.StringListAsType(t, []string{"repository-s3"}),
					},
				},
			)
			assert.Nil(t, diags)

			return res
		}(),
		Kibana: func() types.List {
			res, diags := types.ListValueFrom(
				context.Background(),
				types.ObjectType{
					AttrTypes: resourceKindConfigAttrTypes(util.KibanaResourceKind),
				},
				[]resourceKindConfigModelV0{
					{
						DenyList:               util.StringListAsType(t, []string{"some"}),
						CapacityConstraintsMax: types.Int64Value(8192),
						CapacityConstraintsMin: types.Int64Value(512),
						CompatibleNodeTypes:    util.StringListAsType(t, nil),
						DockerImage:            types.StringValue("docker.elastic.co/cloud-assets/kibana:7.9.1-0"),
					},
				},
			)
			assert.Nil(t, diags)

			return res
		}(),
		EnterpriseSearch: func() types.List {
			res, diags := types.ListValueFrom(
				context.Background(),
				types.ObjectType{
					AttrTypes: resourceKindConfigAttrTypes(util.EnterpriseSearchResourceKind),
				},
				[]resourceKindConfigModelV0{
					{
						DenyList:               util.StringListAsType(t, []string{"some"}),
						CapacityConstraintsMax: types.Int64Value(8192),
						CapacityConstraintsMin: types.Int64Value(512),
						CompatibleNodeTypes:    util.StringListAsType(t, nil),
						DockerImage:            types.StringValue("docker.elastic.co/cloud-assets/enterprise_search:7.9.1-0"),
					},
				},
			)
			assert.Nil(t, diags)

			return res
		}(),
		Apm: func() types.List {
			res, diags := types.ListValueFrom(
				context.Background(),
				types.ObjectType{
					AttrTypes: resourceKindConfigAttrTypes(util.ApmResourceKind),
				},
				[]resourceKindConfigModelV0{
					{
						DenyList:               util.StringListAsType(t, []string{"some"}),
						CapacityConstraintsMax: types.Int64Value(8192),
						CapacityConstraintsMin: types.Int64Value(512),
						CompatibleNodeTypes:    util.StringListAsType(t, nil),
						DockerImage:            types.StringValue("docker.elastic.co/cloud-assets/apm:7.9.1-0"),
					},
				},
			)
			assert.Nil(t, diags)

			return res
		}(),
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
		expr   string
		stacks []*models.StackVersionConfig
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
			name: "returns an error when the expression doesn't match the stackpack",
			args: args{expr: "7.9.1", stacks: []*models.StackVersionConfig{
				{Version: "7.8.0"},
			}},
			err: errors.New(`failed to obtain a stack version matching "7.9.1": please specify a valid version_regex`),
		},
		{
			name: "returns an error when latest has no stacks",
			args: args{expr: "latest", stacks: nil},
			err:  errors.New(`failed to obtain a stack version matching "latest": please specify a valid version_regex`),
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
			got, err := stackFromFilters(tt.args.expr, tt.args.stacks)
			if !assert.Equal(t, tt.err, err) {
				fmt.Println(err, "!= want ", tt.err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_resolveStack(t *testing.T) {
	advertised := []*models.StackVersionConfig{
		{Version: "8.19.21"},
		{Version: "8.19.20"},
		{Version: "8.18.8"},
	}
	withPatch := []*models.StackVersionConfig{
		{Version: "9.5.3"},
		{Version: "8.19.21"},
	}

	type want struct {
		version string
		warning bool
		err     bool
	}

	tests := []struct {
		name string
		p    resolveParams
		want want
	}{
		{
			name: "exact regex uses the advertised match",
			p:    resolveParams{Expr: "8.19.21", Stacks: advertised},
			want: want{version: "8.19.21"},
		},
		{
			name: "exact regex is used when the version is no longer advertised",
			p:    resolveParams{Expr: "9.5.2", Stacks: advertised},
			want: want{version: "9.5.2", warning: true},
		},
		{
			name: "anchored exact regex is used when the version is no longer advertised",
			p:    resolveParams{Expr: "^9.5.2$", Stacks: advertised},
			want: want{version: "9.5.2", warning: true},
		},
		{
			name: "explicit version attribute pins a withdrawn version",
			p:    resolveParams{Expr: "latest", Version: "9.5.2", Stacks: advertised},
			want: want{version: "9.5.2", warning: true},
		},
		{
			name: "explicit version attribute uses the advertised pack when present",
			p:    resolveParams{Expr: "latest", Version: "8.19.20", Stacks: advertised},
			want: want{version: "8.19.20"},
		},
		{
			name: "patch regex uses the advertised match",
			p:    resolveParams{Expr: `9\.5\..+`, Stacks: withPatch},
			want: want{version: "9.5.3"},
		},
		{
			name: "patch regex errors when nothing advertised matches",
			p:    resolveParams{Expr: `9\.5\..+`, Stacks: advertised},
			want: want{err: true},
		},
		{
			name: "latest uses the advertised newest",
			p:    resolveParams{Expr: "latest", Stacks: advertised},
			want: want{version: "8.19.21"},
		},
		{
			name: "latest moves to a newly advertised newer stack",
			p:    resolveParams{Expr: "latest", Stacks: withPatch},
			want: want{version: "9.5.3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, warning, err := resolveStack(tt.p)
			if tt.want.err {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want.warning, warning != "")
			assert.Equal(t, tt.want.version, got.Version)
		})
	}
}

func Test_exactVersion(t *testing.T) {
	tests := []struct {
		expr string
		want string
		ok   bool
	}{
		{expr: "9.5.2", want: "9.5.2", ok: true},
		{expr: "^9.5.2$", want: "9.5.2", ok: true},
		{expr: "9.5.?", ok: false},
		{expr: `9\.5\..+`, ok: false},
		{expr: "latest", ok: false},
		{expr: "", ok: false},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			got, ok := exactVersion(tt.expr)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}
