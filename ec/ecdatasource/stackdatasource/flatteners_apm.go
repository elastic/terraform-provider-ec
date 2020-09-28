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

package stackdatasource

import (
	"github.com/elastic/cloud-sdk-go/pkg/models"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

// flattenApmResources takes in Apm resource models and returns its
// flattened form.
func flattenApmResources(res *models.StackVersionApmConfig) []interface{} {
	var m = make(map[string]interface{})

	if res == nil {
		return nil
	}

	if len(res.Blacklist) > 0 {
		m["denylist"] = util.StringToItems(res.Blacklist...)
	}

	if res.CapacityConstraints != nil {
		m["capacity_constraints_max"] = int(*res.CapacityConstraints.Max)
		m["capacity_constraints_min"] = int(*res.CapacityConstraints.Min)
	}

	if len(res.CompatibleNodeTypes) > 0 {
		m["compatible_node_types"] = res.CompatibleNodeTypes
	}

	if res.DockerImage != nil && *res.DockerImage != "" {
		m["docker_image"] = *res.DockerImage
	}

	if len(m) == 0 {
		return nil
	}

	return []interface{}{m}
}
