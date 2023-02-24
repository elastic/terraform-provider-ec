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

package v2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/testutil"
)

func Test_IntegrationsServerPayload(t *testing.T) {
	tplPath := "../../testdata/template-ece-3.0.0-default.json"
	tpl := func() *models.DeploymentTemplateInfoV2 {
		return testutil.ParseDeploymentTemplate(t, tplPath)
	}
	type args struct {
		srv *IntegrationsServer
		tpl *models.DeploymentTemplateInfoV2
	}
	tests := []struct {
		name  string
		args  args
		want  *models.IntegrationsServerPayload
		diags diag.Diagnostics
	}{
		{
			name: "returns nil when there's no resources",
		},
		{
			name: "parses an Integrations Server resource with explicit topology",
			args: args{
				tpl: tpl(),
				srv: &IntegrationsServer{
					RefId:                     ec.String("main-integrations_server"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
					InstanceConfigurationId:   ec.String("integrations.server"),
					Size:                      ec.String("2g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
			},
			want: &models.IntegrationsServerPayload{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("main-integrations_server"),
				Plan: &models.IntegrationsServerPlan{
					IntegrationsServer: &models.IntegrationsServerConfiguration{},
					ClusterTopology: []*models.IntegrationsServerTopologyElement{{
						ZoneCount:               1,
						InstanceConfigurationID: "integrations.server",
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(2048),
						},
					}},
				},
			},
		},
		{
			name: "parses an Integrations Server resource with invalid instance_configuration_id",
			args: args{
				tpl: tpl(),
				srv: &IntegrationsServer{
					RefId:                     ec.String("main-integrations_server"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
					InstanceConfigurationId:   ec.String("invalid"),
					Size:                      ec.String("2g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
			},
			diags: func() diag.Diagnostics {
				var diags diag.Diagnostics
				diags.AddError("integrations_server topology payload error", `invalid instance_configuration_id: "invalid" doesn't match any of the deployment template instance configurations`)
				return diags
			}(),
		},
		{
			name: "parses an Integrations Server resource with no topology",
			args: args{
				tpl: tpl(),
				srv: &IntegrationsServer{
					RefId:                     ec.String("main-integrations_server"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
				},
			},
			want: &models.IntegrationsServerPayload{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("main-integrations_server"),
				Plan: &models.IntegrationsServerPlan{
					IntegrationsServer: &models.IntegrationsServerConfiguration{},
					ClusterTopology: []*models.IntegrationsServerTopologyElement{{
						ZoneCount:               1,
						InstanceConfigurationID: "integrations.server",
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(1024),
						},
					}},
				},
			},
		},
		{
			name: "parses an Integrations Server resource with a topology element but no instance_configuration_id",
			args: args{
				tpl: tpl(),
				srv: &IntegrationsServer{
					RefId:                     ec.String("main-integrations_server"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
					Size:                      ec.String("2g"),
					SizeResource:              ec.String("memory"),
				},
			},
			want: &models.IntegrationsServerPayload{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("main-integrations_server"),
				Plan: &models.IntegrationsServerPlan{
					IntegrationsServer: &models.IntegrationsServerConfiguration{},
					ClusterTopology: []*models.IntegrationsServerTopologyElement{{
						ZoneCount:               1,
						InstanceConfigurationID: "integrations.server",
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(2048),
						},
					}},
				},
			},
		},
		{
			name: "parses an Integrations Server resource with explicit topology and some config",
			args: args{
				tpl: tpl(),
				srv: &IntegrationsServer{
					RefId:                     ec.String("tertiary-integrations_server"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
					Config: &IntegrationsServerConfig{
						UserSettingsYaml:         ec.String("some.setting: value"),
						UserSettingsOverrideYaml: ec.String("some.setting: value2"),
						UserSettingsJson:         ec.String("{\"some.setting\": \"value\"}"),
						UserSettingsOverrideJson: ec.String("{\"some.setting\": \"value2\"}"),
						DebugEnabled:             ec.Bool(true),
					},
					InstanceConfigurationId: ec.String("integrations.server"),
					Size:                    ec.String("4g"),
					SizeResource:            ec.String("memory"),
					ZoneCount:               1,
				},
			},
			want: &models.IntegrationsServerPayload{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("tertiary-integrations_server"),
				Plan: &models.IntegrationsServerPlan{
					IntegrationsServer: &models.IntegrationsServerConfiguration{
						UserSettingsYaml:         `some.setting: value`,
						UserSettingsOverrideYaml: `some.setting: value2`,
						UserSettingsJSON: map[string]interface{}{
							"some.setting": "value",
						},
						UserSettingsOverrideJSON: map[string]interface{}{
							"some.setting": "value2",
						},
						SystemSettings: &models.IntegrationsServerSystemSettings{
							DebugEnabled: ec.Bool(true),
						},
					},
					ClusterTopology: []*models.IntegrationsServerTopologyElement{{
						ZoneCount:               1,
						InstanceConfigurationID: "integrations.server",
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(4096),
						},
					}},
				},
			},
		},
		{
			name: "tries to parse an integrations_server resource when the template doesn't have an Integrations Server instance set.",
			args: args{
				tpl: nil,
				srv: &IntegrationsServer{
					RefId:                     ec.String("tertiary-integrations_server"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
					Config: &IntegrationsServerConfig{
						DebugEnabled: ec.Bool(true),
					},
					InstanceConfigurationId: ec.String("integrations.server"),
					Size:                    ec.String("4g"),
					SizeResource:            ec.String("memory"),
					ZoneCount:               1,
				},
			},
			diags: func() diag.Diagnostics {
				var diags diag.Diagnostics
				diags.AddError("integrations_server payload error", "integrations_server specified but deployment template is not configured for it. Use a different template if you wish to add integrations_server")
				return diags
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var srv types.Object
			diags := tfsdk.ValueFrom(context.Background(), tt.args.srv, IntegrationsServerSchema().FrameworkType(), &srv)
			assert.Nil(t, diags)

			if got, diags := IntegrationsServerPayload(context.Background(), srv, tt.args.tpl); tt.diags != nil {
				assert.Equal(t, tt.diags, diags)
			} else {
				assert.Nil(t, diags)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
