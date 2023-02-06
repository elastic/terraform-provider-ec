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

package converters

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
)

func TestFlattenClusterEndpoint(t *testing.T) {
	type args struct {
		metadata *models.ClusterMetadataInfo
	}
	type want struct {
		httpEndpoint  *string
		httpsEndpoint *string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "returns nil when the endpoint info is empty",
			args: args{metadata: &models.ClusterMetadataInfo{}},
		},
		{
			name: "parses the endpoint information",
			args: args{metadata: &models.ClusterMetadataInfo{
				Endpoint: "xyz.us-east-1.aws.found.io",
				Ports: &models.ClusterMetadataPortInfo{
					HTTP:  ec.Int32(9200),
					HTTPS: ec.Int32(9243),
				},
			}},
			want: want{
				httpEndpoint:  ec.String("http://xyz.us-east-1.aws.found.io:9200"),
				httpsEndpoint: ec.String("https://xyz.us-east-1.aws.found.io:9243"),
			},
		},
		{
			name: "parses the some more endpoint information",
			args: args{metadata: &models.ClusterMetadataInfo{
				Endpoint: "rst.us-east-1.aws.found.io",
				Ports: &models.ClusterMetadataPortInfo{
					HTTP:  ec.Int32(10000),
					HTTPS: ec.Int32(20000),
				},
			}},
			want: want{
				httpEndpoint:  ec.String("http://rst.us-east-1.aws.found.io:10000"),
				httpsEndpoint: ec.String("https://rst.us-east-1.aws.found.io:20000"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpEndpoint, httpsEndpoint := ExtractEndpoints(tt.args.metadata)
			assert.Equal(t, tt.want.httpEndpoint, httpEndpoint)
			assert.Equal(t, tt.want.httpsEndpoint, httpsEndpoint)
		})
	}
}
