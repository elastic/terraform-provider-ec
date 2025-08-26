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
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDeployment_basic_tf(t *testing.T) {
	resName := "ec_deployment.basic"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	startCfg := "testdata/deployment_basic.tf"
	randomAlias := "alias" + acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	trafficFilterCfg := "testdata/deployment_basic_with_traffic_filter_2.tf"
	trafficFilterUpdateCfg := "testdata/deployment_basic_with_traffic_filter_3.tf"
	resetPasswordCfg := "testdata/deployment_basic_reset_password.tf"
	cfg := fixtureAccDeploymentResourceBasicWithAppsAlias(t, startCfg, randomAlias, randomName, getRegion(), defaultTemplate)
	cfgWithTrafficFilter := fixtureAccDeploymentResourceBasicWithTF(t, trafficFilterCfg, randomName, getRegion(), defaultTemplate)
	cfgWithTrafficFilterUpdate := fixtureAccDeploymentResourceBasicWithTF(t, trafficFilterUpdateCfg, randomName, getRegion(), defaultTemplate)
	cfgResetPassword := fixtureAccDeploymentResourceBasicWithAppsAlias(t, resetPasswordCfg, randomAlias, randomName, getRegion(), defaultTemplate)
	deploymentVersion, err := latestStackVersion()
	if err != nil {
		t.Fatal(err)
	}

	elasticsearchPassword := ""
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: checkBasicDeploymentResource(resName, randomName, deploymentVersion,
					resource.TestCheckResourceAttr(resName, "alias", randomAlias),
					resource.TestCheckNoResourceAttr(resName, "apm.config"),
					resource.TestCheckNoResourceAttr(resName, "enterprise_search.config"),
					resource.TestCheckResourceAttr(resName, "traffic_filter.#", "0"),
					// Ensure at least 1 account is trusted (self).
					resource.TestCheckResourceAttr(resName, "elasticsearch.trust_account.#", "1"),
				),
			},
			{
				Config: cfgWithTrafficFilter,
				Check: checkBasicDeploymentResource(resName, randomName, deploymentVersion,
					// Ensure at least 1 account is trusted (self). It isn't deleted.
					resource.TestCheckResourceAttr(resName, "elasticsearch.trust_account.#", "1"),
					resource.TestCheckResourceAttr(resName, "traffic_filter.#", "1"),
				),
			},
			{
				Config: cfgWithTrafficFilterUpdate,
				Check: checkBasicDeploymentResource(resName, randomName, deploymentVersion,
					resource.TestCheckResourceAttr(resName, "traffic_filter.#", "1"),
				),
			},
			// Unset the traffic filter to remove the traffic filter
			{
				Config: cfg,
				Check: checkBasicDeploymentResource(resName, randomName, deploymentVersion,
					resource.TestCheckResourceAttr(resName, "traffic_filter.#", "0"),
					func(s *terraform.State) error {
						pw, ok := captureElasticsearchPassword(s, resName)
						if !ok {
							return errors.New("unable to capture current elasticsearch_password")
						}

						elasticsearchPassword = pw
						return nil
					},
				),
			},
			// Reset the elasticsearch_password
			{
				Config:             cfgResetPassword,
				ExpectNonEmptyPlan: true, // reset_elasticsearch_password will always result in a non-empty plan
				Check: checkBasicDeploymentResource(resName, randomName, deploymentVersion,
					resource.TestCheckResourceAttr(resName, "traffic_filter.#", "0"),
					resource.TestCheckResourceAttr(resName, "elasticsearch_username", "elastic"),
					func(s *terraform.State) error {
						currentPw, ok := captureElasticsearchPassword(s, resName)
						if !ok {
							return errors.New("unable to capture current elasticsearch_password")
						}

						if currentPw == elasticsearchPassword {
							return fmt.Errorf("expected elasticsearch_password to be reset: %s == %s", elasticsearchPassword, currentPw)
						}

						return nil
					},
				),
			},
		},
	})
}

func captureElasticsearchPassword(s *terraform.State, resName string) (string, bool) {
	res := s.RootModule().Resources[resName]
	pw, ok := res.Primary.Attributes["elasticsearch_password"]
	return pw, ok
}

func TestAccDeployment_basic_config(t *testing.T) {
	resName := "ec_deployment.basic"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	importCfg := "testdata/deployment_basic_settings_config_import.tf"
	settingsConfig := "testdata/deployment_basic_settings_config_2.tf"
	cfg := fixtureAccDeploymentResourceBasicWithApps(t, importCfg, randomName, getRegion(), defaultTemplate)
	settingsConfigCfg := fixtureAccDeploymentResourceBasicWithApps(t, settingsConfig, randomName, getRegion(), defaultTemplate)
	deploymentVersion, err := latestStackVersion()
	if err != nil {
		t.Fatal(err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: settingsConfigCfg,
				Check: checkBasicDeploymentResource(resName, randomName, deploymentVersion,
					resource.TestCheckResourceAttr(resName, "elasticsearch.config.user_settings_yaml", "action.auto_create_index: true"),
					resource.TestCheckResourceAttr(resName, "apm.config.debug_enabled", "true"),
					resource.TestCheckResourceAttr(resName, "apm.config.user_settings_json", `{"apm-server.rum.enabled":true}`),
					resource.TestCheckResourceAttr(resName, "kibana.config.user_settings_yaml", "csp.warnLegacyBrowsers: true"),
				),
			},
			{
				Config: cfg,
				Check: checkBasicDeploymentResource(resName, randomName, deploymentVersion,
					resource.TestCheckResourceAttr(resName, "apm.config.%", "0"),
					resource.TestCheckNoResourceAttr(resName, "elasticsearch.config.user_settings_yaml"),
					resource.TestCheckResourceAttr(resName, "kibana.config.%", "0"),
				),
			},
			// Import resource without complex ID
			{
				ResourceName:            resName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts", "apm_secret_token", "elasticsearch_password", "elasticsearch_username"},
			},
		},
	})
}

func fixtureAccDeploymentResourceBasicWithApps(t *testing.T, fileName, name, region, depTpl string) string {
	t.Helper()
	requiresAPIConn(t)

	deploymentTpl := setDefaultTemplate(region, depTpl)
	// esIC is no longer needed
	_, kibanaIC, apmIC, err := setInstanceConfigurations(deploymentTpl)
	if err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		region, name, region, deploymentTpl, kibanaIC, apmIC,
	)
}

func fixtureAccDeploymentResourceBasicWithAppsAlias(t *testing.T, fileName, alias, name, region, depTpl string) string {
	t.Helper()
	requiresAPIConn(t)

	deploymentTpl := setDefaultTemplate(region, depTpl)
	// esIC is no longer needed
	_, kibanaIC, apmIC, err := setInstanceConfigurations(deploymentTpl)
	if err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		region, alias, name, region, deploymentTpl, kibanaIC, apmIC,
	)
}

func fixtureAccDeploymentResourceBasicWithTF(t *testing.T, fileName, name, region, depTpl string) string {
	t.Helper()

	deploymentTpl := setDefaultTemplate(region, depTpl)
	b, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		region, name, region, deploymentTpl, name, region,
	)
}

func checkBasicDeploymentResource(resName, randomDeploymentName, deploymentVersion string, checks ...resource.TestCheckFunc) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(append([]resource.TestCheckFunc{
		testAccCheckDeploymentExists(resName),
		resource.TestCheckResourceAttr(resName, "name", randomDeploymentName),
		resource.TestCheckResourceAttr(resName, "region", getRegion()),
		resource.TestCheckResourceAttr(resName, "apm.region", getRegion()),
		resource.TestCheckResourceAttr(resName, "apm.size", "1g"),
		resource.TestCheckResourceAttr(resName, "apm.size_resource", "memory"),
		resource.TestCheckResourceAttrSet(resName, "apm_secret_token"),
		resource.TestCheckResourceAttrSet(resName, "elasticsearch_username"),
		resource.TestCheckResourceAttrSet(resName, "elasticsearch_password"),
		resource.TestCheckResourceAttrSet(resName, "apm.http_endpoint"),
		resource.TestCheckResourceAttrSet(resName, "apm.https_endpoint"),
		resource.TestCheckResourceAttr(resName, "elasticsearch.region", getRegion()),
		resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size", "1g"),
		resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size_resource", "memory"),
		resource.TestCheckResourceAttrSet(resName, "elasticsearch.http_endpoint"),
		resource.TestCheckResourceAttrSet(resName, "elasticsearch.https_endpoint"),
		resource.TestCheckResourceAttr(resName, "kibana.region", getRegion()),
		resource.TestCheckResourceAttr(resName, "kibana.size", "1g"),
		resource.TestCheckResourceAttr(resName, "kibana.size_resource", "memory"),
		resource.TestCheckResourceAttrSet(resName, "kibana.http_endpoint"),
		resource.TestCheckResourceAttrSet(resName, "kibana.https_endpoint"),
	}, checks...)...)
}
