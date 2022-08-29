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

package deploymentresource

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	minimumKibanaSize             = 1024
	minimumApmSize                = 512
	minimumEnterpriseSearchSize   = 2048
	minimumIntegrationsServerSize = 1024

	minimumZoneCount = 1
)

// newSchema returns the schema for an "ec_deployment" resource.
func newSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"alias": {
			Type:        schema.TypeString,
			Description: "Optional deployment alias that affects the format of the resource URLs",
			Optional:    true,
			Computed:    true,
		},
		"version": {
			Type:        schema.TypeString,
			Description: "Required Elastic Stack version to use for all of the deployment resources",
			Required:    true,
		},
		"region": {
			Type:        schema.TypeString,
			Description: `Required ESS region where to create the deployment, for ECE environments "ece-region" must be set`,
			Required:    true,
			ForceNew:    true,
		},
		"deployment_template_id": {
			Type:        schema.TypeString,
			Description: "Required Deployment Template identifier to create the deployment from",
			Required:    true,
		},
		"name": {
			Type:        schema.TypeString,
			Description: "Optional name for the deployment",
			Optional:    true,
		},
		"request_id": {
			Type:        schema.TypeString,
			Description: "Optional request_id to set on the create operation, only use when previous create attempts return with an error and a request_id is returned as part of the error",
			Optional:    true,
		},

		// Computed ES Creds
		"elasticsearch_username": {
			Type:        schema.TypeString,
			Description: "Computed username obtained upon creating the Elasticsearch resource",
			Computed:    true,
		},
		"elasticsearch_password": {
			Type:        schema.TypeString,
			Description: "Computed password obtained upon creating the Elasticsearch resource",
			Computed:    true,
			Sensitive:   true,
		},

		// APM secret_token
		"apm_secret_token": {
			Type:      schema.TypeString,
			Computed:  true,
			Sensitive: true,
		},

		// Resources
		"elasticsearch": {
			Type:        schema.TypeList,
			Description: "Required Elasticsearch resource definition",
			MaxItems:    1,
			Required:    true,
			Elem:        newElasticsearchResource(),
		},
		"kibana": {
			Type:        schema.TypeList,
			Description: "Optional Kibana resource definition",
			Optional:    true,
			MaxItems:    1,
			Elem:        newKibanaResource(),
		},
		"apm": {
			Type:        schema.TypeList,
			Description: "Optional APM resource definition",
			Optional:    true,
			MaxItems:    1,
			Elem:        newApmResource(),
		},
		"integrations_server": {
			Type:        schema.TypeList,
			Description: "Optional Integrations Server resource definition",
			Optional:    true,
			MaxItems:    1,
			Elem:        newIntegrationsServerResource(),
		},
		"enterprise_search": {
			Type:        schema.TypeList,
			Description: "Optional Enterprise Search resource definition",
			Optional:    true,
			MaxItems:    1,
			Elem:        newEnterpriseSearchResource(),
		},

		// Settings
		"traffic_filter": {
			Description: "Optional list of traffic filters to apply to this deployment.",
			// This field is a TypeSet since the order of the items isn't
			// important, but the unique list is. This prevents infinite loops
			// for autogenerated IDs.
			Type:     schema.TypeSet,
			Set:      schema.HashString,
			Optional: true,
			MinItems: 1,
			Elem: &schema.Schema{
				MinItems: 1,
				Type:     schema.TypeString,
			},
		},
		"observability": {
			Type:        schema.TypeList,
			Description: "Optional observability settings. Ship logs and metrics to a dedicated deployment.",
			Optional:    true,
			MaxItems:    1,
			Elem:        newObservabilitySettings(),
		},

		"tags": {
			Description: "Optional map of deployment tags",
			Type:        schema.TypeMap,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}

func newObservabilitySettings() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"deployment_id": {
				Type:     schema.TypeString,
				Required: true,
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					// The terraform config can contain 'self' as a deployment target
					// However the API will return the actual deployment-id.
					// This overrides 'self' with the deployment-id so the diff will work correctly.
					var deploymentID = d.Id()
					var mappedOldValue = mapSelfToDeploymentID(oldValue, deploymentID)
					var mappedNewValue = mapSelfToDeploymentID(newValue, deploymentID)

					return mappedOldValue == mappedNewValue
				},
			},
			"ref_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"logs": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"metrics": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func mapSelfToDeploymentID(value string, deploymentID string) string {
	if value == "self" && deploymentID != "" {
		// If the deployment has a deployment-id, replace 'self' with the deployment-id
		return deploymentID
	}

	return value
}

// suppressMissingOptionalConfigurationBlock handles configuration block attributes in the following scenario:
//   - The resource schema includes an optional configuration block with defaults
//   - The API response includes those defaults to refresh into the Terraform state
//   - The operator's configuration omits the optional configuration block
func suppressMissingOptionalConfigurationBlock(k, old, new string, d *schema.ResourceData) bool {
	return old == "1" && new == "0"
}
