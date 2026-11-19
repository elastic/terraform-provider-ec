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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
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
	caCert             string
	tlsCert            string
	tlsKey             string
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

	client, err := newHTTPClient(setup)
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
		Client:          client,
		VerboseSettings: verboseCfg,
		AuthWriter:      authWriter,
		Host:            setup.endpoint,
		SkipTLSVerify:   setup.insecure,
		Timeout:         setup.timeout,
		UserAgent:       userAgent(Version),
		Retries:         DefaultHTTPRetries,
	}, nil
}

// newHTTPClient leaves the default transport to the SDK unless custom TLS is configured.
func newHTTPClient(setup apiSetup) (*http.Client, error) {
	if (setup.tlsCert == "") != (setup.tlsKey == "") {
		return nil, fmt.Errorf("tls_cert and tls_key must be configured together")
	}
	client := &http.Client{}
	if setup.caCert == "" && setup.tlsCert == "" {
		return client, nil
	}
	tlsConfig := &tls.Config{}
	if setup.caCert != "" {
		pem, err := os.ReadFile(setup.caCert)
		if err != nil {
			return nil, fmt.Errorf("failed reading ca_cert: %w", err)
		}
		roots, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("failed loading system CA certificates: %w", err)
		}
		if !roots.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("ca_cert does not contain any valid PEM certificates")
		}
		tlsConfig.RootCAs = roots
	}
	if setup.tlsCert != "" {
		cert, err := tls.LoadX509KeyPair(setup.tlsCert, setup.tlsKey)
		if err != nil {
			return nil, fmt.Errorf("failed loading tls_cert and tls_key: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsConfig
	timeout := setup.timeout
	if timeout <= 0 {
		timeout = api.DefaultTimeout
	}
	transport.DialContext = (&net.Dialer{Timeout: timeout, KeepAlive: 30 * time.Second}).DialContext
	client.Transport = transport
	return client, nil
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
