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
	v1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/apm/v1"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/testutil"
)

func Test_ApmPayload(t *testing.T) {
	tplPath := "../../testdata/template-aws-io-optimized-v2.json"
	tpl := func() *models.DeploymentTemplateInfoV2 {
		return testutil.ParseDeploymentTemplate(t, tplPath)
	}
	type args struct {
		apm *Apm
		tpl *models.DeploymentTemplateInfoV2
	}
	tests := []struct {
		name  string
		args  args
		want  *models.ApmPayload
		diags diag.Diagnostics
	}{
		{
			name: "returns nil when there's no resources",
		},
		{
			name: "parses an APM resource with explicit topology",
			args: args{
				tpl: tpl(),
				apm: &Apm{
					RefId:                     ec.String("main-apm"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
					InstanceConfigurationId:   ec.String("aws.apm.r5d"),
					Size:                      ec.String("2g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
			},
			want: &models.ApmPayload{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("main-apm"),
				Plan: &models.ApmPlan{
					Apm: &models.ApmConfiguration{},
					ClusterTopology: []*models.ApmTopologyElement{{
						ZoneCount:               1,
						InstanceConfigurationID: "aws.apm.r5d",
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(2048),
						},
					}},
				},
			},
		},
		{
			name: "parses an APM resource with invalid instance_configuration_id",
			args: args{
				tpl: tpl(),
				apm: &Apm{
					RefId:                     ec.String("main-apm"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
					InstanceConfigurationId:   ec.String("so invalid"),
					Size:                      ec.String("2g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
			},
			diags: func() diag.Diagnostics {
				var diags diag.Diagnostics
				diags.AddError(
					"cannot match topology element",
					`apm topology: invalid instance_configuration_id: "so invalid" doesn't match any of the deployment template instance configurations`,
				)
				return diags
			}(),
		},
		{
			name: "parses an APM resource with no topology",
			args: args{
				tpl: tpl(),
				apm: &Apm{
					RefId:                     ec.String("main-apm"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
				},
			},
			want: &models.ApmPayload{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("main-apm"),
				Plan: &models.ApmPlan{
					Apm: &models.ApmConfiguration{},
					ClusterTopology: []*models.ApmTopologyElement{{
						ZoneCount:               1,
						InstanceConfigurationID: "aws.apm.r5d",
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(512),
						},
					}},
				},
			},
		},
		{
			name: "parses an APM resource with a topology element but no instance_configuration_id",
			args: args{
				tpl: tpl(),
				apm: &Apm{
					RefId:                     ec.String("main-apm"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
					Size:                      ec.String("2g"),
					SizeResource:              ec.String("memory"),
				},
			},
			want: &models.ApmPayload{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("main-apm"),
				Plan: &models.ApmPlan{
					Apm: &models.ApmConfiguration{},
					ClusterTopology: []*models.ApmTopologyElement{{
						ZoneCount:               1,
						InstanceConfigurationID: "aws.apm.r5d",
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(2048),
						},
					}},
				},
			},
		},
		{
			name: "parses an APM resource with explicit topology and some config",
			args: args{
				tpl: tpl(),
				apm: &Apm{
					RefId:                     ec.String("tertiary-apm"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    ec.String("some-region"),
					Config: &v1.ApmConfig{
						UserSettingsYaml:         ec.String("some.setting: value"),
						UserSettingsOverrideYaml: ec.String("some.setting: value2"),
						UserSettingsJson:         ec.String("{\"some.setting\": \"value\"}"),
						UserSettingsOverrideJson: ec.String("{\"some.setting\": \"value2\"}"),
						DebugEnabled:             ec.Bool(true),
					},
					InstanceConfigurationId: ec.String("aws.apm.r5d"),
					Size:                    ec.String("4g"),
					SizeResource:            ec.String("memory"),
					ZoneCount:               1,
				},
			},
			want: &models.ApmPayload{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("tertiary-apm"),
				Plan: &models.ApmPlan{
					Apm: &models.ApmConfiguration{
						UserSettingsYaml:         `some.setting: value`,
						UserSettingsOverrideYaml: `some.setting: value2`,
						UserSettingsJSON: map[string]interface{}{
							"some.setting": "value",
						},
						UserSettingsOverrideJSON: map[string]interface{}{
							"some.setting": "value2",
						},
						SystemSettings: &models.ApmSystemSettings{
							DebugEnabled: ec.Bool(true),
						},
					},
					ClusterTopology: []*models.ApmTopologyElement{
						{
							ZoneCount:               1,
							InstanceConfigurationID: "aws.apm.r5d",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(4096),
							},
						},
					},
				},
			},
		},
		{
			name: "tries to parse an apm resource when the template doesn't have an APM instance set.",
			args: args{
				tpl: nil,
				apm: &Apm{
					RefId:                     ec.String("tertiary-apm"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    ec.String("some-region"),
					InstanceConfigurationId:   ec.String("aws.apm.r5d"),
					Size:                      ec.String("4g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
					Config: &v1.ApmConfig{
						DebugEnabled: ec.Bool(true),
					},
				},
			},
			diags: func() diag.Diagnostics {
				var diags diag.Diagnostics
				diags.AddError("apm payload error", "apm specified but deployment template is not configured for it. Use a different template if you wish to add apm")
				return diags
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var apm types.Object
			diags := tfsdk.ValueFrom(context.Background(), tt.args.apm, ApmSchema().FrameworkType(), &apm)
			assert.Nil(t, diags)

			if got, diags := ApmPayload(context.Background(), apm, tt.args.tpl); tt.diags != nil {
				assert.Equal(t, tt.diags, diags)
			} else {
				assert.Nil(t, diags)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
