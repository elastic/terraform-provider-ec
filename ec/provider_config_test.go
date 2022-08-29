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
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/auth"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_verboseSettings(t *testing.T) {
	f, err := os.CreateTemp("", "request")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	type args struct {
		name       string
		verbose    bool
		redactAuth bool
	}
	tests := []struct {
		name string
		args args
		want api.VerboseSettings
		err  error
	}{
		{
			name: "creates verbose settings when verbose = true",
			args: args{
				name:       f.Name(),
				verbose:    true,
				redactAuth: true,
			},
			want: api.VerboseSettings{
				Verbose:    true,
				RedactAuth: true,
			},
		},
		{
			name: "creates verbose settings when verbose = true, but without redacting auth",
			args: args{
				name:    f.Name(),
				verbose: true,
			},
			want: api.VerboseSettings{
				Verbose: true,
			},
		},
		{
			name: "skips creating verboose settings when verbose = false",
			args: args{
				name:       f.Name(),
				verbose:    false,
				redactAuth: true,
			},
			want: api.VerboseSettings{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := verboseSettings(tt.args.name, tt.args.verbose, tt.args.redactAuth)
			assert.Equal(t, tt.err, err)
			if tt.args.verbose {
				assert.NotNil(t, got.Device)
				got.Device = nil
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_newAPIConfig(t *testing.T) {
	defer unsetECAPIKey(t)()

	defaultCfg := util.NewResourceData(t, util.ResDataParams{
		ID:     "whocares",
		Schema: newSchema(),
		State:  map[string]interface{}{},
	})
	invalidTimeoutCfg := util.NewResourceData(t, util.ResDataParams{
		ID:     "whocares",
		Schema: newSchema(),
		State: map[string]interface{}{
			"timeout": "invalid",
		},
	})

	apiKeyCfg := util.NewResourceData(t, util.ResDataParams{
		ID:     "whocares",
		Schema: newSchema(),
		State: map[string]interface{}{
			"apikey": "blih",
		},
	})
	apiKeyObj := auth.APIKey("blih")

	userPassCfg := util.NewResourceData(t, util.ResDataParams{
		ID:     "whocares",
		Schema: newSchema(),
		State: map[string]interface{}{
			"username": "my-user",
			"password": "my-pass",
		},
	})
	userPassObj := auth.UserLogin{
		Username: "my-user",
		Password: "my-pass",
		Holder:   new(auth.GenericHolder),
	}

	insecureCfg := util.NewResourceData(t, util.ResDataParams{
		ID:     "whocares",
		Schema: newSchema(),
		State: map[string]interface{}{
			"apikey":   "blih",
			"insecure": true,
		},
	})

	verboseCfg := util.NewResourceData(t, util.ResDataParams{
		ID:     "whocares",
		Schema: newSchema(),
		State: map[string]interface{}{
			"apikey":  "blih",
			"verbose": true,
		},
	})
	defer func() {
		os.Remove("request.log")
	}()

	customFile, err := os.CreateTemp("", "request-custom-verbose")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		customFile.Close()
		os.Remove(customFile.Name())
	}()
	verboseCustomFileCfg := util.NewResourceData(t, util.ResDataParams{
		ID:     "whocares",
		Schema: newSchema(),
		State: map[string]interface{}{
			"apikey":       "blih",
			"verbose":      true,
			"verbose_file": customFile.Name(),
		},
	})
	verboseAndCredsCustomFileCfg := util.NewResourceData(t, util.ResDataParams{
		ID:     "whocares",
		Schema: newSchema(),
		State: map[string]interface{}{
			"apikey":              "blih",
			"verbose":             true,
			"verbose_file":        customFile.Name(),
			"verbose_credentials": true,
		},
	})
	invalidPath := filepath.Join("a", "b", "c", "d", "e", "f", "g", "h", "invalid!")
	verboseInvalidFileCfg := util.NewResourceData(t, util.ResDataParams{
		ID:     "whocares",
		Schema: newSchema(),
		State: map[string]interface{}{
			"apikey":              "blih",
			"verbose":             true,
			"verbose_file":        invalidPath,
			"verbose_credentials": true,
		},
	})
	type args struct {
		d *schema.ResourceData
	}
	tests := []struct {
		name         string
		args         args
		want         api.Config
		wantFileName string
		err          error
	}{
		{
			name: "default config returns with authwriter error",
			args: args{d: defaultCfg},
			err: multierror.NewPrefixed("authwriter",
				errors.New("one of apikey or username and password must be specified"),
			),
		},
		{
			name: "default config with  invalid timeout returns with authwriter error",
			args: args{d: invalidTimeoutCfg},
			err:  errors.New(`time: invalid duration "invalid"`),
		},
		{
			name: "custom config with apikey auth succeeds",
			args: args{d: apiKeyCfg},
			want: api.Config{
				UserAgent:   fmt.Sprintf(providerUserAgentFmt, Version, api.DefaultUserAgent),
				ErrorDevice: os.Stdout,
				Host:        api.ESSEndpoint,
				AuthWriter:  &apiKeyObj,
				Client:      &http.Client{},
				Timeout:     defaultTimeout,
				Retries:     DefaultHTTPRetries,
			},
		},
		{
			name: "custom config with username/password auth succeeds",
			args: args{d: userPassCfg},
			want: api.Config{
				UserAgent:   fmt.Sprintf(providerUserAgentFmt, Version, api.DefaultUserAgent),
				ErrorDevice: os.Stdout,
				Host:        api.ESSEndpoint,
				AuthWriter:  &userPassObj,
				Client:      &http.Client{},
				Timeout:     defaultTimeout,
				Retries:     DefaultHTTPRetries,
			},
		},
		{
			name: "custom config with insecure succeeds",
			args: args{d: insecureCfg},
			want: api.Config{
				UserAgent:     fmt.Sprintf(providerUserAgentFmt, Version, api.DefaultUserAgent),
				ErrorDevice:   os.Stdout,
				Host:          api.ESSEndpoint,
				AuthWriter:    &apiKeyObj,
				Client:        &http.Client{},
				Timeout:       defaultTimeout,
				Retries:       DefaultHTTPRetries,
				SkipTLSVerify: true,
			},
		},
		{
			name: "custom config with insecure succeeds",
			args: args{d: insecureCfg},
			want: api.Config{
				UserAgent:     fmt.Sprintf(providerUserAgentFmt, Version, api.DefaultUserAgent),
				ErrorDevice:   os.Stdout,
				Host:          api.ESSEndpoint,
				AuthWriter:    &apiKeyObj,
				Client:        &http.Client{},
				Timeout:       defaultTimeout,
				Retries:       DefaultHTTPRetries,
				SkipTLSVerify: true,
			},
		},
		{
			name: "custom config with verbose (default file) succeeds",
			args: args{d: verboseCfg},
			want: api.Config{
				UserAgent:   fmt.Sprintf(providerUserAgentFmt, Version, api.DefaultUserAgent),
				ErrorDevice: os.Stdout,
				Host:        api.ESSEndpoint,
				AuthWriter:  &apiKeyObj,
				Client:      &http.Client{},
				Timeout:     defaultTimeout,
				Retries:     DefaultHTTPRetries,
				VerboseSettings: api.VerboseSettings{
					Verbose:    true,
					RedactAuth: true,
				},
			},
			wantFileName: "request.log",
		},
		{
			name: "custom config with verbose (custom file) succeeds",
			args: args{d: verboseCustomFileCfg},
			want: api.Config{
				UserAgent:   fmt.Sprintf(providerUserAgentFmt, Version, api.DefaultUserAgent),
				ErrorDevice: os.Stdout,
				Host:        api.ESSEndpoint,
				AuthWriter:  &apiKeyObj,
				Client:      &http.Client{},
				Timeout:     defaultTimeout,
				Retries:     DefaultHTTPRetries,
				VerboseSettings: api.VerboseSettings{
					Verbose:    true,
					RedactAuth: true,
				},
			},
			wantFileName: filepath.Base(customFile.Name()),
		},
		{
			name: "custom config with verbose and verbose_credentials (custom file) succeeds",
			args: args{d: verboseAndCredsCustomFileCfg},
			want: api.Config{
				UserAgent:   fmt.Sprintf(providerUserAgentFmt, Version, api.DefaultUserAgent),
				ErrorDevice: os.Stdout,
				Host:        api.ESSEndpoint,
				AuthWriter:  &apiKeyObj,
				Client:      &http.Client{},
				Timeout:     defaultTimeout,
				Retries:     DefaultHTTPRetries,
				VerboseSettings: api.VerboseSettings{
					Verbose:    true,
					RedactAuth: false,
				},
			},
			wantFileName: filepath.Base(customFile.Name()),
		},
		{
			name: "custom config with verbose and verbose_credentials (invalid file) fails ",
			args: args{d: verboseInvalidFileCfg},
			err: fmt.Errorf(`failed creating verbose file "%s": %w`,
				invalidPath,
				&os.PathError{
					Op:   "open",
					Path: invalidPath,
					Err:  syscall.ENOENT,
				},
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newAPIConfig(tt.args.d)
			assert.Equal(t, tt.err, err)

			if got.Verbose && err == nil {
				assert.NotNil(t, got.Device)
				if f, ok := got.Device.(*os.File); ok {
					assert.Equal(t, tt.wantFileName, filepath.Base(f.Name()))
				}
				got.Device = nil
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func unsetECAPIKey(t *testing.T) func() {
	t.Helper()
	// This is necessary to avoid any EC_API_KEY which might be set to cause
	// test flakyness.
	if k := os.Getenv("EC_API_KEY"); k != "" {
		if err := os.Unsetenv("EC_API_KEY"); err != nil {
			t.Fatal(err)
		}
		return func() {
			if err := os.Setenv("EC_API_KEY", k); err != nil {
				t.Fatal(err)
			}
		}
	}
	return func() {}
}
