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
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/elastic/cloud-sdk-go/pkg/multierror"
)

func TestAccDeployment_withExtension(t *testing.T) {
	extResName := "ec_deployment_extension.my_extension"
	resName := "ec_deployment.with_extension"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	filePath := filepath.Join(os.TempDir(), "extension.zip")

	// TODO: this causes the test to fail with the invalid file error
	// however we need find a way to delete the temp file
	// defer os.Remove(filePath)

	cfg := fixtureAccDeploymentWithExtensionBundle(t,
		"testdata/deployment_with_extension_bundle_file.tf",
		getRegion(), randomName, "desc", filePath,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy: func(s *terraform.State) error {
			merr := multierror.NewPrefixed("checking resource with extension")

			if err := testAccExtensionDestroy(s); err != nil {
				merr = merr.Append(err)
			}
			if err := testAccDeploymentDestroy(s); err != nil {
				merr = merr.Append(err)
			}

			return merr.ErrorOrNil()
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() { writeFile(t, filePath, "extension.txt", "foo") },
				Config:    cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(extResName, "name", randomName),
					resource.TestCheckResourceAttr(extResName, "version", "*"),
					resource.TestCheckResourceAttr(extResName, "description", "desc"),
					resource.TestCheckResourceAttr(extResName, "extension_type", "bundle"),
					resource.TestCheckResourceAttr(extResName, "file_path", filePath),
					resource.TestCheckResourceAttr(resName, "elasticsearch.extension.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resName, "elasticsearch.extension.*", map[string]string{
						"type": "bundle",
						"name": randomName,
					}),
					func(s *terraform.State) error {
						return checkExtensionFile(t, s, "extension.txt", "foo")
					},
				),
			},
		},
	})
}

func fixtureAccDeploymentWithExtensionBundle(t *testing.T, filepath, region, name, desc, file string) string {
	t.Helper()

	b, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b), region,
		setDefaultTemplate(region, defaultTemplate), name, desc, file,
	)
}
