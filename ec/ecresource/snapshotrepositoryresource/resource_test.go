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

package snapshotrepositoryresource_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	r "github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	provider "github.com/elastic/terraform-provider-ec/ec"
)

func TestResourceSnapshotRepository(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(s3Json1),
				readResponse(s3Json1),
				readResponse(s3Json1),
				readResponse(s3Json1),
				readResponse(s3Json1),
				readResponse(s3Json1),
				updateResponse(s3Json2),
				readResponse(s3Json2),
				readResponse(s3Json2),
				readResponse(s3Json2WithPathStyleAccess),
				readResponse(s3Json2WithPathStyleAccess),
				updateResponse(genericJson),
				readResponse(genericJson),
				readResponse(genericJson),
				readResponse(genericJson),
				readResponse(genericJson),
				deleteResponse(),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: awsSnapshotRepository1,
				Check:  checkS3Resource1(),
			},
			{ // Ensure that it can be successfully read
				PlanOnly: true,
				Config:   awsSnapshotRepository1,
				Check:    checkS3Resource1(),
			},
			{ // Ensure that it can be properly imported
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "ec_snapshot_repository.this",
			},
			{ // Ensure that it can be successfully updated
				Config: awsSnapshotRepository2,
				Check:  checkS3Resource2(),
			},
			{ // Ensure that it can be properly imported
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "ec_snapshot_repository.this",
			},
			{ // Ensure that generic repositories work too
				Config: genericSnapshotRepository,
				Check:  checkGenericResource(),
			},
			{ // Ensure that it can be properly imported
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "ec_snapshot_repository.this",
			},
			{ // Delete resource
				Destroy: true,
				Config:  genericSnapshotRepository,
			},
		},
	})
}

func TestResourceSnapshotRepositoryCreateGeneric(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(genericJson),
				readResponse(genericJson),
				readResponse(genericJson),
				readResponse(genericJson),
				deleteResponse(),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: genericSnapshotRepository,
				Check:  checkGenericResource(),
			},
			{ // Delete resource
				Destroy: true,
				Config:  genericSnapshotRepository,
			},
		},
	})
}

func TestResourceSnapshotRepository_failedCreate(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				failedCreateOrUpdateResponse(s3Json1),
			),
		),
		Steps: []r.TestStep{
			{
				Config:      awsSnapshotRepository1,
				ExpectError: regexp.MustCompile(`internal.server.error: There was an internal server error`),
			},
		},
	})
}

func TestResourceSnapshotRepository_failedReadAfterCreate(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(s3Json1),
				failedReadResponse(),
				deleteResponse(),
			),
		),
		Steps: []r.TestStep{
			{
				Config:      awsSnapshotRepository1,
				ExpectError: regexp.MustCompile(`failed reading snapshot repository`),
			},
		},
	})
}

func TestResourceSnapshotRepository_notFoundAfterCreate(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(s3Json1),
				notFoundReadResponse(),
			),
		),
		Steps: []r.TestStep{
			{
				Config:      awsSnapshotRepository1,
				ExpectError: regexp.MustCompile(`Failed to read snapshot repository after create.`),
			},
		},
	})
}

func TestResourceSnapshotRepository_notFoundAfterUpdate(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(s3Json1),
				readResponse(s3Json1),
				readResponse(s3Json1),
				readResponse(s3Json1),
				updateResponse(s3Json2),
				notFoundReadResponse(),
				deleteResponse(), // required for cleanup
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: awsSnapshotRepository1,
				Check:  checkS3Resource1(),
			},
			{ // Update resource
				Config:      awsSnapshotRepository2,
				ExpectError: regexp.MustCompile(`Failed to read snapshot repository after update.`),
			},
		},
	})
}

func TestResourceSnapshotRepository_notFoundAfterRead(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(s3Json1),
				readResponse(s3Json1),
				notFoundReadResponse(),
				deleteResponse(), // required for cleanup
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config:             awsSnapshotRepository1,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestResourceSnapshotRepository_failedRead(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(s3Json1),
				readResponse(s3Json1),
				failedReadResponse(),
				deleteResponse(), // required for cleanup
			),
		),
		Steps: []r.TestStep{
			{
				Config:      awsSnapshotRepository1,
				Check:       checkS3Resource1(),
				ExpectError: regexp.MustCompile(`internal.server.error: There was an internal server error`),
			},
		},
	})
}

func TestResourceSnapshotRepository_failedUpdate(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(s3Json1),
				readResponse(s3Json1),
				readResponse(s3Json1),
				readResponse(s3Json1),
				failedCreateOrUpdateResponse(s3Json2),
				deleteResponse(), // required for cleanup
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: awsSnapshotRepository1,
				Check:  checkS3Resource1(),
			},
			{ // Update resource
				Config:      awsSnapshotRepository2,
				ExpectError: regexp.MustCompile(`internal.server.error: There was an internal server error`),
			},
		},
	})
}

func TestResourceSnapshotRepository_failedReadAfterUpdate(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(s3Json1),
				readResponse(s3Json1),
				readResponse(s3Json1),
				readResponse(s3Json1),
				updateResponse(s3Json2),
				failedReadResponse(),
				deleteResponse(),
			),
		),
		Steps: []r.TestStep{
			{
				Config: awsSnapshotRepository1,
				Check:  checkS3Resource1(),
			},
			{ // Update resource
				Config:      awsSnapshotRepository2,
				ExpectError: regexp.MustCompile(`internal.server.error: There was an internal server error`),
			},
		},
	})
}

func TestResourceSnapshotRepository_gracefulDeletion(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(s3Json1),
				readResponse(s3Json1),
				readResponse(s3Json1),
				readResponse(s3Json1),
				alreadyDeletedResponse(),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: awsSnapshotRepository1,
				Check:  checkS3Resource1(),
			},
			{ // Delete resource
				Destroy: true,
				Config:  awsSnapshotRepository1,
			},
		},
	})
}

func TestResourceSnapshotRepository_failedDeletion(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewMock(
				createResponse(s3Json1),
				readResponse(s3Json1),
				readResponse(s3Json1),
				readResponse(s3Json1),
				failedDeletionResponse(),
				deleteResponse(), // required for cleanup
			),
		),
		Steps: []r.TestStep{
			{
				Config: awsSnapshotRepository1,
			},
			{
				Destroy:     true,
				Config:      awsSnapshotRepository1,
				ExpectError: regexp.MustCompile(`internal.server.error: There was an internal server error`),
			},
		},
	})
}

const awsSnapshotRepository1 = `
	resource "ec_snapshot_repository" "this" {
	  name = "my-snapshot-repository"
	  s3 = {
		region                 = "us-east-1"
		bucket                 = "my-bucket"
		access_key             = "my-access-key"
		secret_key             = "my-secret-key"
		server_side_encryption = true
		endpoint               = "s3.amazonaws.com"
		path_style_access      = true
	  }
	}
`
const awsSnapshotRepository2 = `
	resource "ec_snapshot_repository" "this" {
	  name = "my-snapshot-repository"
	  s3 = {
		region            = "us-west-1"
		bucket            = "my-bucket2"
		access_key        = "my-access-key2"
		secret_key        = "my-secret-key2"
		endpoint          = "s3.us-west-1.amazonaws.com"
		path_style_access = false
	  }
	}
`

const genericSnapshotRepository = `
	resource "ec_snapshot_repository" "this" {
	  name = "my-snapshot-repository"
	  generic = {
		type = "azure"
		settings = jsonencode({
		  bucket   = "my-bucket"
		  client   = "my_alternate_client"
		  compress = false
		})
	  }
	}
`

const s3Json1 = `{"settings":{"region":"us-east-1","bucket":"my-bucket","access_key":"my-access-key","secret_key":"my-secret-key","server_side_encryption":true,"endpoint":"s3.amazonaws.com","path_style_access":true},"type":"s3"}`
const s3Json2 = `{"settings":{"region":"us-west-1","bucket":"my-bucket2","access_key":"my-access-key2","secret_key":"my-secret-key2","endpoint":"s3.us-west-1.amazonaws.com"},"type":"s3"}`
const s3Json2WithPathStyleAccess = `{"settings":{"region":"us-west-1","bucket":"my-bucket2","access_key":"my-access-key2","secret_key":"my-secret-key2","endpoint":"s3.us-west-1.amazonaws.com","server_side_encryption":false,"path_style_access":false},"type":"s3"}`
const genericJson = `{"settings":{"bucket":"my-bucket","client":"my_alternate_client","compress":false},"type":"azure"}`

func checkS3Resource1() r.TestCheckFunc {
	resource := "ec_snapshot_repository.this"
	return r.ComposeAggregateTestCheckFunc(
		r.TestCheckResourceAttr(resource, "id", "my-snapshot-repository"),
		r.TestCheckResourceAttr(resource, "name", "my-snapshot-repository"),

		r.TestCheckResourceAttr(resource, "s3.region", "us-east-1"),
		r.TestCheckResourceAttr(resource, "s3.bucket", "my-bucket"),
		r.TestCheckResourceAttr(resource, "s3.access_key", "my-access-key"),
		r.TestCheckResourceAttr(resource, "s3.secret_key", "my-secret-key"),
		r.TestCheckResourceAttr(resource, "s3.server_side_encryption", "true"),
		r.TestCheckResourceAttr(resource, "s3.endpoint", "s3.amazonaws.com"),
		r.TestCheckResourceAttr(resource, "s3.path_style_access", "true"),
	)
}
func checkS3Resource2() r.TestCheckFunc {
	resource := "ec_snapshot_repository.this"
	return r.ComposeAggregateTestCheckFunc(
		r.TestCheckResourceAttr(resource, "id", "my-snapshot-repository"),
		r.TestCheckResourceAttr(resource, "name", "my-snapshot-repository"),
		r.TestCheckResourceAttr(resource, "s3.bucket", "my-bucket2"),
		r.TestCheckResourceAttr(resource, "s3.access_key", "my-access-key2"),
		r.TestCheckResourceAttr(resource, "s3.secret_key", "my-secret-key2"),
		r.TestCheckResourceAttr(resource, "s3.endpoint", "s3.us-west-1.amazonaws.com"),
		r.TestCheckResourceAttr(resource, "s3.path_style_access", "false"),
		r.TestCheckResourceAttr(resource, "s3.region", "us-west-1"),
	)
}
func checkGenericResource() r.TestCheckFunc {
	resource := "ec_snapshot_repository.this"
	return r.ComposeAggregateTestCheckFunc(
		r.TestCheckResourceAttr(resource, "id", "my-snapshot-repository"),
		r.TestCheckResourceAttr(resource, "name", "my-snapshot-repository"),
		r.TestCheckResourceAttr(resource, "generic.type", "azure"),
		r.TestCheckResourceAttr(resource, "generic.settings", "{\"bucket\":\"my-bucket\",\"client\":\"my_alternate_client\",\"compress\":false}"),
	)
}

func createResponse(json string) mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "PUT",
			Path:   "/api/v1/regions/ece-region/platform/configuration/snapshots/repositories/my-snapshot-repository",
			Body:   mock.NewStringBody(json + "\n"),
		},
		mock.NewStringBody(json),
	)
}

func updateResponse(json string) mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "PUT",
			Path:   "/api/v1/regions/ece-region/platform/configuration/snapshots/repositories/my-snapshot-repository",
			Body:   mock.NewStringBody(json + "\n"),
		},
		mock.NewStringBody(json),
	)
}

func failedCreateOrUpdateResponse(json string) mock.Response {
	return mock.New500ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "PUT",
			Path:   "/api/v1/regions/ece-region/platform/configuration/snapshots/repositories/my-snapshot-repository",
			Body:   mock.NewStringBody(json + "\n"),
		},
		mock.SampleInternalError().Response.Body,
	)
}

func readResponse(json string) mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "GET",
			Path:   "/api/v1/regions/ece-region/platform/configuration/snapshots/repositories/my-snapshot-repository",
		},
		mock.NewStringBody(`{
  "repository_name" : "my-snapshot-repository",
  "config" : `+json+`

}`),
	)
}

func failedReadResponse() mock.Response {
	return mock.New500ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "GET",
			Path:   "/api/v1/regions/ece-region/platform/configuration/snapshots/repositories/my-snapshot-repository",
		},
		mock.SampleInternalError().Response.Body,
	)
}

func notFoundReadResponse() mock.Response {
	return mock.New404ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "GET",
			Path:   "/api/v1/regions/ece-region/platform/configuration/snapshots/repositories/my-snapshot-repository",
		},
		mock.NewStringBody(`{"errors":[{"code":"root.resource_not_found","message":"The requested resource could not be found"}]}`),
	)
}

func deleteResponse() mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "DELETE",
			Path:   "/api/v1/regions/ece-region/platform/configuration/snapshots/repositories/my-snapshot-repository",
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
			Path:   "/api/v1/regions/ece-region/platform/configuration/snapshots/repositories/my-snapshot-repository",
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
			Path:   "/api/v1/regions/ece-region/platform/configuration/snapshots/repositories/my-snapshot-repository",
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
