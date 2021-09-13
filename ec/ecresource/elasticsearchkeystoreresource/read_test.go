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
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func Test_modelToState(t *testing.T) {
	esKeystoreSchemaArg := schema.TestResourceDataRaw(t, newSchema(), map[string]interface{}{
		"deployment_id": mock.ValidClusterID,
		"setting_name":  "my_secret",
		"value":         "supersecret",
		"as_file":       false, // This field is overridden.
	})
	esKeystoreSchemaArg.SetId(mock.ValidClusterID)

	esKeystoreSchemaArgMissing := schema.TestResourceDataRaw(t, newSchema(), map[string]interface{}{
		"deployment_id": mock.ValidClusterID,
		"setting_name":  "my_secret",
		"value":         "supersecret",
		"as_file":       false,
	})
	esKeystoreSchemaArgMissing.SetId(mock.ValidClusterID)

	type args struct {
		d   *schema.ResourceData
		res *models.KeystoreContents
	}
	tests := []struct {
		name string
		args args
		want *schema.ResourceData
		err  error
	}{
		{
			name: "flattens the keystore secret (not really since the value is not returned)",
			args: args{
				d: esKeystoreSchemaArg,
				res: &models.KeystoreContents{
					Secrets: map[string]models.KeystoreSecret{
						"my_secret": {
							AsFile: ec.Bool(true),
						},
						"some_other_secret": {
							AsFile: ec.Bool(false),
						},
					},
				},
			},
			want: newResourceData(t, resDataParams{
				ID: mock.ValidClusterID,
				Resources: map[string]interface{}{
					"deployment_id": mock.ValidClusterID,
					"setting_name":  "my_secret",
					"value":         "supersecret",
					"as_file":       true,
				},
			}),
		},
		{
			name: "unsets the ID when our secret is not in the returned list of secrets",
			args: args{
				d: esKeystoreSchemaArgMissing,
				res: &models.KeystoreContents{
					Secrets: map[string]models.KeystoreSecret{
						"my_other_secret": {
							AsFile: ec.Bool(true),
						},
						"some_other_secret": {
							AsFile: ec.Bool(false),
						},
					},
				},
			},
			want: newResourceData(t, resDataParams{
				Resources: map[string]interface{}{},
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := modelToState(tt.args.d, tt.args.res)
			if tt.err != nil || err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}

			wantState := tt.want.State()
			gotState := tt.args.d.State()

			if wantState != nil && gotState != nil {
				assert.Equal(t, wantState.Attributes, gotState.Attributes)
				return
			}
		})
	}
}
