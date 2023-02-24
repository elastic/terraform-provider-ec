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

package elasticsearchkeystoreresource_test

import (
	"net/url"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	r "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"

	"github.com/elastic/terraform-provider-ec/ec"
)

func TestResourceElasticsearchKeyStore(t *testing.T) {

	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				readDeployment(),
				createResponse(),
				readDeployment(),
				readResponse(),
				readDeployment(),
				readResponse(),
				readDeployment(),
				readResponse(),
				readDeployment(),
				updateResponse(),
				readDeployment(),
				readResponse(),
				readDeployment(),
				readResponse(),
				readDeployment(),
				readResponse(),
				readDeployment(),
				deleteResponse(),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: externalKeystore1,
				Check:  checkResource1(),
			},
			{ // Update resource
				Config: externalKeystore2,
				Check:  checkResource2(),
			},
			{ // Delete resource
				Destroy: true,
				Config:  externalKeystore1,
			},
		},
	})
}

func TestResourceElasticsearchKeyStore_failedCreate(t *testing.T) {

	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				mock.New500Response(mock.SampleInternalError().Response.Body),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config:      externalKeystore1,
				ExpectError: regexp.MustCompile(`internal.server.error: There was an internal server error`),
			},
		},
	})
}

func TestResourceElasticsearchKeyStore_failedReadAfterCreate(t *testing.T) {

	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				readDeployment(),
				createResponse(),
				mock.New500Response(mock.SampleInternalError().Response.Body),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config:      externalKeystore1,
				ExpectError: regexp.MustCompile(`internal.server.error: There was an internal server error`),
			},
		},
	})
}

func TestResourceElasticsearchKeyStore_notFoundAfterCreate_and_gracefulDeletion(t *testing.T) {

	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				readDeployment(),
				createResponse(),
				readDeployment(),
				emptyReadResponse(),
				readDeployment(),
				emptyReadResponse(),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config:      externalKeystore1,
				Check:       checkResource1(),
				ExpectError: regexp.MustCompile(`Failed to read Elasticsearch keystore after create.`),
			},
		},
	})
}

const externalKeystore1 = `
resource "ec_deployment_elasticsearch_keystore" "test" {
  deployment_id = "0a592ab2c5baf0fa95c77ac62135782e"
  setting_name  = "xpack.notification.slack.account.hello.secure_url"
  value         = "hella"
}
`

const externalKeystore2 = `
resource "ec_deployment_elasticsearch_keystore" "test" {
  deployment_id = "0a592ab2c5baf0fa95c77ac62135782e"
  setting_name  = "xpack.notification.slack.account.hello.secure_url"
  value         = <<EOT
{
  "type": "service_account",
  "project_id": "project-id",
  "private_key_id": "key-id",
  "private_key": "-----BEGIN PRIVATE KEY-----\nprivate-key\n-----END PRIVATE KEY-----\n",
  "client_email": "service-account-email",
  "client_id": "client-id",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://accounts.google.com/o/oauth2/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/service-account-email"
}
EOT
}
`

func checkResource1() r.TestCheckFunc {
	resource := "ec_deployment_elasticsearch_keystore.test"
	return r.ComposeAggregateTestCheckFunc(
		r.TestCheckResourceAttr(resource, "id", "1982613755"),
		r.TestCheckResourceAttr(resource, "deployment_id", "0a592ab2c5baf0fa95c77ac62135782e"),
		r.TestCheckResourceAttr(resource, "setting_name", "xpack.notification.slack.account.hello.secure_url"),
		r.TestCheckResourceAttr(resource, "value", "hella"),
	)
}
func checkResource2() r.TestCheckFunc {
	resource := "ec_deployment_elasticsearch_keystore.test"
	return r.ComposeAggregateTestCheckFunc(
		r.TestCheckResourceAttr(resource, "id", "1982613755"),
		r.TestCheckResourceAttr(resource, "deployment_id", "0a592ab2c5baf0fa95c77ac62135782e"),
		r.TestCheckResourceAttr(resource, "setting_name", "xpack.notification.slack.account.hello.secure_url"),
		r.TestCheckResourceAttr(resource, "value", "{\n  \"type\": \"service_account\",\n  \"project_id\": \"project-id\",\n  \"private_key_id\": \"key-id\",\n  \"private_key\": \"-----BEGIN PRIVATE KEY-----\\nprivate-key\\n-----END PRIVATE KEY-----\\n\",\n  \"client_email\": \"service-account-email\",\n  \"client_id\": \"client-id\",\n  \"auth_uri\": \"https://accounts.google.com/o/oauth2/auth\",\n  \"token_uri\": \"https://accounts.google.com/o/oauth2/token\",\n  \"auth_provider_x509_cert_url\": \"https://www.googleapis.com/oauth2/v1/certs\",\n  \"client_x509_cert_url\": \"https://www.googleapis.com/robot/v1/metadata/x509/service-account-email\"\n}\n"),
	)
}

func readDeployment() mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "GET",
			Path:   "/api/v1/deployments/0a592ab2c5baf0fa95c77ac62135782e",
			Query: url.Values{
				"convert_legacy_plans": []string{"false"},
				"show_metadata":        []string{"false"},
				"show_plan_defaults":   []string{"false"},
				"show_plan_history":    []string{"false"},
				"show_plan_logs":       []string{"false"},
				"show_plans":           []string{"false"},
				"show_settings":        []string{"false"},
				"show_system_alerts":   []string{"5"},
			},
		},
		mock.NewStringBody(`{
  "id" : "0a592ab2c5baf0fa95c77ac62135782e",
  "name" : "test",
  "healthy" : true,
  "resources" : {
    "elasticsearch" : [
      {
        "ref_id" : "main-elasticsearch",
        "id" : "fcf90600779c45008d81364b747a4ff5"
      }
	]
  }
}`,
		),
	)
}
func readResponse() mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "GET",
			Path:   "/api/v1/deployments/0a592ab2c5baf0fa95c77ac62135782e/elasticsearch/main-elasticsearch/keystore",
			Query:  url.Values{},
		},
		mock.NewStringBody(`{
   "secrets" : {
      "xpack.notification.slack.account.hello.secure_url" : {
         "value" : {}
      }
   }
}`,
		),
	)
}

func emptyReadResponse() mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "GET",
			Path:   "/api/v1/deployments/0a592ab2c5baf0fa95c77ac62135782e/elasticsearch/main-elasticsearch/keystore",
			Query:  url.Values{},
		},
		mock.NewStringBody(`{
   "secrets" : {
   }
}`,
		),
	)
}

func createResponse() mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "PATCH",
			Path:   "/api/v1/deployments/0a592ab2c5baf0fa95c77ac62135782e/elasticsearch/main-elasticsearch/keystore",
			Query:  url.Values{},
			Body:   mock.NewStringBody(`{"secrets":{"xpack.notification.slack.account.hello.secure_url":{"as_file":false,"value":"hella"}}}` + "\n"),
		},
		mock.NewStringBody(`{}`),
	)
}
func updateResponse() mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "PATCH",
			Path:   "/api/v1/deployments/0a592ab2c5baf0fa95c77ac62135782e/elasticsearch/main-elasticsearch/keystore",
			Query:  url.Values{},
			Body:   mock.NewStringBody(`{"secrets":{"xpack.notification.slack.account.hello.secure_url":{"as_file":false,"value":{"auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs","auth_uri":"https://accounts.google.com/o/oauth2/auth","client_email":"service-account-email","client_id":"client-id","client_x509_cert_url":"https://www.googleapis.com/robot/v1/metadata/x509/service-account-email","private_key":"-----BEGIN PRIVATE KEY-----\nprivate-key\n-----END PRIVATE KEY-----\n","private_key_id":"key-id","project_id":"project-id","token_uri":"https://accounts.google.com/o/oauth2/token","type":"service_account"}}}}` + "\n"),
		},
		mock.NewStringBody(`{}`),
	)
}

func deleteResponse() mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "PATCH",
			Path:   "/api/v1/deployments/0a592ab2c5baf0fa95c77ac62135782e/elasticsearch/main-elasticsearch/keystore",
			Query:  url.Values{},
			Body:   mock.NewStringBody(`{"secrets":{"xpack.notification.slack.account.hello.secure_url":{"as_file":false,"value":""}}}` + "\n"),
		},
		mock.NewStringBody(`{}`),
	)
}

func protoV6ProviderFactoriesWithMockClient(client *api.API) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"ec": func() (tfprotov6.ProviderServer, error) {
			return providerserver.NewProtocol6(ec.ProviderWithClient(client, "unit-tests"))(), nil
		},
	}
}
