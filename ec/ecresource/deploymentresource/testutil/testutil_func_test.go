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

package testutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
)

func Test_parseDeploymentTemplate(t *testing.T) {
	var contents = `{"id": "deployment-template-name","deployment_template":
	{"resources": {
		"elasticsearch": [{
			"plan": { "deployment_template": {}}
		}]
	}}}`
	if err := os.WriteFile("test.json", []byte(contents), 0660); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test.json")
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want *models.DeploymentTemplateInfoV2
	}{
		{
			name: "Enrich DT",
			args: args{name: "test.json"},
			want: &models.DeploymentTemplateInfoV2{ID: ec.String("deployment-template-name"), DeploymentTemplate: &models.DeploymentCreateRequest{
				Resources: &models.DeploymentCreateResources{Elasticsearch: []*models.ElasticsearchPayload{
					{Plan: &models.ElasticsearchClusterPlan{DeploymentTemplate: &models.DeploymentTemplateReference{
						ID: ec.String("deployment-template-name"),
					}}},
				}},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseDeploymentTemplate(t, tt.args.name)
			assert.Equal(t, tt.want, got)
		})
	}
}
