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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/mocks"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_elasticsearch_project"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_observability_project"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_security_project"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestElasticsearchMinSearchVcuSchema(t *testing.T) {
	ctx := context.Background()

	lake, ok := resource_elasticsearch_project.ElasticsearchProjectResourceSchema(ctx).Attributes["search_lake"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	floor, ok := lake.Attributes["min_search_vcu"].(schema.Int64Attribute)
	require.True(t, ok, "min_search_vcu is an int64 attribute of search_lake")

	t.Run("is optional and not computed", func(t *testing.T) {
		require.True(t, floor.Optional)
		require.False(t, floor.Computed, "a computed floor could never be unset from configuration")
		require.False(t, floor.Required)
		require.True(t, lake.Optional)
		require.True(t, lake.Computed, "the block keeps its computed_optional shape")
	})

	validate := func(value int64) bool {
		resp := &validator.Int64Response{}
		for _, v := range floor.Validators {
			v.ValidateInt64(ctx, validator.Int64Request{
				Path:        path.Root("search_lake").AtName("min_search_vcu"),
				ConfigValue: types.Int64Value(value),
			}, resp)
		}
		return resp.Diagnostics.HasError()
	}

	t.Run("rejects values below 2 at plan time", func(t *testing.T) {
		require.Len(t, floor.Validators, 1)
		require.True(t, validate(1))
		require.True(t, validate(0))
	})

	t.Run("has no upper bound", func(t *testing.T) {
		require.False(t, validate(2))
		require.False(t, validate(5000), "the API's maximum is not enforced client-side")
	})

	t.Run("is absent on the other project resources", func(t *testing.T) {
		_, hasLake := resource_observability_project.ObservabilityProjectResourceSchema(ctx).Attributes["search_lake"]
		require.False(t, hasLake, "observability projects have no search_lake")
		secLake, ok := resource_security_project.SecurityProjectResourceSchema(ctx).Attributes["search_lake"].(schema.SingleNestedAttribute)
		require.True(t, ok)
		_, hasFloor := secLake.Attributes["min_search_vcu"]
		require.False(t, hasFloor)
	})
}

// searchLake builds a search_lake value. A nil pointer is a null attribute.
func searchLake(t *testing.T, ctx context.Context, searchPower, boostWindow, minSearchVcu *int64) resource_elasticsearch_project.SearchLakeValue {
	t.Helper()
	int64Attr := func(v *int64) attr.Value {
		if v == nil {
			return basetypes.NewInt64Null()
		}
		return basetypes.NewInt64Value(*v)
	}
	var model resource_elasticsearch_project.ElasticsearchProjectModel
	return resource_elasticsearch_project.NewSearchLakeValueMust(
		model.SearchLake.AttributeTypes(ctx),
		map[string]attr.Value{
			"boost_window":   int64Attr(boostWindow),
			"min_search_vcu": int64Attr(minSearchVcu),
			"search_power":   int64Attr(searchPower),
		},
	)
}

func minSearchVcuProject(id string, lake resource_elasticsearch_project.SearchLakeValue) resource_elasticsearch_project.ElasticsearchProjectModel {
	model := resource_elasticsearch_project.ElasticsearchProjectModel{
		Name:     types.StringValue("proj"),
		RegionId: types.StringValue("aws-us-east-1"),
	}
	if id != "" {
		model.Id = types.StringValue(id)
	}
	model.SearchLake = lake
	return model
}

// searchLakeJSON returns the raw search_lake member of a marshalled body, or
// nil when the body has no search_lake key.
func searchLakeJSON(t *testing.T, raw []byte) json.RawMessage {
	t.Helper()
	var doc map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &doc))
	return doc["search_lake"]
}

func TestElasticsearchMinSearchVcuCreate(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	created := &serverless.ElasticsearchProjectCreated{
		Id:          "projectid",
		Credentials: serverless.ProjectCredentials{Username: "u", Password: "p"},
	}

	createWith := func(t *testing.T, model resource_elasticsearch_project.ElasticsearchProjectModel) []byte {
		t.Helper()
		var sent []byte
		mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)
		mockClient.EXPECT().
			CreateElasticsearchProjectWithResponse(ctx, gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ *serverless.CreateElasticsearchProjectParams, body serverless.CreateElasticsearchProjectJSONRequestBody, _ ...serverless.RequestEditorFn) (*serverless.CreateElasticsearchProjectResponse, error) {
				var err error
				sent, err = json.Marshal(body)
				require.NoError(t, err)
				return &serverless.CreateElasticsearchProjectResponse{JSON201: created}, nil
			})
		_, diags := elasticsearchApi{client: mockClient}.Create(ctx, model)
		require.False(t, diags.HasError(), diags)
		return sent
	}

	t.Run("sends the floor next to its siblings", func(t *testing.T) {
		sent := createWith(t, minSearchVcuProject("", searchLake(t, ctx, new(int64(100)), nil, new(int64(64)))))
		require.JSONEq(t, `{"search_power":100,"min_search_vcu":64}`, string(searchLakeJSON(t, sent)))
	})

	t.Run("omits the key when the floor is null", func(t *testing.T) {
		sent := createWith(t, minSearchVcuProject("", searchLake(t, ctx, new(int64(100)), nil, nil)))
		require.JSONEq(t, `{"search_power":100}`, string(searchLakeJSON(t, sent)))
		require.NotContains(t, string(sent), "min_search_vcu")
	})

	t.Run("omits search_lake when the block is null", func(t *testing.T) {
		sent := createWith(t, minSearchVcuProject("", resource_elasticsearch_project.NewSearchLakeValueNull()))
		require.Nil(t, searchLakeJSON(t, sent))
	})

	t.Run("surfaces an API rejection with status and body", func(t *testing.T) {
		failed := &serverless.CreateElasticsearchProjectResponse{
			HTTPResponse: &http.Response{Status: "400 Bad Request", StatusCode: 400},
			Body:         []byte(`{"errors":[{"message":"number must be at most 4992"}]}`),
		}
		mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)
		mockClient.EXPECT().
			CreateElasticsearchProjectWithResponse(ctx, gomock.Any(), gomock.Any()).
			Return(failed, nil)
		_, diags := elasticsearchApi{client: mockClient}.Create(ctx, minSearchVcuProject("", searchLake(t, ctx, nil, nil, new(int64(5000)))))
		require.True(t, diags.HasError())
		require.Contains(t, diags[0].Detail(), "400 400 Bad Request")
		require.Contains(t, diags[0].Detail(), "at most 4992")
	})
}

func TestElasticsearchMinSearchVcuPatch(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	const id = "projectid"

	// patchWith runs Patch and returns the marshalled body the client received,
	// whichever client method the provider chose.
	patchWith := func(t *testing.T, plan, state resource_elasticsearch_project.ElasticsearchProjectModel) (sent []byte, raw bool) {
		t.Helper()
		mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)
		mockClient.EXPECT().
			PatchElasticsearchProjectWithResponse(ctx, id, nil, gomock.Any()).
			DoAndReturn(func(_ context.Context, _ serverless.ProjectID, _ *serverless.PatchElasticsearchProjectParams, body serverless.PatchElasticsearchProjectJSONRequestBody, _ ...serverless.RequestEditorFn) (*serverless.PatchElasticsearchProjectResponse, error) {
				var err error
				sent, err = json.Marshal(body)
				require.NoError(t, err)
				return &serverless.PatchElasticsearchProjectResponse{JSON200: &serverless.ElasticsearchProject{}}, nil
			}).MaxTimes(1)
		mockClient.EXPECT().
			PatchElasticsearchProjectWithBodyWithResponse(ctx, id, nil, "application/json", gomock.Any()).
			DoAndReturn(func(_ context.Context, _ serverless.ProjectID, _ *serverless.PatchElasticsearchProjectParams, _ string, body io.Reader, _ ...serverless.RequestEditorFn) (*serverless.PatchElasticsearchProjectResponse, error) {
				var err error
				sent, err = io.ReadAll(body)
				require.NoError(t, err)
				raw = true
				return &serverless.PatchElasticsearchProjectResponse{JSON200: &serverless.ElasticsearchProject{}}, nil
			}).MaxTimes(1)
		diags := elasticsearchApi{client: mockClient}.Patch(ctx, plan, state)
		require.False(t, diags.HasError(), diags)
		require.NotNil(t, sent, "no patch request was sent")
		return sent, raw
	}

	t.Run("raises the floor", func(t *testing.T) {
		sent, raw := patchWith(t,
			minSearchVcuProject(id, searchLake(t, ctx, new(int64(100)), nil, new(int64(128)))),
			minSearchVcuProject(id, searchLake(t, ctx, new(int64(100)), new(int64(7)), new(int64(64)))),
		)
		require.False(t, raw)
		require.JSONEq(t, `{"search_power":100,"min_search_vcu":128}`, string(searchLakeJSON(t, sent)))
	})

	t.Run("adds a floor to a project that had none", func(t *testing.T) {
		sent, raw := patchWith(t,
			minSearchVcuProject(id, searchLake(t, ctx, new(int64(100)), nil, new(int64(64)))),
			minSearchVcuProject(id, searchLake(t, ctx, new(int64(100)), new(int64(7)), nil)),
		)
		require.False(t, raw)
		require.JSONEq(t, `{"search_power":100,"min_search_vcu":64}`, string(searchLakeJSON(t, sent)))
		require.NotContains(t, string(sent), `"min_search_vcu":null`)
	})

	t.Run("omits the key when the floor was never configured", func(t *testing.T) {
		sent, raw := patchWith(t,
			minSearchVcuProject(id, searchLake(t, ctx, new(int64(200)), nil, nil)),
			minSearchVcuProject(id, searchLake(t, ctx, new(int64(100)), new(int64(7)), nil)),
		)
		require.False(t, raw)
		require.JSONEq(t, `{"search_power":200}`, string(searchLakeJSON(t, sent)))
		require.NotContains(t, string(sent), "min_search_vcu")
	})

	t.Run("removes the floor with an explicit null and keeps siblings", func(t *testing.T) {
		sent, raw := patchWith(t,
			minSearchVcuProject(id, searchLake(t, ctx, new(int64(100)), nil, nil)),
			minSearchVcuProject(id, searchLake(t, ctx, new(int64(100)), new(int64(7)), new(int64(64)))),
		)
		require.True(t, raw, "an explicit null needs the raw body path")
		require.JSONEq(t, `{"search_power":100,"min_search_vcu":null}`, string(searchLakeJSON(t, sent)))
		require.NotContains(t, string(sent), `"search_lake":null`)
		require.Contains(t, string(sent), `"name":"proj"`)
	})

	// A computed sibling the configuration omits is unknown in a real plan, and
	// an unknown sibling must stay out of the body.
	unknownSiblings := func(boostWindow attr.Value) resource_elasticsearch_project.SearchLakeValue {
		var model resource_elasticsearch_project.ElasticsearchProjectModel
		return resource_elasticsearch_project.NewSearchLakeValueMust(
			model.SearchLake.AttributeTypes(ctx),
			map[string]attr.Value{
				"boost_window":   boostWindow,
				"min_search_vcu": basetypes.NewInt64Null(),
				"search_power":   basetypes.NewInt64Unknown(),
			},
		)
	}
	stateWithFloor := searchLake(t, ctx, new(int64(100)), new(int64(7)), new(int64(64)))

	t.Run("removes the floor and sends only the siblings that are known", func(t *testing.T) {
		sent, raw := patchWith(t,
			minSearchVcuProject(id, unknownSiblings(basetypes.NewInt64Value(14))),
			minSearchVcuProject(id, stateWithFloor),
		)
		require.True(t, raw)
		require.JSONEq(t, `{"boost_window":14,"min_search_vcu":null}`, string(searchLakeJSON(t, sent)))
	})

	t.Run("removes the floor and sends no siblings when none is known", func(t *testing.T) {
		sent, raw := patchWith(t,
			minSearchVcuProject(id, unknownSiblings(basetypes.NewInt64Unknown())),
			minSearchVcuProject(id, stateWithFloor),
		)
		require.True(t, raw)
		require.JSONEq(t, `{"min_search_vcu":null}`, string(searchLakeJSON(t, sent)))
	})

	for name, state := range map[string]resource_elasticsearch_project.SearchLakeValue{
		"null":    resource_elasticsearch_project.NewSearchLakeValueNull(),
		"unknown": resource_elasticsearch_project.NewSearchLakeValueUnknown(),
	} {
		t.Run("omits the floor key when the state block is "+name, func(t *testing.T) {
			sent, raw := patchWith(t,
				minSearchVcuProject(id, searchLake(t, ctx, new(int64(100)), nil, nil)),
				minSearchVcuProject(id, state),
			)
			require.False(t, raw, "there is no state floor to remove, so no explicit null")
			require.JSONEq(t, `{"search_power":100}`, string(searchLakeJSON(t, sent)))
			require.NotContains(t, string(sent), "min_search_vcu")
		})
	}

	t.Run("sends no search_lake when the plan equals the state", func(t *testing.T) {
		lake := searchLake(t, ctx, new(int64(100)), new(int64(7)), new(int64(64)))
		plan := minSearchVcuProject(id, lake)
		plan.Name = types.StringValue("renamed")
		sent, raw := patchWith(t, plan, minSearchVcuProject(id, lake))
		require.False(t, raw)
		require.Nil(t, searchLakeJSON(t, sent))
	})

	t.Run("sends no search_lake when the plan block is unknown", func(t *testing.T) {
		sent, raw := patchWith(t,
			minSearchVcuProject(id, resource_elasticsearch_project.NewSearchLakeValueUnknown()),
			minSearchVcuProject(id, searchLake(t, ctx, new(int64(100)), new(int64(7)), new(int64(64)))),
		)
		require.False(t, raw)
		require.Nil(t, searchLakeJSON(t, sent))
	})

	t.Run("removes audit and the floor in one body", func(t *testing.T) {
		plan := minSearchVcuProject(id, searchLake(t, ctx, new(int64(100)), nil, nil))
		state := minSearchVcuProject(id, searchLake(t, ctx, new(int64(100)), new(int64(7)), new(int64(64))))
		state.Monitoring = resource_elasticsearch_project.NewMonitoringValueMust(
			state.Monitoring.AttributeTypes(ctx),
			map[string]attr.Value{"logging": loggingWithAudit(t, state.Monitoring.AttributeTypes(ctx))},
		)
		sent, raw := patchWith(t, plan, state)
		require.True(t, raw)
		require.JSONEq(t, `{"search_power":100,"min_search_vcu":null}`, string(searchLakeJSON(t, sent)))
		require.Contains(t, string(sent), `"audit":null`)
		require.NotContains(t, string(sent), `"monitoring":null`)
	})

	t.Run("surfaces an API rejection with status and body", func(t *testing.T) {
		failed := &serverless.PatchElasticsearchProjectResponse{
			HTTPResponse: &http.Response{Status: "400 Bad Request", StatusCode: 400},
			Body:         []byte(`{"errors":[{"message":"can not set a value larger than [42] Min Search VCU in Trial"}]}`),
		}
		mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)
		mockClient.EXPECT().
			PatchElasticsearchProjectWithResponse(ctx, id, nil, gomock.Any()).
			Return(failed, nil)
		diags := elasticsearchApi{client: mockClient}.Patch(ctx,
			minSearchVcuProject(id, searchLake(t, ctx, nil, nil, new(int64(100)))),
			minSearchVcuProject(id, searchLake(t, ctx, nil, nil, nil)),
		)
		require.True(t, diags.HasError())
		require.Equal(t, "Failed to update elasticsearch_project", diags[0].Summary())
		require.Contains(t, diags[0].Detail(), fmt.Sprintf("%d %s", failed.StatusCode(), failed.Status()))
		require.Contains(t, diags[0].Detail(), "larger than [42]")
	})
}

func TestElasticsearchMinSearchVcuModify(t *testing.T) {
	ctx := context.Background()
	mr := elasticsearchModelReader{}
	withFloor := searchLake(t, ctx, new(int64(100)), new(int64(7)), new(int64(64)))

	t.Run("plans an omitted block from state", func(t *testing.T) {
		plan := minSearchVcuProject("id", resource_elasticsearch_project.NewSearchLakeValueUnknown())
		state := minSearchVcuProject("id", withFloor)
		cfg := minSearchVcuProject("", resource_elasticsearch_project.NewSearchLakeValueNull())
		got := mr.Modify(plan, state, cfg)
		require.True(t, got.SearchLake.Equal(withFloor))
	})

	t.Run("leaves the plan alone when the state has no block", func(t *testing.T) {
		plan := minSearchVcuProject("id", resource_elasticsearch_project.NewSearchLakeValueUnknown())
		state := minSearchVcuProject("id", resource_elasticsearch_project.NewSearchLakeValueNull())
		cfg := minSearchVcuProject("", resource_elasticsearch_project.NewSearchLakeValueNull())
		got := mr.Modify(plan, state, cfg)
		require.True(t, got.SearchLake.IsUnknown())
	})

	t.Run("leaves a configured block alone", func(t *testing.T) {
		configured := searchLake(t, ctx, new(int64(100)), nil, nil)
		plan := minSearchVcuProject("id", configured)
		state := minSearchVcuProject("id", withFloor)
		cfg := minSearchVcuProject("", configured)
		got := mr.Modify(plan, state, cfg)
		require.True(t, got.SearchLake.Equal(configured))
		require.True(t, got.SearchLake.MinSearchVcu.IsNull())
	})
}

func TestElasticsearchMinSearchVcuRead(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	const id = "projectid"

	readWith := func(t *testing.T, lake *serverless.ElasticsearchSearchLake) resource_elasticsearch_project.ElasticsearchProjectModel {
		t.Helper()
		mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)
		mockClient.EXPECT().
			GetElasticsearchProjectWithResponse(ctx, id).
			Return(&serverless.GetElasticsearchProjectResponse{JSON200: &serverless.ElasticsearchProject{
				Id:         id,
				Name:       "proj",
				RegionId:   "aws-us-east-1",
				Type:       "elasticsearch",
				SearchLake: lake,
			}}, nil)
		found, model, diags := elasticsearchApi{client: mockClient}.Read(ctx, id, minSearchVcuProject(id, resource_elasticsearch_project.NewSearchLakeValueNull()))
		require.False(t, diags.HasError(), diags)
		require.True(t, found)
		return model
	}

	t.Run("stores the floor the API reports", func(t *testing.T) {
		model := readWith(t, &serverless.ElasticsearchSearchLake{SearchPower: new(100), MinSearchVcu: new(64)})
		require.Equal(t, int64(64), model.SearchLake.MinSearchVcu.ValueInt64())
		require.Equal(t, int64(100), model.SearchLake.SearchPower.ValueInt64())
	})

	t.Run("stores every attribute the API reports", func(t *testing.T) {
		model := readWith(t, &serverless.ElasticsearchSearchLake{SearchPower: new(100), BoostWindow: new(7), MinSearchVcu: new(64)})
		require.Equal(t, int64(100), model.SearchLake.SearchPower.ValueInt64())
		require.Equal(t, int64(7), model.SearchLake.BoostWindow.ValueInt64())
		require.Equal(t, int64(64), model.SearchLake.MinSearchVcu.ValueInt64())
	})

	t.Run("stores null attributes when the API returns no search_lake", func(t *testing.T) {
		model := readWith(t, nil)
		require.False(t, model.SearchLake.IsNull(), "the block itself is known, with null attributes")
		require.False(t, model.SearchLake.IsUnknown())
		require.True(t, model.SearchLake.SearchPower.IsNull())
		require.True(t, model.SearchLake.BoostWindow.IsNull())
		require.True(t, model.SearchLake.MinSearchVcu.IsNull())
	})

	t.Run("stores null when the API has no floor", func(t *testing.T) {
		model := readWith(t, &serverless.ElasticsearchSearchLake{SearchPower: new(100)})
		require.True(t, model.SearchLake.MinSearchVcu.IsNull())
		require.Equal(t, int64(100), model.SearchLake.SearchPower.ValueInt64())
	})
}
