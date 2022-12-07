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
	"fmt"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// FlattenClusterEndpoint receives a ClusterMetadataInfo, parses the http and
// https endpoints and returns a map with two keys: `http_endpoint` and
// `https_endpoint`
func ExtractEndpointsTF(metadata *models.ClusterMetadataInfo) (httpEndpoint, httpsEndpoint types.String) {
	if metadata == nil || metadata.Endpoint == "" || metadata.Ports == nil {
		return
	}

	if metadata.Ports.HTTP != nil {
		httpEndpoint = types.String{Value: fmt.Sprintf("http://%s:%d", metadata.Endpoint, *metadata.Ports.HTTP)}
	}

	if metadata.Ports.HTTPS != nil {
		httpsEndpoint = types.String{Value: fmt.Sprintf("https://%s:%d", metadata.Endpoint, *metadata.Ports.HTTPS)}
	}

	return
}

// FlattenClusterEndpoint receives a ClusterMetadataInfo, parses the http and
// https endpoints and returns a map with two keys: `http_endpoint` and
// `https_endpoint`
func ExtractEndpoints(metadata *models.ClusterMetadataInfo) (httpEndpoint, httpsEndpoint *string) {
	if metadata == nil || metadata.Endpoint == "" || metadata.Ports == nil {
		return
	}

	if metadata.Ports.HTTP != nil {
		httpEndpoint = ec.String(fmt.Sprintf("http://%s:%d", metadata.Endpoint, *metadata.Ports.HTTP))
	}

	if metadata.Ports.HTTPS != nil {
		httpsEndpoint = ec.String(fmt.Sprintf("https://%s:%d", metadata.Endpoint, *metadata.Ports.HTTPS))
	}

	return
}
