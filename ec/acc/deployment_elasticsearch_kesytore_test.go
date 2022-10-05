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

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/eskeystoreapi"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/elastic/cloud-sdk-go/pkg/util/slice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDeploymentElasticsearchKeystore_full(t *testing.T) {
	var previousID, currentID string

	resType := "ec_deployment_elasticsearch_keystore"
	deploymentResName := "ec_deployment.keystore"
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
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

					checkExpectedKeystoreKeysExist(deploymentResName, "xpack.notification.slack.account.hello.secure_url", "gcs.client.secondary.credentials_file"),
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

					checkExpectedKeystoreKeysExist(deploymentResName, "xpack.notification.slack.account.hello.secure_url", "gcs.client.secondary.credentials_file"),
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

					checkExpectedKeystoreKeysExist(deploymentResName, "xpack.notification.slack.account.hello.secure_urla", "gcs.client.secondary.credentials_file"),
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

func checkExpectedKeystoreKeysExist(deploymentResource string, expectedKeys ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := newAPI()
		if err != nil {
			return err
		}

		deployment, ok := s.RootModule().Resources[deploymentResource]
		if !ok {
			return fmt.Errorf("Not found: %s", deploymentResource)
		}

		deploymentID := deployment.Primary.ID

		keystoreContents, err := eskeystoreapi.Get(eskeystoreapi.GetParams{
			API:          client,
			DeploymentID: deploymentID,
		})
		if err != nil {
			return err
		}

		var missingKeys, extraKeys []string
		for _, expectedKey := range expectedKeys {
			if _, ok := keystoreContents.Secrets[expectedKey]; !ok {
				missingKeys = append(missingKeys, expectedKey)
			}
		}

		for key, _ := range keystoreContents.Secrets {
			if !slice.HasString(expectedKeys, key) {
				extraKeys = append(extraKeys, key)
			}
		}

		mErr := multierror.NewPrefixed("unexpected keystore contents")

		if len(missingKeys) > 0 {
			mErr.Append(fmt.Errorf("keys missing from the deployment keystore %v", missingKeys))
		}

		if len(extraKeys) > 0 {
			mErr.Append(fmt.Errorf("extra keys present in the deployment keystore: %v", extraKeys))
		}

		return mErr.ErrorOrNil()
	}
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
