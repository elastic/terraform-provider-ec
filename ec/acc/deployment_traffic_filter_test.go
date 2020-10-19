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

func TestAccDeploymentTrafficFilter_basic(t *testing.T) {
	resName := "ec_deployment_traffic_filter.basic"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	startCfg := "testdata/deployment_traffic_filter_basic.tf"
	updateCfg := "testdata/deployment_traffic_filter_basic_update.tf"
	cfg := fixtureAccDeploymentTrafficFilterResourceBasic(t, startCfg, randomName, getRegion())
	updateConfigCfg := fixtureAccDeploymentTrafficFilterResourceBasic(t, updateCfg, randomName, getRegion())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccDeploymentTrafficFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: checkBasicDeploymentTrafficFilterResource(resName, randomName,
					resource.TestCheckResourceAttr(resName, "include_by_default", "false"),
					resource.TestCheckResourceAttr(resName, "type", "ip"),
					resource.TestCheckResourceAttr(resName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resName, "rule.0.source", "0.0.0.0/0"),
				),
			},
			// Ensure that no diff is generated.
			{Config: cfg, PlanOnly: true},
			{
				Config: updateConfigCfg,
				Check: checkBasicDeploymentTrafficFilterResource(resName, randomName,
					resource.TestCheckResourceAttr(resName, "include_by_default", "false"),
					resource.TestCheckResourceAttr(resName, "type", "ip"),
					resource.TestCheckResourceAttr(resName, "rule.#", "2"),
					resource.TestCheckResourceAttr(resName, "rule.0.source", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resName, "rule.1.source", "1.1.1.0/24"),
				),
			},
			// Ensure that no diff is generated.
			{Config: updateConfigCfg, PlanOnly: true},
			{
				ResourceName:            resName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
		},
	})
}

func fixtureAccDeploymentTrafficFilterResourceBasic(t *testing.T, fileName, name, region string) string {
	t.Helper()
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		name, region,
	)
}

func checkBasicDeploymentTrafficFilterResource(resName, randomDeploymentName string, checks ...resource.TestCheckFunc) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(append([]resource.TestCheckFunc{
		testAccCheckDeploymentTrafficFilterExists(resName),
		resource.TestCheckResourceAttr(resName, "name", randomDeploymentName),
		resource.TestCheckResourceAttr(resName, "region", getRegion())}, checks...)...,
	)
}
