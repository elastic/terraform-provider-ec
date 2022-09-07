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

func TestAccDeploymentExtension_pluginDownload(t *testing.T) {
	resName := "ec_deployment_extension.my_extension"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	downloadURL := "https://artifacts.elastic.co/downloads/elasticsearch-plugins/analysis-icu/analysis-icu-7.10.1.zip"

	cfg := fixtureAccExtensionBundleDownloadWithTF(t, "testdata/extension_plugin_download.tf", randomName, downloadURL)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccExtensionDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "name", randomName),
					resource.TestCheckResourceAttr(resName, "version", "7.10.1"),
					resource.TestCheckResourceAttr(resName, "download_url", downloadURL),
					resource.TestCheckResourceAttr(resName, "extension_type", "plugin"),
				),
			},
		},
	})
}

func fixtureAccExtensionBundleDownloadWithTF(t *testing.T, tfFileName, extensionName, downloadURL string) string {
	t.Helper()

	b, err := os.ReadFile(tfFileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b), extensionName, downloadURL)
}
