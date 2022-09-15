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
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_handleRemoteClusters(t *testing.T) {
	deploymentEmptyRD := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleDeploymentEmptyRD(),
		Schema: newSchema(),
	})
	deploymentWithRemotesRD := util.NewResourceData(t, util.ResDataParams{
		ID: mock.ValidClusterID,
		State: map[string]interface{}{
			"name":                   "my_deployment_name",
			"deployment_template_id": "aws-io-optimized-v2",
			"region":                 "us-east-1",
			"version":                "7.7.0",
			"elasticsearch": []interface{}{map[string]interface{}{
				"remote_cluster": []interface{}{
					map[string]interface{}{
						"alias":            "alias",
						"deployment_id":    "someid",
						"ref_id":           "main-elasticsearch",
						"skip_unavailable": true,
					},
					map[string]interface{}{
						"deployment_id": "some other id",
						"ref_id":        "main-elasticsearch",
					},
				},
			}},
		},
		Schema: newSchema(),
	})
	type args struct {
		d      *schema.ResourceData
		client *api.API
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "returns when the resource has no remote clusters",
			args: args{
				d:      deploymentEmptyRD,
				client: api.NewMock(),
			},
		},
		{
			name: "flattens the remote clusters",
			args: args{
				d: deploymentWithRemotesRD,
				client: api.NewMock(mock.New202ResponseAssertion(
					&mock.RequestAssertion{
						Header: api.DefaultWriteMockHeaders,
						Host:   api.DefaultMockHost,
						Path:   `/api/v1/deployments/320b7b540dfc967a7a649c18e2fce4ed/elasticsearch/main-elasticsearch/remote-clusters`,
						Method: "PUT",
						Body:   mock.NewStringBody(`{"resources":[{"alias":"alias","deployment_id":"someid","elasticsearch_ref_id":"main-elasticsearch","skip_unavailable":true},{"alias":"","deployment_id":"some other id","elasticsearch_ref_id":"main-elasticsearch","skip_unavailable":false}]}` + "\n"),
					},
					mock.NewStringBody("{}"),
				)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handleRemoteClusters(tt.args.d, tt.args.client)
			if !assert.Equal(t, tt.err, err) {
				t.Error(err)
			}
		})
	}
}

func Test_expandRemoteClusters(t *testing.T) {
	type args struct {
		set *schema.Set
	}
	tests := []struct {
		name string
		args args
		want *models.RemoteResources
	}{
		{
			name: "wants no error or empty res",
			args: args{set: newElasticsearchRemoteSet()},
			want: &models.RemoteResources{Resources: []*models.RemoteResourceRef{}},
		},
		{
			name: "expands remotes",
			args: args{set: newElasticsearchRemoteSet([]interface{}{
				map[string]interface{}{
					"alias":            "alias",
					"deployment_id":    "someid",
					"ref_id":           "main-elasticsearch",
					"skip_unavailable": true,
				},
				map[string]interface{}{
					"deployment_id": "some other id",
					"ref_id":        "main-elasticsearch",
				},
			}...)},
			want: &models.RemoteResources{Resources: []*models.RemoteResourceRef{
				{
					DeploymentID:       ec.String("some other id"),
					ElasticsearchRefID: ec.String("main-elasticsearch"),
				},
				{
					Alias:              ec.String("alias"),
					DeploymentID:       ec.String("someid"),
					ElasticsearchRefID: ec.String("main-elasticsearch"),
					SkipUnavailable:    ec.Bool(true),
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandRemoteClusters(tt.args.set)
			assert.Equal(t, tt.want, got)
		})
	}
}
