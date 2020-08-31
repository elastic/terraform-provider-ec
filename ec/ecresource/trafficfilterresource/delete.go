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
	"context"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/client/deployments_traffic_filter"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Delete will delete an existing deployment traffic filter ruleset
func Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var client = meta.(*api.API)

	res, err := client.V1API.DeploymentsTrafficFilter.GetTrafficFilterRulesetDeploymentAssociations(
		deployments_traffic_filter.NewGetTrafficFilterRulesetDeploymentAssociationsParams().
			WithRulesetID(d.Id()),
		client.AuthWriter,
	)
	if err != nil {
		return diag.FromErr(api.UnwrapError(err))
	}

	for _, assoc := range res.Payload.Associations {
		if _, err := client.V1API.DeploymentsTrafficFilter.DeleteTrafficFilterRulesetAssociation(
			deployments_traffic_filter.NewDeleteTrafficFilterRulesetAssociationParams().
				WithRulesetID(d.Id()).
				WithAssociatedEntityID(*assoc.ID).
				WithAssociationType(*assoc.EntityType),
			client.AuthWriter,
		); err != nil {
			return diag.FromErr(api.UnwrapError(err))
		}
	}

	if _, err := client.V1API.DeploymentsTrafficFilter.DeleteTrafficFilterRuleset(
		deployments_traffic_filter.NewDeleteTrafficFilterRulesetParams().
			WithRulesetID(d.Id()),
		client.AuthWriter,
	); err != nil {
		return diag.FromErr(api.UnwrapError(err))
	}

	d.SetId("")
	return nil
}
