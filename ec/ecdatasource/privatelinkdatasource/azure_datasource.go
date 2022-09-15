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

package privatelinkdatasource

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// AzureDataSource returns the ec_gcp_privateserviceconnect_endpoint data source schema.
func AzureDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readContextFor(provider{
			name:             "azure",
			populateResource: populateAzureResource,
		}),

		Schema: newAzureSchema(),

		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func newAzureSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"region": {
			Type:     schema.TypeString,
			Required: true,
		},

		// Computed
		"service_alias": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"domain_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func populateAzureResource(regionData map[string]interface{}, d *schema.ResourceData) error {
	if err := copyToStateAs[string]("service_alias", regionData, d); err != nil {
		return err
	}

	if err := copyToStateAs[string]("domain_name", regionData, d); err != nil {
		return err
	}

	return nil
}
