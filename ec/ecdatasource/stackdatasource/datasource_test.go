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
	"errors"
	"fmt"
	"regexp/syntax"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_modelToState(t *testing.T) {
	deploymentSchemaArg := schema.TestResourceDataRaw(t, newSchema(), nil)
	deploymentSchemaArg.SetId("someid")
	_ = deploymentSchemaArg.Set("region", "us-east-1")
	_ = deploymentSchemaArg.Set("version_regex", "latest")

	wantDeployment := util.NewResourceData(t, util.ResDataParams{
		ID:     "someid",
		State:  newSampleStack(),
		Schema: newSchema(),
	})

	type args struct {
		d   *schema.ResourceData
		res *models.StackVersionConfig
	}
	tests := []struct {
		name string
		args args
		want *schema.ResourceData
		err  error
	}{
		{
			name: "flattens deployment resources",
			want: wantDeployment,
			args: args{
				d: deploymentSchemaArg,
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
			err := modelToState(tt.args.d, tt.args.res)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want.State().Attributes, tt.args.d.State().Attributes)
		})
	}
}

func newSampleStack() map[string]interface{} {
	return map[string]interface{}{
		"id":            "someid",
		"region":        "us-east-1",
		"version_regex": "latest",

		"version":             "7.9.1",
		"accessible":          true,
		"allowlisted":         true,
		"min_upgradable_from": "6.8.0",
		"elasticsearch": []interface{}{map[string]interface{}{
			"denylist":                 []interface{}{"some"},
			"capacity_constraints_max": 8192,
			"capacity_constraints_min": 512,
			"default_plugins":          []interface{}{"repository-s3"},
			"docker_image":             "docker.elastic.co/cloud-assets/elasticsearch:7.9.1-0",
			"plugins": []interface{}{
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
		"kibana": []interface{}{map[string]interface{}{
			"denylist":                 []interface{}{"some"},
			"capacity_constraints_max": 8192,
			"capacity_constraints_min": 512,
			"docker_image":             "docker.elastic.co/cloud-assets/kibana:7.9.1-0",
		}},
		"apm": []interface{}{map[string]interface{}{
			"denylist":                 []interface{}{"some"},
			"capacity_constraints_max": 8192,
			"capacity_constraints_min": 512,
			"docker_image":             "docker.elastic.co/cloud-assets/apm:7.9.1-0",
		}},
		"enterprise_search": []interface{}{map[string]interface{}{
			"denylist":                 []interface{}{"some"},
			"capacity_constraints_max": 8192,
			"capacity_constraints_min": 512,
			"docker_image":             "docker.elastic.co/cloud-assets/enterprise_search:7.9.1-0",
		}},
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
