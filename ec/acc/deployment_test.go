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
	cfg := testAccDeploymentResourceAppsearch(t, randomName, region, deploymentVersionAppsearch)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(resName),
					resource.TestCheckResourceAttr(resName, "name", randomName),
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
				),
			},
			// Ensure that no diff is generated.
			{
				Config:   cfg,
				PlanOnly: true,
			},
			// TODO: Import case when import is ready.
		},
	})
}

func TestAccDeployment_enterpriseSearch(t *testing.T) {
	resName := "ec_deployment.enterprise_search"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	cfg := testAccDeploymentResourceEnterpriseSearch(t, randomName, region, deploymentVersion)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(resName),
					resource.TestCheckResourceAttr(resName, "name", randomName),
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
				),
			},
			// Ensure that no diff is generated.
			{
				Config:   cfg,
				PlanOnly: true,
			},
			// TODO: Import case when import is ready.
		},
	})
}

func testAccDeploymentResourceAppsearch(t *testing.T, name, region, version string) string {
	b, err := ioutil.ReadFile("testdata/deployment_appsearch.tf")
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		name, region, version,
	)
}

func testAccDeploymentResourceEnterpriseSearch(t *testing.T, name, region, version string) string {
	b, err := ioutil.ReadFile("testdata/deployment_enterprise_search.tf")
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		name, region, version,
	)
}
