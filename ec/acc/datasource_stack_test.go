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

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatasourceStack_latest(t *testing.T) {
	datasourceName := "data.ec_stack.latest"
	depCfg := "testdata/datasource_stack_latest.tf"
	cfg := fixtureAccStackDataSource(t, depCfg, getRegion())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		Steps: []resource.TestStep{
			{
				Config:             cfg,
				PreventDiskCleanup: true,
				Check: checkDataSourceStack(datasourceName,
					resource.TestCheckResourceAttr(datasourceName, "version_regex", "latest"),
					resource.TestCheckResourceAttr(datasourceName, "lock", "true"),
					resource.TestCheckResourceAttr(datasourceName, "region", getRegion()),
				),
			},
		},
	})
}

func TestAccDatasourceStack_regex(t *testing.T) {
	datasourceName := "data.ec_stack.regex"
	depCfg := "testdata/datasource_stack_regex.tf"
	cfg := fixtureAccStackDataSource(t, depCfg, getRegion())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		Steps: []resource.TestStep{
			{
				Config:             cfg,
				PreventDiskCleanup: true,
				Check: checkDataSourceStack(datasourceName,
					resource.TestCheckResourceAttr(datasourceName, "version_regex", "8.4.?"),
					resource.TestCheckResourceAttr(datasourceName, "region", getRegion()),
				),
			},
		},
	})
}

func fixtureAccStackDataSource(t *testing.T, fileName, region string) string {
	t.Helper()

	b, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b), region)
}

func checkDataSourceStack(resName string, checks ...resource.TestCheckFunc) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(append([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resName, "version"),
		resource.TestCheckResourceAttrSet(resName, "accessible"),
		resource.TestCheckResourceAttrSet(resName, "min_upgradable_from"),
		resource.TestCheckResourceAttrSet(resName, "allowlisted"),

		// Elasticsearch
		resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.denylist.#"),
		resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.capacity_constraints_max"),
		resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.capacity_constraints_min"),
		resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.docker_image"),
		resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.plugins.#"),
		resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.default_plugins.#"),

		// Kibana
		resource.TestCheckResourceAttrSet(resName, "kibana.0.denylist.#"),
		resource.TestCheckResourceAttrSet(resName, "kibana.0.capacity_constraints_max"),
		resource.TestCheckResourceAttrSet(resName, "kibana.0.capacity_constraints_min"),
		resource.TestCheckResourceAttrSet(resName, "kibana.0.docker_image"),

		// APM
		resource.TestCheckResourceAttrSet(resName, "apm.0.capacity_constraints_max"),
		resource.TestCheckResourceAttrSet(resName, "apm.0.capacity_constraints_min"),
		resource.TestCheckResourceAttrSet(resName, "apm.0.docker_image"),

		// Enterprise Search
		resource.TestCheckResourceAttrSet(resName, "enterprise_search.0.capacity_constraints_max"),
		resource.TestCheckResourceAttrSet(resName, "enterprise_search.0.capacity_constraints_min"),
		resource.TestCheckResourceAttrSet(resName, "enterprise_search.0.docker_image"),
	}, checks...)...)
}
