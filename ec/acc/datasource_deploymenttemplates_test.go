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
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"os"
	"testing"
)

func TestAcc_datasource_deploymenttemplates(t *testing.T) {
	cfg := renderTerraformFile(t, "testdata/datasource_deploymenttemplates.tf", getRegion())
	datasourceName := "data.ec_deployment_templates.test"
	datasourceNameById := "data.ec_deployment_templates.by_id"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		Steps: []resource.TestStep{
			{
				Config:             cfg,
				PreventDiskCleanup: true,
				Check: resource.ComposeTestCheckFunc(
					// Checks that there is at least one template
					// (As the templates are dynamic depending on the region the test runs against)
					resource.TestCheckResourceAttrSet(datasourceName, "templates.0.id"),
					resource.TestCheckResourceAttrSet(datasourceName, "templates.0.name"),
					resource.TestCheckResourceAttrSet(datasourceName, "templates.0.description"),
					resource.TestCheckResourceAttrSet(datasourceName, "templates.0.elasticsearch.hot.instance_configuration_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "templates.0.elasticsearch.hot.default_size"),
					resource.TestCheckResourceAttrSet(datasourceName, "templates.0.elasticsearch.hot.available_sizes.#"),
					resource.TestCheckResourceAttrSet(datasourceName, "templates.0.elasticsearch.hot.size_resource"),
					resource.TestCheckResourceAttrSet(datasourceName, "templates.0.kibana.instance_configuration_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "templates.0.kibana.default_size"),
					resource.TestCheckResourceAttrSet(datasourceName, "templates.0.kibana.size_resource"),

					// Template found by id
					resource.TestCheckResourceAttrSet(datasourceNameById, "templates.0.id"),
					resource.TestCheckResourceAttrSet(datasourceNameById, "templates.0.name"),
					resource.TestCheckResourceAttrSet(datasourceNameById, "templates.0.description"),
				),
			},
		},
	})
}

func renderTerraformFile(t *testing.T, fileName string, region string) string {
	t.Helper()
	b, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b), region, region)
}
