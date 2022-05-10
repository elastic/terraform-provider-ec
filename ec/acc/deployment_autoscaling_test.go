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

func TestAccDeployment_autoscaling(t *testing.T) {
	resName := "ec_deployment.autoscaling"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	startCfg := "testdata/deployment_autoscaling_1.tf"
	disableAutoscale := "testdata/deployment_autoscaling_2.tf"

	cfgF := func(cfg string) string {
		return fixtureAccDeploymentResourceBasic(
			t, cfg, randomName, getRegion(), defaultTemplate,
		)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfgF(startCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.autoscale", "true"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.#", "5"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.instance_configuration_id"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.id", "cold"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size", "0g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.autoscaling.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.autoscaling.0.max_size", "58g"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.id", "frozen"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.size", "0g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.autoscaling.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.autoscaling.0.max_size", "120g"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.2.id", "hot_content"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.2.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.2.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.2.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.2.autoscaling.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.2.autoscaling.0.max_size", "8g"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.3.id", "ml"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.3.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.3.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.3.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.3.autoscaling.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.3.autoscaling.0.max_size", "4g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.3.autoscaling.0.min_size", "1g"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.4.id", "warm"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.4.size", "2g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.4.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.4.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.4.autoscaling.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.4.autoscaling.0.max_size", "15g"),

					resource.TestCheckResourceAttr(resName, "kibana.#", "0"),
					resource.TestCheckResourceAttr(resName, "apm.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "0"),
				),
			},
			// also disables ML
			{
				Config: cfgF(disableAutoscale),
				// When disabling a tier the plan will be non empty on refresh
				// since the topology block is present with size = "0g".
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.autoscale", "false"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.#", "2"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.instance_configuration_id"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.id", "hot_content"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.autoscaling.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.autoscaling.0.max_size", "8g"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.id", "warm"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.size", "2g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.autoscaling.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.1.autoscaling.0.max_size", "15g"),

					resource.TestCheckResourceAttr(resName, "kibana.#", "0"),
					resource.TestCheckResourceAttr(resName, "apm.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "0"),
				),
			},
		},
	})
}
