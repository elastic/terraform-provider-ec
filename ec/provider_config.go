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
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	providerUserAgentFmt = "elastic-terraform-provider/%s (%s)"
)

// configureAPI implements schema.ConfigureFunc
func configureAPI(d *schema.ResourceData) (interface{}, error) {
	timeout, err := time.ParseDuration(d.Get("timeout").(string))
	if err != nil {
		return nil, err
	}

	authWriter, err := auth.NewAuthWriter(auth.Config{
		APIKey:   d.Get("apikey").(string),
		Username: d.Get("username").(string),
		Password: d.Get("password").(string),
	})
	if err != nil {
		return nil, err
	}

	return api.NewAPI(api.Config{
		ErrorDevice:     os.Stdout,
		Client:          &http.Client{},
		VerboseSettings: verboseSettings(d.Get("verbose").(bool)),
		AuthWriter:      authWriter,
		Host:            d.Get("endpoint").(string),
		SkipTLSVerify:   d.Get("insecure").(bool),
		Timeout:         timeout,
		UserAgent:       userAgent(Version),
	})
}

func userAgent(v string) string {
	return fmt.Sprintf(providerUserAgentFmt, v, api.DefaultUserAgent)
}

func verboseSettings(verbose bool) api.VerboseSettings {
	if !verbose {
		return api.VerboseSettings{}
	}

	f, err := os.Create("request.log")
	if err != nil {
		return api.VerboseSettings{}
	}

	return api.VerboseSettings{
		Verbose: true, Device: f,
	}
}
