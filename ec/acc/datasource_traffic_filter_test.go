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

func TestAccDatasource_trafficfilter(t *testing.T) {
	datasourceName := "data.ec_trafficfilter.id"
	depCfg := "testdata/datasource_trafficfilter.tf"
	cfg := fixtureAccTrafficFilterDataSource(t, depCfg, getRegion())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		Steps: []resource.TestStep{
			{
				Config:             cfg,
				PreventDiskCleanup: true,
				Check: checkDataSourceTrafficFilter(datasourceName,
					resource.TestCheckResourceAttr(datasourceName, "region", getRegion()),
				),
			},
		},
	})
}

func fixtureAccTrafficFilterDataSource(t *testing.T, fileName string, region string) string {
	t.Helper()

	b, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b), region)
}

func checkDataSourceTrafficFilter(resName string, checks ...resource.TestCheckFunc) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(append([]resource.TestCheckFunc{
		resource.TestCheckResourceAttr(resName, "rulesets.#", "2"),
		resource.TestCheckResourceAttr(resName, "rulesets.0.region", getRegion()),
		resource.TestCheckResourceAttr(resName, "rulesets.0.name", "example-filter"),
	}, checks...)...)
}
