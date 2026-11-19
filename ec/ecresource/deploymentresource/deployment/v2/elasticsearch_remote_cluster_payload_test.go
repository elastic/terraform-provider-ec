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
	elasticsearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
)

func Test_handleRemoteClusters(t *testing.T) {
	remoteClusters := elasticsearchv2.ElasticsearchRemoteClusters{
		{
			Alias:           new("alias"),
			DeploymentId:    new("someid"),
			RefId:           new("main-elasticsearch"),
			SkipUnavailable: new(true),
		},
		{
			Alias:           new("alias"),
			DeploymentId:    new("some other id"),
			RefId:           new("main-elasticsearch"),
			SkipUnavailable: new(false),
		},
	}

	deploymentWithRemotes := func(remotes elasticsearchv2.ElasticsearchRemoteClusters) *Deployment {
		return &Deployment{
			Name:                 "my_deployment_name",
			Id:                   "320b7b540dfc967a7a649c18e2fce4ed",
			DeploymentTemplateId: "aws-io-optimized-v2",
			Region:               "us-east-1",
			Version:              "7.7.0",
			Elasticsearch: &elasticsearchv2.Elasticsearch{
				RefId:         new("main-elasticsearch"),
				RemoteCluster: remotes,
			},
		}
	}

	remoteClustersUpdate := func(body string) *api.API {
		return api.NewMock(mock.New202ResponseAssertion(
			&mock.RequestAssertion{
				Header: api.DefaultWriteMockHeaders,
				Host:   api.DefaultMockHost,
				Path:   `/api/v1/deployments/320b7b540dfc967a7a649c18e2fce4ed/elasticsearch/main-elasticsearch/remote-clusters`,
				Method: "PUT",
				Body:   mock.NewStringBody(body + "\n"),
			},
			mock.NewStringBody("{}"),
		))
	}

	// Any API call made with this client fails.
	noRequests := api.NewMock()

	type args struct {
		plan   Deployment
		state  *Deployment
		client *api.API
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "does not update remote clusters on create when the plan has no remote clusters",
			args: args{
				plan:   *deploymentWithRemotes(nil),
				client: noRequests,
			},
		},
		{
			name: "does not update remote clusters on create when the plan has an empty set of remote clusters",
			args: args{
				plan:   *deploymentWithRemotes(elasticsearchv2.ElasticsearchRemoteClusters{}),
				client: noRequests,
			},
		},
		{
			name: "updates remote clusters on create when the plan has remote clusters",
			args: args{
				plan:   *deploymentWithRemotes(remoteClusters),
				client: remoteClustersUpdate(`{"resources":[{"alias":"alias","deployment_id":"someid","elasticsearch_ref_id":"main-elasticsearch","skip_unavailable":true},{"alias":"alias","deployment_id":"some other id","elasticsearch_ref_id":"main-elasticsearch","skip_unavailable":false}]}`),
			},
		},
		{
			name: "does not update remote clusters on update when neither the plan nor the state have remote clusters",
			args: args{
				plan:   *deploymentWithRemotes(nil),
				state:  deploymentWithRemotes(nil),
				client: noRequests,
			},
		},
		{
			name: "does not update remote clusters on update when the plan has an empty set and the state has none",
			args: args{
				plan:   *deploymentWithRemotes(elasticsearchv2.ElasticsearchRemoteClusters{}),
				state:  deploymentWithRemotes(nil),
				client: noRequests,
			},
		},
		{
			name: "does not update remote clusters on update when the remote clusters are unchanged",
			args: args{
				plan:   *deploymentWithRemotes(remoteClusters),
				state:  deploymentWithRemotes(elasticsearchv2.ElasticsearchRemoteClusters{remoteClusters[1], remoteClusters[0]}),
				client: noRequests,
			},
		},
		{
			name: "updates remote clusters on update when remote clusters are added",
			args: args{
				plan:   *deploymentWithRemotes(remoteClusters),
				state:  deploymentWithRemotes(nil),
				client: remoteClustersUpdate(`{"resources":[{"alias":"alias","deployment_id":"someid","elasticsearch_ref_id":"main-elasticsearch","skip_unavailable":true},{"alias":"alias","deployment_id":"some other id","elasticsearch_ref_id":"main-elasticsearch","skip_unavailable":false}]}`),
			},
		},
		{
			name: "updates remote clusters on update when the remote clusters change",
			args: args{
				plan:   *deploymentWithRemotes(remoteClusters[:1]),
				state:  deploymentWithRemotes(remoteClusters),
				client: remoteClustersUpdate(`{"resources":[{"alias":"alias","deployment_id":"someid","elasticsearch_ref_id":"main-elasticsearch","skip_unavailable":true}]}`),
			},
		},
		{
			name: "removes remote clusters on update when they are removed from the plan",
			args: args{
				plan:   *deploymentWithRemotes(nil),
				state:  deploymentWithRemotes(remoteClusters),
				client: remoteClustersUpdate(`{"resources":[]}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := DeploymentSchema()

			var planTF DeploymentTF

			diags := tfsdk.ValueFrom(context.Background(), tt.args.plan, schema.Type(), &planTF)
			assert.Nil(t, diags)

			stateES := types.ObjectNull(planTF.Elasticsearch.AttributeTypes(context.Background()))
			if tt.args.state != nil {
				var stateTF DeploymentTF
				diags = tfsdk.ValueFrom(context.Background(), tt.args.state, schema.Type(), &stateTF)
				assert.Nil(t, diags)
				stateES = stateTF.Elasticsearch
			}

			diags = HandleRemoteClusters(context.Background(), tt.args.client, planTF.Id.ValueString(), planTF.Elasticsearch, stateES)
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
						Alias:           new("alias"),
						DeploymentId:    new("someid"),
						RefId:           new("main-elasticsearch"),
						SkipUnavailable: new(true),
					},
					{
						Alias:        new("alias"),
						DeploymentId: new("some other id"),
						RefId:        new("main-elasticsearch"),
					},
				},
			},
			want: &models.RemoteResources{Resources: []*models.RemoteResourceRef{
				{
					Alias:              new("alias"),
					DeploymentID:       new("someid"),
					ElasticsearchRefID: new("main-elasticsearch"),
					SkipUnavailable:    new(true),
				},
				{
					Alias:              new("alias"),
					DeploymentID:       new("some other id"),
					ElasticsearchRefID: new("main-elasticsearch"),
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var remoteClustersTF types.Set
			diags := tfsdk.ValueFrom(context.Background(), tt.args.remoteClusters, elasticsearchv2.ElasticsearchRemoteClusterSchema().GetType(), &remoteClustersTF)
			assert.Nil(t, diags)

			got, diags := elasticsearchv2.ElasticsearchRemoteClustersPayload(context.Background(), remoteClustersTF)
			assert.Nil(t, diags)
			assert.Equal(t, tt.want, got)
		})
	}
}
