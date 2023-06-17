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
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/terraform-provider-ec/ec/ecdatasource/deploymentdatasource"
	"github.com/elastic/terraform-provider-ec/ec/ecdatasource/deploymentsdatasource"
	"github.com/elastic/terraform-provider-ec/ec/ecdatasource/privatelinkdatasource"
	"github.com/elastic/terraform-provider-ec/ec/ecdatasource/stackdatasource"
	"github.com/elastic/terraform-provider-ec/ec/ecdatasource/trafficfilterdatasource"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/elasticsearchkeystoreresource"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/extensionresource"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/snapshotrepositoryresource"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/trafficfilterassocresource"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/trafficfilterresource"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/elastic/terraform-provider-ec/ec/internal/validators"
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

func New(version string) provider.Provider {
	return &Provider{version: version}
}

func ProviderWithClient(client *api.API, version string) provider.Provider {
	return &Provider{client: client, version: version}
}

var _ provider.Provider = (*Provider)(nil)

type Provider struct {
	version string
	client  *api.API
}

func (p *Provider) Metadata(ctx context.Context, request provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "ec"
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		func() datasource.DataSource { return &deploymentdatasource.DataSource{} },
		func() datasource.DataSource { return &deploymentsdatasource.DataSource{} },
		func() datasource.DataSource { return &stackdatasource.DataSource{} },
		func() datasource.DataSource { return &trafficfilterdatasource.DataSource{} },
		privatelinkdatasource.AwsDataSource,
		privatelinkdatasource.GcpDataSource,
		privatelinkdatasource.AzureDataSource,
	}
}

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource { return &elasticsearchkeystoreresource.Resource{} },
		func() resource.Resource { return &extensionresource.Resource{} },
		func() resource.Resource { return &deploymentresource.Resource{} },
		func() resource.Resource { return &snapshotrepositoryresource.Resource{} },
		func() resource.Resource { return &trafficfilterresource.Resource{} },
		func() resource.Resource { return &trafficfilterassocresource.Resource{} },
	}
}

func (p *Provider) Schema(_ context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: fmt.Sprintf(endpointDesc, api.ESSEndpoint),
				Optional:    true,
				Validators: []validator.String{
					validators.Known(),
					validators.IsURLWithSchemeValidator(validURLSchemes),
				},
			},
			"apikey": schema.StringAttribute{
				Description: apikeyDesc,
				Optional:    true,
				Sensitive:   true,
			},
			"username": schema.StringAttribute{
				Description: usernameDesc,
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: passwordDesc,
				Optional:    true,
				Sensitive:   true,
			},
			"insecure": schema.BoolAttribute{
				Description: insecureDesc,
				Optional:    true,
			},
			"timeout": schema.StringAttribute{
				Description: timeoutDesc,
				Optional:    true,
			},
			"verbose": schema.BoolAttribute{
				Description: verboseDesc,
				Optional:    true,
			},
			"verbose_credentials": schema.BoolAttribute{
				Description: verboseCredsDesc,
				Optional:    true,
			},
			"verbose_file": schema.StringAttribute{
				Description: timeoutDesc,
				Optional:    true,
			},
		},
	}
}

// Retrieve provider data from configuration
type providerConfig struct {
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

func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	if p.client != nil {
		// Required for unit tests, because a mock client is pre-created there.
		resp.DataSourceData = p.client
		resp.ResourceData = p.client
		return
	}

	var config providerConfig

	diags := req.Config.Get(ctx, &config)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := config.Endpoint.ValueString()

	if config.Endpoint.ValueString() == "" {
		endpoint = util.MultiGetenvOrDefault([]string{"EC_ENDPOINT", "EC_HOST"}, api.ESSEndpoint)

		diags := validateEndpoint(ctx, endpoint)

		resp.Diagnostics.Append(diags...)

		if diags.HasError() {
			return
		}
	}

	apiKey := config.ApiKey.ValueString()

	if config.ApiKey.ValueString() == "" {
		apiKey = util.MultiGetenvOrDefault([]string{"EC_API_KEY"}, "")
	}

	username := config.Username.ValueString()

	if config.Username.ValueString() == "" {
		username = util.MultiGetenvOrDefault([]string{"EC_USER", "EC_USERNAME"}, "")
	}

	password := config.Password.ValueString()

	if config.Password.ValueString() == "" {
		password = util.MultiGetenvOrDefault([]string{"EC_PASS", "EC_PASSWORD"}, "")
	}

	timeoutStr := config.Timeout.ValueString()

	if config.Timeout.ValueString() == "" {
		timeoutStr = util.MultiGetenvOrDefault([]string{"EC_TIMEOUT"}, defaultTimeout.String())
	}

	timeout, err := time.ParseDuration(timeoutStr)

	if err != nil {
		resp.Diagnostics.AddError("Unable to create client", err.Error())
		return
	}

	insecure := config.Insecure.ValueBool()

	if config.Insecure.IsNull() {
		insecureStr := util.MultiGetenvOrDefault([]string{"EC_INSECURE", "EC_SKIP_TLS_VALIDATION"}, "")

		if insecure, err = util.StringToBool(insecureStr); err != nil {
			resp.Diagnostics.AddError(
				"Unable to create client",
				fmt.Sprintf("Invalid value '%v' in 'EC_INSECURE' or 'EC_SKIP_TLS_VALIDATION'", insecureStr),
			)
			return
		}
	}

	verbose := config.Verbose.ValueBool()

	if config.Verbose.IsNull() {
		verboseStr := util.MultiGetenvOrDefault([]string{"EC_VERBOSE"}, "")

		if verbose, err = util.StringToBool(verboseStr); err != nil {
			resp.Diagnostics.AddError(
				"Unable to create client",
				fmt.Sprintf("Invalid value '%v' in 'EC_VERBOSE'", verboseStr),
			)
			return
		}
	}

	verboseCredentials := config.VerboseCredentials.ValueBool()

	if config.VerboseCredentials.IsNull() {
		verboseCredentialsStr := util.MultiGetenvOrDefault([]string{"EC_VERBOSE_CREDENTIALS"}, "")

		if verboseCredentials, err = util.StringToBool(verboseCredentialsStr); err != nil {
			resp.Diagnostics.AddError(
				"Unable to create client",
				fmt.Sprintf("Invalid value '%v' in 'EC_VERBOSE_CREDENTIALS'", verboseCredentialsStr),
			)
			return
		}
	}

	verboseFile := config.VerboseFile.ValueString()

	if config.VerboseFile.IsNull() {
		verboseFile = util.MultiGetenvOrDefault([]string{"EC_VERBOSE_FILE"}, "request.log")
	}

	cfg, err := newAPIConfig(apiSetup{
		endpoint:           endpoint,
		apikey:             apiKey,
		username:           username,
		password:           password,
		insecure:           insecure,
		timeout:            timeout,
		verbose:            verbose,
		verboseCredentials: verboseCredentials,
		verboseFile:        verboseFile,
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create api Client config",
			err.Error(),
		)
		return
	}

	client, err := api.NewAPI(cfg)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create api Client config",
			err.Error(),
		)
		return
	}

	p.client = client
	resp.DataSourceData = client
	resp.ResourceData = client
}

func validateEndpoint(ctx context.Context, endpoint string) diag.Diagnostics {
	validateReq := validator.StringRequest{
		Path:        path.Root("endpoint"),
		ConfigValue: types.StringValue(endpoint),
	}

	validateResp := validator.StringResponse{}

	validators.IsURLWithSchemeValidator(validURLSchemes).ValidateString(ctx, validateReq, &validateResp)

	return validateResp.Diagnostics
}
