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
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/elastic/cloud-sdk-go/pkg/client/extensions"
)

func TestAccDeploymentExtension_bundleFile(t *testing.T) {
	resName := "ec_deployment_extension.my_extension"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	filePath := filepath.Join(os.TempDir(), "extension.zip")
	defer os.Remove(filePath)

	cfg := fixtureAccExtensionBundleWithTF(t, "testdata/extension_bundle_file.tf", filePath, randomName, "desc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccExtensionDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() { writeFile(t, filePath, "extension.txt", "foo") },
				Config:    cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "name", randomName),
					resource.TestCheckResourceAttr(resName, "version", "*"),
					resource.TestCheckResourceAttr(resName, "description", "desc"),
					resource.TestCheckResourceAttr(resName, "extension_type", "bundle"),
					resource.TestCheckResourceAttr(resName, "file_path", filePath),
					func(s *terraform.State) error {
						return checkExtensionFile(t, s, "extension.txt", "foo")
					},
				),
			},
			{
				PreConfig: func() { writeFile(t, filePath, "extension.txt", "bar") },
				Config:    cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "name", randomName),
					resource.TestCheckResourceAttr(resName, "version", "*"),
					resource.TestCheckResourceAttr(resName, "description", "desc"),
					resource.TestCheckResourceAttr(resName, "extension_type", "bundle"),
					resource.TestCheckResourceAttr(resName, "file_path", filePath),
					func(s *terraform.State) error {
						return checkExtensionFile(t, s, "extension.txt", "bar")
					},
				),
			},
		},
	})
}

func fixtureAccExtensionBundleWithTF(t *testing.T, tfFileName, bundleFilePath, extensionName, description string) string {
	t.Helper()

	b, err := os.ReadFile(tfFileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		bundleFilePath, extensionName, description,
	)
}

func writeFile(t *testing.T, filePath, fileName, content string) {
	t.Helper()

	buf := new(bytes.Buffer)
	writer := zip.NewWriter(buf)

	f, err := writer.Create(fileName)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := f.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
}

func checkExtensionFile(t *testing.T, s *terraform.State, filename string, expected string) error {
	client, err := newAPI()
	if err != nil {
		t.Fatal(err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ec_deployment_extension" {
			continue
		}

		res, err := client.V1API.Extensions.GetExtension(
			extensions.NewGetExtensionParams().WithExtensionID(rs.Primary.ID),
			client.AuthWriter)
		if err != nil {
			t.Fatal(err)
		}

		content, err := downloadAndReadExtension(filename, res.Payload.FileMetadata.URL.String(), res.Payload.FileMetadata.Size)
		if err != nil {
			t.Fatal(err)
		}

		if content == expected {
			return nil // ok
		}
		return fmt.Errorf("extension content is expected: %s, but got: %s", expected, content)
	}

	return fmt.Errorf("extension doesn't exists")
}

func downloadAndReadExtension(filename string, url string, size int64) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	b, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	r, err := zip.NewReader(bytes.NewReader(b), size)
	if err != nil {
		return "", err
	}

	if len(r.File) == 0 {
		return "", fmt.Errorf("the zip file has no content")
	}

	for _, f := range r.File {
		reader, _ := f.Open()
		b, _ := io.ReadAll(reader)
		func() { defer reader.Close() }()
		if filename == f.Name {
			return string(b), nil
		}
	}
	return "", fmt.Errorf("not found: %s", filename)
}
