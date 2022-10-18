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

package trafficfilterresource_test

import (
	"net/url"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	r "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"

	provider "github.com/elastic/terraform-provider-ec/ec"
)

func TestResourceTrafficFilter(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse("true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				updateResponse("false"),
				readResponse("false", "false"),
				readResponse("false", "false"),
				readResponse("false", "false"),
				readResponse("true", "false"),
				deleteResponse(),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: trafficFilter,
				Check:  checkResource("true"),
			},
			{ // Ensure that it can be successfully read
				PlanOnly: true,
				Config:   trafficFilter,
				Check:    checkResource("true"),
			},
			{ // Ensure that it can be successfully updated
				Config: trafficFilterWithoutIncludeByDefault,
				Check:  checkResource("false"),
			},
			{ // Delete resource
				Destroy: true,
				Config:  trafficFilter,
			},
		},
	})
}

func TestResourceTrafficFilterWithoutIncludeByDefault(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse("false"),
				readResponse("false", "false"),
				readResponse("false", "false"),
				readResponse("false", "false"),
				readResponse("false", "false"),
				readResponse("false", "false"),
				readResponse("true", "false"),
				deleteResponse(),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: trafficFilterWithoutIncludeByDefault,
				Check:  checkResource("false"),
			},
			{ // Ensure that it can be successfully read
				PlanOnly: true,
				Config:   trafficFilterWithoutIncludeByDefault,
				Check:    checkResource("false"),
			},
			{ // Delete resource
				Destroy: true,
				Config:  trafficFilterWithoutIncludeByDefault,
			},
		},
	})
}

func TestResourceTrafficFilter_failedRead1(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse("true"),
				readResponse("false", "true"),
				failedReadResponse("false"),
				notFoundReadResponse("true"), // required for cleanup
			),
		),
		Steps: []r.TestStep{
			{
				Config:      trafficFilter,
				ExpectError: regexp.MustCompile(`internal.server.error: There was an internal server error`),
			},
		},
	})
}

func TestResourceTrafficFilter_notFoundAfterUpdate(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse("true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				updateResponse("false"),
				notFoundReadResponse("false"),
				notFoundReadResponse("true"),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: trafficFilter,
				Check:  checkResource("true"),
			},
			{ // Update resource
				Config:      trafficFilterWithoutIncludeByDefault,
				ExpectError: regexp.MustCompile(`Failed to read deployment traffic filter ruleset after update.`),
			},
		},
	})
}

func TestResourceTrafficFilter_failedUpdate1(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse("true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				failedUpdateResponse("false"),
				notFoundReadResponse("true"), // required for cleanup
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: trafficFilter,
				Check:  checkResource("true"),
			},
			{ // Update resource
				Config:      trafficFilterWithoutIncludeByDefault,
				ExpectError: regexp.MustCompile(`internal.server.error: There was an internal server error`),
			},
		},
	})
}

func TestResourceTrafficFilter_failedUpdate2(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse("true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				updateResponse("false"),
				failedReadResponse("false"),
				notFoundReadResponse("true"), // required for cleanup
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: trafficFilter,
				Check:  checkResource("true"),
			},
			{ // Update resource
				Config:      trafficFilterWithoutIncludeByDefault,
				ExpectError: regexp.MustCompile(`internal.server.error: There was an internal server error`),
			},
		},
	})
}

func TestResourceTrafficFilterAssoc_gracefulDeletionOnRead(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse("true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				notFoundReadResponse("false"),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: trafficFilter,
				Check:  checkResource("true"),
			},
			{ // Ensure that it gets unset if deleted externally
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config:             trafficFilter,
				Check:              checkResourceDeleted(),
			},
		},
	})
}

func TestResourceTrafficFilter_gracefulDeletion1(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse("true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				readResponse("true", "true"),
				alreadyDeletedResponse(),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: trafficFilter,
				Check:  checkResource("true"),
			},
			{ // Delete resource
				Destroy: true,
				Config:  trafficFilter,
			},
		},
	})
}

func TestResourceTrafficFilter_gracefulDeletion2(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse("true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				notFoundReadResponse("true"),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: trafficFilter,
				Check:  checkResource("true"),
			},
			{ // Delete resource
				Destroy: true,
				Config:  trafficFilter,
			},
		},
	})
}

func TestResourceTrafficFilter_failedDeletion1(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse("true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				readResponse("true", "true"),
				failedDeletionResponse(),
				notFoundReadResponse("true"), // required for cleanup
			),
		),
		Steps: []r.TestStep{
			{
				Config: trafficFilter,
			},
			{
				Destroy:     true,
				Config:      trafficFilter,
				ExpectError: regexp.MustCompile(`internal.server.error: There was an internal server error`),
			},
		},
	})
}

func TestResourceTrafficFilter_failedDeletion2(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse("true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				failedReadResponse("true"),
				readResponse("true", "true"),
				deleteResponse(),
			),
		),
		Steps: []r.TestStep{
			{
				Config: trafficFilter,
			},
			{
				Destroy:     true,
				Config:      trafficFilter,
				ExpectError: regexp.MustCompile(`internal.server.error: There was an internal server error`),
			},
		},
	})
}

func TestResourceTrafficFilter_deletionWithUnknownAssociationError(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse("true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				mock.New200StructResponse(models.TrafficFilterRulesetInfo{
					Associations: []*models.FilterAssociation{
						{ID: ec.String("some id"), EntityType: ec.String("deployment")},
					},
				}),
				mock.NewErrorResponse(500, mock.APIError{
					Code: "some", Message: "message",
				}),
				readResponse("true", "true"),
				alreadyDeletedResponse(),
			),
		),
		Steps: []r.TestStep{
			{
				Config: trafficFilter,
			},
			{
				Destroy:     true,
				Config:      trafficFilter,
				ExpectError: regexp.MustCompile(`some: message`),
			},
		},
	})
}

func TestResourceTrafficFilter_deletionWithAssociationNotFound(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse("true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				readResponse("false", "true"),
				mock.New200StructResponse(models.TrafficFilterRulesetInfo{
					Associations: []*models.FilterAssociation{
						{ID: ec.String("some id"), EntityType: ec.String("deployment")},
					},
				}),
				mock.NewErrorResponse(404, mock.APIError{
					Code: "some", Message: "message",
				}),
				deleteResponse(),
			),
		),
		Steps: []r.TestStep{
			{
				Config: trafficFilter,
			},
			{
				Destroy: true,
				Config:  trafficFilter,
			},
		},
	})
}

func TestResourceTrafficFilter_importState(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				readResponse("false", "true"),
			),
		),
		Steps: []r.TestStep{
			{
				ImportState:   true,
				ImportStateId: "some-random-id",
				ResourceName:  "ec_deployment_traffic_filter.test1",
				Config:        trafficFilter,
				Check:         checkResource("true"),
			},
		},
	})
}

const trafficFilter = `
	resource "ec_deployment_traffic_filter" "test1" {
	  name   = "my traffic filter"
      description = "Allow access from 1.1.1.1 and 1.1.1.0/16"
	  region = "us-east-1"
	  type   = "ip"

	  include_by_default = true

	  rule {
		source = "1.1.1.1"
	  }
	  rule {
		source = "1.1.1.0/16"
	  }
	}
`
const trafficFilterWithoutIncludeByDefault = `
	resource "ec_deployment_traffic_filter" "test1" {
	  name   = "my traffic filter"
      description = "Allow access from 1.1.1.1 and 1.1.1.0/16"
	  region = "us-east-1"
	  type   = "ip"

	  rule {
		source = "1.1.1.1"
	  }
	  rule {
		source = "1.1.1.0/16"
	  }
	}
`

func checkResource(includeByDefault string) r.TestCheckFunc {
	resource := "ec_deployment_traffic_filter.test1"
	return r.ComposeAggregateTestCheckFunc(
		r.TestCheckResourceAttr(resource, "id", "some-random-id"),
		r.TestCheckResourceAttr(resource, "name", "my traffic filter"),
		r.TestCheckResourceAttr(resource, "description", "Allow access from 1.1.1.1 and 1.1.1.0/16"),
		r.TestCheckResourceAttr(resource, "region", "us-east-1"),
		r.TestCheckResourceAttr(resource, "type", "ip"),
		r.TestCheckResourceAttr(resource, "include_by_default", includeByDefault),
		r.TestCheckResourceAttr(resource, "rule.0.id", "some-random-rule-id-1"),
		r.TestCheckResourceAttr(resource, "rule.0.source", "1.1.1.1"),
		r.TestCheckResourceAttr(resource, "rule.1.id", "some-random-rule-id-2"),
		r.TestCheckResourceAttr(resource, "rule.1.source", "1.1.1.0/16"),
	)
}

func checkResourceDeleted() r.TestCheckFunc {
	resource := "ec_deployment_traffic_filter.test1"
	return r.ComposeAggregateTestCheckFunc(
		r.TestCheckNoResourceAttr(resource, "id"),
		r.TestCheckNoResourceAttr(resource, "name"),
		r.TestCheckNoResourceAttr(resource, "description"),
		r.TestCheckNoResourceAttr(resource, "region"),
		r.TestCheckNoResourceAttr(resource, "type"),
		r.TestCheckNoResourceAttr(resource, "include_by_default"),
		r.TestCheckNoResourceAttr(resource, "rule.0.id"),
		r.TestCheckNoResourceAttr(resource, "rule.0.source"),
		r.TestCheckNoResourceAttr(resource, "rule.1.id"),
		r.TestCheckNoResourceAttr(resource, "rule.1.source"),
	)
}

func createResponse(includeByDefault string) mock.Response {
	return mock.New201ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "POST",
			Path:   "/api/v1/deployments/traffic-filter/rulesets",
			Query:  url.Values{},
			Body:   mock.NewStringBody(`{"description":"Allow access from 1.1.1.1 and 1.1.1.0/16","include_by_default":` + includeByDefault + `,"name":"my traffic filter","region":"us-east-1","rules":[{"source":"1.1.1.0/16"},{"source":"1.1.1.1"}],"type":"ip"}` + "\n"),
		},
		mock.NewStringBody(`{"id" : "some-random-id"}`),
	)
}

func updateResponse(includeByDefault string) mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "PUT",
			Path:   "/api/v1/deployments/traffic-filter/rulesets/some-random-id",
			Query:  url.Values{},
			Body:   mock.NewStringBody(`{"description":"Allow access from 1.1.1.1 and 1.1.1.0/16","include_by_default":` + includeByDefault + `,"name":"my traffic filter","region":"us-east-1","rules":[{"id":"some-random-rule-id-1","source":"1.1.1.1"},{"id":"some-random-rule-id-2","source":"1.1.1.0/16"}],"type":"ip"}` + "\n"),
		},
		mock.NewStringBody(`{"id" : "some-random-id"}`),
	)
}

func failedUpdateResponse(includeByDefault string) mock.Response {
	return mock.New500ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "PUT",
			Path:   "/api/v1/deployments/traffic-filter/rulesets/some-random-id",
			Query:  url.Values{},
			Body:   mock.NewStringBody(`{"description":"Allow access from 1.1.1.1 and 1.1.1.0/16","include_by_default":` + includeByDefault + `,"name":"my traffic filter","region":"us-east-1","rules":[{"id":"some-random-rule-id-1","source":"1.1.1.1"},{"id":"some-random-rule-id-2","source":"1.1.1.0/16"}],"type":"ip"}` + "\n"),
		},
		mock.SampleInternalError().Response.Body,
	)
}

func readResponse(includeAssociations string, includeByDefault string) mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "GET",
			Path:   "/api/v1/deployments/traffic-filter/rulesets/some-random-id",
			Query: url.Values{
				"include_associations": []string{includeAssociations},
			},
		},
		mock.NewStringBody(`{
							  "id" : "some-random-id",
							  "name" : "my traffic filter",
							  "description" : "Allow access from 1.1.1.1 and 1.1.1.0/16",
							  "type": "ip",
							  "include_by_default": `+includeByDefault+`,
							  "region": "us-east-1",
	                          "rules": [
								{
								  "id" : "some-random-rule-id-1",
								  "source" : "1.1.1.1"
								},
								{
								  "id" : "some-random-rule-id-2",
								  "source" : "1.1.1.0/16"
								}
							  ]
							}`,
		),
	)
}

func notFoundReadResponse(includeAssociations string) mock.Response {
	return mock.New404ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "GET",
			Path:   "/api/v1/deployments/traffic-filter/rulesets/some-random-id",
			Query: url.Values{
				"include_associations": []string{includeAssociations},
			},
		},
		mock.NewStringBody(`{	}`),
	)
}
func failedReadResponse(includeAssociations string) mock.Response {
	return mock.New500ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "GET",
			Path:   "/api/v1/deployments/traffic-filter/rulesets/some-random-id",
			Query: url.Values{
				"include_associations": []string{includeAssociations},
			},
		},
		mock.SampleInternalError().Response.Body,
	)
}

func deleteResponse() mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "DELETE",
			Path:   "/api/v1/deployments/traffic-filter/rulesets/some-random-id",
			Query: url.Values{
				"ignore_associations": []string{"false"},
			},
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
			Path:   "/api/v1/deployments/traffic-filter/rulesets/some-random-id",
			Query: url.Values{
				"ignore_associations": []string{"false"},
			},
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
			Path:   "/api/v1/deployments/traffic-filter/rulesets/some-random-id",
			Query: url.Values{
				"ignore_associations": []string{"false"},
			},
		},
		mock.SampleInternalError().Response.Body,
	)
}

func protoV6ProviderFactoriesWithMockClient(client *api.API) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"ec": func() (tfprotov6.ProviderServer, error) {
			return providerserver.NewProtocol6(provider.ProviderWithClient(client, "unit-tests"))(), nil
		},
	}
}
