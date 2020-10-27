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
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func Test_splitID(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
		err  error
	}{
		{
			name: "returns nil when import ID a simple ID",
			args: args{id: mock.ValidClusterID},
			want: map[string]string{
				"id": "320b7b540dfc967a7a649c18e2fce4ed",
			},
		},
		{
			name: "returns err when import keys are invalid",
			args: args{id: strings.Join([]string{
				mock.ValidClusterID,
				"elasticsearch_password=somepass",
				"elasticsearch_username:someuser",
			}, ":")},
			err: multierror.NewPrefixed("invalid import id",
				errors.New(`"elasticsearch_username" not in <key>=<value> format`),
				errors.New(`"someuser" not in <key>=<value> format`),
			),
		},
		{
			name: "parses the keys",
			args: args{id: strings.Join([]string{
				mock.ValidClusterID,
				"elasticsearch_password=somepass",
				"elasticsearch_username=someuser",
			}, ":")},
			want: map[string]string{
				"elasticsearch_password": "somepass",
				"elasticsearch_username": "someuser",
				"id":                     "320b7b540dfc967a7a649c18e2fce4ed",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := splitID(tt.args.id)
			if !assert.Equal(t, tt.err, err) {
				t.Error(err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_importFunc(t *testing.T) {
	t.Skip()
	deploymentWithComplexID := util.NewResourceData(t, util.ResDataParams{
		ID: strings.Join([]string{
			mock.ValidClusterID,
			"elasticsearch_password=somepass",
			"elasticsearch_username=someuser",
			"apm_secret_token=somesecret",
		}, ":"),
		Schema: newSchema(),
		State: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-cross-cluster-search-v2",
			"region":                 "us-east-1",
			"version":                "7.9.2",
			"elasticsearch":          []interface{}{map[string]interface{}{}},
		},
	})
	deploymentWithSimpleID := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		Schema: newSchema(),
		State: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-cross-cluster-search-v2",
			"region":                 "us-east-1",
			"version":                "7.9.2",
			"elasticsearch":          []interface{}{map[string]interface{}{}},
		},
	})

	deploymentWithInvalidSettingsID := util.NewResourceData(t, util.ResDataParams{
		ID: strings.Join([]string{
			mock.ValidClusterID,
			"invalidFormat",
		}, ":"),
		Schema: newSchema(),
		State: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-cross-cluster-search-v2",
			"region":                 "us-east-1",
			"version":                "7.9.2",
			"elasticsearch":          []interface{}{map[string]interface{}{}},
		},
	})

	deploymentWithInvalidAddressID := util.NewResourceData(t, util.ResDataParams{
		ID: strings.Join([]string{
			mock.ValidClusterID, "apm_something=value",
		}, ":"),
		Schema: newSchema(),
		State: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-cross-cluster-search-v2",
			"region":                 "us-east-1",
			"version":                "7.9.2",
			"elasticsearch":          []interface{}{map[string]interface{}{}},
		},
	})
	type args struct {
		ctx context.Context
		d   *schema.ResourceData
		m   interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]string
		err  error
	}{
		{
			name: "succeeds with simple ID",
			args: args{d: deploymentWithSimpleID},
			want: map[string]string{
				"id": "320b7b540dfc967a7a649c18e2fce4ed",

				"name":                   "my_deployment_name",
				"region":                 "us-east-1",
				"version":                "7.9.2",
				"deployment_template_id": "aws-cross-cluster-search-v2",

				"elasticsearch.#":                       "1",
				"elasticsearch.0.cloud_id":              "",
				"elasticsearch.0.config.#":              "0",
				"elasticsearch.0.http_endpoint":         "",
				"elasticsearch.0.https_endpoint":        "",
				"elasticsearch.0.monitoring_settings.#": "0",
				"elasticsearch.0.ref_id":                "main-elasticsearch",
				"elasticsearch.0.region":                "",
				"elasticsearch.0.remote_cluster.#":      "0",
				"elasticsearch.0.resource_id":           "",
				"elasticsearch.0.topology.#":            "0",
				"elasticsearch.0.version":               "",
			},
		},
		{
			name: "succeeds with complex ID",
			args: args{d: deploymentWithComplexID},
			want: map[string]string{
				"apm_secret_token":       "somesecret",
				"elasticsearch_password": "somepass",
				"elasticsearch_username": "someuser",
				"id":                     "320b7b540dfc967a7a649c18e2fce4ed",

				"name":                   "my_deployment_name",
				"region":                 "us-east-1",
				"version":                "7.9.2",
				"deployment_template_id": "aws-cross-cluster-search-v2",

				"elasticsearch.#":                       "1",
				"elasticsearch.0.cloud_id":              "",
				"elasticsearch.0.config.#":              "0",
				"elasticsearch.0.http_endpoint":         "",
				"elasticsearch.0.https_endpoint":        "",
				"elasticsearch.0.monitoring_settings.#": "0",
				"elasticsearch.0.ref_id":                "main-elasticsearch",
				"elasticsearch.0.region":                "",
				"elasticsearch.0.remote_cluster.#":      "0",
				"elasticsearch.0.resource_id":           "",
				"elasticsearch.0.topology.#":            "0",
				"elasticsearch.0.version":               "",
			},
		},
		{
			name: "fails with complex ID (bad format)",
			args: args{d: deploymentWithInvalidSettingsID},
			want: map[string]string{
				"id": "320b7b540dfc967a7a649c18e2fce4ed:invalidFormat",

				"name":                   "my_deployment_name",
				"region":                 "us-east-1",
				"version":                "7.9.2",
				"deployment_template_id": "aws-cross-cluster-search-v2",

				"elasticsearch.#":                       "1",
				"elasticsearch.0.cloud_id":              "",
				"elasticsearch.0.config.#":              "0",
				"elasticsearch.0.http_endpoint":         "",
				"elasticsearch.0.https_endpoint":        "",
				"elasticsearch.0.monitoring_settings.#": "0",
				"elasticsearch.0.ref_id":                "main-elasticsearch",
				"elasticsearch.0.region":                "",
				"elasticsearch.0.remote_cluster.#":      "0",
				"elasticsearch.0.resource_id":           "",
				"elasticsearch.0.topology.#":            "0",
				"elasticsearch.0.version":               "",
			},
			err: multierror.NewPrefixed("invalid import id",
				errors.New(`"invalidFormat" not in <key>=<value> format`),
			),
		},
		{
			name: "fails with complex ID (invalid address)",
			args: args{d: deploymentWithInvalidAddressID},
			want: map[string]string{
				"id": "320b7b540dfc967a7a649c18e2fce4ed",

				"name":                   "my_deployment_name",
				"region":                 "us-east-1",
				"version":                "7.9.2",
				"deployment_template_id": "aws-cross-cluster-search-v2",

				"elasticsearch.#":                       "1",
				"elasticsearch.0.cloud_id":              "",
				"elasticsearch.0.config.#":              "0",
				"elasticsearch.0.http_endpoint":         "",
				"elasticsearch.0.https_endpoint":        "",
				"elasticsearch.0.monitoring_settings.#": "0",
				"elasticsearch.0.ref_id":                "main-elasticsearch",
				"elasticsearch.0.region":                "",
				"elasticsearch.0.remote_cluster.#":      "0",
				"elasticsearch.0.resource_id":           "",
				"elasticsearch.0.topology.#":            "0",
				"elasticsearch.0.version":               "",
			},
			err: multierror.NewPrefixed("failed setting import keys",
				errors.New(`failed setting "apm_something" as "value": Invalid address to set: []string{"apm_something"}`),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := importFunc(tt.args.ctx, tt.args.d, tt.args.m)
			if tt.err != nil {
				if !assert.EqualError(t, err, tt.err.Error()) {
					t.Error(err)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, tt.args.d.State().Attributes)
		})
	}
}
