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
	"testing"

	"github.com/elastic/terraform-provider-ec/ec/internal"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/mocks"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_elasticsearch_project"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestConfigure(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockApi := NewMockapi[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
	mockApiClient := mocks.NewMockClientWithResponsesInterface(ctrl)

	mockApi.EXPECT().WithClient(gomock.Any()).Return(mockApi)
	r := Resource[resource_elasticsearch_project.ElasticsearchProjectModel]{
		api: mockApi,
	}

	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: internal.ProviderClients{
			Serverless: mockApiClient,
		},
	}, &resource.ConfigureResponse{})
}

func TestMetadata(t *testing.T) {
	r := Resource[resource_elasticsearch_project.ElasticsearchProjectModel]{
		name: "test_resource",
	}

	req := resource.MetadataRequest{
		ProviderTypeName: "test_provider",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	require.Equal(t, fmt.Sprintf("%s_%s_project", req.ProviderTypeName, r.name), resp.TypeName)
}

func TestSchema(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()
	req := resource.SchemaRequest{}
	res := resource.SchemaResponse{}

	mockHandler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
	mockHandler.EXPECT().Schema(ctx, req, &res)

	r := Resource[resource_elasticsearch_project.ElasticsearchProjectModel]{
		modelHandler: mockHandler,
	}

	r.Schema(ctx, req, &res)
}

func TestModifyPlan(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Run("should not call the plan modifier if the state model is not set", func(t *testing.T) {
		ctx := context.Background()
		req := resource.ModifyPlanRequest{
			Config: tfsdk.Config{
				Raw: tftypes.NewValue(tftypes.String, "config"),
			},
			State: tfsdk.State{
				Raw: tftypes.NewValue(tftypes.String, "state"),
			},
			Plan: tfsdk.Plan{
				Raw: tftypes.NewValue(tftypes.String, "plan"),
			},
		}
		res := resource.ModifyPlanResponse{}

		mockHandler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
		mockHandler.EXPECT().ReadFrom(ctx, req.Config).Return(&resource_elasticsearch_project.ElasticsearchProjectModel{}, nil)
		mockHandler.EXPECT().ReadFrom(ctx, req.Plan).Return(&resource_elasticsearch_project.ElasticsearchProjectModel{}, nil)
		mockHandler.EXPECT().ReadFrom(ctx, req.State).Return(nil, nil)

		r := Resource[resource_elasticsearch_project.ElasticsearchProjectModel]{
			modelHandler: mockHandler,
		}
		r.ModifyPlan(ctx, req, &res)
	})
	t.Run("should not call the plan modifier if the plan model is not set", func(t *testing.T) {
		ctx := context.Background()
		req := resource.ModifyPlanRequest{
			Config: tfsdk.Config{
				Raw: tftypes.NewValue(tftypes.String, "config"),
			},
			State: tfsdk.State{
				Raw: tftypes.NewValue(tftypes.String, "state"),
			},
			Plan: tfsdk.Plan{
				Raw: tftypes.NewValue(tftypes.String, "plan"),
			},
		}
		res := resource.ModifyPlanResponse{}

		mockHandler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
		mockHandler.EXPECT().ReadFrom(ctx, req.Config).Return(&resource_elasticsearch_project.ElasticsearchProjectModel{}, nil)
		mockHandler.EXPECT().ReadFrom(ctx, req.State).Return(&resource_elasticsearch_project.ElasticsearchProjectModel{}, nil)
		mockHandler.EXPECT().ReadFrom(ctx, req.Plan).Return(nil, nil)

		r := Resource[resource_elasticsearch_project.ElasticsearchProjectModel]{
			modelHandler: mockHandler,
		}
		r.ModifyPlan(ctx, req, &res)
	})
	t.Run("should call the plan modifier with all three models", func(t *testing.T) {
		ctx := context.Background()
		req := resource.ModifyPlanRequest{
			Config: tfsdk.Config{
				Raw: tftypes.NewValue(tftypes.String, "config"),
			},
			State: tfsdk.State{
				Raw: tftypes.NewValue(tftypes.String, "state"),
			},
			Plan: tfsdk.Plan{
				Raw: tftypes.NewValue(tftypes.String, "plan"),
			},
		}
		res := resource.ModifyPlanResponse{
			Plan: tfsdk.Plan{
				Schema: resource_elasticsearch_project.ElasticsearchProjectResourceSchema(ctx),
			},
		}

		planModel := &resource_elasticsearch_project.ElasticsearchProjectModel{
			Id: types.StringValue("plan"),
		}
		stateModel := &resource_elasticsearch_project.ElasticsearchProjectModel{
			Id: types.StringValue("state"),
		}
		cfgModel := &resource_elasticsearch_project.ElasticsearchProjectModel{
			Id: types.StringValue("config"),
		}

		mockHandler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
		mockHandler.EXPECT().ReadFrom(ctx, req.Config).Return(cfgModel, nil)
		mockHandler.EXPECT().ReadFrom(ctx, req.State).Return(stateModel, nil)
		mockHandler.EXPECT().ReadFrom(ctx, req.Plan).Return(planModel, nil)
		mockHandler.EXPECT().Modify(*planModel, *stateModel, *cfgModel).Return(*planModel)

		r := Resource[resource_elasticsearch_project.ElasticsearchProjectModel]{
			modelHandler: mockHandler,
		}
		r.ModifyPlan(ctx, req, &res)

		// Validate that the modified value was set in the response
		var id string
		res.Plan.GetAttribute(ctx, path.Root("id"), &id)
		require.Equal(t, planModel.Id.ValueString(), id)
	})
}
