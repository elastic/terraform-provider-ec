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

package extensionresource_test

import (
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	r "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"

	provider "github.com/elastic/terraform-provider-ec/ec"
)

func TestResourceDeploymentExtension(t *testing.T) {

	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(),
				readResponse1(),
				readResponse1(),
				readResponse1(),
				updateResponse(),

				// Not testing for assertion as the content type is multipart/form-data
				// with a boundary that is a randomly generated string which changes every time.
				mock.Response{
					Response: http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       mock.NewStringBody("{}"),
					},
				},

				readResponse2(),
				readResponse2(),
				readResponse2(),
				deleteResponse(),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: deploymentExtension1,
				Check:  checkResource1(),
			},
			{ // Update resource
				Config: deploymentExtension2,
				Check:  checkResource2(),
			},
			{ // Delete resource
				Destroy: true,
				Config:  deploymentExtension2,
			},
		},
	})
}

func TestResourceDeploymentExtension_failedCreate(t *testing.T) {

	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				mock.New500Response(mock.SampleInternalError().Response.Body),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config:      deploymentExtension1,
				ExpectError: regexp.MustCompile(`internal.server.error: There was an internal server error`),
			},
		},
	})
}

func TestResourceDeploymentExtension_failedReadAfterCreate(t *testing.T) {

	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(),
				mock.New500Response(mock.SampleInternalError().Response.Body),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config:      deploymentExtension1,
				ExpectError: regexp.MustCompile(`internal.server.error: There was an internal server error`),
			},
		},
	})
}

func TestResourceDeploymentExtension_notFoundAfterCreate(t *testing.T) {

	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(),
				mock.New404Response(mock.NewStringBody(`{	}`)),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config:      deploymentExtension1,
				ExpectError: regexp.MustCompile(`Failed to read deployment extension after create.`),
			},
		},
	})
}

func TestResourceDeploymentExtension_failedUpdate(t *testing.T) {

	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(),
				readResponse1(),
				readResponse1(),
				readResponse1(),
				mock.New500Response(mock.SampleInternalError().Response.Body),
				deleteResponse(),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: deploymentExtension1,
				Check:  checkResource1(),
			},
			{ // Update resource
				Config:      deploymentExtension2,
				Check:       checkResource2(),
				ExpectError: regexp.MustCompile(`internal.server.error: There was an internal server error`),
			},
		},
	})
}

func TestResourceDeploymentExtension_notFoundAfterUpdate(t *testing.T) {

	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(),
				readResponse1(),
				readResponse1(),
				readResponse1(),
				updateResponse(),
				mock.Response{
					Response: http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       mock.NewStringBody("{}"),
					},
				},
				mock.New404Response(mock.NewStringBody(`{	}`)),
				deleteResponse(),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: deploymentExtension1,
				Check:  checkResource1(),
			},
			{ // Update resource
				Config:      deploymentExtension2,
				Check:       checkResource2(),
				ExpectError: regexp.MustCompile(`Failed to read deployment extension after update.`),
			},
		},
	})
}

func TestResourceDeploymentExtension_failedDelete(t *testing.T) {

	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(),
				readResponse1(),
				readResponse1(),
				readResponse1(),
				mock.New500Response(mock.SampleInternalError().Response.Body),
				deleteResponse(),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: deploymentExtension1,
				Check:  checkResource1(),
			},
			{ // Delete resource
				Destroy:     true,
				Config:      deploymentExtension2,
				ExpectError: regexp.MustCompile(`internal.server.error: There was an internal server error`),
			},
		},
	})
}

func TestResourceDeploymentExtension_gracefulDeletion(t *testing.T) {

	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(),
				readResponse1(),
				readResponse1(),
				readResponse1(),
				mock.New404ResponseAssertion(
					&mock.RequestAssertion{
						Header: api.DefaultReadMockHeaders,
						Method: "DELETE",
						Host:   api.DefaultMockHost,
						Path:   "/api/v1/deployments/extensions/someid",
					},
					mock.NewStructBody(models.Extension{
						ID: ec.String("{ }"),
					},
					),
				),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: deploymentExtension1,
				Check:  checkResource1(),
			},
			{ // Delete resource
				Destroy: true,
				Config:  deploymentExtension2,
			},
		},
	})
}

const deploymentExtension1 = `
resource "ec_deployment_extension" "my_extension" {
  name           = "My extension"
  description    = "Some description"
  version        = "*"
  extension_type = "bundle"
}
`
const deploymentExtension2 = `
resource "ec_deployment_extension" "my_extension" {
  name           = "My updated extension"
  description    = "Some updated description"
  version        = "7.10.1"
  extension_type = "bundle"
  download_url   = "https://example.com"
  file_path      = "testdata/test_extension_bundle.json"
  file_hash      = "abcd"
}
`

func checkResource1() r.TestCheckFunc {
	resource := "ec_deployment_extension.my_extension"
	return r.ComposeAggregateTestCheckFunc(
		r.TestCheckResourceAttr(resource, "id", "someid"),
		r.TestCheckResourceAttr(resource, "name", "My extension"),
		r.TestCheckResourceAttr(resource, "description", "Some description"),
		r.TestCheckResourceAttr(resource, "version", "*"),
		r.TestCheckResourceAttr(resource, "extension_type", "bundle"),
	)
}

func checkResource2() r.TestCheckFunc {
	resource := "ec_deployment_extension.my_extension"
	return r.ComposeAggregateTestCheckFunc(
		r.TestCheckResourceAttr(resource, "id", "someid"),
		r.TestCheckResourceAttr(resource, "name", "My updated extension"),
		r.TestCheckResourceAttr(resource, "description", "Some updated description"),
		r.TestCheckResourceAttr(resource, "version", "7.10.1"),
		r.TestCheckResourceAttr(resource, "extension_type", "bundle"),
		r.TestCheckResourceAttr(resource, "download_url", "https://example.com"),
		r.TestCheckResourceAttr(resource, "url", "repo://1234"),
		r.TestCheckResourceAttr(resource, "last_modified", "2021-01-07T22:13:42.999Z"),
		r.TestCheckResourceAttr(resource, "size", "1000"),
		r.TestCheckResourceAttr(resource, "file_path", "testdata/test_extension_bundle.json"),
		r.TestCheckResourceAttr(resource, "file_hash", "abcd"),
	)
}

func createResponse() mock.Response {
	return mock.New201ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "POST",
			Path:   "/api/v1/deployments/extensions",
			Query:  url.Values{},
			Body:   mock.NewStringBody(`{"description":"Some description","extension_type":"bundle","name":"My extension","version":"*"}` + "\n"),
		},
		mock.NewStringBody(`{"deployments":null,"description":"Some description","download_url":null,"extension_type":"bundle","id":"someid","name":"My extension","url":null,"version":"*"}`),
	)
}

func updateResponse() mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "POST",
			Path:   "/api/v1/deployments/extensions/someid",
			Query:  url.Values{},
			Body:   mock.NewStringBody(`{"description":"Some updated description","download_url":"https://example.com","extension_type":"bundle","name":"My updated extension","version":"7.10.1"}` + "\n"),
		},
		mock.NewStructBody(models.Extension{
			ID:            ec.String("someid"),
			Name:          ec.String("My updated extension"),
			Description:   "Some updated description",
			ExtensionType: ec.String("bundle"),
			Version:       ec.String("7.10.1"),
			DownloadURL:   "https://example.com",
			URL:           ec.String("repo://1234"),
			FileMetadata: &models.ExtensionFileMetadata{
				LastModifiedDate: lastModified(),
				Size:             1000,
			},
		}))
}

func deleteResponse() mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Header: api.DefaultReadMockHeaders,
			Method: "DELETE",
			Host:   api.DefaultMockHost,
			Path:   "/api/v1/deployments/extensions/someid",
		},
		mock.NewStructBody(models.Extension{
			ID: ec.String("someid"),
		}),
	)
}

func readResponse1() mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Header: api.DefaultReadMockHeaders,
			Method: "GET",
			Host:   api.DefaultMockHost,
			Path:   "/api/v1/deployments/extensions/someid",
			Query:  url.Values{"include_deployments": {"false"}},
		},
		mock.NewStructBody(models.Extension{
			ID:            ec.String("someid"),
			Name:          ec.String("My extension"),
			Description:   "Some description",
			ExtensionType: ec.String("bundle"),
			Version:       ec.String("*"),
		}),
	)
}
func readResponse2() mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Header: api.DefaultReadMockHeaders,
			Method: "GET",
			Host:   api.DefaultMockHost,
			Path:   "/api/v1/deployments/extensions/someid",
			Query:  url.Values{"include_deployments": {"false"}},
		},
		mock.NewStructBody(models.Extension{
			ID:            ec.String("someid"),
			Name:          ec.String("My updated extension"),
			Description:   "Some updated description",
			ExtensionType: ec.String("bundle"),
			Version:       ec.String("7.10.1"),
			DownloadURL:   "https://example.com",
			URL:           ec.String("repo://1234"),
			FileMetadata: &models.ExtensionFileMetadata{
				LastModifiedDate: lastModified(),
				Size:             1000,
			},
		}),
	)
}

func lastModified() strfmt.DateTime {
	lastModified, _ := strfmt.ParseDateTime("2021-01-07T22:13:42.999Z")
	return lastModified
}

func protoV6ProviderFactoriesWithMockClient(client *api.API) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"ec": func() (tfprotov6.ProviderServer, error) {
			return providerserver.NewProtocol6(provider.ProviderWithClient(client, "unit-tests"))(), nil
		},
	}
}
