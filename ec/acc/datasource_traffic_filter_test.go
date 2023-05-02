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

// This test creates a resource of type traffic filter with the randomName
// then it creates a data source that queries for this traffic filter by the id
func TestAccDatasource_trafficfilter(t *testing.T) {
	datasourceName := "data.ec_traffic_filter.name"
	depCfg := "testdata/datasource_trafficfilter.tf"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	cfg := fixtureAccTrafficFilterDataSource(t, depCfg, randomName, getRegion())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		Steps: []resource.TestStep{
			{
				Config:             cfg,
				PreventDiskCleanup: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "rulesets.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "rulesets.0.name", randomName),
					resource.TestCheckResourceAttr(datasourceName, "rulesets.0.region", getRegion()),
				),
			},
		},
	})
}

func fixtureAccTrafficFilterDataSource(t *testing.T, fileName string, name string, region string) string {
	t.Helper()

	b, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b), name, region)
}
