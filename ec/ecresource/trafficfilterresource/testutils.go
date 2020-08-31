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
)

type resDataParams struct {
	Resources map[string]interface{}
	ID        string
}

func newResourceData(t *testing.T, params resDataParams) *schema.ResourceData {
	raw := schema.TestResourceDataRaw(t, NewSchema(), params.Resources)
	raw.SetId(params.ID)

	return raw
}

func newSampleTrafficFilter() map[string]interface{} {
	return map[string]interface{}{
		"name":               "my traffic filter",
		"type":               "ip",
		"include_by_default": false,
		"region":             "us-east-1",
		"rule": []interface{}{
			map[string]interface{}{
				"source": "1.1.1.1",
			},
			map[string]interface{}{
				"source": "0.0.0.0/0",
			},
		},
	}
}
