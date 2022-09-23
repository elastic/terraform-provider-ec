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

package elasticsearchkeystoreresource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
)

func Test_expandModel(t *testing.T) {
	type args struct {
		state modelV0
	}
	tests := []struct {
		name string
		args args
		want *models.KeystoreContents
	}{
		{
			name: "parses the resource with a string value",
			args: args{state: modelV0{

				ID:           types.String{Value: "some-random-id"},
				DeploymentID: types.String{Value: mock.ValidClusterID},
				SettingName:  types.String{Value: "my_secret"},
				Value:        types.String{Value: "supersecret"},
				AsFile:       types.Bool{Value: false},
			}},
			want: &models.KeystoreContents{
				Secrets: map[string]models.KeystoreSecret{
					"my_secret": {
						AsFile: ec.Bool(false),
						Value:  "supersecret",
					},
				},
			},
		},
		{
			name: "parses the resource with a json formatted value",
			args: args{state: modelV0{

				ID:           types.String{Value: "some-random-id"},
				DeploymentID: types.String{Value: mock.ValidClusterID},
				SettingName:  types.String{Value: "my_secret"},
				Value: types.String{Value: `{
    "type": "service_account",
    "project_id": "project-id",
    "private_key_id": "key-id",
    "private_key": "-----BEGIN PRIVATE KEY-----\nprivate-key\n-----END PRIVATE KEY-----\n",
    "client_email": "service-account-email",
    "client_id": "client-id",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "token_uri": "https://accounts.google.com/o/oauth2/token",
    "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
    "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/service-account-email"
}`},
				AsFile: types.Bool{Value: true},
			},
			},
			want: &models.KeystoreContents{
				Secrets: map[string]models.KeystoreSecret{
					"my_secret": {
						AsFile: ec.Bool(true),
						Value: map[string]interface{}{
							"type":                        "service_account",
							"project_id":                  "project-id",
							"private_key_id":              "key-id",
							"private_key":                 "-----BEGIN PRIVATE KEY-----\nprivate-key\n-----END PRIVATE KEY-----\n",
							"client_email":                "service-account-email",
							"client_id":                   "client-id",
							"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
							"token_uri":                   "https://accounts.google.com/o/oauth2/token",
							"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
							"client_x509_cert_url":        "https://www.googleapis.com/robot/v1/metadata/x509/service-account-email",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandModel(context.Background(), tt.args.state)
			assert.Equal(t, tt.want, got)
		})
	}
}
