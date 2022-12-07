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

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDeployment_integrationsServer(t *testing.T) {
	resName := "ec_deployment.basic"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	startCfg := "testdata/deployment_basic_integrations_server_1.tf"
	secondCfg := "testdata/deployment_basic_integrations_server_2.tf"
	cfg := fixtureAccDeploymentResourceBasicDefaults(t, startCfg, randomName, getRegion(), defaultTemplate)
	secondConfigCfg := fixtureAccDeploymentResourceBasicDefaults(t, secondCfg, randomName, getRegion(), defaultTemplate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				// Create an Integrations Server deployment with the default settings.
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "kibana.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "integrations_server.zone_count", "1"),
					resource.TestCheckResourceAttrSet(resName, "integrations_server.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "integrations_server.size", "1g"),
					resource.TestCheckResourceAttr(resName, "integrations_server.size_resource", "memory"),
				),
			},
			{
				// Change the Integrations Server topology (increase zone count to 2).
				Config: secondConfigCfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "kibana.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "integrations_server.zone_count", "2"),
					resource.TestCheckResourceAttrSet(resName, "integrations_server.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "integrations_server.size", "1g"),
					resource.TestCheckResourceAttr(resName, "integrations_server.size_resource", "memory"),
				),
			},
		},
	})
}
