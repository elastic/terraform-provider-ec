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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func modelToState(d *schema.ResourceData, res *models.TrafficFilterRulesetInfo) error {
	if err := d.Set("name", *res.Name); err != nil {
		return err
	}

	if err := d.Set("region", *res.Region); err != nil {
		return err
	}

	if err := d.Set("type", *res.Type); err != nil {
		return err
	}

	if err := d.Set("rule", flattenRules(res.Rules)); err != nil {
		return err
	}

	if err := d.Set("include_by_default", res.IncludeByDefault); err != nil {
		return err
	}

	if res.Description != "" {
		if err := d.Set("description", res.Description); err != nil {
			return err
		}
	}

	return nil
}

func flattenRules(rules []*models.TrafficFilterRule) *schema.Set {
	result := schema.NewSet(trafficFilterRuleHash, []interface{}{})
	for _, rule := range rules {
		var m = make(map[string]interface{})
		m["source"] = rule.Source

		if rule.Description != "" {
			m["description"] = rule.Description
		}

		if rule.ID != "" {
			m["id"] = rule.ID
		}

		result.Add(m)
	}

	return result
}
