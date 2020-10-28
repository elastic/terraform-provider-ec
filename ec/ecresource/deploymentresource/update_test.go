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

package deploymentresource

import (
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_hasDeploymentChange(t *testing.T) {
	unchanged := Resource().Data(util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		Schema: newSchema(),
		State:  newSampleDeployment(),
	}).State())

	changesToTrafficFilter := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		Schema: newSchema(),
		State: map[string]interface{}{
			"traffic_filter": []interface{}{"1.1.1.1"},
		},
	})

	changesToName := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		Schema: newSchema(),
		State:  map[string]interface{}{"name": "some name"},
	})

	changesToRegion := util.NewResourceData(t, util.ResDataParams{
		ID:     mock.ValidClusterID,
		Schema: newSchema(),
		State: map[string]interface{}{
			"name":   "some name",
			"region": "some-region",
		},
	})

	type args struct {
		d *schema.ResourceData
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "when a new resource is persisted and has no changes.",
			args: args{d: unchanged},
			want: false,
		},
		{
			name: "when a new resource has some changes in traffic_filter",
			args: args{d: changesToTrafficFilter},
			want: false,
		},
		{
			name: "when a new resource is has some changes in name",
			args: args{d: changesToName},
			want: true,
		},
		{
			name: "when a new resource is has some changes in name",
			args: args{d: changesToRegion},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasDeploymentChange(tt.args.d)
			assert.Equal(t, tt.want, got)
		})
	}
}
