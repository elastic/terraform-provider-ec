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
	"net/http"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/apierror"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/trafficfilterapi"
	"github.com/elastic/cloud-sdk-go/pkg/client/deployments_traffic_filter"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Read queries the remote deployment traffic filter ruleset state and update
// the local state.
func read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var client = meta.(*api.API)

	res, err := trafficfilterapi.Get(trafficfilterapi.GetParams{
		API: client, ID: d.Id(),
	})
	if err != nil {
		if ruleNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if err := modelToState(d, res); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func ruleNotFound(err error) bool {
	// We're using the As() call since we do not care about the error value
	// but do care about the error type since it's an implicit 404.
	var notDeploymentNotFound *deployments_traffic_filter.GetTrafficFilterRulesetNotFound
	if errors.As(err, &notDeploymentNotFound) {
		return true
	}

	// We also check for the case where a 403 is thrown for ESS.
	return apierror.IsRuntimeStatusCode(err, http.StatusForbidden)
}
