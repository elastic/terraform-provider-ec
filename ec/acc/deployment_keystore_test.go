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
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeployment_keystore(t *testing.T) {
	depResName := "ec_deployment.test"
	keystoreResName := "ec_deployment_elasticsearch_keystore.test"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	configs := []string{
		fixtureAccDeploymentResourceBasicDefaults(t, "testdata/deployment_keystore_create.tf", randomName, getRegion(), "default"),
		fixtureAccDeploymentResourceBasicDefaults(t, "testdata/deployment_keystore_update1.tf", randomName, getRegion(), "default"),
		fixtureAccDeploymentResourceBasicDefaults(t, "testdata/deployment_keystore_update2.tf", randomName, getRegion(), "default"),
		fixtureAccDeploymentResourceBasicDefaults(t, "testdata/deployment_keystore_update3.tf", randomName, getRegion(), "default"),
		fixtureAccDeploymentResourceBasicDefaults(t, "testdata/deployment_keystore_update4.tf", randomName, getRegion(), "default"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				// test deployment creation with an embedded keystore secret and an appropriate entry in ES config
				Config: configs[0],
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(depResName, "elasticsearch.keystore_contents.%", "1"),
				),
			},
			{
				// add `ec_deployment_elasticsearch_keystore resource` with another secret
				Config: configs[1],
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(depResName, "elasticsearch.keystore_contents.%", "1"),

					resource.TestCheckResourceAttr(keystoreResName, "setting_name", "xpack.notification.slack.account.monitoring.secure_url"),
				),
			},
			{
				// remove the deployment's keystore entry and the appropriate entry in ES config
				// test that such removal doesn't affect the secret in `ec_deployment_elasticsearch_keystore` resource
				Config: configs[2],
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(depResName, "elasticsearch.keystore_contents"),

					resource.TestCheckResourceAttr(keystoreResName, "setting_name", "xpack.notification.slack.account.monitoring.secure_url"),
				),
			},
			{
				// test deployment update with new embedded keystore secret and an apppropirate ES config entry
				Config: configs[3],
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(depResName, "elasticsearch.keystore_contents.%", "1"),

					resource.TestCheckResourceAttr(keystoreResName, "setting_name", "xpack.notification.slack.account.monitoring.secure_url"),
				),
			},
			{
				// remove `ec_deployment_elasticsearch_keystore` resource
				// test that such removal doesn't affect the embedded secret
				Config: configs[4],
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(depResName, "elasticsearch.keystore_contents.%", "1"),
				),
			},
		},
	})
}
