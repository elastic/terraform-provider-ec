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
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// This test creates a resource of type traffic filter with the randomName
// then it creates a data source that queries for this traffic filter by the id
func TestAccDatasource_trafficfilter(t *testing.T) {
	datasourceName := "data.ec_traffic_filter.name"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	region := getRegion()

	configVariables := config.Variables{
		"name":   config.StringVariable(randomName),
		"region": config.StringVariable(region),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		Steps: []resource.TestStep{
			{
				ConfigDirectory:    config.StaticDirectory("testdata/datasource_trafficfilter"),
				ConfigVariables:    configVariables,
				PreventDiskCleanup: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "rulesets.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "rulesets.0.name", randomName),
					resource.TestCheckResourceAttr(datasourceName, "rulesets.0.region", region),
				),
			},
		},
	})
}

// This test creates a remote_cluster traffic filter and verifies the data source
// returns the remote_cluster_id and remote_cluster_org_id fields correctly
func TestAccDatasource_trafficfilter_remoteCluster(t *testing.T) {
	datasourceName := "data.ec_traffic_filter.remote_cluster"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	region := getRegion()

	// Use test values for remote cluster - these would need to be valid for a real test
	remoteClusterID := os.Getenv("EC_TEST_REMOTE_CLUSTER_ID")
	remoteClusterOrgID := os.Getenv("EC_TEST_REMOTE_CLUSTER_ORG_ID")

	if remoteClusterID == "" {
		remoteClusterID = "test-remote-cluster-id"
	}
	if remoteClusterOrgID == "" {
		remoteClusterOrgID = "test-org-id"
	}

	configVariables := config.Variables{
		"name":                  config.StringVariable(randomName),
		"region":                config.StringVariable(region),
		"remote_cluster_id":     config.StringVariable(remoteClusterID),
		"remote_cluster_org_id": config.StringVariable(remoteClusterOrgID),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		Steps: []resource.TestStep{
			{
				ConfigDirectory:    config.StaticDirectory("testdata/datasource_trafficfilter_remote_cluster"),
				ConfigVariables:    configVariables,
				PreventDiskCleanup: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "rulesets.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "rulesets.0.name", randomName),
					resource.TestCheckResourceAttr(datasourceName, "rulesets.0.region", region),
					resource.TestCheckResourceAttr(datasourceName, "rulesets.0.rules.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "rulesets.0.rules.0.remote_cluster_id", remoteClusterID),
					resource.TestCheckResourceAttr(datasourceName, "rulesets.0.rules.0.remote_cluster_org_id", remoteClusterOrgID),
				),
			},
		},
	})
}
