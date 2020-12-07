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
	"errors"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/trafficfilterapi"
	"github.com/elastic/cloud-sdk-go/pkg/client/deployments_traffic_filter"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

// Delete will delete an existing deployment traffic filter ruleset
func delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var client = meta.(*api.API)

	res, err := trafficfilterapi.Get(trafficfilterapi.GetParams{
		API: client, ID: d.Id(), IncludeAssociations: true,
	})
	if err != nil {
		if util.TrafficFilterNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	for _, assoc := range res.Associations {
		if err := trafficfilterapi.DeleteAssociation(trafficfilterapi.DeleteAssociationParams{
			API:        client,
			ID:         d.Id(),
			EntityID:   *assoc.ID,
			EntityType: *assoc.EntityType,
		}); err != nil {
			if !associationDeleted(err) {
				return diag.FromErr(err)
			}
		}
	}

	if err := trafficfilterapi.Delete(trafficfilterapi.DeleteParams{
		API: client, ID: d.Id(),
	}); err != nil {
		if !ruleDeleted(err) {
			return diag.FromErr(err)
		}
	}

	d.SetId("")
	return nil
}

func associationDeleted(err error) bool {
	var notFound *deployments_traffic_filter.DeleteTrafficFilterRulesetAssociationNotFound
	return errors.As(err, &notFound)
}

func ruleDeleted(err error) bool {
	var notFound *deployments_traffic_filter.DeleteTrafficFilterRulesetNotFound
	return errors.As(err, &notFound)
}
