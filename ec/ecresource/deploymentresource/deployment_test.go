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

package deploymentresource_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	r "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"

	provider "github.com/elastic/terraform-provider-ec/ec"
)

func Test_createDeploymentWithEmptyFields(t *testing.T) {
	requestId := "cuchxqanal0g8rmx9ljog7qrrpd68iitulaz2mrch1vuuihetgo5ge3f6555vn4s"

	deploymentWithDefaultsIoOptimized := fmt.Sprintf(`
		resource "ec_deployment" "empty-declarations-IO-Optimized" {
			request_id = "%s"
			name = "my_deployment_name"
			deployment_template_id = "aws-io-optimized-v2"
			region = "us-east-1"
			version = "8.4.3"

			elasticsearch = {
				config = {}
				hot = {
					size = "8g"
					autoscaling = {}
				}
			}
		}`,
		requestId,
	)

	createDeploymentResponseJson := []byte(`
	{
		"alias": "my-deployment-name", 
		"created": true, 
		"id": "accd2e61fa835a5a32bb6b2938ce91f3", 
		"resources": [
			{
				"kind": "elasticsearch", 
				"cloud_id": "my_deployment_name:cloud_id", 
				"region": "us-east-1", 
				"ref_id": "main-elasticsearch", 
				"credentials": {
					"username": "elastic", 
					"password": "password"
				}, 
				"id": "resource_id"
			}
		], 
		"name": "my_deployment_name"
	}	
	`)

	templateFileName := "testdata/aws-io-optimized-v2.json"

	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				getTemplate(t, templateFileName),
				createDeployment(t, readFile(t, "testdata/aws-io-optimized-v2-empty-config-create-expected-payload.json"), createDeploymentResponseJson, requestId),
				mock.New200Response(readTestData(t, "testdata/aws-io-optimized-v2-empty-config-expected-deployment1.json")),
				mock.New200Response(readTestData(t, "testdata/aws-io-optimized-v2-empty-config-expected-deployment2.json")),
				mock.New200Response(readTestData(t, "testdata/aws-io-optimized-v2-empty-config-expected-deployment3.json")),
				mock.New200Response(readTestData(t, "testdata/aws-io-optimized-v2-empty-config-expected-deployment3.json")),
				mock.New200Response(readTestData(t, "testdata/aws-io-optimized-v2-empty-config-expected-deployment3.json")),
				mock.New202Response(io.NopCloser(strings.NewReader(""))),
				mock.New200Response(readTestData(t, "testdata/aws-io-optimized-v2-empty-config-expected-deployment3.json")),
				readRemoteClusters(t),
				mock.New200Response(readTestData(t, "testdata/aws-io-optimized-v2-empty-config-expected-deployment3.json")),
				readRemoteClusters(t),
				shutdownDeployment(t),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: deploymentWithDefaultsIoOptimized,
			},
		},
	})
}

func getTemplate(t *testing.T, filename string) mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "GET",
			Path:   "/api/v1/deployments/templates/aws-io-optimized-v2",
			Query:  url.Values{"region": {"us-east-1"}, "show_instance_configurations": {"false"}},
		},
		readTestData(t, filename),
	)
}

func readFile(t *testing.T, fileName string) []byte {
	t.Helper()
	res, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatalf(err.Error())
	}
	return res
}

func readTestData(t *testing.T, filename string) io.ReadCloser {
	t.Helper()
	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf(err.Error())
	}
	return f
}

func createDeployment(t *testing.T, expectedRequestJson, responseJson []byte, requestId string) mock.Response {
	t.Helper()
	var expectedRequest *models.DeploymentCreateRequest
	err := json.Unmarshal(expectedRequestJson, &expectedRequest)
	if err != nil {
		t.Fatalf(err.Error())
	}

	var response *models.DeploymentCreateResponse
	err = json.Unmarshal(responseJson, &response)
	if err != nil {
		t.Fatalf(err.Error())
	}

	return mock.New201ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "POST",
			Path:   "/api/v1/deployments",
			Query:  url.Values{"request_id": {requestId}},
			Body:   mock.NewStructBody(expectedRequest),
		},
		mock.NewStructBody(response),
	)
}

func shutdownDeployment(t *testing.T) mock.Response {
	t.Helper()

	return mock.New201ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "POST",
			Path:   "/api/v1/deployments/accd2e61fa835a5a32bb6b2938ce91f3/_shutdown",
			Query:  url.Values{"skip_snapshot": {"false"}},
			Body:   io.NopCloser(strings.NewReader("")),
		},
		io.NopCloser(strings.NewReader("")),
	)
}

func readRemoteClusters(t *testing.T) mock.Response {

	return mock.New200StructResponse(
		&models.RemoteResources{Resources: []*models.RemoteResourceRef{}},
	)
}

func protoV6ProviderFactoriesWithMockClient(client *api.API) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"ec": providerserver.NewProtocol6WithError(provider.ProviderWithClient(client, "unit-tests")),
	}
}
