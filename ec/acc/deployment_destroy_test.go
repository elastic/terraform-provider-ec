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

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccDeploymentDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	client, err := NewAPI()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ec_deployment" {
			continue
		}

		res, err := deploymentapi.Get(deploymentapi.GetParams{
			API:          client,
			DeploymentID: rs.Primary.ID,
		})

		// The resource will only exist if it can be obtained via the API and
		// the metadata status is not set to hidden. Currently ESS clients
		// cannot delete a deployment, so even when it's been shut down it will
		// show up on the GET call.
		if err == nil && !*res.Metadata.Hidden {
			var merr = multierror.NewPrefixed("ec_deployment found",
				fmt.Errorf("deployment (%s) still exists", rs.Primary.ID),
			)

			// If any of its subresources isn't stopped, return an error
			// which will indicate there's still dangling resources.
			const stoppedStatus = "stopped"
			if res != nil && res.Resources != nil {
				for _, res := range res.Resources.Apm {
					if *res.Info.Status == stoppedStatus {
						continue
					}
					merr = merr.Append(
						fmt.Errorf("resource apm (%s) still exists", *res.ID),
					)
				}
				for _, res := range res.Resources.Elasticsearch {
					if *res.Info.Status == stoppedStatus {
						continue
					}
					merr = merr.Append(
						fmt.Errorf("resource elasticsearch (%s) still exists", *res.ID),
					)
				}
				for _, res := range res.Resources.EnterpriseSearch {
					if *res.Info.Status == stoppedStatus {
						continue
					}
					merr = merr.Append(
						fmt.Errorf("resource enterpriseSearch (%s) still exists", *res.ID),
					)
				}
				for _, res := range res.Resources.Kibana {
					if *res.Info.Status == stoppedStatus {
						continue
					}
					merr = merr.Append(
						fmt.Errorf("resource apm (%s) still exists", *res.ID),
					)
				}
			}

			return merr.ErrorOrNil()
		}
	}

	return nil
}
