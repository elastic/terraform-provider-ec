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

					resource.TestCheckResourceAttrSet(resName, "elasticsearch.cold.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.cold.size", "0g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.cold.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.cold.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.cold.autoscaling.max_size", "58g"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.frozen.size", "0g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.frozen.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.frozen.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.frozen.autoscaling.max_size", "120g"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.autoscaling.max_size", "8g"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.ml.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.ml.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.ml.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.ml.autoscaling.max_size", "4g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.ml.autoscaling.min_size", "1g"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.size", "2g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.autoscaling.max_size", "15g"),

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
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.hot.instance_configuration_id"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.autoscaling.max_size", "8g"),

					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.size", "2g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.zone_count", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.warm.autoscaling.max_size", "15g"),

					resource.TestCheckNoResourceAttr(resName, "kibana"),
					resource.TestCheckNoResourceAttr(resName, "apm"),
					resource.TestCheckNoResourceAttr(resName, "enterprise_search"),
				),
			},
		},
	})
}
