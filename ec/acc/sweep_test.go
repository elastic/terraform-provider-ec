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

package acc

import (
	"net/http"
	"os"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/auth"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func NewAPI() (*api.API, error) {
	var host = api.ESSEndpoint
	if h := os.Getenv("EC_HOST"); h != "" {
		host = h
	}
	if h := os.Getenv("EC_ENDPOINT"); h != "" {
		host = h
	}

	var apikey string
	if k := os.Getenv("EC_APIKEY"); k != "" {
		apikey = k
	}
	if k := os.Getenv("EC_API_KEY"); k != "" {
		apikey = k
	}

	var username string
	if k := os.Getenv("EC_USER"); k != "" {
		username = k
	}
	if k := os.Getenv("EC_USERNAME"); k != "" {
		username = k
	}

	var password string
	if k := os.Getenv("EC_UPASS"); k != "" {
		password = k
	}
	if k := os.Getenv("EC_PASSWORD"); k != "" {
		password = k
	}

	authWriter, err := auth.NewAuthWriter(auth.Config{
		APIKey: apikey, Username: username, Password: password,
	})
	if err != nil {
		return nil, err
	}

	var insecure bool
	if host != api.ESSEndpoint {
		insecure = true
	}

	return api.NewAPI(api.Config{
		ErrorDevice:   os.Stdout,
		Client:        &http.Client{},
		AuthWriter:    authWriter,
		Host:          host,
		SkipTLSVerify: insecure,
	})
}
