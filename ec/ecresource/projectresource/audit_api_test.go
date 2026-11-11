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
	"io"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/mocks"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_elasticsearch_project"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_observability_project"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_security_project"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAuditLoggingCreateAndClearOnEachProject(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	t.Run("elasticsearch", func(t *testing.T) {
		model := resource_elasticsearch_project.ElasticsearchProjectModel{
			Id:       types.StringValue("projectid"),
			Name:     types.StringValue("proj"),
			RegionId: types.StringValue("aws-us-east-1"),
		}
		model.Monitoring = resource_elasticsearch_project.NewMonitoringValueMust(
			model.Monitoring.AttributeTypes(ctx),
			map[string]attr.Value{"logging": loggingWithAudit(t, model.Monitoring.AttributeTypes(ctx))},
		)
		mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)
		mockClient.EXPECT().
			CreateElasticsearchProjectWithResponse(ctx, gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ *serverless.CreateElasticsearchProjectParams, body serverless.CreateElasticsearchProjectJSONRequestBody, _ ...serverless.RequestEditorFn) (*serverless.CreateElasticsearchProjectResponse, error) {
				requireAuditCreate(t, body)
				return &serverless.CreateElasticsearchProjectResponse{
					JSON201: &serverless.ElasticsearchProjectCreated{
						Id:          "projectid",
						Credentials: serverless.ProjectCredentials{Username: "u", Password: "p"},
					},
				}, nil
			})
		_, diags := (elasticsearchApi{}).WithClient(mockClient).Create(ctx, model)
		require.False(t, diags.HasError(), diags)

		cleared := model
		cleared.Monitoring = resource_elasticsearch_project.NewMonitoringValueNull()
		mockClient.EXPECT().
			PatchElasticsearchProjectWithBodyWithResponse(ctx, "projectid", nil, "application/json", gomock.Any()).
			DoAndReturn(func(_ context.Context, _ serverless.ProjectID, _ *serverless.PatchElasticsearchProjectParams, _ string, body io.Reader, _ ...serverless.RequestEditorFn) (*serverless.PatchElasticsearchProjectResponse, error) {
				requireAuditCleared(t, body)
				return &serverless.PatchElasticsearchProjectResponse{JSON200: &serverless.ElasticsearchProject{}}, nil
			})
		diags = elasticsearchApi{client: mockClient}.Patch(ctx, cleared, model)
		require.False(t, diags.HasError(), diags)
	})

	t.Run("observability", func(t *testing.T) {
		model := resource_observability_project.ObservabilityProjectModel{
			Id:       types.StringValue("projectid"),
			Name:     types.StringValue("proj"),
			RegionId: types.StringValue("aws-us-east-1"),
		}
		model.Monitoring = resource_observability_project.NewMonitoringValueMust(
			model.Monitoring.AttributeTypes(ctx),
			map[string]attr.Value{"logging": loggingWithAudit(t, model.Monitoring.AttributeTypes(ctx))},
		)
		mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)
		mockClient.EXPECT().
			CreateObservabilityProjectWithResponse(ctx, gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ *serverless.CreateObservabilityProjectParams, body serverless.CreateObservabilityProjectJSONRequestBody, _ ...serverless.RequestEditorFn) (*serverless.CreateObservabilityProjectResponse, error) {
				requireAuditCreate(t, body)
				return &serverless.CreateObservabilityProjectResponse{
					JSON201: &serverless.ObservabilityProjectCreated{
						Id:          "projectid",
						Credentials: serverless.ProjectCredentials{Username: "u", Password: "p"},
					},
				}, nil
			})
		_, diags := (observabilityApi{}).WithClient(mockClient).Create(ctx, model)
		require.False(t, diags.HasError(), diags)

		cleared := model
		cleared.Monitoring = resource_observability_project.NewMonitoringValueNull()
		mockClient.EXPECT().
			PatchObservabilityProjectWithBodyWithResponse(ctx, "projectid", nil, "application/json", gomock.Any()).
			DoAndReturn(func(_ context.Context, _ serverless.ProjectID, _ *serverless.PatchObservabilityProjectParams, _ string, body io.Reader, _ ...serverless.RequestEditorFn) (*serverless.PatchObservabilityProjectResponse, error) {
				requireAuditCleared(t, body)
				return &serverless.PatchObservabilityProjectResponse{JSON200: &serverless.ObservabilityProject{}}, nil
			})
		diags = observabilityApi{client: mockClient}.Patch(ctx, cleared, model)
		require.False(t, diags.HasError(), diags)
	})

	t.Run("security", func(t *testing.T) {
		model := resource_security_project.SecurityProjectModel{
			Id:       types.StringValue("projectid"),
			Name:     types.StringValue("proj"),
			RegionId: types.StringValue("aws-us-east-1"),
		}
		model.Monitoring = resource_security_project.NewMonitoringValueMust(
			model.Monitoring.AttributeTypes(ctx),
			map[string]attr.Value{"logging": loggingWithAudit(t, model.Monitoring.AttributeTypes(ctx))},
		)
		mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)
		mockClient.EXPECT().
			CreateSecurityProjectWithResponse(ctx, gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ *serverless.CreateSecurityProjectParams, body serverless.CreateSecurityProjectJSONRequestBody, _ ...serverless.RequestEditorFn) (*serverless.CreateSecurityProjectResponse, error) {
				requireAuditCreate(t, body)
				return &serverless.CreateSecurityProjectResponse{
					JSON201: &serverless.SecurityProjectCreated{
						Id:          "projectid",
						Credentials: serverless.ProjectCredentials{Username: "u", Password: "p"},
					},
				}, nil
			})
		_, diags := (securityApi{}).WithClient(mockClient).Create(ctx, model)
		require.False(t, diags.HasError(), diags)

		cleared := model
		cleared.Monitoring = resource_security_project.NewMonitoringValueNull()
		mockClient.EXPECT().
			PatchSecurityProjectWithBodyWithResponse(ctx, "projectid", nil, "application/json", gomock.Any()).
			DoAndReturn(func(_ context.Context, _ serverless.ProjectID, _ *serverless.PatchSecurityProjectParams, _ string, body io.Reader, _ ...serverless.RequestEditorFn) (*serverless.PatchSecurityProjectResponse, error) {
				requireAuditCleared(t, body)
				return &serverless.PatchSecurityProjectResponse{JSON200: &serverless.SecurityProject{}}, nil
			})
		diags = securityApi{client: mockClient}.Patch(ctx, cleared, model)
		require.False(t, diags.HasError(), diags)
	})
}

func TestAuditLoggingReadStoresStatusOnEachProject(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	monitoring := &serverless.Monitoring{
		Logging: &serverless.Logging{
			Audit: &serverless.LoggingCategory{
				Enabled: false,
				Destination: serverless.LoggingDestination{
					ProjectId:   "dest-1",
					ProjectType: "security",
					Status:      "suspended",
				},
			},
		},
	}

	t.Run("elasticsearch", func(t *testing.T) {
		mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)
		mockClient.EXPECT().
			GetElasticsearchProjectWithResponse(ctx, serverless.ProjectID("projectid")).
			Return(&serverless.GetElasticsearchProjectResponse{
				JSON200: &serverless.ElasticsearchProject{
					Id:         "projectid",
					Alias:      "alias-projec",
					CloudId:    "cloud",
					Endpoints:  serverless.ElasticsearchProjectEndpoints{Elasticsearch: "es", Kibana: "kb"},
					Metadata:   serverless.ProjectMetadata{CreatedAt: now, CreatedBy: "me", OrganizationId: "org"},
					Name:       "proj",
					RegionId:   "aws-us-east-1",
					Type:       "elasticsearch",
					Monitoring: monitoring,
				},
			}, nil)
		found, model, diags := elasticsearchApi{client: mockClient}.Read(ctx, "projectid", resource_elasticsearch_project.ElasticsearchProjectModel{
			Id: types.StringValue("projectid"),
		})
		require.False(t, diags.HasError(), diags)
		require.True(t, found)
		require.Equal(t, "suspended", destinationStatus(t, model.Monitoring.Logging))
	})

	t.Run("observability", func(t *testing.T) {
		mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)
		mockClient.EXPECT().
			GetObservabilityProjectWithResponse(ctx, serverless.ProjectID("projectid")).
			Return(&serverless.GetObservabilityProjectResponse{
				JSON200: &serverless.ObservabilityProject{
					Id:         "projectid",
					Alias:      "alias-projec",
					CloudId:    "cloud",
					Endpoints:  serverless.ObservabilityProjectEndpoints{Elasticsearch: "es", Kibana: "kb"},
					Metadata:   serverless.ProjectMetadata{CreatedAt: now, CreatedBy: "me", OrganizationId: "org"},
					Name:       "proj",
					RegionId:   "aws-us-east-1",
					Type:       "observability",
					Monitoring: monitoring,
				},
			}, nil)
		found, model, diags := observabilityApi{client: mockClient}.Read(ctx, "projectid", resource_observability_project.ObservabilityProjectModel{
			Id: types.StringValue("projectid"),
		})
		require.False(t, diags.HasError(), diags)
		require.True(t, found)
		require.Equal(t, "suspended", destinationStatus(t, model.Monitoring.Logging))
	})

	t.Run("security", func(t *testing.T) {
		mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)
		mockClient.EXPECT().
			GetSecurityProjectWithResponse(ctx, serverless.ProjectID("projectid")).
			Return(&serverless.GetSecurityProjectResponse{
				JSON200: &serverless.SecurityProject{
					Id:         "projectid",
					Alias:      "alias-projec",
					CloudId:    "cloud",
					Endpoints:  serverless.SecurityProjectEndpoints{Elasticsearch: "es", Kibana: "kb"},
					Metadata:   serverless.ProjectMetadata{CreatedAt: now, CreatedBy: "me", OrganizationId: "org"},
					Name:       "proj",
					RegionId:   "aws-us-east-1",
					Type:       "security",
					Monitoring: monitoring,
				},
			}, nil)
		found, model, diags := securityApi{client: mockClient}.Read(ctx, "projectid", resource_security_project.SecurityProjectModel{
			Id: types.StringValue("projectid"),
		})
		require.False(t, diags.HasError(), diags)
		require.True(t, found)
		require.Equal(t, "suspended", destinationStatus(t, model.Monitoring.Logging))
	})
}

func loggingWithAudit(t *testing.T, monitoringTypes map[string]attr.Type) basetypes.ObjectValue {
	t.Helper()
	loggingTypes := monitoringTypes["logging"].(attr.TypeWithAttributeTypes).AttributeTypes()
	auditTypes := loggingTypes["audit"].(attr.TypeWithAttributeTypes).AttributeTypes()
	destTypes := auditTypes["destination"].(attr.TypeWithAttributeTypes).AttributeTypes()
	filters, diags := types.SetValue(types.StringType, []attr.Value{types.StringValue("audit-ignore-data-reads")})
	require.False(t, diags.HasError(), diags)
	dest, diags := types.ObjectValue(destTypes, map[string]attr.Value{
		"project_id":   types.StringValue("dest-1"),
		"project_type": types.StringValue("observability"),
		"status":       types.StringNull(),
	})
	require.False(t, diags.HasError(), diags)
	audit, diags := types.ObjectValue(auditTypes, map[string]attr.Value{
		"destination":       dest,
		"enabled":           types.BoolValue(false),
		"ignore_filter_ids": filters,
	})
	require.False(t, diags.HasError(), diags)
	logging, diags := types.ObjectValue(loggingTypes, map[string]attr.Value{"audit": audit})
	require.False(t, diags.HasError(), diags)
	return logging
}

func requireAuditCreate(t *testing.T, body any) {
	t.Helper()
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	var decoded struct {
		Monitoring json.RawMessage `json:"monitoring"`
	}
	require.NoError(t, json.Unmarshal(raw, &decoded))
	require.JSONEq(t, `{
		"logging": {
			"audit": {
				"enabled": false,
				"destination": {"project_id": "dest-1", "project_type": "observability"},
				"ignore_filters": [{"id": "audit-ignore-data-reads"}]
			}
		}
	}`, string(decoded.Monitoring))
	require.NotContains(t, string(decoded.Monitoring), "status")
}

func requireAuditCleared(t *testing.T, body io.Reader) {
	t.Helper()
	raw, err := io.ReadAll(body)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"audit":null`)
	require.NotContains(t, string(raw), `"monitoring":null`)
}

func destinationStatus(t *testing.T, logging basetypes.ObjectValue) string {
	t.Helper()
	audit := logging.Attributes()["audit"].(basetypes.ObjectValue)
	dest := audit.Attributes()["destination"].(basetypes.ObjectValue)
	return dest.Attributes()["status"].(basetypes.StringValue).ValueString()
}
