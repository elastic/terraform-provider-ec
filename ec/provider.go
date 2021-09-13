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
	"time"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/elastic/terraform-provider-ec/ec/ecdatasource/deploymentdatasource"
	"github.com/elastic/terraform-provider-ec/ec/ecdatasource/deploymentsdatasource"
	"github.com/elastic/terraform-provider-ec/ec/ecdatasource/stackdatasource"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/elasticsearchkeystoreresource"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/extensionresource"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/trafficfilterassocresource"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/trafficfilterresource"
)

const (
	eceOnlyText      = "Available only when targeting ECE Installations or Elasticsearch Service Private"
	saasRequiredText = "The only valid authentication mechanism for the Elasticsearch Service"

	endpointDesc     = "Endpoint where the terraform provider will point to. Defaults to \"%s\"."
	insecureDesc     = "Allow the provider to skip TLS validation on its outgoing HTTP calls."
	timeoutDesc      = "Timeout used for individual HTTP calls. Defaults to \"1m\"."
	verboseDesc      = "When set, a \"request.log\" file will be written with all outgoing HTTP requests. Defaults to \"false\"."
	verboseCredsDesc = "When set with verbose, the contents of the Authorization header will not be redacted. Defaults to \"false\"."
)

var (
	apikeyDesc   = fmt.Sprint("API Key to use for API authentication. ", saasRequiredText, ".")
	usernameDesc = fmt.Sprint("Username to use for API authentication. ", eceOnlyText, ".")
	passwordDesc = fmt.Sprint("Password to use for API authentication. ", eceOnlyText, ".")

	validURLSchemes = []string{"http", "https"}

	// defaultTimeout used for all outgoing HTTP requests, keeping it low-ish
	// since any requests which timeout due to network factors are retried
	// automatically by the SDK 2 times.
	defaultTimeout = 40 * time.Second
)

// Provider returns a schema.Provider.
func Provider() *schema.Provider {
	return &schema.Provider{
		ConfigureContextFunc: configureAPI,
		Schema:               newSchema(),
		DataSourcesMap: map[string]*schema.Resource{
			"ec_deployment":  deploymentdatasource.DataSource(),
			"ec_deployments": deploymentsdatasource.DataSource(),
			"ec_stack":       stackdatasource.DataSource(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"ec_deployment":                            deploymentresource.Resource(),
			"ec_deployment_elasticsearch_keystore":     elasticsearchkeystoreresource.Resource(),
			"ec_deployment_traffic_filter":             trafficfilterresource.Resource(),
			"ec_deployment_traffic_filter_association": trafficfilterassocresource.Resource(),
			"ec_deployment_extension":                  extensionresource.Resource(),
		},
	}
}

func newSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"endpoint": {
			Description:  fmt.Sprintf(endpointDesc, api.ESSEndpoint),
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.IsURLWithScheme(validURLSchemes),
			DefaultFunc: schema.MultiEnvDefaultFunc(
				[]string{"EC_ENDPOINT", "EC_HOST"},
				api.ESSEndpoint,
			),
		},
		"apikey": {
			Description: apikeyDesc,
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
			DefaultFunc: schema.MultiEnvDefaultFunc(
				[]string{"EC_API_KEY"}, "",
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
			DefaultFunc: schema.MultiEnvDefaultFunc(
				[]string{"EC_TIMEOUT"}, defaultTimeout.String(),
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
		"verbose_credentials": {
			Description: verboseCredsDesc,
			Type:        schema.TypeBool,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc(
				"EC_VERBOSE_CREDENTIALS", false,
			),
		},
		"verbose_file": {
			Description: timeoutDesc,
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc(
				"EC_VERBOSE_FILE", "request.log",
			),
		},
	}
}
