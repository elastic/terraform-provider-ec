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

func TestAccDeployment_enterpriseSearch(t *testing.T) {
	resName := "ec_deployment.enterprise_search"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	startCfg := "testdata/deployment_enterprise_search_1.tf"
	secondCfg := "testdata/deployment_enterprise_search_2.tf"
	cfg := fixtureAccDeploymentResourceBasicDefaults(t, startCfg, randomName, getRegion(), enterpriseSearchTemplate)
	secondConfigCfg := fixtureAccDeploymentResourceBasicDefaults(t, secondCfg, randomName, getRegion(), enterpriseSearchTemplate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				// Create an Enterprise Search deployment with the default settings.
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size", "4g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_data", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ingest", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_master", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ml", "false"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "kibana.#", "1"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "kibana.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.size", "1g"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "apm.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "1"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "enterprise_search.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.size", "2g"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.size_resource", "memory"),
				),
			},
			{
				// Change the Elasticsearch topology size.
				Config: secondConfigCfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size", "2g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_data", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ingest", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_master", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ml", "false"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "kibana.#", "1"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "kibana.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.size", "1g"),
					resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "apm.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "1"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "enterprise_search.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.size", "2g"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.size_resource", "memory"),
				),
			},
		},
	})
}
