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

// +build acceptance

package acc

import (
	"fmt"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deputil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccCheckDeploymentExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		saved, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("no deployment resource: %s", name)
		}

		if saved.Primary.ID == "" {
			return fmt.Errorf("no deployment id is set")
		}
		client, err := newAPI()
		if err != nil {
			return err
		}

		return api.ReturnErrOnly(deploymentapi.Get(deploymentapi.GetParams{
			API:          client,
			DeploymentID: saved.Primary.ID,
			QueryParams: deputil.QueryParams{
				ShowSettings: true,
				ShowPlans:    true,
				ShowMetadata: true,
			},
		}))
	}
}
