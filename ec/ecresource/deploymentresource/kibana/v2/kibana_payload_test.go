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

func Test_KibanaPayload(t *testing.T) {
	tplPath := "../../testdata/template-aws-io-optimized-v2.json"
	tpl := func() *models.DeploymentTemplateInfoV2 {
		return testutil.ParseDeploymentTemplate(t, tplPath)
	}
	type args struct {
		kibana *Kibana
		tpl    *models.DeploymentTemplateInfoV2
	}
	tests := []struct {
		name  string
		args  args
		want  *models.KibanaPayload
		diags diag.Diagnostics
	}{
		{
			name: "returns nil when there's no resources",
		},
		{
			name: "parses a kibana resource with topology",
			args: args{
				tpl: tpl(),
				kibana: &Kibana{
					RefId:                     ec.String("main-kibana"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
					InstanceConfigurationId:   ec.String("aws.kibana.r5d"),
					Size:                      ec.String("2g"),
					ZoneCount:                 1,
				},
			},
			want: &models.KibanaPayload{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("main-kibana"),
				Plan: &models.KibanaClusterPlan{
					Kibana: &models.KibanaConfiguration{},
					ClusterTopology: []*models.KibanaClusterTopologyElement{
						{
							ZoneCount:               1,
							InstanceConfigurationID: "aws.kibana.r5d",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
						},
					},
				},
			},
		},
		{
			name: "parses a kibana resource with incorrect instance_configuration_id",
			args: args{
				tpl: tpl(),
				kibana: &Kibana{
					RefId:                     ec.String("main-kibana"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
					InstanceConfigurationId:   ec.String("gcp.some.config"),
					Size:                      ec.String("2g"),
					ZoneCount:                 1,
				},
			},
			diags: func() diag.Diagnostics {
				var diags diag.Diagnostics
				diags.AddError("kibana topology payload error", `kibana topology: invalid instance_configuration_id: "gcp.some.config" doesn't match any of the deployment template instance configurations`)
				return diags
			}(),
		},
		{
			name: "parses a kibana resource without topology",
			args: args{
				tpl: tpl(),
				kibana: &Kibana{
					RefId:                     ec.String("main-kibana"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
				},
			},
			want: &models.KibanaPayload{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("main-kibana"),
				Plan: &models.KibanaClusterPlan{
					Kibana: &models.KibanaConfiguration{},
					ClusterTopology: []*models.KibanaClusterTopologyElement{
						{
							ZoneCount:               1,
							InstanceConfigurationID: "aws.kibana.r5d",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(1024),
							},
						},
					},
				},
			},
		},
		{
			name: "parses a kibana resource with a topology but no instance_configuration_id",
			args: args{
				tpl: tpl(),
				kibana: &Kibana{
					RefId:                     ec.String("main-kibana"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
					Size:                      ec.String("4g"),
				},
			},
			want: &models.KibanaPayload{

				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("main-kibana"),
				Plan: &models.KibanaClusterPlan{
					Kibana: &models.KibanaConfiguration{},
					ClusterTopology: []*models.KibanaClusterTopologyElement{
						{
							ZoneCount:               1,
							InstanceConfigurationID: "aws.kibana.r5d",
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
			name: "parses a kibana resource with topology and settings",
			args: args{
				tpl: tpl(),
				kibana: &Kibana{
					RefId:                     ec.String("secondary-kibana"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
					Config: &KibanaConfig{
						UserSettingsYaml:         ec.String("some.setting: value"),
						UserSettingsOverrideYaml: ec.String("some.setting: override"),
						UserSettingsJson:         ec.String(`{"some.setting":"value"}`),
						UserSettingsOverrideJson: ec.String(`{"some.setting":"override"}`),
					},
					InstanceConfigurationId: ec.String("aws.kibana.r5d"),
					Size:                    ec.String("4g"),
					ZoneCount:               1,
				},
			},
			want: &models.KibanaPayload{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("secondary-kibana"),
				Plan: &models.KibanaClusterPlan{
					Kibana: &models.KibanaConfiguration{
						UserSettingsYaml:         "some.setting: value",
						UserSettingsOverrideYaml: "some.setting: override",
						UserSettingsJSON: map[string]interface{}{
							"some.setting": "value",
						},
						UserSettingsOverrideJSON: map[string]interface{}{
							"some.setting": "override",
						},
					},
					ClusterTopology: []*models.KibanaClusterTopologyElement{{
						ZoneCount:               1,
						InstanceConfigurationID: "aws.kibana.r5d",
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(4096),
						},
					}},
				},
			},
		},
		{
			name: "tries to parse an kibana resource when the template doesn't have a kibana instance set.",
			args: args{
				tpl: nil,
				kibana: &Kibana{
					RefId:                     ec.String("tertiary-kibana"),
					ResourceId:                &mock.ValidClusterID,
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
					InstanceConfigurationId:   ec.String("aws.kibana.r5d"),
					Size:                      ec.String("1g"),
					SizeResource:              ec.String("memory"),
					ZoneCount:                 1,
				},
			},
			diags: func() diag.Diagnostics {
				var diags diag.Diagnostics
				diags.AddError("kibana payload error", "kibana specified but deployment template is not configured for it. Use a different template if you wish to add kibana")
				return diags
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var kibana types.Object
			diags := tfsdk.ValueFrom(context.Background(), tt.args.kibana, KibanaSchema().FrameworkType(), &kibana)
			assert.Nil(t, diags)

			if got, diags := KibanaPayload(context.Background(), kibana, tt.args.tpl); tt.diags != nil {
				assert.Equal(t, tt.diags, diags)
			} else {
				assert.Nil(t, diags)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
