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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfgF(startCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "elasticsearch.autoscale", "true"),

					resource.TestCheckResourceAttrSet(resName, "elasticsearch.topology.cold.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.cold.size", "0g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.cold.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.cold.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.cold.autoscaling.max_size", "58g"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.frozen.size", "0g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.frozen.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.frozen.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.frozen.autoscaling.max_size", "120g"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.hot_content.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.hot_content.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.hot_content.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.hot_content.autoscaling.max_size", "8g"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.ml.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.ml.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.ml.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.ml.autoscaling.max_size", "4g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.ml.autoscaling.min_size", "1g"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.warm.size", "2g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.warm.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.warm.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.warm.autoscaling.max_size", "15g"),

					resource.TestCheckNoResourceAttr(resName, "kibana"),
					resource.TestCheckNoResourceAttr(resName, "apm"),
					resource.TestCheckNoResourceAttr(resName, "enterprise_search"),
				),
			},
			// also disables ML
			{
				Config: cfgF(disableAutoscale),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "elasticsearch.autoscale", "false"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.topology.hot_content.instance_configuration_id"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.hot_content.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.hot_content.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.hot_content.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.hot_content.autoscaling.max_size", "8g"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.warm.size", "2g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.warm.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.warm.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.topology.warm.autoscaling.max_size", "15g"),

					resource.TestCheckNoResourceAttr(resName, "kibana"),
					resource.TestCheckNoResourceAttr(resName, "apm"),
					resource.TestCheckNoResourceAttr(resName, "enterprise_search"),
				),
			},
		},
	})
}
