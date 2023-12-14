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

package v2

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	apmv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/apm/v2"
	elasticsearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
	enterprisesearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/enterprisesearch/v2"
	integrationsserverv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/integrationsserver/v2"
	kibanav2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/kibana/v2"
	observabilityv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/observability/v2"
	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifiers"
)

func DeploymentSchema() schema.Schema {
	return schema.Schema{
		Version:             2,
		MarkdownDescription: "Provides an Elastic Cloud deployment resource, which allows deployments to be created, updated, and deleted.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of this deployment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"alias": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "Deployment alias, affects the format of the resource URLs.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				Description: `Elastic Stack version to use for all of the deployment resources.

-> Read the [ESS stack version policy](https://www.elastic.co/guide/en/cloud/current/ec-version-policy.html#ec-version-policy-available) to understand which versions are available.`,
				Required: true,
				Validators: []validator.String{
					isVersion{},
				},
			},
			"region": schema.StringAttribute{
				Description: "Elasticsearch Service (ESS) region where the deployment should be hosted. For Elastic Cloud Enterprise (ECE) installations, set to `\"ece-region\".",
				Required:    true,
			},
			"deployment_template_id": schema.StringAttribute{
				Description: "Deployment template identifier to create the deployment from. See the [full list](https://www.elastic.co/guide/en/cloud/current/ec-regions-templates-instances.html) of regions and deployment templates available in ESS.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name for the deployment",
				Optional:    true,
			},
			"request_id": schema.StringAttribute{
				Description: "Request ID to set when you create the deployment. Use it only when previous attempts return an error and `request_id` is returned as part of the error.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"elasticsearch_username": schema.StringAttribute{
				Description: "Username for authenticating to the Elasticsearch resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"elasticsearch_password": schema.StringAttribute{
				Description: `Password for authenticating to the Elasticsearch resource.

~> **Note on deployment credentials** The <code>elastic</code> user credentials are only available whilst creating a deployment. Importing a deployment will not import the <code>elasticsearch_username</code> or <code>elasticsearch_password</code> attributes.
~> **Note on deployment credentials in state** The <code>elastic</code> user credentials are stored in the state file as plain text. Please follow the official Terraform recommendations regarding senstaive data in state.`,
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					setUnknownIfResetPasswordIsTrue{},
				},
			},
			"apm_secret_token": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					UseNullUnlessAddingAPMOrIntegrationsServer(),
				},
			},
			"traffic_filter": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "List of traffic filters rule identifiers that will be applied to the deployment.",
				PlanModifiers: []planmodifier.Set{
					planmodifiers.SetDefaultValue(types.StringType, []attr.Value{}),
				},
			},
			"tags": schema.MapAttribute{
				Description: "Optional map of deployment tags",
				ElementType: types.StringType,
				Optional:    true,
			},
			"reset_elasticsearch_password": schema.BoolAttribute{
				Description: "Explicitly resets the elasticsearch_password when true",
				Optional:    true,
			},
			"migrate_to_latest_hardware": schema.BoolAttribute{
				Description: "When true, updates deployment according to the latest deployment template values.",
				Optional:    true,
			},
			"elasticsearch":       elasticsearchv2.ElasticsearchSchema(),
			"kibana":              kibanav2.KibanaSchema(),
			"apm":                 apmv2.ApmSchema(),
			"integrations_server": integrationsserverv2.IntegrationsServerSchema(),
			"enterprise_search":   enterprisesearchv2.EnterpriseSearchSchema(),
			"observability":       observabilityv2.ObservabilitySchema(),
		},
	}
}
