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

// This test case takes ensures that several features of the "ec_deployment"
// resource are asserted:
// * Resource defaults.
// * Resource declaration in the <kind> {} format. ("apm {}").
// * Topology field overrides over field defaults.
func TestAccDeployment_basic_defaults(t *testing.T) {
	resName := "ec_deployment.defaults"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	startCfg := "testdata/deployment_basic_defaults_1.tf"
	secondCfg := "testdata/deployment_basic_defaults_2.tf"
	thirdCfg := "testdata/deployment_basic_defaults_3.tf"
	fourthCfg := "testdata/deployment_basic_defaults_4.tf"
	cfg := testAccDeploymentResourceBasic(t, startCfg, randomName, region, deploymentVersion)
	secondConfigCfg := testAccDeploymentResourceBasic(t, secondCfg, randomName, region, deploymentVersion)
	thirdConfigCfg := testAccDeploymentResourceBasic(t, thirdCfg, randomName, region, deploymentVersion)
	hotWarmCfg := testAccDeploymentResourceBasic(t, fourthCfg, randomName, region, deploymentVersion)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				// Checks the defaults which are populated using a mix of
				// Deployment Template and schema defaults.
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.memory_per_node", "8g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_data", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ingest", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_master", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ml", "false"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "kibana.#", "1"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "kibana.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.memory_per_node", "1g"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "apm.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "1"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "enterprise_search.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.memory_per_node", "2g"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.zone_count", "1"),
				),
			},
			{
				// Add an APM resource.
				Config: secondConfigCfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					// changed
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.memory_per_node", "2g"),
					resource.TestCheckResourceAttr(resName, "apm.0.topology.0.memory_per_node", "1g"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.memory_per_node", "8g"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_data", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ingest", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_master", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ml", "false"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "kibana.#", "1"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "kibana.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "apm.#", "1"),
					resource.TestCheckResourceAttr(resName, "apm.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "apm.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "apm.0.topology.0.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "1"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "enterprise_search.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.memory_per_node", "2g"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.zone_count", "1"),
				),
			},
			{
				// Remove all resources except Elasticsearch.
				Config: thirdConfigCfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.memory_per_node", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_data", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ingest", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_master", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ml", "false"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "kibana.#", "0"),
					resource.TestCheckResourceAttr(resName, "apm.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "0"),
				),
			},
			{
				// Change the Elasticsearch resource deployment template to
				// hot warm, use defaults.
				Config: hotWarmCfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.#", "2"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.1.instance_configuration_id"),
					// Hot Warm defaults to 4g.
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.memory_per_node", "4g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.memory_per_node", "4g"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_data", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ingest", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_master", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ml", "false"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.node_type_data", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.node_type_ingest", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.node_type_master", "false"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.node_type_ml", "false"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "kibana.#", "0"),
					resource.TestCheckResourceAttr(resName, "apm.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "0"),
				),
			},
		},
	})
}
