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
	"github.com/elastic/cloud-sdk-go/pkg/client/deployments_traffic_filter"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	res, err := client.V1API.DeploymentsTrafficFilter.GetTrafficFilterRuleset(
		deployments_traffic_filter.NewGetTrafficFilterRulesetParams().
			WithRulesetID(ruleID).
			WithIncludeAssociations(ec.Bool(true)),
		client.AuthWriter,
	)
	if err != nil {
		return api.UnwrapError(err)
	}

	// When the rule has already been associated, return.
	for _, assoc := range res.Payload.Associations {
		if deploymentID == *assoc.ID {
			return nil
		}
	}

	// Create assignment.
	if _, err := client.V1API.DeploymentsTrafficFilter.CreateTrafficFilterRulesetAssociation(
		deployments_traffic_filter.NewCreateTrafficFilterRulesetAssociationParams().
			WithRulesetID(ruleID).
			WithBody(&models.FilterAssociation{
				EntityType: ec.String("deployment"),
				ID:         ec.String(deploymentID),
			}),
		client.AuthWriter,
	// Due to an API bug where there's a mismatch between the spec'ed response
	// status code and the real one, we need to do this.
	); err != nil && err.Error() != "unknown error (status 201): {}" {
		return api.UnwrapError(err)
	}
	return nil
}

func removeRule(ruleID, deploymentID string, client *api.API) error {
	res, err := client.V1API.DeploymentsTrafficFilter.GetTrafficFilterRuleset(
		deployments_traffic_filter.NewGetTrafficFilterRulesetParams().
			WithRulesetID(ruleID).
			WithIncludeAssociations(ec.Bool(true)),
		client.AuthWriter,
	)

	// Removal is a little bit more hairy, the rule might have already been
	// destroyed and the associated Traffic Filter associations too, so if an
	// error is returned, fist, check if exist by iterating over all of the
	// existing rules sets since the GET <rule id> returned with an error.
	// If the rule set doesn't exist, then nil is returned.
	if err != nil {
		r, e := client.V1API.DeploymentsTrafficFilter.GetTrafficFilterRulesets(
			deployments_traffic_filter.NewGetTrafficFilterRulesetsParams(),
			client.AuthWriter,
		)
		if e != nil {
			return api.UnwrapError(e)
		}
		for _, ruleSet := range r.Payload.Rulesets {
			if *ruleSet.ID == ruleID {
				return api.UnwrapError(err)
			}
		}
		return nil
	}

	// If the rule is found, then delete the association.
	for _, assoc := range res.Payload.Associations {
		if deploymentID == *assoc.ID {
			return api.ReturnErrOnly(
				client.V1API.DeploymentsTrafficFilter.DeleteTrafficFilterRulesetAssociation(
					deployments_traffic_filter.NewDeleteTrafficFilterRulesetAssociationParams().
						WithRulesetID(ruleID).
						WithAssociatedEntityID(*assoc.ID).
						WithAssociationType(*assoc.EntityType),
					client.AuthWriter,
				),
			)
		}
	}

	return nil
}
