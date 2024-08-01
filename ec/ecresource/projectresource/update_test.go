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

func TestUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)

	type testData struct {
		modelHandler  modelHandler[resource_elasticsearch_project.ElasticsearchProjectModel]
		api           api[resource_elasticsearch_project.ElasticsearchProjectModel]
		req           resource.UpdateRequest
		expectedDiags diag.Diagnostics
		expectedId    *string
	}
	tests := []struct {
		name     string
		testData func(context.Context) testData
	}{
		{
			name: "should error out if reading model returns an error",
			testData: func(ctx context.Context) testData {
				req := resource.UpdateRequest{
					Plan: tfsdk.Plan{
						Raw: tftypes.NewValue(tftypes.Bool, true),
					},
				}

				expectedDiags := diag.Diagnostics{
					diag.NewErrorDiagnostic("nope", "nope"),
				}

				modelHandler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				modelHandler.EXPECT().ReadFrom(ctx, req.Plan).Return(nil, expectedDiags)

				api := NewMockapi[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				api.EXPECT().Ready().Return(true)

				return testData{
					modelHandler:  modelHandler,
					api:           api,
					req:           req,
					expectedDiags: expectedDiags,
				}
			},
		},
		{
			name: "should error out if it's not possible to read the updated project",
			testData: func(ctx context.Context) testData {
				req := resource.UpdateRequest{
					Plan: tfsdk.Plan{
						Raw: tftypes.NewValue(tftypes.Bool, true),
					},
				}

				model := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: basetypes.NewStringValue("project id"),
				}

				modelHandler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				modelHandler.EXPECT().ReadFrom(ctx, req.Plan).Return(&model, nil)
				modelHandler.EXPECT().GetID(model).Return(model.Id.ValueString())

				api := NewMockapi[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				api.EXPECT().Ready().Return(true)
				api.EXPECT().Patch(ctx, model).Return(nil)
				api.EXPECT().Read(ctx, model.Id.ValueString(), model).Return(false, model, nil)

				return testData{
					modelHandler: modelHandler,
					api:          api,
					req:          req,
					expectedDiags: diag.Diagnostics{
						diag.NewErrorDiagnostic(
							"Failed to read updated elasticsearch project",
							"The elasticsearch project was successfully updated, but could then not be read back from the API",
						),
					},
				}
			},
		},
		{
			name: "should update state with the patched project when successful",
			testData: func(ctx context.Context) testData {
				req := resource.UpdateRequest{
					Plan: tfsdk.Plan{
						Raw: tftypes.NewValue(tftypes.Bool, true),
					},
				}

				model := resource_elasticsearch_project.ElasticsearchProjectModel{
					Id: basetypes.NewStringValue("project id"),
				}
				readModel := model
				readModel.Id = basetypes.NewStringValue("updated project id")

				modelHandler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				modelHandler.EXPECT().ReadFrom(ctx, req.Plan).Return(&model, nil)
				modelHandler.EXPECT().GetID(model).Return(model.Id.ValueString())

				api := NewMockapi[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
				api.EXPECT().Ready().Return(true)
				api.EXPECT().Patch(ctx, model).Return(nil)
				api.EXPECT().Read(ctx, model.Id.ValueString(), model).Return(true, readModel, nil)

				return testData{
					modelHandler: modelHandler,
					api:          api,
					req:          req,
					expectedId:   readModel.Id.ValueStringPointer(),
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
				name:         "elasticsearch",
			}

			res := resource.UpdateResponse{
				State: tfsdk.State{
					Schema: resource_elasticsearch_project.ElasticsearchProjectResourceSchema(ctx),
				},
			}
			r.Update(ctx, td.req, &res)

			require.Equal(t, td.expectedDiags, res.Diagnostics)

			var id basetypes.StringValue
			res.State.GetAttribute(ctx, path.Root("id"), &id)
			require.Equal(t, td.expectedId, id.ValueStringPointer())
		})
	}
}
