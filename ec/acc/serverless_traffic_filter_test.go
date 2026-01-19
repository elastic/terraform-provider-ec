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
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServerlessTrafficFilter_basic(t *testing.T) {
	resName := "ec_serverless_traffic_filter.test"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessTrafficFilterBasic(randomName, getRegion()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "name", randomName),
					resource.TestCheckResourceAttr(resName, "region", getRegion()),
					resource.TestCheckResourceAttr(resName, "type", "ip"),
					resource.TestCheckResourceAttr(resName, "include_by_default", "false"),
					resource.TestCheckResourceAttr(resName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resName, "rule.*", map[string]string{
						"source": "0.0.0.0/0",
					}),
					resource.TestCheckResourceAttrSet(resName, "id"),
				),
			},
			{
				Config: testAccServerlessTrafficFilterUpdate(randomName, getRegion()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "name", randomName),
					resource.TestCheckResourceAttr(resName, "region", getRegion()),
					resource.TestCheckResourceAttr(resName, "type", "ip"),
					resource.TestCheckResourceAttr(resName, "include_by_default", "false"),
					resource.TestCheckResourceAttr(resName, "rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resName, "rule.*", map[string]string{
						"source": "0.0.0.0/0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resName, "rule.*", map[string]string{
						"source": "192.168.1.0/24",
					}),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccServerlessTrafficFilterBasic(name, region string) string {
	return fmt.Sprintf(`
resource "ec_serverless_traffic_filter" "test" {
  name               = "%s"
  region             = "%s"
  type               = "ip"
  include_by_default = false
  
  rule {
    source = "0.0.0.0/0"
  }
}
`, name, region)
}

func testAccServerlessTrafficFilterUpdate(name, region string) string {
	return fmt.Sprintf(`
resource "ec_serverless_traffic_filter" "test" {
  name               = "%s"
  region             = "%s"
  type               = "ip"
  include_by_default = false
  
  rule {
    source = "0.0.0.0/0"
  }
  
  rule {
    source = "192.168.1.0/24"
  }
}
`, name, region)
}
