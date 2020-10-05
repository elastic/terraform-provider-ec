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

package trafficfilterassocresource

import (
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_flatten(t *testing.T) {
	rd := util.NewResourceData(t, util.ResDataParams{
		Resources: newSampleTrafficFilterAssociation(),
		ID:        "123451",
		Schema:    newSchema(),
	})

	wantNotFoundRd := util.NewResourceData(t, util.ResDataParams{
		Resources: newSampleTrafficFilterAssociation(),
		ID:        "123451",
		Schema:    newSchema(),
	})

	_ = wantNotFoundRd.Set("deployment_id", "")
	_ = wantNotFoundRd.Set("traffic_filter_id", "")
	type args struct {
		res *models.TrafficFilterRulesetInfo
		d   *schema.ResourceData
	}
	tests := []struct {
		name string
		args args
		want *schema.ResourceData
		err  error
	}{
		{
			name: "empty response returns nil",
			args: args{d: rd},
		},
		{
			name: "flattens the response",
			args: args{d: rd,
				res: &models.TrafficFilterRulesetInfo{
					Associations: []*models.FilterAssociation{
						{
							EntityType: ec.String("cluster"),
							ID:         ec.String("someid"),
						},
						{
							EntityType: ec.String(entityType),
							ID:         ec.String(mock.ValidClusterID),
						},
					},
				},
			},
			want: rd,
		},
		{
			name: "flattens the response even when the association has been removed externally",
			args: args{d: rd,
				res: &models.TrafficFilterRulesetInfo{
					Associations: []*models.FilterAssociation{{
						EntityType: ec.String("cluster"),
						ID:         ec.String("someid"),
					}},
				},
			},
			want: wantNotFoundRd,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := flatten(tt.args.res, tt.args.d)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.args.res == nil {
				return
			}

			wantState := tt.want.State()
			if wantState == nil {
				tt.want.SetId("some")
				wantState = tt.want.State()
			}

			gotState := tt.args.d.State()
			if gotState == nil {
				tt.args.d.SetId("some")
				gotState = tt.want.State()
			}

			assert.Equal(t, wantState.Attributes, gotState.Attributes)
		})
	}
}
