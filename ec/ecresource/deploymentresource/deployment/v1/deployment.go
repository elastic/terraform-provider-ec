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
	"github.com/hashicorp/terraform-plugin-framework/types"

	apmv1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/apm/v1"
	elasticsearchv1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v1"
	enterprisesearchv1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/enterprisesearch/v1"
	integrationsserverv1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/integrationsserver/v1"
	kibanav1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/kibana/v1"
	observabilityv1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/observability/v1"
)

type DeploymentTF struct {
	Id                    types.String `tfsdk:"id"`
	Alias                 types.String `tfsdk:"alias"`
	Version               types.String `tfsdk:"version"`
	Region                types.String `tfsdk:"region"`
	DeploymentTemplateId  types.String `tfsdk:"deployment_template_id"`
	Name                  types.String `tfsdk:"name"`
	RequestId             types.String `tfsdk:"request_id"`
	ElasticsearchUsername types.String `tfsdk:"elasticsearch_username"`
	ElasticsearchPassword types.String `tfsdk:"elasticsearch_password"`
	ApmSecretToken        types.String `tfsdk:"apm_secret_token"`
	TrafficFilter         types.Set    `tfsdk:"traffic_filter"`
	Tags                  types.Map    `tfsdk:"tags"`
	Elasticsearch         types.List   `tfsdk:"elasticsearch"`
	Kibana                types.List   `tfsdk:"kibana"`
	Apm                   types.List   `tfsdk:"apm"`
	IntegrationsServer    types.List   `tfsdk:"integrations_server"`
	EnterpriseSearch      types.List   `tfsdk:"enterprise_search"`
	Observability         types.List   `tfsdk:"observability"`
}

type Deployment struct {
	Id                    string                                   `tfsdk:"id"`
	Alias                 string                                   `tfsdk:"alias"`
	Version               string                                   `tfsdk:"version"`
	Region                string                                   `tfsdk:"region"`
	DeploymentTemplateId  string                                   `tfsdk:"deployment_template_id"`
	Name                  string                                   `tfsdk:"name"`
	RequestId             string                                   `tfsdk:"request_id"`
	ElasticsearchUsername string                                   `tfsdk:"elasticsearch_username"`
	ElasticsearchPassword string                                   `tfsdk:"elasticsearch_password"`
	ApmSecretToken        *string                                  `tfsdk:"apm_secret_token"`
	TrafficFilter         []string                                 `tfsdk:"traffic_filter"`
	Tags                  map[string]string                        `tfsdk:"tags"`
	Elasticsearch         elasticsearchv1.Elasticsearches          `tfsdk:"elasticsearch"`
	Kibana                kibanav1.Kibanas                         `tfsdk:"kibana"`
	Apm                   apmv1.Apms                               `tfsdk:"apm"`
	IntegrationsServer    integrationsserverv1.IntegrationsServers `tfsdk:"integrations_server"`
	EnterpriseSearch      enterprisesearchv1.EnterpriseSearches    `tfsdk:"enterprise_search"`
	Observability         observabilityv1.Observabilities          `tfsdk:"observability"`
}
