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
	"fmt"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/trafficfilterapi"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccDeploymentTrafficFilterDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	client, err := newAPI()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ec_deployment_traffic_filter" {
			continue
		}

		res, err := trafficfilterapi.Get(trafficfilterapi.GetParams{
			API: client,
			ID:  rs.Primary.ID,
		})

		// The resource will only exist if it can be obtained via the API and
		// the metadata status is not set to hidden. Currently ESS clients
		// cannot delete a deployment, so even when it's been shut down it will
		// show up on the GET call.
		if err == nil && res != nil {
			var merr = multierror.NewPrefixed("ec_deployment_traffic_filter found",
				fmt.Errorf("ruleset (%s) still exists", rs.Primary.ID),
			)

			return merr.ErrorOrNil()
		}
	}

	return nil
}
