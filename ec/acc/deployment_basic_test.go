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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: checkBasicDeploymentResource(resName, randomName, deploymentVersion,
					resource.TestCheckResourceAttr(resName, "alias", randomAlias),
					resource.TestCheckResourceAttr(resName, "apm.0.config.#", "0"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.config.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.config.#", "0"),
					resource.TestCheckResourceAttr(resName, "traffic_filter.#", "0"),
				),
			},
			{
				Config: cfgWithTrafficFilter,
				Check: checkBasicDeploymentResource(resName, randomName, deploymentVersion,
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
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.config.#", "0"),
					resource.TestCheckResourceAttr(resName, "traffic_filter.#", "0"),
					resource.TestCheckResourceAttr(resName, "apm.0.config.#", "0"),
				),
			},
		},
	})
}

func TestAccDeployment_basic_config(t *testing.T) {
	resName := "ec_deployment.basic"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	startCfg := "testdata/deployment_basic_settings_config_1.tf"
	settingsConfig := "testdata/deployment_basic_settings_config_2.tf"
	cfg := fixtureAccDeploymentResourceBasicWithApps(t, startCfg, randomName, getRegion(), defaultTemplate)
	settingsConfigCfg := fixtureAccDeploymentResourceBasicWithApps(t, settingsConfig, randomName, getRegion(), defaultTemplate)
	deploymentVersion, err := latestStackVersion()
	if err != nil {
		t.Fatal(err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: settingsConfigCfg,
				Check: checkBasicDeploymentResource(resName, randomName, deploymentVersion,
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.config.0.user_settings_yaml", "action.auto_create_index: true"),
					resource.TestCheckResourceAttr(resName, "apm.0.config.0.debug_enabled", "true"),
					resource.TestCheckResourceAttr(resName, "apm.0.config.0.user_settings_json", `{"apm-server.rum.enabled":true}`),
					resource.TestCheckResourceAttr(resName, "kibana.0.config.#", "1"),
					resource.TestCheckResourceAttr(resName, "kibana.0.config.0.user_settings_yaml", "csp.warnLegacyBrowsers: true"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.config.#", "1"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.config.0.user_settings_yaml", "ent_search.login_assistance_message: somemessage"),
				),
			},
			{
				Config: cfg,
				Check: checkBasicDeploymentResource(resName, randomName, deploymentVersion,
					resource.TestCheckResourceAttr(resName, "apm.0.config.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.config.#", "0"),
					resource.TestCheckResourceAttr(resName, "apm.0.config.0.debug_enabled", "false"),
					resource.TestCheckResourceAttr(resName, "kibana.0.config.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.0.config.#", "0"),
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
		resource.TestCheckResourceAttr(resName, "apm.#", "1"),
		resource.TestCheckResourceAttr(resName, "apm.0.region", getRegion()),
		resource.TestCheckResourceAttr(resName, "apm.0.topology.0.size", "0.5g"),
		resource.TestCheckResourceAttr(resName, "apm.0.topology.0.size_resource", "memory"),
		resource.TestCheckResourceAttrSet(resName, "apm_secret_token"),
		resource.TestCheckResourceAttrSet(resName, "elasticsearch_username"),
		resource.TestCheckResourceAttrSet(resName, "elasticsearch_password"),
		resource.TestCheckResourceAttrSet(resName, "apm.0.http_endpoint"),
		resource.TestCheckResourceAttrSet(resName, "apm.0.https_endpoint"),
		resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
		resource.TestCheckResourceAttr(resName, "elasticsearch.0.region", getRegion()),
		resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size", "1g"),
		resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size_resource", "memory"),
		resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.http_endpoint"),
		resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.https_endpoint"),
		resource.TestCheckResourceAttr(resName, "kibana.#", "1"),
		resource.TestCheckResourceAttr(resName, "kibana.0.region", getRegion()),
		resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.size", "1g"),
		resource.TestCheckResourceAttr(resName, "kibana.0.topology.0.size_resource", "memory"),
		resource.TestCheckResourceAttrSet(resName, "kibana.0.http_endpoint"),
		resource.TestCheckResourceAttrSet(resName, "kibana.0.https_endpoint"),
		resource.TestCheckResourceAttr(resName, "enterprise_search.#", "1"),
		resource.TestCheckResourceAttr(resName, "enterprise_search.0.region", getRegion()),
		resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.size", "2g"),
		resource.TestCheckResourceAttr(resName, "enterprise_search.0.topology.0.size_resource", "memory"),
		resource.TestCheckResourceAttrSet(resName, "enterprise_search.0.http_endpoint"),
		resource.TestCheckResourceAttrSet(resName, "enterprise_search.0.https_endpoint"),
	}, checks...)...)
}
