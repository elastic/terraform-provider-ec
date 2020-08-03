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
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-ec/ec"
)

const (
	prefix = "terraform_acc_"
)

var testAccProviderFactory = map[string]func() (*schema.Provider, error){
	"ec": providerFactory,
}

func providerFactory() (*schema.Provider, error) {
	return ec.Provider(), nil
}

func testAccPreCheck(t *testing.T) {
	var apikey, username, password string
	if k := os.Getenv("EC_APIKEY"); k != "" {
		apikey = k
	}
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
