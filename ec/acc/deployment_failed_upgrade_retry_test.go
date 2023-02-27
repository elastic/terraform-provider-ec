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
	"net/http"
	"regexp"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDeployment_failed_upgrade_retry(t *testing.T) {
	var esCreds creds
	resName := "ec_deployment.upgrade_retry"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: fixtureDeploymentDefaults(t, "testdata/deployment_upgrade_retry_1.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					readEsCredentials(t, &esCreds),
					checkMajorMinorVersion(t, resName, 7, 10),
				),
			},
			{
				// Creates an Elasticsearch index that will make the kibana upgrade fail.
				PreConfig:   createIndex(t, &esCreds, ".kibana_2"),
				Config:      fixtureDeploymentDefaults(t, "testdata/deployment_upgrade_retry_2.tf"),
				ExpectError: regexp.MustCompile(`\[kibana\].*Plan[ |\t|\n]+change[ |\t|\n]+failed.*`),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkMajorMinorVersion(t, resName, 7, 10),
				),
			},
			{
				// Deletes the Elasticsearch index so that the upgrade succeeds.
				PreConfig: deleteIndex(t, &esCreds, ".kibana_2"),
				Config:    fixtureDeploymentDefaults(t, "testdata/deployment_upgrade_retry_2.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkMajorMinorVersion(t, resName, 7, 11),
				),
			},
		},
	})
}

func createIndex(t *testing.T, esCreds *creds, indexName string) func() {
	t.Helper()
	return func() {
		indexURL := fmt.Sprintf(
			esCreds.URL+"/%s", indexName,
		)
		req, err := http.NewRequest("PUT", indexURL, nil)
		if err != nil {
			t.Fatal(fmt.Errorf("failed creating snapshot request: %w", err))
			return
		}
		req.SetBasicAuth(esCreds.User, esCreds.Pass)

		t.Log("PUT", indexURL)

		// Create a new client with no timeout, just wait for the call to return.
		client := &http.Client{Timeout: 0}
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(fmt.Errorf("failed creating index %s request: %w", indexName, err))
			return
		}
		if res.StatusCode != 200 {
			t.Fatal("index create statuscode != 200")
			return
		}
		t.Logf("Index %s created", indexName)
	}
}

func deleteIndex(t *testing.T, esCreds *creds, indexName string) func() {
	t.Helper()
	return func() {
		indexURL := fmt.Sprintf(
			esCreds.URL+"/%s", indexName,
		)
		req, err := http.NewRequest("DELETE", indexURL, nil)
		if err != nil {
			t.Fatal(fmt.Errorf("failed creating snapshot request: %w", err))
			return
		}
		req.SetBasicAuth(esCreds.User, esCreds.Pass)

		t.Log("DELETE", indexURL)

		// Create a new client with no timeout, just wait for the call to return.
		client := &http.Client{Timeout: 0}
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(fmt.Errorf("failed creating index %s request: %w", indexName, err))
			return
		}
		if res.StatusCode != 200 {
			t.Fatal("index delete statuscode != 200")
			return
		}
		t.Logf("Index %s deleted", indexName)
	}
}

func checkMajorMinorVersion(t *testing.T, resName string, major, minor uint64) resource.TestCheckFunc {
	t.Helper()
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "ec_deployment" {
				continue
			}
			m := rs.Primary.Attributes

			v, err := semver.Parse(m["version"])
			if err != nil {
				return fmt.Errorf("failed parsing deployment semver version %s: %w", m["version"], err)
			}

			if v.Major != major || v.Minor != minor {
				return fmt.Errorf(
					"found version %d.%d != expected version %d.%d",
					v.Major, v.Minor,
					major, minor,
				)
			}
		}
		return nil
	}
}
