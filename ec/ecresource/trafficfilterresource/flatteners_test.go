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
			{Source: ec.String("1.1.1.1")},
			{Source: ec.String("0.0.0.0/0")},
		},
	}

	wantTrafficFilter := util.NewResourceData(t, util.ResDataParams{
		ID:     "some-random-id",
		State:  newSampleTrafficFilter(),
		Schema: newSchema(),
	})
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
