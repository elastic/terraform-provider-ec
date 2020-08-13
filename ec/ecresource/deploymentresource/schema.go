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

// NewSchema returns the schema for an "ec_deployment" resource.
func NewSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"version": {
			Type:     schema.TypeString,
			Required: true,
		},
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"region": {
			Type:     schema.TypeString,
			Required: true,
		},
		"request_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"deployment_template_id": {
			Type:     schema.TypeString,
			Required: true,
		},

		// Computed ES Creds
		"elasticsearch_username": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"elasticsearch_password": {
			Type:      schema.TypeString,
			Computed:  true,
			Sensitive: true,
		},

		// Resources
		"elasticsearch": {
			Type:     schema.TypeList,
			MinItems: 1,
			MaxItems: 1,
			Required: true,
			Elem:     newElasticsearchResource(),
		},
		"kibana": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem:     newKibanaResource(),
		},
		"apm": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem:     newApmResource(),
		},
		"appsearch": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem:     newAppSearchResource(),
		},
		"enterprise_search": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem:     newEnterpriseSearchResource(),
		},
	}
}

// suppressMissingOptionalConfigurationBlock handles configuration block attributes in the following scenario:
//  * The resource schema includes an optional configuration block with defaults
//  * The API response includes those defaults to refresh into the Terraform state
//  * The operator's configuration omits the optional configuration block
func suppressMissingOptionalConfigurationBlock(k, old, new string, d *schema.ResourceData) bool {
	return old == "1" && new == "0"
}
