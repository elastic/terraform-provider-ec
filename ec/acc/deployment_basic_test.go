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

func TestAccDeployment_basic_tf(t *testing.T) {
	resName := "ec_deployment.basic"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	startCfg := "testdata/deployment_basic.tf"
	randomAlias := "alias" + acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	trafficFilterCfg := "testdata/deployment_basic_with_traffic_filter_2.tf"
	trafficFilterUpdateCfg := "testdata/deployment_basic_with_traffic_filter_3.tf"
	cfg := fixtureAccDeploymentResourceBasicWithAppsAlias(t, startCfg, randomAlias, randomName, getRegion(), defaultTemplate)
	cfgWithTrafficFilter := fixtureAccDeploymentResourceBasicWithTF(t, trafficFilterCfg, randomName, getRegion(), defaultTemplate)
	cfgWithTrafficFilterUpdate := fixtureAccDeploymentResourceBasicWithTF(t, trafficFilterUpdateCfg, randomName, getRegion(), defaultTemplate)
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
				Config: cfg,
				Check: checkBasicDeploymentResource(resName, randomName, deploymentVersion,
					resource.TestCheckResourceAttr(resName, "alias", randomAlias),
					resource.TestCheckNoResourceAttr(resName, "apm.config"),
					resource.TestCheckNoResourceAttr(resName, "enterprise_search.config"),
					resource.TestCheckNoResourceAttr(resName, "traffic_filter"),
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
			// Remove traffic filter.
			{
				Config: cfg,
				Check: checkBasicDeploymentResource(resName, randomName, deploymentVersion,
					resource.TestCheckResourceAttr(resName, "traffic_filter.#", "0"),
				),
			},
		},
	})
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
					resource.TestCheckResourceAttr(resName, "enterprise_search.config.user_settings_yaml", "# comment"),
				),
			},
			{
				Config: cfg,
				Check: checkBasicDeploymentResource(resName, randomName, deploymentVersion,
					resource.TestCheckResourceAttr(resName, "apm.config.%", "0"),
					resource.TestCheckNoResourceAttr(resName, "elasticsearch.config.user_settings_yaml"),
					resource.TestCheckResourceAttr(resName, "kibana.config.%", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.config.%", "0"),
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
	_, kibanaIC, apmIC, essIC, err := setInstanceConfigurations(deploymentTpl)
	if err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		region, name, region, deploymentTpl, kibanaIC, apmIC, essIC,
	)
}

func fixtureAccDeploymentResourceBasicWithAppsAlias(t *testing.T, fileName, alias, name, region, depTpl string) string {
	t.Helper()
	requiresAPIConn(t)

	deploymentTpl := setDefaultTemplate(region, depTpl)
	// esIC is no longer needed
	_, kibanaIC, apmIC, essIC, err := setInstanceConfigurations(deploymentTpl)
	if err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b),
		region, alias, name, region, deploymentTpl, kibanaIC, apmIC, essIC,
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
		resource.TestCheckResourceAttr(resName, "enterprise_search.region", getRegion()),
		resource.TestCheckResourceAttr(resName, "enterprise_search.size", "2g"),
		resource.TestCheckResourceAttr(resName, "enterprise_search.size_resource", "memory"),
		resource.TestCheckResourceAttrSet(resName, "enterprise_search.http_endpoint"),
		resource.TestCheckResourceAttrSet(resName, "enterprise_search.https_endpoint"),
	}, checks...)...)
}
