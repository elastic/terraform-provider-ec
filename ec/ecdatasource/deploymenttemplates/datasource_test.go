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

package deploymenttemplates

import (
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_mapResponseToModel(t *testing.T) {
	tests := []struct {
		name        string
		apiResponse []*models.DeploymentTemplateInfoV2
		showHidden  bool
		expected    []deploymentTemplateModel
	}{
		{
			name:       "should filter out any hidden templates for showHidden=false",
			showHidden: false,
			apiResponse: []*models.DeploymentTemplateInfoV2{
				{
					ID:          ec.String("id-nonhidden"),
					Name:        ec.String("name-nonhidden"),
					Description: "description nonhidden",
					MinVersion:  "",
					Metadata:    []*models.MetadataItem{},
				},
				{
					ID:          ec.String("id-hidden"),
					Name:        ec.String("name-hidden"),
					Description: "description hidden",
					MinVersion:  "7.17.0",
					Metadata: []*models.MetadataItem{
						{
							Key:   ec.String("anotherkey"),
							Value: ec.String("false"),
						},
						{
							Key:   ec.String("hidden"),
							Value: ec.String("true"),
						},
					},
				},
			},
			expected: []deploymentTemplateModel{
				{
					ID:              types.StringValue("id-nonhidden"),
					Name:            types.StringValue("name-nonhidden"),
					Description:     types.StringValue("description nonhidden"),
					MinStackVersion: types.StringValue(""),
					Hidden:          types.BoolValue(false),
				},
			},
		},
		{
			name:       "should show all templates for showHidden=true",
			showHidden: true,
			apiResponse: []*models.DeploymentTemplateInfoV2{
				{
					ID:          ec.String("id-nonhidden"),
					Name:        ec.String("name-nonhidden"),
					Description: "description nonhidden",
					MinVersion:  "",
					Metadata:    []*models.MetadataItem{},
				},
				{
					ID:          ec.String("id-hidden"),
					Name:        ec.String("name-hidden"),
					Description: "description hidden",
					MinVersion:  "7.17.0",
					Metadata: []*models.MetadataItem{
						{
							Key:   ec.String("anotherkey"),
							Value: ec.String("false"),
						},
						{
							Key:   ec.String("hidden"),
							Value: ec.String("true"),
						},
					},
				},
			},
			expected: []deploymentTemplateModel{
				{
					ID:              types.StringValue("id-nonhidden"),
					Name:            types.StringValue("name-nonhidden"),
					Description:     types.StringValue("description nonhidden"),
					MinStackVersion: types.StringValue(""),
					Hidden:          types.BoolValue(false),
				},
				{
					ID:              types.StringValue("id-hidden"),
					Name:            types.StringValue("name-hidden"),
					Description:     types.StringValue("description hidden"),
					MinStackVersion: types.StringValue("7.17.0"),
					Hidden:          types.BoolValue(true),
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := mapResponseToModel(test.apiResponse, test.showHidden)
			assert.Equal(t, test.expected, actual)
		})
	}
}
