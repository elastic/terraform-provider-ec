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

func TestAccDeploymentTrafficFilterAssociation_basic(t *testing.T) {
	resName := "ec_deployment_traffic_filter.tf_assoc"
	resNameSecond := "ec_deployment_traffic_filter.tf_assoc_second"
	resAssocName := "ec_deployment_traffic_filter_association.tf_assoc"
	randomName := acctest.RandomWithPrefix(prefix)
	randomNameSecond := acctest.RandomWithPrefix(prefix)
	startCfg := "testdata/deployment_traffic_filter_association_basic.tf"
	updateCfg := "testdata/deployment_traffic_filter_association_basic_update.tf"
	cfg := testAccDeploymentTrafficFilterResourceAssociationBasic(t, startCfg, randomName, region, deploymentVersion)
	updateConfigCfg := testAccDeploymentTrafficFilterResourceAssociationBasic(t, updateCfg, randomNameSecond, region, deploymentVersion)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccDeploymentTrafficFilterDestroy,
		Steps: []resource.TestStep{
			{
				// Expects a non-empty plan since "ec_deployment.traffic_filter"
				// will have changes due to the traffic filter association.
				ExpectNonEmptyPlan: true,
				Config:             cfg,
				Check: checkBasicDeploymentTrafficFilterAssociationResource(
					resName, resAssocName, randomName,
					resource.TestCheckResourceAttr(resName, "include_by_default", "false"),
					resource.TestCheckResourceAttr(resName, "type", "ip"),
					resource.TestCheckResourceAttr(resName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resName, "rule.0.source", "0.0.0.0/0"),
				),
			},
			{
				// Expects a non-empty plan since "ec_deployment.traffic_filter"
				// will have changes due to the traffic filter association.
				ExpectNonEmptyPlan: true,
				Config:             updateConfigCfg,
				Check: checkBasicDeploymentTrafficFilterAssociationResource(
					resNameSecond, resAssocName, randomNameSecond,
					resource.TestCheckResourceAttr(resNameSecond, "include_by_default", "false"),
					resource.TestCheckResourceAttr(resNameSecond, "type", "ip"),
					resource.TestCheckResourceAttr(resNameSecond, "rule.#", "1"),
					resource.TestCheckResourceAttr(resNameSecond, "rule.0.source", "0.0.0.0/0"),
				),
			},
		},
	})
}

func testAccDeploymentTrafficFilterResourceAssociationBasic(t *testing.T, fileName, name, region, version string) string {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		name, region, version, name, region,
	)
}

func checkBasicDeploymentTrafficFilterAssociationResource(resName, assocName, randomDeploymentName string, checks ...resource.TestCheckFunc) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(append([]resource.TestCheckFunc{
		testAccCheckDeploymentTrafficFilterExists(resName),
		resource.TestCheckResourceAttrSet(assocName, "deployment_id"),
		resource.TestCheckResourceAttrSet(assocName, "traffic_filter_id"),
		resource.TestCheckResourceAttr(resName, "name", randomDeploymentName),
		resource.TestCheckResourceAttr(resName, "region", region)}, checks...)...,
	)
}
