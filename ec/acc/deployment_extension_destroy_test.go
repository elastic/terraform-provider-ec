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

	"github.com/elastic/cloud-sdk-go/pkg/api/apierror"
	"github.com/elastic/cloud-sdk-go/pkg/client/extensions"
)

func testAccExtensionDestroy(s *terraform.State) error {
	client, err := newAPI()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ec_deployment_extension" {
			continue
		}

		res, err := client.V1API.Extensions.GetExtension(
			extensions.NewGetExtensionParams().WithExtensionID(rs.Primary.ID),
			client.AuthWriter)

		// If not extension exists, api gets 403 error
		if err != nil && apierror.IsRuntimeStatusCode(err, 403) {
			continue
		}

		return fmt.Errorf("extension (%s) still exists", *res.Payload.ID)
	}

	return nil
}
