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

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/elastic/cloud-sdk-go/pkg/multierror"
)

func TestAccDeploymentElasticsearchKeystore_full(t *testing.T) {
	var previousID, currentID string

	resType := "ec_deployment_elasticsearch_keystore"
	firstResName := resType + ".test"
	secondResName := resType + ".gcs_creds"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	startCfg := "testdata/deployment_elasticsearch_keystore_1.tf"
	updateKeystoreSetting := "testdata/deployment_elasticsearch_keystore_2.tf"
	changeKeystoreSettingName := "testdata/deployment_elasticsearch_keystore_3.tf"
	deleteAllKeystoreSettings := "testdata/deployment_elasticsearch_keystore_4.tf"

	cfgF := func(cfg string) string {
		return fixtureAccDeploymentResourceBasic(
			t, cfg, randomName, getRegion(), defaultTemplate,
		)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccDeploymentDestroy,
			testAccDeploymentElasticsearchKeystoreDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: cfgF(startCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(firstResName, "setting_name", "xpack.notification.slack.account.hello.secure_url"),
					resource.TestCheckResourceAttr(firstResName, "value", "hella"),
					resource.TestCheckResourceAttr(firstResName, "as_file", "false"),
					resource.TestCheckResourceAttrSet(firstResName, "deployment_id"),

					resource.TestCheckResourceAttr(secondResName, "setting_name", "gcs.client.secondary.credentials_file"),
					resource.TestCheckResourceAttr(secondResName, "value", "{\n  \"type\": \"service_account\",\n  \"project_id\": \"project-id\",\n  \"private_key_id\": \"key-id\",\n  \"private_key\": \"-----BEGIN PRIVATE KEY-----\\nprivate-key\\n-----END PRIVATE KEY-----\\n\",\n  \"client_email\": \"service-account-email\",\n  \"client_id\": \"client-id\",\n  \"auth_uri\": \"https://accounts.google.com/o/oauth2/auth\",\n  \"token_uri\": \"https://accounts.google.com/o/oauth2/token\",\n  \"auth_provider_x509_cert_url\": \"https://www.googleapis.com/oauth2/v1/certs\",\n  \"client_x509_cert_url\": \"https://www.googleapis.com/robot/v1/metadata/x509/service-account-email\"\n}"),
					resource.TestCheckResourceAttr(secondResName, "as_file", "false"),
					resource.TestCheckResourceAttrSet(secondResName, "deployment_id"),
				),
			},
			{
				Config: cfgF(updateKeystoreSetting),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkESKeystoreResourceID(firstResName, &previousID),

					resource.TestCheckResourceAttr(firstResName, "setting_name", "xpack.notification.slack.account.hello.secure_url"),
					resource.TestCheckResourceAttr(firstResName, "value", "hello2u"),
					resource.TestCheckResourceAttr(firstResName, "as_file", "false"),
					resource.TestCheckResourceAttrSet(firstResName, "deployment_id"),

					resource.TestCheckResourceAttr(secondResName, "setting_name", "gcs.client.secondary.credentials_file"),
					resource.TestCheckResourceAttr(secondResName, "value", "{\n  \"type\": \"service_account\",\n  \"project_id\": \"project-id\",\n  \"private_key_id\": \"key-id\",\n  \"private_key\": \"-----BEGIN PRIVATE KEY-----\\nprivate-key\\n-----END PRIVATE KEY-----\\n\",\n  \"client_email\": \"service-account-email\",\n  \"client_id\": \"client-id\",\n  \"auth_uri\": \"https://accounts.google.com/o/oauth2/auth\",\n  \"token_uri\": \"https://accounts.google.com/o/oauth2/token\",\n  \"auth_provider_x509_cert_url\": \"https://www.googleapis.com/oauth2/v1/certs\",\n  \"client_x509_cert_url\": \"https://www.googleapis.com/robot/v1/metadata/x509/service-account-email\"\n}"),
					resource.TestCheckResourceAttr(secondResName, "as_file", "false"),
					resource.TestCheckResourceAttrSet(secondResName, "deployment_id"),
				),
			},
			{
				Config: cfgF(changeKeystoreSettingName),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkESKeystoreResourceID(firstResName, &currentID),

					resource.TestCheckResourceAttr(firstResName, "setting_name", "xpack.notification.slack.account.hello.secure_urla"),
					resource.TestCheckResourceAttr(firstResName, "value", "hello2u"),
					resource.TestCheckResourceAttr(firstResName, "as_file", "false"),
					resource.TestCheckResourceAttrSet(firstResName, "deployment_id"),

					resource.TestCheckResourceAttr(secondResName, "setting_name", "gcs.client.secondary.credentials_file"),
					resource.TestCheckResourceAttr(secondResName, "value", "{\n  \"type\": \"service_account\",\n  \"project_id\": \"project-id\",\n  \"private_key_id\": \"key-id\",\n  \"private_key\": \"-----BEGIN PRIVATE KEY-----\\nprivate-key\\n-----END PRIVATE KEY-----\\n\",\n  \"client_email\": \"service-account-email\",\n  \"client_id\": \"client-id\",\n  \"auth_uri\": \"https://accounts.google.com/o/oauth2/auth\",\n  \"token_uri\": \"https://accounts.google.com/o/oauth2/token\",\n  \"auth_provider_x509_cert_url\": \"https://www.googleapis.com/oauth2/v1/certs\",\n  \"client_x509_cert_url\": \"https://www.googleapis.com/robot/v1/metadata/x509/service-account-email\"\n}"),
					resource.TestCheckResourceAttr(secondResName, "as_file", "false"),
					resource.TestCheckResourceAttrSet(secondResName, "deployment_id"),
				),
			},
			{
				Config: cfgF(deleteAllKeystoreSettings),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkNoKeystoreResourcesLeft(firstResName, secondResName),
					func(current, previous *string) resource.TestCheckFunc {
						return func(s *terraform.State) error {
							if *current == *previous {
								return fmt.Errorf("%s id (%s) should not equal %s", firstResName, *current, *previous)
							}
							return nil
						}
					}(&currentID, &previousID),
				),
			},
		},
	})
}

func TestAccDeploymentElasticsearchKeystore_UpgradeFrom0_4_1(t *testing.T) {
	resType := "ec_deployment_elasticsearch_keystore"
	firstResName := resType + ".test"
	secondResName := resType + ".gcs_creds"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	startCfg := "testdata/deployment_elasticsearch_keystore_1_041.tf"
	migratedCfg := "testdata/deployment_elasticsearch_keystore_1_migrated.tf"

	cfgF := func(cfg string) string {
		return fixtureAccDeploymentResourceBasic(
			t, cfg, randomName, getRegion(), defaultTemplate,
		)
	}

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"ec": {
						VersionConstraint: "0.4.1",
						Source:            "elastic/ec",
					},
				},
				Config: cfgF(startCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(firstResName, "setting_name", "xpack.notification.slack.account.hello.secure_url"),
					resource.TestCheckResourceAttr(firstResName, "value", "hella"),
					resource.TestCheckResourceAttr(firstResName, "as_file", "false"),
					resource.TestCheckResourceAttrSet(firstResName, "deployment_id"),

					resource.TestCheckResourceAttr(secondResName, "setting_name", "gcs.client.secondary.credentials_file"),
					resource.TestCheckResourceAttr(secondResName, "value", "{\n  \"type\": \"service_account\",\n  \"project_id\": \"project-id\",\n  \"private_key_id\": \"key-id\",\n  \"private_key\": \"-----BEGIN PRIVATE KEY-----\\nprivate-key\\n-----END PRIVATE KEY-----\\n\",\n  \"client_email\": \"service-account-email\",\n  \"client_id\": \"client-id\",\n  \"auth_uri\": \"https://accounts.google.com/o/oauth2/auth\",\n  \"token_uri\": \"https://accounts.google.com/o/oauth2/token\",\n  \"auth_provider_x509_cert_url\": \"https://www.googleapis.com/oauth2/v1/certs\",\n  \"client_x509_cert_url\": \"https://www.googleapis.com/robot/v1/metadata/x509/service-account-email\"\n}"),
					resource.TestCheckResourceAttr(secondResName, "as_file", "false"),
					resource.TestCheckResourceAttrSet(secondResName, "deployment_id"),
				),
			},
			{
				PlanOnly:                 true,
				ProtoV6ProviderFactories: testAccProviderFactory,
				Config:                   cfgF(migratedCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(firstResName, "setting_name", "xpack.notification.slack.account.hello.secure_url"),
					resource.TestCheckResourceAttr(firstResName, "value", "hella"),
					resource.TestCheckResourceAttr(firstResName, "as_file", "false"),
					resource.TestCheckResourceAttrSet(firstResName, "deployment_id"),

					resource.TestCheckResourceAttr(secondResName, "setting_name", "gcs.client.secondary.credentials_file"),
					resource.TestCheckResourceAttr(secondResName, "value", "{\n  \"type\": \"service_account\",\n  \"project_id\": \"project-id\",\n  \"private_key_id\": \"key-id\",\n  \"private_key\": \"-----BEGIN PRIVATE KEY-----\\nprivate-key\\n-----END PRIVATE KEY-----\\n\",\n  \"client_email\": \"service-account-email\",\n  \"client_id\": \"client-id\",\n  \"auth_uri\": \"https://accounts.google.com/o/oauth2/auth\",\n  \"token_uri\": \"https://accounts.google.com/o/oauth2/token\",\n  \"auth_provider_x509_cert_url\": \"https://www.googleapis.com/oauth2/v1/certs\",\n  \"client_x509_cert_url\": \"https://www.googleapis.com/robot/v1/metadata/x509/service-account-email\"\n}"),
					resource.TestCheckResourceAttr(secondResName, "as_file", "false"),
					resource.TestCheckResourceAttrSet(secondResName, "deployment_id"),
				),
			},
		},
	})
}

func checkESKeystoreResourceID(resourceName string, id *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		*id = rs.Primary.ID
		return nil
	}
}

func checkNoKeystoreResourcesLeft(resourceName ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		merr := multierror.NewPrefixed("found 'ec_deployment_elasticsearch_keystore' resources")
		for _, resName := range resourceName {
			if rs, ok := s.RootModule().Resources[resName]; ok {
				merr = merr.Append(fmt.Errorf("found: %s with ID: %s", resName, rs.Primary.ID))
			}
		}

		return merr.ErrorOrNil()
	}
}
