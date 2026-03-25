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

package serverlesstrafficfilterresource

import (
	"context"
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/mocks"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func ptr(s string) *string {
	return &s
}

// newTestState creates a tfsdk.State from raw attribute values for the traffic filter resource.
func newTestState(ctx context.Context, t *testing.T, id, name, region, typ string) tfsdk.State {
	t.Helper()
	r := &Resource{}
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	state := tfsdk.State{Schema: schemaResp.Schema}
	state.Raw = tftypes.NewValue(state.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":                 tftypes.NewValue(tftypes.String, id),
		"name":               tftypes.NewValue(tftypes.String, name),
		"region":             tftypes.NewValue(tftypes.String, region),
		"type":               tftypes.NewValue(tftypes.String, typ),
		"description":        tftypes.NewValue(tftypes.String, nil),
		"include_by_default": tftypes.NewValue(tftypes.Bool, nil),
		"rules":              tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"source": tftypes.String, "description": tftypes.String}}}, nil),
	})
	return state
}

func TestResourceRead(t *testing.T) {
	ctx := context.Background()

	t.Run("successful read populates state", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)
		mockClient.EXPECT().
			GetTrafficFilterWithResponse(ctx, "filter-123").
			Return(&serverless.GetTrafficFilterResponse{
				HTTPResponse: &http.Response{StatusCode: 200},
				JSON200: &serverless.TrafficFilterInfo{
					Id:               "filter-123",
					Name:             "test-filter",
					Region:           "aws-us-east-1",
					Type:             serverless.Ip,
					Description:      ptr("Test filter"),
					IncludeByDefault: false,
					Rules: []serverless.TrafficFilterRule{
						{Source: "192.168.1.0/24", Description: ptr("Office")},
					},
				},
			}, nil)

		r := &Resource{client: mockClient}
		state := newTestState(ctx, t, "filter-123", "test-filter", "aws-us-east-1", "ip")
		readReq := resource.ReadRequest{State: state}
		readResp := resource.ReadResponse{State: state, Diagnostics: diag.Diagnostics{}}

		r.Read(ctx, readReq, &readResp)

		require.False(t, readResp.Diagnostics.HasError(), "unexpected diags: %v", readResp.Diagnostics)
	})

	t.Run("404 removes resource from state", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)
		mockClient.EXPECT().
			GetTrafficFilterWithResponse(ctx, "nonexistent").
			Return(&serverless.GetTrafficFilterResponse{
				HTTPResponse: &http.Response{StatusCode: 404},
			}, nil)

		r := &Resource{client: mockClient}
		state := newTestState(ctx, t, "nonexistent", "test", "aws-us-east-1", "ip")
		readReq := resource.ReadRequest{State: state}
		readResp := resource.ReadResponse{State: state, Diagnostics: diag.Diagnostics{}}

		r.Read(ctx, readReq, &readResp)

		require.False(t, readResp.Diagnostics.HasError())
		// State should be removed (empty raw)
		assert.True(t, readResp.State.Raw.IsNull())
	})
}

func TestResourceDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)
		mockClient.EXPECT().
			DeleteTrafficFilterWithResponse(ctx, "filter-123").
			Return(&serverless.DeleteTrafficFilterResponse{
				HTTPResponse: &http.Response{StatusCode: 200},
			}, nil)

		r := &Resource{client: mockClient}
		state := newTestState(ctx, t, "filter-123", "test", "aws-us-east-1", "ip")
		deleteReq := resource.DeleteRequest{State: state}
		deleteResp := resource.DeleteResponse{State: state, Diagnostics: diag.Diagnostics{}}

		r.Delete(ctx, deleteReq, &deleteResp)

		require.False(t, deleteResp.Diagnostics.HasError())
	})

	t.Run("successful deletion with 204 No Content", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)
		mockClient.EXPECT().
			DeleteTrafficFilterWithResponse(ctx, "filter-456").
			Return(&serverless.DeleteTrafficFilterResponse{
				HTTPResponse: &http.Response{StatusCode: 204},
			}, nil)

		r := &Resource{client: mockClient}
		state := newTestState(ctx, t, "filter-456", "test", "aws-us-east-1", "ip")
		deleteReq := resource.DeleteRequest{State: state}
		deleteResp := resource.DeleteResponse{State: state, Diagnostics: diag.Diagnostics{}}

		r.Delete(ctx, deleteReq, &deleteResp)

		require.False(t, deleteResp.Diagnostics.HasError())
	})

	t.Run("404 on delete is not an error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)
		mockClient.EXPECT().
			DeleteTrafficFilterWithResponse(ctx, "nonexistent").
			Return(&serverless.DeleteTrafficFilterResponse{
				HTTPResponse: &http.Response{StatusCode: 404},
			}, nil)

		r := &Resource{client: mockClient}
		state := newTestState(ctx, t, "nonexistent", "test", "aws-us-east-1", "ip")
		deleteReq := resource.DeleteRequest{State: state}
		deleteResp := resource.DeleteResponse{State: state, Diagnostics: diag.Diagnostics{}}

		r.Delete(ctx, deleteReq, &deleteResp)

		require.False(t, deleteResp.Diagnostics.HasError())
	})
}

func TestResourceReadyGuard(t *testing.T) {
	r := &Resource{client: nil}
	var diags diag.Diagnostics
	assert.False(t, r.ready(&diags))
	assert.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary(), "Unconfigured API Client")
}
