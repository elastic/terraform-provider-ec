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
	"io/ioutil"
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
	cfg := fixtureAccDeploymentResourceBasicDefaults(t, startCfg, randomName, getRegion(), defaultTemplate)
	secondConfigCfg := fixtureAccDeploymentResourceBasicDefaults(t, secondCfg, randomName, getRegion(), defaultTemplate)
	thirdConfigCfg := fixtureAccDeploymentResourceBasicDefaults(t, thirdCfg, randomName, getRegion(), defaultTemplate)

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
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size", "8g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_data", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ingest", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_master", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ml", "false"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "kibana.#", "1"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "kibana.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.size", "1g"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "apm.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "1"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "enterprise_search.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.size", "2g"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.zone_count", "1"),
				),
			},
			{
				// Add an APM resource.
				Config: secondConfigCfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					// changed
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.size", "2g"),
					resource.TestCheckResourceAttr(resName, "apm.0.topology.0.size", "1g"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size", "8g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_data", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ingest", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_master", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ml", "false"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "kibana.#", "1"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "kibana.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "apm.#", "1"),
					resource.TestCheckResourceAttr(resName, "apm.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "apm.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "apm.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "apm.0.topology.0.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "1"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "enterprise_search.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.size", "2g"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.zone_count", "1"),
				),
			},
			{
				// Remove all resources except Elasticsearch and Kibana.
				Config: thirdConfigCfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_data", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ingest", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_master", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ml", "false"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "kibana.#", "1"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "kibana.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.size_resource", "memory"),

					// In this test we're verifying that the topology for Kibana is not reset.
					// This is due to the terraform SDK stickyness where a removed computed block
					// with a previous value is the same as an empty block, so previous computed
					// values are used.
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.size", "2g"),

					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "apm.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "0"),
				),
			},
		},
	})
}

func TestAccDeployment_basic_defaults_hw(t *testing.T) {
	resName := "ec_deployment.defaults"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	startCfg := "testdata/deployment_basic_defaults_hw_1.tf"
	secondCfg := "testdata/deployment_basic_defaults_hw_2.tf"
	cfg := fixtureAccDeploymentResourceBasicDefaults(t, startCfg, randomName, getRegion(), defaultTemplate)
	hotWarmCfg := fixtureAccDeploymentResourceBasicDefaults(t, secondCfg, randomName, getRegion(), hotWarmTemplate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				// Create a deployment which only uses Elasticsearch resources
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size_resource", "memory"),
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
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size", "4g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.size", "4g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.size_resource", "memory"),
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
					resource.TestCheckResourceAttr(resName, "kibana.#", "1"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.#", "1"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.size", "1g"),
					resource.TestCheckResourceAttrSet(resName, "kibana.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "apm.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "0"),
				),
			},
		},
	})
}

func fixtureAccDeploymentResourceBasicDefaults(t *testing.T, fileName, name, region, depTpl string) string {
	t.Helper()

	deploymentTpl := setDefaultTemplate(region, depTpl)
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		region, name, region, deploymentTpl,
	)
}
