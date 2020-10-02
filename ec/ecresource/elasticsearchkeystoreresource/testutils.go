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

package elasticsearchkeystoreresource

import (
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type resDataParams struct {
	Resources map[string]interface{}
	ID        string
}

func newResourceData(t *testing.T, params resDataParams) *schema.ResourceData {
	raw := schema.TestResourceDataRaw(t, newSchema(), params.Resources)
	raw.SetId(params.ID)

	return raw
}

func newSampleElasticsearchKeystore() map[string]interface{} {
	return map[string]interface{}{
		"deployment_id": mock.ValidClusterID,
		"secrets": []interface{}{map[string]interface{}{
			"setting_name": "my_secret",
			"value":        "supersecret",
			"as_file":      true,
		}},
	}
}
