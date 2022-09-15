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

package trafficfilterresource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_modelToState(t *testing.T) {
	trafficFilterSchemaArg := schema.TestResourceDataRaw(t, newSchema(), nil)
	trafficFilterSchemaArg.SetId("some-random-id")

	remoteState := models.TrafficFilterRulesetInfo{
		ID:               ec.String("some-random-id"),
		Name:             ec.String("my traffic filter"),
		Type:             ec.String("ip"),
		IncludeByDefault: ec.Bool(false),
		Region:           ec.String("us-east-1"),
		Rules: []*models.TrafficFilterRule{
			{Source: "1.1.1.1"},
			{Source: "0.0.0.0/0"},
		},
	}

	trafficFilterSchemaArgMultipleR := schema.TestResourceDataRaw(t, newSchema(), nil)
	trafficFilterSchemaArgMultipleR.SetId("some-random-id")

	remoteStateMultipleRules := models.TrafficFilterRulesetInfo{
		ID:               ec.String("some-random-id"),
		Name:             ec.String("my traffic filter"),
		Type:             ec.String("ip"),
		IncludeByDefault: ec.Bool(false),
		Region:           ec.String("us-east-1"),
		Rules: []*models.TrafficFilterRule{
			{Source: "1.1.1.0/16"},
			{Source: "1.1.1.1/24"},
			{Source: "0.0.0.0/0"},
			{Source: "1.1.1.1"},
		},
	}

	trafficFilterSchemaArgMultipleRWithDesc := schema.TestResourceDataRaw(t, newSchema(), nil)
	trafficFilterSchemaArgMultipleRWithDesc.SetId("some-random-id")

	remoteStateMultipleRulesWithDesc := models.TrafficFilterRulesetInfo{
		ID:               ec.String("some-random-id"),
		Name:             ec.String("my traffic filter"),
		Type:             ec.String("ip"),
		IncludeByDefault: ec.Bool(false),
		Region:           ec.String("us-east-1"),
		Rules: []*models.TrafficFilterRule{
			{Source: "1.1.1.0/16", Description: "some network"},
			{Source: "1.1.1.1/24", Description: "a specific IP"},
			{Source: "0.0.0.0/0", Description: "all internet traffic"},
		},
	}

	wantTrafficFilter := util.NewResourceData(t, util.ResDataParams{
		ID:     "some-random-id",
		State:  newSampleTrafficFilter(),
		Schema: newSchema(),
	})
	wantTrafficFilterMultipleR := util.NewResourceData(t, util.ResDataParams{
		ID: "some-random-id",
		State: map[string]interface{}{
			"name":               "my traffic filter",
			"type":               "ip",
			"include_by_default": false,
			"region":             "us-east-1",
			"rule": []interface{}{
				map[string]interface{}{
					"source": "1.1.1.1/24",
				},
				map[string]interface{}{
					"source": "1.1.1.0/16",
				},
				map[string]interface{}{
					"source": "0.0.0.0/0",
				},
				map[string]interface{}{
					"source": "1.1.1.1",
				},
			},
		},
		Schema: newSchema(),
	})
	wantTrafficFilterMultipleRWithDesc := util.NewResourceData(t, util.ResDataParams{
		ID: "some-random-id",
		State: map[string]interface{}{
			"name":               "my traffic filter",
			"type":               "ip",
			"include_by_default": false,
			"region":             "us-east-1",
			"rule": []interface{}{
				map[string]interface{}{
					"source":      "1.1.1.1/24",
					"description": "a specific IP",
				},
				map[string]interface{}{
					"source":      "1.1.1.0/16",
					"description": "some network",
				},
				map[string]interface{}{
					"source":      "0.0.0.0/0",
					"description": "all internet traffic",
				},
			},
		},
		Schema: newSchema(),
	})

	azurePLSchemaArg := schema.TestResourceDataRaw(t, newSchema(), nil)
	azurePLSchemaArg.SetId("some-random-id")

	azurePLRemoteState := models.TrafficFilterRulesetInfo{
		ID:               ec.String("some-random-id"),
		Name:             ec.String("my traffic filter"),
		Type:             ec.String("azure_private_endpoint"),
		IncludeByDefault: ec.Bool(false),
		Region:           ec.String("azure-australiaeast"),
		Rules: []*models.TrafficFilterRule{
			{
				AzureEndpointGUID: "1231312-1231-1231-1231-1231312",
				AzureEndpointName: "my-azure-pl",
			},
		},
	}

	type args struct {
		d   *schema.ResourceData
		res *models.TrafficFilterRulesetInfo
	}
	tests := []struct {
		name string
		args args
		err  error
		want *schema.ResourceData
	}{
		{
			name: "flattens the resource",
			args: args{d: trafficFilterSchemaArg, res: &remoteState},
			want: wantTrafficFilter,
		},
		{
			name: "flattens the resource with multiple rules",
			args: args{d: trafficFilterSchemaArgMultipleR, res: &remoteStateMultipleRules},
			want: wantTrafficFilterMultipleR,
		},
		{
			name: "flattens the resource with multiple rules with descriptions",
			args: args{d: trafficFilterSchemaArgMultipleRWithDesc, res: &remoteStateMultipleRulesWithDesc},
			want: wantTrafficFilterMultipleRWithDesc,
		},
		{
			name: "flattens the resource with multiple rules with descriptions",
			args: args{d: azurePLSchemaArg, res: &azurePLRemoteState},
			want: util.NewResourceData(t, util.ResDataParams{
				ID: "some-random-id",
				State: map[string]interface{}{
					"name":               "my traffic filter",
					"type":               "azure_private_endpoint",
					"include_by_default": false,
					"region":             "azure-australiaeast",
					"rule": []interface{}{map[string]interface{}{
						"azure_endpoint_guid": "1231312-1231-1231-1231-1231312",
						"azure_endpoint_name": "my-azure-pl",
					}},
				},
				Schema: newSchema(),
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := modelToState(tt.args.d, tt.args.res)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want.State().Attributes, tt.args.d.State().Attributes)
		})
	}
}
