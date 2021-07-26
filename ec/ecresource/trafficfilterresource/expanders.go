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
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandModel(d *schema.ResourceData) *models.TrafficFilterRulesetRequest {
	var ruleSet = d.Get("rule").(*schema.Set)
	var request = models.TrafficFilterRulesetRequest{
		Name:             ec.String(d.Get("name").(string)),
		Type:             ec.String(d.Get("type").(string)),
		Region:           ec.String(d.Get("region").(string)),
		Description:      d.Get("description").(string),
		IncludeByDefault: ec.Bool(d.Get("include_by_default").(bool)),
		Rules:            make([]*models.TrafficFilterRule, 0, ruleSet.Len()),
	}

	for _, r := range ruleSet.List() {
		var m = r.(map[string]interface{})
		var rule = models.TrafficFilterRule{
			Source: m["source"].(string),
		}

		if val, ok := m["id"]; ok {
			rule.ID = val.(string)
		}

		if val, ok := m["description"]; ok {
			rule.Description = val.(string)
		}

		if val, ok := m["azure_endpoint_name"]; ok {
			rule.AzureEndpointName = val.(string)
		}

		if val, ok := m["azure_endpoint_guid"]; ok {
			rule.AzureEndpointGUID = val.(string)
		}

		request.Rules = append(request.Rules, &rule)
	}

	return &request
}
