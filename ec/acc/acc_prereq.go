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

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/auth"

	"github.com/elastic/terraform-provider-ec/ec"
)

const (
	prefix = "terraform_acc_"
)

var testAccProviderFactory = protoV6ProviderFactories()

func protoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"ec": providerserver.NewProtocol6WithError(ec.New("acc-tests")),
	}
}

func testAccPreCheck(t *testing.T) {
	var apikey, username, password string
	if k := os.Getenv("EC_API_KEY"); k != "" {
		apikey = k
	}

	if k := os.Getenv("EC_USER"); k != "" {
		username = k
	}
	if k := os.Getenv("EC_USERNAME"); k != "" {
		username = k
	}

	if k := os.Getenv("EC_PASS"); k != "" {
		password = k
	}
	if k := os.Getenv("EC_PASSWORD"); k != "" {
		password = k
	}

	if apikey == "" && (username == "" || password == "") {
		t.Fatal("No valid credentials found to execute acceptance tests")
	}

	if apikey != "" && (username != "" || password != "") {
		t.Fatal("Only one of API Key or Username / Password can be specified to execute acceptance tests")
	}
}

func newAPI() (*api.API, error) {
	var host = api.ESSEndpoint
	if h := os.Getenv("EC_HOST"); h != "" {
		host = h
	}
	if h := os.Getenv("EC_ENDPOINT"); h != "" {
		host = h
	}

	var apikey string
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
		Retries:       ec.DefaultHTTPRetries,
	})
}

// requiresAPIConn should be called in functions which would be executed by the
// Go testing framework and require external HTTP access, said functions should
// call this one to avoid the tests errorring because of failing prequisites.
func requiresAPIConn(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip()
	}
}
