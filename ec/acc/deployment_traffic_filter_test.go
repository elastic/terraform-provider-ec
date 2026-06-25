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
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeploymentTrafficFilter_basic(t *testing.T) {
	resName := "ec_deployment_traffic_filter.basic"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	region := getRegion()

	configVariables := config.Variables{
		"name":   config.StringVariable(randomName),
		"region": config.StringVariable(region),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentTrafficFilterDestroy,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/deployment_traffic_filter_basic"),
				ConfigVariables: configVariables,
				Check: checkBasicDeploymentTrafficFilterResource(resName, randomName, region,
					resource.TestCheckResourceAttr(resName, "include_by_default", "false"),
					resource.TestCheckResourceAttr(resName, "type", "ip"),
					resource.TestCheckResourceAttr(resName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resName, "rule.0.source", "0.0.0.0/0"),
				),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/deployment_traffic_filter_basic_update"),
				ConfigVariables: configVariables,
				Check: checkBasicDeploymentTrafficFilterResource(resName, randomName, region,
					resource.TestCheckResourceAttr(resName, "include_by_default", "false"),
					resource.TestCheckResourceAttr(resName, "type", "ip"),
					resource.TestCheckResourceAttr(resName, "rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resName, "rule.*", map[string]string{
						"source": "0.0.0.0/0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resName, "rule.*", map[string]string{
						"source": "1.1.1.0/24",
					}),
				),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/deployment_traffic_filter_basic_update_large"),
				ConfigVariables: configVariables,
				Check: checkBasicDeploymentTrafficFilterResource(resName, randomName, region,
					resource.TestCheckResourceAttr(resName, "include_by_default", "false"),
					resource.TestCheckResourceAttr(resName, "type", "ip"),
					resource.TestCheckResourceAttr(resName, "rule.#", "16"),
					resource.TestCheckTypeSetElemNestedAttrs(resName, "rule.*", map[string]string{
						"source": "8.8.8.8/24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resName, "rule.*", map[string]string{
						"source": "8.8.4.4/24",
					}),
				),
			},
			{
				ConfigDirectory:         config.StaticDirectory("testdata/deployment_traffic_filter_basic_update_large"),
				ConfigVariables:         configVariables,
				ResourceName:            resName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
		},
	})
}

func TestAccDeploymentTrafficFilter_azure(t *testing.T) {
	resName := "ec_deployment_traffic_filter.azure"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	region := "azure-australiaeast"

	configVariables := config.Variables{
		"name":   config.StringVariable(randomName),
		"region": config.StringVariable(region),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentTrafficFilterDestroy,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/deployment_traffic_filter_azure"),
				ConfigVariables: configVariables,
				Check: checkBasicDeploymentTrafficFilterResource(resName, randomName, region,
					resource.TestCheckResourceAttr(resName, "include_by_default", "false"),
					resource.TestCheckResourceAttr(resName, "type", "azure_private_endpoint"),
					resource.TestCheckResourceAttr(resName, "rule.#", "1"),
				),
				ExpectError: regexp.MustCompile(`.*traffic_filter.azure_private_link_connection_not_found.*`),
			},
		},
	})
}

func TestAccDeploymentTrafficFilter_remoteCluster(t *testing.T) {
	resName := "ec_deployment_traffic_filter.remote_cluster"
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
		CheckDestroy:             testAccDeploymentTrafficFilterDestroy,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/deployment_traffic_filter_remote_cluster"),
				ConfigVariables: configVariables,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentTrafficFilterExists(resName),
					resource.TestCheckResourceAttr(resName, "name", randomName),
					resource.TestCheckResourceAttr(resName, "region", region),
					resource.TestCheckResourceAttr(resName, "include_by_default", "false"),
					resource.TestCheckResourceAttr(resName, "type", "remote_cluster"),
					resource.TestCheckResourceAttr(resName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resName, "rule.*", map[string]string{
						"remote_cluster_id":     remoteClusterID,
						"remote_cluster_org_id": remoteClusterOrgID,
					}),
				),
			},
			{
				ConfigDirectory:         config.StaticDirectory("testdata/deployment_traffic_filter_remote_cluster"),
				ConfigVariables:         configVariables,
				ResourceName:            resName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
		},
	})
}

func TestAccDeploymentTrafficFilter_UpgradeFrom0_4_1(t *testing.T) {
	t.Skip("skip until `ec_deployment` state upgrade is implemented")

	resName := "ec_deployment_traffic_filter.basic"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	region := getRegion()

	configVariables := config.Variables{
		"name":   config.StringVariable(randomName),
		"region": config.StringVariable(region),
	}

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
				ConfigDirectory: config.StaticDirectory("testdata/deployment_traffic_filter_basic"),
				ConfigVariables: configVariables,
				Check: checkBasicDeploymentTrafficFilterResource(resName, randomName, region,
					resource.TestCheckResourceAttr(resName, "include_by_default", "false"),
					resource.TestCheckResourceAttr(resName, "type", "ip"),
					resource.TestCheckResourceAttr(resName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resName, "rule.0.source", "0.0.0.0/0"),
				),
			},
			{
				PlanOnly:                 true,
				ProtoV6ProviderFactories: testAccProviderFactory,
				ConfigDirectory:          config.StaticDirectory("testdata/deployment_traffic_filter_basic"),
				ConfigVariables:          configVariables,
				Check: checkBasicDeploymentTrafficFilterResource(resName, randomName, region,
					resource.TestCheckResourceAttr(resName, "include_by_default", "false"),
					resource.TestCheckResourceAttr(resName, "type", "ip"),
					resource.TestCheckResourceAttr(resName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resName, "rule.0.source", "0.0.0.0/0"),
				),
			},
		},
	})
}

func checkBasicDeploymentTrafficFilterResource(resName, randomDeploymentName, region string, checks ...resource.TestCheckFunc) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(append([]resource.TestCheckFunc{
		testAccCheckDeploymentTrafficFilterExists(resName),
		resource.TestCheckResourceAttr(resName, "name", randomDeploymentName),
		resource.TestCheckResourceAttr(resName, "region", region)}, checks...)...,
	)
}
