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

func TestRead(t *testing.T) {
	ctrl := gomock.NewController(t)
	type testData struct {
		modelHandler        modelHandler[resource_elasticsearch_project.ElasticsearchProjectModel]
		api                 api[resource_elasticsearch_project.ElasticsearchProjectModel]
		req                 resource.ReadRequest
		expectedDiags       diag.Diagnostics
		expectStateMutation bool
		expectNullState     bool
		expectedId          *string
	}
	tests := []struct {
		name     string
		testData func(context.Context) testData
	}{
		{
			name: "should fail if reading the tf model errors",
			testData: func(ctx context.Context) testData {
				req := resource.ReadRequest{
					State: tfsdk.State{
						Raw: tftypes.NewValue(tftypes.Bool, true),
					},
				}

				readDiags := diag.Diagnostics{
					diag.NewErrorDiagnostic("nope", "nope"),
				}

				handler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				handler.EXPECT().ReadFrom(ctx, req.State).Return(nil, readDiags)

				api := NewMockapi[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				api.EXPECT().Ready().Return(true)

				return testData{
					modelHandler:  handler,
					req:           req,
					api:           api,
					expectedDiags: readDiags,
				}
			},
		},
		{
			name: "should fail if reading the project from the api errors",
			testData: func(ctx context.Context) testData {
				req := resource.ReadRequest{
					State: tfsdk.State{
						Raw: tftypes.NewValue(tftypes.Bool, true),
					},
				}

				model := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: basetypes.NewStringValue("id"),
				}

				handler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				handler.EXPECT().ReadFrom(ctx, req.State).Return(&model, nil)
				handler.EXPECT().GetID(model).Return(model.Id.ValueString())

				readDiags := diag.Diagnostics{
					diag.NewErrorDiagnostic("nope", "nope"),
				}

				api := NewMockapi[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				api.EXPECT().Ready().Return(true)
				api.EXPECT().Read(ctx, model.Id.ValueString(), model).Return(false, model, readDiags)

				return testData{
					modelHandler:  handler,
					req:           req,
					api:           api,
					expectedDiags: readDiags,
				}
			},
		},
		{
			name: "should remove the resource from state if it's not found in the api",
			testData: func(ctx context.Context) testData {
				req := resource.ReadRequest{
					State: tfsdk.State{
						Raw: tftypes.NewValue(tftypes.Bool, true),
					},
				}

				model := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: basetypes.NewStringValue("id"),
				}

				handler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				handler.EXPECT().ReadFrom(ctx, req.State).Return(&model, nil)
				handler.EXPECT().GetID(model).Return(model.Id.ValueString())

				api := NewMockapi[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				api.EXPECT().Ready().Return(true)
				api.EXPECT().Read(ctx, model.Id.ValueString(), model).Return(false, model, nil)

				return testData{
					modelHandler:        handler,
					req:                 req,
					api:                 api,
					expectStateMutation: true,
					expectNullState:     true,
					expectedId:          nil,
				}
			},
		},
		{
			name: "should update state with the model returned by the api",
			testData: func(ctx context.Context) testData {
				req := resource.ReadRequest{
					State: tfsdk.State{
						Raw: tftypes.NewValue(tftypes.Bool, true),
					},
				}

				model := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: basetypes.NewStringValue("id"),
				}

				handler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				handler.EXPECT().ReadFrom(ctx, req.State).Return(&model, nil)
				handler.EXPECT().GetID(model).Return(model.Id.ValueString())

				api := NewMockapi[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				api.EXPECT().Ready().Return(true)
				api.EXPECT().Read(ctx, model.Id.ValueString(), model).Return(true, model, nil)

				return testData{
					modelHandler:        handler,
					req:                 req,
					api:                 api,
					expectStateMutation: true,
					expectNullState:     false,
					expectedId:          model.Id.ValueStringPointer(),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			td := tt.testData(ctx)

			r := Resource[resource_elasticsearch_project.ElasticsearchProjectModel]{
				modelHandler: td.modelHandler,
				api:          td.api,
			}

			res := resource.ReadResponse{
				State: tfsdk.State{
					Schema: resource_elasticsearch_project.ElasticsearchProjectResourceSchema(ctx),
					Raw:    tftypes.NewValue(tftypes.Bool, true),
				},
			}
			r.Read(ctx, td.req, &res)

			require.Equal(t, td.expectedDiags, res.Diagnostics)

			if td.expectStateMutation {
				require.NotEqual(t, tftypes.Bool, res.State.Raw.Type())
				require.Equal(t, td.expectNullState, res.State.Raw.IsNull())

				var id *string
				res.State.GetAttribute(ctx, path.Root("id"), &id)
				require.Equal(t, td.expectedId, id)
			}
		})
	}
}
