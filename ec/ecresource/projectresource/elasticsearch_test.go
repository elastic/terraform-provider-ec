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
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_elasticsearch_project"
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

func TestElasticsearchModelReader_Schema(t *testing.T) {
	mr := elasticsearchModelReader{}
	resp := resource.SchemaResponse{}
	mr.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, resource_elasticsearch_project.ElasticsearchProjectResourceSchema(context.Background()), resp.Schema)
}

func TestElasticsearchModelReader_ReadFrom(t *testing.T) {
	type testData struct {
		expectedModel *resource_elasticsearch_project.ElasticsearchProjectModel
		rawState      tftypes.Value
	}
	tests := []struct {
		name     string
		testData func() testData
	}{
		{
			name: "should read a basic model back",
			testData: func() testData {
				model := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: basetypes.NewStringValue("id"),
				}

				return testData{
					expectedModel: &model,
					rawState:      util.TfTypesValueFromGoTypeValue(t, model, resource_elasticsearch_project.ElasticsearchProjectResourceSchema(context.Background()).Type()),
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
			mr := elasticsearchModelReader{}
			plan := tfsdk.State{
				Raw:    td.rawState,
				Schema: resource_elasticsearch_project.ElasticsearchProjectResourceSchema(context.Background()),
			}

			model, diags := mr.ReadFrom(context.Background(), plan)
			require.False(t, diags.HasError())
			require.Equal(t, td.expectedModel, model)
		})
	}
}

func TestElasticsearchModelReader_GetID(t *testing.T) {
	mr := elasticsearchModelReader{}
	expectedId := "expected_id"
	model := resource_elasticsearch_project.ElasticsearchProjectModel{
		Id: basetypes.NewStringValue(expectedId),
	}

	require.Equal(t, expectedId, mr.GetID(model))
}

func TestElasticsearchModelReader_Modify(t *testing.T) {
	type testData struct {
		state    resource_elasticsearch_project.ElasticsearchProjectModel
		plan     resource_elasticsearch_project.ElasticsearchProjectModel
		cfg      resource_elasticsearch_project.ElasticsearchProjectModel
		expected resource_elasticsearch_project.ElasticsearchProjectModel
	}
	tests := []struct {
		name     string
		testData func() testData
	}{
		{
			name: "should use state for unknown credentials",
			testData: func() testData {
				state := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: types.StringValue("state"),
				}
				state.Credentials = resource_elasticsearch_project.NewCredentialsValueMust(
					state.Credentials.AttributeTypes(context.Background()),
					map[string]attr.Value{
						"username": types.StringValue("username"),
						"password": types.StringValue("password"),
					},
				)

				return testData{
					plan: resource_elasticsearch_project.ElasticsearchProjectModel{
						Id:          types.StringValue("plan"),
						Credentials: resource_elasticsearch_project.NewCredentialsValueUnknown(),
					},
					state: state,
					expected: resource_elasticsearch_project.ElasticsearchProjectModel{
						Id:          types.StringValue("plan"),
						Credentials: state.Credentials,
					},
				}
			},
		},
		{
			name: "should use state for unknown endpoints",
			testData: func() testData {
				state := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: types.StringValue("state"),
				}
				state.Endpoints = resource_elasticsearch_project.NewEndpointsValueMust(
					state.Endpoints.AttributeTypes(context.Background()),
					map[string]attr.Value{
						"elasticsearch": basetypes.NewStringValue("es"),
						"kibana":        basetypes.NewStringValue("kibana"),
					},
				)

				return testData{
					plan: resource_elasticsearch_project.ElasticsearchProjectModel{
						Id:        types.StringValue("plan"),
						Endpoints: resource_elasticsearch_project.NewEndpointsValueUnknown(),
					},
					state: state,
					expected: resource_elasticsearch_project.ElasticsearchProjectModel{
						Id:        types.StringValue("plan"),
						Endpoints: state.Endpoints,
					},
				}
			},
		},
		{
			name: "should use state for unknown metadata",
			testData: func() testData {
				state := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: types.StringValue("state"),
				}
				state.Metadata = resource_elasticsearch_project.NewMetadataValueMust(
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
					plan: resource_elasticsearch_project.ElasticsearchProjectModel{
						Id:       types.StringValue("plan"),
						Metadata: resource_elasticsearch_project.NewMetadataValueUnknown(),
					},
					state: state,
					expected: resource_elasticsearch_project.ElasticsearchProjectModel{
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
					plan: resource_elasticsearch_project.ElasticsearchProjectModel{
						Id:    types.StringValue("plan"),
						Name:  types.StringValue("planned name"),
						Alias: types.StringValue("alias"),
					},
					state: resource_elasticsearch_project.ElasticsearchProjectModel{
						Id:    types.StringValue("state"),
						Name:  types.StringValue("state name"),
						Alias: types.StringValue("alias"),
					},
					cfg: resource_elasticsearch_project.ElasticsearchProjectModel{
						Alias: types.StringValue("alias"),
					},
					expected: resource_elasticsearch_project.ElasticsearchProjectModel{
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
					plan: resource_elasticsearch_project.ElasticsearchProjectModel{
						Id:    types.StringValue("plan"),
						Name:  types.StringValue("name"),
						Alias: types.StringValue("planned alias"),
					},
					state: resource_elasticsearch_project.ElasticsearchProjectModel{
						Id:    types.StringValue("state"),
						Name:  types.StringValue("name"),
						Alias: types.StringValue("state alias"),
					},
					expected: resource_elasticsearch_project.ElasticsearchProjectModel{
						Id:        types.StringValue("plan"),
						Name:      types.StringValue("name"),
						Alias:     types.StringValue("planned alias"),
						CloudId:   types.StringUnknown(),
						Endpoints: resource_elasticsearch_project.NewEndpointsValueUnknown(),
					},
				}
			},
		},
		{
			name: "cloud id, alias, and endpoints should be unknown if name has changed but alias is not configured",
			testData: func() testData {
				return testData{
					plan: resource_elasticsearch_project.ElasticsearchProjectModel{
						Id:   types.StringValue("plan"),
						Name: types.StringValue("planned name"),
					},
					state: resource_elasticsearch_project.ElasticsearchProjectModel{
						Id:   types.StringValue("state"),
						Name: types.StringValue("state name"),
					},
					expected: resource_elasticsearch_project.ElasticsearchProjectModel{
						Id:        types.StringValue("plan"),
						Name:      types.StringValue("planned name"),
						CloudId:   types.StringUnknown(),
						Alias:     types.StringUnknown(),
						Endpoints: resource_elasticsearch_project.NewEndpointsValueUnknown(),
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := tt.testData()
			mr := elasticsearchModelReader{}

			require.Equal(t, td.expected, mr.Modify(td.plan, td.state, td.cfg))
		})
	}
}

func TestElasticsearchApi_Ready(t *testing.T) {
	t.Run("is ready when the client is configured", func(t *testing.T) {
		api := elasticsearchApi{client: &serverless.ClientWithResponses{}}
		require.True(t, api.Ready())
	})
	t.Run("is not ready when the client is not configured", func(t *testing.T) {
		api := elasticsearchApi{}
		require.False(t, api.Ready())
	})
}

func TestElasticsearchApi_WithClient(t *testing.T) {
	var api api[resource_elasticsearch_project.ElasticsearchProjectModel] = elasticsearchApi{}

	require.False(t, api.Ready())

	api = api.WithClient(&serverless.ClientWithResponses{})
	require.True(t, api.Ready())
}

func TestElasticsearchApi_Create(t *testing.T) {
	ctrl := gomock.NewController(t)

	type testData struct {
		client        serverless.ClientWithResponsesInterface
		initialModel  resource_elasticsearch_project.ElasticsearchProjectModel
		expectedModel resource_elasticsearch_project.ElasticsearchProjectModel
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
				mockApiClient.EXPECT().CreateElasticsearchProjectWithResponse(ctx, gomock.Any()).Return(
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
				failedResponse := &serverless.CreateElasticsearchProjectResponse{
					HTTPResponse: &http.Response{
						Status:     "failed",
						StatusCode: 400,
					},
					Body: []byte("api call failed"),
				}
				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().CreateElasticsearchProjectWithResponse(ctx, gomock.Any()).Return(
					failedResponse,
					nil,
				)

				return testData{
					client: mockApiClient,
					expectedDiags: diag.Diagnostics{
						diag.NewErrorDiagnostic(
							"Failed to create elasticsearch_project",
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
				initialModel := resource_elasticsearch_project.ElasticsearchProjectModel{
					Name:     types.StringValue("project name"),
					RegionId: types.StringValue("nether region"),
				}
				createdProject := serverless.ElasticsearchProjectCreated{
					Id: "created id",
					Credentials: serverless.ProjectCredentials{
						Username: "project username",
						Password: "sekret",
					},
				}
				expectedProject := initialModel
				expectedProject.Id = types.StringValue(createdProject.Id)
				expectedProject.Credentials = resource_elasticsearch_project.NewCredentialsValueMust(
					initialModel.Credentials.AttributeTypes(ctx),
					map[string]attr.Value{
						"username": types.StringValue(createdProject.Credentials.Username),
						"password": types.StringValue(createdProject.Credentials.Password),
					},
				)
				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().CreateElasticsearchProjectWithResponse(ctx, serverless.CreateElasticsearchProjectRequest{
					Name:     initialModel.Name.ValueString(),
					RegionId: initialModel.RegionId.ValueString(),
				}).Return(
					&serverless.CreateElasticsearchProjectResponse{
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
				initialModel := resource_elasticsearch_project.ElasticsearchProjectModel{
					Name:         types.StringValue("project name"),
					RegionId:     types.StringValue("nether region"),
					Alias:        types.StringValue("project alias"),
					OptimizedFor: types.StringValue("general_purpose"),
				}
				initialModel.SearchLake = resource_elasticsearch_project.NewSearchLakeValueMust(
					initialModel.SearchLake.AttributeTypes(ctx),
					map[string]attr.Value{
						"boost_window": types.Int64Value(20),
						"search_power": types.Int64Value(60),
					},
				)

				createdProject := serverless.ElasticsearchProjectCreated{
					Id: "created id",
					Credentials: serverless.ProjectCredentials{
						Username: "project username",
						Password: "sekret",
					},
				}
				expectedProject := initialModel
				expectedProject.Id = types.StringValue(createdProject.Id)
				expectedProject.Credentials = resource_elasticsearch_project.NewCredentialsValueMust(
					initialModel.Credentials.AttributeTypes(ctx),
					map[string]attr.Value{
						"username": types.StringValue(createdProject.Credentials.Username),
						"password": types.StringValue(createdProject.Credentials.Password),
					},
				)

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().CreateElasticsearchProjectWithResponse(ctx, serverless.CreateElasticsearchProjectRequest{
					Name:         initialModel.Name.ValueString(),
					RegionId:     initialModel.RegionId.ValueString(),
					Alias:        initialModel.Alias.ValueStringPointer(),
					OptimizedFor: (*serverless.ElasticsearchOptimizedFor)(initialModel.OptimizedFor.ValueStringPointer()),
					SearchLake: &serverless.ElasticsearchSearchLake{
						BoostWindow: util.Ptr(int(initialModel.SearchLake.BoostWindow.ValueInt64())),
						SearchPower: util.Ptr(int(initialModel.SearchLake.SearchPower.ValueInt64())),
					},
				}).Return(
					&serverless.CreateElasticsearchProjectResponse{
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

			api := elasticsearchApi{}.WithClient(td.client)
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

func TestElasticsearchApi_Patch(t *testing.T) {
	ctrl := gomock.NewController(t)

	type testData struct {
		client        serverless.ClientWithResponsesInterface
		model         resource_elasticsearch_project.ElasticsearchProjectModel
		expectedDiags diag.Diagnostics
	}
	tests := []struct {
		name     string
		testData func(ctx context.Context) testData
	}{
		{
			name: "should fail when the api returns an error",
			testData: func(ctx context.Context) testData {
				model := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id:   types.StringValue("project id"),
					Name: types.StringValue("project name"),
				}
				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().PatchElasticsearchProjectWithResponse(ctx, model.Id.ValueString(), nil, serverless.PatchElasticsearchProjectRequest{
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
				model := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id:   types.StringValue("project id"),
					Name: types.StringValue("project name"),
				}
				failedResponse := &serverless.PatchElasticsearchProjectResponse{
					HTTPResponse: &http.Response{
						Status:     "failed",
						StatusCode: 400,
					},
					Body: []byte("api call failed"),
				}
				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().PatchElasticsearchProjectWithResponse(ctx, model.Id.ValueString(), nil, serverless.PatchElasticsearchProjectRequest{
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
							"Failed to update elasticsearch_project",
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
				model := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id:       types.StringValue("project id"),
					Name:     types.StringValue("project name"),
					RegionId: types.StringValue("nether region"),
				}
				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().PatchElasticsearchProjectWithResponse(ctx, model.Id.ValueString(), nil, serverless.PatchElasticsearchProjectRequest{
					Name: model.Name.ValueStringPointer(),
				}).Return(
					&serverless.PatchElasticsearchProjectResponse{
						JSON200: &serverless.ElasticsearchProject{},
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
				model := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id:           types.StringValue("project id"),
					Name:         types.StringValue("project name"),
					RegionId:     types.StringValue("nether region"),
					Alias:        types.StringValue("project alias"),
					OptimizedFor: types.StringValue("general_purpose"),
				}
				model.SearchLake = resource_elasticsearch_project.NewSearchLakeValueMust(
					model.SearchLake.AttributeTypes(ctx),
					map[string]attr.Value{
						"boost_window": types.Int64Value(20),
						"search_power": types.Int64Value(60),
					},
				)

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().PatchElasticsearchProjectWithResponse(ctx, model.Id.ValueString(), nil, serverless.PatchElasticsearchProjectRequest{
					Name:  model.Name.ValueStringPointer(),
					Alias: model.Alias.ValueStringPointer(),
					SearchLake: &serverless.OptionalElasticsearchSearchLake{
						BoostWindow: util.Ptr(int(model.SearchLake.BoostWindow.ValueInt64())),
						SearchPower: util.Ptr(int(model.SearchLake.SearchPower.ValueInt64())),
					},
				}).Return(
					&serverless.PatchElasticsearchProjectResponse{
						JSON200: &serverless.ElasticsearchProject{},
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

			api := elasticsearchApi{}.WithClient(td.client)
			diags := api.Patch(ctx, td.model)

			if td.expectedDiags != nil {
				require.Equal(t, td.expectedDiags, diags)
			} else {
				require.False(t, diags.HasError())
			}
		})
	}
}

type fakeSleeper struct{}

func (f fakeSleeper) Sleep(d time.Duration) {}

func TestElasticsearchApi_EnsureInitialised(t *testing.T) {
	ctrl := gomock.NewController(t)
	type testData struct {
		client        serverless.ClientWithResponsesInterface
		model         resource_elasticsearch_project.ElasticsearchProjectModel
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
				model := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: types.StringValue("project id"),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().GetElasticsearchProjectStatusWithResponse(ctx, model.Id.ValueString()).DoAndReturn(
					func(_ context.Context, id string, _ ...serverless.RequestEditorFn) (*serverless.GetElasticsearchProjectStatusResponse, error) {
						if callsBeforeInitialised > 0 {
							callsBeforeInitialised--
							return &serverless.GetElasticsearchProjectStatusResponse{
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
				model := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: types.StringValue("project id"),
				}

				failedResponse := &serverless.GetElasticsearchProjectStatusResponse{
					HTTPResponse: &http.Response{
						Status:     "failed",
						StatusCode: 400,
					},
					Body: []byte("api call failed"),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().GetElasticsearchProjectStatusWithResponse(ctx, model.Id.ValueString()).DoAndReturn(
					func(_ context.Context, id string, _ ...serverless.RequestEditorFn) (*serverless.GetElasticsearchProjectStatusResponse, error) {
						if callsBeforeInitialised > 0 {
							callsBeforeInitialised--
							return &serverless.GetElasticsearchProjectStatusResponse{
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
							"Failed to get elasticsearch_project status",
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
				model := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: types.StringValue("project id"),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().GetElasticsearchProjectStatusWithResponse(ctx, model.Id.ValueString()).DoAndReturn(
					func(_ context.Context, id string, _ ...serverless.RequestEditorFn) (*serverless.GetElasticsearchProjectStatusResponse, error) {
						phase := serverless.Initialized

						if callsBeforeInitialised > 0 {
							callsBeforeInitialised--
							phase = serverless.Initializing
						}

						return &serverless.GetElasticsearchProjectStatusResponse{
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
			api := elasticsearchApi{sleeper: fakeSleeper{}}.WithClient(td.client)

			diags := api.EnsureInitialised(ctx, td.model)
			if td.expectedDiags != nil {
				require.Equal(t, td.expectedDiags, diags)
			} else {
				require.Nil(t, diags)
			}
		})
	}
}

func TestElasticsearchApi_Read(t *testing.T) {
	ctrl := gomock.NewController(t)

	type testData struct {
		client        serverless.ClientWithResponsesInterface
		id            string
		initialModel  resource_elasticsearch_project.ElasticsearchProjectModel
		expectedModel resource_elasticsearch_project.ElasticsearchProjectModel
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
				initialModel := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: types.StringValue(id),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().GetElasticsearchProjectWithResponse(ctx, id).Return(nil, assert.AnError)

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
				initialModel := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: types.StringValue(id),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().
					GetElasticsearchProjectWithResponse(ctx, id).
					Return(&serverless.GetElasticsearchProjectResponse{
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
				initialModel := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: types.StringValue(id),
				}

				failedResponse := &serverless.GetElasticsearchProjectResponse{
					HTTPResponse: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     "nope",
					},
					Body: []byte("failed"),
				}
				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().
					GetElasticsearchProjectWithResponse(ctx, id).
					Return(failedResponse, nil)

				return testData{
					client:        mockApiClient,
					id:            id,
					initialModel:  initialModel,
					expectedModel: initialModel,
					expectedDiags: diag.Diagnostics{
						diag.NewErrorDiagnostic(
							"Failed to read elasticsearch_project",
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
				initialModel := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: types.StringValue(id),
				}

				readModel := &serverless.ElasticsearchProject{
					Id:      id,
					Alias:   "expected-alias-" + id[0:6],
					CloudId: "cloud-id",
					Endpoints: serverless.ElasticsearchProjectEndpoints{
						Elasticsearch: "es-endpoint",
						Kibana:        "kib-endpoint",
					},
					Metadata: serverless.ProjectMetadata{
						CreatedAt:      time.Now(),
						CreatedBy:      "me",
						OrganizationId: "1",
					},
					Name:         "project-name",
					OptimizedFor: "general_purpose",
					RegionId:     "nether",
					Type:         "elasticsearch",
				}

				expectedModel := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id:      types.StringValue(id),
					Alias:   types.StringValue("expected-alias"),
					CloudId: types.StringValue(readModel.CloudId),
					Endpoints: resource_elasticsearch_project.NewEndpointsValueMust(
						initialModel.Endpoints.AttributeTypes(ctx),
						map[string]attr.Value{
							"elasticsearch": basetypes.NewStringValue(readModel.Endpoints.Elasticsearch),
							"kibana":        basetypes.NewStringValue(readModel.Endpoints.Kibana),
						},
					),
					Metadata: resource_elasticsearch_project.NewMetadataValueMust(
						initialModel.Metadata.AttributeTypes(ctx),
						map[string]attr.Value{
							"created_at":       basetypes.NewStringValue(readModel.Metadata.CreatedAt.String()),
							"created_by":       basetypes.NewStringValue(readModel.Metadata.CreatedBy),
							"organization_id":  basetypes.NewStringValue(readModel.Metadata.OrganizationId),
							"suspended_at":     basetypes.NewStringNull(),
							"suspended_reason": basetypes.NewStringNull(),
						},
					),
					SearchLake: resource_elasticsearch_project.NewSearchLakeValueMust(
						initialModel.SearchLake.AttributeTypes(ctx),
						map[string]attr.Value{
							"boost_window": basetypes.NewInt64Null(),
							"search_power": basetypes.NewInt64Null(),
						},
					),
					Name:         types.StringValue(readModel.Name),
					OptimizedFor: types.StringValue(string(readModel.OptimizedFor)),
					RegionId:     types.StringValue(readModel.RegionId),
					Type:         types.StringValue(string(readModel.Type)),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().
					GetElasticsearchProjectWithResponse(ctx, id).
					Return(&serverless.GetElasticsearchProjectResponse{
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
				initialModel := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: types.StringValue(id),
				}

				now := time.Now()
				readModel := &serverless.ElasticsearchProject{
					Id:      id,
					Alias:   "expected-alias-" + id[0:6],
					CloudId: "cloud-id",
					Endpoints: serverless.ElasticsearchProjectEndpoints{
						Elasticsearch: "es-endpoint",
						Kibana:        "kib-endpoint",
					},
					Metadata: serverless.ProjectMetadata{
						CreatedAt:       now,
						CreatedBy:       "me",
						OrganizationId:  "1",
						SuspendedAt:     util.Ptr(now),
						SuspendedReason: util.Ptr("meh"),
					},
					SearchLake: &serverless.ElasticsearchSearchLake{
						BoostWindow: util.Ptr(20),
						SearchPower: util.Ptr(30),
					},
					Name:         "project-name",
					OptimizedFor: "general_purpose",
					RegionId:     "nether",
					Type:         "elasticsearch",
				}

				expectedModel := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id:      types.StringValue(id),
					Alias:   types.StringValue("expected-alias"),
					CloudId: types.StringValue(readModel.CloudId),
					Endpoints: resource_elasticsearch_project.NewEndpointsValueMust(
						initialModel.Endpoints.AttributeTypes(ctx),
						map[string]attr.Value{
							"elasticsearch": basetypes.NewStringValue(readModel.Endpoints.Elasticsearch),
							"kibana":        basetypes.NewStringValue(readModel.Endpoints.Kibana),
						},
					),
					Metadata: resource_elasticsearch_project.NewMetadataValueMust(
						initialModel.Metadata.AttributeTypes(ctx),
						map[string]attr.Value{
							"created_at":       basetypes.NewStringValue(readModel.Metadata.CreatedAt.String()),
							"created_by":       basetypes.NewStringValue(readModel.Metadata.CreatedBy),
							"organization_id":  basetypes.NewStringValue(readModel.Metadata.OrganizationId),
							"suspended_at":     basetypes.NewStringValue(now.String()),
							"suspended_reason": basetypes.NewStringValue(*readModel.Metadata.SuspendedReason),
						},
					),
					SearchLake: resource_elasticsearch_project.NewSearchLakeValueMust(
						initialModel.SearchLake.AttributeTypes(ctx),
						map[string]attr.Value{
							"boost_window": basetypes.NewInt64Value(int64(*readModel.SearchLake.BoostWindow)),
							"search_power": basetypes.NewInt64Value(int64(*readModel.SearchLake.SearchPower)),
						},
					),
					Name:         types.StringValue(readModel.Name),
					OptimizedFor: types.StringValue(string(readModel.OptimizedFor)),
					RegionId:     types.StringValue(readModel.RegionId),
					Type:         types.StringValue(string(readModel.Type)),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().
					GetElasticsearchProjectWithResponse(ctx, id).
					Return(&serverless.GetElasticsearchProjectResponse{
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

			api := elasticsearchApi{}.WithClient(td.client)
			found, model, diags := api.Read(ctx, td.id, td.initialModel)

			require.Equal(t, td.expectedFound, found)
			require.Equal(t, td.expectedModel, model)
			require.Equal(t, td.expectedDiags, diags)
		})
	}
}

func TestElasticsearchApi_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	type testData struct {
		client        serverless.ClientWithResponsesInterface
		model         resource_elasticsearch_project.ElasticsearchProjectModel
		expectedDiags diag.Diagnostics
	}
	tests := []struct {
		name     string
		testData func(ctx context.Context) testData
	}{
		{
			name: "should error if delete errors",
			testData: func(ctx context.Context) testData {
				model := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: types.StringValue("project id"),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().
					DeleteElasticsearchProjectWithResponse(ctx, model.Id.ValueString(), nil).
					Return(nil, assert.AnError)

				return testData{
					client: mockApiClient,
					model:  model,
					expectedDiags: diag.Diagnostics{
						diag.NewErrorDiagnostic("Failed to delete elasticsearch_project", assert.AnError.Error()),
					},
				}
			},
		},
		{
			name: "should error if delete returns a non-200 and non-404 response",
			testData: func(ctx context.Context) testData {
				model := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: types.StringValue("project id"),
				}

				failedResponse := &serverless.DeleteElasticsearchProjectResponse{
					HTTPResponse: &http.Response{
						Status:     "failed",
						StatusCode: 400,
					},
					Body: []byte("api call failed"),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().
					DeleteElasticsearchProjectWithResponse(ctx, model.Id.ValueString(), nil).
					Return(failedResponse, nil)

				return testData{
					client: mockApiClient,
					model:  model,
					expectedDiags: diag.Diagnostics{

						diag.NewErrorDiagnostic(
							"Request to delete elasticsearch_project failed",
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
				model := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: types.StringValue("project id"),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().
					DeleteElasticsearchProjectWithResponse(ctx, model.Id.ValueString(), nil).
					Return(&serverless.DeleteElasticsearchProjectResponse{
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
				model := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: types.StringValue("project id"),
				}

				mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)
				mockApiClient.EXPECT().
					DeleteElasticsearchProjectWithResponse(ctx, model.Id.ValueString(), nil).
					Return(&serverless.DeleteElasticsearchProjectResponse{
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
			api := elasticsearchApi{sleeper: fakeSleeper{}}.WithClient(td.client)

			diags := api.Delete(ctx, td.model)
			require.Equal(t, td.expectedDiags, diags)
		})
	}
}
