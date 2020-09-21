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

package trafficfilterassocresource

import (
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flatten(res *models.TrafficFilterRulesetInfo, d *schema.ResourceData) error {
	if res == nil {
		return nil
	}

	var found bool
	deploymentID := d.Get("deployment_id").(string)
	for _, assoc := range res.Associations {
		if *assoc.EntityType == entityType && *assoc.ID == deploymentID {
			found = true
		}
	}

	if !found {
		if err := d.Set("deployment_id", ""); err != nil {
			return err
		}
		if err := d.Set("traffic_filter_id", ""); err != nil {
			return err
		}
		d.SetId("")
	}

	return nil
}
