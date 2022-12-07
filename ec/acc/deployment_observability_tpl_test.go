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

func TestAccDeployment_observabilityTpl(t *testing.T) {
	resName := "ec_deployment.observability_tpl"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	startCfg := "testdata/deployment_observability_tpl_1.tf"
	secondCfg := "testdata/deployment_observability_tpl_2.tf"
	cfg := fixtureAccDeploymentResourceBasicDefaults(t, startCfg, randomName, getRegion(), observabilityTemplate)
	secondConfigCfg := fixtureAccDeploymentResourceBasicDefaults(t, secondCfg, randomName, getRegion(), observabilityTemplate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				// Create an Observability deployment with the default settings.
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.hot.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size", "8g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size_resource", "memory"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.hot.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "kibana.zone_count", "1"),
					resource.TestCheckResourceAttrSet(resName, "kibana.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "kibana.size", "1g"),
					resource.TestCheckResourceAttr(resName, "kibana.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "apm.zone_count", "1"),
					resource.TestCheckResourceAttrSet(resName, "apm.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "apm.size", "1g"),
					resource.TestCheckResourceAttr(resName, "apm.size_resource", "memory"),
					resource.TestCheckNoResourceAttr(resName, "enterprise_search"),
				),
			},
			{
				// Change the Elasticsearch topology size.
				Config: secondConfigCfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.hot.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size", "2g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "kibana.zone_count", "1"),
					resource.TestCheckResourceAttrSet(resName, "kibana.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "kibana.size", "1g"),
					resource.TestCheckResourceAttr(resName, "kibana.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "apm.zone_count", "1"),
					resource.TestCheckResourceAttrSet(resName, "apm.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "apm.size", "1g"),
					resource.TestCheckResourceAttr(resName, "apm.size_resource", "memory"),
					resource.TestCheckNoResourceAttr(resName, "enterprise_search"),
				),
			},
		},
	})
}
