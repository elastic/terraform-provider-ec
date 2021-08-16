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

package trafficfilterresource

import (
	"bytes"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// newSchema returns the schema for an "ec_deployment_traffic_filter" resource.
func newSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Description: "Required name of the ruleset",
			Required:    true,
		},
		"type": {
			Type:        schema.TypeString,
			Description: `Required type of the ruleset ("ip", "vpce" or "azure_private_endpoint")`,
			Required:    true,
		},
		"region": {
			Type:        schema.TypeString,
			Description: "Required filter region, the ruleset can only be attached to deployments in the specific region",
			Required:    true,
		},
		"rule": {
			Type:        schema.TypeSet,
			Set:         trafficFilterRuleHash,
			Description: "Required list of rules, which the ruleset is made of.",
			Required:    true,
			MinItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"source": {
						Type:        schema.TypeString,
						Description: "Required traffic filter source: IP address, CIDR mask, or VPC endpoint ID, not required when the type is azure_private_endpoint",
						Optional:    true,
					},

					"description": {
						Type:        schema.TypeString,
						Description: "Optional rule description",
						Optional:    true,
					},

					"azure_endpoint_name": {
						Type:        schema.TypeString,
						Description: "Optional Azure endpoint name",
						Optional:    true,
					},

					"azure_endpoint_guid": {
						Type:        schema.TypeString,
						Description: "Optional Azure endpoint GUID",
						Optional:    true,
					},

					"id": {
						Type:        schema.TypeString,
						Description: "Computed rule ID",
						Computed:    true,
					},
				},
			},
		},

		"include_by_default": {
			Type:        schema.TypeBool,
			Description: "Should the ruleset be automatically included in the new deployments (Defaults to false)",
			Optional:    true,
			Default:     false,
		},
		"description": {
			Type:        schema.TypeString,
			Description: "Optional ruleset description",
			Optional:    true,
		},
	}
}

func trafficFilterRuleHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if m["source"] != nil {
		buf.WriteString(m["source"].(string))
	}
	if m["description"] != nil {
		buf.WriteString(m["description"].(string))
	}
	if m["azure_endpoint_name"] != nil {
		buf.WriteString(m["azure_endpoint_name"].(string))
	}
	if m["azure_endpoint_guid"] != nil {
		buf.WriteString(m["azure_endpoint_guid"].(string))
	}
	return schema.HashString(buf.String())
}
