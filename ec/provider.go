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

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-ec/ec/ecresource"
)

const (
	eceOnlyText      = "Only usable when targeting ECE Installations or Elasticsearch Service Private"
	saasRequiredText = "It's the only valid authentication mechanism for the Elasticsearch Service"

	endpointDesc = "Endpoint where the terraform provider will point to, defaults to \"%s\"."
	insecureDesc = "Allow the provider to skip TLS validation on its outgoing HTTP calls."
	timeoutDesc  = "Timeout used for individual HTTP calls. Defaults to \"1m\"."
	verboseDesc  = "When set, a \"request.log\" file will be written with all outgoing HTTP requests. Defaults to \"false\"."
)

var (
	// DefaultEndpoint is the default provider endpoint.
	DefaultEndpoint = api.ESSEndpoint
)

var (
	apikeyDesc   = fmt.Sprint("API Key to use for API authentication. ", saasRequiredText, ".")
	usernameDesc = fmt.Sprint("Username to use for API authentication. ", eceOnlyText, ".")
	passwordDesc = fmt.Sprint("Password to use for API authentication. ", eceOnlyText, ".")

	validURLSchemes = []string{"http", "https"}
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		ConfigureFunc: configureAPI,
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Description:  fmt.Sprintf(endpointDesc, DefaultEndpoint),
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsURLWithScheme(validURLSchemes),
				DefaultFunc: schema.MultiEnvDefaultFunc(
					[]string{"EC_ENDPOINT", "EC_HOST"},
					DefaultEndpoint,
				),
			},
			"apikey": {
				Description: apikeyDesc,
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.MultiEnvDefaultFunc(
					[]string{"EC_APIKEY", "EC_API_KEY"}, "",
				),
			},
			"username": {
				Description: usernameDesc,
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc(
					[]string{"EC_USER", "EC_USERNAME"}, "",
				),
			},
			"password": {
				Description: passwordDesc,
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.MultiEnvDefaultFunc(
					[]string{"EC_PASS", "EC_PASSWORD"}, "",
				),
			},
			"insecure": {
				Description: insecureDesc,
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				DefaultFunc: schema.MultiEnvDefaultFunc(
					[]string{"EC_INSECURE", "EC_SKIP_TLS_VALIDATION"},
					false,
				),
			},
			"timeout": {
				Description: timeoutDesc,
				Type:        schema.TypeString,
				Optional:    true,
				Default:     false,
				DefaultFunc: schema.MultiEnvDefaultFunc(
					[]string{"EC_TIMEOUT"}, "1m",
				),
			},
			"verbose": {
				Description: verboseDesc,
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc(
					[]string{"EC_VERBOSE"}, false,
				),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"ec_deployment": ecresource.Deployment(),
		},
	}
}
