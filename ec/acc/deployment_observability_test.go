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

// +build acceptance

package acc

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// This test case takes that on a hot/warm "ec_deployment", a select number of
// topology settings can be changed without affecting the underlying Deployment
// Template.
func TestAccDeployment_observability(t *testing.T) {
	resName := "ec_deployment.observability"
	secondResName := "ec_deployment.basic"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	startCfg := "testdata/deployment_observability_1.tf"
	updateCfg := "testdata/deployment_observability_2.tf"
	secondUpdateCfg := "testdata/deployment_observability_3.tf"
	cfg := fixtureAccDeploymentResourceBasicObs(t, startCfg, randomName, getRegion(), defaultTemplate)
	secondCfg := fixtureAccDeploymentResourceBasicObs(t, updateCfg, randomName, getRegion(), defaultTemplate)
	thirdCfg := fixtureAccDeploymentResourceBasicObs(t, secondUpdateCfg, randomName, getRegion(), defaultTemplate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config:             cfg,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resName, "observability.0.deployment_id", secondResName, "id"),
					resource.TestCheckResourceAttr(resName, "observability.0.metrics", "true"),
					resource.TestCheckResourceAttr(resName, "observability.0.logs", "true"),
				),
			},
			{
				Config:             secondCfg,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resName, "observability.0.deployment_id", secondResName, "id"),
					resource.TestCheckResourceAttr(resName, "observability.0.metrics", "false"),
					resource.TestCheckResourceAttr(resName, "observability.0.logs", "true"),
				),
			},
			{
				Config:             thirdCfg,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resName, "observability.0.deployment_id", secondResName, "id"),
					resource.TestCheckResourceAttr(resName, "observability.0.metrics", "true"),
					resource.TestCheckResourceAttr(resName, "observability.0.logs", "false"),
				),
			},
		},
	})
}

func fixtureAccDeploymentResourceBasicObs(t *testing.T, fileName, name, region, depTpl string) string {
	deploymentTpl := setDefaultTemplate(region, depTpl)

	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		region, name, region, deploymentTpl, name, region, deploymentTpl,
	)
}
