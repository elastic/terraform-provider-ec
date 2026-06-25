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
	getUpdateResources := func() *models.DeploymentUpdateResources {
		return testutil.UpdatePayloadsFromTemplate(t, tplPath)
	}
	tplPathWithIcVersion := "../../testdata/template-aws-io-optimized-v2-ic_version.json"
	getUpdateResourcesWithIcVersion := func() *models.DeploymentUpdateResources {
		return testutil.UpdatePayloadsFromTemplate(t, tplPathWithIcVersion)
	}
	type args struct {
		apm             *Apm
		updateResources *models.DeploymentUpdateResources
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
				updateResources: getUpdateResources(),
				apm: &Apm{
					RefId:                     new("main-apm"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    new("some-region"),
					ElasticsearchClusterRefId: new("somerefid"),
					InstanceConfigurationId:   new("aws.apm.r5d"),
					Size:                      new("2g"),
					SizeResource:              new("memory"),
					ZoneCount:                 1,
				},
			},
			want: &models.ApmPayload{
				ElasticsearchClusterRefID: new("somerefid"),
				Region:                    new("some-region"),
				RefID:                     new("main-apm"),
				Plan: &models.ApmPlan{
					Apm: &models.ApmConfiguration{},
					ClusterTopology: []*models.ApmTopologyElement{{
						ZoneCount:               1,
						InstanceConfigurationID: "aws.apm.r5d",
						Size: &models.TopologySize{
							Resource: new("memory"),
							Value:    ec.Int32(2048),
						},
					}},
				},
			},
		},
		{
			name: "parses an APM resource with no topology",
			args: args{
				updateResources: getUpdateResources(),
				apm: &Apm{
					RefId:                     new("main-apm"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    new("some-region"),
					ElasticsearchClusterRefId: new("somerefid"),
				},
			},
			want: &models.ApmPayload{
				ElasticsearchClusterRefID: new("somerefid"),
				Region:                    new("some-region"),
				RefID:                     new("main-apm"),
				Plan: &models.ApmPlan{
					Apm: &models.ApmConfiguration{},
					ClusterTopology: []*models.ApmTopologyElement{{
						ZoneCount:               1,
						InstanceConfigurationID: "aws.apm.r5d",
						Size: &models.TopologySize{
							Resource: new("memory"),
							Value:    ec.Int32(512),
						},
					}},
				},
			},
		},
		{
			name: "parses an APM resource with a topology element but no instance_configuration_id or instance_configuration_version - use values from template",
			args: args{
				updateResources: getUpdateResourcesWithIcVersion(),
				apm: &Apm{
					RefId:                     new("main-apm"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    new("some-region"),
					ElasticsearchClusterRefId: new("somerefid"),
					Size:                      new("2g"),
					SizeResource:              new("memory"),
				},
			},
			want: &models.ApmPayload{
				ElasticsearchClusterRefID: new("somerefid"),
				Region:                    new("some-region"),
				RefID:                     new("main-apm"),
				Plan: &models.ApmPlan{
					Apm: &models.ApmConfiguration{},
					ClusterTopology: []*models.ApmTopologyElement{{
						ZoneCount:                    1,
						InstanceConfigurationID:      "aws.apm.r5d",
						InstanceConfigurationVersion: ec.Int32(4),
						Size: &models.TopologySize{
							Resource: new("memory"),
							Value:    ec.Int32(2048),
						},
					}},
				},
			},
		},
		{
			name: "parses an APM resource with instance_configuration_id and instance_configuration_version",
			args: args{
				updateResources: getUpdateResources(),
				apm: &Apm{
					RefId:                        new("main-apm"),
					ResourceId:                   &mock.ValidClusterID,
					Region:                       new("some-region"),
					ElasticsearchClusterRefId:    new("somerefid"),
					InstanceConfigurationId:      new("testing.ic"),
					InstanceConfigurationVersion: new(5),
					Size:                         new("2g"),
					SizeResource:                 new("memory"),
				},
			},
			want: &models.ApmPayload{
				ElasticsearchClusterRefID: new("somerefid"),
				Region:                    new("some-region"),
				RefID:                     new("main-apm"),
				Plan: &models.ApmPlan{
					Apm: &models.ApmConfiguration{},
					ClusterTopology: []*models.ApmTopologyElement{{
						ZoneCount:                    1,
						InstanceConfigurationID:      "testing.ic",
						InstanceConfigurationVersion: ec.Int32(5),
						Size: &models.TopologySize{
							Resource: new("memory"),
							Value:    ec.Int32(2048),
						},
					}},
				},
			},
		},
		{
			name: "parses an APM resource with instance_configuration_version set to 0",
			args: args{
				updateResources: getUpdateResources(),
				apm: &Apm{
					RefId:                        new("main-apm"),
					ResourceId:                   &mock.ValidClusterID,
					Region:                       new("some-region"),
					ElasticsearchClusterRefId:    new("somerefid"),
					InstanceConfigurationId:      new("testing.ic"),
					InstanceConfigurationVersion: new(0),
					Size:                         new("2g"),
					SizeResource:                 new("memory"),
				},
			},
			want: &models.ApmPayload{
				ElasticsearchClusterRefID: new("somerefid"),
				Region:                    new("some-region"),
				RefID:                     new("main-apm"),
				Plan: &models.ApmPlan{
					Apm: &models.ApmConfiguration{},
					ClusterTopology: []*models.ApmTopologyElement{{
						ZoneCount:                    1,
						InstanceConfigurationID:      "testing.ic",
						InstanceConfigurationVersion: ec.Int32(0),
						Size: &models.TopologySize{
							Resource: new("memory"),
							Value:    ec.Int32(2048),
						},
					}},
				},
			},
		},
		{
			name: "parses an APM resource with explicit topology and some config",
			args: args{
				updateResources: getUpdateResources(),
				apm: &Apm{
					RefId:                     new("tertiary-apm"),
					ElasticsearchClusterRefId: new("somerefid"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    new("some-region"),
					Config: &v1.ApmConfig{
						UserSettingsYaml:         new("some.setting: value"),
						UserSettingsOverrideYaml: new("some.setting: value2"),
						UserSettingsJson:         new("{\"some.setting\": \"value\"}"),
						UserSettingsOverrideJson: new("{\"some.setting\": \"value2\"}"),
						DebugEnabled:             new(true),
					},
					InstanceConfigurationId: new("aws.apm.r5d"),
					Size:                    new("4g"),
					SizeResource:            new("memory"),
					ZoneCount:               1,
				},
			},
			want: &models.ApmPayload{
				ElasticsearchClusterRefID: new("somerefid"),
				Region:                    new("some-region"),
				RefID:                     new("tertiary-apm"),
				Plan: &models.ApmPlan{
					Apm: &models.ApmConfiguration{
						UserSettingsYaml:         `some.setting: value`,
						UserSettingsOverrideYaml: `some.setting: value2`,
						UserSettingsJSON: map[string]any{
							"some.setting": "value",
						},
						UserSettingsOverrideJSON: map[string]any{
							"some.setting": "value2",
						},
						SystemSettings: &models.ApmSystemSettings{
							DebugEnabled: new(true),
						},
					},
					ClusterTopology: []*models.ApmTopologyElement{
						{
							ZoneCount:               1,
							InstanceConfigurationID: "aws.apm.r5d",
							Size: &models.TopologySize{
								Resource: new("memory"),
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
				updateResources: nil,
				apm: &Apm{
					RefId:                     new("tertiary-apm"),
					ElasticsearchClusterRefId: new("somerefid"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    new("some-region"),
					InstanceConfigurationId:   new("aws.apm.r5d"),
					Size:                      new("4g"),
					SizeResource:              new("memory"),
					ZoneCount:                 1,
					Config: &v1.ApmConfig{
						DebugEnabled: new(true),
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
			diags := tfsdk.ValueFrom(context.Background(), tt.args.apm, ApmSchema().GetType(), &apm)
			assert.Nil(t, diags)

			if got, diags := ApmPayload(context.Background(), apm, tt.args.updateResources); tt.diags != nil {
				assert.Equal(t, tt.diags, diags)
			} else {
				assert.Nil(t, diags)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
