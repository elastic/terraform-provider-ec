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

func TestAccDeploymentExtension_basic(t *testing.T) {
	resName := "ec_deployment_extension.my_extension"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	cfg := fixtureAccExtensionBasicWithTF(t, "testdata/extension_basic.tf", randomName, "desc")
	cfg2 := fixtureAccExtensionBasicWithTF(t, "testdata/extension_basic.tf", randomName, "updated desc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccExtensionDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "name", randomName),
					resource.TestCheckResourceAttr(resName, "version", "*"),
					resource.TestCheckResourceAttr(resName, "description", "desc"),
					resource.TestCheckResourceAttr(resName, "extension_type", "bundle"),
				),
			},
			{
				Config: cfg2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "name", randomName),
					resource.TestCheckResourceAttr(resName, "version", "*"),
					resource.TestCheckResourceAttr(resName, "description", "updated desc"),
					resource.TestCheckResourceAttr(resName, "extension_type", "bundle"),
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

func TestAccDeploymentExtension_UpgradeFrom0_4_1(t *testing.T) {
	resName := "ec_deployment_extension.my_extension"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	cfg := fixtureAccExtensionBasicWithTF(t, "testdata/extension_basic.tf", randomName, "desc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccDeploymentTrafficFilterDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"ec": {
						VersionConstraint: "0.4.1",
						Source:            "elastic/ec",
					},
				},
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "name", randomName),
					resource.TestCheckResourceAttr(resName, "version", "*"),
					resource.TestCheckResourceAttr(resName, "description", "desc"),
					resource.TestCheckResourceAttr(resName, "extension_type", "bundle"),
				),
			},
			{
				PlanOnly:                 true,
				ProtoV6ProviderFactories: testAccProviderFactory,
				Config:                   cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "name", randomName),
					resource.TestCheckResourceAttr(resName, "version", "*"),
					resource.TestCheckResourceAttr(resName, "description", "desc"),
					resource.TestCheckResourceAttr(resName, "extension_type", "bundle"),
				),
			},
		},
	})
}

func fixtureAccExtensionBasicWithTF(t *testing.T, tfFileName, extensionName, description string) string {
	t.Helper()

	b, err := os.ReadFile(tfFileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		extensionName, description,
	)
}
