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

func TestAccDeployment_appsearch(t *testing.T) {
	resName := "ec_deployment.appsearch"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	const startCfg = "testdata/deployment_appsearch.tf"
	const topologyConfig = "testdata/deployment_appsearch_topology_config.tf"
	const topConfig = "testdata/deployment_appsearch_top_config.tf"
	cfg := testAccDeploymentResourceAppsearch(t, startCfg, randomName, region, deploymentVersionAppsearch)
	topologyConfigCfg := testAccDeploymentResourceBasic(t, topologyConfig, randomName, region, deploymentVersionAppsearch)
	topConfigCfg := testAccDeploymentResourceBasic(t, topConfig, randomName, region, deploymentVersionAppsearch)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: checkAppSearchDeploymentResource(resName, randomName,
					resource.TestCheckResourceAttr(resName, "appsearch.0.config.#", "0"),
					resource.TestCheckResourceAttr(resName, "appsearch.0.topology.0.config.#", "0"),
				),
			},
			// Ensure that no diff is generated.
			{Config: cfg, PlanOnly: true},
			{
				Config: topologyConfigCfg,
				Check: checkAppSearchDeploymentResource(resName, randomName,
					resource.TestCheckResourceAttr(resName, "appsearch.0.config.#", "0"),
					resource.TestCheckResourceAttr(resName, "appsearch.0.topology.0.config.#", "1"),
					resource.TestCheckResourceAttr(resName, "appsearch.0.topology.0.config.0.user_settings_yaml", "app_search.auth.source: standard"),
				),
			},
			// Ensure that no diff is generated.
			{Config: topologyConfigCfg, PlanOnly: true},
			{
				Config: topConfigCfg,
				Check: checkAppSearchDeploymentResource(resName, randomName,
					resource.TestCheckResourceAttr(resName, "appsearch.0.config.#", "1"),
					resource.TestCheckResourceAttr(resName, "appsearch.0.config.0.user_settings_yaml", "app_search.auth.source: standard"),
					resource.TestCheckResourceAttr(resName, "appsearch.0.topology.0.config.#", "0"),
				),
			},
			// Ensure that no diff is generated.
			{Config: topConfigCfg, PlanOnly: true},
			{
				Config: cfg,
				Check: checkAppSearchDeploymentResource(resName, randomName,
					resource.TestCheckResourceAttr(resName, "appsearch.0.config.#", "0"),
					resource.TestCheckResourceAttr(resName, "appsearch.0.topology.0.config.#", "0"),
				),
			},
			// Ensure that no diff is generated.
			{Config: cfg, PlanOnly: true},
		},
	})
}

func testAccDeploymentResourceAppsearch(t *testing.T, fileName, name, region, version string) string {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		name, region, version,
	)
}

func checkAppSearchDeploymentResource(resName, randomDeploymentName string, checks ...resource.TestCheckFunc) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		testAccCheckDeploymentExists(resName),
		resource.TestCheckResourceAttr(resName, "name", randomDeploymentName),
		resource.TestCheckResourceAttr(resName, "region", region),
		resource.TestCheckResourceAttr(resName, "appsearch.#", "1"),
		resource.TestCheckResourceAttr(resName, "appsearch.0.version", deploymentVersionAppsearch),
		resource.TestCheckResourceAttr(resName, "appsearch.0.region", region),
		resource.TestCheckResourceAttr(resName, "appsearch.0.topology.0.memory_per_node", "2g"),
		resource.TestCheckResourceAttrSet(resName, "appsearch.0.http_endpoint"),
		resource.TestCheckResourceAttrSet(resName, "appsearch.0.https_endpoint"),
		resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
		resource.TestCheckResourceAttr(resName, "elasticsearch.0.version", deploymentVersionAppsearch),
		resource.TestCheckResourceAttr(resName, "elasticsearch.0.region", region),
		resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.memory_per_node", "1g"),
		resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.http_endpoint"),
		resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.https_endpoint"),
		resource.TestCheckResourceAttr(resName, "kibana.#", "1"),
		resource.TestCheckResourceAttr(resName, "kibana.0.version", deploymentVersionAppsearch),
		resource.TestCheckResourceAttr(resName, "kibana.0.region", region),
		resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.memory_per_node", "1g"),
		resource.TestCheckResourceAttrSet(resName, "kibana.0.http_endpoint"),
		resource.TestCheckResourceAttrSet(resName, "kibana.0.https_endpoint"),
		resource.ComposeAggregateTestCheckFunc(checks...),
	)
}
