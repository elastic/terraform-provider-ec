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
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/elastic/terraform-provider-ec/ec/internal/validators"
	"time"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	//"github.com/elastic/terraform-provider-ec/ec/ecdatasource/deploymentdatasource"
	//"github.com/elastic/terraform-provider-ec/ec/ecdatasource/deploymentsdatasource"
	//"github.com/elastic/terraform-provider-ec/ec/ecdatasource/stackdatasource"
	//"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource"
	//"github.com/elastic/terraform-provider-ec/ec/ecresource/elasticsearchkeystoreresource"
	//"github.com/elastic/terraform-provider-ec/ec/ecresource/extensionresource"
	//"github.com/elastic/terraform-provider-ec/ec/ecresource/trafficfilterassocresource"
	//"github.com/elastic/terraform-provider-ec/ec/ecresource/trafficfilterresource"
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

func Provider() provider.Provider {
	return &ecProvider{}
	//ConfigureContextFunc: configureAPI,
	//DataSourcesMap: map[string]*schema.Resource{
	//	"ec_deployment":  deploymentdatasource.DataSource(),
	//	"ec_deployments": deploymentsdatasource.DataSource(),
	//	"ec_stack":       stackdatasource.DataSource(),
	//},
	//ResourcesMap: map[string]*schema.Resource{
	//	"ec_deployment":                            deploymentresource.Resource(),
	//	"ec_deployment_elasticsearch_keystore":     elasticsearchkeystoreresource.Resource(),
	//	"ec_deployment_traffic_filter":             trafficfilterresource.Resource(),
	//	"ec_deployment_traffic_filter_association": trafficfilterassocresource.Resource(),
	//	"ec_deployment_extension":                  extensionresource.Resource(),
	//},
}

var _ provider.Provider = (*ecProvider)(nil)

type ecProvider struct{}

func (p *ecProvider) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	var diags diag.Diagnostics

	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"endpoint": {
				Description: fmt.Sprintf(endpointDesc, api.ESSEndpoint),
				Type:        types.StringType,
				Optional:    true,
				Validators:  []tfsdk.AttributeValidator{validators.Known(), validators.IsURLWithSchemeValidator(validURLSchemes)},
			},
			"apikey": {
				Description: apikeyDesc,
				Type:        types.StringType,
				Optional:    true,
				Sensitive:   true,
			},
			"username": {
				Description: usernameDesc,
				Type:        types.StringType,
				Optional:    true,
			},
			"password": {
				Description: passwordDesc,
				Type:        types.StringType,
				Optional:    true,
				Sensitive:   true,
			},
			"insecure": {
				Description: insecureDesc,
				Type:        types.BoolType,
				Optional:    true,
			},
			"timeout": {
				Description: timeoutDesc,
				Type:        types.StringType,
				Optional:    true,
			},
			"verbose": {
				Description: verboseDesc,
				Type:        types.BoolType,
				Optional:    true,
			},
			"verbose_credentials": {
				Description: verboseCredsDesc,
				Type:        types.BoolType,
				Optional:    true,
			},
			"verbose_file": {
				Description: timeoutDesc,
				Type:        types.StringType,
				Optional:    true,
			},
		},
	}, diags
}

type providerData struct {
	Endpoint           types.String `tfsdk:"endpoint"`
	ApiKey             types.String `tfsdk:"apikey"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	Insecure           types.Bool   `tfsdk:"insecure"`
	Timeout            types.String `tfsdk:"timeout"`
	Verbose            types.Bool   `tfsdk:"verbose"`
	VerboseCredentials types.Bool   `tfsdk:"verbose_credentials"`
	VerboseFile        types.String `tfsdk:"verbose_file"`
}

func (p *ecProvider) Configure(ctx context.Context, req provider.ConfigureRequest, res *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config providerData
	diags := req.Config.Get(ctx, &config)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	var endpoint string
	if config.Endpoint.Null {
		endpoint = util.MultiGetenv([]string{"EC_ENDPOINT", "EC_HOST"}, api.ESSEndpoint)
		/* TODO validate endpoint (see validators used above)
			res.Diagnostics.AddWarning(
				"Unable to create client",
				"Cannot use unknown value as endpoint",
			)
			return
		}
		*/
	} else {
		endpoint = config.Endpoint.Value
	}

	var apiKey string
	if config.ApiKey.Null {
		apiKey = util.MultiGetenv([]string{"EC_API_KEY"}, "")
	} else {
		apiKey = config.ApiKey.Value
	}

	var username string
	if config.Username.Null {
		username = util.MultiGetenv([]string{"EC_USER", "EC_USERNAME"}, "")
	} else {
		username = config.Username.Value
	}

	var password string
	if config.Password.Null {
		password = util.MultiGetenv([]string{"EC_PASS", "EC_PASSWORD"}, "")
	} else {
		password = config.Password.Value
	}

	var err error
	var insecure bool
	if config.Insecure.Null {
		insecureStr := util.MultiGetenv([]string{"EC_INSECURE", "EC_SKIP_TLS_VALIDATION"}, "")
		if insecure, err = util.StringToBool(insecureStr); err != nil {
			res.Diagnostics.AddWarning(
				"Unable to create client",
				fmt.Sprintf("Invalid value %v for insecure", insecureStr),
			)
			return
		}
	} else {
		insecure = config.Insecure.Value
	}

	var timeout string
	if config.Timeout.Null {
		timeout = util.MultiGetenv([]string{"EC_TIMEOUT"}, defaultTimeout.String())
	} else {
		timeout = config.Timeout.Value
	}

	var verbose bool
	if config.Verbose.Null {
		verboseStr := util.MultiGetenv([]string{"EC_VERBOSE"}, "")
		if verbose, err = util.StringToBool(verboseStr); err != nil {
			res.Diagnostics.AddWarning(
				"Unable to create client",
				fmt.Sprintf("Invalid value %v for verbose", verboseStr),
			)
			return
		}
	} else {
		verbose = config.Verbose.Value
	}

	var verboseCredentials bool
	if config.VerboseCredentials.Null {
		verboseCredentialsStr := util.MultiGetenv([]string{"EC_VERBOSE_CREDENTIALS"}, "")
		if verboseCredentials, err = util.StringToBool(verboseCredentialsStr); err != nil {
			res.Diagnostics.AddWarning(
				"Unable to create client",
				fmt.Sprintf("Invalid value %v for verboseCredentials", verboseCredentialsStr),
			)
			return
		}
	} else {
		verboseCredentials = config.VerboseCredentials.Value
	}

	var verboseFile string
	if config.VerboseFile.Null {
		verboseFile = util.MultiGetenv([]string{"EC_VERBOSE_FILE"}, "request.log")
	} else {
		verboseFile = config.VerboseFile.Value
	}
}

func (p *ecProvider) GetResources(_ context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
	return map[string]provider.ResourceType{}, nil
}

func (p *ecProvider) GetDataSources(_ context.Context) (map[string]provider.DataSourceType, diag.Diagnostics) {
	return map[string]provider.DataSourceType{}, nil
}
