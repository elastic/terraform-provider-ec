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

package v2

import (
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/stretchr/testify/assert"
)

func Test_ensurePartialSnapshotStrategy(t *testing.T) {
	type args struct {
		es *models.ElasticsearchPayload
	}
	tests := []struct {
		name string
		args args
		want *models.ElasticsearchPayload
	}{
		{
			name: "ignores resources with no transient block",
			args: args{es: &models.ElasticsearchPayload{
				Plan: &models.ElasticsearchClusterPlan{},
			}},
			want: &models.ElasticsearchPayload{
				Plan: &models.ElasticsearchClusterPlan{},
			},
		},
		{
			name: "ignores resources with no transient.snapshot block",
			args: args{es: &models.ElasticsearchPayload{
				Plan: &models.ElasticsearchClusterPlan{
					Transient: &models.TransientElasticsearchPlanConfiguration{},
				},
			}},
			want: &models.ElasticsearchPayload{
				Plan: &models.ElasticsearchClusterPlan{
					Transient: &models.TransientElasticsearchPlanConfiguration{},
				},
			},
		},
		{
			name: "Sets strategy to partial",
			args: args{es: &models.ElasticsearchPayload{
				Plan: &models.ElasticsearchClusterPlan{
					Transient: &models.TransientElasticsearchPlanConfiguration{
						RestoreSnapshot: &models.RestoreSnapshotConfiguration{
							SourceClusterID: "some",
						},
					},
				},
			}},
			want: &models.ElasticsearchPayload{
				Plan: &models.ElasticsearchClusterPlan{
					Transient: &models.TransientElasticsearchPlanConfiguration{
						RestoreSnapshot: &models.RestoreSnapshotConfiguration{
							SourceClusterID: "some",
							Strategy:        "partial",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ensurePartialSnapshotStrategy(tt.args.es)
			assert.Equal(t, tt.want, tt.args.es)
		})
	}
}
