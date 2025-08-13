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

	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/auth"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
)

func Test_verboseSettings(t *testing.T) {
	f, err := os.CreateTemp("", "request")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("failed to close temp file: %v", err)
		}
		if err := os.Remove(f.Name()); err != nil {
			t.Fatalf("failed to remove temp file: %v",
				err)
		}
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
	apiKeyObj := auth.APIKey("secret")

	userPassObj := auth.UserLogin{
		Username: "my-user",
		Password: "my-pass",
		Holder:   new(auth.GenericHolder),
	}

	defer func() {
		if err := os.Remove("request.log"); err != nil {
			t.Fatalf("failed to remove request.log: %v",
				err)
		}
	}()

	customFile, err := os.CreateTemp("", "request-custom-verbose")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := customFile.Close(); err != nil {
			t.Fatalf("failed to close custom temp file: %v", err)
		}
		if err := os.Remove(customFile.Name()); err != nil {
			t.Fatalf("failed to remove custom temp file: %v", err)
		}
	}()

	invalidPath := filepath.Join("a", "b", "c", "d", "e", "f", "g", "h", "invalid!")

	type args struct {
		apiSetup apiSetup
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
			args: args{
				apiSetup: apiSetup{
					timeout: defaultTimeout,
				},
			},
			err: multierror.NewPrefixed("authwriter",
				errors.New("one of apikey or username and password must be specified"),
			),
		},

		{
			name: "custom config with apikey auth succeeds",
			args: args{
				apiSetup: apiSetup{
					apikey:   "secret",
					timeout:  defaultTimeout,
					endpoint: api.ESSEndpoint,
				},
			},
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
			args: args{
				apiSetup: apiSetup{
					username: "my-user",
					password: "my-pass",
					timeout:  defaultTimeout,
					endpoint: api.ESSEndpoint,
				},
			},
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
			args: args{
				apiSetup: apiSetup{
					apikey:   "secret",
					insecure: true,
					timeout:  defaultTimeout,
					endpoint: api.ESSEndpoint,
				},
			},
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
			args: args{
				apiSetup: apiSetup{
					apikey:      "secret",
					verbose:     true,
					verboseFile: "request.log",
					timeout:     defaultTimeout,
					endpoint:    api.ESSEndpoint,
				},
			},
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
			args: args{
				apiSetup: apiSetup{
					apikey:      "secret",
					verbose:     true,
					verboseFile: customFile.Name(),
					timeout:     defaultTimeout,
					endpoint:    api.ESSEndpoint,
				},
			},
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
			args: args{
				apiSetup: apiSetup{
					apikey:             "secret",
					verbose:            true,
					verboseFile:        customFile.Name(),
					verboseCredentials: true,
					timeout:            defaultTimeout,
					endpoint:           api.ESSEndpoint,
				},
			},
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
			args: args{
				apiSetup: apiSetup{
					apikey:             "secret",
					verbose:            true,
					verboseFile:        invalidPath,
					verboseCredentials: true,
					timeout:            defaultTimeout,
				},
			},
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
			got, err := newAPIConfig(tt.args.apiSetup)
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
