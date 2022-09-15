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

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/eskeystoreapi"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
)

func testAccDeploymentElasticsearchKeystoreDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	client, err := newAPI()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ec_deployment_elasticsearch_keystore" {
			continue
		}

		res, err := eskeystoreapi.Get(eskeystoreapi.GetParams{
			API:          client,
			DeploymentID: rs.Primary.Attributes["deployment_id"],
		})

		if err == nil || res != nil {
			return multierror.NewPrefixed("ec_deployment_elasticsearch_keystore found",
				fmt.Errorf("deployment (%s) still exists", rs.Primary.Attributes["deployment_id"]),
			)
		}
	}

	return nil
}
