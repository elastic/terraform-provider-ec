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

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_expandModel(t *testing.T) {
	trafficFilterRD := util.NewResourceData(t, util.ResDataParams{
		ID:     "some-random-id",
		State:  newSampleTrafficFilter(),
		Schema: newSchema(),
	})
	trafficFilterMultipleRD := util.NewResourceData(t, util.ResDataParams{
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
	type args struct {
		d *schema.ResourceData
	}
	tests := []struct {
		name string
		args args
		want *models.TrafficFilterRulesetRequest
	}{
		{
			name: "parses the resource",
			args: args{d: trafficFilterRD},
			want: &models.TrafficFilterRulesetRequest{
				Name:             ec.String("my traffic filter"),
				Type:             ec.String("ip"),
				IncludeByDefault: ec.Bool(false),
				Region:           ec.String("us-east-1"),
				Rules: []*models.TrafficFilterRule{
					{Source: ec.String("0.0.0.0/0")},
					{Source: ec.String("1.1.1.1")},
				},
			},
		},
		{
			name: "parses the resource with a lot of traffic rules",
			args: args{d: trafficFilterMultipleRD},
			want: &models.TrafficFilterRulesetRequest{
				Name:             ec.String("my traffic filter"),
				Type:             ec.String("ip"),
				IncludeByDefault: ec.Bool(false),
				Region:           ec.String("us-east-1"),
				Rules: []*models.TrafficFilterRule{
					{Source: ec.String("0.0.0.0/0")},
					{Source: ec.String("1.1.1.1")},
					{Source: ec.String("1.1.1.0/16")},
					{Source: ec.String("1.1.1.1/24")},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandModel(tt.args.d)
			assert.Equal(t, tt.want, got)
		})
	}
}
