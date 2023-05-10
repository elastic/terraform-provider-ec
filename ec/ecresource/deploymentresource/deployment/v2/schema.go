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
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	apmv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/apm/v2"
	elasticsearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
	enterprisesearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/enterprisesearch/v2"
	integrationsserverv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/integrationsserver/v2"
	kibanav2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/kibana/v2"
	observabilityv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/observability/v2"
)

func DeploymentSchema() tfsdk.Schema {
	return tfsdk.Schema{
		Version: 2,
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Provides an Elastic Cloud deployment resource, which allows deployments to be created, updated, and deleted.",

		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "Unique identifier of this deployment.",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"alias": {
				Type:        types.StringType,
				Computed:    true,
				Optional:    true,
				Description: "Deployment alias, affects the format of the resource URLs.",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"version": {
				Type: types.StringType,
				Description: `Elastic Stack version to use for all of the deployment resources.

-> Read the [ESS stack version policy](https://www.elastic.co/guide/en/cloud/current/ec-version-policy.html#ec-version-policy-available) to understand which versions are available.`,
				Required: true,
			},
			"region": {
				Type:        types.StringType,
				Description: "Elasticsearch Service (ESS) region where the deployment should be hosted. For Elastic Cloud Enterprise (ECE) installations, set to `\"ece-region\".",
				Required:    true,
			},
			"deployment_template_id": {
				Type:        types.StringType,
				Description: "Deployment template identifier to create the deployment from. See the [full list](https://www.elastic.co/guide/en/cloud/current/ec-regions-templates-instances.html) of regions and deployment templates available in ESS.",
				Required:    true,
			},
			"name": {
				Type:        types.StringType,
				Description: "Name for the deployment",
				Optional:    true,
			},
			"request_id": {
				Type:        types.StringType,
				Description: "Request ID to set when you create the deployment. Use it only when previous attempts return an error and `request_id` is returned as part of the error.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"elasticsearch_username": {
				Type:        types.StringType,
				Description: "Username for authenticating to the Elasticsearch resource.",
				Computed:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"elasticsearch_password": {
				Type: types.StringType,
				Description: `Password for authenticating to the Elasticsearch resource.

~> **Note on deployment credentials** The <code>elastic</code> user credentials are only available whilst creating a deployment. Importing a deployment will not import the <code>elasticsearch_username</code> or <code>elasticsearch_password</code> attributes.`,
				Computed:  true,
				Sensitive: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
					setUnknownIfResetPasswordIsTrue{},
				},
			},
			"apm_secret_token": {
				Type:      types.StringType,
				Computed:  true,
				Sensitive: true,
			},
			"traffic_filter": {
				Type: types.SetType{
					ElemType: types.StringType,
				},
				Optional:    true,
				Description: `List of traffic filters rule identifiers that will be applied to the deployment.`,
			},
			"tags": {
				Description: "Optional map of deployment tags",
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
			"reset_elasticsearch_password": {
				Description: "Explicitly resets the elasticsearch_password when true",
				Type:        types.BoolType,
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

type setUnknownIfResetPasswordIsTrue struct{}

var _ tfsdk.AttributePlanModifier = setUnknownIfResetPasswordIsTrue{}

func (m setUnknownIfResetPasswordIsTrue) Description(ctx context.Context) string {
	return m.MarkdownDescription(ctx)
}

func (m setUnknownIfResetPasswordIsTrue) MarkdownDescription(ctx context.Context) string {
	return "Sets the planned value to unknown if the reset_elasticsearch_password config value is true"
}

func (m setUnknownIfResetPasswordIsTrue) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if resp.AttributePlan == nil || req.AttributeConfig == nil {
		return
	}

	// if the config is the unknown value, use the unknown value otherwise, interpolation gets messed up
	if req.AttributeConfig.IsUnknown() {
		return
	}

	var isResetting *bool
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("reset_elasticsearch_password"), &isResetting)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if isResetting != nil && *isResetting {
		resp.AttributePlan = types.String{Unknown: true}
	}
}
