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
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
	"os"
	"time"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/auth"
)

const (
	providerUserAgentFmt = "elastic-terraform-provider/%s (%s)"
)

var (
	// DefaultHTTPRetries to use for the provider's HTTP Client.
	DefaultHTTPRetries = 2
)

func configureAPI(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	cfg, err := newAPIConfigLegacy(d)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	client, err := api.NewAPI(cfg)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return client, nil
}

func newAPIConfigLegacy(d *schema.ResourceData) (api.Config, error) {
	endpoint := d.Get("endpoint").(string)
	apiKey := d.Get("apikey").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	timeout := d.Get("timeout").(string)
	insecure := d.Get("insecure").(bool)
	verbose := d.Get("verbose").(bool)
	verboseCredentials := d.Get("verbose_credentials").(bool)
	verboseFile := d.Get("verbose_file").(string)
	cfg, err := newAPIConfig(endpoint, apiKey, username, password, insecure, timeout, verbose, verboseCredentials, verboseFile)
	if err != nil {
		return api.Config{}, err
	}
	return cfg, nil
}

func newAPIConfig(endpoint string,
	apiKey string,
	username string,
	password string,
	insecure bool,
	timeout string,
	verbose bool,
	verboseCredentials bool,
	verboseFile string) (api.Config, error) {
	var cfg api.Config

	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return cfg, err
	}

	authWriter, err := auth.NewAuthWriter(auth.Config{
		APIKey:   apiKey,
		Username: username,
		Password: password,
	})
	if err != nil {
		return cfg, err
	}

	verboseCfg, err := verboseSettings(
		verboseFile,
		verbose,
		!verboseCredentials,
	)
	if err != nil {
		return cfg, err
	}

	return api.Config{
		ErrorDevice:     os.Stdout,
		Client:          &http.Client{},
		VerboseSettings: verboseCfg,
		AuthWriter:      authWriter,
		Host:            endpoint,
		SkipTLSVerify:   insecure,
		Timeout:         timeoutDuration,
		UserAgent:       userAgent(Version),
		Retries:         DefaultHTTPRetries,
	}, nil
}

func verboseSettings(name string, verbose, redactAuth bool) (api.VerboseSettings, error) {
	var cfg api.VerboseSettings
	if !verbose {
		return cfg, nil
	}

	f, err := os.Create(name)
	if err != nil {
		return cfg, fmt.Errorf(`failed creating verbose file "%s": %w`, name, err)
	}

	return api.VerboseSettings{
		Verbose:    true,
		RedactAuth: redactAuth,
		Device:     f,
	}, nil
}

func userAgent(v string) string {
	return fmt.Sprintf(providerUserAgentFmt, v, api.DefaultUserAgent)
}
