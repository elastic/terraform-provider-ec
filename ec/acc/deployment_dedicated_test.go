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

func TestAccDeployment_dedicated_coordinating(t *testing.T) {
	resName := "ec_deployment.dedicated_coordinating"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	startCfg := "testdata/deployment_dedicated_coordinating.tf"
	cfg := fixtureAccDeploymentResourceBasicDefaults(t, startCfg, randomName, getRegion(), hotWarmTemplate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				// Create a deployment with dedicated coordinating.
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(

					resource.TestCheckResourceAttrSet(resName, "elasticsearch.coordinating.instance_configuration_id"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.coordinating.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.coordinating.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.coordinating.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.coordinating.size_resource", "memory"),

					resource.TestCheckResourceAttrSet(resName, "elasticsearch.hot.instance_configuration_id"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.hot.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size_resource", "memory"),

					resource.TestCheckResourceAttrSet(resName, "elasticsearch.warm.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.size", "2g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.size_resource", "memory"),

					resource.TestCheckNoResourceAttr(resName, "kibana"),
					resource.TestCheckNoResourceAttr(resName, "apm"),
					resource.TestCheckNoResourceAttr(resName, "enterprise_search"),
				),
			},
		},
	})
}

func TestAccDeployment_dedicated_master(t *testing.T) {
	resName := "ec_deployment.dedicated_master"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	startCfg := "testdata/deployment_dedicated_master.tf"
	cfg := fixtureAccDeploymentResourceBasicDefaults(t, startCfg, randomName, getRegion(), hotWarmTemplate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				// Create a deployment with dedicated master nodes.
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.cold.instance_configuration_id"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.hot.instance_configuration_id"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.master.instance_configuration_id"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.warm.instance_configuration_id"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.cold.size", "2g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.cold.size_resource", "memory"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.cold.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.cold.zone_count", "1"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size_resource", "memory"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.hot.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.zone_count", "3"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.master.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.master.size_resource", "memory"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.master.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.master.zone_count", "3"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.size", "2g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.size_resource", "memory"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.warm.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.zone_count", "2"),

					resource.TestCheckNoResourceAttr(resName, "kibana"),
					resource.TestCheckNoResourceAttr(resName, "apm"),
					resource.TestCheckNoResourceAttr(resName, "enterprise_search"),
				),
			},
		},
	})
}
