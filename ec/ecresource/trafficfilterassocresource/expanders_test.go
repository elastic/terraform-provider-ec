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

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/trafficfilterapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_expand(t *testing.T) {
	rd := util.NewResourceData(t, util.ResDataParams{
		Resources: newSampleTrafficFilterAssociation(),
		ID:        "123451",
		Schema:    newSchema(),
	})
	type args struct {
		d *schema.ResourceData
	}
	tests := []struct {
		name string
		args args
		want trafficfilterapi.CreateAssociationParams
	}{
		{
			name: "expands the resource data",
			args: args{d: rd},
			want: trafficfilterapi.CreateAssociationParams{
				ID:         mockTrafficFilterID,
				EntityID:   mock.ValidClusterID,
				EntityType: entityType,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expand(tt.args.d)
			assert.Equal(t, tt.want, got)
		})
	}
}
