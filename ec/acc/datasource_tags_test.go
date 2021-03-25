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

// This test case takes ensures that the tag metadata of an "ec_deployment"
// datasource is returned:
// * Create a deployment resource with tags.
// * Create a datasource from the resource.
// * Ensure tags exist.
// * Create a datasource with filters for the unique tag
// * Ensure that only a single deployment is returned
func TestAccDatasource_basic_tags(t *testing.T) {

	datasourceName := "data.ec_deployment.tagdata"
	filterDatasourceName := "data.ec_deployments.tagfilter"
	depCfg := "testdata/datasource_tags.tf"
	testID := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	randomName := prefix + testID
	cfg := fixtureAccTagsDataSource(t, depCfg, randomName, getRegion(), defaultTemplate, testID)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check the "ec_deployment" datasource
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(datasourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(datasourceName, "tags.bar", "baz"),
					resource.TestCheckResourceAttr(datasourceName, "tags.test_id", testID),
				),
			},
			{
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check the "ec_deployments" datasource, which filters on tags.
					resource.TestCheckResourceAttrSet(filterDatasourceName, "return_count"),
					resource.TestCheckResourceAttr(filterDatasourceName, "return_count", "1"),
					resource.TestCheckResourceAttr(filterDatasourceName, "deployments.#", "1"),
					resource.TestCheckResourceAttrSet(filterDatasourceName, "deployments.0.deployment_id"),
				),
			},
		},
	})
}

func fixtureAccTagsDataSource(t *testing.T, fileName, name, region string, depTpl string, testID string) string {
	t.Helper()

	deploymentTpl := setDefaultTemplate(region, depTpl)
	b, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		region, name, region, deploymentTpl, testID, testID,
	)
}
