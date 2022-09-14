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

package deploymentresource

import (
	"errors"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/stretchr/testify/assert"
)

func Test_expandIntegrationsServerResources(t *testing.T) {
	tplPath := "testdata/template-ece-3.0.0-default.json"
	tpl := func() *models.IntegrationsServerPayload {
		return integrationsServerResource(parseDeploymentTemplate(t,
			tplPath,
		))
	}
	type args struct {
		ess []interface{}
		tpl *models.IntegrationsServerPayload
	}
	tests := []struct {
		name string
		args args
		want []*models.IntegrationsServerPayload
		err  error
	}{
		{
			name: "returns nil when there's no resources",
		},
		{
			name: "parses an Integrations Server resource with explicit topology",
			args: args{
				tpl: tpl(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":                       "main-integrations_server",
						"resource_id":                  mock.ValidClusterID,
						"region":                       "some-region",
						"elasticsearch_cluster_ref_id": "somerefid",
						"topology": []interface{}{map[string]interface{}{
							"instance_configuration_id": "integrations.server",
							"size":                      "2g",
							"size_resource":             "memory",
							"zone_count":                1,
						}},
					},
				},
			},
			want: []*models.IntegrationsServerPayload{
				{
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
		},
		{
			name: "parses an Integrations Server resource with invalid instance_configuration_id",
			args: args{
				tpl: tpl(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":                       "main-integrations_server",
						"resource_id":                  mock.ValidClusterID,
						"region":                       "some-region",
						"elasticsearch_cluster_ref_id": "somerefid",
						"topology": []interface{}{map[string]interface{}{
							"instance_configuration_id": "so invalid",
							"size":                      "2g",
							"size_resource":             "memory",
							"zone_count":                1,
						}},
					},
				},
			},
			err: errors.New(`IntegrationsServer topology: invalid instance_configuration_id: "so invalid" doesn't match any of the deployment template instance configurations`),
		},
		{
			name: "parses an Integrations Server resource with no topology",
			args: args{
				tpl: tpl(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":                       "main-integrations_server",
						"resource_id":                  mock.ValidClusterID,
						"region":                       "some-region",
						"elasticsearch_cluster_ref_id": "somerefid",
					},
				},
			},
			want: []*models.IntegrationsServerPayload{
				{
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
		},
		{
			name: "parses an Integrations Server resource with a topology element but no instance_configuration_id",
			args: args{
				tpl: tpl(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":                       "main-integrations_server",
						"resource_id":                  mock.ValidClusterID,
						"region":                       "some-region",
						"elasticsearch_cluster_ref_id": "somerefid",
						"topology": []interface{}{map[string]interface{}{
							"size":          "2g",
							"size_resource": "memory",
						}},
					},
				},
			},
			want: []*models.IntegrationsServerPayload{
				{
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
		},
		{
			name: "parses an Integrations Server resource with multiple topology elements but no instance_configuration_id",
			args: args{
				tpl: tpl(),
				ess: []interface{}{
					map[string]interface{}{
						"ref_id":                       "main-integrations_server",
						"resource_id":                  mock.ValidClusterID,
						"region":                       "some-region",
						"elasticsearch_cluster_ref_id": "somerefid",
						"topology": []interface{}{
							map[string]interface{}{
								"size":          "2g",
								"size_resource": "memory",
							}, map[string]interface{}{
								"size":          "2g",
								"size_resource": "memory",
							},
						},
					},
				},
			},
			err: errors.New("IntegrationsServer topology: invalid instance_configuration_id: \"\" doesn't match any of the deployment template instance configurations"),
		},
		{
			name: "parses an Integrations Server resource with explicit topology and some config",
			args: args{
				tpl: tpl(),
				ess: []interface{}{map[string]interface{}{
					"ref_id":                       "tertiary-integrations_server",
					"elasticsearch_cluster_ref_id": "somerefid",
					"resource_id":                  mock.ValidClusterID,
					"region":                       "some-region",
					"config": []interface{}{map[string]interface{}{
						"user_settings_yaml":          "some.setting: value",
						"user_settings_override_yaml": "some.setting: value2",
						"user_settings_json":          "{\"some.setting\": \"value\"}",
						"user_settings_override_json": "{\"some.setting\": \"value2\"}",
						"debug_enabled":               true,
					}},
					"topology": []interface{}{map[string]interface{}{
						"instance_configuration_id": "integrations.server",
						"size":                      "4g",
						"size_resource":             "memory",
						"zone_count":                1,
					}},
				}},
			},
			want: []*models.IntegrationsServerPayload{{
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
			}},
		},
		{
			name: "tries to parse an integrations_server resource when the template doesn't have an Integrations Server instance set.",
			args: args{
				tpl: nil,
				ess: []interface{}{map[string]interface{}{
					"ref_id":                       "tertiary-integrations_server",
					"elasticsearch_cluster_ref_id": "somerefid",
					"resource_id":                  mock.ValidClusterID,
					"region":                       "some-region",
					"topology": []interface{}{map[string]interface{}{
						"instance_configuration_id": "integrations.server",
						"size":                      "4g",
						"size_resource":             "memory",
						"zone_count":                1,
					}},
					"config": []interface{}{map[string]interface{}{
						"debug_enabled": true,
					}},
				}},
			},
			err: errors.New("IntegrationsServer specified but deployment template is not configured for it. Use a different template if you wish to add IntegrationsServer"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandIntegrationsServerResources(tt.args.ess, tt.args.tpl)
			if !assert.Equal(t, tt.err, err) {
				t.Error(err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
