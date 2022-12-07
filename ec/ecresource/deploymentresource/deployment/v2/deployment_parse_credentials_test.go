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
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/stretchr/testify/assert"
)

func Test_parseCredentials(t *testing.T) {
	type args struct {
		dep       Deployment
		resources []*models.DeploymentResource
	}
	tests := []struct {
		name string
		args args
		want Deployment
		err  error
	}{
		{
			name: "Parses credentials",
			args: args{
				dep: Deployment{},
				resources: []*models.DeploymentResource{{
					Credentials: &models.ClusterCredentials{
						Username: ec.String("my-username"),
						Password: ec.String("my-password"),
					},
					SecretToken: "some-secret-token",
				}},
			},
			want: Deployment{
				ElasticsearchUsername: "my-username",
				ElasticsearchPassword: "my-password",
				ApmSecretToken:        ec.String("some-secret-token"),
			},
		},
		{
			name: "when no credentials are passed, it doesn't overwrite them",
			args: args{
				dep: Deployment{
					ElasticsearchUsername: "my-username",
					ElasticsearchPassword: "my-password",
					ApmSecretToken:        ec.String("some-secret-token"),
				},
				resources: []*models.DeploymentResource{
					{},
				},
			},
			want: Deployment{
				ElasticsearchUsername: "my-username",
				ElasticsearchPassword: "my-password",
				ApmSecretToken:        ec.String("some-secret-token"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.dep.parseCredentials(tt.args.resources)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, tt.args.dep)
			}
		})
	}
}
