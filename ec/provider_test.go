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

package ec

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func Test_Configure(t *testing.T) {
	type args struct {
		env    map[string]string
		config providerConfig
	}

	tests := []struct {
		name  string
		args  args
		diags diag.Diagnostics
	}{
		{
			name: `provider config doesn't define "endpoint" and "EC_ENDPOINT" is defined and invalid`,
			args: args{
				env: map[string]string{
					"EC_ENDPOINT": "invalid",
				},
				config: providerConfig{
					Endpoint: types.String{Null: true},
				},
			},
			diags: func() diag.Diagnostics {
				var diags diag.Diagnostics
				diags.AddAttributeError(path.Root("endpoint"), "Value must be a valid URL with scheme (http, https)", "URL is missing host, got invalid")
				return diags
			}(),
		},

		{
			name: `provider config and env vars don't define either api key or user login/passwords`,
			args: args{
				env: map[string]string{
					"EC_ENDPOINT": "https://cloud.elastic.co/api",
				},
				config: providerConfig{
					Endpoint: types.String{Null: true},
					Username: types.String{Null: true},
				},
			},
			diags: func() diag.Diagnostics {
				var diags diag.Diagnostics
				diags.AddError("Unable to create api Client config", "authwriter: 1 error occurred:\n\t* one of apikey or username and password must be specified\n\n")
				return diags
			}(),
		},

		{
			name: `provider config doesn't define "insecure" and "EC_INSECURE" contains invalid value`,
			args: args{
				env: map[string]string{
					"EC_INSECURE": "invalid",
				},
				config: providerConfig{
					Endpoint: types.String{Value: "https://cloud.elastic.co/api"},
					ApiKey:   types.String{Value: "secret"},
					Insecure: types.Bool{Null: true},
				},
			},
			diags: func() diag.Diagnostics {
				var diags diag.Diagnostics
				diags.AddError("Unable to create client", "Invalid value 'invalid' in 'EC_INSECURE' or 'EC_SKIP_TLS_VALIDATION'")
				return diags
			}(),
		},

		{
			name: `provider config doesn't define "verbose" and "EC_VERBOSE" contains invalid value`,
			args: args{
				env: map[string]string{
					"EC_VERBOSE": "invalid",
				},
				config: providerConfig{
					Endpoint: types.String{Value: "https://cloud.elastic.co/api"},
					ApiKey:   types.String{Value: "secret"},
					Verbose:  types.Bool{Null: true},
				},
			},
			diags: func() diag.Diagnostics {
				var diags diag.Diagnostics
				diags.AddError("Unable to create client", "Invalid value 'invalid' in 'EC_VERBOSE'")
				return diags
			}(),
		},

		{
			name: `provider config doesn't define "verbose" and "EC_VERBOSE_CREDENTIALS" contains invalid value`,
			args: args{
				env: map[string]string{
					"EC_VERBOSE_CREDENTIALS": "invalid",
				},
				config: providerConfig{
					Endpoint:           types.String{Value: "https://cloud.elastic.co/api"},
					ApiKey:             types.String{Value: "secret"},
					VerboseCredentials: types.Bool{Null: true},
				},
			},
			diags: func() diag.Diagnostics {
				var diags diag.Diagnostics
				diags.AddError("Unable to create client", "Invalid value 'invalid' in 'EC_VERBOSE_CREDENTIALS'")
				return diags
			}(),
		},

		{
			name: `provider config is read from environment variables`,
			args: args{
				env: map[string]string{
					"EC_ENDPOINT":            "https://cloud.elastic.co/api",
					"EC_API_KEY":             "secret",
					"EC_INSECURE":            "true",
					"EC_TIMEOUT":             "1m",
					"EC_VERBOSE":             "true",
					"EC_VERBOSE_CREDENTIALS": "true",
					"EC_VERBOSE_FILE":        "requests.log",
				},
				config: providerConfig{
					Endpoint:           types.String{Null: true},
					ApiKey:             types.String{Null: true},
					Insecure:           types.Bool{Null: true},
					Timeout:            types.String{Null: true},
					Verbose:            types.Bool{Null: true},
					VerboseCredentials: types.Bool{Null: true},
					VerboseFile:        types.String{Null: true},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Provider

			schema, diags := p.GetSchema(context.Background())

			assert.Nil(t, diags)

			resp := provider.ConfigureResponse{}

			util.GetEnv = func(key string) string {
				return tt.args.env[key]
			}

			var config types.Object

			diags = tfsdk.ValueFrom(context.Background(), &tt.args.config, schema.Type(), &config)

			assert.Nil(t, diags)

			rawConfig, err := config.ToTerraformValue(context.Background())

			assert.Nil(t, err)

			p.Configure(
				context.Background(),
				provider.ConfigureRequest{
					Config: tfsdk.Config{Schema: schema, Raw: rawConfig},
				},
				&resp,
			)

			if tt.diags != nil {
				assert.Equal(t, tt.diags, resp.Diagnostics)
			} else {
				assert.Nil(t, resp.Diagnostics)
			}
		})
	}
}
