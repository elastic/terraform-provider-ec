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
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// issue933Config declares an ec_deployment with an explicit, entirely empty
// `kibana.config = {}` object: no sub-attribute is set at all.
const issue933Config = `
data "ec_stack" "issue_933" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "issue_933" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.issue_933.version
  deployment_template_id = "%s"

  elasticsearch = {
    hot = {
      size = "1g"
    }
  }

  kibana = {
    instance_configuration_id = "%s"
    config                    = {}
  }
}
`

// TestAccReproduceIssue933 reproduces
// https://github.com/elastic/terraform-provider-ec/issues/933.
//
// The Kibana schema's "config" attribute is Optional but not Computed
// (ec/ecresource/deploymentresource/kibana/v2/schema.go), so when a user
// writes an empty object literal, `kibana.config = {}`, Terraform core keeps
// that as the planned value: a non-null object whose leaf attributes
// (docker_image, user_settings_json, user_settings_override_json,
// user_settings_yaml, user_settings_override_yaml) are all null.
//
// After apply, readKibanaConfig
// (ec/ecresource/deploymentresource/kibana/v2/kibana_config.go) builds the
// new state from the API response: since the deployment has no Kibana
// config settings at all, every field stays empty and the function hits
// `if cfg == (KibanaConfig{}) { return nil, nil }`, collapsing the whole
// object to null. Unlike Elasticsearch - which has a dedicated nil-out guard
// for this exact situation in
// ec/ecresource/deploymentresource/read.go - Kibana (and, by the same code
// path, Apm and IntegrationsServer) has no such guard, so the planned
// non-null object and the actual null state disagree and Terraform's
// post-apply consistency check fails with:
//
//	Error: Provider produced inconsistent result after apply
//
//	When applying changes to ec_deployment.issue_933, provider
//	"provider[\"registry.terraform.io/elastic/ec\"]" produced an unexpected
//	new value: .kibana.config: was
//	cty.ObjectVal(map[string]cty.Value{"docker_image":cty.NullVal(cty.String),
//	"user_settings_json":cty.NullVal(cty.String),
//	"user_settings_override_json":cty.NullVal(cty.String),
//	"user_settings_override_yaml":cty.NullVal(cty.String),
//	"user_settings_yaml":cty.NullVal(cty.String)}), but now null.
//
// Once fixed, applying this configuration should succeed with no such
// error, and the check below should pass.
func TestAccReproduceIssue933(t *testing.T) {
	requiresAPIConn(t)

	resName := "ec_deployment.issue_933"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	region := getRegion()
	deploymentTpl := setDefaultTemplate(region, defaultTemplate)

	_, kibanaIC, _, err := getInstanceConfigurations(deploymentTpl)
	if err != nil {
		t.Fatal(err)
	}

	cfg := fmt.Sprintf(issue933Config, region, randomName, region, deploymentTpl, kibanaIC)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resName, "id"),
					resource.TestCheckResourceAttr(resName, "kibana.config.%", "0"),
				),
			},
		},
	})
}
