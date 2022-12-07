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
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// This test case takes that on a ccs "ec_deployment".
func TestAccDeployment_ccs(t *testing.T) {
	ccsResName := "ec_deployment.ccs"
	sourceResName := "ec_deployment.source_ccs.0"

	ccsRandomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	sourceRandomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	startCfg := "testdata/deployment_ccs_1.tf"
	secondCfg := "testdata/deployment_ccs_2.tf"
	cfg := fixtureAccDeploymentResourceBasicCcs(t, startCfg,
		ccsRandomName, getRegion(), ccsTemplate,
		sourceRandomName, getRegion(), defaultTemplate,
	)
	secondConfigCfg := fixtureAccDeploymentResourceBasicDefaults(t, secondCfg, ccsRandomName, getRegion(), ccsTemplate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				// Create a CCS deployment with the default settings.
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(

					// CCS Checks
					resource.TestCheckResourceAttrSet(ccsResName, "elasticsearch.hot.instance_configuration_id"),
					// CCS defaults to 1g.
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.hot.size", "1g"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.hot.size_resource", "memory"),

					// Remote cluster settings
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.remote_cluster.#", "3"),
					resource.TestCheckResourceAttrSet(ccsResName, "elasticsearch.remote_cluster.0.deployment_id"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.remote_cluster.0.alias", fmt.Sprint(sourceRandomName, "-0")),
					resource.TestCheckResourceAttrSet(ccsResName, "elasticsearch.remote_cluster.1.deployment_id"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.remote_cluster.1.alias", fmt.Sprint(sourceRandomName, "-1")),
					resource.TestCheckResourceAttrSet(ccsResName, "elasticsearch.remote_cluster.2.deployment_id"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.remote_cluster.2.alias", fmt.Sprint(sourceRandomName, "-2")),

					resource.TestCheckNoResourceAttr(ccsResName, "elasticsearch.hot.node_type_data"),
					resource.TestCheckNoResourceAttr(ccsResName, "elasticsearch.hot.node_type_ingest"),
					resource.TestCheckNoResourceAttr(ccsResName, "elasticsearch.hot.node_type_master"),
					resource.TestCheckNoResourceAttr(ccsResName, "elasticsearch.hot.node_type_ml"),
					resource.TestCheckResourceAttrSet(ccsResName, "elasticsearch.hot.node_roles.#"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.hot.zone_count", "1"),
					resource.TestCheckNoResourceAttr(sourceResName, "kibana"),
					resource.TestCheckNoResourceAttr(sourceResName, "apm"),
					resource.TestCheckNoResourceAttr(sourceResName, "enterprise_search"),
					// Source Checks

					resource.TestCheckResourceAttrSet(sourceResName, "elasticsearch.hot.instance_configuration_id"),
					resource.TestCheckResourceAttr(sourceResName, "elasticsearch.hot.size", "1g"),
					resource.TestCheckResourceAttr(sourceResName, "elasticsearch.hot.size_resource", "memory"),
					resource.TestCheckNoResourceAttr(ccsResName, "elasticsearch.hot.node_type_data"),
					resource.TestCheckNoResourceAttr(ccsResName, "elasticsearch.hot.node_type_ingest"),
					resource.TestCheckNoResourceAttr(ccsResName, "elasticsearch.hot.node_type_master"),
					resource.TestCheckNoResourceAttr(ccsResName, "elasticsearch.hot.node_type_ml"),
					resource.TestCheckResourceAttrSet(sourceResName, "elasticsearch.hot.node_roles.#"),
					resource.TestCheckResourceAttr(sourceResName, "elasticsearch.hot.zone_count", "1"),
					resource.TestCheckNoResourceAttr(sourceResName, "kibana"),
					resource.TestCheckNoResourceAttr(sourceResName, "apm"),
					resource.TestCheckNoResourceAttr(sourceResName, "enterprise_search"),
				),
			},
			{
				// Change the Elasticsearch topology size and node count.
				Config: secondConfigCfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Changes.
					resource.TestCheckResourceAttrSet(ccsResName, "elasticsearch.hot.instance_configuration_id"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.hot.size", "2g"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.hot.size_resource", "memory"),

					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.remote_cluster.#", "0"),

					resource.TestCheckNoResourceAttr(ccsResName, "elasticsearch.hot.node_type_data"),
					resource.TestCheckNoResourceAttr(ccsResName, "elasticsearch.hot.node_type_ingest"),
					resource.TestCheckNoResourceAttr(ccsResName, "elasticsearch.hot.node_type_master"),
					resource.TestCheckNoResourceAttr(ccsResName, "elasticsearch.hot.node_type_ml"),

					resource.TestCheckResourceAttrSet(ccsResName, "elasticsearch.hot.node_roles.#"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.hot.zone_count", "1"),
					resource.TestCheckResourceAttr(ccsResName, "kibana.zone_count", "1"),
					resource.TestCheckResourceAttrSet(ccsResName, "kibana.instance_configuration_id"),
					resource.TestCheckResourceAttr(ccsResName, "kibana.size", "1g"),
					resource.TestCheckResourceAttr(ccsResName, "kibana.size_resource", "memory"),
					resource.TestCheckNoResourceAttr(ccsResName, "apm"),
					resource.TestCheckNoResourceAttr(ccsResName, "enterprise_search"),
				),
			},
		},
	})
}

func fixtureAccDeploymentResourceBasicCcs(t *testing.T, fileName, name, region, ccsTplName, sourceName, sourceRegion, sourceTplName string) string {
	t.Helper()

	ccsTpl := setDefaultTemplate(region, ccsTplName)
	sourceTpl := setDefaultTemplate(region, sourceTplName)

	b, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		region, name, region, ccsTpl,
		sourceName, sourceRegion, sourceTpl,
	)
}
