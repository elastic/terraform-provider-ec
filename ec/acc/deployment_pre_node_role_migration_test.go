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

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeployment_pre_node_roles(t *testing.T) {
	resName := "ec_deployment.pre_nr"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	startCfg := "testdata/deployment_pre_node_roles_migration_1.tf"
	upgradeVersionCfg := "testdata/deployment_pre_node_roles_migration_2.tf"
	addWarmTopologyCfg := "testdata/deployment_pre_node_roles_migration_3.tf"

	cfgF := func(cfg string) string {
		return fixtureAccDeploymentResourceBasic(
			t, cfg, randomName, getRegion(), defaultTemplate,
		)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfgF(startCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.hot.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.node_type_data", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.node_type_ingest", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.node_type_master", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.node_type_ml", "false"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.node_roles.#", "0"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.zone_count", "1"),

					resource.TestCheckNoResourceAttr(resName, "kibana"),
					resource.TestCheckNoResourceAttr(resName, "apm"),
					resource.TestCheckNoResourceAttr(resName, "enterprise_search"),
				),
			},
			{
				Config: cfgF(upgradeVersionCfg),
				// Expect a non-empty plan here. We explicitly avoid migrating node_roles when the version changes
				// however will the migrate the deployment to node_roles on the next TF application.
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.hot.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.node_type_data", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.node_type_ingest", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.node_type_master", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.node_type_ml", "false"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.node_roles.#", "0"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.zone_count", "1"),

					resource.TestCheckNoResourceAttr(resName, "kibana"),
					resource.TestCheckNoResourceAttr(resName, "apm"),
					resource.TestCheckNoResourceAttr(resName, "enterprise_search"),
				),
			},
			{
				Config: cfgF(addWarmTopologyCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Hot
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.hot.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size_resource", "memory"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.hot.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.zone_count", "1"),

					resource.TestCheckNoResourceAttr(resName, "elastic.hot.node_type_data"),
					resource.TestCheckNoResourceAttr(resName, "elastic.hot.node_type_ingest"),
					resource.TestCheckNoResourceAttr(resName, "elastic.hot.node_type_master"),
					resource.TestCheckNoResourceAttr(resName, "elastic.hot.node_type_ml"),

					// Warm
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.warm.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.size", "2g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.size_resource", "memory"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.warm.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.zone_count", "1"),

					resource.TestCheckNoResourceAttr(resName, "elastic.warm.node_type_data"),
					resource.TestCheckNoResourceAttr(resName, "elastic.warm.node_type_ingest"),
					resource.TestCheckNoResourceAttr(resName, "elastic.warm.node_type_master"),
					resource.TestCheckNoResourceAttr(resName, "elastic.warm.node_type_ml"),

					resource.TestCheckNoResourceAttr(resName, "kibana"),
					resource.TestCheckNoResourceAttr(resName, "apm"),
					resource.TestCheckNoResourceAttr(resName, "enterprise_search"),
				),
			},
		},
	})
}
