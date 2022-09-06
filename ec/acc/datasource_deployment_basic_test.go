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

func TestAccDatasourceDeployment_basic(t *testing.T) {
	resourceName := "ec_deployment.basic_datasource"
	datasourceName := "data.ec_deployment.success"
	depsDatasourceName := "data.ec_deployments.query"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	secondRandomName := prefix + "-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	randomAlias := "alias" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	depCfg := "testdata/datasource_deployment_basic.tf"
	cfg := fixtureAccDeploymentDatasourceBasicAlias(t, depCfg, randomAlias, randomName, secondRandomName, getRegion(), computeOpTemplate)
	var namePrefix = secondRandomName[:22]

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "alias", resourceName, "alias"),
					resource.TestCheckResourceAttrPair(datasourceName, "region", resourceName, "region"),
					resource.TestCheckResourceAttrPair(datasourceName, "deployment_template_id", resourceName, "deployment_template_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "traffic_filter.#", resourceName, "traffic_filter.#"),

					// Elasticsearch
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.ref_id", resourceName, "elasticsearch.0.ref_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.cloud_id", resourceName, "elasticsearch.0.cloud_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.resource_id", resourceName, "elasticsearch.0.resource_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.http_endpoint_id", resourceName, "elasticsearch.0.http_endpoint_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.https_endpoint_id", resourceName, "elasticsearch.0.https_endpoint_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.topology.0.instance_configuration_id", resourceName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.topology.0.size", resourceName, "elasticsearch.0.topology.0.size"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.topology.0.size_resource", resourceName, "elasticsearch.0.topology.0.size_resource"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.topology.0.zone_count", resourceName, "elasticsearch.0.topology.0.zone_count"),
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch.0.topology.*.node_roles.*", resourceName, "elasticsearch.0.topology.*.node_roles.*"),

					// Kibana
					resource.TestCheckResourceAttrPair(datasourceName, "kibana.0.elasticsearch_cluster_ref_id", resourceName, "kibana.0.elasticsearch_cluster_ref_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "kibana.0.ref_id", resourceName, "kibana.0.ref_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "kibana.0.cloud_id", resourceName, "kibana.0.cloud_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "kibana.0.resource_id", resourceName, "kibana.0.resource_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "kibana.0.http_endpoint_id", resourceName, "kibana.0.http_endpoint_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "kibana.0.https_endpoint_id", resourceName, "kibana.0.https_endpoint_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "kibana.0.topology.0.instance_configuration_id", resourceName, "kibana.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "kibana.0.topology.0.size", resourceName, "kibana.0.topology.0.size"),
					resource.TestCheckResourceAttrPair(datasourceName, "kibana.0.topology.0.size_resource", resourceName, "kibana.0.topology.0.size_resource"),
					resource.TestCheckResourceAttrPair(datasourceName, "kibana.0.topology.0.zone_count", resourceName, "kibana.0.topology.0.zone_count"),

					// APM
					resource.TestCheckResourceAttrPair(datasourceName, "apm.0.elasticsearch_cluster_ref_id", resourceName, "apm.0.elasticsearch_cluster_ref_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "apm.0.ref_id", resourceName, "apm.0.ref_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "apm.0.cloud_id", resourceName, "apm.0.cloud_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "apm.0.resource_id", resourceName, "apm.0.resource_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "apm.0.http_endpoint_id", resourceName, "apm.0.http_endpoint_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "apm.0.https_endpoint_id", resourceName, "apm.0.https_endpoint_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "apm.0.topology.0.instance_configuration_id", resourceName, "apm.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "apm.0.topology.0.size", resourceName, "apm.0.topology.0.size"),
					resource.TestCheckResourceAttrPair(datasourceName, "apm.0.topology.0.size_resource", resourceName, "apm.0.topology.0.size_resource"),
					resource.TestCheckResourceAttrPair(datasourceName, "apm.0.topology.0.zone_count", resourceName, "apm.0.topology.0.zone_count"),

					// Enterprise Search
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.elasticsearch_cluster_ref_id", resourceName, "enterprise_search.0.elasticsearch_cluster_ref_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.ref_id", resourceName, "enterprise_search.0.ref_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.cloud_id", resourceName, "enterprise_search.0.cloud_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.resource_id", resourceName, "enterprise_search.0.resource_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.http_endpoint_id", resourceName, "enterprise_search.0.http_endpoint_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.https_endpoint_id", resourceName, "enterprise_search.0.https_endpoint_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.topology.0.instance_configuration_id", resourceName, "enterprise_search.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.topology.0.size", resourceName, "enterprise_search.0.topology.0.size"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.topology.0.size_resource", resourceName, "enterprise_search.0.topology.0.size_resource"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.topology.0.zone_count", resourceName, "enterprise_search.0.topology.0.zone_count"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.topology.0.node_type_appserver", resourceName, "enterprise_search.0.topology.0.node_type_appserver"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.topology.0.node_type_connector", resourceName, "enterprise_search.0.topology.0.node_type_connector"),
					resource.TestCheckResourceAttrPair(datasourceName, "enterprise_search.0.topology.0.node_type_worker", resourceName, "enterprise_search.0.topology.0.node_type_worker"),
				),
			},
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(depsDatasourceName, "name_prefix", namePrefix),
					resource.TestCheckResourceAttrPair(depsDatasourceName, "deployment_template_id", resourceName, "deployment_template_id"),

					// Verify Name and Alias is present
					resource.TestCheckResourceAttrPair(depsDatasourceName, "deployments.0.name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(depsDatasourceName, "deployments.0.alias", resourceName, "alias"),

					// Query results
					resource.TestCheckResourceAttrPair(depsDatasourceName, "deployments.0.elasticsearch_resource_id", resourceName, "elasticsearch.0.resource_id"),
					resource.TestCheckResourceAttrPair(depsDatasourceName, "deployments.0.kibana_resource_id", resourceName, "kibana.0.resource_id"),
					resource.TestCheckResourceAttrPair(depsDatasourceName, "deployments.0.apm_resource_id", resourceName, "apm.0.resource_id"),
					resource.TestCheckResourceAttrPair(depsDatasourceName, "deployments.0.enterprise_search_resource_id", resourceName, "enterprise_search.0.resource_id"),

					// Ref ID check.
					resource.TestCheckResourceAttrPair(depsDatasourceName, "deployments.0.elasticsearch_ref_id", resourceName, "elasticsearch.0.ref_id"),
					resource.TestCheckResourceAttrPair(depsDatasourceName, "deployments.0.kibana_ref_id", resourceName, "kibana.0.ref_id"),
					resource.TestCheckResourceAttrPair(depsDatasourceName, "deployments.0.apm_ref_id", resourceName, "apm.0.ref_id"),
					resource.TestCheckResourceAttrPair(depsDatasourceName, "deployments.0.enterprise_search_ref_id", resourceName, "enterprise_search.0.ref_id"),
				),
			},
		},
	})
}

func fixtureAccDeploymentDatasourceBasicAlias(t *testing.T, fileName, alias, name, secondName, region, depTpl string) string {
	t.Helper()

	deploymentTpl := setDefaultTemplate(region, depTpl)
	b, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		region, name, region, deploymentTpl, alias, secondName, region, deploymentTpl, secondName, region, deploymentTpl,
	)
}
