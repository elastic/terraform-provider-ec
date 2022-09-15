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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_delete(t *testing.T) {
	tc500Err := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleTrafficFilter(),
		Schema: newSchema(),
	})
	wantTC500 := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleTrafficFilter(),
		Schema: newSchema(),
	})

	tc404Err := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleTrafficFilter(),
		Schema: newSchema(),
	})
	wantTC404 := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleTrafficFilter(),
		Schema: newSchema(),
	})
	wantTC404.SetId("")

	tc404AssocErr := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleTrafficFilter(),
		Schema: newSchema(),
	})
	wantTC404Assoc := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleTrafficFilter(),
		Schema: newSchema(),
	})

	tc404DeleteErr := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleTrafficFilter(),
		Schema: newSchema(),
	})
	wantTC404Delete := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleTrafficFilter(),
		Schema: newSchema(),
	})
	wantTC404Delete.SetId("")

	tc500DeleteErr := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleTrafficFilter(),
		Schema: newSchema(),
	})
	wantTC500Delete := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		State:  newSampleTrafficFilter(),
		Schema: newSchema(),
	})
	type args struct {
		ctx  context.Context
		d    *schema.ResourceData
		meta interface{}
	}
	tests := []struct {
		name   string
		args   args
		want   diag.Diagnostics
		wantRD *schema.ResourceData
	}{
		{
			name: "returns an error when it receives a 500",
			args: args{
				d: tc500Err,
				meta: api.NewMock(mock.NewErrorResponse(500, mock.APIError{
					Code: "some", Message: "message",
				})),
			},
			want: diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "api error: 1 error occurred:\n\t* some: message\n\n",
				},
			},
			wantRD: wantTC500,
		},
		{
			name: "returns nil and unsets the state when the error is known",
			args: args{
				d: tc404Err,
				meta: api.NewMock(mock.NewErrorResponse(404, mock.APIError{
					Code: "some", Message: "message",
				})),
			},
			want:   nil,
			wantRD: wantTC404,
		},
		{
			name: "returns error when the error is unknown",
			args: args{
				d: tc404AssocErr,
				meta: api.NewMock(
					mock.New200StructResponse(models.TrafficFilterRulesetInfo{
						Associations: []*models.FilterAssociation{
							{ID: ec.String("some id"), EntityType: ec.String("deployment")},
						},
					}),
					mock.NewErrorResponse(500, mock.APIError{
						Code: "some", Message: "message",
					}),
				),
			},
			want: diag.Diagnostics{
				{
					Summary: "api error: 1 error occurred:\n\t* some: message\n\n",
				},
			},
			wantRD: wantTC404Assoc,
		},
		{
			name: "returns nil and unsets the state when the error is known",
			args: args{
				d: tc404DeleteErr,
				meta: api.NewMock(
					mock.New200StructResponse(models.TrafficFilterRulesetInfo{
						Associations: []*models.FilterAssociation{
							{ID: ec.String("some id"), EntityType: ec.String("deployment")},
						},
					}),
					mock.NewErrorResponse(404, mock.APIError{
						Code: "some", Message: "message",
					}),
					mock.New200StructResponse(map[string]interface{}{}),
				),
			},
			want:   nil,
			wantRD: wantTC404Delete,
		},
		{
			name: "returns error when the delete returns a 500 error",
			args: args{
				d: tc500DeleteErr,
				meta: api.NewMock(
					mock.New200StructResponse(models.TrafficFilterRulesetInfo{
						Associations: []*models.FilterAssociation{
							{ID: ec.String("some id"), EntityType: ec.String("deployment")},
						},
					}),
					mock.NewErrorResponse(404, mock.APIError{
						Code: "some", Message: "message",
					}),
					mock.NewErrorResponse(500, mock.APIError{
						Code: "overload", Message: "server at capacity",
					}),
				),
			},
			want: diag.Diagnostics{
				{
					Summary: "api error: 1 error occurred:\n\t* overload: server at capacity\n\n",
				},
			},
			wantRD: wantTC500Delete,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := delete(tt.args.ctx, tt.args.d, tt.args.meta)
			assert.Equal(t, tt.want, got)
			var want interface{}
			if tt.wantRD != nil {
				if s := tt.wantRD.State(); s != nil {
					want = s.Attributes
				}
			}

			var gotState interface{}
			if s := tt.args.d.State(); s != nil {
				gotState = s.Attributes
			}

			assert.Equal(t, want, gotState)
		})
	}
}
