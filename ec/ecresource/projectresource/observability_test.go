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

package projectresource

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/mocks"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_observability_project"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestObservabilityModelReader_Schema(t *testing.T) {
	mr := observabilityModelReader{}
	resp := resource.SchemaResponse{}
	mr.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, resource_observability_project.ObservabilityProjectResourceSchema(context.Background()), resp.Schema)
}

func TestObservabilityModelReader_ReadFrom(t *testing.T) {
	type testData struct {
		expectedModel *resource_observability_project.ObservabilityProjectModel
		rawState      tftypes.Value
	}
	tests := []struct {
		name     string
		testData func() testData
	}{
		{
			name: "should read a basic model back",
			testData: func() testData {
				model := resource_observability_project.ObservabilityProjectModel{
					Id: basetypes.NewStringValue("id"),
				}

				return testData{
					expectedModel: &model,
					rawState:      util.TfTypesValueFromGoTypeValue(t, model, resource_observability_project.ObservabilityProjectResourceSchema(context.Background()).Type()),
				}
			},
		},
		{
			name: "should return nil for if the config is unset",
			testData: func() testData {
				return testData{}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := tt.testData()
			mr := observabilityModelReader{}
			plan := tfsdk.State{
				Raw:    td.rawState,
				Schema: resource_observability_project.ObservabilityProjectResourceSchema(context.Background()),
			}

			model, diags := mr.ReadFrom(context.Background(), plan)
			require.False(t, diags.HasError())
			require.Equal(t, td.expectedModel, model)
		})
	}
}

func TestObservabilityModelReader_GetID(t *testing.T) {
	mr := observabilityModelReader{}
	expectedId := "expected_id"
	model := resource_observability_project.ObservabilityProjectModel{
		Id: basetypes.NewStringValue(expectedId),
	}

	require.Equal(t, expectedId, mr.GetID(model))
}

func TestObservabilityModelReader_Modify(t *testing.T) {
	type testData struct {
		state    resource_observability_project.ObservabilityProjectModel
		plan     resource_observability_project.ObservabilityProjectModel
		cfg      resource_observability_project.ObservabilityProjectModel
		expected resource_observability_project.ObservabilityProjectModel
	}
	tests := []struct {
		name     string
		testData func() testData
	}{
		{
			name: "should use state for unknown credentials",
			testData: func() testData {
				state := resource_observability_project.ObservabilityProjectModel{
					Id: types.StringValue("state"),
				}
				state.Credentials = resource_observability_project.NewCredentialsValueMust(
					state.Credentials.AttributeTypes(context.Background()),
					map[string]attr.Value{
						"username": types.StringValue("username"),
						"password": types.StringValue("password"),
					},
				)

				return testData{
					plan: resource_observability_project.ObservabilityProjectModel{
						Id:          types.StringValue("plan"),
						Credentials: resource_observability_project.NewCredentialsValueUnknown(),
					},
					state: state,
					expected: resource_observability_project.ObservabilityProjectModel{
						Id:          types.StringValue("plan"),
						Credentials: state.Credentials,
					},
				}
			},
		},
		{
			name: "should use state for unknown endpoints",
			testData: func() testData {
				state := resource_observability_project.ObservabilityProjectModel{
					Id: types.StringValue("state"),
				}
				state.Endpoints = resource_observability_project.NewEndpointsValueMust(
					state.Endpoints.AttributeTypes(context.Background()),
					map[string]attr.Value{
						"elasticsearch": basetypes.NewStringValue("es"),
						"kibana":        basetypes.NewStringValue("kibana"),
						"apm":           basetypes.NewStringValue("apm"),
						"ingest":        basetypes.NewStringValue("ingest"),
					},
				)

				return testData{
					plan: resource_observability_project.ObservabilityProjectModel{
						Id:        types.StringValue("plan"),
						Endpoints: resource_observability_project.NewEndpointsValueUnknown(),
					},
					state: state,
					expected: resource_observability_project.ObservabilityProjectModel{
						Id:        types.StringValue("plan"),
						Endpoints: state.Endpoints,
					},
				}
			},
		},
		{
			name: "should use state for unknown metadata",
			testData: func() testData {
				state := resource_observability_project.ObservabilityProjectModel{
					Id: types.StringValue("state"),
				}
				state.Metadata = resource_observability_project.NewMetadataValueMust(
					state.Metadata.AttributeTypes(context.Background()),
					map[string]attr.Value{
						"created_at":       basetypes.NewStringValue("created_at"),
						"created_by":       basetypes.NewStringValue("created_by"),
						"organization_id":  basetypes.NewStringValue("org_id"),
						"suspended_at":     basetypes.NewStringNull(),
						"suspended_reason": basetypes.NewStringValue("suspension_reason"),
					},
				)

				return testData{
					plan: resource_observability_project.ObservabilityProjectModel{
						Id:       types.StringValue("plan"),
						Metadata: resource_observability_project.NewMetadataValueUnknown(),
					},
					state: state,
					expected: resource_observability_project.ObservabilityProjectModel{
						Id:       types.StringValue("plan"),
						Metadata: state.Metadata,
					},
				}
			},
		},
		{
			name: "cloud id should be unknown if name has changed",
			testData: func() testData {
				return testData{
					plan: resource_observability_project.ObservabilityProjectModel{
						Id:    types.StringValue("plan"),
						Name:  types.StringValue("planned name"),
						Alias: types.StringValue("alias"),
					},
					state: resource_observability_project.ObservabilityProjectModel{
						Id:    types.StringValue("state"),
						Name:  types.StringValue("state name"),
						Alias: types.StringValue("alias"),
					},
					cfg: resource_observability_project.ObservabilityProjectModel{
						Alias: types.StringValue("alias"),
					},
					expected: resource_observability_project.ObservabilityProjectModel{
						Id:      types.StringValue("plan"),
						Name:    types.StringValue("planned name"),
						Alias:   types.StringValue("alias"),
						CloudId: types.StringUnknown(),
					},
				}
			},
		},
		{
			name: "cloud id and endpoints should be unknown if alias has changed",
			testData: func() testData {
				return testData{
					plan: resource_observability_project.ObservabilityProjectModel{
						Id:    types.StringValue("plan"),
						Name:  types.StringValue("name"),
						Alias: types.StringValue("planned alias"),
					},
					state: resource_observability_project.ObservabilityProjectModel{
						Id:    types.StringValue("state"),
						Name:  types.StringValue("name"),
						Alias: types.StringValue("state alias"),
					},
					expected: resource_observability_project.ObservabilityProjectModel{
						Id:        types.StringValue("plan"),
						Name:      types.StringValue("name"),
						Alias:     types.StringValue("planned alias"),
						CloudId:   types.StringUnknown(),
						Endpoints: resource_observability_project.NewEndpointsValueUnknown(),
					},
				}
			},
		},
		{
			name: "cloud id, alias, and endpoints should be unknown if name has changed but alias is not configured",
			testData: func() testData {
				return testData{
					plan: resource_observability_project.ObservabilityProjectModel{
						Id:   types.StringValue("plan"),
						Name: types.StringValue("planned name"),
					},
					state: resource_observability_project.ObservabilityProjectModel{
						Id:   types.StringValue("state"),
						Name: types.StringValue("state name"),
					},
					expected: resource_observability_project.ObservabilityProjectModel{
						Id:        types.StringValue("plan"),
						Name:      types.StringValue("planned name"),
						CloudId:   types.StringUnknown(),
						Alias:     types.StringUnknown(),
						Endpoints: resource_observability_project.NewEndpointsValueUnknown(),
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := tt.testData()
			mr := observabilityModelReader{}

			require.Equal(t, td.expected, mr.Modify(td.plan, td.state, td.cfg))
		})
	}
}

func TestObservabilityApi_Ready(t *testing.T) {
	t.Run("is ready when the client is configured", func(t *testing.T) {
		api := observabilityApi{client: &serverless.ClientWithResponses{}}
		require.True(t, api.Ready())
	})
	t.Run("is not ready when the client is not configured", func(t *testing.T) {
		api := observabilityApi{}
		require.False(t, api.Ready())
	})
}

func TestObservabilityApi_WithClient(t *testing.T) {
	var api api[resource_observability_project.ObservabilityProjectModel] = observabilityApi{}

	require.False(t, api.Ready())

	api = api.WithClient(&serverless.ClientWithResponses{})
	require.True(t, api.Ready())
}

func TestObservabilityApi_Create(t *testing.T) {
	ctrl := gomock.NewController(t)

	type testData struct {
		client        serverless.ClientWithResponsesInterface
		initialModel  resource_observability_project.ObservabilityProjectModel
		expectedModel resource_observability_project.ObservabilityProjectModel
		expectedDiags diag.Diagnostics
	}
	tests := []struct {
		name     string
		testData func(ctx context.Context) testData
	}{
		{
			name: "should fail when the api returns an error",
			testData: func(ctx context.Context) testData {
				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().CreateObservabilityProjectWithResponse(ctx, gomock.Any()).Return(
					nil,
					assert.AnError,
				)

				return testData{
					client: mockApiClient,
					expectedDiags: diag.Diagnostics{
						diag.NewErrorDiagnostic(assert.AnError.Error(), assert.AnError.Error()),
					},
				}
			},
		},
		{
			name: "should fail when the api call does not return a 201 response",
			testData: func(ctx context.Context) testData {
				failedResponse := &serverless.CreateObservabilityProjectResponse{
					HTTPResponse: &http.Response{
						Status:     "failed",
						StatusCode: 400,
					},
					Body: []byte("api call failed"),
				}
				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().CreateObservabilityProjectWithResponse(ctx, gomock.Any()).Return(
					failedResponse,
					nil,
				)

				return testData{
					client: mockApiClient,
					expectedDiags: diag.Diagnostics{
						diag.NewErrorDiagnostic(
							"Failed to create observability_project",
							fmt.Sprintf("The API request failed with: %d %s\n%s",
								failedResponse.StatusCode(),
								failedResponse.Status(),
								failedResponse.Body),
						),
					},
				}
			},
		},
		{
			name: "should not populate unset optional fields in create request",
			testData: func(ctx context.Context) testData {
				initialModel := resource_observability_project.ObservabilityProjectModel{
					Name:     types.StringValue("project name"),
					RegionId: types.StringValue("nether region"),
				}
				createdProject := serverless.ObservabilityProjectCreated{
					Id: "created id",
					Credentials: serverless.ProjectCredentials{
						Username: "project username",
						Password: "sekret",
					},
				}
				expectedProject := initialModel
				expectedProject.Id = types.StringValue(createdProject.Id)
				expectedProject.Credentials = resource_observability_project.NewCredentialsValueMust(
					initialModel.Credentials.AttributeTypes(ctx),
					map[string]attr.Value{
						"username": types.StringValue(createdProject.Credentials.Username),
						"password": types.StringValue(createdProject.Credentials.Password),
					},
				)
				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().CreateObservabilityProjectWithResponse(ctx, serverless.CreateObservabilityProjectRequest{
					Name:     initialModel.Name.ValueString(),
					RegionId: initialModel.RegionId.ValueString(),
				}).Return(
					&serverless.CreateObservabilityProjectResponse{
						JSON201: &createdProject,
					},
					nil,
				)

				return testData{
					client:        mockApiClient,
					initialModel:  initialModel,
					expectedModel: expectedProject,
				}
			},
		},
		{
			name: "should populate provided optional fields in create request",
			testData: func(ctx context.Context) testData {
				initialModel := resource_observability_project.ObservabilityProjectModel{
					Name:     types.StringValue("project name"),
					RegionId: types.StringValue("nether region"),
					Alias:    types.StringValue("project alias"),
				}

				createdProject := serverless.ObservabilityProjectCreated{
					Id: "created id",
					Credentials: serverless.ProjectCredentials{
						Username: "project username",
						Password: "sekret",
					},
				}
				expectedProject := initialModel
				expectedProject.Id = types.StringValue(createdProject.Id)
				expectedProject.Credentials = resource_observability_project.NewCredentialsValueMust(
					initialModel.Credentials.AttributeTypes(ctx),
					map[string]attr.Value{
						"username": types.StringValue(createdProject.Credentials.Username),
						"password": types.StringValue(createdProject.Credentials.Password),
					},
				)

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().CreateObservabilityProjectWithResponse(ctx, serverless.CreateObservabilityProjectRequest{
					Name:     initialModel.Name.ValueString(),
					RegionId: initialModel.RegionId.ValueString(),
					Alias:    initialModel.Alias.ValueStringPointer(),
				}).Return(
					&serverless.CreateObservabilityProjectResponse{
						JSON201: &createdProject,
					},
					nil,
				)

				return testData{
					client:        mockApiClient,
					initialModel:  initialModel,
					expectedModel: expectedProject,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			td := tt.testData(ctx)

			api := observabilityApi{}.WithClient(td.client)
			model, diags := api.Create(ctx, td.initialModel)

			if td.expectedDiags != nil {
				require.Equal(t, td.expectedDiags, diags)
			} else {
				require.False(t, diags.HasError())
			}

			require.Equal(t, td.expectedModel, model)
		})
	}
}

func TestObservabilityApi_Patch(t *testing.T) {
	ctrl := gomock.NewController(t)

	type testData struct {
		client        serverless.ClientWithResponsesInterface
		model         resource_observability_project.ObservabilityProjectModel
		expectedDiags diag.Diagnostics
	}
	tests := []struct {
		name     string
		testData func(ctx context.Context) testData
	}{
		{
			name: "should fail when the api returns an error",
			testData: func(ctx context.Context) testData {
				model := resource_observability_project.ObservabilityProjectModel{
					Id:   types.StringValue("project id"),
					Name: types.StringValue("project name"),
				}
				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().PatchObservabilityProjectWithResponse(ctx, model.Id.ValueString(), nil, serverless.PatchObservabilityProjectRequest{
					Name: model.Name.ValueStringPointer(),
				}).Return(
					nil,
					assert.AnError,
				)

				return testData{
					client: mockApiClient,
					model:  model,
					expectedDiags: diag.Diagnostics{
						diag.NewErrorDiagnostic(assert.AnError.Error(), assert.AnError.Error()),
					},
				}
			},
		},
		{
			name: "should fail when the api call does not return a 201 response",
			testData: func(ctx context.Context) testData {
				model := resource_observability_project.ObservabilityProjectModel{
					Id:   types.StringValue("project id"),
					Name: types.StringValue("project name"),
				}
				failedResponse := &serverless.PatchObservabilityProjectResponse{
					HTTPResponse: &http.Response{
						Status:     "failed",
						StatusCode: 400,
					},
					Body: []byte("api call failed"),
				}
				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().PatchObservabilityProjectWithResponse(ctx, model.Id.ValueString(), nil, serverless.PatchObservabilityProjectRequest{
					Name: model.Name.ValueStringPointer(),
				}).Return(
					failedResponse,
					nil,
				)

				return testData{
					client: mockApiClient,
					model:  model,
					expectedDiags: diag.Diagnostics{
						diag.NewErrorDiagnostic(
							"Failed to update observability_project",
							fmt.Sprintf("The API request failed with: %d %s\n%s",
								failedResponse.StatusCode(),
								failedResponse.Status(),
								failedResponse.Body),
						),
					},
				}
			},
		},
		{
			name: "should not populate unset optional fields in patch request",
			testData: func(ctx context.Context) testData {
				model := resource_observability_project.ObservabilityProjectModel{
					Id:       types.StringValue("project id"),
					Name:     types.StringValue("project name"),
					RegionId: types.StringValue("nether region"),
				}
				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().PatchObservabilityProjectWithResponse(ctx, model.Id.ValueString(), nil, serverless.PatchObservabilityProjectRequest{
					Name: model.Name.ValueStringPointer(),
				}).Return(
					&serverless.PatchObservabilityProjectResponse{
						JSON200: &serverless.ObservabilityProject{},
					},
					nil,
				)

				return testData{
					client: mockApiClient,
					model:  model,
				}
			},
		},
		{
			name: "should populate provided optional fields in create request",
			testData: func(ctx context.Context) testData {
				model := resource_observability_project.ObservabilityProjectModel{
					Id:       types.StringValue("project id"),
					Name:     types.StringValue("project name"),
					RegionId: types.StringValue("nether region"),
					Alias:    types.StringValue("project alias"),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().PatchObservabilityProjectWithResponse(ctx, model.Id.ValueString(), nil, serverless.PatchObservabilityProjectRequest{
					Name:  model.Name.ValueStringPointer(),
					Alias: model.Alias.ValueStringPointer(),
				}).Return(
					&serverless.PatchObservabilityProjectResponse{
						JSON200: &serverless.ObservabilityProject{},
					},
					nil,
				)

				return testData{
					client: mockApiClient,
					model:  model,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			td := tt.testData(ctx)

			api := observabilityApi{}.WithClient(td.client)
			diags := api.Patch(ctx, td.model)

			if td.expectedDiags != nil {
				require.Equal(t, td.expectedDiags, diags)
			} else {
				require.False(t, diags.HasError())
			}
		})
	}
}

func TestObservabilityApi_EnsureInitialised(t *testing.T) {
	ctrl := gomock.NewController(t)
	type testData struct {
		client        serverless.ClientWithResponsesInterface
		model         resource_observability_project.ObservabilityProjectModel
		expectedDiags diag.Diagnostics
	}
	tests := []struct {
		name     string
		testData func(ctx context.Context) testData
	}{
		{
			name: "should error if status check errors eventually",
			testData: func(ctx context.Context) testData {
				callsBeforeInitialised := rand.Intn(20)
				model := resource_observability_project.ObservabilityProjectModel{
					Id: types.StringValue("project id"),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().GetObservabilityProjectStatusWithResponse(ctx, model.Id.ValueString()).DoAndReturn(
					func(_ context.Context, id string, _ ...serverless.RequestEditorFn) (*serverless.GetObservabilityProjectStatusResponse, error) {
						if callsBeforeInitialised > 0 {
							callsBeforeInitialised--
							return &serverless.GetObservabilityProjectStatusResponse{
								JSON200: &serverless.ProjectStatus{Phase: serverless.Initializing},
							}, nil
						}

						return nil, assert.AnError
					},
				).Times(callsBeforeInitialised + 1)

				return testData{
					client: mockApiClient,
					model:  model,
					expectedDiags: diag.Diagnostics{
						diag.NewErrorDiagnostic(assert.AnError.Error(), assert.AnError.Error()),
					},
				}
			},
		},
		{
			name: "should error if status check returns a non-200 response eventually",
			testData: func(ctx context.Context) testData {
				callsBeforeInitialised := rand.Intn(20)
				model := resource_observability_project.ObservabilityProjectModel{
					Id: types.StringValue("project id"),
				}

				failedResponse := &serverless.GetObservabilityProjectStatusResponse{
					HTTPResponse: &http.Response{
						Status:     "failed",
						StatusCode: 400,
					},
					Body: []byte("api call failed"),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().GetObservabilityProjectStatusWithResponse(ctx, model.Id.ValueString()).DoAndReturn(
					func(_ context.Context, id string, _ ...serverless.RequestEditorFn) (*serverless.GetObservabilityProjectStatusResponse, error) {
						if callsBeforeInitialised > 0 {
							callsBeforeInitialised--
							return &serverless.GetObservabilityProjectStatusResponse{
								JSON200: &serverless.ProjectStatus{Phase: serverless.Initializing},
							}, nil
						}

						return failedResponse, nil
					},
				).Times(callsBeforeInitialised + 1)

				return testData{
					client: mockApiClient,
					model:  model,
					expectedDiags: diag.Diagnostics{

						diag.NewErrorDiagnostic(
							"Failed to get observability_project status",
							fmt.Sprintf("The API request failed with: %d %s\n%s",
								failedResponse.StatusCode(),
								failedResponse.Status(),
								failedResponse.Body),
						),
					},
				}
			},
		},
		{
			name: "should return when the model is eventually initialised",
			testData: func(ctx context.Context) testData {
				callsBeforeInitialised := rand.Intn(20)
				model := resource_observability_project.ObservabilityProjectModel{
					Id: types.StringValue("project id"),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().GetObservabilityProjectStatusWithResponse(ctx, model.Id.ValueString()).DoAndReturn(
					func(_ context.Context, id string, _ ...serverless.RequestEditorFn) (*serverless.GetObservabilityProjectStatusResponse, error) {
						phase := serverless.Initialized

						if callsBeforeInitialised > 0 {
							callsBeforeInitialised--
							phase = serverless.Initializing
						}

						return &serverless.GetObservabilityProjectStatusResponse{
							JSON200: &serverless.ProjectStatus{Phase: phase},
						}, nil
					},
				).Times(callsBeforeInitialised + 1)

				return testData{
					client: mockApiClient,
					model:  model,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			td := tt.testData(ctx)
			api := observabilityApi{sleeper: fakeSleeper{}}.WithClient(td.client)

			diags := api.EnsureInitialised(ctx, td.model)
			if td.expectedDiags != nil {
				require.Equal(t, td.expectedDiags, diags)
			} else {
				require.Nil(t, diags)
			}
		})
	}
}

func TestObservabilityApi_Read(t *testing.T) {
	ctrl := gomock.NewController(t)

	type testData struct {
		client        serverless.ClientWithResponsesInterface
		id            string
		initialModel  resource_observability_project.ObservabilityProjectModel
		expectedModel resource_observability_project.ObservabilityProjectModel
		expectedFound bool
		expectedDiags diag.Diagnostics
	}
	tests := []struct {
		name     string
		testData func(context.Context) testData
	}{
		{
			name: "should error if the get call errors",
			testData: func(ctx context.Context) testData {
				id := "project id"
				initialModel := resource_observability_project.ObservabilityProjectModel{
					Id: types.StringValue(id),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().GetObservabilityProjectWithResponse(ctx, id).Return(nil, assert.AnError)

				return testData{
					client:        mockApiClient,
					id:            id,
					initialModel:  initialModel,
					expectedModel: initialModel,
					expectedDiags: diag.Diagnostics{
						diag.NewErrorDiagnostic(assert.AnError.Error(), assert.AnError.Error()),
					},
				}
			},
		},
		{
			name: "should return not found get returns a 404 response",
			testData: func(ctx context.Context) testData {
				id := "project id"
				initialModel := resource_observability_project.ObservabilityProjectModel{
					Id: types.StringValue(id),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().
					GetObservabilityProjectWithResponse(ctx, id).
					Return(&serverless.GetObservabilityProjectResponse{
						HTTPResponse: &http.Response{
							StatusCode: http.StatusNotFound,
						},
					}, nil)

				return testData{
					client:        mockApiClient,
					id:            id,
					initialModel:  initialModel,
					expectedModel: initialModel,
				}
			},
		},
		{
			name: "should error if get returns an error response",
			testData: func(ctx context.Context) testData {
				id := "project id"
				initialModel := resource_observability_project.ObservabilityProjectModel{
					Id: types.StringValue(id),
				}

				failedResponse := &serverless.GetObservabilityProjectResponse{
					HTTPResponse: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     "nope",
					},
					Body: []byte("failed"),
				}
				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().
					GetObservabilityProjectWithResponse(ctx, id).
					Return(failedResponse, nil)

				return testData{
					client:        mockApiClient,
					id:            id,
					initialModel:  initialModel,
					expectedModel: initialModel,
					expectedDiags: diag.Diagnostics{
						diag.NewErrorDiagnostic(
							"Failed to read observability_project",
							fmt.Sprintf("The API request failed with: %d %s\n%s",
								failedResponse.StatusCode(),
								failedResponse.Status(),
								failedResponse.Body),
						),
					},
				}
			},
		},
		{
			name: "should populate model values on a successful response",
			testData: func(ctx context.Context) testData {
				id := "project id"
				initialModel := resource_observability_project.ObservabilityProjectModel{
					Id: types.StringValue(id),
				}

				readModel := &serverless.ObservabilityProject{
					Id:      id,
					Alias:   "expected-alias-" + id[0:6],
					CloudId: "cloud-id",
					Endpoints: serverless.ObservabilityProjectEndpoints{
						Elasticsearch: "es-endpoint",
						Kibana:        "kib-endpoint",
						Apm:           "apm-endpoint",
					},
					Metadata: serverless.ProjectMetadata{
						CreatedAt:      time.Now(),
						CreatedBy:      "me",
						OrganizationId: "1",
					},
					Name:     "project-name",
					RegionId: "nether",
					Type:     "observability",
				}

				expectedModel := resource_observability_project.ObservabilityProjectModel{
					Id:      types.StringValue(id),
					Alias:   types.StringValue("expected-alias"),
					CloudId: types.StringValue(readModel.CloudId),
					Endpoints: resource_observability_project.NewEndpointsValueMust(
						initialModel.Endpoints.AttributeTypes(ctx),
						map[string]attr.Value{
							"elasticsearch": basetypes.NewStringValue(readModel.Endpoints.Elasticsearch),
							"kibana":        basetypes.NewStringValue(readModel.Endpoints.Kibana),
							"apm":           basetypes.NewStringValue(readModel.Endpoints.Apm),
							"ingest":        basetypes.NewStringValue(readModel.Endpoints.Ingest),
						},
					),
					Metadata: resource_observability_project.NewMetadataValueMust(
						initialModel.Metadata.AttributeTypes(ctx),
						map[string]attr.Value{
							"created_at":       basetypes.NewStringValue(readModel.Metadata.CreatedAt.String()),
							"created_by":       basetypes.NewStringValue(readModel.Metadata.CreatedBy),
							"organization_id":  basetypes.NewStringValue(readModel.Metadata.OrganizationId),
							"suspended_at":     basetypes.NewStringNull(),
							"suspended_reason": basetypes.NewStringNull(),
						},
					),
					Name:     types.StringValue(readModel.Name),
					RegionId: types.StringValue(readModel.RegionId),
					Type:     types.StringValue(string(readModel.Type)),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().
					GetObservabilityProjectWithResponse(ctx, id).
					Return(&serverless.GetObservabilityProjectResponse{
						JSON200: readModel,
					}, nil)

				return testData{
					client:        mockApiClient,
					id:            id,
					initialModel:  initialModel,
					expectedModel: expectedModel,
					expectedFound: true,
				}
			},
		},
		{
			name: "should populate optional model values provided on a successful response",
			testData: func(ctx context.Context) testData {
				id := "project id"
				initialModel := resource_observability_project.ObservabilityProjectModel{
					Id: types.StringValue(id),
				}

				now := time.Now()
				readModel := &serverless.ObservabilityProject{
					Id:      id,
					Alias:   "expected-alias-" + id[0:6],
					CloudId: "cloud-id",
					Endpoints: serverless.ObservabilityProjectEndpoints{
						Elasticsearch: "es-endpoint",
						Kibana:        "kib-endpoint",
						Apm:           "apm-endpoint",
					},
					Metadata: serverless.ProjectMetadata{
						CreatedAt:       now,
						CreatedBy:       "me",
						OrganizationId:  "1",
						SuspendedAt:     util.Ptr(now),
						SuspendedReason: util.Ptr("meh"),
					},
					Name:     "project-name",
					RegionId: "nether",
					Type:     "observability",
				}

				expectedModel := resource_observability_project.ObservabilityProjectModel{
					Id:      types.StringValue(id),
					Alias:   types.StringValue("expected-alias"),
					CloudId: types.StringValue(readModel.CloudId),
					Endpoints: resource_observability_project.NewEndpointsValueMust(
						initialModel.Endpoints.AttributeTypes(ctx),
						map[string]attr.Value{
							"elasticsearch": basetypes.NewStringValue(readModel.Endpoints.Elasticsearch),
							"kibana":        basetypes.NewStringValue(readModel.Endpoints.Kibana),
							"apm":           basetypes.NewStringValue(readModel.Endpoints.Apm),
							"ingest":        basetypes.NewStringValue(readModel.Endpoints.Ingest),
						},
					),
					Metadata: resource_observability_project.NewMetadataValueMust(
						initialModel.Metadata.AttributeTypes(ctx),
						map[string]attr.Value{
							"created_at":       basetypes.NewStringValue(readModel.Metadata.CreatedAt.String()),
							"created_by":       basetypes.NewStringValue(readModel.Metadata.CreatedBy),
							"organization_id":  basetypes.NewStringValue(readModel.Metadata.OrganizationId),
							"suspended_at":     basetypes.NewStringValue(now.String()),
							"suspended_reason": basetypes.NewStringValue(*readModel.Metadata.SuspendedReason),
						},
					),
					Name:     types.StringValue(readModel.Name),
					RegionId: types.StringValue(readModel.RegionId),
					Type:     types.StringValue(string(readModel.Type)),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().
					GetObservabilityProjectWithResponse(ctx, id).
					Return(&serverless.GetObservabilityProjectResponse{
						JSON200: readModel,
					}, nil)

				return testData{
					client:        mockApiClient,
					id:            id,
					initialModel:  initialModel,
					expectedModel: expectedModel,
					expectedFound: true,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			td := tt.testData(ctx)

			api := observabilityApi{}.WithClient(td.client)
			found, model, diags := api.Read(ctx, td.id, td.initialModel)

			require.Equal(t, td.expectedFound, found)
			require.Equal(t, td.expectedModel, model)
			require.Equal(t, td.expectedDiags, diags)
		})
	}
}

func TestObservabilityApi_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	type testData struct {
		client        serverless.ClientWithResponsesInterface
		model         resource_observability_project.ObservabilityProjectModel
		expectedDiags diag.Diagnostics
	}
	tests := []struct {
		name     string
		testData func(ctx context.Context) testData
	}{
		{
			name: "should error if delete errors",
			testData: func(ctx context.Context) testData {
				model := resource_observability_project.ObservabilityProjectModel{
					Id: types.StringValue("project id"),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().
					DeleteObservabilityProjectWithResponse(ctx, model.Id.ValueString(), nil).
					Return(nil, assert.AnError)

				return testData{
					client: mockApiClient,
					model:  model,
					expectedDiags: diag.Diagnostics{
						diag.NewErrorDiagnostic("Failed to delete observability_project", assert.AnError.Error()),
					},
				}
			},
		},
		{
			name: "should error if delete returns a non-200 and non-404 response",
			testData: func(ctx context.Context) testData {
				model := resource_observability_project.ObservabilityProjectModel{
					Id: types.StringValue("project id"),
				}

				failedResponse := &serverless.DeleteObservabilityProjectResponse{
					HTTPResponse: &http.Response{
						Status:     "failed",
						StatusCode: 400,
					},
					Body: []byte("api call failed"),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().
					DeleteObservabilityProjectWithResponse(ctx, model.Id.ValueString(), nil).
					Return(failedResponse, nil)

				return testData{
					client: mockApiClient,
					model:  model,
					expectedDiags: diag.Diagnostics{

						diag.NewErrorDiagnostic(
							"Request to delete observability_project failed",
							fmt.Sprintf("The API request failed with: %d %s\n%s",
								failedResponse.StatusCode(),
								failedResponse.Status(),
								failedResponse.Body),
						),
					},
				}
			},
		},
		{
			name: "should succeed if delete returns a 404 response",
			testData: func(ctx context.Context) testData {
				model := resource_observability_project.ObservabilityProjectModel{
					Id: types.StringValue("project id"),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().
					DeleteObservabilityProjectWithResponse(ctx, model.Id.ValueString(), nil).
					Return(&serverless.DeleteObservabilityProjectResponse{
						HTTPResponse: &http.Response{StatusCode: 404},
					}, nil)

				return testData{
					client: mockApiClient,
					model:  model,
				}
			},
		},
		{
			name: "should succeed if delete returns a 200 response",
			testData: func(ctx context.Context) testData {
				model := resource_observability_project.ObservabilityProjectModel{
					Id: types.StringValue("project id"),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().
					DeleteObservabilityProjectWithResponse(ctx, model.Id.ValueString(), nil).
					Return(&serverless.DeleteObservabilityProjectResponse{
						HTTPResponse: &http.Response{StatusCode: 200},
					}, nil)

				return testData{
					client: mockApiClient,
					model:  model,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			td := tt.testData(ctx)
			api := observabilityApi{sleeper: fakeSleeper{}}.WithClient(td.client)

			diags := api.Delete(ctx, td.model)
			require.Equal(t, td.expectedDiags, diags)
		})
	}
}
