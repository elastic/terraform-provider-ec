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

package projectresource

import (
	"context"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// expandTrafficFilterIdsForCreate converts a types.Set of traffic filter ID strings
// into the API's TrafficFilter slice for Create requests.
// Returns nil if the set is null/unknown or empty (API interprets nil as "no change").
func expandTrafficFilterIdsForCreate(ctx context.Context, set types.Set) *[]serverless.TrafficFilter {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}
	var filterIds []string
	set.ElementsAs(ctx, &filterIds, false)
	if len(filterIds) == 0 {
		return nil
	}
	return toTrafficFilterSlice(filterIds)
}

// expandTrafficFilterIdsForPatch converts a types.Set of traffic filter ID strings
// into the API's TrafficFilter slice for Patch/Update requests.
// Unlike the Create variant, this intentionally sends an empty slice when the set
// contains no elements, which tells the API to clear all associated traffic filters.
// Returns nil only if the set is null/unknown (i.e. the field was not configured).
func expandTrafficFilterIdsForPatch(ctx context.Context, set types.Set) *[]serverless.TrafficFilter {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}
	var filterIds []string
	set.ElementsAs(ctx, &filterIds, false)
	return toTrafficFilterSlice(filterIds)
}

func toTrafficFilterSlice(filterIds []string) *[]serverless.TrafficFilter {
	trafficFilters := make([]serverless.TrafficFilter, len(filterIds))
	for i, id := range filterIds {
		trafficFilters[i] = serverless.TrafficFilter{Id: id}
	}
	return &trafficFilters
}
