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

func TestAccDeployment_template_migration(t *testing.T) {
	resName := "ec_deployment.compute_optimized"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	basicCfg := "testdata/deployment_compute_optimized_1.tf"
	region := getRegion()
	cfg := fixtureAccDeploymentResourceBasicDefaults(t, basicCfg, randomName, region, computeOpTemplate)
	secondConfigCfg := fixtureAccDeploymentResourceBasicDefaults(t, basicCfg, randomName, region, memoryOpTemplate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				// Create a Compute Optimized deployment with the default settings.
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "deployment_template_id", setDefaultTemplate(region, computeOpTemplate)),
					resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size", "8g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_data", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ingest", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_master", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ml", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.id", "hot_content"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "kibana.#", "1"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "kibana.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.size", "1g"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "apm.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "0"),
				),
			},
			{
				// Change the deployment to memory optimized
				Config: secondConfigCfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "deployment_template_id", setDefaultTemplate(region, memoryOpTemplate)),
					resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size", "8g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_data", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ingest", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_master", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ml", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.id", "hot_content"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "kibana.#", "1"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "kibana.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.size", "1g"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "apm.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "0"),
				),
			},
		},
	})
}
