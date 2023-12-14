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

func Test_enterpriseSearchPayload(t *testing.T) {
	tplPath := "../../testdata/template-aws-io-optimized-v2.json"
	tplPathWithIcVersion := "../../testdata/template-aws-io-optimized-v2-ic_version.json"
	type args struct {
		es              *EnterpriseSearch
		updateResources *models.DeploymentUpdateResources
	}
	tests := []struct {
		name  string
		args  args
		want  *models.EnterpriseSearchPayload
		diags diag.Diagnostics
	}{
		{
			name: "returns nil when there's no resources",
		},
		{
			name: "parses an enterprise_search resource with explicit topology",
			args: args{
				updateResources: testutil.UpdatePayloadsFromTemplate(t, tplPath),
				es: &EnterpriseSearch{
					RefId:                     ec.String("main-enterprise_search"),
					ResourceId:                ec.String(mock.ValidClusterID),
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
					InstanceConfigurationId:   ec.String("aws.enterprisesearch.m5d"),
					Size:                      ec.String("2g"),
					ZoneCount:                 1,
				},
			},
			want: &models.EnterpriseSearchPayload{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("main-enterprise_search"),
				Plan: &models.EnterpriseSearchPlan{
					EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
					ClusterTopology: []*models.EnterpriseSearchTopologyElement{
						{
							ZoneCount:               1,
							InstanceConfigurationID: "aws.enterprisesearch.m5d",
							Size: &models.TopologySize{
								Resource: ec.String("memory"),
								Value:    ec.Int32(2048),
							},
							NodeType: &models.EnterpriseSearchNodeTypes{
								Appserver: ec.Bool(true),
								Connector: ec.Bool(true),
								Worker:    ec.Bool(true),
							},
						},
					},
				},
			},
		},
		{
			name: "parses an enterprise_search resource with no topology takes the minimum size",
			args: args{
				updateResources: testutil.UpdatePayloadsFromTemplate(t, tplPath),
				es: &EnterpriseSearch{
					RefId:                     ec.String("main-enterprise_search"),
					ResourceId:                ec.String(mock.ValidClusterID),
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
				},
			},
			want: &models.EnterpriseSearchPayload{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("main-enterprise_search"),
				Plan: &models.EnterpriseSearchPlan{
					EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
					ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
						ZoneCount:               2,
						InstanceConfigurationID: "aws.enterprisesearch.m5d",
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(2048),
						},
						NodeType: &models.EnterpriseSearchNodeTypes{
							Appserver: ec.Bool(true),
							Connector: ec.Bool(true),
							Worker:    ec.Bool(true),
						},
					}},
				},
			},
		},
		{
			name: "parses an enterprise_search resource with explicit topology but no instance_configuration_id or instance_configuration_version - use values from template",
			args: args{
				updateResources: testutil.UpdatePayloadsFromTemplate(t, tplPathWithIcVersion),
				es: &EnterpriseSearch{
					RefId:                     ec.String("main-enterprise_search"),
					ResourceId:                ec.String(mock.ValidClusterID),
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
					Size:                      ec.String("4g"),
				},
			},
			want: &models.EnterpriseSearchPayload{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("main-enterprise_search"),
				Plan: &models.EnterpriseSearchPlan{
					EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
					ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
						ZoneCount:                    2,
						InstanceConfigurationID:      "aws.enterprisesearch.m5d",
						InstanceConfigurationVersion: 3,
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(4096),
						},
						NodeType: &models.EnterpriseSearchNodeTypes{
							Appserver: ec.Bool(true),
							Connector: ec.Bool(true),
							Worker:    ec.Bool(true),
						},
					}},
				},
			},
		},
		{
			name: "parses an enterprise_search resource with instance_configuration_id and instance_configuration_version",
			args: args{
				updateResources: testutil.UpdatePayloadsFromTemplate(t, tplPathWithIcVersion),
				es: &EnterpriseSearch{
					RefId:                        ec.String("main-enterprise_search"),
					ResourceId:                   ec.String(mock.ValidClusterID),
					Region:                       ec.String("some-region"),
					ElasticsearchClusterRefId:    ec.String("somerefid"),
					InstanceConfigurationId:      ec.String("testing.ic"),
					InstanceConfigurationVersion: ec.Int(1),
				},
			},
			want: &models.EnterpriseSearchPayload{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("main-enterprise_search"),
				Plan: &models.EnterpriseSearchPlan{
					EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
					ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
						ZoneCount:                    2,
						InstanceConfigurationID:      "testing.ic",
						InstanceConfigurationVersion: 1,
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(2048),
						},
						NodeType: &models.EnterpriseSearchNodeTypes{
							Appserver: ec.Bool(true),
							Connector: ec.Bool(true),
							Worker:    ec.Bool(true),
						},
					}},
				},
			},
		},
		{
			name: "parses an enterprise_search resource with instance_configuration_version set to 0",
			args: args{
				updateResources: testutil.UpdatePayloadsFromTemplate(t, tplPathWithIcVersion),
				es: &EnterpriseSearch{
					RefId:                        ec.String("main-enterprise_search"),
					ResourceId:                   ec.String(mock.ValidClusterID),
					Region:                       ec.String("some-region"),
					ElasticsearchClusterRefId:    ec.String("somerefid"),
					InstanceConfigurationId:      ec.String("testing.ic"),
					InstanceConfigurationVersion: ec.Int(0),
				},
			},
			want: &models.EnterpriseSearchPayload{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("main-enterprise_search"),
				Plan: &models.EnterpriseSearchPlan{
					EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
					ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
						ZoneCount:                    2,
						InstanceConfigurationID:      "testing.ic",
						InstanceConfigurationVersion: 0,
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(2048),
						},
						NodeType: &models.EnterpriseSearchNodeTypes{
							Appserver: ec.Bool(true),
							Connector: ec.Bool(true),
							Worker:    ec.Bool(true),
						},
					}},
				},
			},
		},
		{
			name: "parses an enterprise_search resource with topology and zone_count",
			args: args{
				updateResources: testutil.UpdatePayloadsFromTemplate(t, tplPath),
				es: &EnterpriseSearch{
					RefId:                     ec.String("main-enterprise_search"),
					ResourceId:                ec.String(mock.ValidClusterID),
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
					ZoneCount:                 1,
				},
			},
			want: &models.EnterpriseSearchPayload{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("main-enterprise_search"),
				Plan: &models.EnterpriseSearchPlan{
					EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
					ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
						ZoneCount:               1,
						InstanceConfigurationID: "aws.enterprisesearch.m5d",
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(2048),
						},
						NodeType: &models.EnterpriseSearchNodeTypes{
							Appserver: ec.Bool(true),
							Connector: ec.Bool(true),
							Worker:    ec.Bool(true),
						},
					}},
				},
			},
		},
		{
			name: "parses an enterprise_search resource with explicit topology and config",
			args: args{
				updateResources: testutil.UpdatePayloadsFromTemplate(t, tplPath),
				es: &EnterpriseSearch{
					RefId:                     ec.String("secondary-enterprise_search"),
					ResourceId:                ec.String(mock.ValidClusterID),
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
					Config: &EnterpriseSearchConfig{
						UserSettingsYaml:         ec.String("some.setting: value"),
						UserSettingsOverrideYaml: ec.String("some.setting: override"),
						UserSettingsJson:         ec.String(`{"some.setting":"value"}`),
						UserSettingsOverrideJson: ec.String(`{"some.setting":"override"}`),
					},
					InstanceConfigurationId: ec.String("aws.enterprisesearch.m5d"),
					Size:                    ec.String("4g"),
					ZoneCount:               1,
					NodeTypeAppserver:       ec.Bool(true),
					NodeTypeConnector:       ec.Bool(true),
					NodeTypeWorker:          ec.Bool(true),
				},
			},
			want: &models.EnterpriseSearchPayload{
				ElasticsearchClusterRefID: ec.String("somerefid"),
				Region:                    ec.String("some-region"),
				RefID:                     ec.String("secondary-enterprise_search"),
				Plan: &models.EnterpriseSearchPlan{
					EnterpriseSearch: &models.EnterpriseSearchConfiguration{
						UserSettingsYaml:         "some.setting: value",
						UserSettingsOverrideYaml: "some.setting: override",
						UserSettingsJSON: map[string]interface{}{
							"some.setting": "value",
						},
						UserSettingsOverrideJSON: map[string]interface{}{
							"some.setting": "override",
						},
					},
					ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
						ZoneCount:               1,
						InstanceConfigurationID: "aws.enterprisesearch.m5d",
						Size: &models.TopologySize{
							Resource: ec.String("memory"),
							Value:    ec.Int32(4096),
						},
						NodeType: &models.EnterpriseSearchNodeTypes{
							Appserver: ec.Bool(true),
							Connector: ec.Bool(true),
							Worker:    ec.Bool(true),
						},
					}},
				},
			},
		},
		{
			name: "tries to parse an enterprise_search resource when the template doesn't have an Enterprise Search instance set.",
			args: args{
				updateResources: nil,
				es: &EnterpriseSearch{
					RefId:                     ec.String("tertiary-enterprise_search"),
					ResourceId:                ec.String(mock.ValidClusterID),
					Region:                    ec.String("some-region"),
					ElasticsearchClusterRefId: ec.String("somerefid"),
					Config: &EnterpriseSearchConfig{
						UserSettingsYaml:         ec.String("some.setting: value"),
						UserSettingsOverrideYaml: ec.String("some.setting: value2"),
						UserSettingsJson:         ec.String(`{"some.setting": "value"}`),
						UserSettingsOverrideJson: ec.String(`{"some.setting": "value2"}`),
					},
					InstanceConfigurationId: ec.String("aws.enterprisesearch.m5d"),
					Size:                    ec.String("4g"),
					SizeResource:            ec.String("memory"),
					ZoneCount:               1,
				},
			},
			diags: func() diag.Diagnostics {
				var diags diag.Diagnostics
				diags.AddError("enterprise_search payload error", `enterprise_search specified but deployment template is not configured for it. Use a different template if you wish to add enterprise_search`)
				return diags
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ess types.Object
			diags := tfsdk.ValueFrom(context.Background(), tt.args.es, EnterpriseSearchSchema().GetType(), &ess)
			assert.Nil(t, diags)

			got, diags := EnterpriseSearchesPayload(context.Background(), ess, tt.args.updateResources)
			if tt.diags != nil {
				assert.Equal(t, tt.diags, diags)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
