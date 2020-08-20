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

func TestAccDeployment_enterpriseSearch(t *testing.T) {
	resName := "ec_deployment.enterprise_search"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	const startCfg = "testdata/deployment_enterprise_search.tf"
	const topologyConfig = "testdata/deployment_enterprise_search_topology_config.tf"
	const topConfig = "testdata/deployment_enterprise_search_top_config.tf"
	cfg := testAccDeploymentResourceEnterpriseSearch(t, startCfg, randomName, region, deploymentVersion)
	topologyConfigCfg := testAccDeploymentResourceEnterpriseSearch(t, topologyConfig, randomName, region, deploymentVersion)
	topConfigCfg := testAccDeploymentResourceEnterpriseSearch(t, topConfig, randomName, region, deploymentVersion)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: checkEnterpriseSearchDeploymentResource(resName, randomName,
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.config.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.config.#", "0"),
				),
			},
			// Ensure that no diff is generated.
			{Config: cfg, PlanOnly: true},
			{
				Config: topologyConfigCfg,
				Check: checkEnterpriseSearchDeploymentResource(resName, randomName,
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.config.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.config.#", "1"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.config.0.user_settings_yaml", "ent_search.auth.source: standard"),
				),
			},
			// Ensure that no diff is generated.
			{Config: topologyConfigCfg, PlanOnly: true},
			{
				Config: topConfigCfg,
				Check: checkEnterpriseSearchDeploymentResource(resName, randomName,
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.config.#", "1"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.config.0.user_settings_yaml", "ent_search.auth.source: standard"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.config.#", "0"),
				),
			},
			// Ensure that no diff is generated.
			{Config: topConfigCfg, PlanOnly: true},
			{
				Config: cfg,
				Check: checkEnterpriseSearchDeploymentResource(resName, randomName,
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.config.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.config.#", "0"),
				),
			},
			// Ensure that no diff is generated.
			{Config: cfg, PlanOnly: true},
			// TODO: Import case when import is ready.
		},
	})
}

func testAccDeploymentResourceEnterpriseSearch(t *testing.T, fileName, name, region, version string) string {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		name, region, version,
	)
}

func checkEnterpriseSearchDeploymentResource(resName, randomDeploymentName string, checks ...resource.TestCheckFunc) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		testAccCheckDeploymentExists(resName),
		resource.TestCheckResourceAttr(resName, "name", randomDeploymentName),
		resource.TestCheckResourceAttr(resName, "region", region),
		resource.TestCheckResourceAttr(resName, "enterprise_search.#", "1"),
		resource.TestCheckResourceAttr(resName, "enterprise_search.0.version", deploymentVersion),
		resource.TestCheckResourceAttr(resName, "enterprise_search.0.region", region),
		resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.memory_per_node", "2g"),
		resource.TestCheckResourceAttrSet(resName, "enterprise_search.0.http_endpoint"),
		resource.TestCheckResourceAttrSet(resName, "enterprise_search.0.https_endpoint"),
		resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
		resource.TestCheckResourceAttr(resName, "elasticsearch.0.version", deploymentVersion),
		resource.TestCheckResourceAttr(resName, "elasticsearch.0.region", region),
		resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.memory_per_node", "1g"),
		resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.http_endpoint"),
		resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.https_endpoint"),
		resource.TestCheckResourceAttr(resName, "kibana.#", "1"),
		resource.TestCheckResourceAttr(resName, "kibana.0.version", deploymentVersion),
		resource.TestCheckResourceAttr(resName, "kibana.0.region", region),
		resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.memory_per_node", "1g"),
		resource.TestCheckResourceAttrSet(resName, "kibana.0.http_endpoint"),
		resource.TestCheckResourceAttrSet(resName, "kibana.0.https_endpoint"),
		resource.ComposeAggregateTestCheckFunc(checks...),
	)
}
