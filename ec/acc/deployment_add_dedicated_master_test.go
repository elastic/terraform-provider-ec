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
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeployment_add_dedicated_master(t *testing.T) {
	resName := "ec_deployment.auto_dedicated_master"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	cfg5nodes := buildConfiguration(t, "testdata/deployment_dedicated_master_5_nodes.tf", randomName, getRegion(), defaultTemplate)
	cfg6nodes := buildConfiguration(t, "testdata/deployment_dedicated_master_6_nodes.tf", randomName, getRegion(), defaultTemplate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg6nodes,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Master tier should be enabled
					resource.TestCheckResourceAttrWith(
						resName,
						"elasticsearch.master.size",
						func(v string) error {
							if v == "0g" || v == "" {
								return errors.New("master size should not be empty. size=" + v)
							}
							return nil
						}),
					resource.TestCheckResourceAttrWith(
						resName,
						"elasticsearch.master.zone_count",
						func(v string) error {
							if v == "0" || v == "" {
								return errors.New("master zone_count should not be empty. zone_count=" + v)
							}
							return nil
						}),
				),
			},
			{
				Config: cfg5nodes,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Master tier should be disabled
					resource.TestCheckNoResourceAttr(resName, "elasticsearch.master"),
				),
			},
		},
	})
}

func buildConfiguration(t *testing.T, fileName, name, region, depTpl string) string {
	t.Helper()
	requiresAPIConn(t)

	deploymentTpl := setDefaultTemplate(region, depTpl)

	b, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		region, name, region, deploymentTpl,
	)
}
