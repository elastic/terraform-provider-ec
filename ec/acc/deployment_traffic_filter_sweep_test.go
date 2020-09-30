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

package acc

import (
	"strings"
	"sync"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/trafficfilterapi"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("ec_deployment_traffic_filter", &resource.Sweeper{
		Name: "ec_deployment_traffic_filter",
		F:    testSweepDeploymentTrafficFilter,
	})
}

func testSweepDeploymentTrafficFilter(region string) error {
	client, err := newAPI()
	if err != nil {
		return err
	}

	res, err := trafficfilterapi.List(trafficfilterapi.ListParams{
		API:    client,
		Region: region,
	})
	if err != nil {
		return api.UnwrapError(err)
	}

	var sweepFilters []string
	for _, d := range res.Rulesets {
		if strings.HasPrefix(*d.Name, prefix) {
			sweepFilters = append(sweepFilters, *d.ID)
		}
	}

	var merr = multierror.NewPrefixed("failed sweeping traffic filters")
	var wg sync.WaitGroup
	for _, dep := range sweepFilters {
		wg.Add(1)
		go func(id string) {
			if err := deleteTrafficFilter(client, id, wg.Done); err != nil {
				merr = merr.Append(err)
			}
		}(dep)
	}
	wg.Wait()

	return merr.ErrorOrNil()
}

func deleteTrafficFilter(c *api.API, filter string, done func()) error {
	defer done()
	return trafficfilterapi.Delete(trafficfilterapi.DeleteParams{
		API: c,
		ID:  filter,
	})
}
