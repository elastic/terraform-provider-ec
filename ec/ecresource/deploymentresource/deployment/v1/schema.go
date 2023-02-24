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

package v1

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	apmv1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/apm/v1"
	elasticsearchv1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v1"
	enterprisesearchv1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/enterprisesearch/v1"
	integrationsserverv1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/integrationsserver/v1"
	kibanav1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/kibana/v1"
	observabilityv1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/observability/v1"
)

func DeploymentSchema() tfsdk.Schema {
	return tfsdk.Schema{
		Version: 1,
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Elastic Cloud Deployment resource",

		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "Unique identifier of this resource.",
			},
			"alias": {
				Type:     types.StringType,
				Computed: true,
				Optional: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"version": {
				Type:        types.StringType,
				Description: "Elastic Stack version to use for all of the deployment resources.",
				Required:    true,
			},
			"region": {
				Type:        types.StringType,
				Description: `Region when the deployment should be hosted. For ECE environments this should be set to "ece-region".`,
				Required:    true,
			},
			"deployment_template_id": {
				Type:        types.StringType,
				Description: "Deployment Template identifier to base the deployment from.",
				Required:    true,
			},
			"name": {
				Type:        types.StringType,
				Description: "Name for the deployment.",
				Optional:    true,
			},
			"request_id": {
				Type:        types.StringType,
				Description: "request_id to set on the create operation, only used when a previous create attempt returns an error including a request_id.",
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
			},
			"elasticsearch_password": {
				Type:        types.StringType,
				Description: "Password for authenticating to the Elasticsearch resource",
				Computed:    true,
				Sensitive:   true,
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
				Description: "Optional list of traffic filters to apply to this deployment.",
			},
			"tags": {
				Description: "Optional map of deployment tags",
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
			"elasticsearch":       elasticsearchv1.ElasticsearchSchema(),
			"kibana":              kibanav1.KibanaSchema(),
			"apm":                 apmv1.ApmSchema(),
			"integrations_server": integrationsserverv1.IntegrationsServerSchema(),
			"enterprise_search":   enterprisesearchv1.EnterpriseSearchSchema(),
			"observability":       observabilityv1.ObservabilitySchema(),
		},
	}
}
