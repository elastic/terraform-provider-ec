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
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAcc_ElasticsearchProject(t *testing.T) {
	resId := "my_project"
	resourceName := fmt.Sprintf("ec_elasticsearch_project.%s", resId)
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	alias := "alias-for-acc-test-project"
	newName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	region := getRegion()
	if !strings.HasPrefix(region, "aws-") {
		region = fmt.Sprintf("aws-%s", region)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccElasticsearchProjectDestroy,
		Steps: []resource.TestStep{
			{
				// Create a basic project.
				Config: testAccBasicElasticsearchProject(resId, randomName, region),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckResourceAttrSet(resourceName, "alias"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.elasticsearch"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.kibana"),
					resource.TestCheckResourceAttrSet(resourceName, "credentials.username"),
					resource.TestCheckResourceAttrSet(resourceName, "credentials.password"),
					resource.TestCheckResourceAttrSet(resourceName, "cloud_id"),
				),
			},
			{
				// Explicitly set the alias.
				Config: testAccElasticsearchProjectWithAlias(resId, randomName, region, alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckResourceAttr(resourceName, "alias", alias),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.elasticsearch"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.kibana"),
					resource.TestCheckResourceAttrSet(resourceName, "credentials.username"),
					resource.TestCheckResourceAttrSet(resourceName, "credentials.password"),
					resource.TestCheckResourceAttrSet(resourceName, "cloud_id"),
				),
			},
			{
				// Change the name.
				Config: testAccElasticsearchProjectWithAlias(resId, newName, region, alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", newName),
					resource.TestCheckResourceAttr(resourceName, "alias", alias),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.elasticsearch"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.kibana"),
					resource.TestCheckResourceAttrSet(resourceName, "credentials.username"),
					resource.TestCheckResourceAttrSet(resourceName, "credentials.password"),
					resource.TestCheckResourceAttrSet(resourceName, "cloud_id"),
				),
			},
			{
				// Test import.
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credentials"},
			},
		},
	})
}

func testAccBasicElasticsearchProject(id string, name string, region string) string {
	return fmt.Sprintf(`
resource ec_elasticsearch_project "%s" {
	name = "%s"
	region_id = "%s"
}
`, id, name, region)
}

func testAccElasticsearchProjectWithAlias(id string, name string, region string, alias string) string {
	return fmt.Sprintf(`
resource ec_elasticsearch_project "%s" {
	name = "%s"
	region_id = "%s"
	alias = "%s"
}
`, id, name, region, alias)
}

func TestAcc_ElasticsearchProject_MetadataTags(t *testing.T) {
	resId := "tags_project"
	resourceName := fmt.Sprintf("ec_elasticsearch_project.%s", resId)
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	region := getRegion()
	if !strings.HasPrefix(region, "aws-") {
		region = fmt.Sprintf("aws-%s", region)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccElasticsearchProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchProjectWithMetadataTag(resId, randomName, region, "v1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckResourceAttr(resourceName, "metadata.tags.acc_test", "v1"),
				),
			},
			{
				Config: testAccElasticsearchProjectWithMetadataTag(resId, randomName, region, "v2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckResourceAttr(resourceName, "metadata.tags.acc_test", "v2"),
				),
			},
			{
				Config: testAccElasticsearchProjectWithMetadataTagTeam(resId, randomName, region, "platform"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckResourceAttr(resourceName, "metadata.tags.acc_team", "platform"),
					resource.TestCheckResourceAttr(resourceName, "metadata.tags.%", "1"),
				),
			},
			{
				Config: testAccElasticsearchProjectWithMetadataTagTeam(resId, randomName, region, "platform"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckResourceAttr(resourceName, "metadata.tags.acc_team", "platform"),
					resource.TestCheckResourceAttr(resourceName, "metadata.tags.%", "1"),
				),
			},
			{
				Config: testAccElasticsearchProjectWithEmptyMetadataTags(resId, randomName, region),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckNoResourceAttr(resourceName, "metadata.tags.acc_team"),
					resource.TestCheckResourceAttr(resourceName, "metadata.tags.%", "0"),
				),
			},
		},
	})
}

func testAccElasticsearchProjectWithMetadataTag(id, name, region, tagValue string) string {
	return fmt.Sprintf(`
resource ec_elasticsearch_project "%s" {
	name      = "%s"
	region_id = "%s"
	metadata = {
		tags = {
			acc_test = "%s"
		}
	}
}
`, id, name, region, tagValue)
}

func testAccElasticsearchProjectWithMetadataTagTeam(id, name, region, team string) string {
	return fmt.Sprintf(`
resource ec_elasticsearch_project "%s" {
	name      = "%s"
	region_id = "%s"
	metadata = {
		tags = {
			acc_team = "%s"
		}
	}
}
`, id, name, region, team)
}

func testAccElasticsearchProjectWithEmptyMetadataTags(id, name, region string) string {
	return fmt.Sprintf(`
resource ec_elasticsearch_project "%s" {
	name      = "%s"
	region_id = "%s"
	metadata = {
		tags = {}
	}
}
`, id, name, region)
}

func TestAcc_ElasticsearchProjectImport(t *testing.T) {
	resId := "import_project"
	resourceName := fmt.Sprintf("ec_elasticsearch_project.%s", resId)
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	region := getRegion()
	if !strings.HasPrefix(region, "aws-") {
		region = fmt.Sprintf("aws-%s", region)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccElasticsearchProjectDestroy,
		Steps: []resource.TestStep{
			{
				// Create a project to import.
				Config: testAccBasicElasticsearchProject(resId, randomName, region),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				// Import the project and verify all attributes.
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credentials"},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckResourceAttr(resourceName, "region_id", region),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "alias"),
					resource.TestCheckResourceAttrSet(resourceName, "cloud_id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.elasticsearch"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.kibana"),
					resource.TestCheckResourceAttrSet(resourceName, "metadata.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "metadata.created_by"),
					resource.TestCheckResourceAttrSet(resourceName, "metadata.organization_id"),
					resource.TestCheckResourceAttr(resourceName, "type", "elasticsearch"),
				),
			},
		},
	})
}

func TestAcc_ElasticsearchProject_LinkedProjects(t *testing.T) {
	originID := "origin"
	targetIDA := "target_a"
	targetIDB := "target_b"
	resourceName := fmt.Sprintf("ec_elasticsearch_project.%s", originID)
	targetAResourceName := fmt.Sprintf("ec_observability_project.%s", targetIDA)
	originName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	targetAName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	targetBName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	region := getRegion()
	if !strings.HasPrefix(region, "aws-") {
		region = fmt.Sprintf("aws-%s", region)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccElasticsearchProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchProjectWithLinkedObservability(originID, originName, region, targetIDA, targetAName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", originName),
					resource.TestCheckResourceAttrSet(resourceName, "linked.projects.%"),
					testCheckLinkedProject(resourceName, targetAResourceName, "observability"),
				),
			},
			{
				Config: testAccElasticsearchProjectWithLinkedObservability(originID, originName, region, targetIDA, targetAName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", originName),
					testCheckLinkedProject(resourceName, targetAResourceName, "observability"),
				),
			},
			{
				// Link a second project and verify both are present.
				Config: testAccElasticsearchProjectWithLinkedObservabilityProjects(originID, originName, region, targetIDA, targetAName, targetIDB, targetBName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", originName),
					resource.TestCheckResourceAttr(resourceName, "linked.projects.%", "2"),
					testCheckLinkedProject(resourceName, targetAResourceName, "observability"),
				),
			},
			{
				// Remove the second project from the config; the provider must unlink it.
				Config: testAccElasticsearchProjectWithLinkedObservabilityAndSecondProject(originID, originName, region, targetIDA, targetAName, targetIDB, targetBName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", originName),
					resource.TestCheckResourceAttr(resourceName, "linked.projects.%", "1"),
					testCheckLinkedProject(resourceName, targetAResourceName, "observability"),
				),
			},
		},
	})
}

func testAccElasticsearchProjectWithLinkedObservability(esID, esName, region, targetID, targetName string) string {
	return fmt.Sprintf(`
resource ec_observability_project "%s" {
	name      = "%s"
	region_id = "%s"
}

resource ec_elasticsearch_project "%s" {
	name      = "%s"
	region_id = "%s"

	linked = {
		projects = {
			"${ec_observability_project.%s.id}" = {
				type = "observability"
			}
		}
	}
}
`, targetID, targetName, region, esID, esName, region, targetID)
}
func testAccElasticsearchProjectWithLinkedObservabilityAndSecondProject(esID, esName, region, targetID1, targetName1, targetID2, targetName2 string) string {
	return fmt.Sprintf(`
resource ec_observability_project "%s" {
	name      = "%s"
	region_id = "%s"
}

resource ec_observability_project "%s" {
	name      = "%s"
	region_id = "%s"
}

resource ec_elasticsearch_project "%s" {
	name      = "%s"
	region_id = "%s"

	linked = {
		projects = {
			"${ec_observability_project.%s.id}" = {
				type = "observability"
			}
		}
	}
}
`, targetID1, targetName1, region, targetID2, targetName2, region, esID, esName, region, targetID1)
}

func testAccElasticsearchProjectWithLinkedObservabilityProjects(esID, esName, region, targetID1, targetName1, targetID2, targetName2 string) string {
	return fmt.Sprintf(`
resource ec_observability_project "%s" {
	name      = "%s"
	region_id = "%s"
}

resource ec_observability_project "%s" {
	name      = "%s"
	region_id = "%s"
}

resource ec_elasticsearch_project "%s" {
	name      = "%s"
	region_id = "%s"

	linked = {
		projects = {
			"${ec_observability_project.%s.id}" = {
				type = "observability"
			}
			"${ec_observability_project.%s.id}" = {
				type = "observability"
			}
		}
	}
}
`, targetID1, targetName1, region, targetID2, targetName2, region, esID, esName, region, targetID1, targetID2)
}

func testCheckLinkedProject(resourceName, targetResourceName, targetType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[targetResourceName]
		if !ok {
			return fmt.Errorf("target resource not found: %s", targetResourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("target resource has no ID: %s", targetResourceName)
		}

		origin, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		typeKey := fmt.Sprintf("linked.projects.%s.type", rs.Primary.ID)
		statusKey := fmt.Sprintf("linked.statuses.%s", rs.Primary.ID)

		if got := origin.Primary.Attributes[typeKey]; got != targetType {
			return fmt.Errorf("expected linked project type %q, got %q", targetType, got)
		}
		if got := origin.Primary.Attributes[statusKey]; got == "" {
			return fmt.Errorf("linked project status not set")
		}

		return nil
	}
}

func testAccElasticsearchProjectDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	client, err := newServerlessAPI()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ec_elasticsearch_project" {
			continue
		}

		res, err := client.GetElasticsearchProjectWithResponse(context.Background(), rs.Primary.ID)

		// The resource will only exist if it can be obtained via the API and
		// the metadata status is not set to hidden. Currently ESS clients
		// cannot delete a deployment, so even when it's been shut down it will
		// show up on the GET call.
		if err == nil && res.JSON200 != nil {
			res, err := client.DeleteElasticsearchProjectWithResponse(context.Background(), rs.Primary.ID, nil)
			if err != nil && res.StatusCode() == 200 {
				return nil
			}

			return fmt.Errorf("elasticsearch project [%s] still exists", rs.Primary.ID)
		}
	}

	return nil
}
