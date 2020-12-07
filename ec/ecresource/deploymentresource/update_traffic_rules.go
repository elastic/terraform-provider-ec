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
	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/trafficfilterapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func handleTrafficFilterChange(d *schema.ResourceData, client *api.API) error {
	if !d.HasChange("traffic_filter") {
		return nil
	}

	var additions, deletions = getChange(d.GetChange("traffic_filter"))
	for _, ruleID := range additions.List() {
		if err := associateRule(ruleID.(string), d.Id(), client); err != nil {
			return err
		}
	}

	for _, ruleID := range deletions.List() {
		if err := removeRule(ruleID.(string), d.Id(), client); err != nil {
			return err
		}
	}

	return nil
}

func getChange(oldInterface, newInterface interface{}) (add, delete *schema.Set) {
	var old, new *schema.Set
	if s, ok := oldInterface.(*schema.Set); ok {
		old = s
	}
	if s, ok := newInterface.(*schema.Set); ok {
		new = s
	}

	add = new.Difference(old)
	delete = old.Difference(new)

	return add, delete
}

func associateRule(ruleID, deploymentID string, client *api.API) error {
	res, err := trafficfilterapi.Get(trafficfilterapi.GetParams{
		API: client, ID: ruleID, IncludeAssociations: true,
	})
	if err != nil {
		return err
	}

	// When the rule has already been associated, return.
	for _, assoc := range res.Associations {
		if deploymentID == *assoc.ID {
			return nil
		}
	}

	// Create assignment.
	if err := trafficfilterapi.CreateAssociation(trafficfilterapi.CreateAssociationParams{
		API: client, ID: ruleID, EntityType: "deployment", EntityID: deploymentID,
	}); err != nil {
		return err
	}
	return nil
}

func removeRule(ruleID, deploymentID string, client *api.API) error {
	res, err := trafficfilterapi.Get(trafficfilterapi.GetParams{
		API: client, ID: ruleID, IncludeAssociations: true,
	})

	// If the rule is gone (403 or 404), return nil.
	if err != nil {
		if util.TrafficFilterNotFound(err) {
			return nil
		}
		return err
	}

	// If the rule is found, then delete the association.
	for _, assoc := range res.Associations {
		if deploymentID == *assoc.ID {
			return trafficfilterapi.DeleteAssociation(trafficfilterapi.DeleteAssociationParams{
				API:        client,
				ID:         ruleID,
				EntityID:   *assoc.ID,
				EntityType: *assoc.EntityType,
			})
		}
	}

	return nil
}
