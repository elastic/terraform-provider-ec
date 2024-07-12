package projectresource

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_elasticsearch_project"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	type testData struct {
		api           api[resource_elasticsearch_project.ElasticsearchProjectModel]
		modelHandler  modelHandler[resource_elasticsearch_project.ElasticsearchProjectModel]
		req           resource.CreateRequest
		expectedDiags diag.Diagnostics
		expectedId    *string
	}
	tests := []struct {
		name     string
		testData func(context.Context) testData
	}{
		{
			name: "should error if reading the tf model errors",
			testData: func(ctx context.Context) testData {
				req := resource.CreateRequest{
					Plan: tfsdk.Plan{
						Raw: tftypes.NewValue(tftypes.Bool, true),
					},
				}

				api := NewMockapi[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				api.EXPECT().Ready().Return(true)

				readDiags := diag.Diagnostics{
					diag.NewErrorDiagnostic("nope", "nope"),
				}
				handler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				handler.EXPECT().ReadFrom(ctx, req.Plan).Return(nil, readDiags)

				return testData{
					api:           api,
					modelHandler:  handler,
					req:           req,
					expectedDiags: readDiags,
				}
			},
		},
		{
			name: "should noop if read returns an empty model",
			testData: func(ctx context.Context) testData {
				req := resource.CreateRequest{
					Plan: tfsdk.Plan{
						Raw: tftypes.NewValue(tftypes.Bool, true),
					},
				}

				api := NewMockapi[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				api.EXPECT().Ready().Return(true)

				handler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				handler.EXPECT().ReadFrom(ctx, req.Plan).Return(nil, nil)

				return testData{
					api:          api,
					modelHandler: handler,
					req:          req,
				}
			},
		},
		{
			name: "should set id in state, but ultimately fail if the create call fails, but returns a non-empty model",
			testData: func(ctx context.Context) testData {
				req := resource.CreateRequest{
					Plan: tfsdk.Plan{
						Raw: tftypes.NewValue(tftypes.Bool, true),
					},
				}

				createDiags := diag.Diagnostics{
					diag.NewErrorDiagnostic("nope", "nope"),
				}

				readModel := resource_elasticsearch_project.ElasticsearchProjectModel{
					Name: basetypes.NewStringValue("name"),
				}
				createdModel := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id:   basetypes.NewStringValue("id"),
					Name: basetypes.NewStringValue("name"),
				}

				api := NewMockapi[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				api.EXPECT().Ready().Return(true)
				api.EXPECT().Create(ctx, readModel).Return(createdModel, createDiags)

				handler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				handler.EXPECT().ReadFrom(ctx, req.Plan).Return(&readModel, nil)
				handler.EXPECT().GetID(createdModel).Return(createdModel.Id.ValueString())

				return testData{
					api:           api,
					modelHandler:  handler,
					req:           req,
					expectedDiags: createDiags,
					expectedId:    createdModel.Id.ValueStringPointer(),
				}
			},
		},
		{
			name: "should set id in state, but ultimately fail if initialising fails",
			testData: func(ctx context.Context) testData {
				req := resource.CreateRequest{
					Plan: tfsdk.Plan{
						Raw: tftypes.NewValue(tftypes.Bool, true),
					},
				}

				initDiags := diag.Diagnostics{
					diag.NewErrorDiagnostic("nope", "nope"),
				}

				readModel := resource_elasticsearch_project.ElasticsearchProjectModel{
					Name: basetypes.NewStringValue("name"),
				}
				createdModel := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id:   basetypes.NewStringValue("id"),
					Name: basetypes.NewStringValue("name"),
				}

				api := NewMockapi[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				api.EXPECT().Ready().Return(true)
				api.EXPECT().Create(ctx, readModel).Return(createdModel, nil)
				api.EXPECT().EnsureInitialised(ctx, createdModel).Return(initDiags)

				handler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				handler.EXPECT().ReadFrom(ctx, req.Plan).Return(&readModel, nil)
				handler.EXPECT().GetID(createdModel).Return(createdModel.Id.ValueString())

				return testData{
					api:           api,
					modelHandler:  handler,
					req:           req,
					expectedDiags: initDiags,
					expectedId:    createdModel.Id.ValueStringPointer(),
				}
			},
		},
		{
			name: "should set id in state, but ultimately fail if reading the initialised project fails",
			testData: func(ctx context.Context) testData {
				req := resource.CreateRequest{
					Plan: tfsdk.Plan{
						Raw: tftypes.NewValue(tftypes.Bool, true),
					},
				}

				readDiags := diag.Diagnostics{
					diag.NewErrorDiagnostic("nope", "nope"),
				}

				readModel := resource_elasticsearch_project.ElasticsearchProjectModel{
					Name: basetypes.NewStringValue("name"),
				}
				createdModel := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id:   basetypes.NewStringValue("id"),
					Name: basetypes.NewStringValue("name"),
				}

				api := NewMockapi[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				api.EXPECT().Ready().Return(true)
				api.EXPECT().Create(ctx, readModel).Return(createdModel, nil)
				api.EXPECT().EnsureInitialised(ctx, createdModel).Return(nil)
				api.EXPECT().Read(ctx, createdModel.Id.ValueString(), createdModel).Return(false, createdModel, readDiags)

				handler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				handler.EXPECT().ReadFrom(ctx, req.Plan).Return(&readModel, nil)
				handler.EXPECT().GetID(createdModel).Return(createdModel.Id.ValueString()).AnyTimes()

				return testData{
					api:           api,
					modelHandler:  handler,
					req:           req,
					expectedDiags: readDiags,
					expectedId:    createdModel.Id.ValueStringPointer(),
				}
			},
		},
		{
			name: "should set id in state, but ultimately fail if reading the initialised project returns an empty model",
			testData: func(ctx context.Context) testData {
				req := resource.CreateRequest{
					Plan: tfsdk.Plan{
						Raw: tftypes.NewValue(tftypes.Bool, true),
					},
				}

				readModel := resource_elasticsearch_project.ElasticsearchProjectModel{
					Name: basetypes.NewStringValue("name"),
				}
				createdModel := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id:   basetypes.NewStringValue("id"),
					Name: basetypes.NewStringValue("name"),
				}

				api := NewMockapi[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				api.EXPECT().Ready().Return(true)
				api.EXPECT().Create(ctx, readModel).Return(createdModel, nil)
				api.EXPECT().EnsureInitialised(ctx, createdModel).Return(nil)
				api.EXPECT().Read(ctx, createdModel.Id.ValueString(), createdModel).Return(false, createdModel, nil)

				handler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				handler.EXPECT().ReadFrom(ctx, req.Plan).Return(&readModel, nil)
				handler.EXPECT().GetID(createdModel).Return(createdModel.Id.ValueString()).AnyTimes()

				return testData{
					api:          api,
					modelHandler: handler,
					req:          req,
					expectedDiags: diag.Diagnostics{
						diag.NewErrorDiagnostic(
							"Failed to read created elasticsearch project",
							"The elasticsearch project was successfully created and initialised, but could then not be read back from the API",
						),
					},
					expectedId: createdModel.Id.ValueStringPointer(),
				}
			},
		},
		{
			name: "should set id in state, but ultimately fail if reading the initialised project returns an empty model",
			testData: func(ctx context.Context) testData {
				req := resource.CreateRequest{
					Plan: tfsdk.Plan{
						Raw: tftypes.NewValue(tftypes.Bool, true),
					},
				}

				readModel := resource_elasticsearch_project.ElasticsearchProjectModel{
					Name: basetypes.NewStringValue("name"),
				}
				createdModel := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id:   basetypes.NewStringValue("id"),
					Name: basetypes.NewStringValue("name"),
				}
				finalModel := createdModel
				finalModel.Id = basetypes.NewStringValue("final id")

				api := NewMockapi[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				api.EXPECT().Ready().Return(true)
				api.EXPECT().Create(ctx, readModel).Return(createdModel, nil)
				api.EXPECT().EnsureInitialised(ctx, createdModel).Return(nil)
				api.EXPECT().Read(ctx, createdModel.Id.ValueString(), createdModel).Return(true, finalModel, nil)

				handler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				handler.EXPECT().ReadFrom(ctx, req.Plan).Return(&readModel, nil)
				handler.EXPECT().GetID(createdModel).Return(createdModel.Id.ValueString()).AnyTimes()

				return testData{
					api:          api,
					modelHandler: handler,
					req:          req,
					expectedId:   finalModel.Id.ValueStringPointer(),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			td := tt.testData(ctx)

			res := resource.CreateResponse{
				State: tfsdk.State{
					Raw:    tftypes.NewValue(tftypes.Bool, true),
					Schema: resource_elasticsearch_project.ElasticsearchProjectResourceSchema(ctx),
				},
			}

			r := Resource[resource_elasticsearch_project.ElasticsearchProjectModel]{
				api:          td.api,
				modelHandler: td.modelHandler,
				name:         "elasticsearch",
			}

			r.Create(ctx, td.req, &res)
			require.Equal(t, td.expectedDiags, res.Diagnostics)

			var id *string
			res.State.GetAttribute(ctx, path.Root("id"), &id)
			require.Equal(t, td.expectedId, id)
		})
	}
}
