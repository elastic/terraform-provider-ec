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
	"context"
	"errors"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_importFunc(t *testing.T) {
	deploymentWithImportableVersion := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		Schema: newSchema(),
		State: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-cross-cluster-search-v2",
			"region":                 "us-east-1",
			"version":                "7.9.2",
			"elasticsearch":          []interface{}{map[string]interface{}{}},
		},
	})
	deploymentWithNonImportableVersion := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		Schema: newSchema(),
		State: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-cross-cluster-search-v2",
			"region":                 "us-east-1",
			"version":                "5.6.1",
			"elasticsearch":          []interface{}{map[string]interface{}{}},
		},
	})
	deploymentWithNonImportableVersionSix := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		Schema: newSchema(),
		State: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-cross-cluster-search-v2",
			"region":                 "us-east-1",
			"version":                "6.5.1",
			"elasticsearch":          []interface{}{map[string]interface{}{}},
		},
	})
	type args struct {
		ctx context.Context
		d   *schema.ResourceData
		m   interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]string
		err  error
	}{
		{
			name: "succeeds with an importable version",
			args: args{
				d: deploymentWithImportableVersion,
				m: api.NewMock(mock.New200Response(mock.NewStructBody(models.DeploymentGetResponse{
					Resources: &models.DeploymentResources{Elasticsearch: []*models.ElasticsearchResourceInfo{
						{
							Info: &models.ElasticsearchClusterInfo{
								PlanInfo: &models.ElasticsearchClusterPlansInfo{
									Current: &models.ElasticsearchClusterPlanInfo{
										Plan: &models.ElasticsearchClusterPlan{
											Elasticsearch: &models.ElasticsearchConfiguration{
												Version: "7.9.2",
											},
										},
									},
								},
							},
						},
					}},
				}))),
			},
			want: map[string]string{
				"id": "320b7b540dfc967a7a649c18e2fce4ed",

				"name":                   "my_deployment_name",
				"region":                 "us-east-1",
				"version":                "7.9.2",
				"deployment_template_id": "aws-cross-cluster-search-v2",

				"elasticsearch.#":                   "1",
				"elasticsearch.0.autoscale":         "",
				"elasticsearch.0.cloud_id":          "",
				"elasticsearch.0.snapshot_source.#": "0",
				"elasticsearch.0.config.#":          "0",
				"elasticsearch.0.extension.#":       "0",
				"elasticsearch.0.http_endpoint":     "",
				"elasticsearch.0.https_endpoint":    "",
				"elasticsearch.0.ref_id":            "main-elasticsearch",
				"elasticsearch.0.region":            "",
				"elasticsearch.0.remote_cluster.#":  "0",
				"elasticsearch.0.resource_id":       "",
				"elasticsearch.0.topology.#":        "0",
				"elasticsearch.0.trust_account.#":   "0",
				"elasticsearch.0.trust_external.#":  "0",
			},
		},
		{
			name: "fails with a non importable version (5.6.1)",
			args: args{
				d: deploymentWithNonImportableVersion,
				m: api.NewMock(mock.New200Response(mock.NewStructBody(models.DeploymentGetResponse{
					Resources: &models.DeploymentResources{Elasticsearch: []*models.ElasticsearchResourceInfo{
						{
							Info: &models.ElasticsearchClusterInfo{
								PlanInfo: &models.ElasticsearchClusterPlansInfo{
									Current: &models.ElasticsearchClusterPlanInfo{
										Plan: &models.ElasticsearchClusterPlan{
											Elasticsearch: &models.ElasticsearchConfiguration{
												Version: "5.6.1",
											},
										},
									},
								},
							},
						},
					}},
				}))),
			},
			err: errors.New(`invalid deployment version "5.6.1": minimum supported version is "6.6.0"`),
			want: map[string]string{
				"id": "320b7b540dfc967a7a649c18e2fce4ed",

				"name":                   "my_deployment_name",
				"region":                 "us-east-1",
				"version":                "5.6.1",
				"deployment_template_id": "aws-cross-cluster-search-v2",

				"elasticsearch.#":                   "1",
				"elasticsearch.0.autoscale":         "",
				"elasticsearch.0.cloud_id":          "",
				"elasticsearch.0.snapshot_source.#": "0",
				"elasticsearch.0.config.#":          "0",
				"elasticsearch.0.extension.#":       "0",
				"elasticsearch.0.http_endpoint":     "",
				"elasticsearch.0.https_endpoint":    "",
				"elasticsearch.0.ref_id":            "main-elasticsearch",
				"elasticsearch.0.region":            "",
				"elasticsearch.0.remote_cluster.#":  "0",
				"elasticsearch.0.resource_id":       "",
				"elasticsearch.0.topology.#":        "0",
				"elasticsearch.0.trust_account.#":   "0",
				"elasticsearch.0.trust_external.#":  "0",
			},
		},
		{
			name: "fails with a non importable version (6.5.1)",
			args: args{
				d: deploymentWithNonImportableVersionSix,
				m: api.NewMock(mock.New200Response(mock.NewStructBody(models.DeploymentGetResponse{
					Resources: &models.DeploymentResources{Elasticsearch: []*models.ElasticsearchResourceInfo{
						{
							Info: &models.ElasticsearchClusterInfo{
								PlanInfo: &models.ElasticsearchClusterPlansInfo{
									Current: &models.ElasticsearchClusterPlanInfo{
										Plan: &models.ElasticsearchClusterPlan{
											Elasticsearch: &models.ElasticsearchConfiguration{
												Version: "6.5.1",
											},
										},
									},
								},
							},
						},
					}},
				}))),
			},
			err: errors.New(`invalid deployment version "6.5.1": minimum supported version is "6.6.0"`),
			want: map[string]string{
				"id": "320b7b540dfc967a7a649c18e2fce4ed",

				"name":                   "my_deployment_name",
				"region":                 "us-east-1",
				"version":                "6.5.1",
				"deployment_template_id": "aws-cross-cluster-search-v2",

				"elasticsearch.#":                   "1",
				"elasticsearch.0.autoscale":         "",
				"elasticsearch.0.cloud_id":          "",
				"elasticsearch.0.snapshot_source.#": "0",
				"elasticsearch.0.config.#":          "0",
				"elasticsearch.0.extension.#":       "0",
				"elasticsearch.0.http_endpoint":     "",
				"elasticsearch.0.https_endpoint":    "",
				"elasticsearch.0.ref_id":            "main-elasticsearch",
				"elasticsearch.0.region":            "",
				"elasticsearch.0.remote_cluster.#":  "0",
				"elasticsearch.0.resource_id":       "",
				"elasticsearch.0.topology.#":        "0",
				"elasticsearch.0.trust_account.#":   "0",
				"elasticsearch.0.trust_external.#":  "0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := importFunc(tt.args.ctx, tt.args.d, tt.args.m)
			if tt.err != nil {
				if !assert.EqualError(t, err, tt.err.Error()) {
					t.Error(err)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, tt.args.d.State().Attributes)
		})
	}
}
