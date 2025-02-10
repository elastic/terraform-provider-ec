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

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// This test case takes that on a cross cluster search "ec_deployment".
func TestAccDeployment_ccs(t *testing.T) {
	generalPurposeResName := "ec_deployment.general_purpose"
	sourceResName := "ec_deployment.source_storage_optimized.0"

	generalPurposeRandomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	sourceRandomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	startCfg := "testdata/deployment_ccs_1.tf"
	secondCfg := "testdata/deployment_ccs_2.tf"
	cfg := fixtureAccDeploymentResourceBasicCrossClusterSearch(t, startCfg,
		generalPurposeRandomName, getRegion(), generalPurposeTemplate,
		sourceRandomName, getRegion(), defaultTemplate,
	)
	secondConfigCfg := fixtureAccDeploymentResourceBasicDefaults(t, secondCfg, generalPurposeRandomName, getRegion(), generalPurposeTemplate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				// Create a general purpose deployment with a cross cluster search configuration
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(

					// general purpose template checks
					resource.TestCheckResourceAttrSet(generalPurposeResName, "elasticsearch.hot.instance_configuration_id"),
					// general purpose template defaults to 8g.
					resource.TestCheckResourceAttr(generalPurposeResName, "elasticsearch.hot.size", "8g"),
					resource.TestCheckResourceAttr(generalPurposeResName, "elasticsearch.hot.size_resource", "memory"),

					// Remote cluster settings
					resource.TestCheckResourceAttr(generalPurposeResName, "elasticsearch.remote_cluster.#", "3"),
					resource.TestCheckResourceAttrSet(generalPurposeResName, "elasticsearch.remote_cluster.0.deployment_id"),
					resource.TestCheckResourceAttr(generalPurposeResName, "elasticsearch.remote_cluster.0.alias", fmt.Sprint(sourceRandomName, "-0")),
					resource.TestCheckResourceAttrSet(generalPurposeResName, "elasticsearch.remote_cluster.1.deployment_id"),
					resource.TestCheckResourceAttr(generalPurposeResName, "elasticsearch.remote_cluster.1.alias", fmt.Sprint(sourceRandomName, "-1")),
					resource.TestCheckResourceAttrSet(generalPurposeResName, "elasticsearch.remote_cluster.2.deployment_id"),
					resource.TestCheckResourceAttr(generalPurposeResName, "elasticsearch.remote_cluster.2.alias", fmt.Sprint(sourceRandomName, "-2")),
					resource.TestCheckResourceAttr(generalPurposeResName, "elasticsearch.hot.zone_count", "2"),

					resource.TestCheckResourceAttrSet(generalPurposeResName, "elasticsearch.hot.node_roles.#"),
					resource.TestCheckNoResourceAttr(generalPurposeResName, "elasticsearch.hot.node_type_data"),
					resource.TestCheckNoResourceAttr(generalPurposeResName, "elasticsearch.hot.node_type_ingest"),
					resource.TestCheckNoResourceAttr(generalPurposeResName, "elasticsearch.hot.node_type_master"),
					resource.TestCheckNoResourceAttr(generalPurposeResName, "elasticsearch.hot.node_type_ml"),

					resource.TestCheckNoResourceAttr(generalPurposeResName, "kibana"),
					resource.TestCheckNoResourceAttr(generalPurposeResName, "apm"),
					resource.TestCheckNoResourceAttr(generalPurposeResName, "enterprise_search"),

					// Source Checks
					resource.TestCheckResourceAttrSet(sourceResName, "elasticsearch.hot.instance_configuration_id"),
					resource.TestCheckResourceAttr(sourceResName, "elasticsearch.hot.size", "1g"),
					resource.TestCheckResourceAttr(sourceResName, "elasticsearch.hot.size_resource", "memory"),
					resource.TestCheckResourceAttrSet(sourceResName, "elasticsearch.hot.node_roles.#"),
					resource.TestCheckResourceAttr(sourceResName, "elasticsearch.hot.zone_count", "1"),
					resource.TestCheckNoResourceAttr(sourceResName, "elasticsearch.hot.node_type_data"),
					resource.TestCheckNoResourceAttr(sourceResName, "elasticsearch.hot.node_type_ingest"),
					resource.TestCheckNoResourceAttr(sourceResName, "elasticsearch.hot.node_type_master"),
					resource.TestCheckNoResourceAttr(sourceResName, "elasticsearch.hot.node_type_ml"),
					resource.TestCheckResourceAttrSet(sourceResName, "elasticsearch.hot.node_roles.#"),

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
					resource.TestCheckResourceAttrSet(generalPurposeResName, "elasticsearch.hot.instance_configuration_id"),
					resource.TestCheckResourceAttr(generalPurposeResName, "elasticsearch.hot.size", "4g"),
					resource.TestCheckResourceAttr(generalPurposeResName, "elasticsearch.hot.size_resource", "memory"),

					resource.TestCheckResourceAttr(generalPurposeResName, "elasticsearch.remote_cluster.#", "0"),

					resource.TestCheckNoResourceAttr(generalPurposeResName, "elasticsearch.hot.node_type_data"),
					resource.TestCheckNoResourceAttr(generalPurposeResName, "elasticsearch.hot.node_type_ingest"),
					resource.TestCheckNoResourceAttr(generalPurposeResName, "elasticsearch.hot.node_type_master"),
					resource.TestCheckNoResourceAttr(generalPurposeResName, "elasticsearch.hot.node_type_ml"),

					resource.TestCheckResourceAttrSet(generalPurposeResName, "elasticsearch.hot.node_roles.#"),
					resource.TestCheckResourceAttr(generalPurposeResName, "elasticsearch.hot.zone_count", "2"),

					// TODO: uncomment once bug for kibana instance_configuration_version is fixed
					// resource.TestCheckResourceAttrSet(generalPurposeResName, "kibana.instance_configuration_id"),
					// resource.TestCheckResourceAttrSet(generalPurposeResName, "kibana.instance_configuration_version"),
					// resource.TestCheckResourceAttr(generalPurposeResName, "kibana.size", "1g"),
					// resource.TestCheckResourceAttr(generalPurposeResName, "kibana.size_resource", "memory"),

					resource.TestCheckNoResourceAttr(generalPurposeResName, "apm"),
					resource.TestCheckNoResourceAttr(generalPurposeResName, "enterprise_search"),
				),
			},
		},
	})
}

func fixtureAccDeploymentResourceBasicCrossClusterSearch(t *testing.T, fileName, name, region, targetTplName, sourceName, sourceRegion, sourceTplName string) string {
	t.Helper()

	targetTpl := setDefaultTemplate(region, targetTplName)
	sourceTpl := setDefaultTemplate(region, sourceTplName)

	b, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		region, name, region, targetTpl,
		sourceName, sourceRegion, sourceTpl,
	)
}
