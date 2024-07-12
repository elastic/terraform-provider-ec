package projectresource

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_elasticsearch_project"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Run("should fail if reading the tf model errors", func(t *testing.T) {
		ctx := context.Background()
		req := resource.DeleteRequest{
			State: tfsdk.State{
				Raw: tftypes.NewValue(tftypes.Bool, true),
			},
		}

		readDiags := diag.Diagnostics{
			diag.NewErrorDiagnostic("nope", "nope"),
		}

		api := NewMockapi[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
		api.EXPECT().Ready().Return(true)

		handler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
		handler.EXPECT().ReadFrom(ctx, req.State).Return(nil, readDiags)

		r := Resource[resource_elasticsearch_project.ElasticsearchProjectModel]{
			api:          api,
			modelHandler: handler,
		}

		res := resource.DeleteResponse{}
		r.Delete(ctx, req, &res)

		require.Equal(t, readDiags, res.Diagnostics)
	})
	t.Run("should fail if the delete api call fails", func(t *testing.T) {
		ctx := context.Background()
		req := resource.DeleteRequest{
			State: tfsdk.State{
				Raw: tftypes.NewValue(tftypes.Bool, true),
			},
		}

		deleteDiags := diag.Diagnostics{
			diag.NewErrorDiagnostic("nope", "nope"),
		}

		model := resource_elasticsearch_project.ElasticsearchProjectModel{
			Id: basetypes.NewStringValue("id"),
		}

		api := NewMockapi[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
		api.EXPECT().Ready().Return(true)
		api.EXPECT().Delete(ctx, model).Return(deleteDiags)

		handler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
		handler.EXPECT().ReadFrom(ctx, req.State).Return(&model, nil)

		r := Resource[resource_elasticsearch_project.ElasticsearchProjectModel]{
			api:          api,
			modelHandler: handler,
		}

		res := resource.DeleteResponse{}
		r.Delete(ctx, req, &res)

		require.Equal(t, deleteDiags, res.Diagnostics)
	})
	t.Run("should remove the deleted project from state", func(t *testing.T) {
		ctx := context.Background()
		req := resource.DeleteRequest{
			State: tfsdk.State{
				Raw: tftypes.NewValue(tftypes.Bool, true),
			},
		}

		model := resource_elasticsearch_project.ElasticsearchProjectModel{
			Id: basetypes.NewStringValue("id"),
		}

		api := NewMockapi[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
		api.EXPECT().Ready().Return(true)
		api.EXPECT().Delete(ctx, model).Return(nil)

		handler := NewMockmodelHandler[resource_elasticsearch_project.ElasticsearchProjectModel](ctrl)
		handler.EXPECT().ReadFrom(ctx, req.State).Return(&model, nil)

		r := Resource[resource_elasticsearch_project.ElasticsearchProjectModel]{
			api:          api,
			modelHandler: handler,
		}

		res := resource.DeleteResponse{
			State: tfsdk.State{
				Raw:    tftypes.NewValue(tftypes.Bool, true),
				Schema: resource_elasticsearch_project.ElasticsearchProjectResourceSchema(ctx),
			},
		}
		r.Delete(ctx, req, &res)

		require.Nil(t, res.Diagnostics)
		require.True(t, res.State.Raw.IsNull())
	})
}
