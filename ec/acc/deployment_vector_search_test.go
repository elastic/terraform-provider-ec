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

// This test case takes that on a vector search "ec_deployment".
func TestAccDeployment_vector_search(t *testing.T) {
	vectorSearchResName := "ec_deployment.vector_search"
	sourceResName := "ec_deployment.source_vector_search.0"

	vectorSearchRandomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	sourceRandomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	startCfg := "testdata/deployment_vector_search_1.tf"
	secondCfg := "testdata/deployment_vector_search_2.tf"
	cfg := fixtureAccDeploymentResourceBasicVectorSearch(t, startCfg,
		vectorSearchRandomName, getRegion(), vectorSearchTemplate,
		sourceRandomName, getRegion(), defaultTemplate,
	)
	secondConfigCfg := fixtureAccDeploymentResourceBasicDefaults(t, secondCfg, vectorSearchRandomName, getRegion(), vectorSearchTemplate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				// Create a vector search deployment with the default settings.
				Config: cfg,
				// The legacy vector search DT does not support autoscaling, which leads to autoscaling being 'unknown'.
				// Ideally we would set autoscaling to null if the deployment template does not support autoscaling,
				// but that would require's refactoring our schema and this template is no longer part of the public offering.
				//
				// We can revisit this if there's demand for clean plans when the template does not support autoscaling.
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(

					// vector search Checks
					resource.TestCheckResourceAttrSet(vectorSearchResName, "elasticsearch.hot.instance_configuration_id"),
					// vector search defaults to 1g.
					resource.TestCheckResourceAttr(vectorSearchResName, "elasticsearch.hot.size", "1g"),
					resource.TestCheckResourceAttr(vectorSearchResName, "elasticsearch.hot.size_resource", "memory"),

					// Remote cluster settings
					resource.TestCheckResourceAttr(vectorSearchResName, "elasticsearch.remote_cluster.#", "3"),
					resource.TestCheckResourceAttrSet(vectorSearchResName, "elasticsearch.remote_cluster.0.deployment_id"),
					resource.TestCheckResourceAttr(vectorSearchResName, "elasticsearch.remote_cluster.0.alias", fmt.Sprint(sourceRandomName, "-0")),
					resource.TestCheckResourceAttrSet(vectorSearchResName, "elasticsearch.remote_cluster.1.deployment_id"),
					resource.TestCheckResourceAttr(vectorSearchResName, "elasticsearch.remote_cluster.1.alias", fmt.Sprint(sourceRandomName, "-1")),
					resource.TestCheckResourceAttrSet(vectorSearchResName, "elasticsearch.remote_cluster.2.deployment_id"),
					resource.TestCheckResourceAttr(vectorSearchResName, "elasticsearch.remote_cluster.2.alias", fmt.Sprint(sourceRandomName, "-2")),

					resource.TestCheckNoResourceAttr(vectorSearchResName, "elasticsearch.hot.node_type_data"),
					resource.TestCheckNoResourceAttr(vectorSearchResName, "elasticsearch.hot.node_type_ingest"),
					resource.TestCheckNoResourceAttr(vectorSearchResName, "elasticsearch.hot.node_type_master"),
					resource.TestCheckNoResourceAttr(vectorSearchResName, "elasticsearch.hot.node_type_ml"),
					resource.TestCheckResourceAttrSet(vectorSearchResName, "elasticsearch.hot.node_roles.#"),
					resource.TestCheckResourceAttr(vectorSearchResName, "elasticsearch.hot.zone_count", "1"),
					resource.TestCheckNoResourceAttr(sourceResName, "kibana"),
					resource.TestCheckNoResourceAttr(sourceResName, "apm"),
					resource.TestCheckNoResourceAttr(sourceResName, "enterprise_search"),

					// Source Checks
					resource.TestCheckResourceAttrSet(sourceResName, "elasticsearch.hot.instance_configuration_id"),
					resource.TestCheckResourceAttr(sourceResName, "elasticsearch.hot.size", "1g"),
					resource.TestCheckResourceAttr(sourceResName, "elasticsearch.hot.size_resource", "memory"),
					resource.TestCheckNoResourceAttr(vectorSearchResName, "elasticsearch.hot.node_type_data"),
					resource.TestCheckNoResourceAttr(vectorSearchResName, "elasticsearch.hot.node_type_ingest"),
					resource.TestCheckNoResourceAttr(vectorSearchResName, "elasticsearch.hot.node_type_master"),
					resource.TestCheckNoResourceAttr(vectorSearchResName, "elasticsearch.hot.node_type_ml"),
					resource.TestCheckResourceAttrSet(sourceResName, "elasticsearch.hot.node_roles.#"),
					resource.TestCheckResourceAttr(sourceResName, "elasticsearch.hot.zone_count", "1"),
					resource.TestCheckNoResourceAttr(sourceResName, "kibana"),
					resource.TestCheckNoResourceAttr(sourceResName, "apm"),
					resource.TestCheckNoResourceAttr(sourceResName, "enterprise_search"),
				),
			},
			{
				// Change the Elasticsearch topology size and node count.
				Config:             secondConfigCfg,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Changes.
					resource.TestCheckResourceAttrSet(vectorSearchResName, "elasticsearch.hot.instance_configuration_id"),
					resource.TestCheckResourceAttr(vectorSearchResName, "elasticsearch.hot.size", "2g"),
					resource.TestCheckResourceAttr(vectorSearchResName, "elasticsearch.hot.size_resource", "memory"),

					resource.TestCheckResourceAttr(vectorSearchResName, "elasticsearch.remote_cluster.#", "0"),

					resource.TestCheckNoResourceAttr(vectorSearchResName, "elasticsearch.hot.node_type_data"),
					resource.TestCheckNoResourceAttr(vectorSearchResName, "elasticsearch.hot.node_type_ingest"),
					resource.TestCheckNoResourceAttr(vectorSearchResName, "elasticsearch.hot.node_type_master"),
					resource.TestCheckNoResourceAttr(vectorSearchResName, "elasticsearch.hot.node_type_ml"),

					resource.TestCheckResourceAttrSet(vectorSearchResName, "elasticsearch.hot.node_roles.#"),
					resource.TestCheckResourceAttr(vectorSearchResName, "elasticsearch.hot.zone_count", "1"),
					resource.TestCheckResourceAttr(vectorSearchResName, "kibana.zone_count", "1"),
					resource.TestCheckResourceAttrSet(vectorSearchResName, "kibana.instance_configuration_id"),
					resource.TestCheckResourceAttr(vectorSearchResName, "kibana.size", "1g"),
					resource.TestCheckResourceAttr(vectorSearchResName, "kibana.size_resource", "memory"),
					resource.TestCheckNoResourceAttr(vectorSearchResName, "apm"),
					resource.TestCheckNoResourceAttr(vectorSearchResName, "enterprise_search"),
				),
			},
		},
	})
}

func fixtureAccDeploymentResourceBasicVectorSearch(t *testing.T, fileName, name, region, vectorSearchTplName, sourceName, sourceRegion, sourceTplName string) string {
	t.Helper()

	vectorSearchTpl := setDefaultTemplate(region, vectorSearchTplName)
	sourceTpl := setDefaultTemplate(region, sourceTplName)

	b, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		region, name, region, vectorSearchTpl,
		sourceName, sourceRegion, sourceTpl,
	)
}
