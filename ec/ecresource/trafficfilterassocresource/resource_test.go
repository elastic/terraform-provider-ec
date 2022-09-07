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

package trafficfilterassocresource_test

import (
	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/terraform-provider-ec/ec"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"net/url"
	"regexp"
	"testing"

	r "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceTrafficFilterAssoc(t *testing.T) {

	r.UnitTest(t, r.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(),
				readResponse(),
				readResponse(),
				readResponse(),
				readResponse(),
				deleteResponse(),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: trafficFilterAssoc,
				Check:  checkResource(),
			},
			{ // Ensure that it can be successfully read
				PlanOnly: true,
				Config:   trafficFilterAssoc,
				Check:    checkResource(),
			},
			{ // Delete resource
				Destroy: true,
				Config:  trafficFilterAssoc,
			},
		},
	})
}

func TestResourceTrafficFilterAssoc_externalDeletion1(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(),
				readResponse(),
				readResponseAssociationDeleted(),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: trafficFilterAssoc,
				Check:  checkResource(),
			},
			{ // Ensure that it gets unset if deleted externally
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config:             trafficFilterAssoc,
				Check:              checkResourceDeleted(),
			},
		},
	})
}
func TestResourceTrafficFilterAssoc_externalDeletion2(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(),
				readResponse(),
				readResponseTrafficFilterDeleted(),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: trafficFilterAssoc,
				Check:  checkResource(),
			},
			{ // Ensure that it gets unset if deleted externally
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config:             trafficFilterAssoc,
				Check:              checkResourceDeleted(),
			},
		},
	})
}

func TestResourceTrafficFilterAssoc_gracefulDeletion(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(),
				readResponse(),
				readResponse(),
				alreadyDeletedResponse(),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: trafficFilterAssoc,
				Check:  checkResource(),
			},
			{ // Delete resource
				Destroy: true,
				Config:  trafficFilterAssoc,
			},
		},
	})
}

func TestResourceTrafficFilterAssoc_failedDeletion(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(),
				readResponse(),
				readResponse(),
				failedDeletionResponse(),
				deleteResponse(),
			),
		),
		Steps: []r.TestStep{
			{
				Config: trafficFilterAssoc,
			},
			{
				Destroy:     true,
				Config:      trafficFilterAssoc,
				ExpectError: regexp.MustCompile(`internal.server.error: There was an internal server error`),
			},
		},
	})
}

func TestResourceTrafficFilterAssoc_importState(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactoriesWithMockClient(
			api.NewMock(
				readResponse(),
			),
		),
		Steps: []r.TestStep{
			{
				ImportState:   true,
				ImportStateId: "0a592ab2c5baf0fa95c77ac62135782e,9db94e68e2f040a19dfb664d0e83bc2a",
				ResourceName:  "ec_deployment_traffic_filter_association.test1",
				Config:        trafficFilterAssoc,
				Check:         checkResource(),
			},
		},
	})
}

const trafficFilterAssoc = `
	resource "ec_deployment_traffic_filter_association" "test1" {
	  traffic_filter_id = "9db94e68e2f040a19dfb664d0e83bc2a"
	  deployment_id     = "0a592ab2c5baf0fa95c77ac62135782e"
	}
`

func checkResource() r.TestCheckFunc {
	return r.ComposeAggregateTestCheckFunc(
		r.TestCheckResourceAttr("ec_deployment_traffic_filter_association.test1", "id", "0a592ab2c5baf0fa95c77ac62135782e-9db94e68e2f040a19dfb664d0e83bc2a"),
		r.TestCheckResourceAttr("ec_deployment_traffic_filter_association.test1", "traffic_filter_id", "9db94e68e2f040a19dfb664d0e83bc2a"),
		r.TestCheckResourceAttr("ec_deployment_traffic_filter_association.test1", "deployment_id", "0a592ab2c5baf0fa95c77ac62135782e"),
	)
}

func checkResourceDeleted() r.TestCheckFunc {
	return r.ComposeAggregateTestCheckFunc(
		r.TestCheckNoResourceAttr("ec_deployment_traffic_filter_association.test1", "id"),
		r.TestCheckNoResourceAttr("ec_deployment_traffic_filter_association.test1", "traffic_filter_id"),
		r.TestCheckNoResourceAttr("ec_deployment_traffic_filter_association.test1", "deployment_id"),
	)
}

func createResponse() mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "POST",
			Path:   "/api/v1/deployments/traffic-filter/rulesets/9db94e68e2f040a19dfb664d0e83bc2a/associations",
			Query:  url.Values{},
			Body:   mock.NewStringBody(`{"entity_type":"deployment","id":"0a592ab2c5baf0fa95c77ac62135782e"}` + "\n"),
		},
		mock.NewStringBody(`{"entity_type":"deployment","id":"0a592ab2c5baf0fa95c77ac62135782e"}`),
	)
}

func readResponse() mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "GET",
			Path:   "/api/v1/deployments/traffic-filter/rulesets/9db94e68e2f040a19dfb664d0e83bc2a",
			Query: url.Values{
				"include_associations": []string{"true"},
			},
		},
		mock.NewStringBody(`{
							"id": "9db94e68e2f040a19dfb664d0e83bc2a", 
							"name": "dummy", 
							"type": "ip",  
							"include_by_default": false,  
							"region": "us-east-1",  
							"rules": [{"id": "6e4c8874f90d4793a2290f8199461952","source": "127.0.0.1"}  ],
							"associations": [{"entity_type": "deployment", "id": "0a592ab2c5baf0fa95c77ac62135782e"}],
							"total_associations": 1
						}`,
		),
	)
}

func readResponseAssociationDeleted() mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "GET",
			Path:   "/api/v1/deployments/traffic-filter/rulesets/9db94e68e2f040a19dfb664d0e83bc2a",
			Query: url.Values{
				"include_associations": []string{"true"},
			},
		},
		mock.NewStringBody(`{
							"id": "9db94e68e2f040a19dfb664d0e83bc2a", 
							"name": "dummy", 
							"type": "ip",  
							"include_by_default": false,  
							"region": "us-east-1",  
							"rules": [{"id": "6e4c8874f90d4793a2290f8199461952","source": "127.0.0.1"}  ],
							"associations": [{"entity_type": "deployment", "id": "some-unrelated-id"}],
							"total_associations": 1
						}`,
		),
	)
}

func readResponseTrafficFilterDeleted() mock.Response {
	return mock.New404ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "GET",
			Path:   "/api/v1/deployments/traffic-filter/rulesets/9db94e68e2f040a19dfb664d0e83bc2a",
			Query: url.Values{
				"include_associations": []string{"true"},
			},
		},
		mock.NewStringBody(`{	}`),
	)
}

func deleteResponse() mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "DELETE",
			Path:   "/api/v1/deployments/traffic-filter/rulesets/9db94e68e2f040a19dfb664d0e83bc2a/associations/deployment/0a592ab2c5baf0fa95c77ac62135782e",
			Query:  url.Values{},
		},
		mock.NewStringBody(`{}`),
	)
}

func alreadyDeletedResponse() mock.Response {
	return mock.New404ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "DELETE",
			Path:   "/api/v1/deployments/traffic-filter/rulesets/9db94e68e2f040a19dfb664d0e83bc2a/associations/deployment/0a592ab2c5baf0fa95c77ac62135782e",
			Query:  url.Values{},
		},
		mock.NewStringBody(`{	}`),
	)
}
func failedDeletionResponse() mock.Response {
	mock.SampleInternalError()
	return mock.New500ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "DELETE",
			Path:   "/api/v1/deployments/traffic-filter/rulesets/9db94e68e2f040a19dfb664d0e83bc2a/associations/deployment/0a592ab2c5baf0fa95c77ac62135782e",
			Query:  url.Values{},
		},
		mock.SampleInternalError().Response.Body,
	)
}

func protoV5ProviderFactoriesWithMockClient(client *api.API) map[string]func() (tfprotov5.ProviderServer, error) {
	return map[string]func() (tfprotov5.ProviderServer, error){
		"ec": func() (tfprotov5.ProviderServer, error) {
			return providerserver.NewProtocol5(ec.ProviderWithClient(client, "unit-tests"))(), nil
		},
	}
}
