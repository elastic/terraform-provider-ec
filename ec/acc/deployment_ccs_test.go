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
	"io/ioutil"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// This test case takes that on a ccs "ec_deployment".
func TestAccDeployment_ccs(t *testing.T) {
	ccsResName := "ec_deployment.ccs"
	sourceResName := "ec_deployment.source_ccs"

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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				// Create a CCS deployment with the default settings.
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					// CCS Checks
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(ccsResName, "elasticsearch.0.topology.0.instance_configuration_id"),
					// CCS defaults to 1g.
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.0.topology.0.size", "1g"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.0.topology.0.size_resource", "memory"),
					// Remote cluster settings
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.0.remote_cluster.#", "1"),
					resource.TestCheckResourceAttrSet(ccsResName, "elasticsearch.0.remote_cluster.0.deployment_id"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.0.remote_cluster.0.alias", "my_source_ccs"),

					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.0.topology.0.node_type_data", "true"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.0.topology.0.node_type_ingest", "true"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.0.topology.0.node_type_master", "true"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.0.topology.0.node_type_ml", "false"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.0.topology.0.zone_count", "1"),
					resource.TestCheckResourceAttr(ccsResName, "kibana.#", "0"),
					resource.TestCheckResourceAttr(ccsResName, "apm.#", "0"),
					resource.TestCheckResourceAttr(ccsResName, "enterprise_search.#", "0"),

					// Source Checks

					resource.TestCheckResourceAttr(sourceResName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(sourceResName, "elasticsearch.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(sourceResName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(sourceResName, "elasticsearch.0.topology.0.size", "1g"),
					resource.TestCheckResourceAttr(sourceResName, "elasticsearch.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(sourceResName, "elasticsearch.0.topology.0.node_type_data", "true"),
					resource.TestCheckResourceAttr(sourceResName, "elasticsearch.0.topology.0.node_type_ingest", "true"),
					resource.TestCheckResourceAttr(sourceResName, "elasticsearch.0.topology.0.node_type_master", "true"),
					resource.TestCheckResourceAttr(sourceResName, "elasticsearch.0.topology.0.node_type_ml", "false"),
					resource.TestCheckResourceAttr(sourceResName, "elasticsearch.0.topology.0.zone_count", "1"),
					resource.TestCheckResourceAttr(sourceResName, "kibana.#", "0"),
					resource.TestCheckResourceAttr(sourceResName, "apm.#", "0"),
					resource.TestCheckResourceAttr(sourceResName, "enterprise_search.#", "0"),
				),
			},
			{
				// Change the Elasticsearch topology size and node count.
				Config: secondConfigCfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Changes.
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(ccsResName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.0.topology.0.size", "2g"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.0.topology.0.size_resource", "memory"),

					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.0.remote_cluster.#", "0"),

					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.0.topology.0.node_type_data", "true"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.0.topology.0.node_type_ingest", "true"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.0.topology.0.node_type_master", "true"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.0.topology.0.node_type_ml", "false"),
					resource.TestCheckResourceAttr(ccsResName, "elasticsearch.0.topology.0.zone_count", "1"),
					resource.TestCheckResourceAttr(ccsResName, "kibana.#", "1"),
					resource.TestCheckResourceAttr(ccsResName, "kibana.0.topology.0.zone_count", "1"),
					resource.TestCheckResourceAttr(ccsResName, "kibana.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(ccsResName, "kibana.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(ccsResName, "kibana.0.topology.0.size", "1g"),
					resource.TestCheckResourceAttr(ccsResName, "kibana.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(ccsResName, "apm.#", "0"),
					resource.TestCheckResourceAttr(ccsResName, "enterprise_search.#", "0"),
				),
			},
		},
	})
}

func fixtureAccDeploymentResourceBasicCcs(t *testing.T, fileName, name, region, ccsTplName, sourceName, sourceRegion, sourceTplName string) string {
	t.Helper()

	ccsTpl := setDefaultTemplate(region, ccsTplName)
	sourceTpl := setDefaultTemplate(region, sourceTplName)

	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		region, name, region, ccsTpl,
		sourceName, sourceRegion, sourceTpl,
	)
}
