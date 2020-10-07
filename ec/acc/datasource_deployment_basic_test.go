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

func TestAccDatasourceDeployment_basic(t *testing.T) {
	resourceName := "ec_deployment.basic_datasource"
	datasourceName := "data.ec_deployment.success"
	depsDatasourceName := "data.ec_deployments.query"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	depCfg := "testdata/datasource_deployment_basic.tf"
	cfg := testAccDeploymentDatasourceBasic(t, depCfg, randomName, region, deploymentVersion)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config:             cfg,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "region", resourceName, "region"),
					resource.TestCheckResourceAttrPair(datasourceName, "deployment_template_id", resourceName, "deployment_template_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "traffic_filter.#", resourceName, "traffic_filter.#"),

					// Elasticsearch
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.ref_id", resourceName, "elasticsearch.0.ref_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.cloud_id", resourceName, "elasticsearch.0.cloud_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.resource_id", resourceName, "elasticsearch.0.resource_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.http_endpoint_id", resourceName, "elasticsearch.0.http_endpoint_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.https_endpoint_id", resourceName, "elasticsearch.0.https_endpoint_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.version", resourceName, "elasticsearch.0.version"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.topology.0.instance_configuration_id", resourceName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.topology.0.memory_per_node", resourceName, "elasticsearch.0.topology.0.memory_per_node"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.topology.0.node_count_per_zone", resourceName, "elasticsearch.0.topology.0.node_count_per_zone"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.topology.0.zone_count", resourceName, "elasticsearch.0.topology.0.zone_count"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.topology.0.node_type_data", resourceName, "elasticsearch.0.topology.0.node_type_data"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.topology.0.node_type_master", resourceName, "elasticsearch.0.topology.0.node_type_master"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.topology.0.node_type_ingest", resourceName, "elasticsearch.0.topology.0.node_type_ingest"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.topology.0.node_type_ml", resourceName, "elasticsearch.0.topology.0.node_type_ml"),

					// Kibana
					resource.TestCheckResourceAttrPair(datasourceName, "kibana.0.elasticsearch_cluster_ref_id", resourceName, "kibana.0.elasticsearch_cluster_ref_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "kibana.0.ref_id", resourceName, "kibana.0.ref_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "kibana.0.cloud_id", resourceName, "kibana.0.cloud_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "kibana.0.resource_id", resourceName, "kibana.0.resource_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "kibana.0.http_endpoint_id", resourceName, "kibana.0.http_endpoint_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "kibana.0.https_endpoint_id", resourceName, "kibana.0.https_endpoint_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "kibana.0.version", resourceName, "kibana.0.version"),
					resource.TestCheckResourceAttrPair(datasourceName, "kibana.0.topology.0.instance_configuration_id", resourceName, "kibana.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "kibana.0.topology.0.memory_per_node", resourceName, "kibana.0.topology.0.memory_per_node"),
					resource.TestCheckResourceAttrPair(datasourceName, "kibana.0.topology.0.zone_count", resourceName, "kibana.0.topology.0.zone_count"),

					// APM
					resource.TestCheckResourceAttrPair(datasourceName, "apm.0.elasticsearch_cluster_ref_id", resourceName, "apm.0.elasticsearch_cluster_ref_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "apm.0.ref_id", resourceName, "apm.0.ref_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "apm.0.cloud_id", resourceName, "apm.0.cloud_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "apm.0.resource_id", resourceName, "apm.0.resource_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "apm.0.http_endpoint_id", resourceName, "apm.0.http_endpoint_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "apm.0.https_endpoint_id", resourceName, "apm.0.https_endpoint_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "apm.0.version", resourceName, "apm.0.version"),
					resource.TestCheckResourceAttrPair(datasourceName, "apm.0.topology.0.instance_configuration_id", resourceName, "apm.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "apm.0.topology.0.memory_per_node", resourceName, "apm.0.topology.0.memory_per_node"),
					resource.TestCheckResourceAttrPair(datasourceName, "apm.0.topology.0.zone_count", resourceName, "apm.0.topology.0.zone_count"),

					// Enterprise Search
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.elasticsearch_cluster_ref_id", resourceName, "enterprise_search.0.elasticsearch_cluster_ref_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.ref_id", resourceName, "enterprise_search.0.ref_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.cloud_id", resourceName, "enterprise_search.0.cloud_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.resource_id", resourceName, "enterprise_search.0.resource_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.http_endpoint_id", resourceName, "enterprise_search.0.http_endpoint_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.https_endpoint_id", resourceName, "enterprise_search.0.https_endpoint_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.version", resourceName, "enterprise_search.0.version"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.topology.0.instance_configuration_id", resourceName, "enterprise_search.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.topology.0.memory_per_node", resourceName, "enterprise_search.0.topology.0.memory_per_node"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.topology.0.node_count_per_zone", resourceName, "enterprise_search.0.topology.0.node_count_per_zone"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.topology.0.zone_count", resourceName, "enterprise_search.0.topology.0.zone_count"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.topology.0.node_type_appserver", resourceName, "enterprise_search.0.topology.0.node_type_appserver"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.topology.0.node_type_connector", resourceName, "enterprise_search.0.topology.0.node_type_connector"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.topology.0.node_type_worker", resourceName, "enterprise_search.0.topology.0.node_type_worker"),
				),
			},
			{
				Config:             cfg,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(depsDatasourceName, "name_prefix", "terraform_acc_"),
					resource.TestCheckResourceAttr(depsDatasourceName, "deployment_template_id", "aws-compute-optimized-v2"),

					// Deployment resources
					resource.TestCheckResourceAttr(depsDatasourceName, "elasticsearch.0.version", deploymentVersion),
					resource.TestCheckResourceAttr(depsDatasourceName, "kibana.0.version", deploymentVersion),
					resource.TestCheckResourceAttr(depsDatasourceName, "apm.0.version", deploymentVersion),
					resource.TestCheckResourceAttr(depsDatasourceName, "enterprise_search.0.version", deploymentVersion),

					// Query results
					resource.TestCheckResourceAttrPair(depsDatasourceName, "deployments.0.elasticsearch_resource_id", resourceName, "elasticsearch.0.resource_id"),
					resource.TestCheckResourceAttrPair(depsDatasourceName, "deployments.0.kibana_resource_id", resourceName, "kibana.0.resource_id"),
					resource.TestCheckResourceAttrPair(depsDatasourceName, "deployments.0.apm_resource_id", resourceName, "apm.0.resource_id"),
					resource.TestCheckResourceAttrPair(depsDatasourceName, "deployments.0.enterprise_search_resource_id", resourceName, "enterprise_search.0.resource_id"),
				),
			},
		},
	})
}

func testAccDeploymentDatasourceBasic(t *testing.T, fileName, name, region, version string) string {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		name, region, version, name, region, version, version, version, version,
	)
}
