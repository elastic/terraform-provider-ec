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
	"fmt"
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
	// DefaultHTTPRetries to use for the provider's HTTP client.
	DefaultHTTPRetries = 2
)

type apiSetup struct {
	endpoint           string
	apikey             string
	username           string
	password           string
	insecure           bool
	timeout            time.Duration
	verbose            bool
	verboseCredentials bool
	verboseFile        string
}

func newAPIConfig(setup apiSetup) (api.Config, error) {

	var cfg api.Config

	authWriter, err := auth.NewAuthWriter(auth.Config{
		APIKey:   setup.apikey,
		Username: setup.username,
		Password: setup.password,
	})
	if err != nil {
		return cfg, err
	}

	verboseCfg, err := verboseSettings(
		setup.verboseFile,
		setup.verbose,
		!setup.verboseCredentials,
	)
	if err != nil {
		return cfg, err
	}

	return api.Config{
		ErrorDevice:     os.Stdout,
		Client:          &http.Client{},
		VerboseSettings: verboseCfg,
		AuthWriter:      authWriter,
		Host:            setup.endpoint,
		SkipTLSVerify:   setup.insecure,
		Timeout:         setup.timeout,
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
		Verbose:    verbose,
		RedactAuth: redactAuth,
		Device:     f,
	}, nil
}

func userAgent(v string) string {
	return fmt.Sprintf(providerUserAgentFmt, v, api.DefaultUserAgent)
}
