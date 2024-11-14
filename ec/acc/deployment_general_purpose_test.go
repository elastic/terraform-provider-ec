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

// This test case takes that on a general purpose "ec_deployment", a select number of
// topology settings can be changed without affecting the underlying Deployment Template.
func TestAccDeployment_general_purpose(t *testing.T) {
	resName := "ec_deployment.general_purpose"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	startCfg := "testdata/deployment_general_purpose_1.tf"
	secondCfg := "testdata/deployment_general_purpose_2.tf"
	cfg := fixtureAccDeploymentResourceBasicDefaults(t, startCfg, randomName, getRegion(), generalPurposeTemplate)
	secondConfigCfg := fixtureAccDeploymentResourceBasic(t, secondCfg, randomName, getRegion(), generalPurposeTemplate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				// Create a general purpose deployment with the default settings.
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.hot.instance_configuration_id"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.warm.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size", "8g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size_resource", "memory"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.hot.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.zone_count", "2"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.warm.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.zone_count", "2"),

					resource.TestCheckResourceAttrSet(resName, "kibana.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "kibana.size", "1g"),
					resource.TestCheckResourceAttr(resName, "kibana.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "kibana.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "apm.size", "2g"),
					resource.TestCheckResourceAttrSet(resName, "apm.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "apm.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "apm.zone_count", "1"),
					resource.TestCheckResourceAttrSet(resName, "enterprise_search.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.size", "2g"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.zone_count", "1"),
					resource.TestCheckResourceAttrSet(resName, "integrations_server.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "integrations_server.size", "2g"),
					resource.TestCheckResourceAttr(resName, "integrations_server.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "integrations_server.zone_count", "1"),
				),
			},
			{
				// Change the Elasticsearch toplogy size and node count.
				Config: secondConfigCfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Changes.
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.size", "2g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.zone_count", "1"),

					resource.TestCheckResourceAttrSet(resName, "elasticsearch.hot.instance_configuration_id"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.warm.instance_configuration_id"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.hot.node_roles.#"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.warm.node_roles.#"),

					resource.TestCheckResourceAttrSet(resName, "kibana.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "kibana.size", "1g"),
					resource.TestCheckResourceAttr(resName, "kibana.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "kibana.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "apm.size", "2g"),
					resource.TestCheckResourceAttrSet(resName, "apm.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "apm.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "apm.zone_count", "1"),
					resource.TestCheckResourceAttrSet(resName, "enterprise_search.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.size", "2g"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.zone_count", "1"),
					resource.TestCheckResourceAttrSet(resName, "integrations_server.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "integrations_server.size", "2g"),
					resource.TestCheckResourceAttr(resName, "integrations_server.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "integrations_server.zone_count", "1"),
				),
			},
		},
	})
}

func fixtureAccDeploymentResourceBasic(t *testing.T, fileName, name, region, depTpl string) string {
	t.Helper()
	requiresAPIConn(t)

	b, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		region, name, region, setDefaultTemplate(region, depTpl),
	)
}
