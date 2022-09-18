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
					resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.#", "3"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.1.instance_configuration_id"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.id", "coordinating"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size_resource", "memory"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.id", "hot_content"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.1.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.size_resource", "memory"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.2.id", "warm"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.2.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.2.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.2.size", "2g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.2.size_resource", "memory"),

					resource.TestCheckResourceAttr(resName, "kibana.#", "0"),
					resource.TestCheckResourceAttr(resName, "apm.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "0"),
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
					resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.#", "4"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.1.instance_configuration_id"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.2.instance_configuration_id"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.3.instance_configuration_id"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.id", "cold"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size", "2g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.zone_count", "1"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.id", "hot_content"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.size_resource", "memory"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.1.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.zone_count", "3"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.2.id", "master"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.2.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.2.size_resource", "memory"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.2.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.2.zone_count", "3"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.3.id", "warm"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.3.size", "2g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.3.size_resource", "memory"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.3.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.3.zone_count", "2"),

					resource.TestCheckResourceAttr(resName, "kibana.#", "0"),
					resource.TestCheckResourceAttr(resName, "apm.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "0"),
				),
			},
		},
	})
}
