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

func TestAcc_ElasticsearchProject_AuditLogging(t *testing.T) {
	sourceID := "source"
	destID := "destination"
	resourceName := fmt.Sprintf("ec_elasticsearch_project.%s", sourceID)
	destResourceName := fmt.Sprintf("ec_observability_project.%s", destID)
	sourceName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	destName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	region := getRegion()
	if !strings.HasPrefix(region, "aws-") {
		region = fmt.Sprintf("aws-%s", region)
	}
	// Ids from the global ignore-filter catalog. Drop the filter steps if they are absent.
	dataReadsFilter := "audit-ignore-data-reads"
	uiRequestsFilter := "audit-ignore-ui-requests"

	destinationChecks := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "name", sourceName),
		resource.TestCheckResourceAttrPair(resourceName, "monitoring.logging.audit.destination.project_id", destResourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "monitoring.logging.audit.destination.project_type", "observability"),
		resource.TestCheckResourceAttrSet(resourceName, "monitoring.logging.audit.destination.status"),
	)
	auditChecks := resource.ComposeAggregateTestCheckFunc(
		destinationChecks,
		resource.TestCheckResourceAttr(resourceName, "monitoring.logging.audit.enabled", "true"),
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccElasticsearchProjectDestroy,
			testAccObservabilityProjectDestroy,
		),
		Steps: []resource.TestStep{
			{
				// Create the destination and a source project that audits to it.
				Config: testAccElasticsearchProjectWithAudit(sourceID, sourceName, region, destID, destName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					auditChecks,
					testCheckAttrNull(resourceName, "monitoring.logging.audit.ignore_filter_ids"),
				),
			},
			{
				// Import rebuilds the audit block from the API with no prior state.
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credentials"},
			},
			{
				// Add two ignore filters. The post-apply refresh plan proves the set
				// round-trips whatever order the API returns them in.
				Config: testAccElasticsearchProjectWithAudit(sourceID, sourceName, region, destID, destName, fmt.Sprintf(`ignore_filter_ids = ["%s", "%s"]`, uiRequestsFilter, dataReadsFilter)),
				Check: resource.ComposeAggregateTestCheckFunc(
					auditChecks,
					resource.TestCheckResourceAttr(resourceName, "monitoring.logging.audit.ignore_filter_ids.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "monitoring.logging.audit.ignore_filter_ids.*", dataReadsFilter),
					resource.TestCheckTypeSetElemAttr(resourceName, "monitoring.logging.audit.ignore_filter_ids.*", uiRequestsFilter),
				),
			},
			{
				// Replace the filters with a different single id; this patches the
				// array rather than clearing it.
				Config: testAccElasticsearchProjectWithAudit(sourceID, sourceName, region, destID, destName, fmt.Sprintf(`ignore_filter_ids = ["%s"]`, uiRequestsFilter)),
				Check: resource.ComposeAggregateTestCheckFunc(
					auditChecks,
					resource.TestCheckResourceAttr(resourceName, "monitoring.logging.audit.ignore_filter_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "monitoring.logging.audit.ignore_filter_ids.*", uiRequestsFilter),
				),
			},
			{
				// Pause delivery. The destination and filters must survive.
				Config: testAccElasticsearchProjectWithAudit(sourceID, sourceName, region, destID, destName, fmt.Sprintf("enabled = false\n\t\t\t\tignore_filter_ids = [%q]", uiRequestsFilter)),
				Check: resource.ComposeAggregateTestCheckFunc(
					destinationChecks,
					resource.TestCheckResourceAttr(resourceName, "monitoring.logging.audit.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "monitoring.logging.audit.ignore_filter_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "monitoring.logging.audit.ignore_filter_ids.*", uiRequestsFilter),
				),
			},
			{
				// Resume with the default enabled and remove the ignore filters; the
				// attribute must be null, not an empty set.
				Config: testAccElasticsearchProjectWithAudit(sourceID, sourceName, region, destID, destName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					auditChecks,
					testCheckAttrNull(resourceName, "monitoring.logging.audit.ignore_filter_ids"),
				),
			},
			{
				// Remove audit; state must end up with monitoring null.
				Config: testAccElasticsearchProjectWithDestination(sourceID, sourceName, region, destID, destName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", sourceName),
					testCheckAttrNull(resourceName, "monitoring"),
				),
			},
			{
				// A monitoring block without audit is a shell. It must apply
				// without touching the API and survive refresh: the framework's
				// post-apply plan fails this step if Read drops the shell.
				Config: testAccElasticsearchProjectWithDestination(sourceID, sourceName, region, destID, destName, "monitoring = { logging = {} }"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.TestCheckResourceAttr(resourceName, "name", sourceName),
			},
			{
				// The other shell shape behaves the same way.
				Config: testAccElasticsearchProjectWithDestination(sourceID, sourceName, region, destID, destName, "monitoring = {}"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.TestCheckResourceAttr(resourceName, "name", sourceName),
			},
		},
	})
}

// testAccElasticsearchProjectWithDestination creates an observability project
// and an elasticsearch project with no audit. extra is appended inside the
// elasticsearch resource.
func testAccElasticsearchProjectWithDestination(esID, esName, region, destID, destName, extra string) string {
	return fmt.Sprintf(`
resource ec_observability_project "%s" {
	name      = "%s"
	region_id = "%s"
}

resource ec_elasticsearch_project "%s" {
	name      = "%s"
	region_id = "%s"
	%s
}
`, destID, destName, region, esID, esName, region, extra)
}

// testAccElasticsearchProjectWithAudit audits esID into the observability
// project destID. auditExtra is appended inside the audit block.
func testAccElasticsearchProjectWithAudit(esID, esName, region, destID, destName, auditExtra string) string {
	return fmt.Sprintf(`
resource ec_observability_project "%s" {
	name      = "%s"
	region_id = "%s"
}

resource ec_elasticsearch_project "%s" {
	name      = "%s"
	region_id = "%s"

	monitoring = {
		logging = {
			audit = {
				destination = {
					project_id   = ec_observability_project.%s.id
					project_type = "observability"
				}
				%s
			}
		}
	}
}
`, destID, destName, region, esID, esName, region, destID, auditExtra)
}

// testCheckAttrNull fails when key is present in state, including as an
// empty collection or object, which TestCheckNoResourceAttr allows.
func testCheckAttrNull(resourceName, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		for k := range rs.Primary.Attributes {
			if k == key || strings.HasPrefix(k, key+".") {
				return fmt.Errorf("%s: attribute %q expected to be null, found %q", resourceName, key, k)
			}
		}
		return nil
	}
}

func testAccElasticsearchProjectDestroy(s *terraform.State) error {
	client, err := newServerlessAPI()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ec_elasticsearch_project" {
			continue
		}

		if err := assertServerlessProjectDeleted(
			func(ctx context.Context, id string) (bool, error) {
				res, err := client.GetElasticsearchProjectWithResponse(ctx, id)
				if err != nil {
					return false, err
				}
				return res.JSON200 != nil, nil
			},
			func(ctx context.Context, id string) error {
				res, err := client.DeleteElasticsearchProjectWithResponse(ctx, id, nil)
				return deleteServerlessProjectResponse(res, err)
			},
			"elasticsearch project",
			rs.Primary.ID,
		); err != nil {
			return err
		}
	}

	return nil
}
