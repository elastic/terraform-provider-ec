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

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	elasticsearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
)

func Test_handleRemoteClusters(t *testing.T) {
	type args struct {
		plan   Deployment
		client *api.API
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "returns when the resource has no remote clusters",
			args: args{
				plan: Deployment{
					Id: "320b7b540dfc967a7a649c18e2fce4ed",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId: ec.String("main-elasticsearch"),
					},
				},
				client: api.NewMock(mock.New202ResponseAssertion(
					&mock.RequestAssertion{
						Header: api.DefaultWriteMockHeaders,
						Host:   api.DefaultMockHost,
						Path:   `/api/v1/deployments/320b7b540dfc967a7a649c18e2fce4ed/elasticsearch/main-elasticsearch/remote-clusters`,
						Method: "PUT",
						Body:   mock.NewStringBody(`{"resources":[]}` + "\n"),
					},
					mock.NewStringBody("{}"),
				)),
			},
		},
		{
			name: "read the remote clusters",
			args: args{
				client: api.NewMock(mock.New202ResponseAssertion(
					&mock.RequestAssertion{
						Header: api.DefaultWriteMockHeaders,
						Host:   api.DefaultMockHost,
						Path:   `/api/v1/deployments/320b7b540dfc967a7a649c18e2fce4ed/elasticsearch/main-elasticsearch/remote-clusters`,
						Method: "PUT",
						Body:   mock.NewStringBody(`{"resources":[{"alias":"alias","deployment_id":"someid","elasticsearch_ref_id":"main-elasticsearch","skip_unavailable":true},{"alias":"alias","deployment_id":"some other id","elasticsearch_ref_id":"main-elasticsearch","skip_unavailable":false}]}` + "\n"),
					},
					mock.NewStringBody("{}"),
				)),
				plan: Deployment{
					Name:                 "my_deployment_name",
					Id:                   "320b7b540dfc967a7a649c18e2fce4ed",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.7.0",
					Elasticsearch: &elasticsearchv2.Elasticsearch{
						RefId: ec.String("main-elasticsearch"),
						RemoteCluster: elasticsearchv2.ElasticsearchRemoteClusters{
							{
								Alias:           ec.String("alias"),
								DeploymentId:    ec.String("someid"),
								RefId:           ec.String("main-elasticsearch"),
								SkipUnavailable: ec.Bool(true),
							},
							{
								Alias:           ec.String("alias"),
								DeploymentId:    ec.String("some other id"),
								RefId:           ec.String("main-elasticsearch"),
								SkipUnavailable: ec.Bool(false),
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := DeploymentSchema()

			var planTF DeploymentTF

			diags := tfsdk.ValueFrom(context.Background(), tt.args.plan, schema.Type(), &planTF)
			assert.Nil(t, diags)

			diags = HandleRemoteClusters(context.Background(), tt.args.client, planTF.Id.Value, planTF.Elasticsearch)
			assert.Nil(t, diags)
		})
	}
}

func Test_writeRemoteClusters(t *testing.T) {
	type args struct {
		remoteClusters elasticsearchv2.ElasticsearchRemoteClusters
	}
	tests := []struct {
		name string
		args args
		want *models.RemoteResources
	}{
		{
			name: "wants no error or empty res",
			args: args{
				remoteClusters: elasticsearchv2.ElasticsearchRemoteClusters{},
			},
			want: &models.RemoteResources{Resources: []*models.RemoteResourceRef{}},
		},
		{
			name: "expands remotes",
			args: args{
				remoteClusters: elasticsearchv2.ElasticsearchRemoteClusters{
					{
						Alias:           ec.String("alias"),
						DeploymentId:    ec.String("someid"),
						RefId:           ec.String("main-elasticsearch"),
						SkipUnavailable: ec.Bool(true),
					},
					{
						Alias:        ec.String("alias"),
						DeploymentId: ec.String("some other id"),
						RefId:        ec.String("main-elasticsearch"),
					},
				},
			},
			want: &models.RemoteResources{Resources: []*models.RemoteResourceRef{
				{
					Alias:              ec.String("alias"),
					DeploymentID:       ec.String("someid"),
					ElasticsearchRefID: ec.String("main-elasticsearch"),
					SkipUnavailable:    ec.Bool(true),
				},
				{
					Alias:              ec.String("alias"),
					DeploymentID:       ec.String("some other id"),
					ElasticsearchRefID: ec.String("main-elasticsearch"),
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var remoteClustersTF types.Set
			diags := tfsdk.ValueFrom(context.Background(), tt.args.remoteClusters, elasticsearchv2.ElasticsearchRemoteClusterSchema().FrameworkType(), &remoteClustersTF)
			assert.Nil(t, diags)

			got, diags := elasticsearchv2.ElasticsearchRemoteClustersPayload(context.Background(), remoteClustersTF)
			assert.Nil(t, diags)
			assert.Equal(t, tt.want, got)
		})
	}
}
