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
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// creds is used as a container to pass around ES credentials.
type creds struct {
	User string
	Pass string
	URL  string
}

func TestAccDeployment_snapshot_restore(t *testing.T) {
	t.Skip("skipped due flakiness: https://github.com/elastic/terraform-provider-ec/issues/443")
	var esCreds creds
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: fixtureDeploymentDefaults(t, "testdata/deployment_snapshot_1.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					readEsCredentials(t, &esCreds),
				),
			},
			{
				// Creates a deployment restoring from snapshot. For some reason,
				// this can take quite a long time. It will have an impact on the
				// total run time of the acceptance tests.
				PreConfig: triggerSnapshot(t, &esCreds),
				Config:    fixtureDeploymentDefaults(t, "testdata/deployment_snapshot_2.tf"),
				// Since the `snapshot_source` block is never persisted, it'll
				// always have a non-empty plan after applying.
				ExpectNonEmptyPlan: true,
			},
			{
				// Triggers a new snapshot and restores the last snapshot to the
				// running Elasticsearch cluster. This ensures that the snapshot
				// is restored and the plan returns with no error.
				PreConfig: triggerSnapshot(t, &esCreds),
				Config:    fixtureDeploymentDefaults(t, "testdata/deployment_snapshot_2.tf"),
				// Since the `snapshot_source` block is never persisted, it'll
				// always have a non-empty plan after applying.
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func readEsCredentials(t *testing.T, esCreds *creds) resource.TestCheckFunc {
	t.Helper()
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "ec_deployment" {
				continue
			}

			esCreds.URL = rs.Primary.Attributes["elasticsearch.https_endpoint"]
			esCreds.User = rs.Primary.Attributes["elasticsearch_username"]
			esCreds.Pass = rs.Primary.Attributes["elasticsearch_password"]
		}
		return nil
	}
}

// remove comment after the test is unskipped.
// nolint
func triggerSnapshot(t *testing.T, esCreds *creds) func() {
	t.Helper()
	return func() {
		snapshotURL := fmt.Sprintf(
			esCreds.URL+"/_snapshot/found-snapshots/snap_%d?wait_for_completion=true",
			time.Now().Unix(),
		)
		req, err := http.NewRequest("PUT", snapshotURL, nil)
		if err != nil {
			t.Fatal(fmt.Errorf("failed creating snapshot request: %w", err))
			return
		}
		req.SetBasicAuth(esCreds.User, esCreds.Pass)

		t.Log("PUT", snapshotURL)

		// Create a new client with no timeout, just wait for the call to return.
		client := &http.Client{Timeout: 0}
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(fmt.Errorf("failed performing snapshot request: %w", err))
			return
		}
		if res.StatusCode != 200 {
			t.Fatal("snapshot create statuscode != 200")
			return
		}
		t.Log("Snapshot created")
	}
}
