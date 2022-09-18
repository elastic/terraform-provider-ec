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

// This test case takes ensures that several features of the "ec_deployment"
// resource are asserted:
// * Create a deployment with tags.
// * Remove a tag.
// * Remove the tag block.
func TestAccDeployment_basic_tags(t *testing.T) {
	resName := "ec_deployment.tags"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	startCfg := "testdata/deployment_basic_tags_1.tf"
	secondCfg := "testdata/deployment_basic_tags_2.tf"
	thirdCfg := "testdata/deployment_basic_tags_3.tf"
	fourthCfg := "testdata/deployment_basic_tags_4.tf"
	cfg := fixtureAccDeploymentResourceBasicDefaults(t, startCfg, randomName, getRegion(), defaultTemplate)
	secondConfigCfg := fixtureAccDeploymentResourceBasicDefaults(t, secondCfg, randomName, getRegion(), defaultTemplate)
	thirdConfigCfg := fixtureAccDeploymentResourceBasicDefaults(t, thirdCfg, randomName, getRegion(), defaultTemplate)
	fourthConfigCfg := fixtureAccDeploymentResourceBasicDefaults(t, fourthCfg, randomName, getRegion(), defaultTemplate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				// Create a deployment with tags.
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size", "2g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_data", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ingest", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_master", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ml", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.id", "hot_content"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "kibana.#", "0"),
					resource.TestCheckResourceAttr(resName, "apm.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "0"),

					// Tags
					resource.TestCheckResourceAttr(resName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resName, "tags.owner", "elastic"),
					resource.TestCheckResourceAttr(resName, "tags.cost-center", "rnd"),
				),
			},
			{
				// Remove a tag.
				Config: secondConfigCfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size", "2g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_data", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ingest", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_master", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ml", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.id", "hot_content"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "kibana.#", "0"),
					resource.TestCheckResourceAttr(resName, "apm.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "0"),

					// Tags
					resource.TestCheckResourceAttr(resName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resName, "tags.owner", "elastic"),
				),
			},
			{
				// Remove the tags block.
				Config: thirdConfigCfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size", "2g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_data", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ingest", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_master", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ml", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.id", "hot_content"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "kibana.#", "0"),
					resource.TestCheckResourceAttr(resName, "apm.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "0"),

					// Tags
					resource.TestCheckResourceAttr(resName, "tags.%", "0"),
				),
			},
			{
				// Add the tags block with a single tag.
				Config: fourthConfigCfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size", "2g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_data", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ingest", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_master", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ml", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.id", "hot_content"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "kibana.#", "0"),
					resource.TestCheckResourceAttr(resName, "apm.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "0"),

					// Tags
					resource.TestCheckResourceAttr(resName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resName, "tags.new", "tag"),
				),
			},
		},
	})
}
