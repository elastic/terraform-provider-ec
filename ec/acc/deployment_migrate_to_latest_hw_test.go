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
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeployment_migrate_to_latest_hw(t *testing.T) {
	resName := "ec_deployment.cpu_optimized"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	customHotIc := "testdata/deployment_cpu_optimized_with_custom_hot_ic.tf"
	migrateToLatestHw := "testdata/deployment_cpu_optimized_with_migrate_to_latest_hw.tf"
	region := getRegion()
	customHotIcCfg := fixtureAccDeploymentResourceBasicDefaults(t, customHotIc, randomName, region, cpuOpTemplate)
	migrateToLatestHwCfg, expectedHotIC := fixtureAccDeploymentResourceBasicDefaultsWithExpectedHotIC(t, migrateToLatestHw, randomName, region, cpuOpTemplate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				// Create a Compute Optimized deployment with the default settings.
				Config: customHotIcCfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "deployment_template_id", setDefaultTemplate(region, cpuOpTemplate)),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.instance_configuration_id", "aws.es.datahot.m5d"), // it should contain custom IC
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size", "8g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size_resource", "memory"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.hot.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "kibana.zone_count", "1"),
					resource.TestCheckResourceAttrSet(resName, "kibana.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "kibana.size", "1g"),
					resource.TestCheckResourceAttr(resName, "kibana.size_resource", "memory"),
					resource.TestCheckNoResourceAttr(resName, "apm"),
					resource.TestCheckNoResourceAttr(resName, "enterprise_search"),
				),
			},
			{
				// Create a Compute Optimized deployment with the default settings.
				Config: migrateToLatestHwCfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "deployment_template_id", setDefaultTemplate(region, cpuOpTemplate)),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.instance_configuration_id", expectedHotIC), // it should contain the latest cpu opt IC
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size", "8g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size_resource", "memory"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.hot.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "kibana.zone_count", "1"),
					resource.TestCheckResourceAttrSet(resName, "kibana.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "kibana.size", "1g"),
					resource.TestCheckResourceAttr(resName, "kibana.size_resource", "memory"),
					resource.TestCheckNoResourceAttr(resName, "apm"),
					resource.TestCheckNoResourceAttr(resName, "enterprise_search"),
				),
			},
		},
	})
}

func fixtureAccDeploymentResourceBasicDefaultsWithExpectedHotIC(t *testing.T, fileName, name, region, depTpl string) (string, string) {
	t.Helper()
	requiresAPIConn(t)

	deploymentTpl := setDefaultTemplate(region, depTpl)
	esIC, _, _, err := getInstanceConfigurations(deploymentTpl)
	if err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}

	return fmt.Sprintf(string(b), region, name, region, deploymentTpl), esIC
}
